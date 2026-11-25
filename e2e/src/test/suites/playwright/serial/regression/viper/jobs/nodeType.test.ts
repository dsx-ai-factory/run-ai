/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    BaseWorkload,
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    PodStatus,
    projectResourcesUtils,
    tags,
    TrainingWorkload,
    Utils,
    waitUtils,
    workloadUtils
} from '../../../../../../../modules/helpers';

import { kubeCli } from '../../../../../../../modules/cli';
import { test, expect } from '@playwright/test';
import { shouldRunNativeWorkloadCrdTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

shouldRunNativeWorkloadCrdTests(() => {
test.describe(`Test node type, tags: ${tags.resource.scheduler} ${tags.teams.viper}`, () => {
    let nodeName: string, projectName: string, namespace: string, generatedNodeType: string;

    test.beforeAll(async () => {
        const compatibleNodes = flowRootState.nodesResources.filter((node) => node.gpu >= 1);
        if (compatibleNodes.length < 2) {
            logger.warn('Not enough nodes with GPUs in cluster to run this test');
            test.skip();
        }

        generatedNodeType = objectGenerator.generateNameByPrefix('node-type');
        nodeName = compatibleNodes[0].name;
        await kubeCli.labelNode(nodeName, 'run.ai/type', generatedNodeType);

        projectName = objectGenerator.generateNameByPrefix('node-type');
        namespace = Utils.namespaceFromProjectName(projectName);
        await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            maxAllowedGpus: 1,
            deservedGpus: 1
        });
    });

    test.afterAll(async () => {
        if (nodeName) {
            try {
                await kubeCli.unLabelNode(nodeName, 'run.ai/type');
            } catch (error) {
                logger.error(`Failed removing label 'run.ai/type' from node "${nodeName}"`);
            }
        }
        try {
            await operations.forceCleanup();
        } catch (error) {
            logger.error('Failed to cleanup after test');
        }
    });

    test.describe('single job', () => {
        let job1: BaseWorkload;

        async function submitJob(expectedStatus: PodStatus): Promise<BaseWorkload> {
            const job = new TrainingWorkload(config.quickStartImage.ubuntu, namespace);

            job.gpu = '1';
            job.command = 'sleep infinity';
            job.nodeType = generatedNodeType;

            await workloadUtils.submit(job);
            await waitUtils.waitForWorkloadStatus(job, expectedStatus);

            return job;
        }

        test.beforeAll(async () => {
            job1 = await submitJob(PodStatus.RUNNING);
        });

        test('should schedule job on the correct node', async () => {
            const job1Name = job1.name;
            const job1Pods = await kubeCli.listPodsOfJob(namespace, job1Name);
            const job1Pod = job1Pods[0];
            expect(job1Pod.spec.nodeName).toBe(nodeName);
        });
    });
});
});
