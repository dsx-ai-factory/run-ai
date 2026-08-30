package settings

import (
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	registerListSystemSettings(r, client)
	registerUpdateSystemSettings(r, client)
}
