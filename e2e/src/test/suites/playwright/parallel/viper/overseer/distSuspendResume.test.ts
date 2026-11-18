/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { kubeCli, runaiCliV2 } from '../../../../../../modules/cli';
import {
    cliLoginAsSystemAdmin,
    cliV2TrainingType,
    cliV2WorkloadType,
    SubmitFlagsMerged,
    SuspendFlagsMerged,
    CommonWorkloadFlags,
    ResumeFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { assert } from 'chai';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

import { test, expect } from '@playwright/test';

test.describe(
    `Distributed workload suspend resume tests ${tags.teams.viper}`,
    { tag: [tags.resource.distributedTrainingWorkload, tags.services.workloadOverseer] },
    function () {
        const jobNamePrefix = objectGenerator.generateShortNameByPrefix('w', 5);
        let project: Project;
        const runningCommand = '-- sleep 999999';
        const alreadySuspendedError = 'suspend request is not permitted for Stopped workloads';
        const alreadyRunningError = 'resume request is permitted only for stopped workloads';
        const distTypes = [
            cliV2TrainingType.mpi,
            cliV2TrainingType.xgBoost,
            cliV2TrainingType.tensorFlow,
            cliV2TrainingType.pyTorch
        ];

        function getJobNameByType(type: string): string {
            return `${jobNamePrefix}-${type}`;
        }

        async function deleteProject(projectObj: Project): Promise<void> {
            try {
                await operations.deleteProjectAndNamespace(projectObj.id);
            } catch (error) {
                logger.error(`Failed to delete project ${projectObj.name}`);
            }
        }

        async function generateRunningWorkloads(withGpu = false, projectName: string): Promise<void> {
            const submitFlag = {
                [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu
            } as Partial<typeof SubmitFlagsMerged>;
            submitFlag[SubmitFlagsMerged.command] = runningCommand;
            submitFlag[SubmitFlagsMerged.project] = projectName;
            if (withGpu) {
                submitFlag[SubmitFlagsMerged.gpuRequest] = 1;
            }
            await Promise.all(
                Object.values(distTypes).map(async (type) => {
                    const jobName = getJobNameByType(type);
                    runaiCliV2.submit(
                        submitFlag,
                        jobName,
                        cliV2WorkloadType.training,
                        type as cliV2TrainingType
                    );
                    await waitUtils.waitForBackendWorkloadPhase(jobName, project.id, Phase.Running);
                })
            );
        }

        test.beforeAll(function () {
            cliLoginAsSystemAdmin();
            supportedFrom(SupportedVersions.V2_23)(function () {
                distTypes.push(cliV2TrainingType.jax);
            });
        });
        test.describe('validate suspend and resume', function () {
            test.beforeAll(async function () {
                project = await projectResourcesUtils.createProjectAndNamespace('sus-res');
                await generateRunningWorkloads(false, project.name);
            });
            test.afterAll(async function () {
                await deleteProject(project);
            });

            test('suspend workloads', async function () {
                const suspendFlagsParams = {
                    [CommonWorkloadFlags.project]: project.name
                } as Partial<typeof SuspendFlagsMerged>;
                await Promise.all(
                    Object.values(distTypes).map(async (type) => {
                        runaiCliV2.suspend(suspendFlagsParams, getJobNameByType(type), type);
                        await waitUtils.waitForBackendWorkloadPhase(
                            getJobNameByType(type),
                            project.id,
                            Phase.Stopped,
                            90
                        );
                    })
                );
            });
            test('validate pod count=0 in suspended workloads', async function () {
                for (const type of Object.values(distTypes)) {
                    const workloadId = await workloadsUtilsV2.getWorkloadIdByName(
                        getJobNameByType(type),
                        project.id
                    );
                    const workloadPods = (await workloadsUtilsV2.getWorkloadPods(workloadId)).body.pods;
                    expect(workloadPods.length).toBe(0);
                }
            });
            test('suspend stopped workload', function () {
                const suspendFlagsParams = {
                    [CommonWorkloadFlags.project]: project.name
                } as Partial<typeof SuspendFlagsMerged>;
                for (const type of Object.values(distTypes)) {
                    assert.throw(
                        () => runaiCliV2.suspend(suspendFlagsParams, getJobNameByType(type), type),
                        alreadySuspendedError
                    );
                }
            });
            test('resume workloads', async function () {
                const resumeFlagsParams = {
                    [CommonWorkloadFlags.project]: project.name
                } as Partial<typeof ResumeFlagsMerged>;
                await Promise.all(
                    Object.values(distTypes).map(async (type) => {
                        runaiCliV2.resume(resumeFlagsParams, getJobNameByType(type), type);
                        await waitUtils.waitForBackendWorkloadPhase(
                            getJobNameByType(type),
                            project.id,
                            Phase.Running
                        );
                    })
                );
            });

            test('validate pod count > 0 in resumed workloads', async function () {
                for (const type of Object.values(distTypes)) {
                    const workloadId = await workloadsUtilsV2.getWorkloadIdByName(
                        getJobNameByType(type),
                        project.id
                    );
                    const workloadPods = (await workloadsUtilsV2.getWorkloadPods(workloadId)).body.pods;
                    expect(workloadPods.length).toBeGreaterThan(0);
                }
            });
            test('Resume running workload', function () {
                const resumeFlagsParams = {
                    [CommonWorkloadFlags.project]: project.name
                } as Partial<typeof ResumeFlagsMerged>;
                for (const type of Object.values(distTypes)) {
                    assert.throw(
                        () => runaiCliV2.resume(resumeFlagsParams, getJobNameByType(type), type),
                        alreadyRunningError
                    );
                }
            });
        });
        test.describe('Overseer suspend resume', function () {
            const defaultJobTimeLimitSec = 5;
            const timeoutTime = 90;
            let projectName: string;
            let namespace: string;
            const trainMaxGpuIdle = 'trainingJobMaxIdleDurationSecs';
            const trainJobTimeLimit = 'trainingJobTimeLimitSecs';

            async function setIdleGpuAnnotation(): Promise<void> {
                const idleGpuAnnotation = 'run.ai/simulated-gpu-utilization=0-1';
                await Promise.all(
                    Object.values(distTypes).map(async (type) => {
                        const podsNames = (
                            await workloadsUtilsV2.getWorkloadPodsByNameAndProjectName(
                                getJobNameByType(type),
                                project.name
                            )
                        ).body.pods.map((pod) => pod.name);
                        for (const podName of podsNames) {
                            await kubeCli.annotatePod(podName, project.namespace, idleGpuAnnotation);
                        }
                    })
                );
            }

            async function createProjectWithTimeLimit(
                field: string,
                jobTimeLimitSecs = defaultJobTimeLimitSec
            ): Promise<Project> {
                projectName = objectGenerator.generateNameByPrefix('project');
                const projectCreateRequest = {
                    name: projectName,
                    deservedGpus: 10,
                    maxAllowedGpus: 10
                };
                namespace = Utils.namespaceFromProjectName(projectName);
                projectCreateRequest[field] = jobTimeLimitSecs;
                project = await projectResourcesUtils.createProject(namespace, projectCreateRequest);
                project.namespace = namespace;
                return project;
            }

            test.afterEach(async function () {
                await deleteProject(project);
            });

            test('Should suspend workload - job time limit', async function () {
                project = await createProjectWithTimeLimit(trainJobTimeLimit);
                generateRunningWorkloads(false, project.name);
                await Promise.all(
                    Object.values(distTypes).map(async (type) => {
                        await waitUtils.waitForBackendWorkloadPhase(
                            getJobNameByType(type),
                            project.id,
                            Phase.Stopped,
                            timeoutTime
                        );
                    })
                );
            });
            test('Should suspend workload - idle gpu limit', async function () {
                test.setTimeout(240 * 1000);
                project = await createProjectWithTimeLimit(trainMaxGpuIdle);
                await generateRunningWorkloads(true, project.name);
                await setIdleGpuAnnotation();
                await Promise.all(
                    Object.values(distTypes).map(async (type) => {
                        await waitUtils.waitForBackendWorkloadPhase(
                            getJobNameByType(type),
                            project.id,
                            Phase.Stopped,
                            timeoutTime * 2
                        );
                    })
                );
            });
        });
    }
);
