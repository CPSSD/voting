package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
	"strings"
	"time"
)

type Transaction struct {
	Header    TransactionHeader
	Ballot    []byte
	Signature []byte
	Proof     [32]byte
}

type TransactionHeader struct {
	VoteToken  []byte   // so that we know what token is authorizing the vote
	BallotSize uint32   // in case we lose track of the size of our ballot
	BallotHash [32]byte // hash of the ballot to tie it to the header
	Timestamp  uint32   // timestamp so we know when to count this vote for
	Nonce      uint32   // the incremented value for proof of work
}

func (t Transaction) String() (str string) {
	// str = str + "\n // Time:          " + fmt.Sprint(t.Header.Timestamp)
	// str = str + "\n // Proof of Work: " + hex.EncodeToString(t.Proof[:5]) + "..."
	str = str + "\n // Ballot:        " + string(t.Ballot[:])
	str = str + "\n // Vote Token:    " + string(t.Header.VoteToken[:])
	// str = str + "\n // Nonce:         " + fmt.Sprint(t.Header.Nonce)
	// str = str + "\n"
	return str
}

func NewTransaction(token, ballot []byte) (t *Transaction) {
	t = &Transaction{
		Header: TransactionHeader{
			VoteToken: token,
		},
		Ballot: ballot,
	}

	t.Header.BallotSize = uint32(len(t.Ballot))
	t.Header.BallotHash = sha256.Sum256(t.Ballot)
	t.Header.Timestamp = uint32(time.Now().Unix())
	t.Header.Nonce = uint32(0)

	return t
}

func (t *Transaction) createProof(prefixLen int) (nonce uint32) {

	// We need the first prefixLen characters of the hex
	// representation of hash(t) to be equal to 0 for our proof
	// to be valid.

	altT := t
	prefix := strings.Repeat("0", prefixLen)

	for {
		var data []byte
		var b bytes.Buffer
		enc := gob.NewEncoder(&b)
		err := enc.Encode(&altT.Header)
		if err != nil {
			log.Fatalln(err)
		}
		data = b.Bytes()

		hash := sha256.Sum256(data)
		if checkProof(prefix, prefixLen, hash) {
			t.Proof = hash
			break
		}
		altT.Header.Nonce++
	}
	return altT.Header.Nonce
}

func checkProof(prefix string, len int, hash [32]byte) bool {
	s := hex.EncodeToString(hash[:])
	return s[:len] == prefix
}
