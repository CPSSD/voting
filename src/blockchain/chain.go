package blockchain

import (
	"github.com/CPSSD/voting/src/crypto"
	"log"
	"strconv"
	"sync"
	"time"
)

type Chain struct {
	Peers               chan map[string]bool
	TransactionPool     chan []Transaction
	TransactionsReady   chan []Transaction
	CurrentTransactions chan []Transaction
	BlockUpdate         chan BlockUpdate
	KeyShares           chan map[string]ElectionSecret
	SeenTrs             chan map[string]bool
	head                *Block
	blocks              chan []Block
	conf                Configuration
}

func NewChain() (c *Chain, err error) {
	c = &Chain{
		Peers:               make(chan map[string]bool, 1),
		TransactionPool:     make(chan []Transaction, 1),
		TransactionsReady:   make(chan []Transaction, 1),
		CurrentTransactions: make(chan []Transaction, 1),
		BlockUpdate:         make(chan BlockUpdate, 1),
		KeyShares:           make(chan map[string]ElectionSecret, 1),
		SeenTrs:             make(chan map[string]bool, 1),
		head:                NewBlock(),
		blocks:              make(chan []Block, 1),
	}
	pool := make([]Transaction, 0)
	c.TransactionPool <- pool
	seenTrs := make(map[string]bool, 0)
	c.SeenTrs <- seenTrs
	keyShares := make(map[string]ElectionSecret, 0)
	c.KeyShares <- keyShares
	blocks := make([]Block, 0)
	c.blocks <- blocks
	return c, nil
}

func (c *Chain) ReconstructElectionKey() {
	shares := <-c.KeyShares
	c.KeyShares <- shares

	lambdaShares := make([]crypto.Share, len(shares))
	muShares := make([]crypto.Share, len(shares))

	var i int
	for _, s := range shares {
		lambdaShares[i] = s.Lambda
		muShares[i] = s.Mu
		i++
	}

	lambda, err := crypto.Interpolate(lambdaShares, c.conf.ElectionLambdaModulus)
	if err != nil {
		log.Println("Error reconstructing the lambda value for the election key")
		log.Fatalln(err)
	}

	mu, err := crypto.Interpolate(muShares, c.conf.ElectionMuModulus)
	if err != nil {
		log.Println("Error reconstructing the mu value for the election key")
		log.Fatalln(err)
	}

	c.conf.ElectionKey.Lambda = lambda
	c.conf.ElectionKey.Mu = mu
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
			log.Println("Peer syncing process received signal to shutdown")
			quit <- true
			wg.Done()
			break loop
		case <-timer.C:
			log.Println("About to sync peers")
			c.syncPeers()
			timer = time.NewTimer(time.Second * time.Duration(syncDelay))
		}
	}
}

// removeSeenTransactions will return an array of transactions which do not
// occur in the map of seen transaction tokens
func (c *Chain) removeSeenTransactions(trs []Transaction, seen map[string]bool) (out []Transaction) {

	for _, tr := range trs {
		if _, ok := seen[tr.Header.VoteToken]; !ok {
			if valid := c.ValidateSignature(&tr); valid {
				out = append(out, tr)
				seen[tr.Header.VoteToken] = true
			}
		}
	}

	return out
}

func (c *Chain) scheduleMining(quit, stopMining, startMining, confirmStopped chan bool, wg *sync.WaitGroup) {
	timer := time.NewTimer(time.Second)
start:

	log.Println("Waiting for the signal to start mining")
	_ = <-startMining
	log.Println("Got the signal, about to start mining")

loop:
	for {
		select {

		default:
			// By default, we wait for timer to expire, then we will check
			// to see if there are enough transactions in the pool that we
			// can create a block from.
			_ = <-timer.C

			// Get the pool and see if it is longer than the constant blockSize
			pool := <-c.TransactionPool
			if len(pool) >= blockSize {
				// if so, we will put blockSize worth of transactions into
				// the TransactionsReady channel, and replace the rest of the
				// transactions
				c.TransactionsReady <- pool[:blockSize]
				c.TransactionPool <- pool[blockSize:]
			} else {
				c.TransactionPool <- pool
			}
			// Reset the timer
			timer = time.NewTimer(time.Second * time.Duration(hashingDelay))

		case <-quit:
			log.Println("Mining process received signal to shutdown")
			quit <- true
			wg.Done()
			break loop

		case <-stopMining:
			log.Println("Mining process received signal to stop activities")
			c.CurrentTransactions <- make([]Transaction, 0)
			confirmStopped <- true
			goto start

		case blockPool := <-c.TransactionsReady:
			log.Println("We have enough transactions to create a block")
			// make a backup in case we need to stop mining
			tmpTrs := blockPool

			for _, tr := range blockPool {
				// signatures have been verified before being added to the pool
				c.head.addTransaction(&tr)
			}

			blocks := <-c.blocks
			c.blocks <- blocks

			if len(blocks) != 0 {
				c.head.Header.ParentHash = blocks[len(blocks)-1].Proof
			} else {
				c.head.Header.ParentHash = *new([32]byte)
			}

			// compute block hash until created or stopped by new longest chain
			stopped := c.head.createProof(proofDifficultyBl, stopMining)

			if stopped {

				log.Println("Mining process received signal to stop activities")

				// notify what transactions we were working with
				c.CurrentTransactions <- tmpTrs
				c.head = NewBlock()
				confirmStopped <- true

				goto start

			} else {

				log.Println("Mining process created a block")

				seenTrs := <-c.SeenTrs
				for _, tr := range c.head.Transactions {
					seenTrs[tr.Header.VoteToken] = true
				}
				c.SeenTrs <- seenTrs

				blocks := <-c.blocks
				c.blocks <- append(blocks, *c.head)

				bl := *c.head
				c.head = NewBlock()

				go c.sendBlock(&bl)
			}
		}
	}
}

