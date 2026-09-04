/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IConfig } from "@/models/config.model";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";

import type { AuthConfiguration } from "@/swagger-models/identity-manager-client";

export const configService = {
  get,
};

async function get(): Promise<IConfig> {
  const response = await identityManagerApi.configApi.getTenantConfiguration();
  const config: AuthConfiguration = response.data;
  config.authHost = config.authHost + "/auth/realms/" + config.tenantName;
  return config;
}
