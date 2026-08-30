package connection

import (
	"nhooyr.io/websocket"
)

type MessageHandler interface {
	Process(msgType websocket.MessageType, msgRead []byte, msgBuf []byte) (int, error)
	Close() error
}
