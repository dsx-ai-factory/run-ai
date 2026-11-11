/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, Page } from '@playwright/test';
import {
    objectGenerator,
    Utils,
    waitUtils,
    logger,
    projectResourcesUtils,
    flowRootState,
    operations,
    apiUtils,
    unleashClient
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { WorkspacePage } from '../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { TensorBoardPage } from '../../../../../../modules/ui/runaiV2/pages/ToolsTensorBoardPage';

import { JupyterPage } from '../../../../../../modules/ui/runaiV2/pages/ToolsJupyterPage';
import { tags } from '../../../../../../modules/helpers/config/tags';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { EToolName } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { EFeatureToggleKeys } from '../../../../../../modules/helpers/utils/Unleash/featureToggle';

test.describe(`Create interactive jobs, tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.loong} ${tags.runOn.dgxc}`, () => {
    supportedFrom(SupportedVersions.V2_15)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workloads: WorkloadPage;
        let workspace: WorkspacePage;
        let projectPage: ProjectPage;
        let toolsBasePages;
        let workspaceName: string;
        let projectName: string;
        let namespace: string;
        let project;
        let videoName: string;
        let baseUrl: string;
        let initialGpuCount: number;
        let initialCpuCount: number;
        let host: string;

        test.beforeAll(async () => {
            baseUrl = process.env.CONTROL_PLANE_URL?.replace(/\/$/, '');
            host = baseUrl.replace(/^https?:\/\//, '');
            await unleashClient.disableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            await apiUtils.disableHybridSubmission();
        });

        test.afterAll(async () => {
            await unleashClient.enableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
            await apiUtils.enableHybridSubmission();
            await apiUtils.setFlexTemplateFeature(true);
        });

        test.beforeEach(async () => {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
            workspace = new WorkspacePage(browserEngine.currentPage);
            projectPage = new ProjectPage(browserEngine.currentPage);
            toolsBasePages = [];
            workspaceName = objectGenerator.generateShortNameByPrefix('e2e-wl', 5);
            projectName = objectGenerator.generateShortNameByPrefix('e2e-proj', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName || 'interactive-jobs-test');
            for (const toolsBasePage of toolsBasePages) {
                await toolsBasePage.savePageVideo(this);
            }
            if (project && project.id) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch {
                    logger.error('failed to delete project ' + projectName + ' and namespace ' + namespace);
                }
            }
        });

        test('T3794050: L1 researcher can create interactive jobs with tensorboard', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'interactive-jobs-with-tensorboard';
            await test.step('Submit workspace workload', async () => {
                await workspace.submitWorkspaceWorkloadUI(projectName, workspaceName, {
                    environment: 'tensorboard-tensorflow',
                    computeResource: 'cpu-only'
                });
            });

            await test.step('Verify workspace is go to running state', async () => {
                await workspace.waitForWorkspaceStatus(workspaceName, 'Running');
            });

            await test.step('Wait for workspace to be fully ready', async () => {
                await new Promise((resolve) => setTimeout(resolve, 15000));
            });

            let tbTab: Page | undefined;
            await waitUtils.waitForCondition(
                async () => {
                    try {
                        if (tbTab && !tbTab.isClosed()) await tbTab.close();
                        await workloads.reload();
                        const newPagePromise = browserEngine.currentPage
                            .context()
                            .waitForEvent('page', { timeout: 30000 });
                        await workloads.connectWorkloadToolByName(workspaceName, EToolName.TensorBoard);
                        tbTab = await newPagePromise;
                        const tensorBoardPage = new TensorBoardPage(tbTab, 'T3794050-tensorboard-page');
                        const TensorBoardUrl = `https://${projectName}-${workspaceName}.${host}/`;
                        await tensorBoardPage.checkForTensorBoardPage(TensorBoardUrl);
                        toolsBasePages.push(tensorBoardPage);
                        return true;
                    } catch (error) {
                        logger.error(`Failed to connect to workspace and verify TensorBoard: ${error}`);
                        return false;
                    }
                },
                'Connect to workspace and verify TensorBoard Page',
                true,
                120,
                10
            );

            await workloads.deleteWorkload(workspaceName);
        });

        test('T3738304: L1 researcher can create interactive jobs with Jupyter Notebook and Tensorboard', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'interactive-jobs-with-jupyter-and-tensorboard';
            await test.step('Submit workspace workload', async () => {
                await workspace.submitWorkspaceWorkloadUI(projectName, workspaceName, {
                    environment: 'jupyter-scipy',
                    computeResource: 'cpu-only'
                });
            });

            await test.step('Verify workspace is go to running state', async () => {
                await workspace.waitForWorkspaceStatus(workspaceName, 'Running');
            });

            await test.step('Wait for workspace to be fully ready', async () => {
                await new Promise((resolve) => setTimeout(resolve, 15000));
            });

            let jup1Tab: Page | undefined;
            await waitUtils.waitForCondition(
                async () => {
                    try {
                        if (jup1Tab && !jup1Tab.isClosed()) await jup1Tab.close();
                        await workloads.reload();
                        const newPagePromise = browserEngine.currentPage
                            .context()
                            .waitForEvent('page', { timeout: 30000 });
                        await workloads.connectWorkloadToolByName(workspaceName, EToolName.Jupyter);
                        jup1Tab = await newPagePromise;
                        const jupyterPage = new JupyterPage(jup1Tab, 'T3793926-jupyter-page');
                        const jupyterUrl = `https://${projectName}-${workspaceName}.${host}/lab`;
                        await jupyterPage.checkForJupyterPage(jupyterUrl);
                        toolsBasePages.push(jupyterPage);
                        return true;
                    } catch (error) {
                        logger.error(`Failed to connect to workspace and verify Jupyter: ${error}`);
                        return false;
                    }
                },
                'Connect to workspace and verify Jupyter Page',
                true,
                120,
                10
            );

            await workloads.deleteWorkload(workspaceName);
        });

        test('T3793926: L1 researcher can create interactive jobs with jupyter-lab', async () => {
            test.setTimeout(300 * 1000);
            videoName = 'interactive-jobs-with-jupyter-lab';
            initialGpuCount = await projectPage.getAllocatedGpuCount(projectName);
            initialCpuCount = await projectPage.getAllocatedCpuCount(projectName);
            await test.step('Submit workspace workload', async () => {
                await workspace.submitWorkspaceWorkloadUI(projectName, workspaceName, {
                    environment: 'jupyter-scipy',
                    computeResource: 'cpu-only',
                    cmd: 'cat start-notebook.sh && start-notebook.sh',
                    backoffLimit: 1
                });
            });

            await test.step('Checked the job is created and goes into fail state', async () => {
                await workspace.waitForWorkspaceStatus(workspaceName, 'Failed');
            });

            await test.step('Delete the workspace', async () => {
                await workloads.deleteWorkload(workspaceName);
                await waitUtils.waitForElementToNotExist(
                    workloads,
                    'td:has-text(' + `"${workspaceName}"` + ')'
                );
            });

            await test.step('Create workspace again, but change the runtime command', async () => {
                await workspace.submitWorkspaceWorkloadUI(projectName, workspaceName, {
                    environment: 'jupyter-scipy',
                    computeResource: 'cpu-only',
                    cmd: '/bin/bash -c "cat /usr/local/bin/start-notebook.sh; start-notebook.sh --NotebookApp.token=\'\'"'
                });
            });

            await test.step('Verify workspace is go to running state', async () => {
                await workspace.waitForWorkspaceStatus(workspaceName, 'Running');
            });

            await test.step('Wait for workspace to be fully ready', async () => {
                await new Promise((resolve) => setTimeout(resolve, 15000));
            });

            let jup2Tab: Page | undefined;
            await waitUtils.waitForCondition(
                async () => {
                    try {
                        if (jup2Tab && !jup2Tab.isClosed()) await jup2Tab.close();
                        await workloads.reload();
                        const newPagePromise = browserEngine.currentPage
                            .context()
                            .waitForEvent('page', { timeout: 30000 });
                        await workloads.connectWorkloadToolByName(workspaceName, EToolName.Jupyter);
                        jup2Tab = await newPagePromise;
                        const jupyterPage = new JupyterPage(jup2Tab, 'T3793926-jupyter-page');
                        const jupyterUrl = `https://${projectName}-${workspaceName}.${host}/lab`;
                        await jupyterPage.checkForJupyterPage(jupyterUrl);
                        toolsBasePages.push(jupyterPage);
                        return true;
                    } catch (error) {
                        logger.error(`Failed to connect to workspace and verify Jupyter: ${error}`);
                        return false;
                    }
                },
                'Connect to workspace and verify Jupyter Page',
                true,
                120,
                10
            );

            await test.step('Stop the workspace, should be stopped successfully', async () => {
                await workspace.stopWorkloadByName(workspaceName);
                await workspace.waitForWorkspaceStatus(workspaceName, 'Stopped');
            });

            await test.step('Delete the workspace', async () => {
                await workloads.deleteWorkload(workspaceName);
                await waitUtils.waitForElementToNotExist(
                    workloads,
                    'td:has-text(' + `"${workspaceName}"` + ')'
                );
            });

            await test.step('Check the active quota for the project used by the first workspace is released', async () => {
                const finalGpuCount = await projectPage.getAllocatedGpuCount(projectName);
                const finalCpuCount = await projectPage.getAllocatedCpuCount(projectName);
                expect(finalGpuCount).toBe(initialGpuCount);
                expect(finalCpuCount).toBe(initialCpuCount);
            });
        });
    });
});
