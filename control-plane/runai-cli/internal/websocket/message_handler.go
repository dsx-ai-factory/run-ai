package websocket

import "nhooyr.io/websocket"

type MessageHandler interface {
	Process(msgType websocket.MessageType, msgRead []byte, msgBuffer []byte) (int, error)
	Close() error
}
