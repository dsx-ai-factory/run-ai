/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { backendClusterApi } from '../../../../../modules/api';
import { kubeCli } from '../../../../../modules/cli';
import {
    authUtils,
    clusterUtils,
    config,
    logger,
    runaiHealthChecker,
    settings,
    supportedPlatforms,
    tags,
    waitUtils
} from '../../../../../modules/helpers';
import { runaiV1 } from '../../../../../modules/ui';

test.describe('Environment Health Check', () => {
    test.beforeAll('skipper', () => {
        if (process.env.CONTROL_PLANE_URL.startsWith('https://multi-')) {
            logger.info(`Skipping test due to multitenancy + ${config.tenantUrl}`);
            test.skip();
        }
    });

    test.describe(`Nodes Check ${tags.teams.phoenix}`, () => {
        test('should check that nodes condition ready', async () => {
            const nodesCheck = await runaiHealthChecker.getNodeHealthStatus();
            expect(nodesCheck).toBeTruthy();
            expect(nodesCheck.length).toBeGreaterThan(0);
            const nodes = await kubeCli.CoreV1Api.listNode();
            expect(nodesCheck.length).toBe(nodes.items.length);
        });

        test('should validate available gpu in the cluster', async () => {
            await waitUtils.waitForCondition(
                async () => (await kubeCli.getNumOfGpus()) > 0,
                'wait for gpus to be available in the cluster',
                true,
                300
            );
        });
    });

    test.describe('Control Plane, Cluster', () => {
        test(`should check that pods by ns runai condition ready ${tags.teams.phoenix}`, async () => {
            const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('runai');
            expect(podByNs).toBe(true);
            logger.debug(`runai pods: ${JSON.stringify(podByNs)}`);
        });

        test(`should check that pods by ns runai-backend condition ready ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, async () => {
            if (await runaiHealthChecker.isSelfHostedCluster()) {
                const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('runai-backend');
                expect(podByNs).toBe(true);
                logger.debug(`runai-backend pods: ${JSON.stringify(podByNs)}`);
            }
        });

        test(`should check that pods by ns gpu-operator condition ready ${tags.teams.phoenix}`, async () => {
            const gpuOperatorNamespace = await runaiHealthChecker.resolveNamespaceByContains('gpu-operator');
            const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace(gpuOperatorNamespace);
            expect(podByNs).toBe(true);
            logger.debug(`gpu-operator pods: ${JSON.stringify(podByNs)}`);
        });

        test(`should check that pods by ns monitoring condition ready ${tags.teams.phoenix}`, async () => {
            if (settings.platform !== supportedPlatforms.ocp) {
                const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('monitoring');
                expect(podByNs).toBe(true);
                logger.debug(`gpu-operator pods: ${JSON.stringify(podByNs)}`);
            } else {
                test.skip();
            }
        });

        test(`should validate Operator status ${tags.teams.phoenix}`, async ({}, testInfo) => {
            if (await runaiHealthChecker.isClusterVersionHaveOperator()) {
                await runaiHealthChecker.waitForRetryBackoff(testInfo.retry);
                await runaiHealthChecker.validateOperatorStatusConditions();
                logger.debug('Operator status is healthy');
            } else {
                logger.debug('Operator runai-public not supported under 2.9');
                test.skip();
            }
        });
    });

    test.describe(`UI health check tests ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, () => {
        test.beforeEach('start browser', async () => {
            await runaiV1.cluster.start();
        });

        test.afterEach('close browser', async () => {
            try {
                await runaiV1.cluster.stopAndArchive();
            } catch (error) {
                logger.error('Failed to stop browser' + error);
            }
        });

        test('should login admin by the UI', async () => {
            await runaiV1.cluster.redirectAndSignIn(
                config.controlPlaneAdminUsername,
                config.controlPlaneAdminPassword,
                true,
                false
            );
        });
    });

    test.describe(`API health check ${tags.teams.phoenix}`, () => {
        test('should cluster be connected', async () => {
            const adminTokens = await authUtils.authenticateUser(
                config.controlPlaneAdminUsername,
                config.controlPlaneAdminPassword
            );
            const clusterUuid = (await clusterUtils.list(adminTokens.idToken)).filter((cluster) =>
                cluster.name.includes(config.clusterNamePrefix)
            );
            const clusterStatus = await clusterUtils.getClusterStatus(adminTokens.idToken, clusterUuid[0].uuid);
            expect(clusterStatus).toBe('Connected');
        });

        test('should validate get cluster api', async () => {
            const adminTokens = await authUtils.authenticateUser(
                config.controlPlaneAdminUsername,
                config.controlPlaneAdminPassword
            );
            backendClusterApi.clusterApi.accessToken = adminTokens.idToken;
            const clusters = (await backendClusterApi.clusterApi.v1K8sClustersGet()).body;
            if (await runaiHealthChecker.isSelfHostedCluster()) {
                expect(clusters.length).toBeGreaterThan(0);
                logger.debug(`clusters: ${JSON.stringify(clusters)}`);
            } else {
                expect(clusters.length).toBeGreaterThan(0);
                logger.debug(`clusters: ${JSON.stringify(clusters)}`);
            }
        });
    });
});
