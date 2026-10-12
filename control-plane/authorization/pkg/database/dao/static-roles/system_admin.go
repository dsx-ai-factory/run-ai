/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_legacy_roles

import (
	static_permission_sets "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var SystemAdmin = models.Role{
	Name: "System administrator",
	Description: "A system administrator can:\n" +
		"Control all aspects of the system: users, organizations, compute resources, general settings, etc.\n" +
		"This role has global system control and should be limited to a small group of skilled IT administrators.",
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
		{Name: static_permission_sets.ClusterEdit.Name},
		{Name: static_permission_sets.ClusterConfigEdit.Name},
		{Name: static_permission_sets.NetworkTopologyEdit.Name},
		{Name: static_permission_sets.NodePoolsEdit.Name},
		{Name: static_permission_sets.NodesRead.Name},
		{Name: static_permission_sets.PoliciesEdit.Name},
		{Name: static_permission_sets.DepartmentEdit.Name},
		{Name: static_permission_sets.ProjectEdit.Name},
		{Name: static_permission_sets.RolesEdit.Name},
		{Name: static_permission_sets.AccessRulesEdit.Name},
		{Name: static_permission_sets.UsersEdit.Name},
		{Name: static_permission_sets.ServiceAccountsEdit.Name},
		{Name: static_permission_sets.EventHistoryRead.Name},
		{Name: static_permission_sets.AccountRead.Name},
		{Name: static_permission_sets.SettingsEdit.Name},
		{Name: static_permission_sets.BrandingSettingsEdit.Name},
		{Name: static_permission_sets.SecuritySettingsEdit.Name},
		{Name: static_permission_sets.NotificationSettingsEdit.Name},
		{Name: static_permission_sets.IntegrationMetricsRead.Name},
		{Name: static_permission_sets.AccessKeysEdit.Name},
		{Name: static_permission_sets.UsersMinimalRead.Name},
	},
}
