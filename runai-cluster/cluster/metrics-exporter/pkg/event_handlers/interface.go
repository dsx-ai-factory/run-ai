package event_handlers

import "context"

type ObjectIdentifier struct {
	Name      string
	Namespace string
}

type EventHandler interface {
	ResetMetrics(identifier ObjectIdentifier)
	SetMetrics(ctx context.Context, eventObject interface{}) error
}
