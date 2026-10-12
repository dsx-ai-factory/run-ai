/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_legacy_roles

import (
	static_permission_sets "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var DepartmentViewer = models.Role{
	Name: "Department viewer",
	Description: "A department viewer can do the following within their assigned scope:\n" +
		"View: departments, projects, assigned subordinates (jobs, deployments, workspaces, environments, data sources, compute resources, templates), dashboards, datavolumes",
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	ClusterRolePredefinedId: nil,
	PermissionSets: &[]models.PermissionSet{
		{Name: static_permission_sets.WorkloadRead.Name},
		{Name: static_permission_sets.ApplicationRead.Name},
		{Name: static_permission_sets.WorkloadPropertiesRead.Name},
		{Name: static_permission_sets.TemplatesRead.Name},
		{Name: static_permission_sets.DataStorageRead.Name},
		{Name: static_permission_sets.ClusterRead.Name},
		{Name: static_permission_sets.PoliciesRead.Name},
		{Name: static_permission_sets.DepartmentRead.Name},
		{Name: static_permission_sets.ProjectRead.Name},
		{Name: static_permission_sets.AccountRead.Name},
		{Name: static_permission_sets.SettingsRead.Name},
		{Name: static_permission_sets.BrandingSettingsRead.Name},
		{Name: static_permission_sets.SecuritySettingsRead.Name},
		{Name: static_permission_sets.NotificationSettingsRead.Name},
		{Name: static_permission_sets.UsersMinimalRead.Name},
	},
}
