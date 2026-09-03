/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";

import type { IAssetsFilter } from "@/models/filter.model";
import type {
  EnvironmentAsset,
  EnvironmentCreationRequest,
  EnvironmentUpdateRequest,
} from "@/swagger-models/assets-service-client";
import type { Action } from "@/swagger-models/authorization-client";
import { type PermittedScopes, ResourceType } from "@/swagger-models/authorization-client";
import { environmentService } from "@/services/control-plane/environment.service/environment.service";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
import { useSettingStore } from "./setting.store";
import { GPT2_ASSET_ID } from "@/models/gpt2-model-spec.model";

export const useEnvironmentStore = defineStore("Environment", {
  state: () => ({
    environments: [] as Array<EnvironmentAsset>,
    actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,
  }),
  getters: {
    environmentList(): Array<EnvironmentAsset> {
      return this.environments.filter((env: EnvironmentAsset) =>
        useSettingStore().isAllowDeployingGPT2ModelEnabled ? true : env.meta.id !== GPT2_ASSET_ID,
      );
    },
  },
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(ResourceType.Environments);
    },
    async loadEnvironments(filterBy: IAssetsFilter = {}): Promise<EnvironmentAsset[]> {
      this.environments = await environmentService.list(filterBy);
      return this.environmentList;
    },
    async loadById(environmentId: string): Promise<EnvironmentAsset> {
      let environment: EnvironmentAsset | undefined = this.environmentList.find(
        (env: EnvironmentAsset) => env.meta.id === environmentId,
      );
      if (!environment) environment = await environmentService.getById(environmentId);
      return environment;
    },
    async remove(environmentId: string): Promise<void> {
      await environmentService.remove(environmentId);
    },
    async update(environmentId: string, environment: EnvironmentUpdateRequest): Promise<EnvironmentAsset> {
      return environmentService.update(environmentId, environment);
    },
    async createEnvironment(environment: EnvironmentCreationRequest): Promise<EnvironmentAsset> {
      return await environmentService.save(environment);
    },
  },
});
