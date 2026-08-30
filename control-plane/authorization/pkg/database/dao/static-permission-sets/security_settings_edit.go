package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var SecuritySettingsEdit = models.PermissionSet{
	Name:        "Security settings edit",
	Description: "Grants permissions to create, read, update, and delete security settings.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_SECURITY_SETTINGS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
	},
}
