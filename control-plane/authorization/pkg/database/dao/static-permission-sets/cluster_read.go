/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_permission_sets

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var ClusterRead = models.PermissionSet{
	Name:        "Cluster read",
	Description: "Grants permissions to view clusters.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_CLUSTER,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
	},
}
