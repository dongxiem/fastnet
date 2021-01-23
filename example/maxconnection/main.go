package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/Dongxiem/gfaio"
	"github.com/Dongxiem/gfaio/connection"
	"github.com/Dongxiem/gfaio/tool/sync/atomic"
)

// Server example
type Server struct {
	clientNum     atomic.Int64
	maxConnection int64
	server        *fastnet.Server
}

// New server
func New(ip, port string, maxConnection int64) (*Server, error) {
	var err error
	s := new(Server)
	s.maxConnection = maxConnection
	s.server, err = fastnet.NewServer(s,
		fastnet.Address(ip+":"+port))
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Start server
func (s *Server) Start() {
	s.server.Start()
}

// Stop server
func (s *Server) Stop() {
	s.server.Stop()
}

// OnConnect callback
func (s *Server) OnConnect(c *connection.Connection) {
	s.clientNum.Add(1)
	log.Println(" OnConnect ： ", c.PeerAddr())

	if s.clientNum.Get() > s.maxConnection {
		_ = c.ShutdownWrite()
		log.Println("Refused connection")
		return
	}
}

// OnMessage callback
func (s *Server) OnMessage(c *connection.Connection, ctx interface{}, data []byte) (out []byte) {
	log.Println("OnMessage")
	out = data
	return
}

// OnClose callback
func (s *Server) OnClose(c *connection.Connection) {
	s.clientNum.Add(-1)
	log.Println("OnClose")
}

func main() {
	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			panic(err)
		}
	}()

	s, err := New("", "1833", 1)
	if err != nil {
		panic(err)
	}
	defer s.Stop()

	s.Start()
}
