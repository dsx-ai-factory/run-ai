/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { clusterServiceApi } from "@/services/infra/client-apis/cluster-service-api/cluster-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import {
  ClusterSettingsKeyEnum,
  type ClusterSettingsGetResponse,
  type EnableAdvancedGpuMetricsSetting,
} from "@/swagger-models/cluster-service-client";
import type { IEnableAdvancedGpuMetricsSetting } from "@/models/setting.model";

export interface IClusterServiceSettings {
  enableAdvancedGpuMetrics: IEnableAdvancedGpuMetricsSetting;
}

export const clusterServiceSettingsService = {
  getAll,
  getEnableAdvancedGpuMetrics,
  updateEnableAdvancedGpuMetrics,
};

async function getAll(): Promise<IClusterServiceSettings> {
  try {
    const response = await clusterServiceApi.settingsApi.getClusterSettings();
    const data = response.data as ClusterSettingsGetResponse;

    return {
      enableAdvancedGpuMetrics: {
        value: Boolean(data.enableAdvancedGpuMetrics?.enabled),
        displayName: data.enableAdvancedGpuMetrics?.displayName,
        description: data.enableAdvancedGpuMetrics?.description,
        stage: data.enableAdvancedGpuMetrics?.stage,
      },
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getEnableAdvancedGpuMetrics(): Promise<IEnableAdvancedGpuMetricsSetting> {
  try {
    const response = await clusterServiceApi.settingsApi.getClusterSettingByKey(
      ClusterSettingsKeyEnum.EnableAdvancedGpuMetrics,
    );
    const data = response.data as EnableAdvancedGpuMetricsSetting;
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

async function updateEnableAdvancedGpuMetrics(enabled: boolean): Promise<void> {
  try {
    await clusterServiceApi.settingsApi.updateClusterSettingByKey(ClusterSettingsKeyEnum.EnableAdvancedGpuMetrics, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
