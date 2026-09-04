/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_permission_sets

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var BrandingSettingsEdit = models.PermissionSet{
	Name:        "Branding settings edit",
	Description: "Grants permissions to create, read, update, and delete branding settings.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_BRANDING_SETTINGS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
	},
}
