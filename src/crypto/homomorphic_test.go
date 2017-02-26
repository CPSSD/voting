package crypto_test

import (
	"github.com/CPSSD/voting/src/crypto"
	"math/big"
	"testing"
)

func TestHomomorphicProperties(t *testing.T) {

	var tests = []struct {
		input   []*big.Int
		correct *big.Int
	}{
		{[]*big.Int{big.NewInt(1), big.NewInt(1), big.NewInt(-123422), big.NewInt(-2341317)}, big.NewInt(2)},
		{[]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(1), big.NewInt(0)}, big.NewInt(1)},
		{[]*big.Int{big.NewInt(1), big.NewInt(1), big.NewInt(1), big.NewInt(1)}, big.NewInt(4)},
		{[]*big.Int{big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)}, big.NewInt(0)},
		{[]*big.Int{big.NewInt(799), big.NewInt(0), big.NewInt(201), big.NewInt(2)}, big.NewInt(1002)},
	}

	for _, c := range tests {
		got, err := CheckHomomorphicAddition(c.input...)
		if got.Cmp(c.correct) != 0 {
			t.Error("For input", c.input, "expected", c.correct, "got", got, "with error", err, "\n")
		}
	}
}

func CheckHomomorphicAddition(inputs ...*big.Int) (total *big.Int, err error) {

	priv, err := crypto.GenerateKeyPair(512)
	if err != nil {
		return nil, err
	}

	zero := new(big.Int)
	total, err = priv.Encrypt(zero)
	if err != nil {
		return nil, err
	}

	// create encryptions of each input, and add the ciphertexts incrementally
	for _, input := range inputs {
		ciphertext, err := priv.Encrypt(input)
		if err != nil {
			return nil, err
		}
		total, err = priv.AddCipherTexts(ciphertext, total)
		if err != nil {
			return nil, err
		}
	}

	// decrypt the total of the ciphertexts
	total, err = priv.Decrypt(total)
	return total, err
}
