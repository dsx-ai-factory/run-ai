/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    crds,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    tags,
    Utils,
    waitUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import { isOCP } from '../../../../../../../modules/helpers';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';
import { Phase } from '../../../../../../../modules/api/apis/assetsApi';
import {
    cliLoginAsSystemAdmin,
    cliV2workloadNoun,
    cliV2WorkloadType,
    DeleteFlagsMerged,
    preemptibilityOptions,
    ResumeFlagsMerged,
    SubmitFlagsMerged
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';

import { test, TestInfo } from '@playwright/test';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(
    `Workload overseer: ${tags.teams.viper}`,
    { tag: [tags.runOn.serialCI, tags.services.workloadOverseer] },
    function () {
        let projectName: string;
        let namespace: string;
        let workloadName: string;
        let projects: Project[] = [];
        let workspaceTimeLimitProject: Project;
        let workspaceTimeLimitZeroProject: Project;
        let trainingTimeLimitProject: Project;
        let trainingTimeLimitZeroProject: Project;
        let workspaceIdleGpuProject: Project;
        let workspacePreemptibleIdleGpuProject: Project;
        let trainingIdleGpuProject: Project;

        const WorkloadOverseerParams = [
            {
                wlType: cliV2WorkloadType.workspace,
                image: config.quickStartImage.interactive,
                jobTimeLimitProject: (): Project => workspaceTimeLimitProject,
                jobTimeLimitZeroProject: (): Project => workspaceTimeLimitZeroProject,
                jobTimeLimitSecsField: 'interactiveJobTimeLimitSecs',
                jobKind: crds.runaiInteractiveWorkloadCrd.kind,
                idleGpuProject: (): Project => workspaceIdleGpuProject,
                preemptible: false
            },
            {
                wlType: cliV2WorkloadType.training,
                image: config.quickStartImage.training,
                jobTimeLimitProject: (): Project => trainingTimeLimitProject,
                jobTimeLimitZeroProject: (): Project => trainingTimeLimitZeroProject,
                jobTimeLimitSecsField: 'trainingJobTimeLimitSecs',
                jobKind: crds.runaiTrainingWorkloadCrd.kind,
                idleGpuProject: (): Project => trainingIdleGpuProject,
                preemptible: false
            }
        ];
        const WorkloadOverseerParamsPreemptible = [
            {
                wlType: cliV2WorkloadType.workspace,
                image: config.quickStartImage.interactive,
                jobTimeLimitProject: (): Project => workspaceTimeLimitProject,
                jobTimeLimitZeroProject: (): Project => workspaceTimeLimitZeroProject,
                jobTimeLimitSecsField: 'interactiveJobTimeLimitSecs',
                jobKind: crds.runaiInteractiveWorkloadCrd.kind,
                idleGpuProject: (): Project => workspacePreemptibleIdleGpuProject,
                preemptible: true
            }
        ];
        const runningCommand = '-- sleep infinity';
        const workloadIdleGpuAnnotation = 'run.ai/simulated-gpu-utilization=0-1';
        const defaultJobTimeLimitSec = 5;
        const jobTimeLimitSecTimeout = 60;
        const waitForPhaseTime = 180;
        const deleteFlag = {
            [DeleteFlagsMerged.byPassDialog]: ''
        } as Partial<typeof DeleteFlagsMerged>;

        function getParmas(): any {
            supportedFrom(SupportedVersions.V2_24)(function () {
                return WorkloadOverseerParams.push(...WorkloadOverseerParamsPreemptible);
            });
            return WorkloadOverseerParams;
        }

        async function createProjectWithTimeLimit(
            field: string,
            jobTimeLimitSecs = defaultJobTimeLimitSec
        ): Promise<Project> {
            projectName = objectGenerator.generateNameByPrefix('project');
            namespace = Utils.namespaceFromProjectName(projectName);
            const projectCreateRequest = {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            };
            projectCreateRequest[field] = jobTimeLimitSecs;
            const project = await projectResourcesUtils.createProject(namespace, projectCreateRequest);
            projects.push(project);
            return project;
        }

        async function createWorkload(
            project: Project,
            workloadType: cliV2WorkloadType,
            image: string,
            preemptible = false,
            gpu = false,
            annotation = ''
        ): Promise<string> {
            workloadName = objectGenerator.generateNameByPrefix('cli-job');
            const submitFlag = {
                [SubmitFlagsMerged.image]: image,
                [SubmitFlagsMerged.project]: project.name,
                [SubmitFlagsMerged.command]: runningCommand
            } as Partial<typeof SubmitFlagsMerged>;
            if (preemptible) {
                submitFlag[SubmitFlagsMerged.preemptibility] = preemptibilityOptions.preemptible;
            }
            if (gpu) {
                submitFlag[SubmitFlagsMerged.gpuRequest] = 1;
            }
            if (annotation !== '') {
                submitFlag[SubmitFlagsMerged.annotation] = annotation;
            }
            runaiCliV2.submit(submitFlag, workloadName, workloadType);
            await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Running);
            return workloadName;
        }

        test.beforeAll(async function ({}, testInfo: TestInfo) {
            skipPlaywrightTest(
                test,
                testInfo,
                'Any Workload Suspend/Resume: block suspend/resume on Failed workloads (match V1 behavior)'
            );
            cliLoginAsSystemAdmin();
            [
                workspaceTimeLimitZeroProject,
                workspaceTimeLimitProject,
                trainingTimeLimitZeroProject,
                trainingTimeLimitProject,
                workspaceIdleGpuProject,
                workspacePreemptibleIdleGpuProject,
                trainingIdleGpuProject
            ] = await Promise.all([
                createProjectWithTimeLimit('interactiveJobTimeLimitSecs', 0),
                createProjectWithTimeLimit('interactiveJobTimeLimitSecs'),
                createProjectWithTimeLimit('trainingJobTimeLimitSecs', 0),
                createProjectWithTimeLimit('trainingJobTimeLimitSecs'),
                createProjectWithTimeLimit('interactiveJobMaxIdleDurationSecs'),
                createProjectWithTimeLimit('interactivePreemptibleJobMaxIdleDurationSecs'),
                createProjectWithTimeLimit('trainingJobMaxIdleDurationSecs')
            ]);
        });

        test.afterEach(function () {
            deleteFlag[DeleteFlagsMerged.project] = projectName;
            runaiCliV2.delete(deleteFlag, workloadName, cliV2workloadNoun, false);
        });
        test.afterAll(async function () {
            for (const project of projects) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    projects = projects.filter((p) => p !== project);
                } catch (error) {
                    logger.error(`Failed to delete project ${project.name} and namespace ${namespace}`);
                }
            }
        });

        getParmas().forEach(function (params) {
            test.describe(`${params.wlType} workload preemptible=${params.preemptible}`, () => {
                test(`${params.wlType} should suspend workload`, async function () {
                    projectName = params.jobTimeLimitProject().name;
                    await createWorkload(
                        params.jobTimeLimitProject(),
                        params.wlType,
                        params.image,
                        params.preemptible
                    );
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        params.jobTimeLimitProject().id,
                        Phase.Stopped,
                        jobTimeLimitSecTimeout
                    );
                });

                test(`${params.wlType} resume suspended workload via CLI`, async function () {
                    projectName = params.jobTimeLimitProject().name;
                    const resumeFlag = {
                        [ResumeFlagsMerged.project]: params.jobTimeLimitProject().name
                    } as Partial<typeof ResumeFlagsMerged>;
                    await createWorkload(
                        params.jobTimeLimitProject(),
                        params.wlType,
                        params.image,
                        params.preemptible
                    );
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        params.jobTimeLimitProject().id,
                        Phase.Stopped,
                        jobTimeLimitSecTimeout
                    );
                    runaiCliV2.resume(resumeFlag, workloadName, params.wlType);
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        params.jobTimeLimitProject().id,
                        Phase.Running,
                        jobTimeLimitSecTimeout
                    );
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        params.jobTimeLimitProject().id,
                        Phase.Stopped,
                        jobTimeLimitSecTimeout
                    );
                });

                test(`${params.wlType} change time limit`, async function () {
                    const project = params.jobTimeLimitProject();
                    projectName = project.name;
                    project[params.jobTimeLimitSecsField] = 300;
                    await projectResourcesUtils.updateProject(project);
                    // Ensure the 300s limit reaches the cluster Project CRD BEFORE we change it back
                    // to 5s, otherwise cluster-sync can process the two rapid updates out of order and
                    // leave the CRD stuck at 300, so the overseer never suspends the workload (flaky
                    // 180s timeout).
                    await projectUtils.validateProjectCrdSpec(
                        project.name,
                        60,
                        undefined,
                        undefined,
                        params.jobTimeLimitSecsField,
                        300
                    );
                    await createWorkload(project, params.wlType, params.image, params.preemptible);
                    project[params.jobTimeLimitSecsField] = defaultJobTimeLimitSec;
                    await projectResourcesUtils.updateProject(project);
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        params.jobTimeLimitProject().id,
                        Phase.Stopped,
                        waitForPhaseTime
                    );
                });
                test(`${params.wlType} workload with ${params.jobTimeLimitSecsField}=0 should not be suspended`, async function () {
                    const projectA = params.jobTimeLimitProject();
                    const workloadAName = await createWorkload(
                        projectA,
                        params.wlType,
                        params.image,
                        params.preemptible
                    );
                    const projectB = params.jobTimeLimitZeroProject();
                    const workloadBName = await createWorkload(
                        projectB,
                        params.wlType,
                        params.image,
                        params.preemptible
                    );
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadAName,
                        projectA.id,
                        Phase.Stopped,
                        waitForPhaseTime
                    );
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadBName,
                        projectB.id,
                        Phase.Running,
                        waitForPhaseTime
                    );

                    deleteFlag[DeleteFlagsMerged.project] = projectA.name;
                    runaiCliV2.delete(deleteFlag, workloadAName, cliV2workloadNoun, false);
                    deleteFlag[DeleteFlagsMerged.project] = projectB.name;
                    runaiCliV2.delete(deleteFlag, workloadBName, cliV2workloadNoun, false);
                });

                test(`${params.wlType} workload with idle gpu should be suspended`, async ({}, testInfo) => {
                    if (isOCP()) {
                        skipPlaywrightTest(
                            test,
                            testInfo,
                            '[Open Shift] idle gpu test is not working with fake gpu'
                        );
                        return;
                    }
                    workloadName = await createWorkload(
                        params.idleGpuProject(),
                        params.wlType,
                        params.image,
                        params.preemptible,
                        true,
                        workloadIdleGpuAnnotation
                    );
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        params.idleGpuProject().id,
                        Phase.Running
                    );
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        params.idleGpuProject().id,
                        Phase.Stopped,
                        waitForPhaseTime
                    );
                });
            });
        });
    }
);
