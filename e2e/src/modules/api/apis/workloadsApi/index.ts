/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    WorkloadsApi,
    PodsApi,
    EventsApi,
    WorkspacesApi,
    TrainingsApi,
    InferencesApi,
    DistributedApi,
    WorkloadsBatchApi,
    RevisionsApi,
    DistributedInferencesApi,
    WorkloadTemplatesApi,
    WorkloadPropertiesApi,
    AIApplicationsApi,
    MetricsApi
} from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export class WorkloadsService {
    workloadsApi: WorkloadsApi;

    podsApi: PodsApi;

    eventsApi: EventsApi;

    workloadsBatchApi: WorkloadsBatchApi;

    revisionsApi: RevisionsApi;

    workloadTemplatesApi: WorkloadTemplatesApi;

    workloadPropertiesApi: WorkloadPropertiesApi;

    aiApplicationsApi: AIApplicationsApi;

    metricsApi: MetricsApi;

    constructor(url: string) {
        this.workloadsApi = new WorkloadsApi(url);
        this.workloadsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.podsApi = new PodsApi(url);
        this.podsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.eventsApi = new EventsApi(url);
        this.eventsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.workloadsBatchApi = new WorkloadsBatchApi(url);
        this.workloadsBatchApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.revisionsApi = new RevisionsApi(url);
        this.revisionsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.workloadTemplatesApi = new WorkloadTemplatesApi(url);
        this.workloadTemplatesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.workloadPropertiesApi = new WorkloadPropertiesApi(url);
        this.workloadPropertiesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.aiApplicationsApi = new AIApplicationsApi(url);
        this.aiApplicationsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.metricsApi = new MetricsApi(url);
        this.metricsApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}

export class WorkloadsCrudService {
    workspacesApi: WorkspacesApi;

    trainingsApi: TrainingsApi;

    inferencesApi: InferencesApi;

    distributedInferencesApi: DistributedInferencesApi;

    distributedApi: DistributedApi;

    constructor(url: string) {
        this.workspacesApi = new WorkspacesApi(url);
        this.workspacesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.trainingsApi = new TrainingsApi(url);
        this.trainingsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.inferencesApi = new InferencesApi(url);
        this.inferencesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.distributedInferencesApi = new DistributedInferencesApi(url);
        this.distributedInferencesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.distributedApi = new DistributedApi(url);
        this.distributedApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
