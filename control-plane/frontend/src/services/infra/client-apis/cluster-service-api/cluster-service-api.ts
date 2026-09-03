/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  Configuration,
  type ConfigurationParameters,
  ClustersApi,
  NodesApi,
  NodePoolsApi,
  NetworkTopologiesApi,
  SettingsApi,
} from "@/swagger-models/cluster-service-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class ClusterServiceApi {
  configuration: Configuration;

  clustersApi: ClustersApi;
  nodesApi: NodesApi;
  nodePoolsApi: NodePoolsApi;
  networkTopologiesApi: NetworkTopologiesApi;
  settingsApi: SettingsApi;

  constructor() {
    this.configuration = new Configuration();

    this.clustersApi = new ClustersApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.nodesApi = new NodesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.nodePoolsApi = new NodePoolsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.networkTopologiesApi = new NetworkTopologiesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.settingsApi = new SettingsApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const clusterServiceApi = new ClusterServiceApi();
