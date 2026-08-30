package base

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	clusterErrors "github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/errors"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/status"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/websocket"
)

// StreamBase provides common streaming functionality with retry logic
type StreamBase struct {
	handler websocket.Handler
}

// NewStreamBase creates a new StreamBase with the given configuration
func NewStreamBase(dialer websocket.Dialer, baseUrl string, maxRetryDuration, timeBetweenTries time.Duration) *StreamBase {
	return &StreamBase{
		handler: websocket.Handler{
			Dialer:           dialer,
			BaseUrl:          baseUrl,
			MaxRetryDuration: maxRetryDuration,
			TimeBetweenTries: timeBetweenTries,
		},
	}
}

// ConnectWithRetries performs the streaming operation with retry logic
func (sb *StreamBase) ConnectWithRetries(ctx context.Context, streamUrl *url.URL, header *http.Header) (io.ReadWriteCloser, error) {
	var err error
	startTime := time.Now()

	for {
		remote, response, innerError := sb.handler.Dialer.Connect(ctx, streamUrl, header)
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		if innerError == nil {
			return remote, nil
		}

		err = clusterErrors.HandleError(response, innerError)
		if clusterErrors.IgnoreNotReadyError(err) != nil {
			return nil, err
		}

		if time.Since(startTime) >= sb.handler.MaxRetryDuration {
			break
		}

		time.Sleep(sb.handler.TimeBetweenTries)
	}

	return nil, err
}

// GetHandler returns the websocket handler for access to its fields
func (sb *StreamBase) GetHandler() websocket.Handler {
	return sb.handler
}

// SetupMessageHandler sets up the appropriate message handler for the connection based on TTY and metadata
func (sb *StreamBase) SetupMessageHandler(enc *websocket.ExtendedNetConnection, clusterMetadata *status.Metadata, handler websocket.MessageHandler) (io.ReadWriteCloser, error) {
	// backward compatibility
	if clusterMetadata == nil || handler == nil {
		enc.SetMessageHandler(&websocket.Keepalive{})
		return enc, nil
	}

	// handler exist
	enc.SetMessageHandler(handler)
	return enc, nil
}

func IsMessageHandlerCapable(remote io.ReadWriteCloser, clusterMetadata *status.Metadata) bool {
	if _, ok := remote.(*websocket.ExtendedNetConnection); !ok {
		return false
	}

	if clusterMetadata == nil || clusterMetadata.Features.MessageEnabled {
		return false
	}

	return true
}

func SetMessageKeepAlive(remote io.ReadWriteCloser) {
	enc := remote.(*websocket.ExtendedNetConnection)
	enc.SetMessageHandler(&websocket.Keepalive{})
}
