package enrichers

import (
	"context"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
)

const (
	StaticEnricherName = "static"
)

type StaticEnricherConfig map[string]string

type StaticEnricher struct {
	Config *StaticEnricherConfig
}

func NewStaticEnricher(config *StaticEnricherConfig) (*StaticEnricher, error) {
	return &StaticEnricher{
		Config: config,
	}, nil
}

func (se *StaticEnricher) Enrich(ctx context.Context, event *events.Event) error {
	for key, val := range *se.Config {
		if val != "" {
			event.Properties[key] = val
		}
	}
	return nil
}
