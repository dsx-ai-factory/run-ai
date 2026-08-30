package orgunits

import (
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	registerListProjects(r, client)
	registerListDepartments(r, client)
	registerManageOrgUnit(r, client)
}
