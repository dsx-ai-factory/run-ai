/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import {
  WorkloadsSettingsKeyEnum,
  type WorkloadsSettingsGetResponse,
  type GetEnablePolicies,
  type GetGpuOverProvisioning,
} from "@/swagger-models/workloads-client";
import type { IEnablePoliciesSetting, IGpuOverProvisioningSetting } from "@/models/setting.model";

export interface IWorkloadsServiceSettings {
  enablePolicies: IEnablePoliciesSetting;
  gpuOverProvisioning: IGpuOverProvisioningSetting;
}

export const workloadsServiceSettingsService = {
  getAll,
  getEnablePolicies,
  updateEnablePolicies,
  getGpuOverProvisioning,
  updateGpuOverProvisioning,
};

async function getAll(): Promise<IWorkloadsServiceSettings> {
  try {
    const response = await workloadServiceApi.settingsApi.getWorkloadsServiceSettings();
    const data = response.data as WorkloadsSettingsGetResponse;

    return {
      enablePolicies: {
        value: Boolean(data.enablePolicies?.value?.enabled),
        displayName: data.enablePolicies?.displayName,
        description: data.enablePolicies?.description,
        stage: data.enablePolicies?.stage,
      },
      gpuOverProvisioning: {
        value: Boolean(data.gpuOverProvisioning?.value?.enabled),
        displayName: data.gpuOverProvisioning?.displayName,
        description: data.gpuOverProvisioning?.description,
        stage: data.gpuOverProvisioning?.stage,
      },
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getEnablePolicies(): Promise<IEnablePoliciesSetting> {
  try {
    const response = await workloadServiceApi.settingsApi.getWorkloadsServiceSettingByKey(
      WorkloadsSettingsKeyEnum.EnablePolicies,
    );
    const data = response.data as GetEnablePolicies;
    return {
      value: Boolean(data.value?.enabled),
      displayName: data.displayName,
      description: data.description,
      stage: data.stage,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateEnablePolicies(enabled: boolean): Promise<void> {
  try {
    await workloadServiceApi.settingsApi.updateWorkloadsServiceSettingByKey(WorkloadsSettingsKeyEnum.EnablePolicies, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getGpuOverProvisioning(): Promise<IGpuOverProvisioningSetting> {
  try {
    const response = await workloadServiceApi.settingsApi.getWorkloadsServiceSettingByKey(
      WorkloadsSettingsKeyEnum.GpuOverProvisioning,
    );
    const data = response.data as GetGpuOverProvisioning;
    return {
      value: Boolean(data.value?.enabled),
      displayName: data.displayName,
      description: data.description,
      stage: data.stage,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateGpuOverProvisioning(enabled: boolean): Promise<void> {
  try {
    await workloadServiceApi.settingsApi.updateWorkloadsServiceSettingByKey(
      WorkloadsSettingsKeyEnum.GpuOverProvisioning,
      { enabled },
    );
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
