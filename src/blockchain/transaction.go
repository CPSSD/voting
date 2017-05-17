package blockchain

import (
	"crypto/sha256"
	"github.com/CPSSD/voting/src/crypto"
	"log"
	"math/big"
	"time"
)

type Transaction struct {
	Header TransactionHeader
	Ballot *big.Int // the encrypted vote
}

type TransactionHeader struct {
	VoteToken  string           // so that we know what token is authorizing the vote
	BallotHash [32]byte         // hash of the ballot to tie it to the header
	Signature  crypto.Signature // signature of the ballot hash
	Timestamp  uint32           // timestamp so we know when to count this vote for
}

func (t Transaction) String() (str string) {
	// str = str + "\n // Time:          " + fmt.Sprint(t.Header.Timestamp)
	str = str + "\n // Ballot:        " + t.Ballot.String()
	str = str + "\n // Vote Token:    " + string(t.Header.VoteToken)

	return str
}

func (c *Chain) NewTransaction(token string, vote *big.Int) (t *Transaction) {

	// TODO: encrypt the vote using the public election key here to form ballot
	ballot := vote

	t = &Transaction{
		Header: TransactionHeader{
			VoteToken: token,
		},
		Ballot: ballot,
	}

	t.Header.BallotHash = sha256.Sum256(t.Ballot.Bytes())
	t.Header.Signature = *crypto.SignHash(&c.conf.PrivateKey, &t.Header.BallotHash)
	t.Header.Timestamp = uint32(time.Now().Unix())

	return t
}

func (c *Chain) ValidateSignature(t *Transaction) (valid bool) {
	pubkey, ok := c.conf.VoteTokens[t.Header.VoteToken]
	if !ok {
		log.Println("Transaction contains fake vote token:", t.Header.VoteToken)
		return false
	}
	valid = crypto.Verify(&pubkey, &t.Header.BallotHash, &t.Header.Signature)
	if !valid {
		log.Println("Transaction signature invalid")
	}
	return valid
}
