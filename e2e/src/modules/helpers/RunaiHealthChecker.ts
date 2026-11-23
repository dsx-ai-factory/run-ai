/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { apiUtils, authUtils, infoUtils } from '.';
import { kubeCli } from '../cli';
import { config, flowRootState } from './config/config';
import { logger } from './logger/logger';
import { expect } from 'chai';

const sleep = (ms: number): Promise<void> => new Promise((resolve) => setTimeout(resolve, ms));

export class RunaiHealthChecker {
    async getNodeHealthStatus(): Promise<string[]> {
        const nodes = (await kubeCli.CoreV1Api.listNode()).items;
        return nodes
            .filter((node) => node.status?.conditions.find((condition) => condition.type === 'Ready'))
            .map((node) => node.metadata?.name);
    }

    async getPodsByNamespace(namespace: string): Promise<string[]> {
        const res = await kubeCli.CoreV1Api.listNamespacedPod({ namespace });
        const pods = res.items;
        const podNames = [];
        pods.forEach((pod) => {
            podNames.push(pod?.metadata?.name);
        });
        return podNames;
    }

    async resolveNamespaceByContains(namespaceSubstring: string): Promise<string> {
        const namespaces = (await kubeCli.CoreV1Api.listNamespace()).items;
        const match = namespaces.find((namespace) => namespace.metadata?.name?.includes(namespaceSubstring));
        return match?.metadata?.name ?? namespaceSubstring;
    }

    async validatePodHealthStatusByNamespace(namespace: string): Promise<boolean> {
        const res = await kubeCli.CoreV1Api.listNamespacedPod({ namespace });
        const pods = res.items;
        if (pods.length === 0) {
            logger.info('No pods are found');
            return false;
        }
        let isHealthy = true;
        pods.forEach((pod) => {
            const podName = pod?.metadata?.name;
            const podStatus = pod?.status?.phase;
            const podConditions = pod?.status?.conditions;
            const podStatusValidation = ['Running', 'Completed', 'Succeeded'].includes(podStatus);
            if (!podStatusValidation) {
                logger.error(`Pod ${podName} is not in a healthy status - Current status: ${podStatus}`);
                isHealthy = false;
            }
            const condition = podConditions.find(
                (condition) =>
                    condition.type === 'Ready' &&
                    condition.status !== 'True' &&
                    condition.reason !== 'PodCompleted'
            );
            if (condition) {
                logger.error(
                    `Pod ${podName} is not in a healthy status, due to condition ${condition.type} with status ${condition.status} and reason ${condition.reason}`
                );
                isHealthy = false;
            }
        });
        logger.debug(`Pods health status is ${isHealthy} in namespace ${namespace}`);
        return isHealthy;
    }

    async validateOperatorStatusConditions(consecutiveFailuresToFail = 3, intervalMs = 3000): Promise<void> {
        let lastError: unknown;
        for (let attempt = 1; attempt <= consecutiveFailuresToFail; attempt++) {
            if (attempt > 1) {
                await sleep(intervalMs);
            }
            try {
                await this.checkOperatorStatusConditionsOnce();
                if (attempt > 1) {
                    logger.info(
                        `operator status check passed on attempt ${attempt}/${consecutiveFailuresToFail} after prior failure(s)`
                    );
                }
                return;
            } catch (err) {
                lastError = err;
                logger.warn(
                    `operator status check failed on attempt ${attempt}/${consecutiveFailuresToFail}: ${err}`
                );
            }
        }
        throw lastError;
    }

    private async checkOperatorStatusConditionsOnce(): Promise<void> {
        const res = await kubeCli.CoreV1Api.readNamespacedConfigMap({
            name: 'runai-public',
            namespace: 'runai'
        });
        const runaiPublic = res.data['runai-public'];
        const types = runaiPublic.split('conditions:')[1].split('optional:')[0].split('- type:');
        logger.debug('going to check the conditions in configmap runai-public');

        const deployedCondition = types.find((t) => t.includes('Deployed'));
        logger.debug(deployedCondition);
        expect(deployedCondition).to.include('Deployed');
        expect(deployedCondition).to.include('status: "True"');
        expect(deployedCondition).to.include('reason: deployed');

        const availableCondition = types.find((t) => t.includes('Available'));
        logger.debug(availableCondition);
        expect(availableCondition).to.include('Available');
        expect(availableCondition).to.include('status: "True"');
        expect(availableCondition).to.include('reason: available');
        expect(availableCondition).to.include('message: System Available');
    }

    async waitForRetryBackoff(retry: number, baseMs = 5000): Promise<void> {
        if (retry <= 0) return;
        const waitMs = baseMs * 2 ** (retry - 1);
        logger.info(`Playwright retry ${retry}: backing off ${waitMs}ms before health check`);
        await sleep(waitMs);
    }

    async isSelfHostedCluster(): Promise<boolean> {
        try {
            const runai = await kubeCli.CoreV1Api.readNamespace({ name: 'runai' });
            const runai_backend = await kubeCli.CoreV1Api.readNamespace({ name: 'runai-backend' });
            if (runai_backend.metadata.name === 'runai-backend' && runai.metadata.name === 'runai')
                return true;
        } catch (result) {
            logger.debug(`Cluster is not self hosted: ${result}`);
            return false;
        }
    }

    async isClusterVersionHaveOperator(): Promise<boolean> {
        const adminTokens = await authUtils.authenticateUser(
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword
        );
        flowRootState.cluster = await apiUtils.getClusterByPrefix(
            config.clusterNamePrefix,
            adminTokens.idToken
        );
        const cluster = await infoUtils.getCluster(config.clusterNamePrefix, adminTokens.idToken);
        logger.debug('cluster version is: ' + cluster.version);
        if (cluster.version.startsWith('2.7') || cluster.version.startsWith('2.8')) {
            return false;
        } else {
            return true;
        }
    }
}