// Start will begin some of the background routines required for the running
// of the blockchain such as searching for new peers, and mining blocks.
func (c *Chain) Start(delay int, quit, stop, start, confirm chan bool, w *sync.WaitGroup) {

	// check for new peers every "delay" seconds
	log.Println("Starting peer syncing process...")
	go c.schedulePeerSync(delay, quit, w)

	// be processing transactions aka making blocks
	log.Println("Starting mining process...")
	go c.scheduleMining(quit, stop, start, confirm, w)

	// be ready to process new blocks and consensus forming
	log.Println("Starting chain management process...")
	go c.scheduleChainUpdates(quit, stop, start, confirm, w)

	// be listening for new shares
	log.Println("Starting key share collection process...")
	go c.scheduleKeyShareBroadcasting(delay, quit, w)
}

func (c *Chain) scheduleKeyShareBroadcasting(delay int, quit chan bool, wg *sync.WaitGroup) {
	timer := time.NewTimer(time.Second * time.Duration(delay))
loop:
	for {
		select {
		case <-quit:
			log.Println("Key share collection process received signal to shutdown")
			quit <- true
			wg.Done()
			break loop

		case <-timer.C:
			log.Println("About to broadcast key shares")
			c.broadcastKeyShares()
			timer = time.NewTimer(time.Second * time.Duration(delay))

		}
	}
}

func (c *Chain) scheduleChainUpdates(quit, stopMining, startMining, confirmStopped chan bool, wg *sync.WaitGroup) {
loop:
	for {
		select {
		case <-quit:
			log.Println("Chain update process received signal to shutdown")
			quit <- true
			wg.Done()
			break loop

		case blu := <-c.BlockUpdate:
			log.Println("Handling block update")

			blocks := <-c.blocks
			c.blocks <- blocks
			newBlocks := append(blocks, blu.LatestBlock)

			// validate the proposed new chain
			valid, seen := c.validate(&newBlocks)

			if valid {

				log.Println("Update contains valid next block")

			} else if !valid && blu.ChainLength > uint32(len(blocks)) {

				log.Println("Possible new longer chain;", blu.ChainLength, "vs", uint32(len(blocks)))
				log.Println("Getting alt chain")
				altChain, err := c.getChainUpdateFrom(blu.Peer)
				if err != nil {
					log.Println("There was a problem getting the alt chain")
					continue
				}

				// make sure it is longer
				if len(*altChain) < len(blocks) {
					log.Println("Alt chain is shorter")
					continue
				}

				// validate the new chain
				newBlocks = *altChain

				valid, seen = c.validate(altChain)
				if valid {
					log.Println("Alt chain is valid")
				}
			}

			// if newBlocks is a valid chain...
			if valid {

				log.Println("Sending signal to stop mining")
				stopMining <- true

				_ = <-confirmStopped
				log.Println("We have stopped mining")

				// set the new chain of blocks
				oldBlocks := <-c.blocks
				c.blocks <- newBlocks

				// set the new map of seen transactions
				_ = <-c.SeenTrs
				c.SeenTrs <- seen

				// set the new pool of transactions still to be mined
				oldPool := <-c.TransactionPool
				currentTrs := <-c.CurrentTransactions

				oldChainTrs := extractTransactions(&oldBlocks)

				allTrs := append(oldPool, currentTrs...)
				allTrs = append(allTrs, *oldChainTrs...)

				newPool := c.removeSeenTransactions(allTrs, seen)

				// TODO: broadcast new pool to peers (share workload)

				c.TransactionPool <- newPool

				go c.sendBlock(&blu.LatestBlock)

				log.Println("Sending signal to start mining again")
				startMining <- true
			} else {
				log.Println("Alt chain was not valid")
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
			if valid := c.ValidateSignature(&tr); !valid {
				log.Println("Invalid chain - badly signed transaction:", tr.Header.VoteToken)
				return false, seen
			}
			if _, ok := seen[tr.Header.VoteToken]; ok {
				log.Println("Invalid chain - duplicated transactions:", tr.Header.VoteToken)
				return false, seen
			}
			seen[tr.Header.VoteToken] = true
		}

		valid, hash := bl.validate(parent)

		if !valid {
			log.Println("Invalid chain - bad hash of block to parent")
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
