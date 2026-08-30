package filter

import (
	"context"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
)

type Filter interface {
	// Filter - should return true if event dropped the filter
	Filter(ctx context.Context, event *events.Event) (bool, error)
}

type Factory interface {
	CreateFilter(fc *FilterConfig) (Filter, error)
}
