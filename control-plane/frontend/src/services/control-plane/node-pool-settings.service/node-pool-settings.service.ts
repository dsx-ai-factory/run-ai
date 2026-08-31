/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { clusterServiceApi } from "@/services/infra/client-apis/cluster-service-api/cluster-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import {
  NodePoolSettingsKeyEnum,
  type NodePoolSettingsGetResponse,
  type NodePoolSetting,
} from "@/swagger-models/cluster-service-client";
import type { IProjectManualOverQuotaSetting, ICPUResourcesQuotaSetting } from "@/models/setting.model";

export interface INodePoolServiceSettings {
  projectManualOverQuota: IProjectManualOverQuotaSetting;
  cpuResourcesQuota: ICPUResourcesQuotaSetting;
}

export const nodePoolSettingsService = {
  getAll,
  getProjectManualOverQuota,
  getCpuResourcesQuota,
  updateProjectManualOverQuota,
  updateCpuResourcesQuota,
};

async function getAll(): Promise<INodePoolServiceSettings> {
  try {
    const response = await clusterServiceApi.settingsApi.getNodePoolSettings();
    const data = response.data as NodePoolSettingsGetResponse;

    return {
      projectManualOverQuota: {
        value: Boolean(data.projectManualOverQuota?.enabled),
        displayName: data.projectManualOverQuota?.displayName,
        description: data.projectManualOverQuota?.description,
        stage: data.projectManualOverQuota?.stage,
      },
      cpuResourcesQuota: {
        value: Boolean(data.cpuResourcesQuota?.enabled),
        displayName: data.cpuResourcesQuota?.displayName,
        description: data.cpuResourcesQuota?.description,
        stage: data.cpuResourcesQuota?.stage,
      },
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getProjectManualOverQuota(): Promise<IProjectManualOverQuotaSetting> {
  try {
    const response = await clusterServiceApi.settingsApi.getNodePoolSettingByKey(
      NodePoolSettingsKeyEnum.ProjectManualOverQuota,
    );
    const data = response.data as NodePoolSetting;
    return {
      value: Boolean(data.enabled),
      displayName: data.displayName,
      description: data.description,
      stage: data.stage,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getCpuResourcesQuota(): Promise<ICPUResourcesQuotaSetting> {
  try {
    const response = await clusterServiceApi.settingsApi.getNodePoolSettingByKey(
      NodePoolSettingsKeyEnum.CpuResourcesQuota,
    );
    const data = response.data as NodePoolSetting;
    return {
      value: Boolean(data.enabled),
      displayName: data.displayName,
      description: data.description,
      stage: data.stage,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateProjectManualOverQuota(enabled: boolean): Promise<void> {
  try {
    await clusterServiceApi.settingsApi.updateNodePoolSettingByKey(NodePoolSettingsKeyEnum.ProjectManualOverQuota, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateCpuResourcesQuota(enabled: boolean): Promise<void> {
  try {
    await clusterServiceApi.settingsApi.updateNodePoolSettingByKey(NodePoolSettingsKeyEnum.CpuResourcesQuota, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
