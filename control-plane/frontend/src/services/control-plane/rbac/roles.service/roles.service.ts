/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Permission, Role } from "@/swagger-models/authorization-client";
import { authorizationApi } from "@/services/infra/client-apis/authorization-api/authorization-api";

const ROLES_PAGE_SIZE = 500;

export const rolesService = {
  list,
  isPermittedRole,
};
async function list(): Promise<Role[]> {
  const roles: Role[] = [];
  let offset: number | undefined = 0;
  while (offset !== undefined) {
    const response = await authorizationApi.rolesApi.getRoles(ROLES_PAGE_SIZE, offset);
    roles.push(...response.data.roles);
    const next = response.data.next;
    offset = typeof next === "number" && next > offset ? next : undefined;
  }
  return roles;
}

function isPermittedRole(role: Role, permissions: Array<Permission>): boolean {
  // Check that the permissions of the current user contain at least all the resource types of the role
  for (const permission of role.permissions) {
    const rolePermission = permissions.find((p) => p.resourceType === permission.resourceType);
    if (!rolePermission) {
      return false;
    }
    // Check that the role's permission object has at least all the actions that the first permission object has
    for (const action of permission.actions) {
      if (!rolePermission.actions.includes(action)) {
        return false;
      }
    }
  }
  // If all checks pass, the role is valid
  return true;
}

export default rolesService;
