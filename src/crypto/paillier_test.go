package crypto_test

import (
	"fmt"
	"github.com/CPSSD/voting/src/crypto"
	"testing"
)

func TestSimpleTest(t *testing.T) {
	var tests = []struct {
		equal bool
	}{
        {false},
	}
	for _, c := range tests {
		got, err := CheckPaillier()
		if got != c.equal {
			t.Error("Simple test failed.\n")
			fmt.Printf("Error: %s\n", err)
		}
	}
}

func CheckPaillier() (success bool, err error) {

    // if the file doesn't exist, the test will pass
    err = crypto.EncryptFile("simple-non-existant-file")
    return false, err

}
