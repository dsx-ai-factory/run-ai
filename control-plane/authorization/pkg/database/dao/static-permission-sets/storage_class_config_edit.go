package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var StorageClassConfigEdit = models.PermissionSet{
	Name:        "Storage class configuration edit",
	Description: "Grants full permissions to create, read, update, and delete storage class configurations used by volume and PVC assets.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_STORAGE_CLASS_CONFIGURATION,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
	},
}
