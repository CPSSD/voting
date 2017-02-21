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
*/
package crypto
