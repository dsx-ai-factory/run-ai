/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { PolicyApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class PolicyApiHandler {
    policyApi: PolicyApi;

    constructor(url: string) {
        this.policyApi = new PolicyApi(url);
        this.policyApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
