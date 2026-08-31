/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  Configuration,
  WorkloadsApi,
  type ConfigurationParameters,
  PodsApi,
  EventsApi,
  DistributedApi,
  TrainingsApi,
  WorkspacesApi,
  InferencesApi,
  WorkloadsBatchApi,
  UnifiedApi,
  WorkloadTemplatesApi,
  WorkloadPropertiesApi,
  DistributedInferencesApi,
  AIApplicationsApi,
  SettingsApi,
} from "@/swagger-models/workloads-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class WorkloadServiceApi {
  configuration: Configuration;

  eventsApi: EventsApi;
  podsApi: PodsApi;
  workloadsApi: WorkloadsApi;
  distributedApi: DistributedApi;
  trainingsApi: TrainingsApi;
  workspacesApi: WorkspacesApi;
  inferencesApi: InferencesApi;
  distributedInferencesApi: DistributedInferencesApi;
  workloadsBatchApi: WorkloadsBatchApi;
  unifiedApi: UnifiedApi;
  workloadTemplatesApi: WorkloadTemplatesApi;
  workloadPropertiesApi: WorkloadPropertiesApi;
  aiApplicationsApi: AIApplicationsApi;
  settingsApi: SettingsApi;

  constructor() {
    this.configuration = new Configuration();

    this.eventsApi = new EventsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.podsApi = new PodsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.workloadsApi = new WorkloadsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.workloadsBatchApi = new WorkloadsBatchApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.distributedApi = new DistributedApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.trainingsApi = new TrainingsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.workspacesApi = new WorkspacesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.inferencesApi = new InferencesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.distributedInferencesApi = new DistributedInferencesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.unifiedApi = new UnifiedApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.workloadTemplatesApi = new WorkloadTemplatesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.workloadPropertiesApi = new WorkloadPropertiesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.aiApplicationsApi = new AIApplicationsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.settingsApi = new SettingsApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const workloadServiceApi = new WorkloadServiceApi();
