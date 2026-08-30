package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var EnvironmentsOnlyRead = models.PermissionSet{
	Name:        "Environments only read",
	Description: "Grants permissions to view environments.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_ENVIRONMENTS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS_MINIMAL,
			Read:         true,
		},
	},
}
