## fastnet 简介

- `fastnet` 是一个高性能、轻量级、非阻塞的 TCP 网络库，用户可以使用 `fastnet` 来实现自己的应用层网络协议，从而构建出自己的应用层网络应用；
- 在某些极端的网络业务场景，比如海量连接、高频短连接、网络小包等场景，`fastnet` 在性能和资源占用上都超过 `Go` 原生 `net` 包的`goroutine-per-connection` 模式。

## fastnet 特性

- 使用 `Epoll` 水平触发的IO多路复用技术，非阻塞IO，使用 `Reactor` 模式；
- 使用多线程充分利用多核CPU，使用动态扩容 `Ring Buffer` 实现读写缓冲区；
- 支持异步读写操作、支持 `SO_REUSEPORT` 端口重用；
- 灵活的事件定时器，可以定时任务，延时任务；
- 支持 `WebSocket`，同时支持自定义协议，处理 `TCP` 粘包；

**TODO：**

- 更多的负载方式（轮询（Round-Robin）、一致性哈希（`consistent hashing`）等）；



## Installation

1. 获得并安装 `fastnet`

```shell
$ go get -u github.com/dongxiem/fastnet
```

2. 进行 `import`

```go
import "github.com/dongxiem/fastnet"
```



## fastnet 性能测试

- [bench-echo](https://github.com/dongxiem/fastnet/blob/main/benchmarks/bench-echo.sh)
- [bench-pingpong](https://github.com/dongxiem/fastnet/blob/main/benchmarks/bench-pingpong.sh)



## Example

```go
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
```



