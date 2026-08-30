package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var ClusterEdit = models.PermissionSet{
	Name:        "Cluster edit",
	Description: "Grants permissions to create, read, update, and delete clusters.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_CLUSTER,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
	},
}
