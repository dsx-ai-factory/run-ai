package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var RolesEdit = models.PermissionSet{
	Name:        "Roles edit",
	Description: "Grants permissions to create, read, update, and delete roles.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_ROLES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
	},
}
