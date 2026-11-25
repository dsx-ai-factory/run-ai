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
import { QTable } from '../../../../../../modules/ui/runaiV2/components/QTable';
import { MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { ComputeResourcePage } from '../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { DepartmentPage } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPage';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';

test.describe(`Compute resource administrator user - role testing, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.users}`, () => {
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
        let computeResourceName: string;
        let namespace: string;
        let project: Project;
        let videoName: string;

        const computeResourceAdmin = 'compute resource admin user';
        const gpuRequest = 1;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            projectName = objectGenerator.generateShortNameByPrefix('proj-compute-resource-admin', 5);
            computeResourceName = objectGenerator.generateShortNameByPrefix('compute-res', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
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
                flowRootState.users.compute_resource_administrator.email,
                flowRootState.users.compute_resource_administrator.password
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

        test('T3756056: test compute resource administrator user roles', async () => {
            videoName = 'compute-resource-admin-user-roles';
            const newComputeResourceName = `new-${computeResourceName}`;
            const copiedComputeResourceName = `copied-${computeResourceName}`;

            await test.step(`verify ${computeResourceAdmin} can create a compute resource`, async () => {
                await computeResourcePage.goToPage();
                await computeResourcePage.addComputeResource();
                await computeResourcePage.createNewComputeResource(computeResourceName, {
                    scope: Scope.Project,
                    scopeEntityName: projectName
                });
            });

            await test.step(`verify ${computeResourceAdmin} can edit a compute resource`, async () => {
                await computeResourcePage.openEditForm(computeResourceName);
                await computeResourcePage.clearComputeResourceNameInput();
                await computeResourcePage.fillComputeName(newComputeResourceName);
                await computeResourcePage.clickOnCreateComputeBtn();

                const envData = await qTable.getEntityData('Compute resource', newComputeResourceName);
                expect(parseInt(envData['GPU devices request per pod'])).toEqual(gpuRequest);
            });

            await test.step(`verify ${computeResourceAdmin} can copy a compute resource`, async () => {
                await computeResourcePage.openCopyAndEditForm(newComputeResourceName);
                await computeResourcePage.clearComputeResourceNameInput();
                await computeResourcePage.fillComputeName(copiedComputeResourceName);
                await computeResourcePage.clickOnCreateComputeBtn();

                const envData = await qTable.getEntityData('Compute resource', copiedComputeResourceName);
                expect(parseInt(envData['GPU devices request per pod'])).toEqual(gpuRequest);
            });

            await test.step(`verify ${computeResourceAdmin} can delete a compute resource`, async () => {
                const computeResourceToDelete = [newComputeResourceName, copiedComputeResourceName];

                for (const compResName of computeResourceToDelete) {
                    await computeResourcePage.removeCompute(compResName);
                }
            });

            await test.step(`verify ${computeResourceAdmin} cannot see Resources menu`, async () => {
                await expect(
                    authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Resources}"]`)
                ).not.toBeVisible();
            });

            await test.step(`verify ${computeResourceAdmin} cannot see Access menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Access}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${computeResourceAdmin} cannot see Admin menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Admin}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${computeResourceAdmin} can see Analytics menu`, async () => {
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
                await test.step(`verify ${computeResourceAdmin} has view only permission on ${stepData.pageName}`, async () => {
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
