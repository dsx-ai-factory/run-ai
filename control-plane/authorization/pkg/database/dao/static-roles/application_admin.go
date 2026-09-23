/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_legacy_roles

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	static_permission_sets "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var ApplicationAdmin = models.Role{
	Name: "Application administrator",
	Description: "An application administrator can:\n" +
		"Manage: departments, projects.\n" +
		"Assign: roles (department administrator, researcher, ML engineer, research manager, viewer).\n" +
		"View: users, dashboards (including Consumption), nodepools.",
	RoleType:                models.DGXC,
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	ClusterRolePredefinedId: nil,
	PermissionSets: &[]models.PermissionSet{
		{Name: static_permission_sets.WorkspaceEdit.Name},
		{Name: static_permission_sets.TrainingEdit.Name},
		{Name: static_permission_sets.InferenceEdit.Name},
		{Name: static_permission_sets.ApplicationEdit.Name},
		{Name: static_permission_sets.YamlWorkloadEdit.Name},
		{Name: static_permission_sets.WorkloadRead.Name},
		{Name: static_permission_sets.WorkloadPropertiesEdit.Name},
		{Name: static_permission_sets.TemplatesEdit.Name},
		{Name: static_permission_sets.DataStorageEdit.Name},
		{Name: static_permission_sets.ClusterRead.Name},
		{Name: static_permission_sets.NetworkTopologyRead.Name},
		{Name: static_permission_sets.NodePoolsRead.Name},
		{Name: static_permission_sets.NodesRead.Name},
		{Name: static_permission_sets.PoliciesEdit.Name},
		{Name: static_permission_sets.DepartmentEdit.Name},
		{Name: static_permission_sets.ProjectEdit.Name},
		{Name: static_permission_sets.RolesRead.Name},
		{Name: static_permission_sets.AccessRulesEdit.Name},
		{Name: static_permission_sets.UsersRead.Name},
		{Name: static_permission_sets.ServiceAccountsEdit.Name},
		{Name: static_permission_sets.EventHistoryRead.Name},
		{Name: static_permission_sets.AccountRead.Name},
		{Name: static_permission_sets.SettingsRead.Name},
		{Name: static_permission_sets.BrandingSettingsRead.Name},
		{Name: static_permission_sets.SecuritySettingsRead.Name},
		{Name: static_permission_sets.NotificationSettingsRead.Name},
		{Name: static_permission_sets.UsersMinimalRead.Name},
	},
	Permissions: []*models.Permission{
		{
			ResourceType: api.RESOURCETYPE_TENANT,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_USERS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_USERS_MINIMAL,
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
			ResourceType: api.RESOURCETYPE_EVENTS_HISTORY,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ROLES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DEPARTMENT,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_APPS,
			Read:         true,
			Create:       true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_SERVICE_ACCOUNT,
			Read:         true,
			Create:       true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_PROJECT,
			Create:       true,
			Read:         true,
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
			ResourceType: api.RESOURCETYPE_ACCESS_RULES,
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
			ResourceType: api.RESOURCETYPE_APPLICATIONS,
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
			ResourceType: api.RESOURCETYPE_CREDENTIALS,
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
			ResourceType: api.RESOURCETYPE_WORKSPACES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
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
			ResourceType: api.RESOURCETYPE_DEVICE_CLASSES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CM_VOLUME_ASSETS,
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
			ResourceType: api.RESOURCETYPE_POLICIES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_INFERENCES,
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
			ResourceType: api.RESOURCETYPE_DATAVOLUMES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_REGISTRIES,
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
			ResourceType: api.RESOURCETYPE_NETWORK_TOPOLOGIES,
			Read:         true,
		},
	},
}
