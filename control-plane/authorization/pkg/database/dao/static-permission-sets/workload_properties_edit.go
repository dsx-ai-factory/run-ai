package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var WorkloadPropertiesEdit = models.PermissionSet{
	Name:        "Workload properties edit",
	Description: "Grants full permissions to create, read, update, and delete workload properties, and to access all related data and resources required for their management.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_WORKLOAD_PROPERTIES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
	},
}
