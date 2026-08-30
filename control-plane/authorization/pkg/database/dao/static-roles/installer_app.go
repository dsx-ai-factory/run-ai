package static_legacy_roles

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	static_permission_sets "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var InstallerApp = models.Role{
	Name:                    "Installer application",
	Description:             "An internal installer app used to create clusters and fetch its control-plane configuration (currently used during the installation process)",
	InternalOnly:            true,
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	ClusterRolePredefinedId: nil,
	PermissionSets: &[]models.PermissionSet{
		{Name: static_permission_sets.InstallerAppPS.Name},
		{Name: static_permission_sets.UsersMinimalRead.Name},
	},
	Permissions: []*models.Permission{
		{
			ResourceType: api.RESOURCETYPE_TENANT,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTER_CONFIG,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTER,
			Create:       true,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
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
			ResourceType: api.RESOURCETYPE_SETTINGS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_USERS_MINIMAL,
			Read:         true,
		},
	},
}
