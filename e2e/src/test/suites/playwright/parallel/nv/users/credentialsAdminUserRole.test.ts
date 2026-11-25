/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    credentialUtils,
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
import { AssetSyncStatusEnum, Scope } from '../../../../../../modules/api/apis/assetsApi';
import { MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { ComputeResourcePage } from '../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { DepartmentPage } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPage';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';

test.describe(`Credentials administrator user - role testing, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.users}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let projectPage: ProjectPage;
        let departmentPage: DepartmentPage;
        let envPage: EnvironmentPage;
        let computeResourcePage: ComputeResourcePage;
        let workloadPage: WorkloadPage;
        let credentialsPage: CredentialPage;
        let dataSourcePage: DataSourcePage;
        let templatePage: TemplatePageV2;
        let projectName: string;
        let namespace: string;
        let project: Project;
        let videoName: string;
        let credentialName: string;
        let credentialDescription: string;
        let username: string;
        let password: string;

        const credentialsAdmin = 'credentials admin user';

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            credentialName = objectGenerator.generateShortNameByPrefix('user-pass', 5);
            credentialDescription = objectGenerator.generateShortNameByPrefix('user-pass-desc', 5);
            username = objectGenerator.generateNameByPrefix('user');
            password = objectGenerator.generateNameByPrefix('psswd');

            projectName = objectGenerator.generateShortNameByPrefix('proj-cred-admin', 5);
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
            credentialsPage = new CredentialPage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);
            computeResourcePage = new ComputeResourcePage(browserEngine.currentPage);
            templatePage = new TemplatePageV2(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.credentials_administrator.email,
                flowRootState.users.credentials_administrator.password
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

        test('T3756054: test credentials administrator user roles', async () => {
            videoName = 'credentials-admin-user-roles';
            const editedCredDesc = `new-${credentialDescription}`;

            await test.step(`verify ${credentialsAdmin} can create username & password secret`, async () => {
                await test.step('create username & password secret', async () => {
                    await credentialsPage.createUserPassCredential(
                        credentialName,
                        username,
                        password,
                        {
                            scope: Scope.Project,
                            scopeEntityName: project.name
                        },
                        credentialDescription
                    );

                    await credentialUtils.waitForPasswordStatus(
                        credentialName,
                        AssetSyncStatusEnum.NoIssuesFound
                    );
                });
            });

            await test.step(`verify ${credentialsAdmin} can edit credentials`, async () => {
                await credentialsPage.updateCredentialDesc(credentialName, editedCredDesc);
                await credentialsPage.clickOnRenameCredentialBtn();
                const desc = await credentialsPage.getEditModalDescriptionInputText();
                await credentialsPage.clickOnEditModalCancelBtn();

                expect(desc).toEqual(editedCredDesc);
            });

            await test.step(`verify ${credentialsAdmin} can delete credentials`, async () => {
                await credentialsPage.goToPage();
                await credentialsPage.deleteCredential(credentialName);
            });

            await test.step(`verify ${credentialsAdmin} cannot see Resources menu`, async () => {
                await expect(
                    authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Resources}"]`)
                ).not.toBeVisible();
            });

            await test.step(`verify ${credentialsAdmin} cannot see Access menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Access}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${credentialsAdmin} cannot see Admin menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Admin}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${credentialsAdmin} can see Analytics menu`, async () => {
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
                    page: envPage,
                    pageName: 'environment page',
                    disabledElem: envPage.addNewEnvironmentBtn
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
                await test.step(`verify data source admin user has view only permission on ${stepData.pageName}`, async () => {
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
