package crypto_test

import (
	"github.com/CPSSD/voting/src/crypto"
	"testing"
    "errors"
    "strconv"
)

func TestInterpolation(t *testing.T) {

    var tests = []struct {
        threshold int
        shares int
        collaborators int
        interpolates bool
    }{
        {2, 2, 2, true},
        {2, 3, 2, true},
        {2, 3, 3, true},
        {4, 3, 3, false},
        {4, 10, 3, false},
        {5, 10, 4, false},
        {5, 10, 5, true},
    }

    for _, c := range tests {
    	success, err := CheckInterpolation(c.threshold, c.shares, c.collaborators)
        if err != nil {
            t.Error(err)
        }
    	if success != c.interpolates {
    		t.Error("For input",c.collaborators,"collaborators with pool of ",
                c.shares,"shares and threshold of",c.threshold,"expected ",c.interpolates,"got",success,"\n")
    	}
    }
}

func CheckInterpolation(threshold, shares, collaborators int) (success bool, err error) {

    priv, err := crypto.GenerateKeyPair(10)
	if err != nil {
		return false, err
	}

    secrets, err := crypto.DivideSecret(priv.Lambda, threshold, shares)
    if err != nil {
        return false, err
    }

    if collaborators > shares {
        return false, errors.New("Cannot get slice of "+
            strconv.Itoa(collaborators)+
            " collaborators from "+
            strconv.Itoa(shares)+
            " shares")
    }

    secret, err := crypto.Interpolate(secrets)
    if err != nil {
        return false, err
    }

    if secret.Cmp(priv.Lambda) != 0 && collaborators >= threshold {
        return false, err
    }
    if secret.Cmp(priv.Lambda) == 0 && collaborators <= threshold {
        return false, err
    }
    return true, err
}
