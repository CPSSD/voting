package utils

import (
    "io/ioutil"
)

// only simple function for testing purposes
func EncryptFile(filepath string) (err error) {

	// Open the item to be encrypted (the plaintext)
	_, err = ioutil.ReadFile(filepath)

    return err
}
