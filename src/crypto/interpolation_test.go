package crypto_test

import (
	"github.com/CPSSD/voting/src/crypto"
	"math"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"errors"
	"fmt"
	"strconv"
)

var debug = false

func TestInterpolation(t *testing.T) {

	var tests = []struct {
		threshold     int
		shares        int
		collaborators int
		interpolates  bool
	}{
		{0, 0, 0, false},
		{1, 1, 1, true},
		{2, 2, 2, true},
		{2, 10, 10, true},
		{2, 3, 3, true},
		{4, 3, 3, false},
		{4, 10, 3, false},
		{4, 10, 4, true},
		{4, 10, 5, true},
		{5, 10, 4, false},
		{5, 10, 5, true},
		{200, 10000, 199, false},
	}

	for i, c := range tests {
		success, err := CheckInterpolation(c.threshold, c.shares, c.collaborators)
		if success != c.interpolates {
			debug = true
			success, err = CheckInterpolation(c.threshold, c.shares, c.collaborators)
			debug = false
			t.Error("Test no:", i, "For input", c.collaborators, "collaborators with pool of ",
				c.shares, "shares and threshold of", c.threshold, "expected ", c.interpolates, "got", success, "\n")
			if err != nil {
				t.Error(err)
			}
		}
	}

	success, err := CheckAltInterpolation()
	if success != true {
		t.Error("Did not interpolate simple example")
	}
	if err != nil {
		t.Error(err)
	}
}

func CheckAltInterpolation() (success bool, err error) {
	secret := big.NewInt(1234)
	mod := big.NewInt(1613)

	var points = []crypto.Share{
		crypto.Share{X: big.NewInt(1), Y: big.NewInt(1494)},
		crypto.Share{X: big.NewInt(2), Y: big.NewInt(329)},
		crypto.Share{X: big.NewInt(3), Y: big.NewInt(965)},
		crypto.Share{X: big.NewInt(4), Y: big.NewInt(176)},
		crypto.Share{X: big.NewInt(5), Y: big.NewInt(1188)},
		crypto.Share{X: big.NewInt(6), Y: big.NewInt(775)},
	}

	recovered, err := crypto.Interpolate(points, mod)

	if debug {
		fmt.Println("The secret is", secret)
		fmt.Println("The recovered val is", recovered)
	}

	if secret.Cmp(recovered) != 0 {
		return false, err
	}

	return true, err

}

func CheckInterpolation(threshold, shares, collaborators int) (success bool, err error) {

	priv, err := crypto.GenerateKeyPair(10)
	if err != nil {
		return false, err
	}

	secrets, prime, err := crypto.DivideSecret(priv.Lambda, threshold, shares)
	if err != nil {
		return false, err
	}

	if collaborators > shares {
		return false, errors.New("Cannot get slice of " +
			strconv.Itoa(collaborators) +
			" collaborators from " +
			strconv.Itoa(shares) +
			" shares")
	}

	if debug {
		fmt.Println("Pre Secrets =", secrets)
	}

	collabSecrets := randomSlice(collaborators, secrets)

	if debug {
		fmt.Println("Collab Secrets @", collaborators, "collabs =", collabSecrets)
	}

	secret, err := crypto.Interpolate(collabSecrets, prime)
	if err != nil {
		return false, err
	}

	if debug {
		fmt.Println("secret is", priv.Lambda)
		fmt.Println("recove is", secret)
	}

	// if interpolated correctly
	if secret.Cmp(priv.Lambda) == 0 {
		if collaborators < threshold {
			return true, errors.New("Unexpected successful interpolation of key.")
		} else {
			return true, nil
		}
	} else {
		if collaborators >= threshold {
			return false, errors.New("Collaborators should have interpolated the secret.")
		} else {
			return false, nil
		}
	}
}

func randomSlice(num int, in []crypto.Share) (out []crypto.Share) {

	slice := append([]crypto.Share(nil), in...)

	rand.Seed(time.Now().UnixNano())
	num = int(math.Min(float64(num), float64(len(slice))))

	for ; num > 0; num-- {
		selection := rand.Intn(len(slice))
		out = append(out, slice[selection])
		slice = append(slice[:selection], slice[selection+1:]...)

	}
	return out
}
