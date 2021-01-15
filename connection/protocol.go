package connection

import (
	"github.com/Dongxiem/gfaio/tool/ringbuffer"
)

var _ Protocol = &DefaultProtocol{}

// Protocol：自定义协议编解码接口
type Protocol interface {
	// 拆包
	UnPacket(c *Connection, buffer *ringbuffer.RingBuffer) (interface{}, []byte)
	// 装包
	Packet(c *Connection, data []byte) []byte
}

// DefaultProtocol：默认 Protocol
type DefaultProtocol struct{}

// UnPacket：拆包
func (d *DefaultProtocol) UnPacket(c *Connection, buffer *ringbuffer.RingBuffer) (interface{}, []byte) {
	ret := buffer.Bytes()
	buffer.RetrieveAll()
	return nil, ret
}

// Packet：装包
func (d *DefaultProtocol) Packet(c *Connection, data []byte) []byte {
	return data
}
