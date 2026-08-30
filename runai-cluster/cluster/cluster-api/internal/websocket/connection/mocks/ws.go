package mocks

import (
	"net"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/connection"
)

type MockWebSocket struct {
	net.Conn
}

func NewMockWebSocket() *MockWebSocket {
	conn, _ := net.Pipe()
	return &MockWebSocket{conn}
}

func (mws *MockWebSocket) WriteText(p []byte) error {
	return nil
}

func (mws *MockWebSocket) SetMessageHandler(_ connection.MessageHandler) {
}
