package crypto_test

import (
	"github.com/CPSSD/voting/src/crypto"
	"testing"
)

func TestSecretSharing(t *testing.T) {

	success, err := CheckSecretSharing()
	if !success {
		t.Error("Expected ", !success, "got", success, "with err", err, "\n")
	}
}

func CheckSecretSharing() (success bool, err error) {

	priv, err := crypto.GenerateKeyPair(10)
	if err != nil {
		return false, err
	}

	_, _, err = crypto.DivideSecret(priv.Lambda, 2, 2)
	if err != nil {
		return false, err
	}

	return true, err
}
