package notify

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
)

func NewNotifyingHandler(configProvider NotifyingConfigProvider, sinkConfigProvider ConfigProvider) events.EventHandler {
	return &NotifyingEventHandler{
		ConfigProvider:     configProvider,
		SinkConfigProvider: sinkConfigProvider,
		Notifier:           NewEventNotifier(),
		SinkFactory:        NewSinkFactory(),
	}
}

type NotifyingEventHandler struct {
	ConfigProvider     NotifyingConfigProvider
	SinkConfigProvider ConfigProvider
	Notifier           Notifier
	SinkFactory        Factory
}

func (nh NotifyingEventHandler) Handle(ctx context.Context, event *events.Event) error {
	log.Debug().Msgf("event %s is in notify handler", event.UID)

	sinks, err := nh.SinkFactory.CreateSinks(nh.SinkConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create sinks %w", err)
	}

	for _, sink := range sinks {
		if err := sink.Sink(event); err != nil {
			return fmt.Errorf("failed to sink event with sink %s: %w", sink.Type(), err)
		}
		log.Debug().Msgf("event %s triggered notify with sink %s", event.UID, sink.Type())
	}
	return nil
}
