package blockchain

import (
	"strconv"
)

type Chain struct {
	head   *Block
	blocks []Block
}

func NewChain() (c *Chain, err error) {
	c = &Chain{
		head:   NewBlock(),
		blocks: make([]Block, 0),
	}
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
		return
	}

	// if we haven't, create proof
	tr.createProof(proofDifficultyTr)

	// then add tr to a block
	c.addTransaction(tr)

	return

}

// Chain.addTransaction will add a transaction to the
// head of the chain. If the head is then full, it will
// be appended to the chain and cleared for re-use.
func (c *Chain) addTransaction(tr *Transaction) {
	if isFull := c.head.addTransaction(tr); isFull {

		if len(c.blocks) != 0 {
			c.head.Header.ParentHash = c.blocks[len(c.blocks)-1].Proof
		} else {
			c.head.Header.ParentHash = *new([32]byte)
		}

		c.head.createProof(proofDifficultyBl)

		c.blocks = append(c.blocks, *c.head)
		c.head = NewBlock()
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
