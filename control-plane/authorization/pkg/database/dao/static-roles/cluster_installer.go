/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_legacy_roles

import (
	"k8s.io/utils/ptr"

	static_permission_sets "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao/static-permission-sets"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var ClusterInstaller = models.Role{
	Name:                    "Cluster installer",
	Description:             "Allows a service account to read the control-plane configuration of a cluster. Used by non-interactive installers such as Argo CD",
	InternalOnly:            false,
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	ClusterRolePredefinedId: nil,
	NoClusterPermissions:    ptr.To(true),
	PermissionSets: &[]models.PermissionSet{
		{Name: static_permission_sets.ClusterConfigRead.Name},
	},
}
