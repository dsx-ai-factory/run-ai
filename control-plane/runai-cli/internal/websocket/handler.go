package websocket

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/network"
)

type Dialer interface {
	Connect(ctx context.Context, url *url.URL, header *http.Header) (network.Conn, *http.Response, error)
}

type Handler struct {
	Dialer           Dialer
	BaseUrl          string
	MaxRetryDuration time.Duration
	TimeBetweenTries time.Duration
}
