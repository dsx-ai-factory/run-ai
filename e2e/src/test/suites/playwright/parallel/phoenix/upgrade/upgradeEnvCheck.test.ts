/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { backendClusterApi } from '../../../../../../modules/api';
import { kubeCli } from '../../../../../../modules/cli';
import {
    clusterUtils,
    config,
    infoUtils,
    logger,
    runaiHealthChecker,
    settings,
    supportedPlatforms,
    versionUtils,
    waitUtils,
    tags,
    upgradesHealthChecker,
    upgradeUtils,
    ensureControlPlaneCredentials
} from '../../../../../../modules/helpers';
import { runaiV1 } from '../../../../../../modules/ui';
import { shouldRunTests } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    getRunaiRemoteWrites,
    monitoringConfigWithoutRunaiRemoteWrites
} from '../../../../../../modules/helpers/utils/OpenShiftMonitoringConfigUtils';

const monitoringConfigNamespace = 'openshift-monitoring';
const monitoringConfigName = 'cluster-monitoring-config';
const monitoringSnapshotNamespace = 'default';

const monitoringSnapshotName = (): string => {
    const runId = process.env.GITHUB_RUN_ID || config.clusterNamePrefix;
    return `runai-e2e-monitoring-${runId}`
        .toLowerCase()
        .replace(/[^a-z0-9-]/g, '-')
        .slice(0, 63);
};

test.describe(`Environment Health Check: ${tags.runOn.upgradeHealthCheck} ${tags.teams.phoenix}`, () => {
        test.skip(!settings.isUpgradesTests, 'Upgrade health checks only run in upgrade environments');

        test.beforeAll(async ({}, testInfo) => {
            if (process.env.CONTROL_PLANE_URL?.startsWith('https://multi-')) {
                logger.info(`Skipping test due to multitenancy + ${config.tenantUrl}`);
                testInfo.skip();
            }
            await ensureControlPlaneCredentials();
        });

        test.describe('Nodes Check', () => {
            test('should check that nodes condition ready', async () => {
                const nodesCheck = await runaiHealthChecker.getNodeHealthStatus();
                expect(nodesCheck).not.toBeNull();
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
            test('should check that pods by ns runai condition ready', async () => {
                const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('runai');
                expect(podByNs).toBe(true);
            });

            shouldRunTests(settings.isSelfHosted)(() => {
                test('should check that pods by ns runai-backend condition ready', async () => {
                    const podByNs =
                        await runaiHealthChecker.validatePodHealthStatusByNamespace('runai-backend');
                    expect(podByNs).toBe(true);
                });
            });

            test('should check that pods by ns gpu-operator condition ready', async () => {
                const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('gpu-operator');
                expect(podByNs).toBe(true);
            });

            test('should check that pods by ns monitoring condition ready', async () => {
                test.skip(
                    settings.platform === supportedPlatforms.ocp,
                    'Monitoring namespace is managed by OpenShift on OCP'
                );
                const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('monitoring');
                expect(podByNs).toBe(true);
            });

            test('should validate Operator status', async ({}, testInfo) => {
                if (await upgradesHealthChecker.isClusterVersionHaveOperator()) {
                    await runaiHealthChecker.waitForRetryBackoff(testInfo.retry);
                    await runaiHealthChecker.validateOperatorStatusConditions();
                } else {
                    logger.debug('Operator runai-public not supported under 2.9');
                    testInfo.skip();
                }
            });

            test('should preserve OpenShift monitoring configuration across upgrade', async () => {
                test.skip(settings.platform !== supportedPlatforms.ocp, 'Only applicable to OpenShift');

                const monitoringConfigMap = await kubeCli.getConfigMap(
                    monitoringConfigName,
                    monitoringConfigNamespace
                );
                const configYaml = monitoringConfigMap.data?.['config.yaml'];
                if (!configYaml) {
                    throw new Error('OpenShift monitoring ConfigMap is missing config.yaml');
                }
                const runaiRemoteWrites = getRunaiRemoteWrites(configYaml);
                const currentConfig = monitoringConfigWithoutRunaiRemoteWrites(
                    configYaml
                );
                const cluster = await infoUtils.getCluster(config.clusterNamePrefix);
                const snapshotName = monitoringSnapshotName();
                let snapshot;

                try {
                    snapshot = await kubeCli.getConfigMap(snapshotName, monitoringSnapshotNamespace);
                } catch (error) {
                    if ((error as any)?.code !== 404) {
                        throw error;
                    }
                }

                if (!snapshot) {
                    await kubeCli.CoreV1Api.createNamespacedConfigMap({
                        namespace: monitoringSnapshotNamespace,
                        body: {
                            metadata: { name: snapshotName },
                            data: {
                                config: JSON.stringify(currentConfig),
                                clusterVersion: cluster.version
                            }
                        }
                    });
                    logger.info(`Saved pre-upgrade OpenShift monitoring configuration in ${snapshotName}`);
                    return;
                }

                expect(
                    currentConfig,
                    'OpenShift monitoring configuration changed outside Run:ai remote writes'
                ).toEqual(JSON.parse(snapshot.data?.config || '{}'));

                test.skip(
                    snapshot.data?.clusterVersion === cluster.version,
                    'Run:ai remote write is validated after the cluster version changes'
                );
                expect(runaiRemoteWrites, 'expected exactly one Run:ai remote write').toHaveLength(1);
                expect(runaiRemoteWrites[0].url).toMatch(/^https?:\/\/.+/);
                expect(runaiRemoteWrites[0].basicAuth).toMatchObject({
                    username: { key: 'remoteWriteUsername-runai' },
                    password: { key: 'remoteWritePassword-runai' }
                });
                expect(runaiRemoteWrites[0].tlsConfig).toBeDefined();
                expect(Array.isArray(runaiRemoteWrites[0].writeRelabelConfigs)).toBe(true);
                await kubeCli.CoreV1Api.deleteNamespacedConfigMap({
                    name: snapshotName,
                    namespace: monitoringSnapshotNamespace
                });
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

        test.describe('API health check', () => {
            test('should cluster be connected', async () => {
                const version = await infoUtils.getControlPlaneVersion();
                test.skip(
                    !versionUtils.isVersionSupported(version, '2.16'),
                    'Cluster connectivity API is only supported from version 2.16'
                );
                const adminTokens = await upgradeUtils.authenticateUser(
                    config.controlPlaneAdminUsername,
                    config.controlPlaneAdminPassword
                );
                const clusterList = await clusterUtils.list(adminTokens.idToken);
                const cluster = clusterList.find((c) => c.name.includes(config.clusterNamePrefix));
                const status = await clusterUtils.getClusterStatus(adminTokens.idToken, cluster.uuid);
                expect(status).toBe('Connected');
            });

            test('should validate get cluster api', async () => {
                const adminTokens = await upgradeUtils.authenticateUser(
                    config.controlPlaneAdminUsername,
                    config.controlPlaneAdminPassword
                );
                backendClusterApi.clusterApi.accessToken = adminTokens.idToken;
                const clusters = (await backendClusterApi.clusterApi.v1K8sClustersGet()).body;
                expect(clusters.length).toBeGreaterThan(0);
            });
        });
});
