package blockchain

import (
	"fmt"
	"strconv"
	"sync"
	"time"
    "reflect"
)

type Chain struct {
	TransactionPool chan []Transaction
    TransactionReady chan bool
    CurrentTransactions chan []Transaction
	head            *Block
	blocks          []Block
	conf            Configuration
}

func NewChain() (c *Chain, err error) {
	c = &Chain{
        TransactionPool: make(chan []Transaction, 1),
		TransactionReady: make(chan bool, 1),
        CurrentTransactions: make(chan []Transaction, 1),
		head:            NewBlock(),
		blocks:          make([]Block, 0),
	}
	pool := make([]Transaction, 0)
    c.TransactionPool <- pool
	c.CurrentTransactions <- pool
	return c, nil
}

func (c Chain) String() (str string) {
	for i, b := range c.blocks {
		str = str + "Block " + strconv.Itoa(i) + ": \b" + b.String() + "\n"
	}
	return "Chain:\n " + str
}

// CreateProof is a function for RPC. It should be called
// by a client who wishes to broadcast a vote to peers on
// the network.

func (c *Chain) AddTransaction(tr *Transaction, _ *struct{}) (err error) {

	// have we seen this tr?
	if c.contains(tr) {
		fmt.Printf("We have seen this transaction already:\n%s\n", tr.String())
		return
	}

	// if we haven't, add it to the pool
	//c.TransactionPool <- tr

	// create proof
	// tr.createProof(proofDifficultyTr)

	// then add tr to a block
	// c.addTransaction(tr)

	return

}

func (c *Chain) Start(delay int, quit, stop chan bool, w *sync.WaitGroup) {

	// be updating peers
	go func(syncDelay int, quit chan bool, wg *sync.WaitGroup) {
		timer := time.NewTimer(time.Second)
	loop:
		for {
			select {
			case <-quit:
				quit <- true
				wg.Done()
				break loop
			case <-timer.C:
				go func() {
					c.syncPeers()
				}()
				timer = time.NewTimer(time.Second * time.Duration(syncDelay))

			}
		}
	}(delay, quit, w)

	// be processing transactions aka making blocks
	go func(quit, stop chan bool, wg *sync.WaitGroup) {
	loop:
		for {
			select {
			case <-quit:
				quit <- true
				wg.Done()
				break loop
			case <- c.TransactionReady:
                fmt.Println("Transaction ready")
                pool := <- c.TransactionPool
                _ = <- c.CurrentTransactions

				if len(pool[blockSize:]) >= blockSize {
                    c.TransactionReady <- true
				}

                c.TransactionPool <- pool[blockSize:]
                c.CurrentTransactions <- pool[:blockSize]

                blockPool := pool[:blockSize]

				fmt.Println("Computing tr hashes")

                for _, tr := range blockPool {
					tr.createProof(proofDifficultyTr)
					c.head.addTransaction(&tr)
				}
				if len(c.blocks) != 0 {
					c.head.Header.ParentHash = c.blocks[len(c.blocks)-1].Proof
				} else {
					c.head.Header.ParentHash = *new([32]byte)
				}

				fmt.Println("Computing block hash")

                start := time.Now()
				// create the proof for the chain
                stopped := c.head.createProof(proofDifficultyBl, stop)

                elapsed := time.Since(start)
                fmt.Printf("Block hash took %s\n", elapsed)

                if stopped {
    				c.head = NewBlock()
                    tp := <- c.TransactionPool
                    _ = <- c.CurrentTransactions

                    for _, t := range blockPool {
                        inPool := false
                        for _, tr := range tp {
                    		if reflect.DeepEqual(t.Header.VoteToken, tr.Header.VoteToken) {
                    			inPool = true
                    			break
                    		}
                    	}
                    	if inPool || c.contains(&t) {
                    		continue
                    	}
                        tp = append(tp, t)
                    }

                    c.TransactionPool <- tp
                    c.CurrentTransactions <- make([]Transaction, 0)
                    fmt.Println("Block hashing interrupted")
                } else {
                    c.blocks = append(c.blocks, *c.head)
    				c.head = NewBlock()
                    fmt.Println("Done hashing")
                }


				// send the block to our peers
				// c.sendBlock(*c.head)

				// clear the head block
			}
		}
	}(quit, stop, w)
}

// Chain.addTransaction will add a transaction to the
// head of the chain. If the head is then full, it will
// be appended to the chain and cleared for re-use.
// func (c *Chain) addTransaction(tr *Transaction) {
// 	if isFull := c.head.addTransaction(tr); isFull {
//
// 		// The block is full, so let us:
//
// 		// link it to the previous block in the chain
// 		if len(c.blocks) != 0 {
// 			c.head.Header.ParentHash = c.blocks[len(c.blocks)-1].Proof
// 		} else {
// 			c.head.Header.ParentHash = *new([32]byte)
// 		}
//
// 		// create the proof for the chain
// 		c.head.createProof(proofDifficultyBl)
//
// 		// add it to our chain
// 		c.blocks = append(c.blocks, *c.head)
//
// 		// send the block to our peers
// 		//c.sendBlock(*c.head)
//
// 		// clear the head block
// 		c.head = NewBlock()
// 	}
// }

// TODO: check the chain in reverse order ie. most
// recent blocks first: hypothesis is that if a
// transaction has been seen before, it will be
// seen more recently.
func (c *Chain) contains(t *Transaction) bool {
	for _, b := range c.blocks {
		if b.contains(t) {
			return true
		}
	}
	return false
}
