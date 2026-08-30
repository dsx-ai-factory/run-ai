package websocket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/network"
	"nhooyr.io/websocket"
)

type TransportDialer struct {
	token string
}

func NewTransportDialer(token string) *TransportDialer {
	return &TransportDialer{token: token}
}

func (td *TransportDialer) Connect(ctx context.Context, url *url.URL, header *http.Header) (network.Conn, *http.Response, error) {
	if header == nil {
		header = &http.Header{}
	}
	header.Set(config.AuthorizationHeader, fmt.Sprintf(config.BearerTokenStr, td.token))

	connection, reg, err := websocket.Dial(ctx, url.String(), &websocket.DialOptions{HTTPHeader: *header})
	if err != nil {
		return nil, reg, err
	}
	transport := NewConnection(connection, ctx)
	return transport, reg, nil
}
