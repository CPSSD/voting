package blockchain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"reflect"
	"sync"
)

type Configuration struct {
	MyAddr string
	MyPort string
	Lock   sync.RWMutex
	Peers  map[string]bool
}

func (c *Chain) ReceiveTransaction(t *Transaction, _ *struct{}) (err error) {
	pool := <-c.TransactionPool
    ctrs := <-c.CurrentTransactions
	inPool := false
    for _, tr := range ctrs {
        if reflect.DeepEqual(t.Header.VoteToken, tr.Header.VoteToken) {
            inPool = true
            break
        }
    }
    if !inPool {
        for _, tr := range pool {
    		if reflect.DeepEqual(t.Header.VoteToken, tr.Header.VoteToken) {
    			inPool = true
    			break
    		}
    	}
    }
	if inPool || c.contains(t) {
		c.TransactionPool <- pool
        c.CurrentTransactions <- ctrs
		return
	}

	pool = append(pool, *t)
	c.TransactionPool <- pool
    c.CurrentTransactions <- ctrs

    if len(pool) >= blockSize {
        select {
        case <- c.TransactionReady:
            c.TransactionReady <- true
        default:
            c.TransactionReady <- true
        }
    }

	go c.SendTransaction(t)
	return
}

func (c *Chain) SendTransaction(tr *Transaction) {

	c.conf.Lock.RLock()
	peers := c.conf.Peers
	c.conf.Lock.RUnlock()

	for k, _ := range peers {
		if k == c.conf.MyAddr+c.conf.MyPort {
			continue
		}
		//fmt.Println(c.conf.MyPort + ": sending transaction to -> " + k)
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

	return
}

func (c *Chain) syncPeers() {

	c.conf.Lock.RLock()
	peers := c.conf.Peers
	c.conf.Lock.RUnlock()

	for k, _ := range peers {
		if k == c.conf.MyAddr+c.conf.MyPort {
			continue
		}
		conn, err := rpc.DialHTTP("tcp", k)
		if err != nil {
			continue
		}

		var newPeers map[string]bool

		c.conf.Lock.RLock()
		err = conn.Call("Chain.GetPeers", c.conf.Peers, &newPeers)
		c.conf.Lock.RUnlock()
        conn.Close()
		if err != nil {
			continue
		}

		c.conf.Lock.Lock()
		c.conf.Peers = newPeers
		c.conf.Lock.Unlock()
	}

	return
}

func (c *Chain) PrintPeers() {

	c.conf.Lock.RLock()
	peers := c.conf.Peers
	c.conf.Lock.RUnlock()
	fmt.Printf("Peers:\n")
	for k, _ := range peers {
		fmt.Printf("\t%v\n", k)
	}
}

func (c *Chain) PrintPool() {

    pool := <- c.TransactionPool
    c.TransactionPool <- pool
    for _, tr := range pool {
        fmt.Println(tr)
    }
}

func (c *Chain) GetPeers(myPeers *map[string]bool, r *map[string]bool) error {

	c.conf.Lock.Lock()
	for key, _ := range *myPeers {
		c.conf.Peers[key] = true
	}
	*r = c.conf.Peers
	c.conf.Lock.Unlock()

	return nil
}

func (c *Chain) SavePeers(filename string) {
	bytes, err := json.Marshal(c.conf)
	err = ioutil.WriteFile(filename, bytes, 0777)
	if err != nil {
		panic(err)
	}
}

func (c *Chain) addPeer(p string) {
	c.conf.Lock.Lock()
	c.conf.Peers[p] = true
	c.conf.Lock.Unlock()
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
