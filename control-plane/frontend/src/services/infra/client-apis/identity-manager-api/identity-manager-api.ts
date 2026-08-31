/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  Configuration,
  ClientsApi,
  IdpsApi,
  RealmApi,
  TokensApi,
  UsersApi,
  type ConfigurationParameters,
  ConfigApi,
  AccessKeysApi,
  SettingsApi,
  ServiceAccountsApi,
} from "@/swagger-models/identity-manager-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class IdentityManagerApi {
  configuration: Configuration;

  clientsApi: ClientsApi;
  idpsApi: IdpsApi;
  realmApi: RealmApi;
  tokensApi: TokensApi;
  usersApi: UsersApi;
  configApi: ConfigApi;
  accessKeysApi: AccessKeysApi;
  settingsApi: SettingsApi;
  serviceAccountsApi: ServiceAccountsApi;

  constructor() {
    this.configuration = new Configuration();

    this.clientsApi = new ClientsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.idpsApi = new IdpsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.realmApi = new RealmApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.tokensApi = new TokensApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.usersApi = new UsersApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.configApi = new ConfigApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.accessKeysApi = new AccessKeysApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.settingsApi = new SettingsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.serviceAccountsApi = new ServiceAccountsApi(this.configuration, this.baseUrl, authenticatedAxios);
  }
  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const identityManagerApi = new IdentityManagerApi();
