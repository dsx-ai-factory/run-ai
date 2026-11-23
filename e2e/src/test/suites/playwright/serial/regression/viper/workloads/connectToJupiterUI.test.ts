/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { HybridForm } from '../../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
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
    environmentUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { severity, Severity } from 'allure-js-commons';
import { expect, test } from '@playwright/test';

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
        let workspace: HybridForm;
        let workload: WorkloadPage;

        test.beforeEach('start browser', async function () {
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
            workspace = new HybridForm(browserEngine.currentPage);
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

        test('should connect successfully', async function () {
            test.setTimeout(180 * 1000);
            await severity(Severity.CRITICAL);
            videoName = 'workspace-connect';
            await workload.navigateToWorkloads();
            await workspace.fillWorkloadCreationFirstPage(
                flowRootState.cluster.name,
                projectName,
                workspaceName
            );
            const envName = (await environmentUtils.presetNames()).jupyter;
            await workspace.openEnvironmentsDrawer();
            await workspace.clickOnDrawerItemByLabel(envName);
            await workspace.fillComputeSection('one-gpu');
            await workspace.clickOnSubmit();
            await workload.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);
            const [newTab] = await Promise.all([
                browserEngine.currentPage.context().waitForEvent('page'),
                workload.connectWorkloadToolByName(workspaceName, EToolName.Jupyter)
            ]);
            await expect(newTab).toHaveURL(/\/(lab|tree)(?:[/?#]|$)/);
            await workload.openPythonConsole(newTab);
            await workload.deleteWorkload(workspaceName);
        });
    }
);
