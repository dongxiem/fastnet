package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/Dongxiem/fastnet"
	"github.com/Dongxiem/fastnet/connection"
	"github.com/Dongxiem/fastnet/tool/sync/atomic"
)

// Server example
type Server struct {
	clientNum     atomic.Int64
	maxConnection int64
	server        *fastnet.Server
}

// New：创建一个新的 server
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

// Start：开启 Server
func (s *Server) Start() {
	s.server.Start()
}

// Stop：停止 Server
func (s *Server) Stop() {
	s.server.Stop()
}

// OnConnect：回调函数
func (s *Server) OnConnect(c *connection.Connection) {
	// 每连接一个客户端用户则计数原子加 1
	s.clientNum.Add(1)
	log.Println(" OnConnect ： ", c.PeerAddr())
	// 判断用户连接数是否大于最大限制连接数
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

// OnClose：关闭时回调函数
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
	// 设置最大连接数为 10000
	s, err := New("", "1833", 10000)
	if err != nil {
		panic(err)
	}
	defer s.Stop()

	s.Start()
}
