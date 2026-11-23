/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { PermissionAction } from '../../../../../../modules/api/apis/authorizationApi/generated/model/permissionAction';
import { ResourceType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/resourceType';

export const SETTINGS_RESOURCES: ResourceType[] = [
    ResourceType.Settings,
    ResourceType.SecuritySettings,
    ResourceType.BrandingSettings,
    ResourceType.NotificationSettings,
    ResourceType.Registries
];

export const BASELINE_RESOURCES: ResourceType[] = [
    ResourceType.Tenant,
    ResourceType.Cluster,
    ResourceType.ClusterConfig,
    ResourceType.ClustersMinimal,
    ResourceType.Department
];

export const CRUD_ACTIONS: PermissionAction[] = [
    PermissionAction.Create,
    PermissionAction.Read,
    PermissionAction.Update,
    PermissionAction.Delete
];
