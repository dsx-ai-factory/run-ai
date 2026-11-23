/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    nodeUtils,
    tags,
    workloadsCrudUtils,
    waitUtils,
    distributedUtils,
    orgUnitUtils,
    flowRootState
} from '../../../../../../modules/helpers';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { kubeCli } from '../../../../../../modules/cli';
import { DistributedFramework } from '../../../../../../modules/api/apis/policyApi';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi/generated/model/project';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';


test.describe(`[Multi-Node] Distributed Training API Tests, tags: ${tags.tool.api} ${tags.resource.workloads} ${tags.teams.viper} ${tags.runOn.k8sFlavWorkloadsKit}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let project: Project;
        let namespace: string;
        let workloadName: string;
        let workerNodeNames: string[];
        let anyWorkloadSubmissionEnabled: boolean;
        const workloadIds: string[] = [];

        const alpineImage = 'alpine:latest';
        const sleepCmd = 'sleep infinity';

        async function cleanupWorkloads(workloadIds: string[]): Promise<void> {
            for (const workloadId of workloadIds) {
                try {
                    await workloadsCrudUtils.deleteDistributed(workloadId);
                } catch (error) {
                    logger.error(`Failed to delete workload ${workloadId}`, error);
                }
            }
        }

        test.beforeAll(async ({}, testInfo) => {
            anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
                AnyWorkloadSubmissionType.Distributed
            );
            workerNodeNames = await nodeUtils.getWorkerNodeNames();
            logger.info(
                `Found ${workerNodeNames.length} schedulable worker nodes: ${workerNodeNames.join(', ')}`
            );

            if (workerNodeNames.length < 2) {
                skipPlaywrightTest(test, testInfo, 'Requires at least 2 worker nodes');
                return;
            }

            const projectName = objectGenerator.generateShortNameByPrefix('proj-mn-dist', 5);
            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: [
                    {
                        gpu: { deserved: 8 }
                    }
                ]
            });
            namespace = Utils.namespaceFromProjectName(project.name);
        });

        test.afterAll(async () => {
            await cleanupWorkloads(workloadIds);

            if (!project) return;
            try {
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                logger.info(`Deleted project and namespace ${namespace}`);
            } catch (error) {
                logger.error(`Failed to delete project and namespace ${namespace}`, error);
            }
        });

        test.beforeEach(() => {
            workloadName = objectGenerator.generateShortNameByPrefix('mn-dist', 5);
        });

        [
            { name: 'MPI', framework: DistributedFramework.Mpi },
            { name: 'PyTorch', framework: DistributedFramework.PyTorch },
            { name: 'TensorFlow', framework: DistributedFramework.Tf }
        ].forEach(({ name, framework }) => {
            test(`Multi-node ${name} distributed training with node exclusion`, async () => {
                const numberOfNodes = workerNodeNames.length;
                const workerExcludeNodes = workerNodeNames.slice(0, numberOfNodes-1);
                const masterExcludeNodes = workerNodeNames.slice(1);
                let workloadId: string;

                await test.step(`Create ${name} distributed workload with node exclusion`, async () => {
                    const response = await distributedUtils.createDistributedTraining(
                        project.id,
                        workloadName,
                        {
                            distributedFramework: framework,
                            image: alpineImage,
                            command: sleepCmd,
                            numWorkers: 2,
                            compute: {
                                gpuDevicesRequest: 0,
                                cpuCoreRequest: 0.1,
                                cpuMemoryRequest: '100M'
                            }
                        },
                        {
                            distributedFramework: framework,
                            image: alpineImage,
                            command: sleepCmd,
                            compute: {
                                gpuDevicesRequest: 0,
                                cpuCoreRequest: 0.1,
                                cpuMemoryRequest: '100M'
                            }
                        },
                        workerExcludeNodes,
                        masterExcludeNodes
                    );

                    expect(response.response.statusCode).toEqual(202);
                    workloadId = response.body.workloadId;
                    workloadIds.push(workloadId);
                });

                await test.step('Wait for workload to be Running', async () => {
                    if (anyWorkloadSubmissionEnabled) {
                        await waitUtils.waitForBackendWorkloadPhase(
                            workloadName,
                            project.id,
                            Phase.Running,
                            120
                        );
                    } else {
                        await waitUtils.waitForCondition(
                            async () =>
                                (await workloadsCrudUtils.getDistributed(workloadId)).body.actualPhase ===
                                Phase.Running,
                            `Wait for ${workloadName} to be Running`,
                            true,
                            120
                        );
                    }
                });

                await test.step('Validate pods are scheduled on different nodes', async () => {
                    const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, workloadName);
                    expect(pods.length).toBeGreaterThanOrEqual(2);

                    const nodeSet = new Set(pods.map((p) => p.spec.nodeName));
                    logger.info(
                        `Workload ${workloadName} pods scheduled on nodes: ${Array.from(nodeSet).join(', ')}`
                    );
                    expect(nodeSet.size).toBeGreaterThan(1);
                });
            });
        });
    });
});
