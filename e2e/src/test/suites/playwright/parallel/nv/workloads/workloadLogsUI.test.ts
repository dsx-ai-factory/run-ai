/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    objectGenerator,
    Utils,
    waitUtils,
    logger,
    projectResourcesUtils,
    flowRootState,
    operations
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { WorkloadPage, EWorkloadStatus } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { WorkloadDetailsDrawer } from '../../../../../../modules/ui/runaiV2/pages/WorkloadDetailsDrawer';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { WorkspacePage } from '../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { tags } from '../../../../../../modules/helpers/config/tags';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Create single node workloads (Training & Workspace), tags: ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_15)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workloads: WorkloadPage;
        let detailsDrawer: WorkloadDetailsDrawer;
        let training: TrainingPage;
        let workspace: WorkspacePage;
        let projectPage: ProjectPage;
        let trainingName: string;
        let workspaceName: string;
        let projectName: string;
        let namespace: string;
        let project;
        let videoName: string;

        test.beforeEach(async () => {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
            detailsDrawer = new WorkloadDetailsDrawer(browserEngine.currentPage);
            training = new TrainingPage(browserEngine.currentPage);
            workspace = new WorkspacePage(browserEngine.currentPage);
            projectPage = new ProjectPage(browserEngine.currentPage);
            trainingName = objectGenerator.generateNameByPrefix('e2e-training');
            workspaceName = objectGenerator.generateNameByPrefix('e2e-workspace');
            projectName = objectGenerator.generateNameByPrefix('e2e-project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName || 'workload-logs-ui');
            if (project && project.id) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch {
                    logger.error('failed to delete project ' + projectName + ' and namespace ' + namespace);
                }
            }
        });

        test('T3738310: L1 researcher can get log of single node workload', async () => {
            videoName = 'log-of-single-node-workload';

            await training.createFlexibleTraining(projectName, trainingName);
            await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
            // Verify the workload log content
            await detailsDrawer.clickShowDetailsBtn(trainingName);
            await detailsDrawer.clickOnLogsTab();
            await detailsDrawer.waitForLogContentToAppear();

            // Verify the workload log info will be refreshed when the new log is generated
            const initialLogContent = await detailsDrawer.getLogContent();
            await waitUtils.waitForCondition(
                async () => {
                    try {
                        const logContent = await detailsDrawer.getLogContent();
                        const logIncludesInitialLog = logContent.includes(initialLogContent);
                        const logChanged = logContent.length > initialLogContent.length;
                        if (logIncludesInitialLog && logChanged) {
                            return true;
                        }
                        return false;
                    } catch {
                        return false;
                    }
                },
                'wait for log content changed',
                true,
                60,
                10
            );
        });
    });
});
