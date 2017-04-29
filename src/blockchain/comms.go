package blockchain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"reflect"
)

type Configuration struct {
	MyAddr string
	MyPort string
	Peers  map[string]bool
}

func (c *Chain) ReceiveTransaction(t *Transaction, _ *struct{}) (err error) {
	pool := <-c.TransactionPool
    seen := <-c.SeenTrs

    // if the tr was seen in our chain, then don't add it
    if _, ok := seen[string(t.Header.VoteToken[:])]; ok {
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

    // else add it
	pool = append(pool, *t)
    c.SeenTrs <- seen
    c.TransactionPool <- pool

    // send it
    go c.SendTransaction(t)

    return nil
}

type ChainUpdate struct {
    Blocks []Block
    //SeenTrs []string
}

func (c *Chain) getChainUpdateFrom(peer string) (altChain *[]Block, err error) {
    conn, err := rpc.DialHTTP("tcp", peer)
    if err != nil {
        fmt.Println("c.getChainUpdateFrom:",peer,"error with HTTP:")
        fmt.Println(err)
        return altChain, err
    }
    empty := true
    altChain = new([]Block)
    err = conn.Call("Chain.GetChain", empty, &altChain)
    if err != nil {
        fmt.Println("c.getChainUpdateFrom:",peer,"error with RPC:")
        fmt.Println(altChain)
        fmt.Println(err)
        return altChain, err
    }
    conn.Close()
    return altChain, err
}

func (c *Chain) GetChain(empty bool, altChain *[]Block) error {
    b := <- c.blocks
    c.blocks <- b
    // smap := <- c.SeenTrs
    // c.SeenTrs <- s
    //
    // sarr := make([]string, 0, len(s))
    // for tr, seen := range smap {
    //     if seen {
    //         sarr = append(sarr, tr)
    //     }
    // }

    *altChain = b
    // altChain.SeenTrs = sarr

    return nil
}

type BlockUpdate struct {
    LatestBlock Block
    Peer        string
    ChainLength uint32
}

func (c *Chain) ReceiveBlockUpdate(blu *BlockUpdate, _ *struct{}) (err error) {

    fmt.Println("c.ReceiveBlockUpdate: writing to c.BlockUpdate...")
	c.BlockUpdate <- *blu
    fmt.Println("c.ReceiveBlockUpdate: wrote to c.BlockUpdate")
	return
}

func (c *Chain) sendBlock(bl *Block) {

    fmt.Println("c.sendBlock: sending block to peers...")
    peers := <- c.Peers
    c.Peers <- peers
    blocks := <- c.blocks
    c.blocks <- blocks
    update := &BlockUpdate{
        LatestBlock: *bl,
        Peer:        c.conf.MyAddr+c.conf.MyPort,
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
    fmt.Println("c.sendBlock: done sending to peers")
	return
}

func (c *Chain) SendTransaction(tr *Transaction) {

    fmt.Println("c.SendTransaction: sending to peers")
    peers := <- c.Peers

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
    fmt.Println("c.SendTransaction: done sending to peers")
	return
}

func (c *Chain) syncPeers() {

	peers := <- c.Peers
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
            fmt.Println("Error getting peers")
			continue
		}

        _ = <- c.Peers
        c.Peers <- newPeers
	}

	return
}

func (c *Chain) PrintPeers() {

    peers := <- c.Peers
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

	peers := <- c.Peers

    for key, _ := range *myPeers {
		peers[key] = true
	}
	*r = peers
    c.Peers <- peers

	return nil
}

func (c *Chain) SavePeers(filename string) {
    bl:= <- c.blocks
    c.blocks <- bl
	bytes, err := json.Marshal(bl)
	err = ioutil.WriteFile(filename, bytes, 0777)
	if err != nil {
		panic(err)
	}
}

func (c *Chain) addPeer(p string) {
    peers := <- c.Peers
	peers[p] = true
    c.Peers <- peers
}

func (c *Chain) Init(filename string) (err error) {

	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bs, &c.conf)
	if err != nil {
		panic(err)
	}
    c.Peers <- c.conf.Peers
	c.addPeer(c.conf.MyAddr + c.conf.MyPort)

	rpc.Register(c)
	rpc.HandleHTTP()

	ln, err := net.Listen("tcp", c.conf.MyPort)
	if err != nil {
		panic(err)
	}

	go http.Serve(ln, nil)

	return err
}
