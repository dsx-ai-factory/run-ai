/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    NodesApi,
    NodePoolsApi,
    ClustersApi,
    HealthApi,
    SettingsApi,
    ControlPlaneConfigApi
} from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class ClusterApi {
    healthApi: HealthApi;

    nodesApi: NodesApi;

    nodepoolsApi: NodePoolsApi;

    clustersApi: ClustersApi;

    settingsApi: SettingsApi;

    controlPlaneConfigApi: ControlPlaneConfigApi;

    constructor(url: string) {
        this.healthApi = new HealthApi(url);
        this.healthApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.nodesApi = new NodesApi(url);
        this.nodesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.nodepoolsApi = new NodePoolsApi(url);
        this.nodepoolsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.clustersApi = new ClustersApi(url);
        this.clustersApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.settingsApi = new SettingsApi(url);
        this.settingsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.controlPlaneConfigApi = new ControlPlaneConfigApi(url);
        this.controlPlaneConfigApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
