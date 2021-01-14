package poller

import "errors"

// ErrClosed 错误： 重复 close poller 错误
var ErrClosed = errors.New("poller instance is not running")

// waitEventsBegin：开始进行事件等待的阈值
const waitEventsBegin = 1024

// Event：poller 返回事件
type Event uint32

// Event poller 返回事件值
const (
	EventRead  Event = 0x1
	EventWrite Event = 0x2
	EventErr   Event = 0x80
	EventNone  Event = 0
)
