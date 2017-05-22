// Generate is used to create a set of user configs for the
// system, along with the public and private key of the
// election. The private key is shared amongst the nodes.
// This program will also create DSA keys for users. The
// completed configuration files will be named json files
// in the range 0 to N-1, where N is the number of users to
// generate.
package main

import (
	"crypto/dsa"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/CPSSD/voting/src/blockchain"
	"github.com/CPSSD/voting/src/crypto"
	"github.com/CPSSD/voting/src/election"
	"io/ioutil"
	"math/big"
	mrand "math/rand"
	"strconv"
)


func main() {
	var numVoters int      // how many "registered" voterList are there
	var shareThreshold int // amount of collaborators to recreate election key
	var allowPeerSync bool // are peers allowed to discover new peers
	var portNumber int     // first port to use for the network
	var tokenLen int       // number of characters in a vote token
	var degree int         // minimum number of known peers per node
	var input string

	fmt.Printf("Number of voters to generate: ")
	fmt.Scanf("%v\n", &numVoters)

	fmt.Printf("Threshold number of shares to construct election key: ")
	fmt.Scanf("%v\n", &shareThreshold)

	fmt.Printf("Allow peers to sync? (y/n): ")
	fmt.Scanf("%v\n", &input)
	switch string(input[0]) {
	case "y":
		allowPeerSync = true
	case "n":
		allowPeerSync = false
	default:
		fmt.Println("Allowing peers to sync by default")
		allowPeerSync = true
	}
	fmt.Printf("Minimum known starting peers? (min recommended = 1): ")
	fmt.Scanf("%v\n", &degree)

	fmt.Printf("Initial port number for nodes: ")
	fmt.Scanf("%v\n", &portNumber)

	fmt.Printf("Number of characters in a vote node: ")
	fmt.Scanf("%v\n", &tokenLen)

	format := election.CreateFormat()

	fmt.Println("Building election config...")
	// create the election key
	priv, err := crypto.GenerateKeyPair(512)
	if err != nil {
		panic(err)
	}

	// create the shares of the election key lambda value
	lambdaShares, lambdaPrimeModulus, err := crypto.DivideSecret(priv.Lambda, shareThreshold, numVoters)
	if err != nil {
		panic(err)
	}

	// create the shares of the election key's mu value
	muShares, muPrimeModulus, err := crypto.DivideSecret(priv.Mu, shareThreshold, numVoters)
	if err != nil {
		panic(err)
	}

	voteTokens := make(map[string]dsa.PublicKey, numVoters)

	voterList := make([]blockchain.Configuration, numVoters)
	var i int

	for ; numVoters > 0; numVoters-- {
		var conf blockchain.Configuration
		privateKey := createKey()
		vt := createVoteToken(tokenLen)
		_, exists := voteTokens[vt]
		for exists {
			vt = createVoteToken(tokenLen)
			_, exists = voteTokens[vt]
		}
		voteTokens[vt] = privateKey.PublicKey

		conf = blockchain.Configuration{
			MyAddr:     "localhost",
			MyPort:     ":" + strconv.Itoa(portNumber+i),
			Peers:      make(map[string]bool, 0),
			SyncPeers:  allowPeerSync,
			PrivateKey: *privateKey,
			MyToken:    vt,

			ElectionFormat: *format,

			ElectionKey: crypto.PrivateKey{
				Lambda:    new(big.Int),
				Mu:        new(big.Int),
				PublicKey: priv.PublicKey,
			},

			ElectionKeyShare: blockchain.ElectionSecret{
				Lambda: lambdaShares[i],
				Mu:     muShares[i],
			},
			ElectionLambdaModulus: lambdaPrimeModulus,
			ElectionMuModulus:     muPrimeModulus,
		}

		voterList[i] = conf
		i++
	}

	for i, _ := range voterList {
		voterList[i].VoteTokens = voteTokens
	}

	voterList = generateUndirectedGraph(voterList, degree)

	for i, v := range voterList {
		bytes, err := json.MarshalIndent(v, "", "    ")
		err = ioutil.WriteFile(strconv.Itoa(i)+".peer.json", bytes, 0777)
		if err != nil {
			fmt.Println("Could not save configuration to json file")
			panic(err)
		}
	}

	fmt.Println("Done")

}

func generateUndirectedGraph(voterList []blockchain.Configuration, degree int) (out []blockchain.Configuration) {
	tmp := voterList
	unconn := make(map[string]bool)
	for _, conf := range tmp {
		unconn[conf.MyAddr+conf.MyPort] = true
	}
	conn := make(map[string]bool)
	for i, conf := range tmp {
		delete(unconn, conf.MyAddr+conf.MyPort)
		for p, _ := range unconn {
			if len(voterList[i].Peers) < degree {
				voterList[i].Peers[p] = true
				conn[p] = true
				delete(unconn, p)
			} else {
				break
			}
		}
		for p, _ := range conn {
			if len(voterList[i].Peers) < degree {
				voterList[i].Peers[p] = true
			} else {
				break
			}
		}
		conn[conf.MyAddr+conf.MyPort] = true
	}
	return voterList
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func createVoteToken(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[mrand.Intn(len(letterRunes))]
	}
	return string(b)
}

func createKey() (privateKey *dsa.PrivateKey) {
	params := new(dsa.Parameters)

	err := dsa.GenerateParameters(params, crand.Reader, dsa.L2048N256)
	if err != nil {
		fmt.Println("Could not generate DSA parameters")
		panic(err)
	}

	privateKey = new(dsa.PrivateKey)
	privateKey.PublicKey.Parameters = *params
	dsa.GenerateKey(privateKey, crand.Reader)

	return privateKey
}
