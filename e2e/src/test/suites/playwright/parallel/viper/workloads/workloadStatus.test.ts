/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliLoginAsSystemAdmin,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { kubeCli, runaiCliV2 } from '../../../../../../modules/cli';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadParams, initWorkloadParams, runningWorkloadParams } from './workloadStatus.consts';
import { test, TestInfo, expect } from '@playwright/test';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { shouldRunNativeWorkloadCrdTests } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

shouldRunNativeWorkloadCrdTests(() => {
test.describe(
    `Validate workload status ${tags.teams.viper}`,
    { tag: [tags.runOn.parallelCI, tags.resource.trainingWorkload] },
    function () {
        const jobNameSuffix = objectGenerator.generateShortNameByPrefix('w', 5);
        let project: Project;
        const workloads = {};
        const sleepTimeMs = 1000;

        function getJobNameByWorkload(workloadParam: WorkloadParams): string {
            return `${workloadParam.name}-${jobNameSuffix}`;
        }

        async function deleteProject(projectObj: Project): Promise<void> {
            try {
                await operations.deleteProjectAndNamespace(projectObj.id);
            } catch (error) {
                logger.error(`Failed to delete project ${projectObj.name}`);
            }
        }

        async function generateWorkloads(): Promise<void> {
            const submitAndWait = async (workload: WorkloadParams, phase: Phase): Promise<void> => {
                const jobName = getJobNameByWorkload(workload);
                workload.submitFlags[SubmitFlagsMerged.project] = project.name;
                runaiCliV2.submit(workload.submitFlags, jobName, workload.wlType, workload.trainingType);
                await waitUtils.waitForBackendWorkloadPhase(jobName, project.id, phase);
            };
            const getWorkloadData = async (workload: WorkloadParams): Promise<void> => {
                const jobName = getJobNameByWorkload(workload);
                workloads[jobName] = await workloadsUtilsV2.getWorkloadByNameAndProjectName(
                    jobName,
                    project.name
                );
            };
            await Promise.all([
                ...Object.values(initWorkloadParams).map((w) => submitAndWait(w, Phase.Initializing)),
                ...Object.values(runningWorkloadParams).map((w) => submitAndWait(w, Phase.Running))
            ]);
            await waitUtils.sleep(sleepTimeMs); // sleep to allow the yaml to be updated with the new phase data
            await Promise.all(
                Object.values([...initWorkloadParams, ...runningWorkloadParams]).map((w) =>
                    getWorkloadData(w)
                )
            );
        }

        test.beforeAll(async ({}, testInfo: TestInfo) => {
            const redirectedWorkloadTypes = [
                await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Workspace),
                await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Training),
                await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Distributed)
            ];
            if (redirectedWorkloadTypes.some(Boolean)) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'This suite depends on native phase, condition, and CRD status semantics'
                );
            }

            cliLoginAsSystemAdmin();
            project = await projectResourcesUtils.createProjectAndNamespace('wl-status');
            await generateWorkloads();
        });

        test.afterAll(async function () {
            if (project) {
                await deleteProject(project);
            }
        });

        test.describe('validate workloads init phase message', function () {
            for (const workload of initWorkloadParams) {
                test(`validate ${workload.name} init phase message`, function () {
                    const workloadData = workloads[getJobNameByWorkload(workload)];
                    expect(workloadData.phaseMessage).toBe(workload.expectedPhaseMsg);
                });
            }
        });
        test.describe('validate workload condition', function () {
            for (const workload of [...initWorkloadParams]) {
                test(`validate ${workload.name} condition`, function () {
                    const workloadData = workloads[getJobNameByWorkload(workload)];
                    const conditionMatch = workloadData.conditions.filter(
                        (condition: Record<string, string>) =>
                            condition.type === workload.condition.type &&
                            condition.status === workload.condition.status &&
                            condition.message === workload.condition.message
                    );
                    expect(conditionMatch.length).toBeGreaterThan(0);
                });
            }
        });
        test.describe('validate workload pods', function () {
            for (const workload of [...runningWorkloadParams, ...initWorkloadParams]) {
                test(`validate ${workload.name} pods`, function () {
                    const workloadData = workloads[getJobNameByWorkload(workload)];
                    expect(workloadData.runningPods).toBe(workload.runningPods);
                });
            }
            for (const workload of [...runningWorkloadParams]) {
                test(`validate ${workload.name} minRequiredReadyPods`, async function () {
                    const workloadYaml = await kubeCli.getWorkload(
                        workload.name,
                        project.namespace,
                        workload.crd
                    );
                    expect(workloadYaml['status']['minRequiredReadyPods']).toBe(
                        workload.minRequiredReadyPods
                    );
                });
            }
        });
    }
);
});
