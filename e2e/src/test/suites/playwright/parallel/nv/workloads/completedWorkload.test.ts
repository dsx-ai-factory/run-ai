/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { HybridForm } from '../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { kubeCli } from '../../../../../../modules/cli';

test.describe(`Workload completed status tests, tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let trainingPage: TrainingPage;
        let hybridForm: HybridForm;
        let projectName: string;
        let namespace: string;
        let workloadName: string;
        let project: Project;
        let videoName: string;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            projectName = objectGenerator.generateShortNameByPrefix('proj-com', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`, error);
                throw error;
            }
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            trainingPage = new TrainingPage(browserEngine.currentPage);
            hybridForm = new HybridForm(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );

            workloadName = objectGenerator.generateShortNameByPrefix('workload', 5);
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test('T3952691: test automatic cleanup of pods when training reaches the completed status', async () => {
            videoName = 'workload-completed-status';

            await test.step('create a pytorch training workload', async () => {
                await trainingPage.goToPage();
                await trainingPage.fillFirstNewTrainingPage(projectName, workloadName);
                await trainingPage.page.waitForLoadState('load');
                await hybridForm.fillImageInput('alpine:latest');
                await hybridForm.clickOnSubSectionByLabel('Runtime settings');
                await hybridForm.clickOnCommandAndArguments();
                await hybridForm.fillCmdInput('sleep 5');
                await hybridForm.selectExistingComputeResource('cpu-only');
                await trainingPage.clickOnSubmitBtn();
            });

            await test.step('wait for the training to be running', async () => {
                await trainingPage.waitForWorkloadUIStatus(workloadName, EWorkloadStatus.Running);
            });

            await test.step('wait for the training to be completed', async () => {
                await trainingPage.waitForWorkloadUIStatus(workloadName, EWorkloadStatus.Completed);
            });

            await test.step('verify workload with succeeded status and pod is stopped', async () => {
                await waitUtils.waitForCondition(
                    async () => {
                        const pods = await kubeCli.listPodsByNamespaceAndPrefix(namespace, workloadName);
                        return (
                            pods[0].status.phase === 'Succeeded' &&
                            pods[0].status.containerStatuses?.every(
                                (container) =>
                                    container.state?.running === undefined &&
                                    container.state?.terminated !== undefined
                            )
                        );
                    },
                    'Wait for pod to be stopped.',
                    true,
                    60
                );
            });

            await test.step('delete the workload', async () => {
                await trainingPage.deleteWorkload(workloadName);
            });
        });
    });
});
