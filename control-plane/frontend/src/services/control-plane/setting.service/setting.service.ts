/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { SettingKeys, TSettingValue } from "@/models/setting.model";

import { backendApi } from "@/services/infra/client-apis/backend-api/backend-api";
import { httpService } from "@/services/infra/https.service/http.service";
import type { TenantSettingCreationResponse } from "@/swagger-models/temp-backend-settings-client";
import type { Setting } from "@/models/setting.model";

export const settingService = {
  get,
  update,
  getByKey,
};

async function get(excludeMigrated = true): Promise<Map<string, Setting>> {
  try {
    const settings = await backendApi.tenantApi.getTenantSettings(excludeMigrated);
    return settings.data.reduce((acc: Map<string, Setting>, item: Setting) => {
      acc.set(item.key as string, item);
      return acc;
    }, new Map());
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function update(key: SettingKeys, value: TSettingValue): Promise<TenantSettingCreationResponse> {
  try {
    const response = await backendApi.tenantApi.updateTenantSetting(undefined, { key, value });
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getByKey(key: SettingKeys): Promise<TSettingValue | undefined> {
  try {
    const response = await backendApi.tenantApi.getTenantSettingsByKey(key);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
