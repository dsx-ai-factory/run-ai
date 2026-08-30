package orchestrator2

import (
	"context"
	"fmt"
	"strings"
)

type PerformAction interface {
	Kind() string
	Args() map[string]any
	Describe() string
	Execute(ctx context.Context, deps Deps) (ActionResult, error)
}

func describeActions(actions []PerformAction) string {
	lines := make([]string, 0, len(actions))
	for _, a := range actions {
		lines = append(lines, fmt.Sprintf("%s: %s", a.Kind(), a.Describe()))
	}
	return strings.Join(lines, "\n")
}
