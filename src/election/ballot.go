package election

import (
	"errors"
	"fmt"
	"github.com/CPSSD/voting/src/crypto"
	"math/big"
	"strings"
)

var (
	InvalidFormatError = errors.New("Invalid format was supplied; bad number of selections.")
)

type Ballot struct {
	VoteToken     string      // VT of the voter who owns ballot
	NumSelections int         // number of selections in the ballot
	Selections    []Selection // list of selections on the ballot
}

type Selection struct {
	Name  string   // name of this selection option
	Vote  *big.Int // value should be encrypted with PrivKE
	Proof []byte   // value used as the zero-knowledge proof
}

type Format struct {
	NumSelections int
	Selections    []Selection
}

func (b *Ballot) Fill(f Format, vt string) (err error) {
	if len(f.Selections) != f.NumSelections {
		return InvalidFormatError
	}

	b.VoteToken = vt
	b.NumSelections = f.NumSelections
	b.Selections = make([]Selection, f.NumSelections)

	for i, s := range f.Selections {
		fmt.Printf("Enter your selection (0 or 1) for Candidate %v: ", s.Name)
		var input int
		fmt.Scanf("%v\n", &input)

		vote := big.NewInt(int64(input))

		selection := Selection{
			Name:  s.Name,
			Vote:  vote,
			Proof: make([]byte, 0),
		}

		b.Selections[i] = selection
	}

	// TODO: let user review inputs before returning

	return nil
}

func CreateFormat() (f *Format) {
	fmt.Printf("How many selections are on the ballot? ")
	var input int
	fmt.Scanf("%v\n", &input)

	f = &Format{
		NumSelections: input,
		Selections:    make([]Selection, input),
	}

	fmt.Println("Use double quotes for description entries")
	for i := 0; i < input; i++ {
		fmt.Printf("Enter user description for selection %v: ", i+1)
		var desc string

		fmt.Scanf("%q\n", &desc)
		fmt.Println("You entered:", desc)

		desc = strings.Trim(desc, " \n")

		s := Selection{
			Name: desc,
		}

		f.Selections[i] = s
	}

	return f
}

type Tally struct {
	Totals map[string]*big.Int
}

func (t Tally) String() (str string) {
	for name, result := range t.Totals {
		str = str + name + ": " + result.String() + " votes\n"
	}
	return "Totals for the election are as follows:\n" + str
}

func (f *Format) Tally(bs *[]Ballot, key *crypto.PrivateKey) (t *Tally, err error) {

	t = &Tally{
		Totals: make(map[string]*big.Int, 0),
	}

	selectionCounts := make(map[string][]*big.Int, 0)

	for _, s := range f.Selections {
		selectionCounts[s.Name] = make([]*big.Int, len(*bs))
	}

	for i, b := range *bs {
		for _, s := range b.Selections {
			if _, ok := selectionCounts[s.Name]; ok {
				selectionCounts[s.Name][i] = s.Vote
			}
		}
	}

	// TODO: decrypt each sub tally
	for name, count := range selectionCounts {
		sum, err := key.AddCipherTexts(count...)
		if err != nil {
			return t, err
		}
		result, err := key.Decrypt(sum)
		if err != nil {
			return t, err
		}
		t.Totals[name] = result
	}

	return t, err
}
