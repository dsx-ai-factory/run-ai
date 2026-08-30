package static_legacy_roles

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var ResearchManager = models.Role{
	Name: "Research manager",
	Description: "A research manager can do the following within their assigned scope:\n" +
		"Create, view, edit, delete: environments, data sources, compute resources, templates, datavolumes\n" +
		"View: projects, related jobs and workspaces, dashboards, storage class configurations",
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
			ResourceType: api.RESOURCETYPE_PROJECT,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_ACCESS_RULES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_ROLES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DEPARTMENT,
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
			ResourceType: api.RESOURCETYPE_WORKLOADS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKLOAD_PROPERTIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKSPACES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_TRAININGS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_INFERENCES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CREDENTIALS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ENVIRONMENTS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_HOST_PATH_ASSETS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_NFS_ASSETS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_PVC_ASSETS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_CM_VOLUME_ASSETS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_SECRET_VOLUME_ASSETS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_GIT_ASSETS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_S3_ASSETS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_COMPUTE_RESOURCES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_TEMPLATES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_REGISTRIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_POLICIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DATAVOLUMES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
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
			ResourceType: api.RESOURCETYPE_USERS_MINIMAL,
			Read:         true,
		},
	},
}
