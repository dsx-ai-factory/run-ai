package events

import (
	"errors"
	"fmt"
	"os"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/terminal"
)

// Manager is not safe for concurrent use. HandleEvent and GetEvents must be
// called from a single goroutine (the orchestrator emits sequentially).
type Manager struct {
	sinks  []sink
	events []orchestrator2.Event
}

func New(cfg Config) (*Manager, error) {
	m := &Manager{}

	switch cfg.Format {
	case spec.OutputFormatNone:
	case spec.OutputFormatJSON:
		m.sinks = append(m.sinks, newNDJSONSink(os.Stdout, nil))
	case spec.OutputFormatText:
		term := cfg.Terminal
		if term == nil {
			term = terminal.New()
		}
		m.sinks = append(m.sinks, newTextSink(term))
	default:
		return nil, fmt.Errorf("events: unknown output format %q", cfg.Format)
	}

	if cfg.EventsFile != "" {
		f, err := os.Create(cfg.EventsFile)
		if err != nil {
			_ = m.Close()
			return nil, fmt.Errorf("events: open events file %q: %w", cfg.EventsFile, err)
		}
		m.sinks = append(m.sinks, newNDJSONSink(f, f))
	}

	return m, nil
}

func (m *Manager) HandleEvent(e orchestrator2.Event) error {
	m.events = append(m.events, e)
	for _, s := range m.sinks {
		if err := s.Handle(e); err != nil {
			return err
		}
	}
	return nil
}

// progressReporter is implemented by sinks that render live progress lines.
type progressReporter interface {
	Progress(lines []string)
}

// Progress forwards live progress lines to sinks that render them. It records
// nothing and is safe to call from a separate goroutine.
func (m *Manager) Progress(lines []string) {
	for _, s := range m.sinks {
		if pr, ok := s.(progressReporter); ok {
			pr.Progress(lines)
		}
	}
}

func (m *Manager) GetEvents() []orchestrator2.Event {
	out := make([]orchestrator2.Event, len(m.events))
	copy(out, m.events)
	return out
}

func (m *Manager) Close() error {
	var errs []error
	for _, s := range m.sinks {
		if err := s.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
