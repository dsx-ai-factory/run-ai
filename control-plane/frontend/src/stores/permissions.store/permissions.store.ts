/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Stores
import { defineStore } from "pinia";

// Models
import { type Permission, ResourceType, Action } from "@/swagger-models/authorization-client";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
export type PermissionCheck = {
  resourceType: ResourceType;
  action: Action;
};

export const usePermissionStore = defineStore("Permission", {
  state: () => ({
    permissions: [] as Array<Permission>,
    resourceTypeMap: {} as Record<ResourceType, Permission>,
  }),
  getters: {
    permissionsList(): Array<Permission> {
      return this.permissions;
    },
  },
  actions: {
    async loadPermissions(forceLoad = false): Promise<void> {
      if (this.permissions.length && !forceLoad) return;
      const permissions: Permission[] = await permissionService.getPermissions();
      this.setPermissions(permissions);
    },
    setPermissions(permissions: Permission[]): void {
      this.permissions = permissions;
      this.resourceTypeMap = this.permissions.reduce(
        (acc: Record<ResourceType, Permission>, permission: Permission) => {
          acc[permission.resourceType] = permission;
          return acc;
        },
        {} as Record<ResourceType, Permission>,
      );
    },
    hasAllPermissions(permissions: Array<PermissionCheck>): boolean {
      return permissions.every((permission) => this.hasPermission(permission.resourceType, permission.action));
    },

    getPermittedActions(resourceType: ResourceType): Action[] {
      return (
        this.permissions.find((permission: Permission) => {
          return permission.resourceType === resourceType;
        })?.actions || []
      );
    },

    isReadOnly(resourceType: ResourceType): boolean {
      const permittedActions = this.getPermittedActions(resourceType);
      return permittedActions.length === 1 && permittedActions[0] === Action.Read;
    },

    hasPermission(resourceType: ResourceType | ResourceType[], action: Action): boolean {
      if (Array.isArray(resourceType)) {
        return resourceType.some((res) => this.resourceTypeMap[res]?.actions.includes(action));
      }
      return this.resourceTypeMap[resourceType]?.actions.includes(action) || false;
    },
    hasDeletePermission(resourceType: ResourceType | ResourceType[]): boolean {
      return this.hasPermission(resourceType, Action.Delete);
    },
    isActionGrantedForEveryResource(resourceTypes: ResourceType[], action: Action): boolean {
      return resourceTypes.every((resourceType: ResourceType) =>
        this.resourceTypeMap[resourceType]?.actions.includes(action),
      );
    },
    isActionGrantedForAnyResource(resourceTypes: ResourceType[], action: Action): boolean {
      return resourceTypes.some((resourceType: ResourceType) =>
        this.resourceTypeMap[resourceType]?.actions.includes(action),
      );
    },
    hasSomePermittedActions(resourceType: ResourceType | ResourceType[], actions: Action[]): boolean {
      return actions.some((action: Action) =>
        Array.isArray(resourceType)
          ? resourceType.some((res) => this.resourceTypeMap[res]?.actions.includes(action))
          : this.resourceTypeMap[resourceType]?.actions.includes(action) || false,
      );
    },
    hasOverviewPermission(): boolean {
      return this.isActionGrantedForAnyResource(
        [
          ResourceType.Cluster,
          ResourceType.Nodes,
          ResourceType.Department,
          ResourceType.Project,
          ResourceType.Workloads,
          ResourceType.Nodepools,
        ],
        Action.Read,
      );
    },
    // Permission helpers for Overview dashboard widgets
    hasReadNodesPermission(): boolean {
      return this.hasPermission(ResourceType.Nodes, Action.Read);
    },
    hasReadClustersPermission(): boolean {
      return this.hasPermission(ResourceType.Cluster, Action.Read);
    },
    hasReadWorkloadsPermission(): boolean {
      return this.hasPermission(ResourceType.Workloads, Action.Read);
    },
    hasReadNodePoolsPermission(): boolean {
      return this.hasPermission(ResourceType.Nodepools, Action.Read);
    },
    hasReadDepartmentsPermission(): boolean {
      return this.hasPermission(ResourceType.Department, Action.Read);
    },
    hasReadProjectsPermission(): boolean {
      return this.hasPermission(ResourceType.Project, Action.Read);
    },
  },
});
