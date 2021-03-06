package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/Dongxiem/fastnet"
	"github.com/Dongxiem/fastnet/connection"
	pb "github.com/Dongxiem/fastnet/example/protobuf/proto"
	"github.com/Dongxiem/fastnet/plugins/protobuf"
	"github.com/golang/protobuf/proto"
)

type example struct{}

func (s *example) OnConnect(c *connection.Connection) {
	log.Println(" OnConnect ： ", c.PeerAddr())
}
func (s *example) OnMessage(c *connection.Connection, ctx interface{}, data []byte) (out []byte) {
	msgType := ctx.(string)

	switch msgType {
	case "msg1":
		msg := &pb.Msg1{}
		if err := proto.Unmarshal(data, msg); err != nil {
			log.Println(err)
		}
		log.Println(msgType, msg)
	case "msg2":
		msg := &pb.Msg2{}
		if err := proto.Unmarshal(data, msg); err != nil {
			log.Println(err)
		}
		log.Println(msgType, msg)
	default:
		log.Println("unknown msg type")
	}

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

	s, err := fastnet.NewServer(handler,
		fastnet.Network("tcp"),
		fastnet.Address(":"+strconv.Itoa(port)),
		fastnet.NumLoops(loops),
		fastnet.Protocol(&protobuf.Protocol{}))
	if err != nil {
		panic(err)
	}

	log.Println("server start")
	s.Start()
}
