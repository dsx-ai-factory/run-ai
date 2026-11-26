/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    inferenceUtils,
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
import { MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { ComputeResourcePage } from '../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { DepartmentPage } from '../../../../../../modules/ui/runaiV2/pages/DepartmentPage';
import { TemplatePageV2 } from '../../../../../../modules/ui/runaiV2/pages/TemplatePageV2';
import { Phase, ServingPortProtocol } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { NodePage } from '../../../../../../modules/ui/runaiV2/pages/NodePage';
import { NodePoolPage } from '../../../../../../modules/ui/runaiV2/pages/NodePoolPage';

test.describe(`ML engineer user - role testing, tags: ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.users}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let nodePage: NodePage;
        let nodePoolPage: NodePoolPage;
        let projectPage: ProjectPage;
        let departmentPage: DepartmentPage;
        let envPage: EnvironmentPage;
        let workloadPage: WorkloadPage;
        let computeResourcePage: ComputeResourcePage;
        let dataSourcePage: DataSourcePage;
        let templatePage: TemplatePageV2;
        let credentialsPage: CredentialPage;
        let projectName: string;
        let workloadName: string;
        let workloadId: string;
        let namespace: string;
        let project: Project;
        let videoName: string;

        const mlEngineer = 'ML engineer';

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();

            workloadName = objectGenerator.generateShortNameByPrefix('inference', 5);
            projectName = objectGenerator.generateShortNameByPrefix('proj-ml-engineer', 5);
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
            credentialsPage = new CredentialPage(browserEngine.currentPage);
            nodePage = new NodePage(browserEngine.currentPage);
            nodePoolPage = new NodePoolPage(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.ml_engineer.email,
                flowRootState.users.ml_engineer.password
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

        test('T3756063: test ML engineer user roles', async () => {
            videoName = 'ml-engineer-user-roles';

            await test.step(`verify ${mlEngineer} can create inference workload`, async () => {
                [workloadId, workloadName] = await inferenceUtils.createInference(
                    workloadName,
                    {
                        image: config.quickStartImage.inferenceHelloWorld,
                        servingPort: { container: 8033, protocol: ServingPortProtocol.Http }
                    },
                    project.id.toString(),
                    false,
                    flowRootState.users.ml_engineer.accessToken
                );

                await waitUtils.waitForCondition(
                    async () =>
                        (await workloadsCrudUtils.getInference(workloadId)).body.actualPhase ===
                        Phase.Running,
                    `Wait for workload ${workloadName} to be Running`,
                    true,
                    60
                );
            });

            await test.step(`verify ${mlEngineer} can delete inference workload`, async () => {
                await workloadPage.goToPage();
                await workloadPage.deleteWorkload(workloadName);
                await waitUtils.waitForElementToNotExist(workloadPage, workloadName);
            });

            await test.step(`verify ${mlEngineer} can not create training workload`, async () => {
                await workloadPage.openTrainingForm();
                await workloadPage.waitForVisible(workloadPage.locator(workloadPage.noRolesDiv));
                await workloadPage.page.goBack();
            });

            await test.step(`verify ${mlEngineer} can not create workspace workload`, async () => {
                await workloadPage.openWorkspaceForm();
                await workloadPage.waitForVisible(workloadPage.locator(workloadPage.noRolesDiv));
                await workloadPage.page.goBack();
            });

            await test.step(`verify ${mlEngineer} cannot see Access menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Access}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${mlEngineer} cannot see Admin menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Admin}"]`)).not.toBeVisible();
            });

            await test.step(`verify ${mlEngineer} can see Analytics menu`, async () => {
                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Analytics}"]`)).toBeVisible();
            });

            const viewOnlyTestData = [
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
                await test.step(`verify ML engineer user has view only permission on ${stepData.pageName}`, async () => {
                    await stepData.page.goToPage();
                    const viewOnlyBadgeShown = await stepData.page.viewOnlyBadgeVisible();
                    expect(viewOnlyBadgeShown).toBeTruthy();
                    const createBtnEnabled = await stepData.page.elementEnabled(stepData.disabledElem);
                    expect(createBtnEnabled).toBeFalsy();
                });
            }

            await test.step('verify ML engineer user has create permission on Credential page', async () => {
                await credentialsPage.goToPage();

                const createBtnEnabled = await credentialsPage.elementEnabled(
                    credentialsPage.addNewCredentialBtn
                );
                expect(createBtnEnabled).toBeTruthy();
            });
        });
    });
});
