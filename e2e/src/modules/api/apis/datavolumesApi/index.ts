/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { DatavolumesApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export class DatavolumesService {
    datavolumesApi: DatavolumesApi;

    constructor(url: string) {
        this.datavolumesApi = new DatavolumesApi(url);
        this.datavolumesApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
