package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var InferenceEdit = models.PermissionSet{
	Name:        "Inference edit",
	Description: "Grants full permissions to create, read, update, and delete inference workloads, and to access all related data and resources required for their management.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_WORKLOADS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_INFERENCES,
			Create:       true,
			Update:       true,
			Delete:       true,
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
			ResourceType: api.RESOURCETYPE_POLICIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_TEMPLATES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ENVIRONMENTS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_COMPUTE_RESOURCES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_PVC_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_GIT_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_HOST_PATH_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NFS_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_S3_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DATAVOLUMES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_REGISTRIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CREDENTIALS,
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
		{
			ResourceType: api.RESOURCETYPE_USERS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NETWORK_TOPOLOGIES,
			Read:         true,
		},
	},
}
