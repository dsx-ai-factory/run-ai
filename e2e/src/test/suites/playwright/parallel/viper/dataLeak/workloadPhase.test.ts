/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    logger,
    objectGenerator,
    operations,
    PodStatus,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadsUtilsV2,
    workloadUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { WorkloadStatus } from '../../../../../../modules/helpers/models/baseWorkload';
import { EWorkloadStatusToPhase } from '../../../../../../modules/helpers/utils/WorkloadUtils';
import {
    cliLoginAsSystemAdmin,
    cliV2WorkloadType,
    listFlags,
    SubmitFlagsMerged,
    SuspendFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { test, expect } from '@playwright/test';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

const WorkloadPhaseParams = [
    {
        wlType: cliV2WorkloadType.training,
        image: config.quickStartImage.training
    },
    {
        wlType: cliV2WorkloadType.workspace,
        image: config.quickStartImage.interactive
    }
];

test.describe(
    `Workload phase, tags: ${tags.teams.viper}`,
    { tag: [tags.runOn.parallelCI, tags.resource.trainingWorkload] },
    function () {
        const defaultImagePath = config.quickStartImage.ubuntu;
        const sleepInfinity = '-- sleep infinity';
        const failPod = '-- /bin/sh -c "exit 1"';

        let project: Project;
        let jobName: string;
        test.beforeAll(function () {
            cliLoginAsSystemAdmin();
        });

        test.beforeEach(async function () {
            project = await projectResourcesUtils.createProjectAndNamespace('jobs-phases', undefined, 1, -1);
        });
        test.afterEach(async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch (error) {
                logger.error(`failed to delete project: ${project.name}`, error);
            }
        });

        async function validateWorkloadPhase(
            jobName: string,
            project: Project,
            expectedStatus: WorkloadStatus
        ): Promise<void> {
            await workloadUtils.waitForWorkloadRunnerPhase(
                jobName,
                project.namespace,
                EWorkloadStatusToPhase[expectedStatus]
            );
        }

        WorkloadPhaseParams.forEach(function (params) {
            test.describe(`${params.wlType} workload`, () => {
                test('completed', async function () {
                    jobName = await workloadUtils.createWorkload(
                        params.wlType,
                        'ubuntu',
                        undefined,
                        project,
                        WorkloadStatus.SUCCEEDED
                    );
                    await validateWorkloadPhase(jobName, project, WorkloadStatus.SUCCEEDED);
                });

                test('running', async function () {
                    jobName = await workloadUtils.createWorkload(
                        params.wlType,
                        params.image,
                        sleepInfinity,
                        project,
                        WorkloadStatus.RUNNING
                    );
                    await validateWorkloadPhase(jobName, project, WorkloadStatus.RUNNING);
                });

                test('stopped', async function ({}, testInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'WorkloadRunner phase semantics need investigation'
                    );
                    jobName = await workloadUtils.createWorkload(
                        params.wlType,
                        params.image,
                        sleepInfinity,
                        project,
                        WorkloadStatus.RUNNING
                    );
                    const suspendFlagsParams = {
                        [listFlags.project]: project.name
                    } as Partial<typeof SuspendFlagsMerged>;
                    runaiCliV2.suspend(suspendFlagsParams, jobName, params.wlType);
                    try {
                        await waitUtils.waitForPodsDeletion(jobName, project.namespace);
                    } catch (error) {
                        logger.error('pod of a suspended job was not deleted');
                        throw error;
                    }
                    await validateWorkloadPhase(jobName, project, WorkloadStatus.SUSPENDED);
                    const workloadId = await workloadsUtilsV2.getWorkloadIdByName(jobName, project.id);
                    const workloadPods = (await workloadsUtilsV2.getWorkloadPods(workloadId)).body.pods;
                    expect(workloadPods.length).toBe(0);
                });

                test('pending', async function ({}, testInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'WorkloadRunner phase semantics need investigation'
                    );
                    jobName = await workloadUtils.createWorkload(
                        params.wlType,
                        params.image,
                        sleepInfinity,
                        project,
                        WorkloadStatus.PENDING,
                        '100'
                    );
                    await validateWorkloadPhase(jobName, project, WorkloadStatus.PENDING);
                });

                test('failed', async function () {
                    const jobName = objectGenerator.generateNameByPrefix('cli-job');
                    const submitFlags = {
                        [SubmitFlagsMerged.image]: defaultImagePath,
                        [SubmitFlagsMerged.project]: project.name,
                        [SubmitFlagsMerged.gpuRequest]: '1',
                        [SubmitFlagsMerged.command]: failPod,
                        ...(params.wlType === cliV2WorkloadType.training
                            ? { [SubmitFlagsMerged.backoffLimit]: 0 }
                            : {})
                    } as Partial<typeof SubmitFlagsMerged>;
                    runaiCliV2.submit(submitFlags, jobName, params.wlType);
                    try {
                        await waitUtils.waitForPods(jobName, project.namespace, 1, 120, [PodStatus.FAILED]);
                    } catch (e) {
                        logger.error('pod is not in error');
                        throw e;
                    }
                    const expectedWlStatus = WorkloadStatus.FAILED;
                    await validateWorkloadPhase(jobName, project, expectedWlStatus);
                });
            });
        });
    }
);
