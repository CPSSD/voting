package crypto_test

import (
	"github.com/CPSSD/voting/src/crypto"
	"math"
	"math/big"
	"testing"
)

func TestSimpleTests(t *testing.T) {

	var above_max, _ = new(big.Int).SetString("922337203685477580700001", 10)
	var tests = []struct {
		input *big.Int
		equal bool
	}{
		{big.NewInt(-1), true},
		{big.NewInt(0), true},
		{big.NewInt(1), true},
		{big.NewInt(math.MaxInt64), true},
		{above_max, true},
		{nil, false},
	}

	for _, c := range tests {
		got, err := CheckEncryption(c.input)
		if got != c.equal {
			t.Error("For input", c.input, "expected", c.equal, "got", got, "with error", err, "\n")
		}

		got, err = CheckNonDeterministicProperties(c.input)
		if got != c.equal {
			t.Error("For input", c.input, "expected", c.equal, "got", got, "with error", err, "\n")
		}
	}

	got, err := CheckNilEncryptionKey()
	if got != true {
		t.Error("For nil key expected error was", crypto.InvalidPublicKeyError, "but got", err, "\n")
	}

	got, err = CheckNilDecryptionKey()
	if got != true {
		t.Error("For nil key expected error was", crypto.InvalidPrivateKeyError, "but got", err, "\n")
	}
}

func CheckEncryption(input *big.Int) (success bool, err error) {

	priv, err := crypto.GenerateKeyPair(512)
	if err != nil {
		return false, err
	}
	_, err = priv.Encrypt(input)
	if err != nil {
		return false, err
	} else {
		return true, err
	}

	_, err = priv.PublicKey.Encrypt(input)
	if err != nil {
		return false, err
	} else {
		return true, err
	}
}

func CheckNilEncryptionKey() (success bool, err error) {

	input := new(big.Int)
	var key crypto.PublicKey
	_, err = key.Encrypt(input)
	if err != crypto.InvalidPublicKeyError {
		return false, err
	}
	return true, err
}

func CheckNilDecryptionKey() (success bool, err error) {

	input := new(big.Int)
	var key crypto.PrivateKey
	_, err = key.Decrypt(input)
	if err != crypto.InvalidPrivateKeyError {
		return false, err
	}
	return true, err
}

func CheckNonDeterministicProperties(input *big.Int) (success bool, err error) {

	priv, err := crypto.GenerateKeyPair(512)
	if err != nil {
		return false, err
	}
	ciphertext_a, err := priv.Encrypt(input)
	if err != nil {
		return false, err
	}
	ciphertext_b, err := priv.Encrypt(input)
	if err != nil {
		return false, err
	}

	// same input should yield different output
	// ie. E(m) != E(m) for same key
	if ciphertext_a.Cmp(ciphertext_b) == 0 {
		return false, err
	}

	deciphered_a, err := priv.Decrypt(ciphertext_a)
	if err != nil {
		return false, err
	}
	deciphered_b, err := priv.Decrypt(ciphertext_b)
	if err != nil {
		return false, err
	}

	// deciphered texts of different encryptions of the same plaintext
	// should be the same
	// ie. D(E(m)) == D(E(m)) for same key
	if deciphered_a.Cmp(deciphered_b) != 0 {
		return false, err
	}

	return true, err
}
