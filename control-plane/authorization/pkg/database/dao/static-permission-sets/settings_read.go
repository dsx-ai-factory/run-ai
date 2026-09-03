/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_permission_sets

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var SettingsRead = models.PermissionSet{
	Name:        "Settings read",
	Description: "Grants permissions to view settings.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_SETTINGS,
			Read:         true,
		},
	},
}
