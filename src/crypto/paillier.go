package crypto

import (
	"crypto/rand"
	"io/ioutil"
	"math/big"
    "github.com/CPSSD/voting/src/utils"
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

func generatePrimePair(bits int) (n, phiN *big.Int) {

    gcd := new(big.Int)

    for gcd.Cmp(one) != 0 {

		p, err := rand.Prime(rand.Reader, bits)
		utils.Check(err)
		q, err := rand.Prime(rand.Reader, bits)
		utils.Check(err)

		n = new(big.Int).Mul(p, q)
		phiN = getPhi(p, q)

		gcd = new(big.Int).GCD(nil, nil, phiN, n)
	}

    return
}

func GenerateKeyPair(bits int) (privateKey *PrivateKey, err error) {

	n, lambda := generatePrimePair(bits)
    mu := getMu(lambda, n)
    generator := new(big.Int).Add(n, one)

    nSquared := new(big.Int).Mul(n, n)

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

func getMu(phi, n *big.Int) (ans *big.Int) {

	ans = new(big.Int).ModInverse(phi, n)
	return ans
}

func getL(x, n *big.Int) (ans *big.Int) {

	ans = new(big.Int).Div(new(big.Int).Sub(x, one), n)
	return ans
}

func getPhi(a, b *big.Int) (phi *big.Int) {

	x := new(big.Int).Sub(a, one)
	y := new(big.Int).Sub(b, one)
	phi = new(big.Int).Mul(x, y)
	return phi
}

func Encrypt(m *big.Int, key *PublicKey) (c *big.Int) {

    r, err := rand.Int(rand.Reader, key.N)
    utils.Check(err)

    // c = ((g^m).(r^n)) mod (n^2)
    c = new(big.Int).Mod(
            new(big.Int).Mul(
                new(big.Int).Exp(key.Generator, m, key.NSquared),
                new(big.Int).Exp(r, key.N, key.NSquared)), key.NSquared)

	return c
}

func Decrypt(c *big.Int, key *PrivateKey) (m *big.Int) {

    // m = L(c^lambda mod n^2).mu mod n
    // where L(x) = (x-1)/n
	m = new(big.Int).Exp(c, key.Lambda, key.PublicKey.NSquared)
	m = getL(m, key.PublicKey.N)
	m.Mul(m, key.Mu)
	m.Mod(m, key.PublicKey.N)

	return m
}
