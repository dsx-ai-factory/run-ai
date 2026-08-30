package websocket

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/connection"

	"nhooyr.io/websocket"
)

const (
	ProtocolBearer = "Bearer"
)

var SupportedProtocols = []string{ProtocolBearer}

type Upgrader interface {
	Upgrade(ctx context.Context, w http.ResponseWriter, r *http.Request) (connection.WebSocket, error)
}

type UpgradeConnection struct {
	options websocket.AcceptOptions
}

func NewUpgradeConnection() *UpgradeConnection {
	// we are using insecure for preventing CORS blocking, for example if the CP
	// and the cluster be in a different domains, the call will be blocked at the UI.
	return &UpgradeConnection{
		options: websocket.AcceptOptions{Subprotocols: SupportedProtocols, InsecureSkipVerify: true},
	}
}

func (u *UpgradeConnection) Upgrade(
	ctx context.Context, w http.ResponseWriter, r *http.Request,
) (connection.WebSocket, error) {
	conn, err := websocket.Accept(w, r, &u.options)
	if err != nil {
		return nil, err
	}

	return NewConnection(conn, ctx), nil
}
