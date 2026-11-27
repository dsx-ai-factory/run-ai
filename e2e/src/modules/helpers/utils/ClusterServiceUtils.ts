/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { clusterApi } from '../../api';
import {
    ClusterControlPlaneConfig,
    ClusterInstallationInfoResponse,
    DisplayedCluster
} from '../../api/apis/clusterApi';
import { ClusterInfoSyncRequest } from '../../api/apis/clusterApi/generated/model/clusterInfoSyncRequest';
import { ClusterLivenessSyncRequest } from '../../api/apis/clusterApi/generated/model/clusterLivenessSyncRequest';
export class ClusterService {
    async updateClusterLiveness(accessToken: string, clusterUuid: string): Promise<any> {
        clusterApi.clustersApi.accessToken = accessToken;
        const clusterLivenessSyncRequest: ClusterLivenessSyncRequest = {
            timeTag: Date.now()
        };
        const req = await clusterApi.clustersApi.updateClusterLiveness(
            clusterUuid,
            clusterLivenessSyncRequest
        );
        return req.response.statusCode;
    }

    async getClusterVersions(accessToken: string): Promise<any> {
        clusterApi.clustersApi.accessToken = accessToken;
        const response = await clusterApi.clustersApi.getClusterVersions();
        return response.body;
    }

    async getClusterInfo(
        accessToken: string,
        clusterUuid: string,
        version: string
    ): Promise<ClusterInstallationInfoResponse> {
        clusterApi.clustersApi.accessToken = accessToken;
        const response = await clusterApi.clustersApi.getClusterInstallInfoByUuid(clusterUuid, version);
        return response.body;
    }

    async getClusterControlPlaneConfig(
        accessToken: string,
        clusterUuid: string
    ): Promise<ClusterControlPlaneConfig> {
        clusterApi.controlPlaneConfigApi.accessToken = accessToken;
        const response = await clusterApi.controlPlaneConfigApi.getClusterControlPlaneConfig(clusterUuid);
        return response.body;
    }

    async getCluster(accessToken: string, clusterUuid: string): Promise<DisplayedCluster> {
        clusterApi.clustersApi.accessToken = accessToken;
        const response = await clusterApi.clustersApi.getClusterByUuid(clusterUuid);
        return response.body;
    }

    async getClusters(accessToken: string): Promise<Array<DisplayedCluster>> {
        clusterApi.clustersApi.accessToken = accessToken;
        const response = await clusterApi.clustersApi.getClusters();
        return response.body;
    }

    async updateCluster(accessToken: string, clusterUuid: string, cluster: DisplayedCluster): Promise<any> {
        clusterApi.clustersApi.accessToken = accessToken;
        const response = await clusterApi.clustersApi.updateCluster(clusterUuid, cluster);
        return response.response.statusCode;
    }
}
