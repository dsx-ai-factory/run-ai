package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var ClusterConfigRead = models.PermissionSet{
	Name:        "Cluster config read",
	Description: "Grants permissions to view cluster configuration.",
	Internal:    true,
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_CLUSTER_CONFIG,
			Read:         true,
		},
	},
}
