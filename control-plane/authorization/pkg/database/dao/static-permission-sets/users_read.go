package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var UsersRead = models.PermissionSet{
	Name:        "Users read",
	Description: "Grants permissions to view users.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_USERS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_USERS_MINIMAL,
			Read:         true,
		},
	},
}
