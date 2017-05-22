package main

import (
	"fmt"
	"github.com/CPSSD/voting/src/blockchain"
	"github.com/CPSSD/voting/src/election"
	"log"
	"os"
	"sync"
)

var (
	tokenMsg    string = "Please enter your unique token"
	voteMsg     string = "Please enter your ballot message"
	badInputMsg string = "Unrecognised input"
	waitMsg     string = "Waiting for processes to quit"
)

func main() {

	f, err := os.OpenFile(os.Args[1]+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.SetFlags(log.Ltime | log.Lmicroseconds | log.Lshortfile)

	c, err := blockchain.NewChain()
	if err != nil {
		panic(err)
	}

	log.Println("Setting up network config")
	filename := string(os.Args[1])

	c.Init(filename)

	// to quit entirely
	quit := make(chan bool, 1)

	// to signal to stop mining
	stop := make(chan bool, 1)

	// to signal to start mining
	start := make(chan bool, 1)

	// to signal to confirm stopped mining
	confirm := make(chan bool, 1)

	var syncDelay int = 10
	var wg sync.WaitGroup
	wg.Add(4)
	c.Start(syncDelay, quit, stop, start, confirm, &wg)
	start <- true

	fmt.Println("Welcome to voting system.")
	vt := c.GetVoteToken()
	fmt.Println("Your vote token is:", vt)

loop:
	for {
		fmt.Printf("What next? (h for help): ")
		var input string
		fmt.Scanf("%v\n", &input)

		switch input {
		case "h":
			fmt.Printf("\th\t\tPrint this help\n")
			fmt.Printf("\tpeers\t\tPrint known peers\n")
			fmt.Printf("\tpool\t\tPrint pool of transactions\n")
			fmt.Printf("\tchain\t\tPrint current chain\n")
			fmt.Printf("\tv\t\tCast a vote\n")
			fmt.Printf("\tq\t\tQuit program\n")
			fmt.Printf("\tb\t\tBroadcast share\n")
			fmt.Printf("\tr\t\tReconstruct election key\n")
			fmt.Printf("\ttally\t\tTally the votes\n")
		case "peers":
			c.PrintPeers()
		case "pool":
			c.PrintPool()
		case "chain":
			fmt.Println("Entering print chain")
			fmt.Println(c)
			fmt.Println("Exited print chain")
		case "q":
			quit <- true
			break loop
		case "b":
			fmt.Printf("Broadcasting our share of the election key\n")
			c.BroadcastShare()
		case "r":
			fmt.Printf("Attempting to reconstruct the election key\n")
			c.ReconstructElectionKey()
			c.PrintKey()
		case "v":

			token := vt

			ballot := new(election.Ballot)
			err := ballot.Fill(c.GetFormat(), tokenMsg)
			if err != nil {
				log.Printf("Error filling out the ballot")
			} else {
				tr := c.NewTransaction(token, ballot)
				go c.ReceiveTransaction(tr, nil)
			}
		case "tally":
			ballots := c.CollectBallots()
			format := c.GetFormat()
			key := c.GetElectionKey()
			fmt.Println("Calculating the tally...")
			tally, err := format.Tally(ballots, &key)
			if err != nil {
				fmt.Println("Error calculating tally")
			}
			fmt.Println(tally)
		default:
			fmt.Println(badInputMsg)
		}

	}

	fmt.Printf("%v\n", waitMsg)
	log.Printf("%v\n", waitMsg)
	wg.Wait()
	log.Println(c)
}
