package main

import (
	"encoding/binary"
	"github.com/Dongxiem/fastnet/connection"
	"github.com/Dongxiem/fastnet/tool/ringbuffer"
	"github.com/gobwas/pool/pbytes"
)

const exampleHeaderLen = 4

type ExampleProtocol struct{}

// UnPacket：拆包
func (d *ExampleProtocol) UnPacket(c *connection.Connection, buffer *ringbuffer.RingBuffer) (interface{}, []byte) {
	if buffer.VirtualLength() > exampleHeaderLen {
		// 根据头部得到 data 包长度得到一个空的 buf
		buf := pbytes.GetLen(exampleHeaderLen)
		// 延迟调用
		defer pbytes.Put(buf)
		// 进行读取数据到 buf 中
		_, _ = buffer.VirtualRead(buf)
		// 并获取 data 长度
		dataLen := binary.BigEndian.Uint32(buf)
		// 根据 dataLen 进行数据的获取
		if buffer.VirtualLength() >= int(dataLen) {
			ret := make([]byte, dataLen)
			_, _ = buffer.VirtualRead(ret)

			buffer.VirtualFlush()
			return nil, ret
		} else {
			buffer.VirtualRevert()
		}
	}
	return nil, nil
}

// Packet：装包
func (d *ExampleProtocol) Packet(c *connection.Connection, data []byte) []byte {
	// 获取 data 的长度
	dataLen := len(data)
	// 根据 data 的长度和 header 的长度，创建一个新的 byte 切片
	ret := make([]byte, exampleHeaderLen+dataLen)
	// 将 data 长度写入 ret 的前四个字节位置当中，且是大端序
	binary.BigEndian.PutUint32(ret, uint32(dataLen))
	// 拷贝剩余
	copy(ret[4:], data)
	return ret
}
