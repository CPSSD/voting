package blockchain

import (
	"crypto/dsa"
	"encoding/json"
	"fmt"
	"github.com/CPSSD/voting/src/crypto"
	"github.com/CPSSD/voting/src/election"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"reflect"
)

type Configuration struct {
	MyAddr    string
	MyPort    string
	Peers     map[string]bool
	SyncPeers bool

	PrivateKey dsa.PrivateKey

	VoteTokens map[string]dsa.PublicKey
	MyToken    string

	ElectionFormat election.Format

	ElectionKey           crypto.PrivateKey
	ElectionKeyShare      ElectionSecret
	ElectionLambdaModulus *big.Int
	ElectionMuModulus     *big.Int
}

type ElectionSecret struct {
	Lambda crypto.Share
	Mu     crypto.Share
}

func (c *Chain) PrintKey() {
	fmt.Println(c.conf.ElectionKey.Lambda)
	fmt.Println(c.conf.ElectionKey.Mu)
}

func (c *Chain) GetFormat() election.Format {
	return c.conf.ElectionFormat
}

func (c *Chain) GetElectionKey() crypto.PrivateKey {
	return c.conf.ElectionKey
}

func (c *Chain) GetVoteToken() string {
	return c.conf.MyToken
}

func (c *Chain) CollectBallots() *[]election.Ballot {
	log.Println("Gathering ballots from the chain")
	blocks := <-c.blocks
	c.blocks <- blocks

	ballots := make([]election.Ballot, 0)

	for _, bl := range blocks {
		for _, tr := range bl.Transactions {
			ballots = append(ballots, tr.Ballot)
		}
	}
	log.Println("Collected the ballots from the chain")

	return &ballots
}

func (c *Chain) ReceiveTransaction(t *Transaction, _ *struct{}) (err error) {
	pool := <-c.TransactionPool
	seen := <-c.SeenTrs

	if valid := c.ValidateSignature(t); !valid {
		log.Println("Received a new transaction with invalid signature")
		c.SeenTrs <- seen
		c.TransactionPool <- pool
		return nil
	}
	// if the tr was seen in our chain, then don't add it
	if _, ok := seen[t.Header.VoteToken]; ok {
		c.SeenTrs <- seen
		c.TransactionPool <- pool
		return nil
	}

	// if the tr is in our pool, don't add it
	for _, tr := range pool {
		if reflect.DeepEqual(t.Header.VoteToken, tr.Header.VoteToken) {
			c.SeenTrs <- seen
			c.TransactionPool <- pool
			return nil
		}
	}

	log.Println("We received a new transaction")
	pool = append(pool, *t)
	c.SeenTrs <- seen
	c.TransactionPool <- pool

	go c.SendTransaction(t)

	return nil
}

type ChainUpdate struct {
	Blocks []Block
	//SeenTrs []string
}

func (c *Chain) sendKeyShareTo(share *ElectionSecret, peer string) {
	log.Println("Opening connection to ", peer)
	conn, err := rpc.DialHTTP("tcp", peer)
	if err != nil {
		return
	}
	shareCall := conn.Go("Chain.ReceiveKeyShare", share, nil, nil)
	_ = <-shareCall.Done
	conn.Close()
	log.Println("Done sending key, connection closed with", peer)
}

func (c *Chain) ReceiveKeyShare(share *ElectionSecret, _ *struct{}) (err error) {

	log.Println("Received a key share, writing to respective channel")
	c.addShare(*share)
	log.Println("Written key share to channel")
	return
}

func (c *Chain) getChainUpdateFrom(peer string) (altChain *[]Block, err error) {
	conn, err := rpc.DialHTTP("tcp", peer)
	if err != nil {
		return altChain, err
	}
	empty := true
	altChain = new([]Block)
	err = conn.Call("Chain.GetChain", empty, &altChain)
	if err != nil {
		return altChain, err
	}
	conn.Close()
	return altChain, err
}

func (c *Chain) GetChain(empty bool, altChain *[]Block) error {
	b := <-c.blocks
	c.blocks <- b

	*altChain = b
	return nil
}

type BlockUpdate struct {
	LatestBlock Block
	Peer        string
	ChainLength uint32
}

func (c *Chain) ReceiveBlockUpdate(blu *BlockUpdate, _ *struct{}) (err error) {

	log.Println("Received block update, writing to respective channel")
	c.BlockUpdate <- *blu
	return
}

