package tools

import (
	"testing"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/stretchr/testify/assert"
)

func TestToolsLint(t *testing.T) {
	errs := runaimcp.Lint(func(r *runaimcp.Registry) {
		Register(r, nil)
	})
	assert.Empty(t, errs)
}
