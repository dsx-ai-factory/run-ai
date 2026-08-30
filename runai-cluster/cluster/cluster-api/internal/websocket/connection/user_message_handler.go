package connection

import (
	"nhooyr.io/websocket"
)

type ClientHandler struct {
	base BaseMessageHandler
}

func (ch *ClientHandler) Process(msgType websocket.MessageType, msgRead []byte, msgBuffer []byte) (int, error) {
	statusMsg, bytesRead, err := ch.base.ProcessCommon(msgType, msgRead, msgBuffer)
	if err != nil {
		return 0, err
	}

	// If it's a binary message, we're done
	if statusMsg == nil {
		return bytesRead, nil
	}

	return bytesRead, nil
}

func (ch *ClientHandler) Close() error {
	return nil
}
