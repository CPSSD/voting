package blockchain

import (
	"fmt"
)

type Chain struct {
	head   Block
	blocks []Block
}

func NewChain() (c *Chain, err error) {
	c = &Chain{
		head:   NewBlock(),
		blocks: make([]Block, 0),
	}
	return c
}

// CreateProof is a function for RPC. It should be called
// by a client who wishes to broadcast a vote to peers on
// the network.
func (c *Chain) AddTransaction(tr *Transaction, _ *struct{}) (err error) {

	// have we seen this tr?
	if c.contains(tr) {
		return
	}

	// if we haven't, create proof
	tr.createProof(3)

	// then add tr to a block
	c.addTransaction(tr)

	return
}

// Chain.addTransaction will add a transaction to the
// head of the chain. If the head is then full, it will
// be appended to the chain and cleared for re-use.
func (c *Chain) addTransaction(tr *Transaction) {
	if isFull := c.head.addTransaction(tr); isFull {
		blocks := append(blocks, head)
		c.head = c.head[:0]
	}
}

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
