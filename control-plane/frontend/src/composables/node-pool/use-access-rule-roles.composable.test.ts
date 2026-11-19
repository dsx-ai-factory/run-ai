/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useAccessRuleRoles } from "@/composables/node-pool/use-access-rule-roles.composable";
import { useRoleStore } from "@/stores/role.store";
import { usePermissionStore } from "@/stores/permissions.store/index";
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";
import { Action, ResourceType, type Role } from "@/swagger-models/authorization-client/index";
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
describe("scheduling department administrator role selection", () => {
  it("loads API roles and offers eligible built-in and custom roles", async () => {
    usePermissionStore().setPermissions([
      { displayName: "Test", groupId: "iam", resourceType: ResourceType.Roles, actions: [Action.Read] },
      { displayName: "Test", groupId: "iam", resourceType: ResourceType.AccessRules, actions: [Action.Create] },
      { displayName: "Test", groupId: "iam", resourceType: ResourceType.Department, actions: [Action.Read] },
    ]);
    const list = vi.spyOn(rolesService, "list").mockResolvedValue(roles);
    const nodePoolModal = useAccessRuleRoles();
    await nodePoolModal.loadRoles();
    expect(list).toHaveBeenCalledTimes(1);
    expect(nodePoolModal.roleOptions.value.map((role) => role.value)).toEqual([2, 1]);
    expect(nodePoolModal.validateRole({ label: "Custom reader", value: 2 })).toBe(true);
    expect(nodePoolModal.validateRole({ label: "Custom manager", value: 3 })).not.toBe(true);
    expect(nodePoolModal.validateRole(null)).not.toBe(true);
    usePermissionStore().setPermissions([
      { displayName: "Test", groupId: "iam", resourceType: ResourceType.Roles, actions: [Action.Read] },
    ]);
    expect(nodePoolModal.validateRole({ label: "Custom reader", value: 2 })).not.toBe(true);
  });
  it("excludes disabled built-in and custom roles from scheduling department creation and rejects stale selections", () => {
    usePermissionStore().setPermissions([
      { displayName: "Test", groupId: "iam", resourceType: ResourceType.Roles, actions: [Action.Read] },
      { displayName: "Test", groupId: "iam", resourceType: ResourceType.AccessRules, actions: [Action.Create] },
      { displayName: "Test", groupId: "iam", resourceType: ResourceType.Department, actions: [Action.Read] },
    ]);
    const disabledBuiltIn = { ...roles[0], id: 4, name: "Disabled administrator", effectiveEnabled: false };
    const disabledCustom = { ...roles[1], id: 5, name: "Disabled custom reader", effectiveEnabled: false };
    const legacyRole = { ...roles[1], id: 6, name: "Legacy reader" };
    delete (legacyRole as Partial<Role>).effectiveEnabled;
    useRoleStore().roles = [...roles.map((role) => ({ ...role })), disabledBuiltIn, disabledCustom, legacyRole];
    const nodePoolModal = useAccessRuleRoles();
    expect(nodePoolModal.roleOptions.value.map((role) => role.value)).toEqual([2, 1, 6]);
    expect(nodePoolModal.validateRole({ label: "Disabled administrator", value: 4 })).not.toBe(true);
    expect(nodePoolModal.validateRole({ label: "Disabled custom reader", value: 5 })).not.toBe(true);
    expect(nodePoolModal.validateRole({ label: "Custom reader", value: 2 })).toBe(true);
    expect(nodePoolModal.validateRole({ label: "Legacy reader", value: 6 })).toBe(true);
    const customReader = useRoleStore().roles.find((role) => role.id === 2);
    if (!customReader) throw new Error("The custom reader fixture must exist");
    customReader.effectiveEnabled = false;
    expect(nodePoolModal.validateRole({ label: "Custom reader", value: 2 })).not.toBe(true);
    expect(nodePoolModal.roleOptions.value.map((role) => role.value)).toEqual([1, 6]);
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
