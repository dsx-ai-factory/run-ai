package filters

import (
	"context"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
)

type DummyFilter struct{}

func (df DummyFilter) Filter(ctx context.Context, event *events.Event) (bool, error) {
	return false, nil
}
