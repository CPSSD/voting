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

// Configuration contains information about our node, along with
// information about peers in the network, VoteTokens for the
// election, and information required for reconstruction of the
// election private key.
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

// ElectionSecret contains two shares which are required in the
// reconstruction of the private election key.
type ElectionSecret struct {
	Lambda crypto.Share
	Mu     crypto.Share
}

// PrintKey prints our current interpolation of the private
// election key.
func (c *Chain) PrintKey() {
	fmt.Println(c.conf.ElectionKey.Lambda)
	fmt.Println(c.conf.ElectionKey.Mu)
}

// GetFormat returns the format defined for the election.
func (c *Chain) GetFormat() election.Format {
	return c.conf.ElectionFormat
}

// GetElectionKey returns the election key as currently
// interpolated by the node.
func (c *Chain) GetElectionKey() crypto.PrivateKey {
	return c.conf.ElectionKey
}

// GetVoteToken returns the vote token of the
// user associated with this node.
func (c *Chain) GetVoteToken() string {
	return c.conf.MyToken
}

// CollectBallots will gather all the ballots from the current
// chain and return them.
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

// ReceiveTransaction is an RPC function which allows a node to
// recieve transactions from the network.
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

// ChainUpdate contains the blocks associated with a given chain.
type ChainUpdate struct {
	Blocks []Block
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

// ReceiveKeyShare is an RPC function which allows a node to
// receive a share of the private key from other nodes.
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

// GetChain is an RPC call which will set the value of altChain
// to the value of this node's current set of blocks. The value
// of empty is unused.
func (c *Chain) GetChain(empty bool, altChain *[]Block) error {
	b := <-c.blocks
	c.blocks <- b

	*altChain = b
	return nil
}

// BlockUpdate contains the latest block, along with details of
// the peer who created it and the length of the chain which it
// was added to.
type BlockUpdate struct {
	LatestBlock Block
	Peer        string
	ChainLength uint32
}

// ReceiveBlockUpdate is an RPC function which allows a node
// to receive an update about a block for further processing.
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

// SendTransaction will invoke the broadcasting of a
// transaction to peers on the network.
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

// BroadcastShare will add a user's share of the election
// key to the pool of shares which are broadcast regularly.
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

// PrintPeers displays the list of peers known to a node
func (c *Chain) PrintPeers() {

	peers := <-c.Peers
	c.Peers <- peers

	fmt.Printf("Peers:\n")
	for k, _ := range peers {
		fmt.Printf("\t%v\n", k)
	}
}

// PrintPool displays a list of transactions which are
// not yet incorporated into the chain.
func (c *Chain) PrintPool() {

	pool := <-c.TransactionPool
	c.TransactionPool <- pool

	for _, tr := range pool {
		fmt.Println(tr)
	}
}

// GetPeers is an RPC function which allows peers to
// combine their peer lists for syncing.
func (c *Chain) GetPeers(myPeers *map[string]bool, r *map[string]bool) error {

	peers := <-c.Peers

	for key, _ := range *myPeers {
		peers[key] = true
	}
	*r = peers
	c.Peers <- peers

	return nil
}

func (c *Chain) addPeer(p string) {
	peers := <-c.Peers
	peers[p] = true
	c.Peers <- peers
}

// Init will read in a configration file and set up
// a new chain. The RPC functions are made available
// during the call of this method.
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
