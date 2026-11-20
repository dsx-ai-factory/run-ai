/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { HealthApi, AuditLogsApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class AuditApi {
    healthApi: HealthApi;

    auditLogsApi: AuditLogsApi;

    constructor(url: string) {
        this.healthApi = new HealthApi(url);
        this.healthApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.auditLogsApi = new AuditLogsApi(url);
        this.auditLogsApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
