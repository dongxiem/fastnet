package eventloop

import (
	"sync"

	"github.com/Dongxiem/fastnet/log"
	"github.com/Dongxiem/fastnet/poller"
	"github.com/Dongxiem/fastnet/tool/sync/atomic"
	"github.com/Dongxiem/fastnet/tool/sync/spinlock"
)

// Socket：socket 接口
type Socket interface {
	HandleEvent(fd int, events poller.Event)
	Close() error
}

// EventLoop：事件循环
type EventLoop struct {
	poll    *poller.Poller // Poller
	sockets sync.Map //
	packet  []byte // 临时缓冲区

	eventHandling atomic.Bool // eventHandling 表明事件是否正在处理

	pendingFunc []func()          // 添加 EventLoop 待执行函数到 pendingFunc 中，是一个函数切片
	mu          spinlock.SpinLock // 自旋锁
}

// New：创建一个 EventLoop
func New() (*EventLoop, error) {
	p, err := poller.Create()
	if err != nil {
		return nil, err
	}

	return &EventLoop{
		poll:   p,
		packet: make([]byte, 0xFFFF),
	}, nil
}

// PacketBuf：内部使用，临时缓冲区
func (l *EventLoop) PacketBuf() []byte {
	return l.packet
}

// DeleteFdInLoop 删除 fd
func (l *EventLoop) DeleteFdInLoop(fd int) {
	if err := l.poll.Del(fd); err != nil {
		log.Error("[DeleteFdInLoop]", err)
	}
	l.sockets.Delete(fd)
}

// AddSocketAndEnableRead：增加 Socket 到时间循环中，并注册可读事件
func (l *EventLoop) AddSocketAndEnableRead(fd int, s Socket) error {
	var err error
	// 并发添加当前 socket 到 sync.map 当中
	l.sockets.Store(fd, s)
	// 并将该 socket 添加到可写事件，如果失败则从 sync.map 中进行删除
	if err = l.poll.AddRead(fd); err != nil {
		l.sockets.Delete(fd)
		return err
	}
	return nil
}

// EnableReadWrite：注册可读可写事件
func (l *EventLoop) EnableReadWrite(fd int) error {
	return l.poll.EnableReadWrite(fd)
}

// EnableRead：只注册可写事件
func (l *EventLoop) EnableRead(fd int) error {
	return l.poll.EnableRead(fd)
}

// RunLoop：启动事件循环
func (l *EventLoop) RunLoop() {
	l.poll.Poll(l.handlerEvent)
}

// Stop：关闭事件循环
func (l *EventLoop) Stop() error {
	// sync.map 自身提供了Range方法，通过回调的方式遍历 sync.map
	l.sockets.Range(func(key, value interface{}) bool {
		// 这里进行了一次接口类型判断，判断 value 是否为 Socket 接口类型，并得到匹配之后的 s
		s, ok := value.(Socket)
		if !ok {
			log.Error("value.(Socket) fail")
		} else {
			// 关闭 socket
			if err := s.Close(); err != nil {
				log.Error(err)
			}
		}
		return true
	})
	// 最后并关闭 poll，返回其成功与否标志位
	return l.poll.Close()
}

// QueueInLoop：添加 func 到事件循环中执行
func (l *EventLoop) QueueInLoop(f func()) {
	l.mu.Lock()
	l.pendingFunc = append(l.pendingFunc, f)
	l.mu.Unlock()

	if !l.eventHandling.Get() {
		if err := l.poll.Wake(); err != nil {
			log.Error("QueueInLoop Wake loop, ", err)
		}
	}
}

// handlerEvent：进行事件处理
func (l *EventLoop) handlerEvent(fd int, events poller.Event) {
	l.eventHandling.Set(true)

	if fd != -1 {
		// 根据 fd 取出对应的 socket
		s, ok := l.sockets.Load(fd)
		if ok {
			// 然后调用 socket 自身的函数进行相对应的事件处理
			s.(Socket).HandleEvent(fd, events)
		}
	}

	l.eventHandling.Set(false)

	l.doPendingFunc()
}

// doPendingFunc：进行待处理函数的处理
func (l *EventLoop) doPendingFunc() {
	l.mu.Lock()
	pf := l.pendingFunc
	l.pendingFunc = nil
	l.mu.Unlock()

	length := len(pf)
	for i := 0; i < length; i++ {
		pf[i]()
	}
}
