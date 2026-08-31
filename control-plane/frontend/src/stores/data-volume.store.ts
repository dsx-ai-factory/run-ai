/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";

// models
import { ResourceType, type Action, type PermittedScopes } from "@/swagger-models/authorization-client";

// services
import { permissionService } from "@/services/authorization/permission.service/permission.service";

export const useDataVolumeStore = defineStore("DataVolume", {
  state: () => ({
    dataVolumeActionPermissions: {} as Record<Action, PermittedScopes>,
  }),
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.dataVolumeActionPermissions).length) return;
      this.dataVolumeActionPermissions = await permissionService.getPermittedScopesForAction(ResourceType.Datavolumes);
    },
  },
});
