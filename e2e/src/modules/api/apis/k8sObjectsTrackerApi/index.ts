/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { HealthApi } from './generated/api/healthApi';
import { PVCsApi } from './generated/api/pVCsApi';
import { SecretsApi } from './generated/api/secretsApi';
import { ConfigMapsApi } from './generated/api/configMapsApi';
import { StorageClassesApi } from './generated/api/storageClassesApi';
import { StorageClassConfigurationApi } from './generated/api/storageClassConfigurationApi';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export class k8sObjectsTrackerApi {
    healthApi: HealthApi;

    pvcsApi: PVCsApi;

    secretsApi: SecretsApi;

    configMapsApi: ConfigMapsApi;

    storageClassesApi: StorageClassesApi;

    storageClassConfigurationApi: StorageClassConfigurationApi;

    constructor(url: string) {
        this.healthApi = new HealthApi(url);
        this.healthApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.pvcsApi = new PVCsApi(url);
        this.pvcsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.secretsApi = new SecretsApi(url);
        this.secretsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.configMapsApi = new ConfigMapsApi(url);
        this.configMapsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.storageClassesApi = new StorageClassesApi(url);
        this.storageClassesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.storageClassConfigurationApi = new StorageClassConfigurationApi(url);
        this.storageClassConfigurationApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
