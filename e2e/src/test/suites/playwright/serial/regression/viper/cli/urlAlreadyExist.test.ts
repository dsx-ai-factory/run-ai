/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    TrainingWorkload,
    Utils,
    waitUtils,
    workloadUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { runaiCliV2 } from '../../../../../../../modules/cli';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    cliLoginAsSystemAdmin,
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

import { test } from '@playwright/test';

test.describe(`Test URL Already Exist, tags: ${tags.tool.ui} ${tags.resource.jobs} ${tags.teams.viper}`, function () {
    shouldRunTests(!isOCP())(function () {
        let projectName: string;
        let project: Project;
        let namespace: string;
        let redirectedSubmissionEnabled = false;

        test.beforeAll('creating project', async function () {
            projectName = objectGenerator.generateNameByPrefix('url-exist');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 2,
                maxAllowedGpus: 2
            });
            redirectedSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
                AnyWorkloadSubmissionType.Training
            );
            logger.info(`created project ${projectName}`);
        });

        test.afterAll(async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch (error) {
                logger.error(`Failed to delete project ${projectName}`);
            }
        });

        test.describe('Workload API', function () {
            const runningWorkloadName = objectGenerator.generateNameByPrefix('job');
            const failWorkloadName = runningWorkloadName + '-failed';
            const newRunningWorkloadName = runningWorkloadName + '-new';

            function getSubmitFlags(): Partial<typeof SubmitFlagsMerged> {
                return {
                    [SubmitFlagsMerged.image]: config.quickStartImage.sleepInfinity,
                    [SubmitFlagsMerged.externalUrl]: `container=9443,url=${config.appUrl}/${projectName}`
                } as Partial<typeof SubmitFlagsMerged>;
            }

            test.beforeAll(function () {
                cliLoginAsSystemAdmin();
                runaiCliV2.projectSet(projectName);
            });

            test('Reject a duplicate URL and reuse it after deleting the owner', async function () {
                await test.step('Submit two workloads with the same URL', async () => {
                    runaiCliV2.submit(getSubmitFlags(), runningWorkloadName, cliV2WorkloadType.training);
                    await waitUtils.waitForBackendWorkloadPhase(
                        runningWorkloadName,
                        project.id,
                        Phase.Running
                    );
                    runaiCliV2.submit(getSubmitFlags(), failWorkloadName, cliV2WorkloadType.training);
                    await waitUtils.waitForBackendWorkloadPhase(
                        failWorkloadName,
                        project.id,
                        redirectedSubmissionEnabled ? Phase.Degraded : Phase.Failed
                    );
                });

                await test.step('Release and reuse the URL', async () => {
                    runaiCliV2.delete({}, runningWorkloadName, cliV2WorkloadType.training);

                    if (redirectedSubmissionEnabled) {
                        await waitUtils.waitForWorkloadToBeDeleted(
                            runningWorkloadName,
                            70,
                            flowRootState.users.system_administrator.accessToken
                        );
                        await waitUtils.waitForBackendWorkloadPhase(
                            failWorkloadName,
                            project.id,
                            Phase.Running
                        );
                        return;
                    }

                    const runningWorkload = new TrainingWorkload(
                        config.quickStartImage.sleepInfinity,
                        namespace,
                        runningWorkloadName
                    );
                    await workloadUtils.isWorkloadDeleted(runningWorkload, 30);
                    runaiCliV2.submit(getSubmitFlags(), newRunningWorkloadName, cliV2WorkloadType.training);
                    await waitUtils.waitForBackendWorkloadPhase(
                        newRunningWorkloadName,
                        project.id,
                        Phase.Running
                    );
                    await waitUtils.waitForBackendWorkloadPhase(failWorkloadName, project.id, Phase.Failed);
                });
            });
        });
    });
});
