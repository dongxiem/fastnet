package main

import (
	"github.com/Dongxiem/fastnet"
	"github.com/Dongxiem/fastnet/plugins/websocket"
	"github.com/Dongxiem/fastnet/plugins/websocket/ws"
)

// NewWebSocketServer 创建 WebSocket Server
func NewWebSocketServer(handler websocket.WSHandler, u *ws.Upgrader, opts ...fastnet.Option) (server *fastnet.Server, err error) {
	opts = append(opts, fastnet.Protocol(websocket.New(u)))
	return fastnet.NewServer(websocket.NewHandlerWrap(u, handler), opts...)
}
