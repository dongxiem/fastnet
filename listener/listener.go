package listener

import (
	"errors"
	"net"
	"os"

	"github.com/Dongxiem/fastnet/eventloop"
	"github.com/Dongxiem/fastnet/log"
	"github.com/Dongxiem/fastnet/poller"
	reuseport "github.com/libp2p/go-reuseport"
	"golang.org/x/sys/unix"
)

// HandleConnFunc：处理新连接回调方法
type HandleConnFunc func(fd int, sa unix.Sockaddr)

// Listener：监听TCP连接
type Listener struct {
	file     *os.File				// 文件
	fd       int 					// 文件描述符
	handleC  HandleConnFunc 		// 处理新连接函数
	listener net.Listener 			// Listener 监听
	loop     *eventloop.EventLoop 	// 事件循环
}

// New：创建一个新的 Listener 监听
func New(network, addr string, reusePort bool, loop *eventloop.EventLoop, handlerConn HandleConnFunc) (*Listener, error) {
	var listener net.Listener
	var err error
	// 判断是否端口重用，如果是端口复用的话在原端口启动监听，否则启动一个新的监听
	if reusePort {
		listener, err = reuseport.Listen(network, addr)
	} else {
		listener, err = net.Listen(network, addr)
	}
	if err != nil {
		return nil, err
	}

	// 得到一个 TCP 监听
	l, ok := listener.(*net.TCPListener)
	if !ok {
		return nil, errors.New("could not get file descriptor")
	}

	// 得到该 TCP监听对应的文件
	file, err := l.File()
	if err != nil {
		return nil, err
	}
	// 然后去取该文件的文件描述符
	fd := int(file.Fd())
	// 通过文件描述符进行设置，将该文件设置为 Nonblock
	if err = unix.SetNonblock(fd, true); err != nil {
		return nil, err
	}
	// 最后进行 Listener 的填充并返回
	return &Listener{
		file:     file,
		fd:       fd,
		handleC:  handlerConn,
		listener: listener,
		loop:     loop}, nil
}

// HandleEvent ：内部使用，供 event loop 回调处理事件
func (l *Listener) HandleEvent(fd int, events poller.Event) {
	// 如果 events 有读事件，也即有客户端进行了请求连接
	if events & poller.EventRead != 0 {
		// 进行 Accept，并得到 Accept 之后的文件描述符 nfd
		nfd, sa, err := unix.Accept(fd)
		// 进行 err 错误判断
		if err != nil {
			if err != unix.EAGAIN {
				log.Error("accept:", err)
			}
			return
		}
		// 对该 nfd 文件描述符设置为不阻塞：Nonblock
		if err := unix.SetNonblock(nfd, true); err != nil {
			_ = unix.Close(nfd)
			log.Error("set nonblock:", err)
			return
		}
		// 然后调用 handleC 继续处理
		l.handleC(nfd, sa)
	}
}

// Close ：关闭 listener
func (l *Listener) Close() error {
	// 进行一个队列循环，将队列中的所有 Listener 都进行关闭
	l.loop.QueueInLoop( func() {
		l.loop.DeleteFdInLoop(l.fd)
		if err := l.listener.Close(); err != nil {
			log.Error("[Listener] close error: ", err)
		}
	})

	return nil
}

// Fd ：返回 listener 的文件句柄
func (l *Listener) Fd() int {
	return l.fd
}
