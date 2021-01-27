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

// 定义 example 类型，并实现 Handler 接口对应的所有方法
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

	// new 一个 server，根据传递进来的 loops 进行工作线程循环的调配
	s, err := fastnet.NewServer(handler,
		fastnet.Network("tcp"),
		fastnet.Address(":"+strconv.Itoa(port)),
		fastnet.NumLoops(loops))
	if err != nil {
		panic(err)
	}

	// 每两秒打印一下 handle 统计
	s.RunEvery(time.Second*2, func() {
		log.Info("connections :", handler.Count.Get())
	})
	// 启动 server
	s.Start()
}
