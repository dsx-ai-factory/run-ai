/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var WorkloadPropertiesRead = models.PermissionSet{
	Name:        "Workload properties read",
	Description: "Grants permissions to view workload properties, including all related data and resources associated with them.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_WORKLOAD_PROPERTIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
	},
}
