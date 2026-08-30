package listen

import (
	"context"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
)

type Factory interface {
	CreateListener(id string, eventCh chan *events.Event, configProvider ListeningConfigProvider) (Listener, error)
}

type Listener interface {
	Id() string
	Listen(ctx context.Context)
}
