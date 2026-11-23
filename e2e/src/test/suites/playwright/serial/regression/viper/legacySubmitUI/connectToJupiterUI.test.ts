/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { WorkspacePage } from '../../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import {
    WorkloadPage,
    EWorkloadStatus,
    EToolName
} from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    flowRootState,
    objectGenerator,
    logger,
    tags,
    projectResourcesUtils,
    Utils,
    operations,
    waitUtils,
    unleashClient,
    apiUtils,
    environmentUtils,
    settings
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { expect } from 'chai';
import { severity, Severity } from 'allure-js-commons';
import { Page } from 'playwright';
import { test } from '@playwright/test';
import { EFeatureToggleKeys } from '../../../../../../../modules/helpers/utils/Unleash/featureToggle';

test.describe(
    `Workspaces UI - Sanity checks, tags: ${tags.tool.ui} ${tags.resource.workspace} ${tags.teams.viper} ${tags.runOn.parallelCI}`,
    { tag: [tags.services.frontend] },
    () => {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let videoName: string;
        let workspaceName: string;
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workspace: WorkspacePage;
        let workload: WorkloadPage;

        test.beforeEach('start browser', async function () {
            if (!settings.isSelfHosted) {
                await unleashClient.disableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            }
            await apiUtils.disableHybridSubmission();
            projectName = objectGenerator.generateNameByPrefix('p-workspace');
            workspaceName = objectGenerator.generateNameByPrefix('workspace');
            namespace = Utils.namespaceFromProjectName(projectName);
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });

            auth = new AuthPage(browserEngine.currentPage);
            workspace = new WorkspacePage(browserEngine.currentPage);
            workload = new WorkloadPage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            try {
                await browserEngine.stopAndArchive(testInfo, 'workspaces-' + videoName);
            } catch (e) {
                logger.error(e);
            }
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} an namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
                throw error;
            }
        });

        function validateUrl(newTab: Page): void {
            const isValidUrl = ['/lab', '/tree'].some((validUrl) => newTab.url().includes(validUrl));
            expect(isValidUrl).to.be.true;
            logger.debug('The url is valid');
        }

        test('should connect successfully', async function () {
            test.setTimeout(180 * 1000);
            await severity(Severity.CRITICAL);
            videoName = 'workspace-connect';
            await workload.navigateToWorkloads();
            await workload.openWorkspaceForm();
            // fill the workspace form details
            await workspace.selectClusterFromDropDown(flowRootState.cluster.name);
            await workspace.workloadFormSelectProject(projectName);
            await workspace.fillWorkspaceName(workspaceName);
            await workspace.waitForWorkspaceNameCount(workspaceName);
            await workspace.clickOnContinue();
            const envName = (await environmentUtils.presetNames()).jupyter;
            await workspace.selectEnvironment(envName);
            await workspace.selectComputeResource('one-gpu');
            await workspace.clickOnCreateWorkspaceBtn();
            logger.debug('workspace created successfully');
            await workload.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running); //TODO: fix it, not working
            await waitUtils.sleep(10000); // TODO: remove after the prev line is fixed
            logger.debug('waiting for to jupyter loading');
            await workload.connectWorkloadToolByName(workspaceName, EToolName.Jupyter);
            logger.debug('wait for new tab loading');
            await browserEngine.page.waitForTimeout(5000);
            const newTab = browserEngine.context.pages()[1];
            validateUrl(newTab);
            await workload.openPythonConsole(newTab);
            await workload.deleteWorkload(workspaceName);
        });
    }
);
