/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_legacy_roles

import (
	static_permission_sets "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var ClusterServiceAccount = models.Role{
	Name:                    "Cluster service account",
	Description:             "An internal cluster service account used to sync the data from the cluster to the control plane and in the opposite way.",
	InternalOnly:            true,
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	ClusterRolePredefinedId: nil,
	PermissionSets: &[]models.PermissionSet{
		{Name: static_permission_sets.ClusterServiceAccountPS.Name}, {Name: static_permission_sets.UsersMinimalRead.Name},
	},
}
