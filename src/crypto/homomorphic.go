package crypto

import (
	"math/big"
)

func (key *PrivateKey) AddCipherTexts(ciphertexts ...*big.Int) (total *big.Int, err error) {
	total, err = key.PublicKey.AddCipherTexts(ciphertexts...)
	return
}

// Accepts one or more ciphertexts and homomorphically sums them
func (key *PublicKey) AddCipherTexts(ciphertexts ...*big.Int) (total *big.Int, err error) {

	if err = key.Validate(); err != nil {
		return nil, err
	}

	// create an encryption of voting value zero to start off
	zero := new(big.Int)
	total, err = key.Encrypt(zero)
	if err != nil {
		return nil, err
	}

	// D(E(m1,r1).E(m2,r2) mod n^2) = m1 + m2 mod n
	for _, ciphertext := range ciphertexts {
		total = new(big.Int).Mul(total, ciphertext)
		total.Mod(total, key.NSquared)
	}

	return total, nil
}
