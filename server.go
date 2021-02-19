package fastnet

import (
	"errors"
	"runtime"
	"time"

	"github.com/Dongxiem/fastnet/connection"
	"github.com/Dongxiem/fastnet/eventloop"
	"github.com/Dongxiem/fastnet/listener"
	"github.com/Dongxiem/fastnet/log"
	"github.com/Dongxiem/fastnet/tool/sync"
	"github.com/RussellLuo/timingwheel"
	"golang.org/x/sys/unix"
)

// Handler：Server 注册接口
type Handler interface {
	connection.CallBack
	OnConnect(c *connection.Connection)
}

// Server：fastnet Server
type Server struct {
	loop          *eventloop.EventLoop 		// 主事件循环，负责监听客户端连接
	workLoops     []*eventloop.EventLoop 	// 其他负责处理已连接客户端的读写事件
	nextLoopIndex int 						// 下一个循环索引
	callback      Handler 					// 回调处理

	timingWheel *timingwheel.TimingWheel	// 定时器
	opts        *Options 					// 配置选项
}

// NewServer：创建 Server
func NewServer(handler Handler, opts ...Option) (server *Server, err error) {
	if handler == nil {
		return nil, errors.New("handler is nil")
	}
	options := newOptions(opts...)
	// server 创建及配置
	server = new(Server)
	server.callback = handler
	server.opts = options
	server.timingWheel = timingwheel.NewTimingWheel(server.opts.tick, server.opts.wheelSize)
	server.loop, err = eventloop.New()
	if err != nil {
		_ = server.loop.Stop()
		return nil, err
	}

	// 生成新的监听者 listener
	l, err := listener.New(server.opts.Network, server.opts.Address, options.ReusePort, server.loop, server.handleNewConnection)
	if err != nil {
		return nil, err
	}
	// 将该 listener 添加到服务器监听循环，监听可读事件
	if err = server.loop.AddSocketAndEnableRead(l.Fd(), l); err != nil {
		return nil, err
	}

	// 如果 server.opts.NumLoops 小于等于0，则设置为现机器 CPU 的个数
	if server.opts.NumLoops <= 0 {
		server.opts.NumLoops = runtime.NumCPU()
	}

	// 根据 server.opts.NumLoops 创建对应数量的 goroutine（work 协程）负责处理已连接客户端的读写事件
	wloops := make([]*eventloop.EventLoop, server.opts.NumLoops)
	for i := 0; i < server.opts.NumLoops; i++ {
		l, err := eventloop.New()
		if err != nil {
			for j := 0; j < i; j++ {
				_ = wloops[j].Stop()
			}
			return nil, err
		}
		wloops[i] = l
	}
	server.workLoops = wloops

	return
}

// RunAfter：延时任务开启
func (s *Server) RunAfter(d time.Duration, f func()) *timingwheel.Timer {
	return s.timingWheel.AfterFunc(d, f)
}

// RunEvery：定时任务，定时每 Duration 时间执行 f
func (s *Server) RunEvery(d time.Duration, f func()) *timingwheel.Timer {
	return s.timingWheel.ScheduleFunc(&everyScheduler{Interval: d}, f)
}

// nextLoop：下一个循环
func (s *Server) nextLoop() *eventloop.EventLoop {
	// TODO 更多的负载方式
	loop := s.workLoops[s.nextLoopIndex]
	s.nextLoopIndex = (s.nextLoopIndex + 1) % len(s.workLoops)
	return loop
}

// handleNewConnection：进行监听事件的分发，也即 Listener 中的调用方法
func (s *Server) handleNewConnection(fd int, sa unix.Sockaddr) {
	// 取得下一个循环的 work 线程
	loop := s.nextLoop()
	// 生成新的 connection 连接
	c := connection.New(fd, loop, sa, s.opts.Protocol, s.timingWheel, s.opts.IdleTime, s.callback)
	// 调用回调函数中的 OnConnect 方法
	s.callback.OnConnect(c)
	// 将该 socket 添加进监听循环，并且置为读监听事件
	if err := loop.AddSocketAndEnableRead(fd, c); err != nil {
		log.Error("[AddSocketAndEnableRead]", err)
	}
}

// Start：启动 Server
func (s *Server) Start() {
	// 使用 WaitGroup 进行并发模型构建
	sw := sync.WaitGroupWrapper{}
	s.timingWheel.Start()
	// 获取循环工作线程的大小
	length := len(s.workLoops)
	// 然后让每个工作线程都启动
	for i := 0; i < length; i++ {
		sw.AddAndRun(s.workLoops[i].RunLoop)
	}
	// 并开启主事件循环
	sw.AddAndRun(s.loop.RunLoop)
	// 在这里等待所有线程结束、退出
	sw.Wait()
}

// Stop：关闭 Server
func (s *Server) Stop() {
	// 先停止 timingWheel
	s.timingWheel.Stop()
	// 关闭主循环线程
	if err := s.loop.Stop(); err != nil {
		log.Error(err)
	}
	// 关闭其他工作线程
	for k := range s.workLoops {
		if err := s.workLoops[k].Stop(); err != nil {
			log.Error(err)
		}
	}
}

// Options：返回 options
func (s *Server) Options() Options {
	return *s.opts
}
