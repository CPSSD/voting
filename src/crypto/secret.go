package crypto

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// polynomial struct represents a slice of monomials.
type polynomial struct {
	monomials []monomial
}

// String returns a formatted string representing a polynomial
// in the format:
// f(x) = v1.x^d1 + v2.x^d2 + ... + vn.x^dn
// for a polynomial of n degrees.
func (p polynomial) String() (str string) {
	for _, m := range p.monomials {
		str = str + " " + m.String() + " +"
	}
	str = strings.TrimRight(str, "+")
	return "f(x) = " + strings.TrimSpace(str)
}

// monomial struct represents a single degree of a
// larger polynomial.
type monomial struct {
	Value  *big.Int
	Degree *big.Int
}

// Share represents a secret point (X, Y) in a polynomial.
// A Share is required to reconstruct the polynomial based
// on the original secret value. The X value is some arbitrary
// incremental value, where X > 0. The Y value is the result of
// solving the polynomial f(X) = Y.
type Share struct {
	X *big.Int
	Y *big.Int
}

// String returns a formatted string representing a monomial
// in the format:
// v.x^d
// It is used in building the larger polynomial string
// representation.
func (m monomial) String() string {
	return m.Value.String() + ".x^" + m.Degree.String()
}

// solve will return the result of f(x) = y for a polynomial
// function p.
func (p polynomial) solve(x *big.Int) (y *big.Int) {
	y = new(big.Int)
	for _, m := range p.monomials {
		y = new(big.Int).Add(y, m.solve(x))
	}
	return
}

// solve will return the result of solving for x in
// a single monomial which is part of a larger polynomial.
func (m monomial) solve(x *big.Int) (y *big.Int) {
	// return value.x^degree
	y = new(big.Int).Mul(m.Value, new(big.Int).Exp(x, m.Degree, nil))
	return
}

// DivideSecret creates a k-threshold secret and returns
// a slice of m shares. If m < k, then k shares will be
// created. The prime modulus used to create the shares is
// also returned, as this is required in order to interpolate
// the shares into the correct polynomial.
func DivideSecret(secret *big.Int, k, m int) (shares []Share, prime *big.Int, err error) {

	// generate a prime P > s
	prime, _ = rand.Prime(rand.Reader, 1+secret.BitLen())

	// We need k total parts for reconstruction, so
	// we will use s as the first monomial, and k-1 extra parts.
	poly := polynomial{
		monomials: []monomial{{secret, new(big.Int)}},
	}

	for i := int64(1); i < int64(k); i++ {
		// select ai where a < P, s < P
		value, err := rand.Int(rand.Reader, prime)
		if err != nil {
			return nil, nil, err
		}
		poly.monomials = append(poly.monomials, monomial{value, big.NewInt(i)})
	}

	// Using the polynomial, construct n shares
	// of the secret. Constructing a share involves
	// soling the polynimial for x to get the point (x, y)
	for i := int64(1); i <= int64(m); i++ {
		yVal := poly.solve(big.NewInt(i))
		shares = append(shares, Share{big.NewInt(i), new(big.Int).Mod(yVal, prime)})
	}

	// Return the set of shares.
	return
}
