/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  ComputeApi,
  CredentialsApi,
  DatasourcesApi,
  EnvironmentApi,
  GitApi,
  HostPathApi,
  ModelApi,
  ModelInferenceApi,
  NFSApi,
  PVCApi,
  S3Api,
  SecretApi,
  SynchronizationApi,
  TemplateApi,
  TrainingsApi,
  WorkloadApi,
  WorkspacesApi,
  Configuration,
  ConfigMapApi,
  InferencesApi,
  AssetsApi,
  type ConfigurationParameters,
} from "@/swagger-models/assets-service-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class AssetsServiceApi {
  configuration: Configuration;

  environmentApi: EnvironmentApi;
  dataSourceApi: DatasourcesApi;
  computeApi: ComputeApi;
  credentialApi: CredentialsApi;
  templateApi: TemplateApi;
  workloadApi: WorkloadApi;
  workspaceApi: WorkspacesApi;
  trainingApi: TrainingsApi;
  s3Api: S3Api;
  pvcApi: PVCApi;
  nfsApi: NFSApi;
  hostPathApi: HostPathApi;
  gitApi: GitApi;
  secretApi: SecretApi;
  modelInferenceApi: ModelInferenceApi;
  modelApi: ModelApi;
  synchronizationApi: SynchronizationApi;
  configMapApi: ConfigMapApi;
  inferenceApi: InferencesApi;
  assetsApi: AssetsApi;

  constructor() {
    this.configuration = new Configuration();

    this.environmentApi = new EnvironmentApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.dataSourceApi = new DatasourcesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.computeApi = new ComputeApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.credentialApi = new CredentialsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.templateApi = new TemplateApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.workloadApi = new WorkloadApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.workspaceApi = new WorkspacesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.trainingApi = new TrainingsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.s3Api = new S3Api(this.configuration, this.baseUrl, authenticatedAxios);
    this.pvcApi = new PVCApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.nfsApi = new NFSApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.hostPathApi = new HostPathApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.gitApi = new GitApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.secretApi = new SecretApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.modelInferenceApi = new ModelInferenceApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.modelApi = new ModelApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.synchronizationApi = new SynchronizationApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.configMapApi = new ConfigMapApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.inferenceApi = new InferencesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.assetsApi = new AssetsApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const assetsServiceApi = new AssetsServiceApi();
