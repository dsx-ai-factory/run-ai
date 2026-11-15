/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { computed } from "vue";
import { useRoleStore } from "@/stores/role.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import type { ISelectOption } from "@/models/global.model";
import { errorMessages } from "@/common/error-message.constant";

// Shared by the department/project access-rule form and the node-pool preview.
export function useAccessRuleRoles() {
  const roleStore = useRoleStore();
  const permissionStore = usePermissionStore();
  const canReadRoles = computed(() => permissionStore.hasPermission(ResourceType.Roles, Action.Read));
  const canAssignRoles = computed(
    () => canReadRoles.value && permissionStore.hasPermission(ResourceType.AccessRules, Action.Create),
  );
  const roleOptions = computed<ISelectOption[]>(() =>
    canReadRoles.value
      ? roleStore.roleList
          .filter((role) => rolesService.isPermittedRole(role, permissionStore.permissionsList))
          .map((role) => ({ label: role.name, value: role.id, roleStageBadge: role.name }))
      : [],
  );
  async function loadRoles() {
    if (canReadRoles.value && roleStore.roleList.length === 0) await roleStore.loadRoles();
  }
  function validateRole(option: ISelectOption | null | undefined): boolean | string {
    if (!option) return errorMessages.ROLE_TYPE_NOT_EMPTY;
    return (
      (canAssignRoles.value && roleOptions.value.some((role) => role.value === option.value)) ||
      "You cannot assign this role."
    );
  }
  return {
    roleOptions,
    canReadRoles,
    canAssignRoles,
    loadRoles,
    validateRole,
    loading: computed(() => roleStore.loading),
  };
}
