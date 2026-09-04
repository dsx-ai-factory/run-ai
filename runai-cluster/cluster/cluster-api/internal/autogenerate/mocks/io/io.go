package io

import (
	"context"
	"net"
	"time"
)

// Mock implementations
type MockReadCloser struct{}

func (m *MockReadCloser) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (m *MockReadCloser) Close() error {
	return nil
}

type MockConn struct{}

func (m *MockConn) Close() error {
	return nil
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return nil
}

func (m *MockConn) RemoteAddr() net.Addr {
	return nil
}

func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type MockProxy struct{}

func (m *MockProxy) Start(ctx context.Context) error {
	return nil
}
