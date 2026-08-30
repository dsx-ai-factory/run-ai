package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var UsersMinimalRead = models.PermissionSet{
	Name:        "Users minimal read",
	Description: "Grants permissions to view minimal users data.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_USERS_MINIMAL,
			Read:         true,
		},
	},
}
