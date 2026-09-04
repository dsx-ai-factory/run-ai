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

var CloudOperatorDgxc = models.Role{
	Name: "Cloud operator-DGXC",
	Description: "A cloud operator can do the following:\n" +
		"Manage: clusters, settings, nodepools.\n" +
		"View: nodes, users, roles, event history.",
	ShowForDGXC:             true,
	RoleType:                models.DGXC,
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	ClusterRolePredefinedId: nil,
	PermissionSets: &[]models.PermissionSet{
		{Name: static_permission_sets.ClusterEdit.Name},
		{Name: static_permission_sets.NodePoolsEdit.Name},
		{Name: static_permission_sets.NodesRead.Name},
		{Name: static_permission_sets.RolesEdit.Name},
		{Name: static_permission_sets.AccessRulesRead.Name},
		{Name: static_permission_sets.UsersRead.Name},
		{Name: static_permission_sets.ServiceAccountsEdit.Name},
		{Name: static_permission_sets.EventHistoryRead.Name},
		{Name: static_permission_sets.AccountRead.Name},
		{Name: static_permission_sets.SettingsEdit.Name},
		{Name: static_permission_sets.BrandingSettingsEdit.Name},
		{Name: static_permission_sets.SecuritySettingsEdit.Name},
		{Name: static_permission_sets.NotificationSettingsEdit.Name},
		{Name: static_permission_sets.StorageClassConfigEdit.Name},
		{Name: static_permission_sets.UsersMinimalRead.Name},
	},
	Permissions: []*models.Permission{
		{
			ResourceType:           api.RESOURCETYPE_TENANT,
			Create:                 true,
			Read:                   true,
			Update:                 true,
			Delete:                 true,
			ReadOnlyForDGXCManaged: true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTER,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_SETTINGS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_SECURITY_SETTINGS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_NOTIFICATION_SETTINGS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_BRANDING_SETTINGS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
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
			ResourceType: api.RESOURCETYPE_APPS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_SERVICE_ACCOUNT,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_ROLES,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_ACCESS_RULES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_EVENTS_HISTORY,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_STORAGE_CLASS_CONFIGURATION,
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
	},
}
