package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var SettingsRead = models.PermissionSet{
	Name:        "Settings read",
	Description: "Grants permissions to view settings.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_SETTINGS,
			Read:         true,
		},
	},
}
