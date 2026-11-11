/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { AccessKeysApi, ConfigApi, LoginsApi, SettingsApi, UserApplicationsApi } from './generated/api';
import { ApplicationsApi } from './generated/api/applicationsApi';
import { ClientsApi } from './generated/api/clientsApi';
import { IdpsApi } from './generated/api/idpsApi';
import { RealmApi } from './generated/api/realmApi';
import { TokensApi } from './generated/api/tokensApi';
import { UsersApi } from './generated/api/usersApi';
import { ServiceAccountsApi } from './generated/api/serviceAccountsApi';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export class IdentityManagerApi {
    applicationsApi: ApplicationsApi;

    clientsApi: ClientsApi;

    idpsApi: IdpsApi;

    realmApi: RealmApi;

    tokensApi: TokensApi;

    usersApi: UsersApi;

    configApi: ConfigApi;

    loginApi: LoginsApi;

    userApplicationsApi: UserApplicationsApi;

    settingsApi: SettingsApi;

    accessKeysApi: AccessKeysApi;

    serviceAccountsApi: ServiceAccountsApi;

    constructor(url: string) {
        this.applicationsApi = new ApplicationsApi(url);
        this.applicationsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.clientsApi = new ClientsApi(url);
        this.clientsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.idpsApi = new IdpsApi(url);
        this.idpsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.realmApi = new RealmApi(url);
        this.realmApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.tokensApi = new TokensApi(url);
        this.tokensApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.usersApi = new UsersApi(url);
        this.usersApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.configApi = new ConfigApi(url);
        this.configApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.loginApi = new LoginsApi(url);
        this.loginApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.userApplicationsApi = new UserApplicationsApi(url);
        this.userApplicationsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.settingsApi = new SettingsApi(url);
        this.settingsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.accessKeysApi = new AccessKeysApi(url);
        this.accessKeysApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.serviceAccountsApi = new ServiceAccountsApi(url);
        this.serviceAccountsApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
