/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
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
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { NodePoolPage } from '../../../../../../modules/ui/runaiV2/pages/NodePoolPage';
import { NodePage } from '../../../../../../modules/ui/runaiV2/pages/NodePage';
import { GeneralSettings } from '../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { ApplicationsPage } from '../../../../../../modules/ui/runaiV2/pages/ApplicationPage';
import { ClusterPage } from '../../../../../../modules/ui/runaiV2/pages/ClusterPage';
import { UsersV2 } from '../../../../../../modules/ui/runaiV2/pages/UsersV2';
import { AccessRulesPage } from '../../../../../../modules/ui/runaiV2/pages/AccessRulesPage';

test.describe(`System administrator user - role testing, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.users}`, () => {
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
        let nodePoolPage: NodePoolPage;
        let nodePage: NodePage;
        let templatePage: TemplatePageV2;
        let generalSettingsPage: GeneralSettings;
        let applicationsPage: ApplicationsPage;
        let userPage: UsersV2;
        let accessRulesPage: AccessRulesPage;
        let projectName: string;
        let namespace: string;
        let project: Project;
        let videoName: string;

        const sysAdmin = 'system administrator user';

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            projectName = objectGenerator.generateShortNameByPrefix('proj-system-admin', 5);
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
            nodePoolPage = new NodePoolPage(browserEngine.currentPage);
            nodePage = new NodePage(browserEngine.currentPage);
            generalSettingsPage = new GeneralSettings(browserEngine.currentPage);
            applicationsPage = new ApplicationsPage(browserEngine.currentPage);
            userPage = new UsersV2(browserEngine.currentPage);
            accessRulesPage = new AccessRulesPage(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
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

        test('T3756057: test system administrator user roles', async () => {
            videoName = 'system-admin-user-roles';

            await test.step(`verify ${sysAdmin} can see Admin menus`, async () => {
                await generalSettingsPage.goToPage();
                await expect(
                    authPage.page.locator(`[aid="${MENU_SUB_ITEMS.GeneralSettings}"]`)
                ).toBeVisible();
                await expect(authPage.page.locator(`[aid="${MENU_SUB_ITEMS.Audit}"]`)).toBeVisible();
            });

            await test.step(`verify ${sysAdmin} can see Roles menu`, async () => {
                await applicationsPage.goToPage();
                await expect(authPage.page.locator(`[aid="${MENU_SUB_ITEMS.Roles}"]`)).toBeVisible();
            });

            await test.step(`verify ${sysAdmin} can see Nodes menu`, async () => {
                await nodePage.goToPage();
                await expect(authPage.page.locator(`[aid="${MENU_SUB_ITEMS.Nodes}"]`)).toBeVisible();
            });

            await test.step(`verify ${sysAdmin} can see Analytics menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Analytics}"]`)).toBeVisible();
            });

            await test.step(`verify ${sysAdmin} can see training/workspace/inference menus and they are enabled`, async () => {
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
                    page: clusterPage,
                    pageName: 'cluster page',
                    enabledElem: clusterPage.addNewClusterBtn
                },
                {
                    page: nodePoolPage,
                    pageName: 'node pools page',
                    enabledElem: nodePoolPage.addNewNodePoolBtn
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
                await test.step(`verify ${sysAdmin} has create permission on ${stepData.pageName}`, async () => {
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
