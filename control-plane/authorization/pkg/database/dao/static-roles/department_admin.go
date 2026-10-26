/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_legacy_roles

import (
	static_permission_sets "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var DepartmentAdminLegacy = models.Role{
	Name: "Department administrator (Legacy)",
	Description: "A department administrator can do the following within their assigned scope:\n" +
		"Create, view, edit, delete: departments, projects, datavolumes\n" +
		"Assign: roles (researcher, ML engineer, research manager, viewer) within those departments and projects\n" +
		"View: dashboards (including Consumption)",
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	Deprecated:              true,
	ClusterRolePredefinedId: nil,
	PermissionSets: &[]models.PermissionSet{
		{Name: static_permission_sets.WorkspaceEdit.Name},
		{Name: static_permission_sets.TrainingEdit.Name},
		{Name: static_permission_sets.InferenceEdit.Name},
		{Name: static_permission_sets.ApplicationEdit.Name},
		{Name: static_permission_sets.YamlWorkloadEdit.Name},
		{Name: static_permission_sets.WorkloadRead.Name},
		{Name: static_permission_sets.WorkloadPropertiesRead.Name},
		{Name: static_permission_sets.TemplatesEdit.Name},
		{Name: static_permission_sets.DataStorageEdit.Name},
		{Name: static_permission_sets.ClusterRead.Name},
		{Name: static_permission_sets.NetworkTopologyRead.Name},
		{Name: static_permission_sets.NodePoolsRead.Name},
		{Name: static_permission_sets.NodesRead.Name},
		{Name: static_permission_sets.PoliciesEdit.Name},
		{Name: static_permission_sets.DepartmentRead.Name},
		{Name: static_permission_sets.ProjectEdit.Name},
		{Name: static_permission_sets.RolesRead.Name},
		{Name: static_permission_sets.AccessRulesEdit.Name},
		{Name: static_permission_sets.UsersEdit.Name},
		{Name: static_permission_sets.ServiceAccountsEdit.Name},
		{Name: static_permission_sets.AccountRead.Name},
		{Name: static_permission_sets.SettingsRead.Name},
		{Name: static_permission_sets.BrandingSettingsRead.Name},
		{Name: static_permission_sets.SecuritySettingsRead.Name},
		{Name: static_permission_sets.NotificationSettingsRead.Name},
		{Name: static_permission_sets.UsersMinimalRead.Name},
	},
}

var DepartmentAdmin = models.Role{
	Name: "Department administrator",
	Description: "A department administrator can do the following within their assigned scope:\n" +
		"Create, view, edit, delete: departments, projects, datavolumes\n" +
		"Assign: roles (researcher, ML engineer, research manager, viewer) within those departments and projects\n" +
		"View: dashboards (including Consumption)",
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	Deprecated:              false,
	ClusterRolePredefinedId: nil,
	PermissionSets: &[]models.PermissionSet{
		{Name: static_permission_sets.WorkspaceEdit.Name},
		{Name: static_permission_sets.TrainingEdit.Name},
		{Name: static_permission_sets.InferenceEdit.Name},
		{Name: static_permission_sets.ApplicationEdit.Name},
		{Name: static_permission_sets.YamlWorkloadEdit.Name},
		{Name: static_permission_sets.WorkloadRead.Name},
		{Name: static_permission_sets.WorkloadPropertiesRead.Name},
		{Name: static_permission_sets.TemplatesEdit.Name},
		{Name: static_permission_sets.DataStorageEdit.Name},
		{Name: static_permission_sets.NetworkTopologyRead.Name},
		{Name: static_permission_sets.NodePoolsRead.Name},
		{Name: static_permission_sets.NodesRead.Name},
		{Name: static_permission_sets.PoliciesEdit.Name},
		{Name: static_permission_sets.DepartmentRead.Name},
		{Name: static_permission_sets.ProjectEdit.Name},
		{Name: static_permission_sets.RolesRead.Name},
		{Name: static_permission_sets.AccessRulesEdit.Name},
		{Name: static_permission_sets.UsersEdit.Name},
		{Name: static_permission_sets.ServiceAccountsEdit.Name},
		{Name: static_permission_sets.AccountRead.Name},
		{Name: static_permission_sets.SettingsRead.Name},
		{Name: static_permission_sets.BrandingSettingsRead.Name},
		{Name: static_permission_sets.SecuritySettingsRead.Name},
		{Name: static_permission_sets.NotificationSettingsRead.Name},
	},
}
