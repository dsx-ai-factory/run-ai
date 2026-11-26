/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger, config, flowRootState, BackendApiError, UtilsError, settings, apiUtils } from '..';
import { kubeCli } from '../../cli';
import axios from 'axios';
import { Cluster } from '../../api/apis/backendClusterApi';
import gitBranch from 'git-branch';

export class InfoUtils {
    getControlPlaneVersion = async (accessToken?: string): Promise<string> => {
        const url = `${config.tenantUrl}/_version`;
        logger.debug(`get cp version from url ${url}`);
        try {
            const token = accessToken ?? flowRootState.users?.system_administrator?.accessToken;
            const res = await axios.get(url, token ? { headers: { Authorization: `Bearer ${token}` } } : {}).catch((error) => {
                throw new BackendApiError(error);
            });
            logger.debug(`cp version is ${res.data.version}`);
            flowRootState.controlPlaneVersion = res.data.version;
            return res.data.version;
        } catch (error) {
            logger.error('error while trying to fetch control plane version ', error);
            throw new UtilsError(error);
        }
    };

    getCluster = async (clusterNamePrefix?: string, accessToken?: string): Promise<Cluster> => {
        return flowRootState.cluster
            ? flowRootState.cluster
            : await apiUtils.getClusterByPrefix(clusterNamePrefix, accessToken);
    };

    getAutomationBranch = async (): Promise<string> => {
        try {
            let automationVersion = await gitBranch();
            if (automationVersion === null) {
                automationVersion = process.env.GITHUB_HEAD_REF;
            }
            return automationVersion;
        } catch (error) {
            logger.error('error while trying to fetch automation version by branch ', error);
        }
    };

    printInitSummary = async (): Promise<void> => {
        logger.info(`
    Versions Summary Details:
    - Current Flow uuid: ${flowRootState.currentFlowUuid}
    - Automation Branch: ${await this.getAutomationBranch()}
    - Control Plane: ${await this.getControlPlaneVersion()}
    - Cluster: ${flowRootState.cluster.version}
    - Runai CLI: ${flowRootState.cliVersion}
    - GPU Type: ${settings.gpuType}
    - Cluster Name (backend): ${flowRootState.cluster.name}
    - Cluster URL (backend): ${config.tenantUrl}
    - Cluster Context (kubeconfig): ${kubeCli.getCurrentContext()}
    `);
    };
}
