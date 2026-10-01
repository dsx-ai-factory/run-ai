/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Stores
import { defineStore } from "pinia";

// Models
import type { IFilterBy } from "@/models/filter.model";
import type { Role } from "@/swagger-models/authorization-client";
import { defaultRolesFilterBy } from "@/table-models/roles.table-model";
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";
import { filterService } from "@/services/filter.service/filter.service";
import { allRoleColumns } from "@/table-models/roles.table-model";

export const useRoleStore = defineStore("Role", {
  state: () => ({
    roles: [] as Array<Role>,
    loading: false as boolean,
    namespaces: {} as Record<string, string | undefined>,
    filterBy: {} as IFilterBy,
  }),
  getters: {
    roleList(): Array<Role> {
      return this.roles.sort((a: Role, b: Role) => a.name.localeCompare(b.name));
    },
  },
  actions: {
    async loadRoles(filterBy?: IFilterBy): Promise<void> {
      this.loading = true;
      this.filterBy = filterBy || defaultRolesFilterBy;
      const rolesList = await rolesService.list();
      this.roles = filterService.filterListByTableFilters(rolesList, this.filterBy, allRoleColumns);
      this.loading = false;
    },
    getRoleById(id: number): Role {
      return this.roles.find((role: Role) => role.id === id) as Role;
    },
  },
});
