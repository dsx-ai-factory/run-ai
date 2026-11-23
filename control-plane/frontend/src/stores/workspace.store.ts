/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";

// models
import { ResourceType, type Action, type PermittedScopes } from "@/swagger-models/authorization-client";
// services
import { permissionService } from "@/services/authorization/permission.service/permission.service";

import type { THybridFormState } from "@/models/workload-hybrid.model";

export const useWorkspaceStore = defineStore("Workspace", {
  state: () => ({
    actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,

    flatFormState: null as THybridFormState | null,
  }),
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(ResourceType.Workspaces);
    },
    setFlatFormState(flatData: THybridFormState | null): void {
      this.flatFormState = flatData;
    },
    getFlatFormState(): THybridFormState | null {
      return this.flatFormState;
    },
    clearFlatFormState(): void {
      this.flatFormState = null;
    },
  },
});
