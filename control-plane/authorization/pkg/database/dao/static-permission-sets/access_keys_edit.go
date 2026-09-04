/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_permission_sets

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var AccessKeysEdit = models.PermissionSet{
	Name:        "Access keys edit",
	Description: "Grants full permissions to create, read, update, and delete user access keys.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_ACCESS_KEYS,
			Create:       true,
			Read:         true,
			Update:       true,
			Delete:       true,
		},
	},
}
