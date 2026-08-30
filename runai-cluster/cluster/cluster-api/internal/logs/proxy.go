package logs

import (
	"io"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/keepalive"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/proxy"
)

type ProxyFactoryFn func(
	remote io.ReadCloser, local io.ReadWriteCloser, keepalive keepalive.Keepalive,
) proxy.ProxyHandler

func NewOneWayProxyFactory() ProxyFactoryFn {
	return func(remote io.ReadCloser, local io.ReadWriteCloser, keepalive keepalive.Keepalive) proxy.ProxyHandler {
		return proxy.NewOneWayProxy(remote, local, keepalive)
	}
}
