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

var IntegrationMetricsCollector = models.Role{
	Name:                    "Integration metrics collector",
	Description:             "An integration metrics collector can collect integration metrics from the control plane.",
	ScopeType:               models.ROLE_SCOPE_SYSTEM,
	Custom:                  false,
	Enabled:                 true,
	InternalOnly:            true,
	ClusterRolePredefinedId: nil,
	PermissionSets: &[]models.PermissionSet{
		{Name: static_permission_sets.IntegrationMetricsRead.Name},
		{Name: static_permission_sets.UsersMinimalRead.Name},
	},
	Permissions: []*models.Permission{
		{
			ResourceType: api.RESOURCETYPE_WORKLOAD_INTEGRATION_METRICS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_USERS_MINIMAL,
			Read:         true,
		},
	},
}
