package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"

	"github.com/Dongxiem/fastnet"
	"github.com/Dongxiem/fastnet/connection"
	"github.com/Dongxiem/fastnet/log"
	"github.com/Dongxiem/fastnet/tool/sync/atomic"
)

type example struct {
	Count atomic.Int64
}

func (s *example) OnConnect(c *connection.Connection) {
	s.Count.Add(1)
	//log.Println(" OnConnect ： ", c.PeerAddr())
}
func (s *example) OnMessage(c *connection.Connection, ctx interface{}, data []byte) (out []byte) {
	//log.Println("OnMessage")
	out = data
	return
}

func (s *example) OnClose(c *connection.Connection) {
	s.Count.Add(-1)
	//log.Println("OnClose")
}

func main() {
	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			panic(err)
		}
	}()

	handler := new(example)
	var port int
	var loops int

	flag.IntVar(&port, "port", 1833, "server port")
	flag.IntVar(&loops, "loops", -1, "num loops")
	flag.Parse()

	s, err := fastnet.NewServer(handler,
		fastnet.Network("tcp"),
		fastnet.Address(":"+strconv.Itoa(port)),
		fastnet.NumLoops(loops))
	if err != nil {
		panic(err)
	}

	s.RunEvery(time.Second*2, func() {
		log.Info("connections :", handler.Count.Get())
	})

	s.Start()
}
