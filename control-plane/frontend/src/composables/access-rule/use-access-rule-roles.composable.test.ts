/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useAccessRuleRoles } from "./use-access-rule-roles.composable";
import { useRoleStore } from "@/stores/role.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";
import { Action, ResourceType, type Role } from "@/swagger-models/authorization-client";
const roles: Role[] = [
  {
    id: 1,
    description: "",
    createdAt: "",
    updatedAt: "",
    createdBy: "",
    custom: false,
    effectiveEnabled: true,
    name: "Department administrator",
    permissions: [{ resourceType: ResourceType.Department, actions: [Action.Read] }],
  },
  {
    id: 2,
    description: "",
    createdAt: "",
    updatedAt: "",
    createdBy: "",
    custom: true,
    effectiveEnabled: true,
    name: "Custom reader",
    permissions: [{ resourceType: ResourceType.Department, actions: [Action.Read] }],
  },
  {
    id: 3,
    description: "",
    createdAt: "",
    updatedAt: "",
    createdBy: "",
    custom: true,
    effectiveEnabled: true,
    name: "Custom manager",
    permissions: [{ resourceType: ResourceType.Department, actions: [Action.Delete] }],
  },
];
beforeEach(() => {
  setActivePinia(createPinia());
  vi.restoreAllMocks();
});
describe("shared department access-rule role selection", () => {
  it("loads API roles and gives both consumers the same eligible built-in and custom roles", async () => {
    usePermissionStore().setPermissions([
      { displayName: "Test", groupId: "iam", resourceType: ResourceType.Roles, actions: [Action.Read] },
      { displayName: "Test", groupId: "iam", resourceType: ResourceType.AccessRules, actions: [Action.Create] },
      { displayName: "Test", groupId: "iam", resourceType: ResourceType.Department, actions: [Action.Read] },
    ]);
    const list = vi.spyOn(rolesService, "list").mockResolvedValue(roles);
    const departmentForm = useAccessRuleRoles();
    const nodePoolModal = useAccessRuleRoles();
    await departmentForm.loadRoles();
    await nodePoolModal.loadRoles();
    expect(list).toHaveBeenCalledTimes(1);
    expect(nodePoolModal.roleOptions.value).toEqual(departmentForm.roleOptions.value);
    expect(nodePoolModal.roleOptions.value.map((role) => role.value)).toEqual([2, 1]);
    expect(nodePoolModal.validateRole({ label: "Custom reader", value: 2 })).toBe(true);
    expect(nodePoolModal.validateRole({ label: "Custom manager", value: 3 })).not.toBe(true);
    expect(nodePoolModal.validateRole(null)).not.toBe(true);
    usePermissionStore().setPermissions([
      { displayName: "Test", groupId: "iam", resourceType: ResourceType.Roles, actions: [Action.Read] },
    ]);
    expect(nodePoolModal.validateRole({ label: "Custom reader", value: 2 })).not.toBe(true);
  });
  it("does not load or offer roles without read permission", async () => {
    useRoleStore().roles = roles;
    const load = vi.spyOn(useRoleStore(), "loadRoles");
    const selection = useAccessRuleRoles();
    await selection.loadRoles();
    expect(load).not.toHaveBeenCalled();
    expect(selection.roleOptions.value).toEqual([]);
    expect(selection.canAssignRoles.value).toBe(false);
  });
});
