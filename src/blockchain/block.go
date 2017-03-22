package blockchain

import (
//"fmt"
)

type Block struct {
	transactions []Transaction
}

func NewBlock() (b *Block) {
	b = &Block{
		transactions: make([]Transaction, 0, blockSize),
	}
	return b
}

func (b *block) addTransaction(t *Transaction) (isFull bool) {
	b.transactions = append(b.transactions, t)
	return len(b.transactions) == cap(b.transactions)
}

func (b *block) contains(t *Transaction) bool {
	for _, tr := range transactions {
		if t.Header.VoteToken == tr.Header.VoteToken {
			return true
		}
	}
	return false
}
