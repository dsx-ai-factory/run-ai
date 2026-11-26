/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { TenantsApi, LogoApi, TenantApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export class TenantsManagerApi {
    tenantsApi: TenantsApi;

    logoApi: LogoApi;

    tenantApi: TenantApi;

    constructor(url: string) {
        this.tenantsApi = new TenantsApi(url);
        this.tenantsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.logoApi = new LogoApi(url);
        this.logoApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.tenantApi = new TenantApi(url);
        this.tenantApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
