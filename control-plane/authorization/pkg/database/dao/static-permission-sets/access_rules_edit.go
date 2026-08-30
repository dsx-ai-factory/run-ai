package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var AccessRulesEdit = models.PermissionSet{
	Name:        "Access rules edit",
	Description: "Grants full permissions to create, read, update, and delete access rules, and to access all related data and resources required for their management.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_ACCESS_RULES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_PROJECT,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DEPARTMENT,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ROLES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS_MINIMAL,
			Read:         true,
		},
	},
}
