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
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { ComputeResourcePage } from '../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { ApplicationsPage } from '../../../../../../modules/ui/runaiV2/pages/ApplicationPage';
import { ClusterPage } from '../../../../../../modules/ui/runaiV2/pages/ClusterPage';
import { UsersV2 } from '../../../../../../modules/ui/runaiV2/pages/UsersV2';
import { AccessRulesPage } from '../../../../../../modules/ui/runaiV2/pages/AccessRulesPage';
import { RolesPage } from '../../../../../../modules/ui/runaiV2/pages/RolesPage';
import { DepartmentPage } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPage';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';

test.describe(`Researcher manager user - role testing, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.users}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let projectPage: ProjectPage;
        let clusterPage: ClusterPage;
        let envPage: EnvironmentPage;
        let workloadPage: WorkloadPage;
        let computeResourcePage: ComputeResourcePage;
        let dataSourcePage: DataSourcePage;
        let credentialsPage: CredentialPage;
        let templatePage: TemplatePageV2;
        let applicationsPage: ApplicationsPage;
        let departmentPage: DepartmentPage;
        let userPage: UsersV2;
        let accessRulesPage: AccessRulesPage;
        let rolesPage: RolesPage;
        let projectName: string;
        let namespace: string;
        let project: Project;
        let videoName: string;

        const researcherManager = 'researcher manager user';

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            projectName = objectGenerator.generateShortNameByPrefix('proj-researcher-manager', 5);
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
            clusterPage = new ClusterPage(browserEngine.currentPage);
            envPage = new EnvironmentPage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);
            computeResourcePage = new ComputeResourcePage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);
            templatePage = new TemplatePageV2(browserEngine.currentPage);
            credentialsPage = new CredentialPage(browserEngine.currentPage);
            applicationsPage = new ApplicationsPage(browserEngine.currentPage);
            userPage = new UsersV2(browserEngine.currentPage);
            accessRulesPage = new AccessRulesPage(browserEngine.currentPage);
            rolesPage = new RolesPage(browserEngine.currentPage);
            departmentPage = new DepartmentPage(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.research_manager.email,
                flowRootState.users.research_manager.password
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

        test('T3756061: test research manager user roles', async () => {
            videoName = 'researcher-manager-user-roles';

            await test.step(`verify ${researcherManager} cannot see Admin menu`, async () => {
                await expect(
                    authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Resources}"]`)
                ).not.toBeVisible();
            });

            await test.step(`verify ${researcherManager} cannot see Admin menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Admin}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${researcherManager} can see Analytics menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Analytics}"]`)).toBeVisible();
            });

            await test.step(`${researcherManager} cannot see Users and Service accounts page`, async () => {
                await rolesPage.goToPage();
                await expect(authPage.page.locator(`[aid="${MENU_SUB_ITEMS.Users}"]`)).not.toBeVisible();
                await expect(
                    authPage.page.locator(`[aid="${MENU_SUB_ITEMS.ServiceAccounts}"]`)
                ).not.toBeVisible();
            });

            const viewOnlyTestData = [
                {
                    page: workloadPage,
                    pageName: 'workload page',
                    disabledElem: workloadPage.addNewWorkloadBtn
                },
                {
                    page: departmentPage,
                    pageName: 'department page',
                    disabledElem: departmentPage.addNewDepartmentBtn
                },
                {
                    page: credentialsPage,
                    pageName: 'credentials page',
                    enabledElem: credentialsPage.addNewCredentialBtn
                },
                {
                    page: rolesPage,
                    pageName: 'roles page',
                    disabledElem: undefined
                }
            ];

            for (const stepData of viewOnlyTestData) {
                await test.step(`verify ${researcherManager} has view only permission on ${stepData.pageName}`, async () => {
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
                    page: projectPage,
                    pageName: 'project page',
                    enabledElem: projectPage.addNewProjectBtn
                },
                {
                    page: accessRulesPage,
                    pageName: 'access rules page',
                    enabledElem: accessRulesPage.newAccessRulesBtn
                }
            ];

            for (const stepData of canCreateTestData) {
                await test.step(`verify ${researcherManager} has create permission on ${stepData.pageName}`, async () => {
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
