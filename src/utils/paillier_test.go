package utils_test

import (
	"fmt"
	"github.com/CPSSD/voting/src/utils"
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
    err = utils.EncryptFile("simple-non-existant-file")
    return false, err

}
