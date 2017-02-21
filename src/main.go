package main

import (
	"crypto/rand"
	"fmt"
    "github.com/CPSSD/voting/src/utils"
	"github.com/CPSSD/voting/src/crypto"
	"math/big"
)

func main() {

	priv, err := crypto.GenerateKeyPair(1024)
	utils.Check(err)

	upperBound := big.NewInt(10000000000)

	plaintext, err := rand.Int(rand.Reader, upperBound)
	utils.Check(err)

	fmt.Println("\nplaintext:", plaintext, "\n")

	ciphertext, _ := crypto.Encrypt(plaintext, &priv.PublicKey)

	fmt.Println("\nciphertext:", ciphertext, "\n")

	deciphered, _ := crypto.Decrypt(ciphertext, priv)

	fmt.Println("\ndeciphered:", deciphered, "\n")

}
