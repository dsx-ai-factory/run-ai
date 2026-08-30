package websocket

import (
	"context"
	"errors"
	"net"
	"time"

	"nhooyr.io/websocket"
)

//TODO: remove this after moving to mono repo

// ExtendedNetConnection this is a wrapper around the websocket connection to use it as a net.Conn
// this is useful for the proxy function
// this is warped the third party library websocket.netConn In case we want to change the library in the future
type ExtendedNetConnection struct {
	InnerConnection    net.Conn
	InnerWebSocketConn *websocket.Conn
	Cancel             context.CancelFunc
	ctx                context.Context
	messageHandler     MessageHandler
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
	// If no message handler exist, just read
	if e.messageHandler == nil {
		return e.InnerConnection.Read(p)
	}

	// reading from the websocket message
	msgType, msgRead, errRead := e.InnerWebSocketConn.Read(e.ctx)
	if errRead != nil {
		return 0, errRead
	}

	return e.messageHandler.Process(msgType, msgRead, p)
}

func (e *ExtendedNetConnection) Write(p []byte) (n int, err error) {
	return e.InnerConnection.Write(p)
}

func (e *ExtendedNetConnection) Close() error {
	if e.messageHandler != nil {
		_ = e.messageHandler.Close()
	}
	err := e.InnerWebSocketConn.Close(websocket.StatusNormalClosure, "")
	defer e.Cancel()
	if IgnoreClosedError(err) != nil {
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

func (e *ExtendedNetConnection) SetMessageHandler(messageHandler MessageHandler) {
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
