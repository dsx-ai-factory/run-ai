package websocket

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/metrics"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/connection"

	"github.com/rs/zerolog/log"

	"nhooyr.io/websocket"
)

// ExtendedNetConnection this is a wrapper around the websocket connection to use it as a net.Conn
// this is useful for the proxy function
// this is warped the third party library websocket.netConn In case we want to change the library in the future
type ExtendedNetConnection struct {
	InnerConnection    net.Conn
	InnerWebSocketConn *websocket.Conn
	Cancel             context.CancelFunc
	ctx                context.Context
	readSize           atomic.Uint64
	writeSize          atomic.Uint64
	messageHandler     connection.MessageHandler
}

func NewConnection(conn *websocket.Conn, ctx context.Context) *ExtendedNetConnection {
	ctx, cancel := context.WithCancel(ctx)
	innerConnection := websocket.NetConn(ctx, conn, websocket.MessageBinary)

	return &ExtendedNetConnection{
		InnerConnection:    innerConnection,
		InnerWebSocketConn: conn,
		Cancel:             cancel,
		ctx:                ctx,
	}
}

func (e *ExtendedNetConnection) Read(p []byte) (n int, err error) {
	readLen := uint64(len(p))
	e.readSize.Add(readLen)

	if readLen > 0 {
		metrics.MessageSentReport(metrics.FromContext(e.ctx), readLen)
	}

	if e.messageHandler == nil {
		return e.InnerConnection.Read(p)
	}

	msgType, msgRead, err := e.InnerWebSocketConn.Read(e.ctx)
	if err != nil {
		return 0, err
	}

	return e.messageHandler.Process(msgType, msgRead, p)
}

func (e *ExtendedNetConnection) Write(p []byte) (n int, err error) {
	writeLen := uint64(len(p))
	e.writeSize.Add(writeLen)

	if writeLen > 0 {
		metrics.MessageReceivedReport(metrics.FromContext(e.ctx), writeLen)
	}

	return e.InnerConnection.Write(p)
}

func (e *ExtendedNetConnection) WriteText(p []byte) error {
	return e.InnerWebSocketConn.Write(e.ctx, websocket.MessageText, p)
}

func (e *ExtendedNetConnection) Close() error {
	if e.messageHandler != nil {
		_ = e.messageHandler.Close()
	}

	err := e.InnerWebSocketConn.Close(websocket.StatusNormalClosure, "")
	defer e.Cancel()

	if IgnoreClosedError(err) != nil {
		log.Warn().Err(err).Msg("failed to close websocket connection")
		log.Debug().Msg("force closing the connection")

		err = e.InnerWebSocketConn.CloseNow()
		if IgnoreClosedError(err) != nil {
			return err
		}
	}

	return nil
}

func (e *ExtendedNetConnection) LocalAddr() net.Addr {
	return e.InnerConnection.LocalAddr()
}

func (e *ExtendedNetConnection) RemoteAddr() net.Addr {
	return e.InnerConnection.RemoteAddr()
}

func (e *ExtendedNetConnection) SetDeadline(t time.Time) error {
	return e.InnerConnection.SetDeadline(t)
}

func (e *ExtendedNetConnection) SetReadDeadline(t time.Time) error {
	return e.InnerConnection.SetReadDeadline(t)
}

func (e *ExtendedNetConnection) SetWriteDeadline(t time.Time) error {
	return e.InnerConnection.SetWriteDeadline(t)
}

func (e *ExtendedNetConnection) ReadSize() uint64 {
	return e.readSize.Swap(0)
}

func (e *ExtendedNetConnection) WriteSize() uint64 {
	return e.writeSize.Swap(0)
}

func (e *ExtendedNetConnection) SetMessageHandler(messageHandler connection.MessageHandler) {
	e.messageHandler = messageHandler
}

func IgnoreClosedError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, net.ErrClosed) {
		return nil
	}

	return err
}

func SetClientMessageHandler(conn connection.WebSocket) {
	enc, ok := conn.(*ExtendedNetConnection)
	if !ok {
		// should not reach this section otherwise in tests
		log.Warn().Msg("cannot convert to ExtendedNetConnection")
		return
	}

	log.Debug().Msgf("enbale client message handler")
	enc.SetMessageHandler(&connection.ClientHandler{})
}
