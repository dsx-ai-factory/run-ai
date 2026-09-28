/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_permission_sets

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var TemplatesRead = models.PermissionSet{
	Name:        "Templates read",
	Description: "Grants permissions to view workload templates, including all related data and resources associated with them.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_TEMPLATES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ENVIRONMENTS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_COMPUTE_RESOURCES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_RESOURCE_CLAIM_TEMPLATES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_GIT_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_HOST_PATH_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NFS_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_S3_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_REGISTRIES,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CREDENTIALS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CM_VOLUME_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_SECRET_VOLUME_ASSETS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS_MINIMAL,
			Read:         true,
		},
	},
}
