/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";
import type { Action } from "@/swagger-models/authorization-client";
import { ResourceType, type PermittedScopes } from "@/swagger-models/authorization-client";
import { permissionService } from "@/services/authorization/permission.service/permission.service";

export const useWorkloadTemplateStore = defineStore("WorkloadTemplate", {
  state: () => ({
    actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,
  }),
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(ResourceType.Templates);
    },
  },
});
