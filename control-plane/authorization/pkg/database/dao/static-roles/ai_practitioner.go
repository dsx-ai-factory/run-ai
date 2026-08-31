/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_legacy_roles

import (
	static_permission_sets "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var AiPractitioner = models.Role{
	Name:                    "AI practitioner",
	Description:             "Can create, edit, and manage all workload types (workspace, training, inference, and YAML) and view related resources such as projects, templates and data sources.",
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
		{Name: static_permission_sets.DataStorageRead.Name},
		{Name: static_permission_sets.ProjectRead.Name},
		{Name: static_permission_sets.AccountRead.Name},
		{Name: static_permission_sets.SettingsRead.Name},
		{Name: static_permission_sets.SecuritySettingsRead.Name},
		{Name: static_permission_sets.NotificationSettingsRead.Name},
		{Name: static_permission_sets.BrandingSettingsRead.Name},
		{Name: static_permission_sets.UsersMinimalRead.Name},
	},
}
