/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package static_permission_sets

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
)

var ClusterServiceAccountPS = models.PermissionSet{
	Name:        "Cluster service account permission set",
	Description: "",
	Internal:    true,
	Permissions: []models.RolePermission{
		{
			ResourceType: api.RESOURCETYPE_TENANT,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DEPARTMENT,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_PROJECT,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKLOADS,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_APPLICATIONS,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKLOAD_PROPERTIES,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_USERS,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_USERS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_EVENTS_HISTORY,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTER,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTERS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODEPOOLS_MINIMAL,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NODES,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ROLES,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_ACCESS_RULES,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_WORKSPACES,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_TRAININGS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_INFERENCES,
			Read:         true,
			Delete:       true,
			Sync:         true,
			Update:       true,
			Create:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_CREDENTIALS,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_PVC_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DEVICE_CLASSES,
			Read:         true,
			Sync:         true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_RESOURCE_CLAIM_TEMPLATES,
			Read:         true,
			Sync:         true,
			Delete:       true,
		},
		{
			ResourceType: api.RESOURCETYPE_CM_VOLUME_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_SECRET_VOLUME_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_GIT_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_HOST_PATH_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NFS_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_S3_ASSETS,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_POLICIES,
			Read:         true,
			Delete:       true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_DATAVOLUMES,
			Read:         true,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_CLUSTER_CONFIG,
			Sync:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_SECURITY_SETTINGS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_BRANDING_SETTINGS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_SETTINGS,
			Read:         true,
		},
		{
			ResourceType: api.RESOURCETYPE_NETWORK_TOPOLOGIES,
			Read:         true,
			Sync:         true,
		},
	},
}
