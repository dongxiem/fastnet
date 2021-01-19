package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/Dongxiem/gfaio"
	"github.com/Dongxiem/gfaio/connection"
)

type example struct{}

func (s *example) OnConnect(c *connection.Connection) {
	log.Println(" OnConnect ： ", c.PeerAddr())
}
func (s *example) OnMessage(c *connection.Connection, ctx interface{}, data []byte) (out []byte) {
	log.Println("OnMessage：", data)
	out = data
	return
}

func (s *example) OnClose(c *connection.Connection) {
	log.Println("OnClose")
}

func main() {
	handler := new(example)
	var port int
	var loops int

	flag.IntVar(&port, "port", 1833, "server port")
	flag.IntVar(&loops, "loops", -1, "num loops")
	flag.Parse()

	s, err := gfaio.NewServer(handler,
		gfaio.Network("tcp"),
		gfaio.Address(":"+strconv.Itoa(port)),
		gfaio.NumLoops(loops),
		gfaio.Protocol(&ExampleProtocol{}))
	if err != nil {
		panic(err)
	}

	log.Println("server start")
	s.Start()
}
