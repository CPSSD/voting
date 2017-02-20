package utils

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
)

// only simple function for testing purposes
func EncryptFile(filepath string) (err error) {

	// Open the item to be encrypted (the plaintext)
	_, err = ioutil.ReadFile(filepath)

	return err
}

type PrivateKey struct {
	Lambda *big.Int
	Mu     *big.Int
	PublicKey
}

type PublicKey struct {
	N         *big.Int
	NSquared  *big.Int
	Generator *big.Int
}

var one = big.NewInt(1)

func GenerateKeyPair(bits int) (privateKey *PrivateKey, err error) {

	p, err := rand.Prime(rand.Reader, bits/2)
	Check(err)

	q, err := rand.Prime(rand.Reader, bits/2)
	Check(err)

    // p = big.NewInt(5)
    // q = big.NewInt(7)


	n := new(big.Int).Mul(p, q)
	phiN := getPhi(p, q)
    fmt.Println(p,q,phiN)

	gcd := new(big.Int).GCD(nil, nil, phiN, n)

	for gcd.Cmp(one) != 0 {

		p, err = rand.Prime(rand.Reader, bits)
		Check(err)

		q, err = rand.Prime(rand.Reader, bits)
		Check(err)

		n := new(big.Int).Mul(p, q)
		phiN = getPhi(p, q)

		gcd = new(big.Int).GCD(nil, nil, phiN, n)
	}

	lambda := altLCM(p, q)

	nSquared := new(big.Int).Mul(n, n)

	// generator, err := rand.Int(rand.Reader, nSquared)
	// Check(err)

    generator := new(big.Int).Add(n, one)

    mu := altMu(lambda, n)

	//mu := getMu(generator, lambda, n, nSquared)

	// checking relatieve primality of n and g
	// gcd.GCD(nil, nil, generate(g, l, n), n)
	//
	// for gcd.Cmp(one) != 0 {
	//
	// 	g, err = rand.Int(rand.Reader, nSquared)
	// 	Check(err)
	// 	gcd.GCD(nil, nil, generate(g, l, n), n)
	// 	fmt.Println("bad generator")
	// }

    privateKey = &PrivateKey{
        PublicKey: PublicKey{
            N: n,
            NSquared: nSquared,
            Generator: generator,
        },
        Lambda: lambda,
        Mu: mu,
    }

	return
}

func getMu(g, l, n, n2 *big.Int) (ans *big.Int) {

	x := new(big.Int).Exp(g, l, n2)
	ans = new(big.Int).ModInverse(getL(x, n), n)
	return ans
}

func getL(x, n *big.Int) (ans *big.Int) {

	ans = new(big.Int).Div(new(big.Int).Sub(x, one), n)
	return ans
}

func generate(g, l, n *big.Int) (ans *big.Int) {

	//unused
	mod := new(big.Int)
	mod.Mul(n, n)
	mod.Sub(mod, one)
	ans = new(big.Int)
	ans.Exp(g, l, mod)
	ans.Div(ans, n)
	return ans
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func getPhi(a, b *big.Int) (phi *big.Int) {

	x := new(big.Int).Sub(a, one)
	y := new(big.Int).Sub(b, one)
	phi = new(big.Int).Mul(x, y)
	return phi
}

func altLCM(a, b *big.Int) (lambda *big.Int) {

    x := new(big.Int).Sub(a, one)
    y := new(big.Int).Sub(b, one)

    lambda = new(big.Int).Mul(x, y)
    return lambda
}

func altMu(phi, n *big.Int) (ans *big.Int) {

	ans = new(big.Int).ModInverse(phi, n)
	return ans
}

func carmichaelLCM(a, b *big.Int) (lambda *big.Int) {

	x := new(big.Int).Sub(a, one)
	y := new(big.Int).Sub(b, one)

	top := new(big.Int).Mul(x, y)
	gcd := new(big.Int).GCD(nil, nil, x, y)

	lambda = new(big.Int).Div(top, gcd)

	return lambda
}

func Encrypt(m *big.Int, key *PublicKey) (c *big.Int) {

	g := key.Generator
	n := key.N
    nSquared := key.NSquared

    r, err := rand.Int(rand.Reader, n)
    Check(err)

    c = new(big.Int).Mod(
            new(big.Int).Mul(
                new(big.Int).Exp(g, m, nSquared),
                new(big.Int).Exp(r, n, nSquared)), nSquared)


	// c = new(big.Int)
	// cA := new(big.Int)
	// cB := new(big.Int)
    //
    //
    //
	// cA.Exp(g, m, nSquared)
	// cB.Exp(r, n, nSquared)
    //
	// c.Mul(cA, cB)
	// c.Mod(c, nSquared)

	return c
}

func Decrypt(c *big.Int, key *PrivateKey) (m *big.Int) {

	lambda := key.Lambda
	mu := key.Mu
	n := key.PublicKey.N
    nSquared := key.PublicKey.NSquared

	m = new(big.Int)

	m.Exp(c, lambda, nSquared)
	m = getL(m, n)
	m.Mul(m, mu)
	m.Mod(m, n)

	return m
}
