/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var AccountRead = models.PermissionSet{
	Name:        "Account read",
	Description: "Grants permissions to view accounts.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_TENANT,
			Read:         true,
		},
	},
}
