package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var WorkloadRead = models.PermissionSet{
	Name:        "Workload read",
	Description: "Grants permissions to view workloads, including all related data and resources associated with them.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_WORKLOADS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKSPACES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_TRAININGS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_INFERENCES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKLOAD_PROPERTIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_PROJECT,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
	},
}
