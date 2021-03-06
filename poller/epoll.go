// +build linux

package poller

import (
	"github.com/Dongxiem/fastnet/log"
	"github.com/Dongxiem/fastnet/tool/sync/atomic"
	"golang.org/x/sys/unix"
)

// readEvent：读事件，默认为水平触发
const readEvent = unix.EPOLLIN | unix.EPOLLPRI
// writeEvent：写事件，默认为水平触发
const writeEvent = unix.EPOLLOUT

// Poller：结构体封装
type Poller struct {
	fd       int           // 文件句柄
	eventFd  int           // 事件句柄
	running  atomic.Bool   // 判断 Poller 是否在执行当中
	waitDone chan struct{} // 通过空结构体 chan 进行 goroutine 同步
}

// Create：创建一个 Poller
func Create() (*Poller, error) {
	// 使用 unix.EpollCreate1 创建一个原始 Epoll
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	// 使用 unix.Syscall 系统调用得到 r0，并得到其 fd
	// 这里需要了解关于 SYS_EVENTFD2 相关的原理内容，是一个用来通知事件的文件描述符
	r0, _, errno := unix.Syscall(unix.SYS_EVENTFD2, 0, 0, 0)
	if errno != 0 {
		return nil, errno
	}
	eventFd := int(r0)

	// 使用 unix.EpollCtl 系统调用进行 Epoll 事件注册添加
	err = unix.EpollCtl(fd, unix.EPOLL_CTL_ADD, eventFd, &unix.EpollEvent{
		// 传入需要关注的事件及该事件 Fd
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

// Wake：唤醒调用
func (ep *Poller) Wake() error {
	// 进行 unix.Write 系统调用，对 ep.eventFd 对应的文件进行写入 wakeBytes 任意字符即可唤醒
	_, err := unix.Write(ep.eventFd, wakeBytes)
	return err
}

var buf = make([]byte, 8)

// wakeHandlerRead: 唤醒读取处理
func (ep *Poller) wakeHandlerRead() {
	// 通过 unix.Read 系统调用，对 ep.eventFd 对应的文件进行读取
	n, err := unix.Read(ep.eventFd, buf)
	// 只是读了，但是并没有对数据进行啥处理
	if err != nil || n != 8 {
		log.Error("wakeHandlerRead", err, n)
	}
}

// Close：关闭
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

// add：对指定 fd 进行指定 events 事件的添加
func (ep *Poller) add(fd int, events uint32) error {
	return unix.EpollCtl(ep.fd, unix.EPOLL_CTL_ADD, fd, &unix.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
}

// AddRead：注册需要关注的 fd 到 epoll，并注册为可读事件
func (ep *Poller) AddRead(fd int) error {
	return ep.add(fd, readEvent)
}

// AddWrite：注册 fd 到 epoll，并注册可写事件
func (ep *Poller) AddWrite(fd int) error {
	return ep.add(fd, writeEvent)
}

// Del：从 epoll 中删除对应的 fd 事件
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

// EnableReadWrite：使能 fd 注册事件为可读可写事件
func (ep *Poller) EnableReadWrite(fd int) error {
	return ep.mod(fd, readEvent|writeEvent)
}

// EnableWrite：使能 fd 注册事件为可写事件
func (ep *Poller) EnableWrite(fd int) error {
	return ep.mod(fd, writeEvent)
}

// EnableRead：使能 fd 注册事件为可读事件
func (ep *Poller) EnableRead(fd int) error {
	return ep.mod(fd, readEvent)
}

// Poll：启动 epoll 进行事件读写等待循环，handler 为事件到来时的处理函数
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
		// EpollWait 调用，返回触发事件的个数 n，这个函数中是一个死循环，程序会阻塞在此处等待 epoll 的”通知“，然后处理就绪的 fd
		// 当有 fd 就绪的时候，syscall.EpollWait 函数返回，并且将就绪的 fd 通过 events 传出，返回值 n 为就绪 fd 的个数。
		// 然后循环逐个遍历就绪的 fd，调用回调函数处理。
		// msec 设置为 -1，也即 timeout 设置为 -1，会无限期阻塞
		// 所谓 Reactor 『非阻塞 I/O』的核心思想是指避免阻塞在 read() 或者 write() 或者其他的 I/O 系统调用上，这样可以最大限度的复用 event-loop 线程，让一个线程能服务于多个 sockets。
		// 在 Reactor 模式中，I/O 线程只能阻塞在 I/O multiplexing 函数上（select/poll/epoll_wait）。
		n, err := unix.EpollWait(ep.fd, events, -1)
		if err != nil && err != unix.EINTR {
			log.Error("EpollWait: ", err)
			continue
		}
		// 对 n 进行一个循环遍历
		for i := 0; i < n; i++ {
			// 先得到当前对应的文件描述符 fd
			fd := int(events[i].Fd)
			// 如果该 fd 不是我们当前 Poller 的 eventFd，需要进行 event 事件的获取，了解是什么事件发生了，然后再对应处理
			if fd != ep.eventFd {
				var rEvents Event
				if ((events[i].Events & unix.POLLHUP) != 0) && ((events[i].Events & unix.POLLIN) == 0) {
					// 错误事件
					rEvents |= EventErr
				}
				if (events[i].Events&unix.EPOLLERR != 0) || (events[i].Events&unix.EPOLLOUT != 0) {
					// 写事件
					rEvents |= EventWrite
				}
				if events[i].Events&(unix.EPOLLIN|unix.EPOLLPRI|unix.EPOLLRDHUP) != 0 {
					// 读事件
					rEvents |= EventRead
				}
				// 当 epoll 检测到有就绪的 fd 时，会逐个调用上面的回调函数，主要逻辑也在这里。
				handler(fd, rEvents)
			} else {
				ep.wakeHandlerRead()
				wake = true
			}
		}
		// 如果 wake 置为 True，意思即被唤醒
		if wake {
			// 使用 handler 去查看并处理剩余事件，进行完美退出
			handler(-1, 0)
			// 再将 wake 置为 false
			wake = false
			// 进行退出，退出时候会延迟调用 close(ep.waitDone)
			if !ep.running.Get() {
				return
			}
		}

		// 如果此时监听事件达到最大，则进行扩容
		if n == len(events) {
			events = make([]unix.EpollEvent, n*2)
		}
	}
}
