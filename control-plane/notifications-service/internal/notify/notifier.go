package notify

import (
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks"
)

func NewEventNotifier() Notifier {
	return &EventNotifier{
		sinks: []sinks.Sinker{},
	}
}

type EventNotifier struct {
	sinks []sinks.Sinker
}

func (en *EventNotifier) Register(sink sinks.Sinker) {
	en.sinks = append(en.sinks, sink)
}

func (en *EventNotifier) Notify(events.Event) error {
	return nil
}
