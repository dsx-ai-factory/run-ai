package aiapplications

import (
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	registerListAIApplications(r, client)
	registerSubmitAIApplication(r, client)
}
