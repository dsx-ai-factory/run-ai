/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { WorkloadsV2Api, NVIDIANIMApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export class WorkloadsManagerService {
    workloadsV2Api: WorkloadsV2Api;

    nimServicesApi: NVIDIANIMApi;

    constructor(url: string) {
        this.workloadsV2Api = new WorkloadsV2Api(url);
        this.workloadsV2Api.defaultHeaders = API_DEFAULT_HEADERS;

        this.nimServicesApi = new NVIDIANIMApi(url);
        this.nimServicesApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
