/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { clusterUtils, config, kubeConfigUtils, Utils } from '../index';
import { Cluster } from '../../api/apis/backendClusterApi';
import { setKubeCli } from '../../cli';
import { backendClusterApi } from '../../api';
import { isEmpty } from 'lodash-es';

export class MultiClusterUtils {
    numberOfClusters = 3;

    getClusterPrefixNameByNum(clusterNum: number): string {
        return config.clusterNamePrefix.replace(/\d$/, clusterNum.toString());
    }

    async getClusterDetails(clusterNamePrefix = config.clusterNamePrefix): Promise<Cluster> {
        const clusters = (await backendClusterApi.clusterApi.v1K8sClustersGet()).body;
        const filteredClusters = clusters.filter((cls) => !isEmpty(cls.version));
        return filteredClusters.find((cls) => Utils.strStartsWith(cls.name, clusterNamePrefix));
    }

    async switchClusterKubeConfig(clusterPrefix: string): Promise<void> {
        const cluster = await this.getClusterDetails(clusterPrefix);
        const clusterNamePrefix = clusterUtils.getClusterPrefix(cluster.name);
        await kubeConfigUtils.setContextAndSave(clusterNamePrefix);
        setKubeCli(clusterNamePrefix);
    }

    async getAllClustersDetails(): Promise<Map<string, Cluster>> {
        const clusters = new Map<string, Cluster>();
        for (let i = 1; i <= this.numberOfClusters; i++) {
            const clusterPrefix = this.getClusterPrefixNameByNum(i);
            clusters.set(clusterPrefix, await this.getClusterDetails(clusterPrefix));
        }
        return clusters;
    }
}
