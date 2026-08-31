/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  Configuration,
  PVCsApi,
  type ConfigurationParameters,
  SecretsApi,
  ConfigMapsApi,
  StorageClassesApi,
  StorageClassConfigurationApi,
} from "@/swagger-models/k8s-objects-tracker-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class K8sObjectTrackerApi {
  configuration: Configuration;

  pvcApi: PVCsApi;
  secretsApi: SecretsApi;
  configMapsApi: ConfigMapsApi;
  storageClassesApi: StorageClassesApi;
  storageClassConfigurationApi: StorageClassConfigurationApi;
  constructor() {
    this.configuration = new Configuration();

    this.pvcApi = new PVCsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.secretsApi = new SecretsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.storageClassesApi = new StorageClassesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.configMapsApi = new ConfigMapsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.storageClassConfigurationApi = new StorageClassConfigurationApi(
      this.configuration,
      this.baseUrl,
      authenticatedAxios,
    );
  }
  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const k8sObjectTrackerApi = new K8sObjectTrackerApi();
