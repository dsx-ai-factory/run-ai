/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { kubeCli } from '../../../../../../modules/cli';
import { tags } from '../../../../../../modules/helpers/config/tags';
import {
    config,
    flowRootState,
    logger,
    runaiHealthChecker,
    settings,
    supportedPlatforms,
    waitUtils,
    ensureControlPlaneCredentials
} from '../../../../../../modules/helpers';
import { runaiV1 } from '../../../../../../modules/ui';
import { shouldRunTests } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

shouldRunTests(settings.isUpgradesTests)(() => {
    test.describe(`Upgrade tests: ${tags.runOn.preUpgrade} ${tags.teams.phoenix}`, () => {
        test.beforeAll(async () => {
            await ensureControlPlaneCredentials();
        });

        test.describe('Nodes Check', () => {
            test('should check that nodes condition ready', async () => {
                const nodesCheck = await runaiHealthChecker.getNodeHealthStatus();
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

        test.describe('Pods, ns are ready', () => {
            test('should check that pods by ns runai condition ready', async () => {
                const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('runai');
                expect(podByNs).toBe(true);
                logger.debug(`runai pods: ${JSON.stringify(podByNs)}`);
            });

            shouldRunTests(settings.isSelfHosted)(() => {
                test('should check that pods by ns runai-backend condition ready', async () => {
                    const podByNs =
                        await runaiHealthChecker.validatePodHealthStatusByNamespace('runai-backend');
                    expect(podByNs).toBe(true);
                    logger.debug(`runai-backend pods: ${JSON.stringify(podByNs)}`);
                });
            });

            test('should check that pods by ns gpu-operator condition ready', async () => {
                const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('gpu-operator');
                expect(podByNs).toBe(true);
                logger.debug(`gpu-operator pods: ${JSON.stringify(podByNs)}`);
            });

            test('should check that pods by ns monitoring condition ready', async () => {
                test.skip(
                    settings.platform === supportedPlatforms.ocp,
                    'Monitoring namespace is managed by OpenShift on OCP'
                );
                const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('monitoring');
                expect(podByNs).toBe(true);
                logger.debug(`monitoring pods: ${JSON.stringify(podByNs)}`);
            });
        });

        test.describe('UI health check tests', () => {
            test.beforeEach(async () => {
                await runaiV1.cluster.start();
            });

            test.afterEach(async () => {
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
    });
});
