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
    waitUtils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { EnvironmentPage } from '../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { ComputeResourcePage } from '../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { workloadsCrudService } from '../../../../../../modules/api';

test.describe(`L1 researcher user - role testing, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.users}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let projectPage: ProjectPage;
        let envPage: EnvironmentPage;
        let workloadPage: WorkloadPage;
        let computeResourcePage: ComputeResourcePage;
        let dataSourcePage: DataSourcePage;
        let templatePage: TemplatePageV2;
        let credentialsPage: CredentialPage;
        let projectName: string;
        let trainingName: string;
        let trainingId: string;
        let workspaceName: string;
        let workspaceId: string;
        let namespace: string;
        let project: Project;
        let videoName: string;

        const l1Researcher = 'L1 reseacher';

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            trainingName = objectGenerator.generateShortNameByPrefix('training', 5);
            workspaceName = objectGenerator.generateShortNameByPrefix('workspace', 5);
            projectName = objectGenerator.generateShortNameByPrefix('proj-l1-researcher', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            projectPage = new ProjectPage(browserEngine.currentPage);
            envPage = new EnvironmentPage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);
            computeResourcePage = new ComputeResourcePage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);
            templatePage = new TemplatePageV2(browserEngine.currentPage);
            credentialsPage = new CredentialPage(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test('T3756062: test L1 researcher user roles', async () => {
            videoName = 'l1-researcher-user-roles';

            await test.step(`verify ${l1Researcher} can create training workload`, async () => {
                workloadsCrudService.trainingsApi.accessToken = flowRootState.users.l1_researcher.accessToken;

                const { body: workload } = await workloadsCrudUtils.createTraining(
                    trainingName,
                    false,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    { image: 'alpine:latest', command: 'sleep infinity' }
                );

                trainingId = workload.workloadId;

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getTraining(workload.workloadId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${workload.name} to be Running`,
                    true,
                    90
                );
            });

            await test.step(`verify ${l1Researcher} can delete training workload`, async () => {
                workloadsCrudService.trainingsApi.accessToken = flowRootState.users.l1_researcher.accessToken;
                const response = await workloadsCrudUtils.deleteTraining(trainingId);
                expect(response.response.statusCode).toEqual(202);
            });

            await test.step(`verify ${l1Researcher} can create workspace workload`, async () => {
                workloadsCrudService.workspacesApi.accessToken =
                    flowRootState.users.l1_researcher.accessToken;

                const { body: workload } = await workloadsCrudUtils.createWorkspace(
                    workspaceName,
                    false,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    { image: 'alpine:latest', command: 'sleep infinity' }
                );

                workspaceId = workload.workloadId;

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getWorkspace(workload.workloadId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${workload.name} to be Running`,
                    true,
                    90
                );
            });

            await test.step(`verify ${l1Researcher} can delete workspace workload`, async () => {
                workloadsCrudService.workspacesApi.accessToken =
                    flowRootState.users.l1_researcher.accessToken;
                const response = await workloadsCrudUtils.deleteWorkspace(workspaceId);
                expect(response.response.statusCode).toEqual(202);
            });

            await test.step(`verify ${l1Researcher} can not create inference workload`, async () => {
                await workloadPage.goToPage();
                await workloadPage.locator(workloadPage.addNewWorkloadBtn).click();
                await expect(workloadPage.locator(workloadPage.newInferenceBtn)).toHaveAttribute(
                    'class',
                    /disabled/
                );
            });

            await test.step(`verify ${l1Researcher} cannot see Resources menu`, async () => {
                await expect(
                    authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Resources}"]`)
                ).not.toBeVisible();
            });

            await test.step(`verify ${l1Researcher} cannot see Department menu`, async () => {
                await projectPage.goToPage();
                await expect(
                    projectPage.page.locator(`[aid="${MENU_SUB_ITEMS.Departments}"]`)
                ).not.toBeVisible();
            });

            await test.step(`verify ${l1Researcher} cannot see Access menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Access}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${l1Researcher} cannot see Admin menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Admin}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${l1Researcher} can see Analytics menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Analytics}"]`)).toBeVisible();
            });

            const viewOnlyTestData = [
                {
                    page: projectPage,
                    pageName: 'project page',
                    disabledElem: projectPage.addNewProjectBtn
                }
            ];

            for (const stepData of viewOnlyTestData) {
                await test.step(`verify data source admin user has view only permission on ${stepData.pageName}`, async () => {
                    await stepData.page.goToPage();
                    const viewOnlyBadgeShown = await stepData.page.viewOnlyBadgeVisible();
                    expect(viewOnlyBadgeShown).toBeTruthy();
                    const createBtnEnabled = await stepData.page.elementEnabled(stepData.disabledElem);
                    expect(createBtnEnabled).toBeFalsy();
                });
            }

            const canCreateTestData = [
                {
                    page: templatePage,
                    pageName: 'template page',
                    enabledElem: templatePage.newTemplateBtn
                },
                {
                    page: envPage,
                    pageName: 'environment page',
                    enabledElem: envPage.addNewEnvironmentBtn
                },
                {
                    page: computeResourcePage,
                    pageName: 'compute resource page',
                    enabledElem: computeResourcePage.addNewComputeBtn
                },
                {
                    page: dataSourcePage,
                    pageName: 'data source page',
                    enabledElem: dataSourcePage.addNewDataSourceBtn
                },
                {
                    page: credentialsPage,
                    pageName: 'credentials page',
                    enabledElem: credentialsPage.addNewCredentialBtn
                }
            ];

            for (const stepData of canCreateTestData) {
                await test.step(`verify ${l1Researcher} has create permission on ${stepData.pageName}`, async () => {
                    await stepData.page.goToPage();

                    if (stepData.enabledElem) {
                        const createBtnEnabled = await stepData.page.elementEnabled(stepData.enabledElem);
                        expect(createBtnEnabled).toBeTruthy();
                    }
                });
            }
        });
    });
});
