/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    credentialUtils,
    dataSourcesUtils,
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
import { Password, Scope } from '../../../../../../modules/api/apis/assetsApi';
import { MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { ComputeResourcePage } from '../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { DepartmentPage } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPage';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';

test.describe(`Data source administrator user - role testing, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.users}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let projectPage: ProjectPage;
        let departmentPage: DepartmentPage;
        let envPage: EnvironmentPage;
        let workloadPage: WorkloadPage;
        let computeResourcePage: ComputeResourcePage;
        let credentials: Password;
        let dataSourcePage: DataSourcePage;
        let templatePage: TemplatePageV2;
        let qTable: QTable;
        let projectName: string;
        let dsName: string;
        let namespace: string;
        let project: Project;
        let videoName: string;
        let credentialName: string;
        let credentialId: string;

        const mountPath = '/mnt/data';
        const dataSourceAdmin = 'data source admin user';

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            dsName = objectGenerator.generateShortNameByPrefix('ds', 5);
            credentialName = objectGenerator.generateShortNameByPrefix('cred', 5);

            projectName = objectGenerator.generateShortNameByPrefix('proj-data-source-admin', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            credentials = await credentialUtils.createUserPasswordWithArbitraryFields();
            credentialName = credentials.meta.name;
            credentialId = credentials.meta.id;
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
                flowRootState.users.data_source_administrator.email,
                flowRootState.users.data_source_administrator.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                credentialUtils.deletePasswordAsset(credentialId);
            } catch (error) {
                logger.error('failed to delete credentials.', error);
            }

            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test('T3756055: test data source administrator user roles', async () => {
            videoName = 'data-source-admin-user-roles';
            const newDsName = `new-${dsName}`;
            const copiedDsName = `copied-${dsName}`;

            await test.step(`verify ${dataSourceAdmin} can create data source using username & password secret`, async () => {
                await dataSourcePage.createSecretDataSource(
                    dsName,
                    Scope.Project,
                    projectName,
                    credentialName,
                    mountPath
                );
            });

            await test.step(`verify ${dataSourceAdmin} can rename a data source`, async () => {
                await dataSourcePage.renameDataSource(dsName, newDsName);
                // unselect the data source
                await dataSourcePage.clickOnElementByText('td', newDsName);

                const dsData = await qTable.getEntityData('Data source', newDsName);
                expect(dsData['Type']).toEqual('Secret');
            });

            await test.step(`verify ${dataSourceAdmin} can copy a data source`, async () => {
                await dataSourcePage.openCopyAndEditForm(newDsName);
                await dataSourcePage.clearSecretAssetNameInput();
                await dataSourcePage.fillSecretAssetNameInput(copiedDsName);
                await dataSourcePage.clickOnSubmitBtn();

                const dsData = await qTable.getEntityData('Data source', copiedDsName);
                expect(dsData['Type']).toEqual('Secret');
            });

            await test.step(`verify ${dataSourceAdmin} can delete a data source`, async () => {
                const dsToDelete = [newDsName, copiedDsName];
                for (const dsName of dsToDelete) {
                    await dataSourcePage.deleteDataSource(dsName);
                    await dataSourcesUtils.waitForDataSourceToBeDeleted(dsName);
                }
            });

            await test.step(`verify ${dataSourceAdmin} cannot see Resources menu`, async () => {
                await expect(
                    authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Resources}"]`)
                ).not.toBeVisible();
            });

            await test.step(`verify ${dataSourceAdmin} cannot see Access menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Access}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${dataSourceAdmin} cannot see Admin menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Admin}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${dataSourceAdmin} can see Analytics menu`, async () => {
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
                    page: envPage,
                    pageName: 'environment page',
                    disabledElem: envPage.addNewEnvironmentBtn
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
