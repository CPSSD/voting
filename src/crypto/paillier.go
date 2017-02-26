package crypto

import (
	"crypto/rand"
	"errors"
	"math/big"
)

var (
	InvalidPlaintextError  = errors.New("Invalid plaintext was submitted for encryption.")
	InvalidCiphertextError = errors.New("Invalid ciphertext was submitted for decryption.")
	InvalidPublicKeyError  = errors.New("Invalid public key.")
	InvalidPrivateKeyError = errors.New("Invalid private key.")
)

var one = big.NewInt(1)

// PrivateKey contains the private components Lambda and Mu,
// and the public components in PublicKey.
type PrivateKey struct {
	Lambda *big.Int
	Mu     *big.Int
	PublicKey
}

// PublicKey contains the public components N, NSquared and
// Generator.
type PublicKey struct {
	N         *big.Int
	NSquared  *big.Int
	Generator *big.Int
}

// Encrypt returns the ciphertext c which is created by
// encrypting the message m with the PublicKey associated
// with PrivateKey key. The encryption of a given message
// m is non-deterministic.
func (key *PrivateKey) Encrypt(m *big.Int) (c *big.Int, err error) {
	c, err = key.PublicKey.Encrypt(m)
	return
}

// Encrypt returns the ciphertext c which is created by
// encrypting the message m with the PrivateKey key. The
// encryption of a given message m is non-deterministic.
func (key *PublicKey) Encrypt(m *big.Int) (c *big.Int, err error) {

	if m == nil {
		return nil, InvalidPlaintextError
	}
	if err = key.Validate(); err != nil {
		return nil, err
	}

	r, err := rand.Int(rand.Reader, key.N)
	if err != nil {
		return nil, err
	}

	// c = ((g^m).(r^n)) mod (n^2)
	c = new(big.Int).Mod(
		new(big.Int).Mul(
			new(big.Int).Exp(key.Generator, m, key.NSquared),
			new(big.Int).Exp(r, key.N, key.NSquared)), key.NSquared)

	return c, err
}

// Decrypt returns the message m which is obtained from
// decrypting the ciphertext c using PrivateKey key. If
// a nil ciphertext is passed to this function, an
// InvalidCiphertextError will be returned along with a
// nil value for m.
// If an invalid key is used, a corresponding error will
// be returned with a nil value for m.
func (key *PrivateKey) Decrypt(c *big.Int) (m *big.Int, err error) {

	if c == nil {
		return nil, InvalidCiphertextError
	}
	if err = key.Validate(); err != nil {
		return nil, err
	}

	// m = L(c^lambda mod n^2).mu mod n
	// where L(x) = (x-1)/n
	m = new(big.Int).Exp(c, key.Lambda, key.PublicKey.NSquared)
	m = getL(m, key.PublicKey.N)
	m.Mul(m, key.Mu)
	m.Mod(m, key.PublicKey.N)

	return m, err
}

// GenerateKeyPair returns a PrivateKey struct containing the
// private components of a key-pair and the corresponding
// PublicKey struct. The value bits determines the size of
// prime numbers to be used in the generation of the key-pair.
func GenerateKeyPair(bits int) (privateKey *PrivateKey, err error) {

	n, lambda, err := generatePrimePair(bits)
	if err != nil {
		return nil, err
	}

	mu := getMu(lambda, n)
	generator := new(big.Int).Add(n, one)

	nSquared := new(big.Int).Mul(n, n)

	privateKey = &PrivateKey{
		PublicKey: PublicKey{
			N:         n,
			NSquared:  nSquared,
			Generator: generator,
		},
		Lambda: lambda,
		Mu:     mu,
	}

	err = privateKey.Validate()

	return
}

// Validate returns an InvalidPrivateKeyError if the
// key is nil, or if the values of Mu or Lambda are nil.
// If the PublicKey is invalid an InvalidPublicKeyError
// will be returned, else nil is returned.
func (key *PrivateKey) Validate() (err error) {

	if key == nil || key.Mu == nil || key.Lambda == nil {
		return InvalidPrivateKeyError
	}
	if err = key.PublicKey.Validate(); err != nil {
		return err
	}
	return
}


// Validate returns an InvalidPublicKeyError if the key
// is nil, or if the values of N, NSquared or Generator
// are nil, else nil is returned.
func (key *PublicKey) Validate() (err error) {
	if key == nil || key.N == nil ||
		key.NSquared == nil || key.Generator == nil {

		return InvalidPublicKeyError
	}
	return
}

// generatePrimePair returns n and phiN, where n = p.q,
// and phiN = (p-1).(q-1), and p and q are primes
// with a length specified by the bits argument.
func generatePrimePair(bits int) (n, phiN *big.Int, err error) {

	gcd := new(big.Int)

	for gcd.Cmp(one) != 0 {

		p, err := rand.Prime(rand.Reader, bits)
		if err != nil {
			return nil, nil, err
		}

		q, err := rand.Prime(rand.Reader, bits)
		if err != nil {
			return nil, nil, err
		}

		n = new(big.Int).Mul(p, q)
		phiN = getPhi(p, q)

		gcd = new(big.Int).GCD(nil, nil, phiN, n)
	}

	return
}

// getMu returns the modular inverse of phi mod n,
// and is used to generate the value Mu for a
// PrivateKey.
func getMu(phi, n *big.Int) (ans *big.Int) {

	ans = new(big.Int).ModInverse(phi, n)
	return ans
}

// getL returns the value of (x-1)/n. Note that this
// performs a division, and not a multiplication of a
// modular inverse. It is used during the decryption
// of a ciphertext.
func getL(x, n *big.Int) (ans *big.Int) {

	ans = new(big.Int).Div(new(big.Int).Sub(x, one), n)
	return ans
}

// getPhi returns the result of the totient function
// on two primes a and b. It is used to calculate
// phi(n), where n = p.q.
func getPhi(a, b *big.Int) (phi *big.Int) {

	x := new(big.Int).Sub(a, one)
	y := new(big.Int).Sub(b, one)
	phi = new(big.Int).Mul(x, y)
	return phi
}
