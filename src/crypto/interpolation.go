package crypto

import (
	"math/big"
)

func Interpolate(points []Secret, prime *big.Int) (secret *big.Int, err error) {

	secret = new(big.Int)

	// get the sum from j = 0, to k-1 of:
	// f(xj) . the product from m = 0, m != j, to k-1 of:
	// xm /( xm - xj )

	for _, j := range points {
		subProduct := calculateProduct(j, points, prime)
		subSecret := new(big.Int).Mul(j.Y, subProduct)
		//fmt.Println("(",j.Y,".",subProduct,") +",)
		secret = new(big.Int).Add(secret, subSecret)
	}

	secret = new(big.Int).Mod(secret, prime)
	return
}

func calculateProduct(j Secret, points []Secret, prime *big.Int) (product *big.Int) {

	product = big.NewInt(1)

	for _, s := range points {
		if s.X.Cmp(j.X) != 0 {

			negated := new(big.Int).Mul(s.X, big.NewInt(-1))
			modInverse := new(big.Int).ModInverse(new(big.Int).Sub(j.X, s.X), prime)
			term := new(big.Int).Mul(negated, modInverse)
			product = new(big.Int).Mul(product, term)
		}
	}

	return
}
