/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  Configuration,
  TenantsApi,
  TenantApi,
  LogoApi,
  type ConfigurationParameters,
} from "@/swagger-models/tenants-manager-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class TenantManagerApi {
  configuration: Configuration;
  tenantsApi: TenantsApi;
  tenantApi: TenantApi;
  logoApi: LogoApi;

  constructor() {
    this.configuration = new Configuration();

    this.tenantsApi = new TenantsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.tenantApi = new TenantApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.logoApi = new LogoApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const tenantManagerApi = new TenantManagerApi();
