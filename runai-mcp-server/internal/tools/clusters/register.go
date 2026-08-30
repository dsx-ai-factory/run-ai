package clusters

import (
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

// Register wires the cluster-service-backed tools.
func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	registerListNodePools(r, client)
	registerManageNodePool(r, client)
	registerGetClusterPhysicalInventory(r, client)
	registerGetClusterInfrastructureHealth(r, client)
}
