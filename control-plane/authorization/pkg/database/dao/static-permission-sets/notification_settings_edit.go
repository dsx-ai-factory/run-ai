package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var NotificationSettingsEdit = models.PermissionSet{
	Name:        "Notification settings edit",
	Description: "Grants permissions to create, read, update, and delete notification settings.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_NOTIFICATION_SETTINGS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
	},
}
