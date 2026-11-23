/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { RolesApi } from './generated/api/rolesApi';
import { AccessRulesApi } from './generated/api/accessRulesApi';
import { PermissionsApi } from './generated/api/permissionsApi';
import { SettingsApi } from './generated/api/settingsApi';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export class AuthorizationApi {
    rolesApi: RolesApi;

    accessRulesApi: AccessRulesApi;

    permissionsApi: PermissionsApi;

    settingsApi: SettingsApi;

    constructor(url: string) {
        this.rolesApi = new RolesApi(url);
        this.rolesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.accessRulesApi = new AccessRulesApi(url);
        this.accessRulesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.permissionsApi = new PermissionsApi(url);
        this.permissionsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.settingsApi = new SettingsApi(url);
        this.settingsApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
