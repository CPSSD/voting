package blockchain

import (
	"fmt"
	"strconv"
	"sync"
	"time"
    "log"
)

type Chain struct {
    Peers               chan map[string]bool
	TransactionPool     chan []Transaction
	TransactionsReady   chan []Transaction
	CurrentTransactions chan []Transaction
    BlockUpdate         chan BlockUpdate
	SeenTrs             chan map[string]bool
	head                *Block
	blocks              chan []Block
	conf                Configuration
}

func NewChain() (c *Chain, err error) {
    log.Println("This is chain printing to the log")
	c = &Chain{
        Peers:               make(chan map[string]bool, 1),
		TransactionPool:     make(chan []Transaction, 1),
		TransactionsReady:   make(chan []Transaction, 1),
		CurrentTransactions: make(chan []Transaction, 1),
        BlockUpdate:         make(chan BlockUpdate, 1),
		SeenTrs:             make(chan map[string]bool, 1),
		head:                NewBlock(),
		blocks:              make(chan []Block, 1),
	}
	pool := make([]Transaction, 0)
	c.TransactionPool <- pool
	seenTrs := make(map[string]bool, 0)
	c.SeenTrs <- seenTrs
    blocks := make([]Block, 0)
    c.blocks <- blocks
	return c, nil
}

func (c Chain) String() (str string) {
	blocks := <-c.blocks
	c.blocks <- blocks
	for i, b := range blocks {
		str = str + "Block " + strconv.Itoa(i) + ": \b" + b.String() + "\n"
	}
	return "Chain:\n " + str
}

func (c *Chain) schedulePeerSync(syncDelay int, quit chan bool, wg *sync.WaitGroup) {
	timer := time.NewTimer(time.Second)
loop:
	for {
		select {
		case <-quit:
			quit <- true
			wg.Done()
			break loop
		case <-timer.C:
			c.syncPeers()
			timer = time.NewTimer(time.Second * time.Duration(syncDelay))
		}
	}
}

// removeSeenTransactions will return an array of transactions which do not
// occur in the map of seen transaction tokens
func (c *Chain) removeSeenTransactions(trs []Transaction, seen map[string]bool) (out []Transaction) {

	for _, val := range trs {
		if _, ok := seen[string(val.Header.VoteToken[:])]; !ok {
			out = append(out, val)
			seen[string(val.Header.VoteToken[:])] = true
		}
	}

	return out
}

func (c *Chain) scheduleMining(quit, stopMining, startMining, confirmStopped chan bool, wg *sync.WaitGroup) {
    timer := time.NewTimer(time.Second)
start:
fmt.Println("c.scheduleMining: Waiting to start mining")

// wait for the signal to start mining
    _ = <- startMining
    fmt.Println("c.scheduleMining: Got signal to start mining")

loop:
	for {
		select {
        default:
            _ = <-timer.C
            pool := <- c.TransactionPool
            if len(pool) >= blockSize {
                c.TransactionsReady <- pool[:blockSize]
                c.TransactionPool <- pool[blockSize:]
            } else {
                c.TransactionPool <- pool
            }
            timer = time.NewTimer(time.Second * time.Duration(3))
		case <-quit:
			quit <- true
			wg.Done()
			break loop
        case <-stopMining:
            c.CurrentTransactions <- make([]Transaction, 0)
            confirmStopped <- true
            goto start
		case blockPool := <- c.TransactionsReady:

            // make a backup in case we need to stop mining
            tmpTrs := blockPool

            for _, tr := range blockPool {
				tr.createProof(proofDifficultyTr)
				c.head.addTransaction(&tr)
                // TODO: change this to signing the transactions
			}

			blocks := <-c.blocks
            c.blocks <- blocks

			if len(blocks) != 0 {
				c.head.Header.ParentHash = blocks[len(blocks)-1].Proof
			} else {
				c.head.Header.ParentHash = *new([32]byte)
			}

			// compute block hash until created or stopped by new longest chain
            fmt.Println("c.scheduleMining: Working on a hash")
			stopped := c.head.createProof(proofDifficultyBl, stopMining)
            fmt.Println("c.scheduleMining: Finished hashing ")

			if stopped {
                fmt.Println("c.scheduleMining: We were stopped")
                fmt.Println("c.scheduleMining: writing current trs to c.CurrentTransactions")


                // notify what transactions we were working with
                c.CurrentTransactions <- tmpTrs

                fmt.Println("c.scheduleMining: clearing head")

                // clear our head block
				c.head = NewBlock()

                fmt.Println("c.scheduleMining: saying we have stopped ")


                // let the rest of the program know we have stopped mining
                confirmStopped <- true
                fmt.Println("c.scheduleMining: said we have stopped, going to start")

                goto start

                // tp := <-c.TransactionPool
                // updatedPool := c.consolidateTransactions(tp, tmpTrs)
				// ctrs := <-c.CurrentTransactions

			} else {

                // otherwise broadcast our new block to the network
                fmt.Println("c.scheduleMining: We created a block")
				seenTrs := <-c.SeenTrs
				for _, tr := range c.head.Transactions {
					seenTrs[string(tr.Header.VoteToken[:])] = true
				}
				c.SeenTrs <- seenTrs
                blocks := <- c.blocks
				c.blocks <- append(blocks, *c.head)
                bl := *c.head
				c.head = NewBlock()
                fmt.Println("c.scheduleMining: We are going to send the block")
                go c.sendBlock(&bl)
			}
		}
	}
}

