package connection

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/meta"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/stream/resize"
	"k8s.io/client-go/tools/remotecommand"

	"nhooyr.io/websocket"
)

type InteractiveHandler struct {
	base              BaseMessageHandler
	conn              WebSocket
	terminalSizeQueue *resize.TerminalSizeQueue
	id                atomic.Int32
}

func NewInteractiveHandler(conn WebSocket, terminalSizeQueue *resize.TerminalSizeQueue) *InteractiveHandler {
	return &InteractiveHandler{conn: conn, terminalSizeQueue: terminalSizeQueue}
}

func (ih *InteractiveHandler) Process(msgType websocket.MessageType, msgRead []byte, msgBuf []byte) (int, error) {
	statusMsg, bytesRead, err := ih.base.ProcessCommon(msgType, msgRead, msgBuf)
	if err != nil {
		return 0, err
	}

	// If it's a binary message, we're done
	if statusMsg == nil {
		return bytesRead, nil
	}

	msgErr := ih.handleMessage(statusMsg)
	if msgErr != nil {
		log.Warn().Msgf("handle message error: %v", msgErr)
		ih.handleError(msgErr)
	}
	// must return the message length to acknowledge the message, otherwise the message will not be removed from the queue.
	return len(msgRead), nil
}

func (ih *InteractiveHandler) handleMessage(message *Message) *MessageError {
	if message.Type != TypeTerminalSize {
		return &MessageError{
			error: fmt.Errorf("unsupported message, id: %d type: %s", message.Id, string(message.Type)),
			Id:    message.Id,
		}
	}

	messageError := ih.handleTerminalSize(message)
	if messageError != nil {
		return messageError
	}

	return nil
}

func (ih *InteractiveHandler) handleTerminalSize(message *Message) *MessageError {
	width := message.GetInt("width", 0)
	height := message.GetInt("height", 0)

	if width <= 0 || height <= 0 {
		return &MessageError{
			error: fmt.Errorf("window size must be positive number, width: %d, height: %d",
				width, height),
			Id: message.Id,
		}
	}

	ih.terminalSizeQueue.Resize(uint16(width), uint16(height))

	return nil
}

func (ih *InteractiveHandler) Close() error {
	ih.terminalSizeQueue.Stop()
	return nil
}

func (ih *InteractiveHandler) Next() *remotecommand.TerminalSize {
	next := ih.terminalSizeQueue.Next()

	err := ih.sendResizeUpdate(next)
	if err != nil {
		log.Warn().Msgf("error sending CLI update, error: %v", err)
	}

	return next
}

func (ih *InteractiveHandler) sendResizeUpdate(next *remotecommand.TerminalSize) error {
	message, err := json.Marshal(next)
	if err != nil {
		return err
	}

	response := Message{
		Version:    meta.GetVersion(),
		Id:         int(ih.id.Add(1)),
		StatusCode: http.StatusOK,
		Message:    string(message),
		Type:       TypeTerminalSize,
	}

	p, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return ih.conn.WriteText(p)
}

func (ih *InteractiveHandler) handleError(errMsg *MessageError) {
	response := Message{
		Version:    meta.GetVersion(),
		Id:         errMsg.Id,
		StatusCode: http.StatusBadRequest,
		Message:    errMsg.Error(),
		Type:       TypeTerminalSize,
	}

	p, err := json.Marshal(response)
	if err != nil {
		log.Warn().Msgf("error marshaling error message, error: %v", err)
	}

	err = ih.conn.WriteText(p)
	if err != nil {
		log.Warn().Msgf("error sending error message, error: %v", err)
	}
}
