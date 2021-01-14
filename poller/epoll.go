// +build linux

package poller

import (
	"github.com/Allenxuxu/gev/log"
	"github.com/Allenxuxu/toolkit/sync/atomic"
	"golang.org/x/sys/unix"
)

// readEvent：读事件
const readEvent = unix.EPOLLIN | unix.EPOLLPRI
// writeEvent：写事件
const writeEvent = unix.EPOLLOUT

// Poller：Epoll封装
type Poller struct {
	fd       int // 文件句柄
	eventFd  int // 事件句柄
	running  atomic.Bool // 是否在执行当中
	waitDone chan struct{} //  通过阻塞通道进行 goroutine 同步
}

// Create：创建一个 Poller
func Create() (*Poller, error) {
	// 使用 unix.EpollCreate1 创建一个原始 Epoll
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	// 使用 unix.Syscall 系统调用得到 r0
	r0, _, errno := unix.Syscall(unix.SYS_EVENTFD2, 0, 0, 0)
	if errno != 0 {
		return nil, errno
	}
	eventFd := int(r0)

	// 使用 unix.EpollCtl 系统调用进行 Epoll 事件注册
	err = unix.EpollCtl(fd, unix.EPOLL_CTL_ADD, eventFd, &unix.EpollEvent{
		// 传入 Epoll 需要关注的事件及 Fd
		Events: unix.EPOLLIN,
		Fd:     int32(eventFd),
	})
	if err != nil {
		_ = unix.Close(fd)
		_ = unix.Close(eventFd)
		return nil, err
	}
	// 返回这个处理之后的 Poller 封装
	return &Poller{
		fd:       fd,
		eventFd:  eventFd,
		waitDone: make(chan struct{}),
	}, nil
}

// wakeBytes：唤醒字符切片
var wakeBytes = []byte{1, 0, 0, 0, 0, 0, 0, 0}

// Wake：唤醒 epoll
func (ep *Poller) Wake() error {
	// 进行 unix.Write 系统调用，对 ep.eventFd 对应的文件进行写入 wakeBytes 任意字符即可唤醒
	_, err := unix.Write(ep.eventFd, wakeBytes)
	return err
}

var buf = make([]byte, 8)

// wakeHandlerRead: 唤醒处理读取
func (ep *Poller) wakeHandlerRead() {
	// 通过 unix.Read 系统调用，对 ep.eventFd 对应的文件进行读取
	n, err := unix.Read(ep.eventFd, buf)
	// 只是读了，但是并没有对数据进行啥处理
	if err != nil || n != 8 {
		log.Error("wakeHandlerRead", err, n)
	}
}

// Close：关闭 epoll
func (ep *Poller) Close() (err error) {
	// 如果 Poller 的状态并没有在运行，也没关闭一说
	if !ep.running.Get() {
		return ErrClosed
	}
	// 然后对该原子状态置为 false
	ep.running.Set(false)
	// 然后调用 Wake 进行唤醒所有的 Epoll
	if err = ep.Wake(); err != nil {
		return
	}
	// 在此处阻塞等待通道事件的到来
	<-ep.waitDone
	// 进行系统调用关闭所有的文件
	_ = unix.Close(ep.fd)
	_ = unix.Close(ep.eventFd)
	return
}

// add：对指定 fd 进行 events 事件的添加
func (ep *Poller) add(fd int, events uint32) error {
	return unix.EpollCtl(ep.fd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
}

// AddRead：注册 fd 到 epoll，并注册可读事件
func (ep *Poller) AddRead(fd int) error {
	return ep.add(fd, readEvent)
}

// AddWrite：注册 fd 到 epoll，并注册可写事件
func (ep *Poller) AddWrite(fd int) error {
	return ep.add(fd, writeEvent)
}

// Del：从 epoll 中删除 fd
func (ep *Poller) Del(fd int) error {
	return unix.EpollCtl(ep.fd, unix.EPOLL_CTL_DEL, fd, nil)
}

// mod ：修改已经注册的 fd 的监听事件
func (ep *Poller) mod(fd int, events uint32) error {
	return unix.EpollCtl(ep.fd, unix.EPOLL_CTL_MOD, fd, &unix.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
}

// EnableReadWrite：修改 fd 注册事件为可读可写事件
func (ep *Poller) EnableReadWrite(fd int) error {
	return ep.mod(fd, readEvent|writeEvent)
}

// EnableWrite：修改 fd 注册事件为可写事件
func (ep *Poller) EnableWrite(fd int) error {
	return ep.mod(fd, writeEvent)
}

// EnableRead：修改 fd 注册事件为可读事件
func (ep *Poller) EnableRead(fd int) error {
	return ep.mod(fd, readEvent)
}

// Poll：启动 epoll wait 循环
func (ep *Poller) Poll(handler func(fd int, event Event)) {
	// 延迟关闭
	defer func() {
		close(ep.waitDone)
	}()

	// 监听事件切片
	events := make([]unix.EpollEvent, waitEventsBegin)
	// wake 布尔值
	var wake bool
	// 置当前 Poller 运行状态为 True
	ep.running.Set(true)
	// 死循环进行监听
	for {
		// EpollWait 调用，返回触发事件的个数 n
		n, err := unix.EpollWait(ep.fd, events, -1)
		if err != nil && err != unix.EINTR {
			log.Error("EpollWait: ", err)
			continue
		}
		// 对 n 进行一个循环遍历
		for i := 0; i < n; i++ {
			// 先得到其对应的文件描述符 fd
			fd := int(events[i].Fd)
			// 如果该 fd 不是我们当前 Poller 的 eventFd，需要进行 event 事件的获取，了解是什么事件发生了，然后再对应处理
			if fd != ep.eventFd {
				var rEvents Event
				if ((events[i].Events & unix.POLLHUP) != 0) && ((events[i].Events & unix.POLLIN) == 0) {
					rEvents |= EventErr
				}
				if (events[i].Events&unix.EPOLLERR != 0) || (events[i].Events&unix.EPOLLOUT != 0) {
					rEvents |= EventWrite
				}
				if events[i].Events&(unix.EPOLLIN|unix.EPOLLPRI|unix.EPOLLRDHUP) != 0 {
					rEvents |= EventRead
				}

				handler(fd, rEvents)
			} else {
				// 该 fd 是当前 Poller 的 eventFd，证明当前 Poller 中对应的 Epoll 被唤醒
				ep.wakeHandlerRead()
				wake = true
			}
		}
		// 如果 wake 置为 True
		if wake {
			handler(-1, 0)
			// 再将 wake 置为 false
			wake = false
			if !ep.running.Get() {
				return
			}
		}

		// 如果此时监听事件达到最大，，则进行扩容
		if n == len(events) {
			events = make([]unix.EpollEvent, n*2)
		}
	}
}
