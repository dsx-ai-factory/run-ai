package websocket

import "io"

// ReadWriteCloser this for testing purposes, we can mock this interface
type ReadWriteCloser interface {
	io.Reader
	io.Writer
	io.Closer
}

// ReadCloser this for testing purposes, we can mock this interface
type ReadCloser interface {
	io.Reader
	io.Closer
}

// WriteCloser this for testing purposes, we can mock this interface
type WriteCloser interface {
	io.Writer
	io.Closer
}
