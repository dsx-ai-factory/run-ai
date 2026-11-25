/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils
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
import { QTable } from '../../../../../../modules/ui/runaiV2/components/QTable';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { ComputeResourcePage } from '../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { DepartmentPage } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPage';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';

test.describe(`Environment administrator user - role testing, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.users}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let projectPage: ProjectPage;
        let departmentPage: DepartmentPage;
        let envPage: EnvironmentPage;
        let workloadPage: WorkloadPage;
        let computeResourcePage: ComputeResourcePage;
        let dataSourcePage: DataSourcePage;
        let templatePage: TemplatePageV2;
        let qTable: QTable;
        let projectName: string;
        let envName: string;
        let namespace: string;
        let project: Project;
        let videoName: string;

        const sleep10s = 'sleep 10';
        const envAdmin = 'env admin user';

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            projectName = objectGenerator.generateShortNameByPrefix('proj-env-admin', 5);
            envName = objectGenerator.generateShortNameByPrefix('env', 5);
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
            departmentPage = new DepartmentPage(browserEngine.currentPage);
            envPage = new EnvironmentPage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);
            computeResourcePage = new ComputeResourcePage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);
            templatePage = new TemplatePageV2(browserEngine.currentPage);
            qTable = new QTable(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.environment_administrator.email,
                flowRootState.users.environment_administrator.password
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

        test('T3756053: test environment administrator user roles', async () => {
            videoName = 'env-admin-user-roles';
            const copiedEnvName = `copied-${envName}`;

            await test.step(`verify ${envAdmin} can create an environment`, async () => {
                await envPage.goToPage();
                await envPage.clickOnAddNewEnvironmentBtn();
                await envPage.createNewEnvironment(envName, projectName, {
                    scope: Scope.Project,
                    scopeEntityName: projectName,
                    selectProjectScope: true,
                    command: sleep10s
                });
            });

            await test.step(`verify ${envAdmin} can edit an environment`, async () => {
                await envPage.openEditEnvironmentForm(envName);
                await envPage.clearImageInput();
                await envPage.fillEnvironmentImage(config.quickStartImage.ubuntu);
                await envPage.clickOnCreateEnvironmentBtn();

                const envData = await qTable.getEntityData('Environment', envName);
                expect(envData['Image']).toEqual(config.quickStartImage.ubuntu);
            });

            await test.step(`verify ${envAdmin} can copy an environment`, async () => {
                await envPage.copyEnvironment(envName);
                await envPage.clearEnvNameInput();
                await envPage.fillEnvironmentName(copiedEnvName);
                await envPage.clickOnCreateEnvironmentBtn();

                environmentUtils.waitForEnvironmentToAppear(copiedEnvName);
            });

            await test.step(`verify ${envAdmin} can delete an environment`, async () => {
                const namesToRemove = [copiedEnvName, envName];
                for (const name of namesToRemove) {
                    await envPage.removeEnvironment(name);
                }

                for (const name of namesToRemove) {
                    environmentUtils.waitForEnvironmentToBeDeleted(name);
                }
            });

            await test.step(`verify ${envAdmin} cannot see Resources menu`, async () => {
                await expect(
                    authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Resources}"]`)
                ).not.toBeVisible();
            });

            await test.step(`verify ${envAdmin} cannot see Access menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Access}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${envAdmin} cannot see Admin menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Admin}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${envAdmin} can see Analytics menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Analytics}"]`)).toBeVisible();
            });

            const viewOnlyTestData = [
                {
                    page: workloadPage,
                    pageName: 'workload page',
                    disabledElem: workloadPage.addNewWorkloadBtn
                },
                {
                    page: templatePage,
                    pageName: 'template page',
                    disabledElem: templatePage.newTemplateBtn
                },
                {
                    page: computeResourcePage,
                    pageName: 'compute resource page',
                    disabledElem: computeResourcePage.addNewComputeBtn
                },
                {
                    page: dataSourcePage,
                    pageName: 'data source page',
                    disabledElem: dataSourcePage.addNewDataSourceBtn
                },
                {
                    page: projectPage,
                    pageName: 'project page',
                    disabledElem: projectPage.addNewProjectBtn
                },
                {
                    page: departmentPage,
                    pageName: 'department page',
                    disabledElem: departmentPage.addNewDepartmentBtn
                }
            ];

            for (const stepData of viewOnlyTestData) {
                await test.step(`verify ${envAdmin} has view only permission on ${stepData.pageName}`, async () => {
                    await stepData.page.goToPage();
                    const viewOnlyBadgeShown = await stepData.page.viewOnlyBadgeVisible();
                    expect(viewOnlyBadgeShown).toBeTruthy();
                    const createBtnEnabled = await stepData.page.elementEnabled(stepData.disabledElem);
                    expect(createBtnEnabled).toBeFalsy();
                });
            }
        });
    });
});
