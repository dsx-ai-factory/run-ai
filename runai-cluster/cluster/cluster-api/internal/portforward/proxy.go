package portforward

import (
	"io"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/keepalive"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/proxy"
)

type ProxyFactoryFn func(
	remote io.ReadWriteCloser, local io.ReadWriteCloser, keepalive keepalive.Keepalive,
) proxy.ProxyHandler

func NewTwoWayProxyFactory() ProxyFactoryFn {
	return func(remote io.ReadWriteCloser, local io.ReadWriteCloser, keepalive keepalive.Keepalive) proxy.ProxyHandler {
		return proxy.NewTwoWayProxy(remote, local, keepalive)
	}
}
