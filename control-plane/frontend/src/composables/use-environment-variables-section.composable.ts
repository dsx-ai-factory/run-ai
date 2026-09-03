/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { TEnvVarNoneCustomSourceOption } from "@/models/environment.model";
import type { IAssetsFilter } from "@/models/filter.model";
import { credentialService } from "@/services/control-plane/credential.service/credential.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import {
  AssetKind,
  type EnvironmentVariable,
  type EnvironmentVariableOfAsset,
} from "@/swagger-models/assets-service-client";
import { ScopeType } from "@/swagger-models/authorization-client";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";
import { scopeUtil } from "@/utils/scope.util";
import { computed, type Ref, ref } from "vue";

export function useEnvironmentVariablesSection() {
  async function loadSecretsForEnvironmentVariables(scope: ScopeType, scopeId: string) {
    const secrets = await credentialService.listExistingSecrets(scope, scopeId, undefined, "all", true);

    return secrets.map((s: SecretListInfo) => ({
      label: s.name,
      value: s.name,
      type: s.type,
    }));
  }

  async function loadCredentialsForEnvironmentVariables(
    clusterId?: string,
    departmentId?: string,
    projectId?: number,
  ): Promise<TEnvVarNoneCustomSourceOption[]> {
    const credentials = (
      await credentialService.listCredentials({
        clusterUuid: clusterId || undefined,
        departmentId: departmentId,
        projectId: projectId,
      })
    ).map((c) => ({
      label: c.meta.name,
      value: c.meta.id,
      type: c.meta.kind,
    }));

    return credentials;
  }

  async function loadConfigMapsForEnvironmentVariables(
    scope: ScopeType,
    scopeId: string,
  ): Promise<TEnvVarNoneCustomSourceOption[]> {
    const configMaps = (await dataSourceService.listExistingConfigMaps(scope, scopeId, true)).map((cm) => ({
      label: cm.name,
      value: cm.name,
      type: AssetKind.ConfigMap,
    }));

    return configMaps;
  }

  const environmentVariablesSourceValueOptions: Ref<TEnvVarNoneCustomSourceOption[]> = ref<
    TEnvVarNoneCustomSourceOption[]
  >([]);

  async function loadUserCredentialsForEnvironmentVariables(
    scope: ScopeType,
    scopeId: string,
  ): Promise<TEnvVarNoneCustomSourceOption[]> {
    const { projectId, departmentId, clusterId } = scopeUtil.getScopeModelFromTypeAndId(scope, scopeId);
    const filter: IAssetsFilter = {
      projectId,
      departmentId,
      clusterUuid: clusterId || undefined,
      privateOnly: true,
    };
    const userCredentials = await credentialService.listCredentials(filter);
    return userCredentials.map((uc) => ({
      label: uc.meta.name,
      value: uc.meta.name,
      type: uc.meta.kind,
      isPrivate: true,
    }));
  }

  const environmentVariablesOptionsMap = computed(() =>
    environmentVariablesSourceValueOptions.value.reduce(
      (acc, option: TEnvVarNoneCustomSourceOption) => {
        acc[option.value] = option.value;
        return acc;
      },
      {} as Record<string, string>,
    ),
  );

  async function loadSourceValueOptions(
    isHybrid: boolean,
    projectId?: string,
    departmentId?: string,
    clusterId?: string,
  ) {
    let scope: ScopeType = ScopeType.Project;
    let scopeId = "";
    projectId && (scope = ScopeType.Project) && (scopeId = projectId);
    departmentId && (scope = ScopeType.Department) && (scopeId = departmentId);
    clusterId && (scope = ScopeType.Cluster) && (scopeId = clusterId);
    if (!scopeId) {
      console.warn("No scope ID provided for environment variables source value options");
      return;
    }

    if (isHybrid) {
      const [secrets, configMaps, userCredentials] = await Promise.all([
        loadSecretsForEnvironmentVariables(scope, scopeId),
        loadConfigMapsForEnvironmentVariables(scope, scopeId),
        loadUserCredentialsForEnvironmentVariables(scope, scopeId),
      ]);
      environmentVariablesSourceValueOptions.value = [...secrets, ...configMaps, ...userCredentials];
    } else {
      const [credentials, configMaps] = await Promise.all([
        loadCredentialsForEnvironmentVariables(clusterId, departmentId, projectId ? Number(projectId) : undefined),
        loadConfigMapsForEnvironmentVariables(scope, scopeId),
      ]);
      environmentVariablesSourceValueOptions.value = [...credentials, ...configMaps];
    }
  }

  // validateEnvironmentVariables is used to validate that the environment variables values are still valid (after scope change or copy&edit)
  function validateEnvironmentVariablesOfAsset(environmentVariables: EnvironmentVariableOfAsset[]): void {
    if (!environmentVariables) return;
    environmentVariables.forEach((envVar: EnvironmentVariableOfAsset) => {
      if (envVar.credential?.assetId) {
        const credential = environmentVariablesOptionsMap.value[envVar.credential.assetId];
        if (!credential) {
          envVar.credential.assetId = null;
        }
      }
      if (envVar.configMap?.name) {
        const configMap = environmentVariablesOptionsMap.value[envVar.configMap.name];
        if (!configMap) {
          envVar.configMap.name = null;
        }
      }
    });
  }

  // validateEnvironmentVariables is used to validate that the environment variables values are still valid (after scope change or copy&edit)
  function validateEnvironmentVariables(environmentVariables: EnvironmentVariable[]): void {
    if (!environmentVariables) return;
    environmentVariables.forEach((envVar: EnvironmentVariable) => {
      if (envVar.secret?.name) {
        const secret = environmentVariablesOptionsMap.value[envVar.secret.name];
        if (!secret) {
          envVar.secret.name = null;
        }
      }
      if (envVar.configMap?.name) {
        const configMap = environmentVariablesOptionsMap.value[envVar.configMap.name];
        if (!configMap) {
          envVar.configMap.name = null;
        }
      }
    });
  }

  return {
    loadSourceValueOptions,
    environmentVariablesSourceValueOptions,
    validateEnvironmentVariables,
    validateEnvironmentVariablesOfAsset,
  };
}
