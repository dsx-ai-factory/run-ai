package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var NotificationSettingsRead = models.PermissionSet{
	Name:        "Notification settings read",
	Description: "Grants permissions to view notification settings.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_NOTIFICATION_SETTINGS,
			Read:         true,
		},
	},
}
