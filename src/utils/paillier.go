package utils

import (
    "io/ioutil"
    "fmt"
    "math/big"
    "crypto/rand"
)

// only simple function for testing purposes
func EncryptFile(filepath string) (err error) {

	// Open the item to be encrypted (the plaintext)
	_, err = ioutil.ReadFile(filepath)

    return err
}

type PaillierPrivateKey struct {

    p, q, g big.Int
}

type PaillierPublicKey struct {

    lambda, mu big.Int
}

func (key *PaillierPrivateKey) set(p, q, g, n *big.Int) {
    key.p = *p
    key.q = *q
    key.g = *g
    key.n = *n
}
func (key *PaillierPublicKey) set(l, m, n *big.Int) {
    key.lambda = *l
    key.mu = *m
    key.n = *n
}

var one = big.NewInt(1)

func GenerateKeyPair(bits int) (publicKey PaillierPublicKey, privateKey PaillierPrivateKey, err error) {

    privateKey = PaillierPrivateKey{
        p: *new(big.Int),
        q: *new(big.Int),
        g: *new(big.Int),
        n: *new(big.Int),
    }

    publicKey = PaillierPublicKey{
        lambda: *new(big.Int),
        mu: *new(big.Int),
        n: *new(big.Int),
    }

    gcd := new(big.Int)
    n := new(big.Int)
    phiN := new(big.Int)

    p := new(big.Int)
    q := new(big.Int)
    g := new(big.Int)

    l := new(big.Int)
    m := new(big.Int)

    fmt.Println("gcd of",phiN,"and",n,"is",gcd)

    for gcd.Cmp(one) != 0 {

        p, err = rand.Prime(rand.Reader, bits)
        check(err)

        q, err = rand.Prime(rand.Reader, bits)
        check(err)

        n.Mul(p,q)
        phiN = getPhi(p, q)

        gcd.GCD(nil, nil, phiN, n)

        l = carmichael(p, q)
        fmt.Println("gcd of",phiN,"and",n,"is",gcd,"\nand carmichael is",l)
    }

    nSquared := new(big.Int)
    nSquared.Mul(n,n)

    g, err = rand.Int(rand.Reader, nSquared)
    check(err)
    gcd.GCD(nil, nil, generate(g, l, n), n)

    for gcd.Cmp(one) != 0 {

        g, err = rand.Int(rand.Reader, nSquared)
        check(err)
        gcd.GCD(nil, nil, generate(g, l, n), n)
        fmt.Println("bad generator")
    }

    m = getMu(g, l, n)

    privateKey.set(p,q,g,n)
    publicKey.set(l,m,n)

    fmt.Println("privateKey:\n\tp:",privateKey.p,"\n\tq:",privateKey.q,"\npublicKey:\n\tlambda:",publicKey.lambda,"\n\tmu:",publicKey.mu)
    return
}

func getMu(g, l, n *big.Int) (ans *big.Int) {
    mod := new(big.Int)
    mod.Mul(n,n)
    u := new(big.Int)
    u.Exp(g,l,mod)
    res := getL(u, n)
    ans = new(big.Int)
    ans.ModInverse(res, n)
    return ans
}

func getL(u, n *big.Int) (ans *big.Int) {
    ans = new(big.Int)
    ans.Sub(u, one)
    ans.Div(ans, n)
    return ans
}

func generate(g, l, n *big.Int) (ans *big.Int) {
    mod := new(big.Int)
    mod.Mul(n,n)
    mod.Sub(mod,one)
    ans = new(big.Int)
    ans.Exp(g,l,mod)
    ans.Div(ans,n)
    return ans
}

func check(err error) {
    if err != nil {
        panic(err)
    }
}

func getPhi(a, b *big.Int) (phi *big.Int) {

    x := a.Sub(a, one)
    y := b.Sub(b, one)
    phi = x.Mul(x, y)
    return phi
}

func carmichael(a, b *big.Int) (lambda *big.Int) {

    x := a.Sub(a, one)
    y := b.Sub(b, one)
    phi := getPhi(a, b)

    gcd := new(big.Int)
    gcd.GCD(nil, nil, x, y)

    lambda = new(big.Int)
    lambda.Div(phi, gcd)

    return lambda
}
