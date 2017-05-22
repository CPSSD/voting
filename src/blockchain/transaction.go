package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"github.com/CPSSD/voting/src/crypto"
	"github.com/CPSSD/voting/src/election"
	"log"
	"time"
)

// Transaction contains the user's Ballot along
// with a TransactionHeader containing information
// about the transaction.
type Transaction struct {
	Header TransactionHeader
	Ballot election.Ballot // the encrypted vote
}

// TransactionHeader contains information required to
// verify a transaction, such as the VoteToken, the
// BallotHash and Signature.
type TransactionHeader struct {
	VoteToken  string           // so that we know what token is authorizing the vote
	BallotHash [32]byte         // hash of the ballot to tie it to the header
	Signature  crypto.Signature // signature of the ballot hash
	Timestamp  uint32           // timestamp so we know when to count this vote for
}

// String representation of a Transaction
func (t Transaction) String() (str string) {
	// str = str + "\n // Time:          " + fmt.Sprint(t.Header.Timestamp)
	// str = str + "\n // Ballot:        " + t.Ballot.String()
	str = str + "\n // Vote Token:    " + string(t.Header.VoteToken)

	return str
}

// NewTransaction will take a filled ballot and encrypt
// its contents with the election key.
func (c *Chain) NewTransaction(token string, ballot *election.Ballot) (t *Transaction) {

	tmp := ballot
	for i, s := range ballot.Selections {
		enc, err := c.conf.ElectionKey.Encrypt(s.Vote)
		if err != nil {
			log.Println("Error while encrypting vote with the public election key")
			log.Fatalln(err)
		}
		tmp.Selections[i].Vote = enc
	}
	ballot = tmp

	t = &Transaction{
		Header: TransactionHeader{
			VoteToken: token,
		},
		Ballot: *ballot,
	}

	var ballot_buf bytes.Buffer
	binary.Write(&ballot_buf, binary.BigEndian, t.Ballot)

	t.Header.BallotHash = sha256.Sum256(ballot_buf.Bytes())
	t.Header.Signature = *crypto.SignHash(&c.conf.PrivateKey, &t.Header.BallotHash)
	t.Header.Timestamp = uint32(time.Now().Unix())

	return t
}

// ValidateSignature will allow a signature of a transaction
// to be validated. The result is returned in the boolean value
// valid.
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
