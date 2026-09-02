/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_permission_sets

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var ClusterConfigRead = models.PermissionSet{
	Name:        "Cluster config read",
	Description: "Grants permissions to view cluster configuration.",
	Internal:    true,
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_CLUSTER_CONFIG,
			Read:         true,
		},
	},
}
