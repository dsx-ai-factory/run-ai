/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { HealthApi, NotificationChannelsApi, EmailApi, NotificationStateApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class NotificationsProxyApi {
    healthApi: HealthApi;

    notificationsChannelApi: NotificationChannelsApi;

    emailApi: EmailApi;

    notificationStateApi: NotificationStateApi;

    constructor(url: string) {
        this.healthApi = new HealthApi(url);
        this.healthApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.notificationsChannelApi = new NotificationChannelsApi(url);
        this.notificationsChannelApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.emailApi = new EmailApi(url);
        this.emailApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.notificationStateApi = new NotificationStateApi(url);
        this.notificationStateApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
