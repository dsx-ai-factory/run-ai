/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { AdminCLIApi, NodeTypesApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class BackendUndocumentedApi {
    adminCliApi: AdminCLIApi;

    nodeTypesApi: NodeTypesApi;

    constructor(url: string) {
        this.adminCliApi = new AdminCLIApi(url);
        this.adminCliApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.nodeTypesApi = new NodeTypesApi(url);
        this.nodeTypesApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
