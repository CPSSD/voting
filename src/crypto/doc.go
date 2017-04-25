/*
   Package crypto implements the Paillier cryptography system
   and provides functionality to allow homomorphic addition
   of ciphertexts.

   Key Structure

   Private key:
       type PrivateKey struct {
       	Lambda *big.Int
       	Mu     *big.Int
       	PublicKey
       }

   The private key consists of the secret values Lambda and Mu. It also
   contains a reference to the corresponding public key:

       type PublicKey struct {
       	N         *big.Int
       	NSquared  *big.Int
       	Generator *big.Int
       }

   The public key consists of the public values N and the Generator. There
   is also a reference to the value of N^2 as this value is used frequently
   in computations.

   The values for the keys are stored as pointers to a big.Int for ease of
   computation across cryptographic functions.

   Key Generation

   Key-pairs are generated as follows:

       var bits int = 1024
       keyPair, err := crypto.GenerateKeyPair(bits)

   The value of err should be checked to ensure that keyPair is valid. To
   validate a key-pair you could also run:

       err := keyPair.Validate()

   and check the value of err. This is actually run internally when a
   key-pair is created, but it is nice to be able to validate a key-pair which
   you have not generated yourself. The Validate() function will validate
   either a private or public key.

   The validation process essentially checks that no null values have entered
   the key.

   Encryption

   Encryption is performed as follows:

       var plaintext *big.Int = big.NewInt(23)
       ciphertext, err := key.Encrypt(plaintext)

   where key either a PrivateKey or PublicKey. The value of err should be
   checked to ensure that the encryption was successfully performed.

   Decryption

   Decryption is performed as follows:

       decipheredText, err := key.Decrypt(ciphertext)

   where key must be a PrivateKey. Again, the value of err should be checked
   for errors with the ciphertext or key.

   Homomorphic addition

   Homomorphic addition of ciphertexts can be performed as follows:

       ciphertext_a, _ := key.Encrypt(plaintext_a)
       ciphertext_b, _ := key.Encrypt(plaintext_b)
       ...
       ciphertext_sum, err := key.AddCipherTexts(ciphertext_a, ciphertext_b...)

   The value of ciphertext_sum, when decrypted, will be equal to the value
   of plaintext_a + plaintext_b mod N. To add ciphertexts, the key used
   can be either the public or private key.

   Secret sharing

   Secret sharing of a *big.Int can be performed as follows:

       secret := big.NewInt(1234)
       threshold := 18
       numShares := 20
       shares, prime, err := crypto.DivideSecret(secret, threshold, numShares)

   The contents of the slice shares should be distributed amongst users of the
   system. The ratio of threshold:numShares will determine the security vs
   redundancy of the system. The value of prime should be made public to allow
   the reconstruction of the polynomial.

   Interpolation

   Shares can be interpolated to reconstruct a secret as follows. Given:

       var collaboratorShares []Share

   containing a set of shares greater than the required threshold,
   and the prime modulus from the secret sharing step:

       secret, err := crypto.Interpolate(collaboratorShares, prime)

   If the amount of shares used is not at least equal to the threshold,
   then the value of secret will not be correct.
*/
package crypto
