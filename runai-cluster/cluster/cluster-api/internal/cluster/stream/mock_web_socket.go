package stream

import (
	"net"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/connection"
)

// MockWebSocket is a mock implementation of the WebSocket interface for testing
type MockWebSocket struct {
	messages [][]byte
	index    int
}

func (m *MockWebSocket) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (m *MockWebSocket) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *MockWebSocket) Close() error {
	return nil
}

func (m *MockWebSocket) LocalAddr() net.Addr {
	const port = 8080
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: port}
}

func (m *MockWebSocket) RemoteAddr() net.Addr {
	const port = 8081
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: port}
}

func (m *MockWebSocket) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockWebSocket) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockWebSocket) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *MockWebSocket) WriteText(p []byte) error {
	return nil
}

func (m *MockWebSocket) ReadMessage() (messageType int, data []byte, err error) {
	if m.index >= len(m.messages) {
		return 0, nil, nil
	}

	data = m.messages[m.index]
	m.index++

	return 1, data, nil // 1 = text message
}

func (m *MockWebSocket) SetMessageHandler(_ connection.MessageHandler) {
}
