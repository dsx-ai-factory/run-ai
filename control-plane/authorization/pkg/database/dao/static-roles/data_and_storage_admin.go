/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_legacy_roles

import (
	static_permission_sets "github.com/run-ai/runai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var DataAndStorageAdmin = models.Role{
	Name:                    "Data and storage administrator",
	Description:             "Manages data-related resources, including data sources and data volumes. This role allows you to create and modify data assets, with read-only access to workloads and templates.",
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	ClusterRolePredefinedId: nil,
	PermissionSets: &[]models.PermissionSet{
		{Name: static_permission_sets.DataStorageEdit.Name},
		{Name: static_permission_sets.TemplatesRead.Name},
		{Name: static_permission_sets.WorkloadRead.Name},
		{Name: static_permission_sets.ApplicationRead.Name},
		{Name: static_permission_sets.AccountRead.Name},
		{Name: static_permission_sets.SettingsRead.Name},
		{Name: static_permission_sets.SecuritySettingsRead.Name},
		{Name: static_permission_sets.NotificationSettingsRead.Name},
		{Name: static_permission_sets.BrandingSettingsRead.Name},
		{Name: static_permission_sets.UsersMinimalRead.Name},
	},
}
