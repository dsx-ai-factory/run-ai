package events

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/terminal"
)

type Config struct {
	Format     spec.OutputFormat
	EventsFile string

	// Terminal renders the text sink. When Format == OutputFormatText and
	// Terminal is nil, the manager constructs a default one writing to
	// stdout/stderr.
	Terminal *terminal.Terminal
}

func ConfigFromSpec(s *spec.Spec, t *terminal.Terminal) Config {
	return Config{
		Format:     s.Output.Format,
		EventsFile: s.Output.EventsFile,
		Terminal:   t,
	}
}
