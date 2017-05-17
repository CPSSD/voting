package main

import (
	"fmt"
	"github.com/CPSSD/voting/src/blockchain"
	"log"
	"math/big"
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
	wg.Add(3)
	c.Start(syncDelay, quit, stop, start, confirm, &wg)
	start <- true

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
			fmt.Printf("\tsp\t\tSave current peer list\n")
			fmt.Printf("\tv\t\tCast a vote\n")
			fmt.Printf("\tq\t\tQuit program\n")
			fmt.Printf("\tb\t\tBlock interrupt\n")
		case "peers":
			c.PrintPeers()
		case "pool":
			c.PrintPool()
		case "chain":
			fmt.Println("Entering print chain")
			fmt.Println(c)
			fmt.Println("Exited print chain")
		case "sp":
			fmt.Printf("\tEnter file name to save to: ")
			fmt.Scanf("%v\n", &input)
			c.SavePeers(input)
		case "q":
			quit <- true
			break loop
		case "b":
			stop <- true
		case "v":
			var tokenStr string
			var vote int64

			fmt.Printf("%s: ", tokenMsg)
			fmt.Scanf("%v\n", &tokenStr)
			fmt.Printf("%s: ", voteMsg)
			fmt.Scanf("%v\n", &vote)

			token := tokenStr
			ballot := big.NewInt(vote)

			tr := c.NewTransaction(token, ballot)

			go c.ReceiveTransaction(tr, nil)
		default:
			fmt.Println(badInputMsg)
		}

	}

	fmt.Printf("%v\n", waitMsg)
	log.Printf("%v\n", waitMsg)
	wg.Wait()
	log.Println(c)
}
