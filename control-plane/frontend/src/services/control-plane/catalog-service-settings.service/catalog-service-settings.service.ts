/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { catalogServiceApi } from "@/services/infra/client-apis/catalog-service-api/catalog-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import {
  CatalogSettingsKeyEnum,
  type CatalogSettingsGetResponse,
  type GetEnableNgcIntegration,
} from "@/swagger-models/catalog-service-client";
import type { IEnableNgcIntegrationSetting } from "@/models/setting.model";

export interface ICatalogServiceSettings {
  enableNgcIntegration: IEnableNgcIntegrationSetting;
}

export const catalogServiceSettingsService = {
  getAll,
  getEnableNgcIntegration,
  updateEnableNgcIntegration,
};

async function getAll(): Promise<ICatalogServiceSettings> {
  try {
    const response = await catalogServiceApi.settingsApi.getCatalogSettings();
    const data = response.data as CatalogSettingsGetResponse;

    return {
      enableNgcIntegration: {
        value: Boolean(data.enableNgcIntegration?.value?.enabled),
        displayName: data.enableNgcIntegration?.displayName,
        description: data.enableNgcIntegration?.description,
        stage: data.enableNgcIntegration?.stage,
      },
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getEnableNgcIntegration(): Promise<IEnableNgcIntegrationSetting> {
  try {
    const response = await catalogServiceApi.settingsApi.getCatalogSettingByKey(
      CatalogSettingsKeyEnum.EnableNgcIntegration,
    );
    const data = response.data as GetEnableNgcIntegration;
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

async function updateEnableNgcIntegration(enabled: boolean): Promise<void> {
  try {
    await catalogServiceApi.settingsApi.updateCatalogSettingByKey(CatalogSettingsKeyEnum.EnableNgcIntegration, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
