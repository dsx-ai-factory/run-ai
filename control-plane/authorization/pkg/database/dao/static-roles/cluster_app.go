package static_legacy_roles

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	static_permission_sets "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var ClusterApp = models.Role{
	Name:                    "Cluster application",
	Description:             "An internal cluster app used to sync the data from the cluster to the control plane and in the opposite way.",
	InternalOnly:            true,
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	ClusterRolePredefinedId: nil,
	PermissionSets: &[]models.PermissionSet{
		{Name: static_permission_sets.ClusterAppPS.Name}, {Name: static_permission_sets.UsersMinimalRead.Name},
	},
	Permissions: []*models.Permission{
		{
			ResourceType: api.RESOURCETYPE_TENANT,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DEPARTMENT,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_PROJECT,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKLOADS,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_APPLICATIONS,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKLOAD_PROPERTIES,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_USERS,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_EVENTS_HISTORY,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTER,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODES,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ROLES,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ACCESS_RULES,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKSPACES,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_TRAININGS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_INFERENCES,
			Read:         true,
			Delete:       true,
			Sync:         true,
			Update:       true,
			Create:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_CREDENTIALS,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_PVC_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CM_VOLUME_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_SECRET_VOLUME_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_GIT_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_HOST_PATH_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NFS_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_S3_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_POLICIES,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DATAVOLUMES,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTER_CONFIG,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_SECURITY_SETTINGS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NOTIFICATION_SETTINGS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_BRANDING_SETTINGS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_SETTINGS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NETWORK_TOPOLOGIES,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_USERS_MINIMAL,
			Read:         true,
		},
	},
}
