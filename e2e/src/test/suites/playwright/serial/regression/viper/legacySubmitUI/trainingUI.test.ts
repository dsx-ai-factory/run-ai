/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    Utils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    projectUtils,
    settings,
    unleashClient,
    apiUtils
} from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    DISTRIBUTED_TOOLS_LABELS,
    TrainingPage
} from '../../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { WorkloadPage, EWorkloadStatus } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { EnvironmentPage } from '../../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { test } from '@playwright/test';
import { EFeatureToggleKeys } from '../../../../../../../modules/helpers/utils/Unleash/featureToggle';

test.describe(
    `Training UI - sanity, tags: ${tags.tool.ui} ${tags.resource.jobs} ${tags.teams.viper} `,
    { tag: [tags.services.frontend] },
    () => {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let training: TrainingPage;
        let environment: EnvironmentPage;
        let workloads: WorkloadPage;
        let videoName: string;
        let projectName: string;
        let namespace: string;
        let project: Project;
        let trainingName: string;
        let environmentName: string;

        async function fillFirstNewTrainingPage(
            projectName: string,
            isMultiNode = false,
            toolName?: DISTRIBUTED_TOOLS_LABELS,
            configureMaster?: boolean
        ): Promise<void> {
            await workloads.navigateToWorkloads();
            await workloads.openTrainingForm();
            await workloads.selectClusterFromDropDown(flowRootState.cluster.name);
            await workloads.workloadFormSelectProject(projectName);
            if (isMultiNode) {
                await waitUtils.sleep(1000);
                await training.selectedMultiNodeOption();
                await training.selectedDistributedTool(toolName);
                if (!configureMaster) await training.selectedWorkersOnlyOption();
            }
            await workloads.workloadFormSelectTemplate();
            await training.fillTrainingName(trainingName);
            await training.waitForTrainingNameCount(trainingName);
            await training.clickOnContinueBtn();
        }

        test.beforeAll(async function () {
            if (!settings.isSelfHosted) {
                await unleashClient.disableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            }
            await apiUtils.disableHybridSubmission();
            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateShortNameByPrefix('project-e2e-for-training');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 4,
                maxAllowedGpus: 4
            });

            const roleBindingName = projectUtils.getSystemAdministratorRoleBindingName();
            await projectUtils.waitForRoleBindingForUserInProject(
                namespace,
                flowRootState.users.system_administrator.email,
                roleBindingName,
                true
            );
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} an namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
                throw error;
            }
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            training = new TrainingPage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
            environment = new EnvironmentPage(browserEngine.currentPage);
            trainingName = objectGenerator.generateShortNameByPrefix('training');
            environmentName = objectGenerator.generateShortNameByPrefix('environment');

            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'training-' + videoName);
        });

        test('copy and edit training', async () => {
            videoName = 'copy-and-edit-training';
            if (settings.isSelfHosted) test.setTimeout(180 * 1000);
            const copiedTrainingName = objectGenerator.generateShortNameByPrefix('copy-training');
            await fillFirstNewTrainingPage(projectName);
            await training.clickOnNewEnvironmentBtn();
            await environment.createNewEnvironment(environmentName, projectName, {
                command: 'sleep infinity'
            });
            await training.selectComputeResource('cpu-only');
            await training.clickOnSubmitBtn();
            await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
            await workloads.toggleWorkload(trainingName);
            await workloads.clickOnCopyAndEditWorkload();
            await training.fillTrainingName(copiedTrainingName);
            await training.clickOnSubmitBtn();
            await workloads.waitForWorkloadUIStatus(copiedTrainingName, EWorkloadStatus.Running);
            await workloads.deleteWorkload(trainingName);
            await workloads.deleteWorkload(copiedTrainingName);
            await environment.goToPage();
            await environment.removeEnvironment(environmentName);
        });
    }
);
