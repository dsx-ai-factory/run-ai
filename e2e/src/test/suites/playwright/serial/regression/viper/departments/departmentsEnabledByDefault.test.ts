/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
    BaseWorkload,
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    orgUnitUtils,
    tags,
    TrainingWorkload,
    Utils,
    waitUtils,
    workloadUtils
} from '../../../../../../../modules/helpers';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    EWidgetComponentName,
    OverviewPage
} from '../../../../../../../modules/ui/runaiV2/pages/OverviewPage';
import { test, expect } from '@playwright/test';
import { App } from '../../../../../../../modules/ui/runaiV2/pages/App';
import { DepartmentPage } from '../../../../../../../modules/ui/runaiV2/pages/DepartmentPage';
import { DepartmentForList, Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { kubeCli } from '../../../../../../../modules/cli';
import {
    EQuotaWidgetComponentName,
    QuotaManagementPage
} from '../../../../../../../modules/ui/runaiV2/pages/QuotaManagementPage';
import { departmentUiUtils } from '../../../../../../../modules/ui/utils/departmentUiUtils';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { shouldRunNativeWorkloadCrdTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe.serial(
    `Departments UI, tags: ${tags.tool.ui} ${tags.resource.overview} ${tags.teams.viper}`,
    { tag: [tags.services.orgUnitService, tags.runOn.serialCI] },
    () => {
        const defaultDepartmentName: string = 'default';
        const clickWaitTime = 1000;
        const drillDownWaitTime = 2000;

        let browserEngine: BrowserEngine;
        let appPage: App;
        let departmentPage: DepartmentPage;
        let project1: Project;
        let project2: Project;
        let project3: Project;
        let projects: Project[] = [];
        let department: DepartmentForList;
        let workloads: BaseWorkload[] = [];
        let authPage: AuthPage;
        let videoName: string;
        let overviewPage: OverviewPage;
        let quotaManagementPage: QuotaManagementPage;

        test.beforeAll(async () => {
            try {
                await operations.forceCleanup();
            } catch (error) {
                logger.error(error);
            }

            // Enable manual over quota feature to allow department creation with over quota weights
            await apiUtils.enableProjectOverQuotaFeature();

            await orgUnitUtils.validateAndSetDefaultDepartmentGpu();

            browserEngine = new BrowserEngine();

            // create 2 projects on default department
            let projectName = objectGenerator.generateShortNameByPrefix('p1');
            project1 = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                resources: [{ gpu: { deserved: 4 } }],
                clusterId: flowRootState.cluster.uuid
            });
            projects.push(project1);
            projectName = objectGenerator.generateShortNameByPrefix('p2');
            project2 = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                resources: [{ gpu: { deserved: 4 } }],
                clusterId: flowRootState.cluster.uuid
            });
            projects.push(project2);
        });

        test.afterAll(async () => {
            try {
                await operations.deleteWorkloads(Object.values(workloads));
            } catch (error) {
                logger.error(error);
            }
            workloads = [];
            for (const project of projects) {
                try {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (error) {
                    logger.error(error);
                }
            }
            projects = [];

            // Disable manual over quota feature after tests
            try {
                await apiUtils.disableProjectOverQuotaFeature();
            } catch (error) {
                logger.error('Error disabling project over quota feature: ' + error);
            }
        });

        const startSession = async function (): Promise<void> {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            const systemAdmin = flowRootState.users.system_administrator;
            await authPage.signIn(systemAdmin.email, systemAdmin.password);
        };

        async function createRunningWorkload(project: Project, gpu: string): Promise<TrainingWorkload> {
            const workload = new TrainingWorkload(
                config.quickStartImage.ubuntu,
                Utils.namespaceFromProjectName(project.name),
                objectGenerator.generateShortNameByPrefix('w')
            );
            workload.arguments = 'sleep infinity';
            workload.gpu = gpu;
            await kubeCli.createWorkload(workload);
            await workloadUtils.waitForWorkloadBackendPhase(workload.name, project.name, Phase.Running);
            return workload;
        }

        async function waitForExpectedWidgetLabels(
            getLabels: () => Promise<string>,
            expectedLabels: string[],
            description: string
        ): Promise<string> {
            await waitUtils.waitForCondition(
                async () => {
                    const labels = await getLabels();
                    return expectedLabels.every((expectedLabel) => labels.includes(expectedLabel));
                },
                description,
                true,
                60
            );

            return getLabels();
        }

        async function validateOverviewDrillDown(
            widgetComponentName: EWidgetComponentName,
            firstDrillDown1: string,
            firstDrillDown2: string,
            secondDrillDown1: string = undefined,
            secondDrillDown2: string = undefined
        ): Promise<void> {
            await waitUtils.waitForCondition(
                async () => {
                    const isVisible = await overviewPage.isDrillableCategoryVisible(
                        widgetComponentName,
                        config.clusterNamePrefix
                    );
                    if (!isVisible) {
                        await overviewPage.page.reload();
                        await overviewPage.page.waitForLoadState('networkidle');
                    }
                    return isVisible;
                },
                `wait for ${widgetComponentName} x-axis labels to be visible`,
                true,
                60
            );

            let xLabel = await overviewPage.getWidgetXLabel(widgetComponentName);
            expect(xLabel).toContain(config.clusterNamePrefix);

            await overviewPage.clickWidgetXLabel(widgetComponentName, config.clusterNamePrefix);
            xLabel = await waitForExpectedWidgetLabels(
                () => overviewPage.getWidgetXLabel(widgetComponentName),
                [firstDrillDown1, firstDrillDown2],
                `wait for ${widgetComponentName} first drilldown x-axis labels`
            );
            expect(xLabel).toContain(firstDrillDown1);
            expect(xLabel).toContain(firstDrillDown2);

            if (secondDrillDown1) {
                await overviewPage.clickWidgetXLabelWithText(widgetComponentName, firstDrillDown1);
                const expectedLabels = secondDrillDown2
                    ? [secondDrillDown1, secondDrillDown2]
                    : [secondDrillDown1];
                xLabel = await waitForExpectedWidgetLabels(
                    () => overviewPage.getWidgetXLabel(widgetComponentName),
                    expectedLabels,
                    `wait for ${widgetComponentName} second drilldown x-axis labels`
                );
                expect(xLabel).toContain(secondDrillDown1);
                if (secondDrillDown2) {
                    expect(xLabel).toContain(secondDrillDown2);
                }
            }
        }

        async function validateQuotaManagementDrillDown(
            widgetComponentName: EQuotaWidgetComponentName,
            firstDrillDown1: string,
            firstDrillDown2: string,
            secondDrillDown1: string = undefined,
            secondDrillDown2: string = undefined
        ): Promise<void> {
            await waitUtils.waitForCondition(
                () =>
                    quotaManagementPage.isDrillableCategoryVisible(
                        widgetComponentName,
                        config.clusterNamePrefix
                    ),
                `wait for ${widgetComponentName} x-axis labels to be visible`,
                true,
                60
            );

            let xLabel = await quotaManagementPage.getWidgetXLabel(widgetComponentName);
            expect(xLabel).toContain(config.clusterNamePrefix);

            await quotaManagementPage.clickWidgetXLabel(widgetComponentName, config.clusterNamePrefix);
            xLabel = await waitForExpectedWidgetLabels(
                () => quotaManagementPage.getWidgetXLabel(widgetComponentName),
                [firstDrillDown1, firstDrillDown2],
                `wait for ${widgetComponentName} first drilldown x-axis labels`
            );
            expect(xLabel).toContain(firstDrillDown1);
            expect(xLabel).toContain(firstDrillDown2);

            if (secondDrillDown1) {
                await quotaManagementPage.clickWidgetXLabelWithText(widgetComponentName, firstDrillDown1);
                const expectedLabels = secondDrillDown2
                    ? [secondDrillDown1, secondDrillDown2]
                    : [secondDrillDown1];
                xLabel = await waitForExpectedWidgetLabels(
                    () => quotaManagementPage.getWidgetXLabel(widgetComponentName),
                    expectedLabels,
                    `wait for ${widgetComponentName} second drilldown x-axis labels`
                );
                expect(xLabel).toContain(secondDrillDown1);
                if (secondDrillDown2) {
                    expect(xLabel).toContain(secondDrillDown2);
                }
            }
        }

        test.describe.serial('test departments ui - one department', () => {
            test.beforeAll(async () => {
                await startSession();

                overviewPage = new OverviewPage(browserEngine.currentPage);
                quotaManagementPage = new QuotaManagementPage(browserEngine.currentPage);
            });

            test.afterAll(async ({}, testInfo) => {
                try {
                    await browserEngine.stopAndArchive(testInfo, 'departments-' + videoName);
                    videoName = 'departments ui';
                } catch (e) {
                    logger.error(e);
                }
            });

            shouldRunNativeWorkloadCrdTests(() =>
                test('validate overview drill down with one department', async () => {
                    // create workload on each project
                    const workload1 = await createRunningWorkload(project1, '1');
                    workloads.push(workload1);
                    const workload2 = await createRunningWorkload(project2, '1');
                    workloads.push(workload2);

                    await waitUtils.sleep(drillDownWaitTime);
                    await overviewPage.goToPage();

                    await validateOverviewDrillDown(
                        EWidgetComponentName.WorkloadStatus,
                        project1.name,
                        project2.name
                    );
                    await validateOverviewDrillDown(
                        EWidgetComponentName.WorkloadByTypeWidget,
                        project1.name,
                        project2.name
                    );
                })
            );

            test('validate quota management drill down with one department', async () => {
                appPage = new App(browserEngine.currentPage);

                await quotaManagementPage.goToPage();
                await validateQuotaManagementDrillDown(
                    EQuotaWidgetComponentName.QuotaByNodepoolWidget,
                    project1.name,
                    project2.name
                );
                await validateQuotaManagementDrillDown(
                    EQuotaWidgetComponentName.AllocationByNodepoolWidget,
                    project1.name,
                    project2.name
                );
                // PendingWorkloadsByNodepoolWidget is intentionally not drilled here:
                // its chart x-axis is empty when there are no pending workloads
                // (this flow only creates running workloads). The widget's render
                // is covered by quotaManagementUI.test.ts.
                // DeptAllocationRatioWidget is conditional on a non-default
                // department existing — covered by the multi-department test below.
                expect(
                    await appPage
                        .locator(
                            quotaManagementPage.getWidgetSelector(
                                EQuotaWidgetComponentName.ProjectAllocationRatioWidget
                            )
                        )
                        .isVisible()
                ).toBe(true);
            });
        });

        test.describe.serial('test departments ui - more than one department', () => {
            test.beforeAll(async () => {
                await startSession();

                appPage = new App(browserEngine.currentPage);
                overviewPage = new OverviewPage(browserEngine.currentPage);
                quotaManagementPage = new QuotaManagementPage(browserEngine.currentPage);
            });

            test.afterAll(async ({}, testInfo) => {
                try {
                    await browserEngine.stopAndArchive(testInfo, 'departments-' + videoName);
                    videoName = 'departments ui';
                } catch (e) {
                    logger.error(e);
                }
                try {
                    await orgUnitUtils.deleteDepartment(department.id);
                } catch (error) {
                    logger.error(error);
                }
                department = null;
            });

            test('departments grid', async () => {
                departmentPage = new DepartmentPage(browserEngine.currentPage);
                await departmentPage.goToPage();
                await waitUtils.sleep(3 * clickWaitTime);
                const isDepartmentBtnEnabled = await departmentPage
                    .locator(departmentPage.addNewDepartmentBtn)
                    .isEnabled({ timeout: 10000 });
                expect(isDepartmentBtnEnabled).toBe(true);

                const departmentName = objectGenerator.generateShortNameByPrefix('d');
                await departmentUiUtils.createBasicDepartment(departmentPage, departmentName);
                await orgUnitUtils.waitForDepartmentByNameInOrgUnitService(departmentName);

                await waitUtils.sleep(clickWaitTime);
                department = (await orgUnitUtils.getDepartments()).find(
                    (departmentFromGet) => departmentFromGet.name === departmentName
                );
                expect(department).toBeDefined();
            });

            shouldRunNativeWorkloadCrdTests(() =>
                test('validate overview drill down with more than one department', async () => {
                    const projectName = objectGenerator.generateShortNameByPrefix('p3');
                    project3 = await orgUnitUtils.createProjectAndValidate({
                        name: projectName,
                        resources: [{ gpu: { deserved: 4 } }],
                        clusterId: flowRootState.cluster.uuid,
                        parentId: department.id
                    });
                    projects.push(project3);
                    const workload = await createRunningWorkload(project3, '1');
                    workloads.push(workload);

                    await waitUtils.sleep(drillDownWaitTime);
                    await overviewPage.goToPage();
                    await validateOverviewDrillDown(
                        EWidgetComponentName.WorkloadStatus,
                        defaultDepartmentName,
                        department.name,
                        project1.name,
                        project2.name
                    );
                    await validateOverviewDrillDown(
                        EWidgetComponentName.WorkloadByTypeWidget,
                        defaultDepartmentName,
                        department.name,
                        project1.name,
                        project2.name
                    );
                })
            );

            test('validate quota management drill down with more than one department', async ({}, testInfo) => {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'QuotaByNodepoolWidget only lists departments that have projects with quota, so a newly created department never appears in the drilldown'
                );
                await quotaManagementPage.goToPage();
                await validateQuotaManagementDrillDown(
                    EQuotaWidgetComponentName.QuotaByNodepoolWidget,
                    defaultDepartmentName,
                    department.name,
                    project1.name,
                    project2.name
                );
                await validateQuotaManagementDrillDown(
                    EQuotaWidgetComponentName.AllocationByNodepoolWidget,
                    defaultDepartmentName,
                    department.name,
                    project1.name,
                    project2.name
                );
                // PendingWorkloadsByNodepoolWidget drilldown is omitted here
                // because the test fixtures only spawn running workloads, leaving
                // its chart x-axis is empty. Visibility is covered by
                // quotaManagementUI.test.ts.
                expect(
                    await appPage
                        .locator(
                            quotaManagementPage.getWidgetSelector(
                                EQuotaWidgetComponentName.DeptAllocationRatioWidget
                            )
                        )
                        .isVisible()
                ).toBe(true);
                expect(
                    await appPage
                        .locator(
                            quotaManagementPage.getWidgetSelector(
                                EQuotaWidgetComponentName.ProjectAllocationRatioWidget
                            )
                        )
                        .isVisible()
                ).toBe(true);
            });
        });
    }
);
