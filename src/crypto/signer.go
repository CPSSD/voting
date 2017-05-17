package crypto

import (
	"crypto/dsa"
	"crypto/rand"
	"log"
	"math/big"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

func CreateKey() (privateKey *dsa.PrivateKey) {
	params := new(dsa.Parameters)

	err := dsa.GenerateParameters(params, rand.Reader, dsa.L2048N256)
	if err != nil {
		log.Println("Could not generate DSA parameters")
		log.Fatalln(err)
	}

	privateKey = new(dsa.PrivateKey)
	privateKey.PublicKey.Parameters = *params
	dsa.GenerateKey(privateKey, rand.Reader)

	return privateKey
}

func SignHash(privateKey *dsa.PrivateKey, hash *[32]byte) (sig *Signature) {

	r := big.NewInt(0)
	s := big.NewInt(0)

	r, s, err := dsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		log.Println("Error signing the hash")
		log.Fatalln(err)
	}

	sig = &Signature{
		R: r,
		S: s,
	}

	return sig
}

func Verify(pubkey *dsa.PublicKey, hash *[32]byte, sig *Signature) (valid bool) {

	return dsa.Verify(pubkey, hash[:], sig.R, sig.S)
}
