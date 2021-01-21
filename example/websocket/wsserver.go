package main

import (
	"github.com/Dongxiem/fastnet"
	"github.com/Dongxiem/fastnet/plugins/websocket"
	"github.com/Dongxiem/fastnet/plugins/websocket/ws"
)

// NewWebSocketServer 创建 WebSocket Server
func NewWebSocketServer(handler websocket.WSHandler, u *ws.Upgrader, opts ...gfaio.Option) (server *gfaio.Server, err error) {
	opts = append(opts, gfaio.Protocol(websocket.New(u)))
	return gfaio.NewServer(websocket.NewHandlerWrap(u, handler), opts...)
}
