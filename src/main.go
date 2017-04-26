package main

import (
	"fmt"
	"github.com/CPSSD/voting/src/blockchain"
	"os"
	"sync"
)

var (
	tokenMsg    string = "Please enter your unique token"
	ballotMsg   string = "Please enter your ballot message"
	badInputMsg string = "Unrecognised input"
	waitMsg     string = "Waiting for processes to quit"
)

func main() {

	c, err := blockchain.NewChain()
	if err != nil {
		panic(err)
	}

	fmt.Println("Setting up network config")
	filename := string(os.Args[1])

	c.Init(filename)

	quit := make(chan bool, 1)
	stop := make(chan bool, 1)
	var syncDelay int = 10
	var wg sync.WaitGroup
	wg.Add(2)
	c.Start(syncDelay, quit, stop, &wg)

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
			fmt.Println(c)
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
			var ballotStr string

			fmt.Printf("%s: ", tokenMsg)
			fmt.Scanf("%v\n", &tokenStr)
			fmt.Printf("%s: ", ballotMsg)
			fmt.Scanf("%v\n", &ballotStr)

			token := []byte("Token" + tokenStr)
			ballot := []byte("Ballot" + ballotStr)

			tr := blockchain.NewTransaction(token, ballot)

			go c.SendTransaction(tr)
		default:
			fmt.Println(badInputMsg)
		}

	}

	fmt.Printf("%v\n", waitMsg)
	wg.Wait()
	fmt.Println("\n\n\nDONE\n\n\n")
	fmt.Println(c)
	fmt.Println("\n\n\nDONE\n\n\n")
}
