package events

import "github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"

type sink interface {
	Handle(e orchestrator2.Event) error
	Close() error
}
