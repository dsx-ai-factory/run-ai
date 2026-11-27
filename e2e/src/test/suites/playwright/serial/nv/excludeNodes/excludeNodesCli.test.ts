/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    nodeUtils,
    tags,
    waitUtils,
    flowRootState,
    UtilsError,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    SubmitFlagsMerged,
    cliV2WorkloadType,
    cliV2TrainingType
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { verifyNodeExclusion } from '../../../../../../modules/helpers/Validations';

test.describe(`Exclude Nodes CLI Tests, tags: ${tags.resource.workloads} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let workloadName: string;
        let workerNodeNames: string[];
        const alpineImage = 'alpine:latest';
        const sleepCmd = '-- sleep infinity';

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-cli-node-excl', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            workerNodeNames = await nodeUtils.getWorkerNodeNames();
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test.beforeEach(function () {
            workloadName = objectGenerator.generateShortNameByPrefix('excl-node-wl', 5);

            runaiCliV2.loginAsUser(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
            runaiCliV2.projectSet(project.name);
        });

        test.afterEach(function () {
            runaiCliV2.logout();
        });

        const dataSets = [
            {
                testCase: 'T5167072: L1 researcher can create a standard training with multi-nodes excluded',
                image: alpineImage,
                cmd: sleepCmd,
                workloadType: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.standard,
                excludedNodeCnt: 2,
                testEnable: true
            },
            {
                testCase: 'T5167073: L1 researcher can create a distributed training with 1 node excluded',
                image: alpineImage,
                cmd: sleepCmd,
                workloadType: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.mpi,
                excludedNodeCnt: 1,
                testEnable: true
            },
            {
                testCase: 'T5167074: L1 researcher can create a workspace workload with multi-nodes excluded',
                image: alpineImage,
                cmd: sleepCmd,
                workloadType: cliV2WorkloadType.workspace,
                trainingType: undefined,
                excludedNodeCnt: 2,
                testEnable: true
            },
            {
                testCase: 'T5167075: L1 researcher can create an inference workload with 1 node excluded',
                image: config.quickStartImage.inferenceHelloWorld,
                serverPort: 8083,
                workloadType: cliV2WorkloadType.inference,
                trainingType: undefined,
                excludedNodeCnt: 1,
                testEnable: true
            }
        ];

        dataSets.forEach((caseData) => {
            if (!caseData.testEnable) return;

            test(caseData.testCase, async function () {
                const excludedNodeNames = Utils.getRandomElements(workerNodeNames, caseData.excludedNodeCnt);

                if (caseData.workloadType === cliV2WorkloadType.inference) {
                    runaiCliV2.loginAsUser(
                        flowRootState.users.ml_engineer.email,
                        flowRootState.users.ml_engineer.password
                    );
                    runaiCliV2.projectSet(project.name);
                }

                await test.step(`create a ${caseData.workloadType} workload and exclude ${caseData.excludedNodeCnt} worker node(s)`, async () => {
                    let submitFlag: Partial<typeof SubmitFlagsMerged>;
                    if (caseData.cmd) {
                        submitFlag = {
                            [SubmitFlagsMerged.image]: caseData.image,
                            [SubmitFlagsMerged.excludeNode]: excludedNodeNames,
                            [SubmitFlagsMerged.command]: caseData.cmd
                        } as Partial<typeof SubmitFlagsMerged>;
                    } else {
                        submitFlag = {
                            [SubmitFlagsMerged.image]: caseData.image,
                            [SubmitFlagsMerged.servingPort]: caseData.serverPort,
                            [SubmitFlagsMerged.excludeNode]: excludedNodeNames
                        } as Partial<typeof SubmitFlagsMerged>;
                    }

                    runaiCliV2.submit(submitFlag, workloadName, caseData.workloadType, caseData.trainingType);

                    await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Running, 180);
                });

                await test.step('Verify workload is not running on excluded node(s)', async () => {
                    await verifyNodeExclusion(excludedNodeNames, namespace, workloadName);
                });

                await test.step('Delete workload', async () => {
                    logger.info(`case name: ${caseData.testCase}`);

                    if (caseData.trainingType === cliV2TrainingType.mpi) {
                        runaiCliV2.delete(undefined, workloadName, caseData.trainingType);
                    } else {
                        runaiCliV2.delete(undefined, workloadName, caseData.workloadType);
                    }

                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                });
            });
        });

        test('T5167076: Create standard training with all nodes excluded', async function () {
            await test.step('Create a stanadard training with all worker nodes excluded', function () {
                const submitFlag = {
                    [SubmitFlagsMerged.image]: alpineImage,
                    [SubmitFlagsMerged.excludeNode]: workerNodeNames,
                    [SubmitFlagsMerged.command]: sleepCmd
                } as Partial<typeof SubmitFlagsMerged>;

                runaiCliV2.submit(
                    submitFlag,
                    workloadName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.standard
                );
            });

            await test.step('Verify workload is stuck at pending', async function () {
                try {
                    await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Running, 60);
                } catch (error) {
                    if (error instanceof UtilsError) {
                        const workload = await workloadsUtilsV2.getWorkloadByNameAndProjectName(
                            workloadName,
                            projectName
                        );
                        expect(workload.phase).toBe('Pending');
                    } else {
                        throw error;
                    }
                }
            });

            await test.step('Delete workload', async () => {
                runaiCliV2.delete(undefined, workloadName, cliV2WorkloadType.training);
                await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
            });
        });
    });
});
