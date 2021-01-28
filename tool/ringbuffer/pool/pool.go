package pool

import (
	"sync"

	"github.com/Dongxiem/fastnet/tool/ringbuffer"
)

var DefaultPool = New(1024)

// Get：获取元素
func Get() *ringbuffer.RingBuffer {
	return DefaultPool.Get()
}

// Put：存放元素
func Put(r *ringbuffer.RingBuffer) {
	DefaultPool.Put(r)
}

// RingBufferPool：定义 RingBufferPool 结构体
type RingBufferPool struct {
	pool *sync.Pool
}

// New：根据初始化大小 initSize，初始化一个 RingBufferPool
func New(initSize int) *RingBufferPool {
	return &RingBufferPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return ringbuffer.New(initSize)
			},
		},
	}
}

func (p *RingBufferPool) Get() *ringbuffer.RingBuffer {
	r, _ := p.pool.Get().(*ringbuffer.RingBuffer)
	return r
}

// Put：存放元素
func (p *RingBufferPool) Put(r *ringbuffer.RingBuffer) {
	p.pool.Put(r)
}
