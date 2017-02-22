package crypto

import (
    "math/big"
    "strings"
    "crypto/rand"
)

type Polynomial struct {
    Monomials []Monomial
}

func (p Polynomial) String() (str string) {
    for _, m := range p.Monomials {
        str = str + " " + m.String() + " +"
    }
    str = strings.TrimRight(str, "+")
    return "f(x) = "+ strings.TrimSpace(str)
}

type Monomial struct {
    Value *big.Int
    Degree *big.Int
}

// Secret represents a point (X, Y) in a polynomial
type Secret struct {
    X *big.Int
    Y *big.Int
}

func (m Monomial) String() string {
    return m.Value.String() + "x^" + m.Degree.String()
}

func (p Polynomial) solve(x *big.Int) (y *big.Int) {
    y = new(big.Int)
    for _, m := range p.Monomials {
        y = new(big.Int).Add(y, m.solve(x))
    }
    return
}

func (m Monomial) solve(x *big.Int) (y *big.Int) {
    // return value.x^degree
    y = new(big.Int).Mul(m.Value, new(big.Int).Exp(x, m.Degree, nil))
    return
}

//   dividesecret( S    ,    k   ,       n        )
func DivideSecret(secret *big.Int, threshold, shares int) (secrets []Secret, err error) {

    // we need k total parts for reconstruction, so
    // we will use s, and k-1 extra parts.
    // Construct the polynomial of k-1 degrees here

    // generate a prime P > s
    prime, _ := rand.Prime(rand.Reader, 60 + secret.BitLen())

    poly := Polynomial{
        Monomials: []Monomial{{secret, new(big.Int)}},
    }

    // select ai where a < P, s < P
    for i := int64(1); i <= int64(threshold); i++ {
        value, err := rand.Int(rand.Reader, prime)
        if err != nil {
            return nil, err
        }
        poly.Monomials = append(poly.Monomials, Monomial{value, big.NewInt(i)})

    }

    for i := int64(1); i <= int64(shares); i++ {
        secrets = append(secrets, Secret{big.NewInt(i), poly.solve(big.NewInt(i))})
    }

    // Using the polynomial, construct n shares
    // of the secret. Constructing a share involves
    // soling the polynimial for x to get the point (x, y)

    // Return the set of shares.
    return
}
