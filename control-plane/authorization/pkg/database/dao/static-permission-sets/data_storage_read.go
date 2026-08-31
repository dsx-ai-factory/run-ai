/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_permission_sets

import (
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
)

var DataStorageRead = models.PermissionSet{
	Name:        "Data & storage read",
	Description: "Grants permissions to view data and storage, including all related data and resources associated with them.",
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_STORAGE_CLASS_CONFIGURATION,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_PVC_ASSETS,
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
			ResourceType: api.RESOURCETYPE_DATAVOLUMES,
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
