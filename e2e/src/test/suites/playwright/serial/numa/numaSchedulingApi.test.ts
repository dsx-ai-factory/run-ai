/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    flowRootState,
    logger,
    nodePoolUtils,
    nodeUtils,
    numaUtils,
    operations,
    settings,
    tags,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../modules/helpers';
import { shouldRunTests } from '../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { kubeCli } from '../../../../../modules/cli';
import { TopologyManagerPolicy } from '../../../../../modules/api/apis/clusterApi';
import { Phase } from '../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { Training } from '../../../../../modules/api/apis/workloadsApi/generated/model/training';
import { NumaTestNodePool, NumaUtils } from '../../../../../modules/helpers/utils/NumaUtils';

test.describe(`NUMA-aware scheduling, tags: ${tags.resource.nodepools} ${tags.tool.api} ${tags.runOn.numa} ${tags.teams.scheduler}`, () => {
    shouldRunTests(settings.isNumaWorkflow)(() => {
        let nodePool: NumaTestNodePool;
        let project;
        let nodeGpus: number;
        let zoneGpus: number;
        let numaZones: number;
        let createdWorkloads: Training[] = [];

        async function createTraining(gpus: number): Promise<Training> {
            const { body: workload } = await workloadsCrudUtils.createTraining(
                'numa',
                true,
                flowRootState.cluster.uuid,
                String(project.id),
                {
                    image: config.quickStartImage.ubuntu,
                    command: 'sleep infinity',
                    nodePools: [nodePool.nodePoolName],
                    compute: { gpuDevicesRequest: gpus }
                }
            );
            createdWorkloads.push(workload);
            return workload;
        }

        async function enableNumaAndWait(enabled: boolean): Promise<void> {
            await nodePoolUtils.setNumaAwareScheduling(
                nodePool.nodePoolId,
                nodePool.labelKey,
                nodePool.labelValue,
                enabled
            );
            await nodeUtils.waitForNodeTopologyManagerPolicy(
                nodePool.nodeName,
                enabled ? TopologyManagerPolicy.SingleNumaNode : null
            );
        }

        async function getWorkloadEventMessages(workloadId: string): Promise<string[]> {
            const { body } = await workloadsUtilsV2.getWorkloadEvents(workloadId);
            return (body.events ?? []).map((event) => `${event.reason ?? ''}: ${event.message}`);
        }

        test.beforeAll(async () => {
            nodePool = await numaUtils.createNodePoolForGpuNode();
            nodeGpus = await kubeCli.getNumOfGpusOnNode(nodePool.nodeName);
            await numaUtils.setDefaultDepartmentGpuQuota(nodePool.nodePoolName, nodeGpus);
            project = await nodePoolUtils.createProjectWithSomeNodePoolResources([
                {
                    nodePoolName: nodePool.nodePoolName,
                    nodePoolId: nodePool.nodePoolId,
                    deservedGPUs: nodeGpus
                }
            ]);

            await enableNumaAndWait(true);

            const nrtZoneGpus = await numaUtils.getNrtZoneGpus(nodePool.nodeName);
            numaZones = nrtZoneGpus.length;
            zoneGpus = Math.max(0, ...nrtZoneGpus);
            logger.info(
                `node ${nodePool.nodeName}: ${nodeGpus} GPUs, ${numaZones} NUMA zones, up to ${zoneGpus} GPUs per zone`
            );
        });

        test.afterEach(async () => {
            for (const workload of createdWorkloads) {
                try {
                    await workloadsCrudUtils.deleteTraining(workload.workloadId);
                    await workloadsUtilsV2.waitForDeletedWorkload(workload.workloadId, project.id);
                } catch (error) {
                    logger.error(`Failed to delete workload ${workload.name}`, error);
                }
            }
            createdWorkloads = [];
            // a test may have turned NUMA off on the node pool
            await enableNumaAndWait(true);
        });

        test.afterAll(async () => {
            try {
                if (project) {
                    await operations.deleteProjectAndNamespace(project.id);
                }
            } catch (error) {
                logger.error(`Failed to delete project ${project?.name}`, error);
            }
            try {
                await numaUtils.deleteNumaTestNodePool(nodePool);
            } catch (error) {
                logger.error(`Failed to clean up node pool ${nodePool?.nodePoolName}`, error);
            }
        });

        test('Workload that fits in a single NUMA zone is scheduled', async () => {
            expect(zoneGpus, 'the node should report GPUs per NUMA zone').toBeGreaterThan(0);

            const workload = await createTraining(zoneGpus);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workload.workloadId, Phase.Running, 180);
        });

        test('Workload that cannot fit in a single NUMA zone stays pending with a NUMA reason', async () => {
            test.skip(
                numaZones < 2 || zoneGpus >= nodeGpus,
                'The node needs at least two NUMA zones with GPUs to split a workload across them'
            );

            let workload: Training;

            await test.step(`Submit a workload requesting ${zoneGpus + 1} of ${nodeGpus} GPUs`, async () => {
                workload = await createTraining(zoneGpus + 1);
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workload.workloadId, Phase.Pending, 180);
            });

            await test.step('Pending reason mentions NUMA alignment', async () => {
                let messages: string[] = [];
                await waitUtils.waitForCondition(
                    async () => {
                        messages = await getWorkloadEventMessages(workload.workloadId);
                        return messages.some((message) => NumaUtils.numaRejectionPattern.test(message));
                    },
                    `wait for a NUMA pending reason on workload ${workload.name}`,
                    true,
                    180,
                    5
                );
                logger.info(`events for ${workload.name}: ${JSON.stringify(messages)}`);
            });

            await test.step('Disabling NUMA-aware scheduling lets the workload run', async () => {
                await enableNumaAndWait(false);
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workload.workloadId, Phase.Running, 180);
            });
        });
    });
});
