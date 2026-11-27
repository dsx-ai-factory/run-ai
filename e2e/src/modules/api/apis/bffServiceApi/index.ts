/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { HealthApi } from './generated/api/healthApi';
import { SettingsApi } from './generated/api/settingsApi';
import { SystemNotificationApi } from './generated/api/systemNotificationApi';
import { VersionApi } from './generated/api/versionApi';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class BffServiceApi {
    healthApi: HealthApi;

    settingsApi: SettingsApi;

    systemNotificationApi: SystemNotificationApi;

    versionApi: VersionApi;

    constructor(url: string) {
        this.healthApi = new HealthApi(url);
        this.healthApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.settingsApi = new SettingsApi(url);
        this.settingsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.systemNotificationApi = new SystemNotificationApi(url);
        this.systemNotificationApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.versionApi = new VersionApi(url);
        this.versionApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
