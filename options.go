package fastnet

import (
	"time"

	"github.com/Dongxiem/fastnet/connection"
)

// Options：服务配置
type Options struct {
	Network   string				// 网络协议
	Address   string				// 端口地址
	NumLoops  int					// work 协程，负责处理已连接客户端的读写事件
	ReusePort bool					// 是否开启端口复用

	tick      time.Duration			// 事件持续
	wheelSize int64
	IdleTime  time.Duration			// 最大空闲时间（秒）
	Protocol  connection.Protocol	// 连接协议
}

// Option ...
type Option func(*Options)

// newOptions：返回一个新的 Options 配置
func newOptions(opt ...Option) *Options {
	opts := Options{}

	for _, o := range opt {
		o(&opts)
	}
	// 默认使用 Tcp
	if len(opts.Network) == 0 {
		opts.Network = "tcp"
	}
	// 默认端口 1388
	if len(opts.Address) == 0 {
		opts.Address = ":1388"
	}
	// 默认 tick 1s
	if opts.tick == 0 {
		opts.tick = 1 * time.Millisecond
	}
	// 默认最大连接 1000
	if opts.wheelSize == 0 {
		opts.wheelSize = 1000
	}
	// 默认协议
	if opts.Protocol == nil {
		opts.Protocol = &connection.DefaultProtocol{}
	}

	return &opts
}

// ReusePort：设置 SO_REUSEPORT
func ReusePort(reusePort bool) Option {
	return func(o *Options) {
		o.ReusePort = reusePort
	}
}

// Network：暂时只支持tcp
func Network(n string) Option {
	return func(o *Options) {
		o.Network = n
	}
}

// Address：server 监听地址
func Address(a string) Option {
	return func(o *Options) {
		o.Address = a
	}
}

// NumLoops：work eventloop 的数量
func NumLoops(n int) Option {
	return func(o *Options) {
		o.NumLoops = n
	}
}

// Protocol：数据包处理
func Protocol(p connection.Protocol) Option {
	return func(o *Options) {
		o.Protocol = p
	}
}

// IdleTime：最大空闲时间（秒）
func IdleTime(t time.Duration) Option {
	return func(o *Options) {
		o.IdleTime = t
	}
}
