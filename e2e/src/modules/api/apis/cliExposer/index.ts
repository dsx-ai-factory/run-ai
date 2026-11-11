/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { CLIApi, HealthApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class CliExposer {
    cliApi: CLIApi;

    healthApi: HealthApi;

    constructor(url: string) {
        this.cliApi = new CLIApi(url);
        this.cliApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.healthApi = new HealthApi(url);
        this.healthApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
