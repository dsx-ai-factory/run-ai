package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var ServiceAccountsRead = models.PermissionSet{
	Name:        "Service accounts read",
	Description: "Grants permissions to view service accounts.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_APPS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_SERVICE_ACCOUNT,
			Read:         true,
		},
	},
}
