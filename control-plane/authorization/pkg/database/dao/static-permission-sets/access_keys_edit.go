package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var AccessKeysEdit = models.PermissionSet{
	Name:        "Access keys edit",
	Description: "Grants full permissions to create, read, update, and delete user access keys.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_ACCESS_KEYS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
	},
}