func (c *Chain) sendBlock(bl *Block) {

	log.Println("Sending block to peers")
	peers := <-c.Peers
	c.Peers <- peers
	blocks := <-c.blocks
	c.blocks <- blocks
	update := &BlockUpdate{
		LatestBlock: *bl,
		Peer:        c.conf.MyAddr + c.conf.MyPort,
		ChainLength: uint32(len(blocks)),
	}

	for k, _ := range peers {
		if k == c.conf.MyAddr+c.conf.MyPort {
			continue
		}

		conn, err := rpc.DialHTTP("tcp", k)
		if err != nil {
			continue
		}
		go func() {
			blCall := conn.Go("Chain.ReceiveBlockUpdate", update, nil, nil)
			_ = <-blCall.Done
			conn.Close()
		}()
	}
	log.Println("Done sending block to peers")
	return
}

func (c *Chain) SendTransaction(tr *Transaction) {

	log.Println("Sending transaction to peers")
	peers := <-c.Peers

	for k, _ := range peers {
		if k == c.conf.MyAddr+c.conf.MyPort {
			continue
		}

		conn, err := rpc.DialHTTP("tcp", k)
		if err != nil {
			continue
		}
		go func() {
			trCall := conn.Go("Chain.ReceiveTransaction", tr, nil, nil)
			_ = <-trCall.Done
			conn.Close()
		}()
	}

	c.Peers <- peers
	log.Println("Done sending transaction to peers")
	return
}

func (c *Chain) BroadcastShare() {

	log.Println("Broadcasting our share of the election key")
	c.addShare(c.conf.ElectionKeyShare)
}

func (c *Chain) addShare(sh ElectionSecret) (exists bool) {

	shares := <-c.KeyShares
	var ok bool
	if _, ok := shares[sh.Lambda.X.String()]; !ok {
		shares[sh.Lambda.X.String()] = sh
		log.Println("Added a new share:", sh.Lambda.X.String())
	}
	c.KeyShares <- shares
	return ok
}

func (c *Chain) broadcastKeyShares() {

	shares := <-c.KeyShares
	c.KeyShares <- shares

	if len(shares) == 0 {
		return
	}

	peers := <-c.Peers
	c.Peers <- peers

	log.Println("Broadcasting our known shares")
	for _, s := range shares {
		for p, _ := range peers {
			if p != c.conf.MyAddr+c.conf.MyPort {
				log.Println("Broadcasting share", s, "to peer:", p)
				c.sendKeyShareTo(&s, p)
			}
		}
	}
}

func (c *Chain) syncPeers() {

	log.Println("Syncing peer list with peers")

	peers := <-c.Peers
	c.Peers <- peers

	for k, _ := range peers {
		if k == c.conf.MyAddr+c.conf.MyPort {
			continue
		}
		conn, err := rpc.DialHTTP("tcp", k)
		if err != nil {
			continue
		}

		var newPeers map[string]bool

		err = conn.Call("Chain.GetPeers", peers, &newPeers)
		conn.Close()
		if err != nil {
			continue
		}

		_ = <-c.Peers
		c.Peers <- newPeers
	}
	log.Println("Done syncing peer list with peers")
	return
}

func (c *Chain) PrintPeers() {

	peers := <-c.Peers
	c.Peers <- peers

	fmt.Printf("Peers:\n")
	for k, _ := range peers {
		fmt.Printf("\t%v\n", k)
	}
}

func (c *Chain) PrintPool() {

	pool := <-c.TransactionPool
	c.TransactionPool <- pool

	for _, tr := range pool {
		fmt.Println(tr)
	}
}

func (c *Chain) GetPeers(myPeers *map[string]bool, r *map[string]bool) error {

	peers := <-c.Peers

	for key, _ := range *myPeers {
		peers[key] = true
	}
	*r = peers
	c.Peers <- peers

	return nil
}

func (c *Chain) SavePeers(filename string) {
	log.Println("Saving peer list to disk")
	bl := <-c.blocks
	c.blocks <- bl
	bytes, err := json.Marshal(bl)
	err = ioutil.WriteFile(filename, bytes, 0777)
	if err != nil {
		log.Println("Could not save peer list to disk")
		log.Println(err)
	}
}

func (c *Chain) addPeer(p string) {
	peers := <-c.Peers
	peers[p] = true
	c.Peers <- peers
}

func (c *Chain) Init(filename string) (err error) {

	log.Println("Reading configuration file")
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.Unmarshal(bs, &c.conf)
	if err != nil {
		log.Fatalln(err)
	}
	c.Peers <- c.conf.Peers
	c.addPeer(c.conf.MyAddr + c.conf.MyPort)

	rpc.Register(c)
	rpc.HandleHTTP()

	ln, err := net.Listen("tcp", c.conf.MyPort)
	if err != nil {
		log.Fatalln(err)
	}

	go http.Serve(ln, nil)

	return err
}
