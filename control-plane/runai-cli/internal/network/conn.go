package network

import "net"

// Conn is an interface that extends net.Conn for now only for mocking purposes
type Conn interface {
	net.Conn
}