// Start will begin some of the background routines required for the running
// of the blockchain such as searching for new peers, and mining blocks.
func (c *Chain) Start(delay int, quit, stop, start, confirm chan bool, w *sync.WaitGroup) {

	// check for new peers every "delay" seconds
	go c.schedulePeerSync(delay, quit, w)

	// be processing transactions aka making blocks
	go c.scheduleMining(quit, stop, start, confirm, w)

	// be ready to process new blocks and consensus forming
	go c.scheduleChainUpdates(quit, stop, start, confirm, w)
}

func (c *Chain) scheduleChainUpdates(quit, stopMining, startMining, confirmStopped chan bool, wg *sync.WaitGroup) {
loop:
	for {
		select {
		case <-quit:
			quit <- true
			wg.Done()
			break loop
		case blu := <-c.BlockUpdate:
            fmt.Println("c.scheduleChainUpdates: Handling block update")
            fmt.Println(blu.LatestBlock)

			blocks := <-c.blocks
            c.blocks <- blocks
            newBlocks := append(blocks, blu.LatestBlock)
            fmt.Println("c.scheduleChainUpdates: Checking if this is the next block in chain...")

            // validate the proposed new chain
            valid, seen := c.validate(&newBlocks)

            if valid {
                fmt.Println("c.scheduleChainUpdates: Received next block in the chain")
            }

            // if it was not the next block in a sequence, it could be from
            // a new longer chain
            if !valid && blu.ChainLength > uint32(len(blocks)) {

                fmt.Println("c.scheduleChainUpdates: Not next block in chain")
                fmt.Println("c.scheduleChainUpdates: Possible new longer chain of len",blu.ChainLength, "compared to",uint32(len(blocks)))

                // get the claimed chain
                fmt.Println("c.scheduleChainUpdates: Getting alt chain...")
                altChain, err := c.getChainUpdateFrom(blu.Peer)
                if err != nil {
                    fmt.Println("c.scheduleChainUpdates: Error getting alt chain!")
                    continue
                }

                // make sure it is longer
                fmt.Println("c.scheduleChainUpdates: Checking length of alt chain...")
                if len(*altChain) < len(blocks) {
                    fmt.Println("c.scheduleChainUpdates: Chain was not longer!")
                    continue
                }

                // validate the new chain
                newBlocks = *altChain
                fmt.Println("c.scheduleChainUpdates: Validating alt chain...")

                valid, seen = c.validate(altChain)
                if valid {
                    log.Println("c.scheduleChainUpdates: New longer valid chain received")
                }
            }

            // if newBlocks is a valid chain...
            if valid {

                // tell the chain to stop mining
                fmt.Println("c.scheduleChainUpdates: Telling to stop mining")
                stopMining <- true
                // confirm that it has stopped
                fmt.Println("c.scheduleChainUpdates: Confirming stopped mining")
                _ = <- confirmStopped
                fmt.Println("c.scheduleChainUpdates: We have stopped mining")

                // set the new chain of blocks
                fmt.Println("c.scheduleChainUpdates: Waiting to delete old blocks...")
                _ = <- c.blocks
                fmt.Println("c.scheduleChainUpdates: We are now writing new blocks")
                c.blocks <- newBlocks


                // set the map of seen transactions (trs in valid blocks)
                fmt.Println("c.scheduleChainUpdates: Waiting to delete old seen trs...")
                _ = <- c.SeenTrs
                fmt.Println("c.scheduleChainUpdates: We are now writing new seen trs")
                c.SeenTrs <- seen

                // set the pool of transactions still to be mined
                fmt.Println("c.scheduleChainUpdates: Waiting to copy tr pool...")

                oldPool := <- c.TransactionPool

                fmt.Println("c.scheduleChainUpdates: Waiting to copy current trs...")
                currentTrs := <- c.CurrentTransactions

                allTrs := append(oldPool, currentTrs...)
                newPool := c.removeSeenTransactions(allTrs, seen)

                fmt.Println("c.scheduleChainUpdates: We are now writing the new tr pool...")
                c.TransactionPool <- newPool


                fmt.Println("c.scheduleChainUpdates: We wrote the new pool and are sending the update to peers")
                go c.sendBlock(&blu.LatestBlock)
                fmt.Println("c.scheduleChainUpdates: We have sent the update")


                // we have finished making changes, so we can tell the
                // chain to start mining again
                fmt.Println("c.scheduleChainUpdates: We are informing to start mining again...")

                startMining <- true
                fmt.Println("c.scheduleChainUpdates: We have informed to start mining.")

            } else {
                fmt.Println("c.scheduleChainUpdates: Chain was not valid")

            }
		}
	}
}

func (c *Chain) validate(blocks *[]Block) (valid bool, seen map[string]bool) {

    seen = make(map[string]bool, 0)
    parent := *new([32]byte)

    for _, bl := range *blocks {

        // validate the transactions in the block
        for _, tr := range bl.Transactions {
            if _, ok := seen[string(tr.Header.VoteToken[:])]; ok {
                return false, seen
            }
            seen[string(tr.Header.VoteToken[:])] = true
            // TODO: should also verify signatures once implemented
        }

        // verify the hash of the block
        valid, hash := bl.validate(parent)

        if !valid {
            return false, seen
        }
        parent = hash
    }
    return true, seen
}

// TODO: check the chain in reverse order ie. most
// recent blocks first: hypothesis is that if a
// transaction has been seen before, it will be
// seen more recently.
func (c *Chain) contains(t *Transaction) bool {
	blocks := <-c.blocks
	for _, b := range blocks {
		if b.contains(t) {
			c.blocks <- blocks
			return true
		}
	}
	c.blocks <- blocks
	return false
}
