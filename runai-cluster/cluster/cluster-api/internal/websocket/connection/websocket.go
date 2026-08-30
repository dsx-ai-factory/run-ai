package connection

import (
	"net"
)

type WebSocket interface {
	net.Conn
	WriteText(p []byte) error
	SetMessageHandler(messageHandler MessageHandler)
}
