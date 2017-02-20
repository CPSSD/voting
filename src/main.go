package main

import (
	"crypto/rand"
	"fmt"
	"github.com/CPSSD/voting/src/utils"
	"math/big"
)

func main() {
	fmt.Println("Hello voting test file")

	priv, err := utils.GenerateKeyPair(512)
	utils.Check(err)

    fmt.Println(priv.Lambda, priv.PublicKey.N)

	upperBound := big.NewInt(1000000)

	plaintext, err := rand.Int(rand.Reader, upperBound)
	utils.Check(err)

	fmt.Println("\nplaintext:", plaintext, "\n")

	ciphertext := utils.Encrypt(plaintext, &priv.PublicKey)

	fmt.Println("\nplaintext:", plaintext, "\n", "\nciphertext:", ciphertext, "\n")

	deciphered := utils.Decrypt(ciphertext, priv)

	fmt.Println("\nplaintext:", plaintext, "\n", "\nciphertext:", ciphertext, "\ndeciphered:", deciphered, "\n")

}
