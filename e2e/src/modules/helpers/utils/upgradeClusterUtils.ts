/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { backendClusterApi } from '../../api';
import { Cluster, CreateCluster } from '../../api/apis/backendClusterApi';
import { flowRootState } from '../config/config';

export enum ClusterPlatform {
    OnPremises = 'op',
    AMAZON = 'aws',
    AZURE = 'aks',
    OPENSHIFT = 'openshift',
    Google = 'gcp'
}
export class UpgradeClusterUtils {
    async createClusterByName(
        clusterName: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Cluster> {
        backendClusterApi.clusterApi.accessToken = accessToken;
        const createCluster: CreateCluster = {
            name: clusterName
        };
        return (await backendClusterApi.clusterApi.createCluster(createCluster)).body;
    }
}
