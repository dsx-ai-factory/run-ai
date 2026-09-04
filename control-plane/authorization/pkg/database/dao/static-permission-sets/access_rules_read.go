/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_permission_sets

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var AccessRulesRead = models.PermissionSet{
	Name:        "Access rules read",
	Description: "Grants permissions to view access rules, including all related data and resources associated with them.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_ACCESS_RULES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ROLES,
			Read:         true,
		},
	},
}
