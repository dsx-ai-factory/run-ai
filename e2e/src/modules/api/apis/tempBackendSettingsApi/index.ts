/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { TenantApi } from './generated/api/tenantApi';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class TempBackendSettingsApi {
    tenantApi: TenantApi;

    constructor(url: string) {
        this.tenantApi = new TenantApi(url);
        this.tenantApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
