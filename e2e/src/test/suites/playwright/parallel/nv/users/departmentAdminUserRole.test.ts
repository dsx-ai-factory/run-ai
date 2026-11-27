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
    usersManager,
    Utils,
    waitUtils
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
import { MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { ComputeResourcePage } from '../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { DepartmentPage } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPage';
import { NodePoolPage } from '../../../../../../modules/ui/runaiV2/pages/NodePoolPage';
import { NodePage } from '../../../../../../modules/ui/runaiV2/pages/NodePage';
import { RolesPage } from '../../../../../../modules/ui/runaiV2/pages/RolesPage';
import { UsersV2 } from '../../../../../../modules/ui/runaiV2/pages/UsersV2';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { AccessRulesPage } from '../../../../../../modules/ui/runaiV2/pages/AccessRulesPage';
import { ApplicationsPage } from '../../../../../../modules/ui/runaiV2/pages/ApplicationPage';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';

test.describe(`Department administrator user - role testing, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.users}`, () => {
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
        let nodePoolPage: NodePoolPage;
        let nodePage: NodePage;
        let rolesPage: RolesPage;
        let userPage: UsersV2;
        let credentialsPage: CredentialPage;
        let accessRulesPage: AccessRulesPage;
        let applicationsPage: ApplicationsPage;
        let projectName: string;
        let namespace: string;
        let project: Project;
        let videoName: string;

        const departmentAdmin = 'department admin user';

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            projectName = objectGenerator.generateShortNameByPrefix('proj-depmt-admin', 5);
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
            nodePoolPage = new NodePoolPage(browserEngine.currentPage);
            nodePage = new NodePage(browserEngine.currentPage);
            rolesPage = new RolesPage(browserEngine.currentPage);
            userPage = new UsersV2(browserEngine.currentPage);
            credentialsPage = new CredentialPage(browserEngine.currentPage);
            accessRulesPage = new AccessRulesPage(browserEngine.currentPage);
            applicationsPage = new ApplicationsPage(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.department_admin.email,
                flowRootState.users.department_admin.password
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

        test('T3756058: test department administrator user roles', async () => {
            videoName = 'department-admin-user-roles';
            const username = 'test-user333@example.com';
            const namePart = username.split('@')[0];

            await test.step(`verify ${departmentAdmin} can create a user`, async () => {
                const response = (
                    await usersManager.createLocalUser(
                        flowRootState.users.department_admin.accessToken,
                        username
                    )
                ).body;
                expect(response.username).toEqual(username);
            });

            await test.step(`verify ${departmentAdmin} can delete created user`, async () => {
                await userPage.goToPage();
                await userPage.searchByName(username);
                await userPage.clickOnUser(username);
                await userPage.clickOnDeleteUserButton();
                await waitUtils.waitForElementToNotExist(userPage, namePart);
                await userPage.clearSearchInput();
            });

            await test.step(`verify ${departmentAdmin} can see Admin menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Admin}"]`)).toBeVisible();
            });

            await test.step(`verify ${departmentAdmin} can see Analytics menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Analytics}"]`)).toBeVisible();
            });

            await test.step(`verify ${departmentAdmin} can see training/workspace/inference menus and they are enabled`, async () => {
                await workloadPage.goToPage();
                await workloadPage.locator(workloadPage.addNewWorkloadBtn).click();
                await expect(workloadPage.locator(workloadPage.newTrainingBtn)).not.toHaveAttribute(
                    'class',
                    /disabled/
                );
                await expect(workloadPage.locator(workloadPage.newWorkspaceBtn)).not.toHaveAttribute(
                    'class',
                    /disabled/
                );
                await expect(workloadPage.locator(workloadPage.newInferenceBtn)).not.toHaveAttribute(
                    'class',
                    /disabled/
                );
            });

            const viewOnlyTestData = [
                {
                    page: departmentPage,
                    pageName: 'department page',
                    disabledElem: departmentPage.addNewDepartmentBtn
                },
                {
                    page: nodePoolPage,
                    pageName: 'node pools page',
                    disabledElem: nodePoolPage.addNewNodePoolBtn
                },
                {
                    page: nodePage,
                    pageName: 'node page',
                    disabledElem: undefined
                },
                {
                    page: rolesPage,
                    pageName: 'roles page',
                    disabledElem: undefined
                }
            ];

            for (const stepData of viewOnlyTestData) {
                await test.step(`verify ${departmentAdmin} has view only permission on ${stepData.pageName}`, async () => {
                    await stepData.page.goToPage();
                    const viewOnlyBadgeShown = await stepData.page.viewOnlyBadgeVisible();
                    expect(viewOnlyBadgeShown).toBeTruthy();

                    if (stepData.disabledElem) {
                        const createBtnEnabled = await stepData.page.elementEnabled(stepData.disabledElem);
                        expect(createBtnEnabled).toBeFalsy();
                    }
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
                },
                {
                    page: projectPage,
                    pageName: 'project page',
                    enabledElem: projectPage.addNewProjectBtn
                },
                {
                    page: accessRulesPage,
                    pageName: 'access rules page',
                    enabledElem: accessRulesPage.newAccessRulesBtn
                },
                {
                    page: userPage,
                    pageName: 'users page',
                    enabledElem: userPage.addNewUserButton
                },
                {
                    page: applicationsPage,
                    pageName: 'applications page',
                    enabledElem: applicationsPage.newApplicationBtn
                }
            ];

            for (const stepData of canCreateTestData) {
                await test.step(`verify ${departmentAdmin} has create permission on ${stepData.pageName}`, async () => {
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
