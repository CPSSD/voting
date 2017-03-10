// Node.go is a representation of a peer that
// will exist on the network. A peer should be able
// to connect to multiple peers, and should be able
// to discover new peers given its current connections.
package main

import (
    "log"
    "bytes"
    "os"
    "time"
	"fmt"
	"net"
    "net/rpc"
    "net/http"
    "io/ioutil"
	"sync"
    "encoding/json"
)

type Configuration struct {
	MyAddr  string
	MyPort  string
    IsMaster bool
	Master  string
	Message string
    Lock sync.RWMutex
    Peers map[string]bool
}

var conf Configuration
var logger *log.Logger
func main() {
    var buf bytes.Buffer

    logger = log.New(&buf, "logger: ", log.Lshortfile)
    defer fmt.Print(&buf)

    filename := string(os.Args[1])

    // bytes, err := SetupConf(&conf)
    // if err != nil {
    //     panic(err)
    // }
    //
    // err = ioutil.WriteFile(filename, bytes, 0777)
    // if err != nil {
    //     panic(err)
    // }

    bs, err := ioutil.ReadFile(filename)
    if err != nil {
        panic(err)
    }

    err = json.Unmarshal(bs, &conf)
    if err != nil {
        panic(err)
    }

    addPeer(conf.Master)
    addPeer(conf.MyAddr + conf.MyPort)

    fmt.Println("ready to do stuff")


    rpc.Register(&conf)
    rpc.HandleHTTP()

    ln, err := net.Listen("tcp", conf.MyPort)
    if err != nil {
        logger.Println(err)
    }

    go http.Serve(ln, nil)
    go func() {
        for {
            timer := time.NewTimer(time.Second * 3)
            <- timer.C
            syncPeers()
        }
    }()
    for {

    }
}

func SetupConf(c *Configuration) (bytes []byte, err error) {
    c.MyAddr = "localhost"
    c.MyPort = ":8092"
    c.IsMaster = true
    c.Master = ":8092"
    c.Peers = make(map[string]bool, 0)
    c.Message = "I am the Master node"

    bytes, err = json.Marshal(c)
    return bytes, err
}

func (c *Configuration) GetPeers(myPeers *map[string]bool, r *map[string]bool) error {

    conf.Lock.Lock()
    for key, _ := range *myPeers {
        conf.Peers[key] = true
    }
    *r = conf.Peers
    conf.Lock.Unlock()

    return nil
}

func syncPeers() {

    conf.Lock.RLock()
    peers := conf.Peers
    conf.Lock.RUnlock()

    for k, _ := range peers {
        if k == conf.MyAddr + conf.MyPort || conf.IsMaster {
            continue
        }
        fmt.Println(conf.MyPort+": dialing to -> "+k)
    	conn, err := rpc.DialHTTP("tcp", k)
    	if err != nil {
    		logger.Println("RPC dial error:",err)
            continue
    	}

        var newPeers map[string]bool

        conf.Lock.RLock()
        err = conn.Call("Configuration.GetPeers", conf.Peers, &newPeers)
        conf.Lock.RUnlock()
        if err != nil {
            logger.Println("RPC call error:",err)
            continue
        }
        conf.Lock.Lock()
        conf.Peers = newPeers
        conf.Lock.Unlock()

        fmt.Println(conf.MyPort+": got peers -> ")
        for k, _ := range newPeers {
            fmt.Println("peer: " + k)
        }
    }

	return
}

func addPeer(p string) {
    conf.Lock.Lock()
    conf.Peers[p] = true
    conf.Lock.Unlock()
}
