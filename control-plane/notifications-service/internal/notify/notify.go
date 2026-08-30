package notify

import (
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks"
)

type Factory interface {
	CreateSinks(configProvider ConfigProvider) ([]sinks.Sinker, error)
}

type Notifier interface {
	Register(sinks.Sinker)
	Notify(events.Event) error
}

type ConfigProvider interface {
	GetSinks(id ...string) (map[string]*sinks.SinkConfig, error)
}
