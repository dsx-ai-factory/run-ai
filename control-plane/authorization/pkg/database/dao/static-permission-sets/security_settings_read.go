package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var SecuritySettingsRead = models.PermissionSet{
	Name:        "Security settings read",
	Description: "Grants permissions to view security settings.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_SECURITY_SETTINGS,
			Read:         true,
		},
	},
}
