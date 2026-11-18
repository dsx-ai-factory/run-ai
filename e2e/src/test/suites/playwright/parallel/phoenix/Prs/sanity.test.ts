/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { kubeCli } from '../../../../../../modules/cli';
import {
    clusterService,
    clusterUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    runaiHealthChecker,
    tags,
    waitUtils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';

test.describe('Environment sanity tests', { tag: [tags.runOn.basicCI, tags.runOn.parallelCI] }, () => {
    // Smoke Test 1: Check runai pods health
    test('should verify runai namespace pods are healthy', { tag: [tags.teams.phoenix] }, async () => {
        const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('runai');
        expect(podByNs, 'all pods in the runai namespace should be healthy').toBe(true);
        logger.info(' All runai namespace pods are healthy');
    });

    // Smoke Test 2: Check runai-backend pods health (only self-hosted)
    test(
        'should verify runai-backend namespace pods are healthy (self-hosted only)',
        { tag: [tags.teams.phoenix, tags.suites.legacySphinx] },
        async () => {
            const isSelfHosted = await runaiHealthChecker.isSelfHostedCluster();
            test.skip(!isSelfHosted, 'Not a self-hosted cluster');
            const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('runai-backend');
            expect(podByNs, 'all pods in the runai-backend namespace should be healthy').toBe(true);
            logger.info('All runai-backend namespace pods are healthy');
        }
    );

    // Smoke Test 3: Check nodes health
    test('should verify all cluster nodes are ready', { tag: [tags.teams.phoenix] }, async () => {
        const nodesCheck = await runaiHealthChecker.getNodeHealthStatus();
        expect(nodesCheck, 'cluster node health status should be returned').toBeTruthy();
        expect(nodesCheck.length, 'at least one node should be present in the cluster').toBeGreaterThan(0);

        const nodes = await kubeCli.CoreV1Api.listNode();
        expect(nodesCheck.length, 'healthy node count should match total node count in cluster').toBe(
            nodes.items.length
        );
        logger.info(`All ${nodesCheck.length} nodes are in Ready state`);
    });

    // Smoke Test 4: Check GPU operator pods
    test('should verify gpu-operator namespace pods are healthy', { tag: [tags.teams.phoenix] }, async () => {
        const podByNs = await runaiHealthChecker.validatePodHealthStatusByNamespace('gpu-operator');
        expect(podByNs, 'all pods in the gpu-operator namespace should be healthy').toBe(true);
        logger.info('All gpu-operator namespace pods are healthy');
    });

    // Smoke Test 5: Check GPUs are available in cluster
    test('should verify GPUs are available in the cluster', { tag: [tags.teams.phoenix] }, async () => {
        await waitUtils.waitForCondition(
            async () => (await kubeCli.getNumOfGpus()) > 0,
            'wait for gpus to be available in the cluster',
            true,
            300
        );
        const gpuCount = await kubeCli.getNumOfGpus();
        expect(gpuCount, 'cluster should have at least one GPU available').toBeGreaterThan(0);
        logger.info(`Cluster has ${gpuCount} GPUs available`);
    });

    // Job Test: Submit and verify a training job completes successfully
    test(
        'should submit and verify a training job completes successfully',
        { tag: [tags.teams.viper] },
        async () => {
            let project: Project;
            let workloadName: string;

            try {
                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix(`smoke-test-project-${Date.now()}`),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 1, limit: 1 } }]
                });
                logger.info(`Created temporary project: ${project.name}`);

                workloadName = `smoke-test-workload-${Date.now()}`;
                const { body: workload } = await workloadsCrudUtils.createTraining(
                    workloadName,
                    false,
                    flowRootState.cluster.uuid,
                    String(project.id),
                    { image: config.quickStartImage.ubuntu, command: 'sleep 5' }
                );
                logger.info(`Submitted workload: ${workload.name}`);

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workload.workloadId, Phase.Completed, 120);
                logger.info('Workload completed successfully');
            } finally {
                if (project && project.id) {
                    await orgUnitUtils.deleteProjectAndNamespace(
                        project.id,
                        project.name,
                        flowRootState.users.system_administrator.accessToken
                    );
                    logger.info('Cleaned up temporary project');
                }
            }
        }
    );

    // Random Test: Verify cluster connectivity and API access
    test(
        'should verify cluster is connected and API is accessible',
        { tag: [tags.teams.phoenix] },
        async () => {
            const clusters = await clusterService.getClusters(
                flowRootState.users.system_administrator.accessToken
            );
            expect(clusters, 'cluster list response should be defined').toBeDefined();
            expect(clusters.length, 'at least one cluster should be registered').toBeGreaterThan(0);
            logger.info(`Cluster API is accessible - found ${clusters.length} clusters`);
            const cluster = clusters[0];
            const clusterStatus = await clusterUtils.getClusterStatus(
                flowRootState.users.system_administrator.accessToken,
                cluster.uuid
            );
            expect(clusterStatus, 'cluster should be in Connected status').toBe('Connected');
        }
    );
});
