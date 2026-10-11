/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_legacy_roles

import (
	static_permission_sets "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var CloudOperator = models.Role{
	Name: "Cloud operator",
	Description: "A cloud operator can do the following:\n" +
		"Manage: clusters, settings, nodepools.\n" +
		"View: nodes, users, roles, event history.",
	InternalOnly:            true,
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
		{Name: static_permission_sets.AccountEdit.Name},
		{Name: static_permission_sets.SettingsEdit.Name},
		{Name: static_permission_sets.BrandingSettingsEdit.Name},
		{Name: static_permission_sets.SecuritySettingsEdit.Name},
		{Name: static_permission_sets.NotificationSettingsEdit.Name},
		{Name: static_permission_sets.StorageClassConfigEdit.Name},
		{Name: static_permission_sets.UsersMinimalRead.Name},
	},
}
