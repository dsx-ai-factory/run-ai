/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    computeResourceUtils,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    settings,
    tags,
    trainingUtils,
    Utils,
    waitUtils
} from '../../../../../../modules/helpers';
import { shouldRunTests } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { test } from '@playwright/test';

test.describe(`Bulk Deletion UI ${tags.teams.viper} ${tags.tool.ui}`, function () {
    shouldRunTests(settings.isSelfHosted)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let browserEngine: BrowserEngine;
        let workloadPage: WorkloadPage;
        let authPage: AuthPage;
        let videoName: string;

        let environmentId: string;
        let computeResourceId: string;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateShortNameByPrefix('project-e2e-workloads-bulk-deletion');
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

            [environmentId, computeResourceId] = await Promise.all([
                environmentUtils.getId('jupyter-lab'),
                computeResourceUtils.getId('cpu-only')
            ]);
        });

        test.afterAll('deleting project and namespace', async function () {
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
            authPage = new AuthPage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'workloads-bulk-deletion-' + videoName);
        });

        test('should be able to select multiple workloads and delete them', async function () {
            videoName = 'select-multiple-workloads-and-delete-them';
            const workloadNamePrefix = objectGenerator.generateShortNameByPrefix('t-bulk');
            const workloadNames = [0, 1, 2].map((i) => `${workloadNamePrefix}-${i}`);
            await trainingUtils.createTrainingWorkloads(workloadNames, project, {
                assets: { environment: environmentId, compute: computeResourceId }
            });
            await workloadPage.navigateToWorkloads();

            // The workloads table is shared by all parallel tests and only loads the first page.
            // Search by this test's unique prefix so all three workloads are rendered together.
            await Promise.all([
                workloadPage.waitForWorkloadsCountResponse({ nameInUrl: workloadNamePrefix }),
                workloadPage.searchByName(workloadNamePrefix)
            ]);

            // wait for workloads to appear on UI
            for (const name of workloadNames) {
                await workloadPage.waitForWorkloadToAppear(name);
            }

            await workloadPage.deleteWorkloads(workloadNames);
            await Promise.all(
                workloadNames.map((name) =>
                    waitUtils.waitForElementToNotExist(
                        workloadPage,
                        `td${workloadPage.workloadNameCol}:has-text("${name}")`,
                        120
                    )
                )
            );
        });
    });
});
