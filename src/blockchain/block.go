package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	transactions []Transaction
	Header       BlockHeader
	Proof        [32]byte
}

type BlockHeader struct {
	MerkleHash [32]byte
	ParentHash [32]byte
	Timestamp  uint32
	Nonce      uint32
}

func NewBlock() (b *Block) {
	b = &Block{
		transactions: make([]Transaction, 0, blockSize),
	}
	return b
}

func (b Block) String() (str string) {
	str = str + "\n // Time:          " + fmt.Sprint(b.Header.Timestamp)
	str = str + "\n // Proof of Work: " + hex.EncodeToString(b.Proof[:5]) + "..."
	str = str + "\n // Merkle Hash:   " + hex.EncodeToString(b.Header.MerkleHash[:])
	str = str + "\n // Parent Proof:  " + hex.EncodeToString(b.Header.ParentHash[:])
	str = str + "\n // Nonce:         " + fmt.Sprint(b.Header.Nonce)
	str = str + "\n\n"
	for i, t := range b.transactions {
		str = str + "Transaction " + strconv.Itoa(i) + ": " + t.String() + "\n"
	}
	return str
}

func (b *Block) addTransaction(t *Transaction) (isFull bool) {
	b.transactions = append(b.transactions, *t)
	return len(b.transactions) == cap(b.transactions)
}

func (b *Block) createProof(prefixLen int, stop chan bool) (stopped bool) {

	merkle := b.MerkleHash()

	altB := b
	prefix := strings.Repeat("0", prefixLen)

	b.Header.Timestamp = uint32(time.Now().Unix())

    loop:
    for {
        select {
        case <-stop:
            return true
        default:
    		var data []byte
    		var buf bytes.Buffer
    		enc := gob.NewEncoder(&buf)
    		err := enc.Encode(&altB.Header)
    		if err != nil {
    			panic(err)
    		}

    		data = append(merkle, buf.Bytes()...)

    		hash := sha256.Sum256(data)
    		if checkProof(prefix, prefixLen, hash) {
    			b.Proof = hash
    			break loop
    		}
    		altB.Header.Nonce++
        }
	}
	return false
}

func (b *Block) MerkleHash() (hash []byte) {
	h := merkleHash(b.transactions)
	b.Header.MerkleHash = h
	return h[:]
}

func merkleHash(trs []Transaction) (hash [32]byte) {
	l := len(trs)
	if l == 1 {
		return trs[0].Proof
	}
	hl := merkleHash(trs[:l/2])
	hr := merkleHash(trs[l/2:])
	return sha256.Sum256(append(hl[:], hr[:]...))
}

func (b *Block) contains(t *Transaction) bool {
	for _, tr := range b.transactions {
		if reflect.DeepEqual(t.Header.VoteToken, tr.Header.VoteToken) {
			return true
		}
	}
	return false
}
