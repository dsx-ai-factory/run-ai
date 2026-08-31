/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_legacy_roles

import (
	static_permission_sets "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var ProjectAdmin = models.Role{
	Name: "Project administrator",
	Description: "Manages all aspects of a specific project, including workloads, templates, policies, and data resources. " +
		"This role allows you to edit workloads of all types, view project configurations, and oversee access rules for project users.",
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
		{Name: static_permission_sets.WorkloadPropertiesRead.Name},
		{Name: static_permission_sets.TemplatesEdit.Name},
		{Name: static_permission_sets.DataStorageEdit.Name},
		{Name: static_permission_sets.PoliciesEdit.Name},
		{Name: static_permission_sets.ProjectRead.Name},
		{Name: static_permission_sets.RolesRead.Name},
		{Name: static_permission_sets.AccessRulesEdit.Name},
		{Name: static_permission_sets.AccountRead.Name},
		{Name: static_permission_sets.SettingsRead.Name},
		{Name: static_permission_sets.SecuritySettingsRead.Name},
		{Name: static_permission_sets.NotificationSettingsRead.Name},
		{Name: static_permission_sets.BrandingSettingsRead.Name},
		{Name: static_permission_sets.UsersMinimalRead.Name},
	},
}
