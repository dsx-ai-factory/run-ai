package static_legacy_roles

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var MLEngineer = models.Role{
	Name: "ML engineer",
	Description: "An ML engineer can do the following within their assigned scope:\n" +
		"Create, edit, view, delete: deployments, credentials\n" +
		"View: departments, projects, clusters, node-pools, nodes, dashboards, datavolumes, storage class configurations",
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	ClusterRolePredefinedId: nil,
	Deprecated:              true,
	Permissions: []*models.Permission{
		{
			ResourceType: api.RESOURCETYPE_TENANT,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DEPARTMENT,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_PROJECT,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_INFERENCES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKLOADS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKLOAD_PROPERTIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKSPACES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTER,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CREDENTIALS,
			Read:         true,
			Create:       true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_POLICIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DATAVOLUMES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_STORAGE_CLASS_CONFIGURATION,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_SETTINGS,
			Read:         true,
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
			ResourceType: api.RESOURCETYPE_COMPUTE_RESOURCES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_TEMPLATES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_REGISTRIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_GIT_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_SECRET_VOLUME_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CM_VOLUME_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_PVC_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NFS_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_HOST_PATH_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ENVIRONMENTS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NETWORK_TOPOLOGIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_USERS_MINIMAL,
			Read:         true,
		},
	},
}
