/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { TrainingCreationRequestV2, TrainingV2 } from '../../../../../../modules/api/apis/assetsApi';
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

            function getTrainingWorkloadPayload(name: string): TrainingCreationRequestV2 {
                return {
                    meta: {
                        projectId: project.id.toString(),
                        name: name,
                        clusterId: project.clusterUuid
                    },
                    assets: {
                        environment: environmentId,
                        compute: computeResourceId
                    }
                };
            }

            async function createTrainingWorkloads(
                amount: number,
                accessToken?: string
            ): Promise<TrainingV2[]> {
                const trainingWorkloads = [];
                for (let i = 0; i < amount; i++) {
                    const name = objectGenerator.generateShortNameByPrefix(`training-workload-${i}`);
                    const workloadPayload = getTrainingWorkloadPayload(name);
                    await trainingUtils.createTraining(workloadPayload, accessToken);
                    trainingWorkloads.push(workloadPayload);
                }
                return trainingWorkloads;
            }

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
                        const workloads = await createTrainingWorkloads(
                            3,
                            flowRootState.users.l1_researcher.accessToken
                        );
                        await workloadPage.navigateToWorkloads();

                        // wait for workloads to appear on UI
                        for (const workload of workloads) {
                            await workloadPage.waitForWorkloadToAppear(workload.meta.name);
                        }

                        // The deleteWorkloads function only works when all workloads are visible in the UI,
                        // due to the last change made in the toggleWorkload function in Workload Page
                        await workloadPage.deleteWorkloads(workloads.map((workload) => workload.meta.name));

                        await Promise.all(
                            workloads.map(async (workload) => {
                                await waitUtils.waitForElementToNotExist(
                                    workloadPage,
                                    `td:has-text("${workload.meta.name}")`,
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
