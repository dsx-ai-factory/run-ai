package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var AccessRulesRead = models.PermissionSet{
	Name:        "Access rules read",
	Description: "Grants permissions to view access rules, including all related data and resources associated with them.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_ACCESS_RULES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ROLES,
			Read:         true,
		},
	},
}
