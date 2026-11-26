/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    ClustersApi,
    DeploymentsApi,
    JobsApi,
    NodePoolsApi,
    PodGroupsApi,
    PodsApi,
    ServicesApi
} from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class BackendClusterApi {
    clusterApi: ClustersApi;

    jobsApi: JobsApi;

    podGroupsApi: PodGroupsApi;

    podsApi: PodsApi;

    servicesApi: ServicesApi;

    nodePoolsApi: NodePoolsApi;

    constructor(url: string) {
        this.clusterApi = new ClustersApi(url);
        this.clusterApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.nodePoolsApi = new NodePoolsApi(url);
        this.nodePoolsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.jobsApi = new JobsApi(url);
        this.jobsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.podGroupsApi = new PodGroupsApi(url);
        this.podGroupsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.podsApi = new PodsApi(url);
        this.podsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.servicesApi = new ServicesApi(url);
        this.servicesApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
