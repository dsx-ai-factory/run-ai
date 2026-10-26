/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ResourceTypeGroupId, Role, RolePermission } from "@/swagger-models/authorization-client";

// Each resource type with its entity group, as the permission store holds it; the roles API returns a
// role's permissions without the group.
const permissionsWithGroup: Array<RolePermission & { groupId: ResourceTypeGroupId }> = [
  { resourceType: "tenant", displayName: "Account", groupId: "organization", actions: ["read"] },
  { resourceType: "nodes", displayName: "Nodes", groupId: "physical-resource", actions: ["read"] },
  { resourceType: "users", displayName: "Users", groupId: "iam", actions: ["create", "read", "update", "delete"] },
  {
    resourceType: "apps",
    displayName: "Applications",
    groupId: "iam",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "access_rules",
    displayName: "Access rules",
    groupId: "iam",
    actions: ["create", "read", "update", "delete"],
  },
  { resourceType: "roles", displayName: "Roles", groupId: "iam", actions: ["read"] },
  { resourceType: "department", displayName: "Departments", groupId: "organization", actions: ["read"] },
  {
    resourceType: "project",
    displayName: "Projects",
    groupId: "organization",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "workloads",
    displayName: "Workloads",
    groupId: "workload",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "workspaces",
    displayName: "Workspaces",
    groupId: "workload",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "trainings",
    displayName: "Trainings",
    groupId: "workload",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "inferences",
    displayName: "Inferences",
    groupId: "workload",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "environments",
    displayName: "Environments",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "host-path-assets",
    displayName: "Data sources",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "nfs-assets",
    displayName: "Data sources",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "pvc-assets",
    displayName: "Data sources",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "cm-volume-assets",
    displayName: "Data sources",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "secret-volume-assets",
    displayName: "Data sources",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "git-assets",
    displayName: "Data sources",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "s3-assets",
    displayName: "Data sources",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "compute-resources",
    displayName: "Compute resources",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "templates",
    displayName: "Templates",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  { resourceType: "cluster", displayName: "Clusters", groupId: "physical-resource", actions: ["read"] },
  { resourceType: "nodepools", displayName: "Node pools", groupId: "physical-resource", actions: ["read"] },
  {
    resourceType: "credentials",
    displayName: "Credentials",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "policies",
    displayName: "Policies",
    groupId: "organization",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "datavolumes",
    displayName: "Data volumes",
    groupId: "workload-asset",
    actions: ["create", "read", "update", "delete"],
  },
  {
    resourceType: "cluster-config",
    displayName: "Cluster configuration",
    groupId: "physical-resource",
    actions: [],
  },
  { resourceType: "settings", displayName: "Settings", groupId: "organization", actions: [] },
  { resourceType: "events-history", displayName: "Event history", groupId: "organization", actions: [] },
];

const roles: Array<Role> = [
  {
    custom: false,
    effectiveEnabled: true,
    name: "Department administrator",
    description:
      "A department administrator can do the following within their assigned scope:\nCreate, view, edit, delete: departments, projects, datavolumes, datavolumes/sharedscopes\nAssign: roles (researcher, ML engineer, research manager, viewer) within those departments and projects\nView: dashboards (including Consumption)",
    permissions: permissionsWithGroup.map(({ resourceType, displayName, actions }) => ({
      resourceType,
      displayName,
      actions,
    })),
    id: 2,
    createdAt: "2024-06-18T11:45:00.71943Z",
    updatedAt: "2024-06-23T10:25:46.987415Z",
    createdBy: "Run:ai",
  },
];

export const roleReviewMock = {
  roles,
  permissionsWithGroup,
};
