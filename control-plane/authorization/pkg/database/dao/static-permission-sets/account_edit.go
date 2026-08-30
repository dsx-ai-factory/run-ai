package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var AccountEdit = models.PermissionSet{
	Name:        "Account edit",
	Description: "Grants permissions to create, read, update, and delete accounts.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_TENANT,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
	},
}
