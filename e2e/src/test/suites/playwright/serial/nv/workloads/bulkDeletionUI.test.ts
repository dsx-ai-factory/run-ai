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
import {
    shouldRunTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { test } from '@playwright/test';

test.describe(`Bulk Deletion UI ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.workloads}`, function () {
    supportedFrom(SupportedVersions.V2_19)(function () {
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
                projectName = objectGenerator.generateShortNameByPrefix(
                    'project-e2e-workloads-bulk-deletion'
                );
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
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
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch (error) {
                    logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                    throw error;
                }
            });

            test.beforeEach(async function () {
                await browserEngine.start();
                authPage = new AuthPage(browserEngine.currentPage);
                workloadPage = new WorkloadPage(browserEngine.currentPage);
            });

            test.afterEach('stop browser', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, 'workloads-bulk-deletion-' + videoName);
            });

            for (const [username, passwd] of [
                [flowRootState.users.l1_researcher.email, flowRootState.users.l1_researcher.password],
                [
                    flowRootState.users.system_administrator.email,
                    flowRootState.users.system_administrator.password
                ]
            ]) {
                test.describe('test workload bulk deletion', function () {
                    test.beforeEach(async function () {
                        await authPage.signIn(username, passwd);
                    });

                    test(`T4040413: ${username} should be able to select multiple workloads and delete them`, async function () {
                        videoName = 'select-multiple-workloads-and-delete-them';
                        const workloadNames = [0, 1, 2].map((i) =>
                            objectGenerator.generateShortNameByPrefix(`training-workload-${i}`)
                        );
                        await trainingUtils.createTrainingWorkloads(
                            workloadNames,
                            project,
                            { assets: { environment: environmentId, compute: computeResourceId } },
                            flowRootState.users.l1_researcher.accessToken
                        );
                        await workloadPage.navigateToWorkloads();

                        // wait for workloads to appear on UI
                        for (const name of workloadNames) {
                            await workloadPage.waitForWorkloadToAppear(name);
                        }

                        // The deleteWorkloads function only works when all workloads are visible in the UI,
                        // due to the last change made in the toggleWorkload function in Workload Page
                        await workloadPage.deleteWorkloads(workloadNames);

                        await Promise.all(
                            workloadNames.map(async (name) => {
                                await waitUtils.waitForElementToNotExist(
                                    workloadPage,
                                    `td:has-text("${name}")`,
                                    120
                                );
                            })
                        );
                    });
                });
            }
        });
    });
});
