/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    flowRootState,
    logger,
    nodePoolUtils,
    operations,
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
    generateNodePools,
    generateProjects,
    generateWorkloads,
    IWidgetTest,
    OverviewPage
} from '../../../../../../../modules/ui/runaiV2/pages/OverviewPage';
import { kubeCli } from '../../../../../../../modules/cli';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { shouldRunNativeWorkloadCrdTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Overview UI, tags: ${tags.tool.ui} ${tags.resource.overview} ${tags.teams.viper}`, () => {
    let browserEngine: BrowserEngine;
    let authPage: AuthPage;
    let overviewPage: OverviewPage;
    let totalNumberOfNodesWithGpus = 0;
    let videoName: string;

    test.beforeAll(async function () {
        try {
            await operations.forceCleanup();
        } catch (error) {
            logger.error(error);
        }

        browserEngine = new BrowserEngine();
        totalNumberOfNodesWithGpus = await kubeCli.getNumOfNodesWithGpus();
    });

    const startSession = async function (): Promise<void> {
        await browserEngine.start();
        authPage = new AuthPage(browserEngine.currentPage);
        overviewPage = new OverviewPage(browserEngine.currentPage);
        await authPage.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
    };

    const validateEmptyWidgetsState = async function (): Promise<void> {
        for (const widgetName of [
            EWidgetComponentName.WorkloadByTypeWidget,
            EWidgetComponentName.WorkloadStatus
        ]) {
            if (
                widgetName === EWidgetComponentName.GpuAllocationRatioWidget &&
                !overviewPage.isDepartmentMetricSupported()
            ) {
                logger.info('GpuAllocationRatioWidget is not supported in this version, skipping...');
                test.skip();
                return;
            }
            logger.info(`Validating empty state for widget: ${widgetName}`);
            await waitUtils.waitForElementToNotExist(
                overviewPage,
                `${overviewPage.getWidgetSelector(widgetName)} .loading-section`
            );
            await expect(
                overviewPage.page.locator(
                    `${overviewPage.getWidgetSelector(widgetName)} [aid="empty-widget-message"]`
                )
            ).toBeVisible();
        }
    };

    test.describe('overview page initial widgets state', function () {
        test.beforeEach(async function () {
            await startSession();
        });

        test.afterEach(async ({}, testInfo) => {
            try {
                await browserEngine.stopAndArchive(testInfo, 'overview-' + videoName);
            } catch (e) {
                logger.error(e);
            }
        });

        test('should redirect to overview page', async () => {
            videoName = 'redirect-overview-page';
            const title = await overviewPage.getPageTitle();
            expect(title).toContain('Overview');
        });

        test('should verify widgets presence', async () => {
            videoName = 'widgets-presence';
            for (const name of overviewPage.widgetsNames) {
                await overviewPage.page.waitForSelector(`[aid="${name}"]`);
            }
        });

        test('should validate correct number of Ready GPU nodes', async () => {
            videoName = 'number-of-ready-gpu-nodes';
            const totalNumberOfGpus = await kubeCli.getNumOfGpus();
            const widgetText = await overviewPage.getStatWidgetText(EWidgetComponentName.ReadyGpuWidget);
            expect(widgetText).toContain(totalNumberOfGpus.toString());
        });

        test('should validate correct number of Ready GPU devices', async () => {
            videoName = 'number-of-ready-gpu-devices';
            const widgetText = await overviewPage.getStatWidgetText(EWidgetComponentName.ReadyNodesWidget);
            expect(widgetText).toContain(totalNumberOfNodesWithGpus.toString());
        });

        test('should validate correct number of Allocated GPU devices', async () => {
            videoName = 'number-of-allocated-gpu-devices';
            const widgetText = await overviewPage.getStatWidgetText(EWidgetComponentName.AllocatedGpuWidget);
            expect(widgetText).toContain('0');
        });

        test('should validate correct number of Running Workloads', async () => {
            videoName = '0-running-workloads';
            const widgetText = await overviewPage.getStatWidgetText(EWidgetComponentName.RunningWorkloads);
            expect(widgetText).toBe('0');
        });

        test('should validate correct number of Pending Workloads', async () => {
            videoName = '0-pending-workloads';
            const widgetText = await overviewPage.getStatWidgetText(EWidgetComponentName.PendingWorkloads);
            expect(widgetText).toBe('0');
        });

        test('should validate node pool filter is display or hidden correctly', async () => {
            videoName = 'node-pool-filter';
            const nodePoolFilter = overviewPage.page.locator('.runai-page-filter');
            if (!overviewPage.isNodePoolFilterSupported()) {
                await expect(nodePoolFilter).not.toBeVisible();
            } else {
                await expect(nodePoolFilter).toBeVisible();
            }
        });
    });

    shouldRunNativeWorkloadCrdTests(() =>
    test.describe.serial('overview page with running workloads', () => {
        const waitTimeBeforeRunningTests = 20 * 1000; // 20 seconds
        let projects: Array<Project>;
        const projectCount = 2;
        const nodepoolsCount = 3;
        let nodepools: { nodePoolName: string; nodePoolId: number }[] = [];

        let runningWorkloads: TrainingWorkload[] = [];
        let pendingWorkloads: TrainingWorkload[] = [];

        let widgetTests: IWidgetTest[] = [];

        let initialData: IWidgetTest[] = [
            {
                name: EWidgetComponentName.RunningWorkloads,
                desiredResult: '0'
            },
            {
                name: EWidgetComponentName.PendingWorkloads,
                desiredResult: '0'
            },
            {
                name: EWidgetComponentName.AllocatedGpuWidget,
                desiredResult: '0'
            }
        ];

        test.beforeAll(async function ({}, testInfo) {
            try {
                // Current state of the env before creating workloads
                await startSession();
                const promises = initialData.map((widget: IWidgetTest) => {
                    return overviewPage
                        .getStatWidgetText(widget.name)
                        .then((desiredResult) => ({ ...widget, desiredResult }));
                });
                initialData = await Promise.all(promises);
            } catch (error) {
                logger.error('Failed to get initial data', JSON.stringify(error));
            }

            try {
                await browserEngine.stopAndArchive(testInfo, 'overview-data-init');
            } catch (error) {
                logger.error('Failed to stop and archive browser engine', JSON.stringify(error));
            }

            nodepools = await generateNodePools(nodepoolsCount);
            projects = await generateProjects(
                flowRootState.users.system_administrator.accessToken,
                projectCount,
                nodepools
            );

            logger.info('creating projects for metrics api tests');

            //submit workloads
            [runningWorkloads, pendingWorkloads] = await generateWorkloads(
                flowRootState.users.system_administrator.accessToken,
                projects,
                nodepools,
                runningWorkloads,
                pendingWorkloads
            );

            // Calculate expected workload counts based on created workloads
            const expectedRunningWorkloadCount = runningWorkloads.length;
            const expectedPendingWorkloadCount = pendingWorkloads.length;

            logger.info(
                `set up environment for metrics api tests successfully, waiting ${
                    waitTimeBeforeRunningTests / 1000
                } seconds before running tests. Expected running workloads: ${expectedRunningWorkloadCount}, Expected pending workloads: ${expectedPendingWorkloadCount}`
            );
            await Utils.delay(waitTimeBeforeRunningTests);

            // Update widget tests with actual workload counts
            widgetTests = [
                {
                    name: EWidgetComponentName.RunningWorkloads,
                    desiredResult: expectedRunningWorkloadCount.toString()
                },
                {
                    name: EWidgetComponentName.PendingWorkloads,
                    desiredResult: expectedPendingWorkloadCount.toString()
                },
                {
                    name: EWidgetComponentName.AllocatedGpuWidget,
                    desiredResult: expectedRunningWorkloadCount.toString()
                }
            ];
        });

        test.afterAll(async () => {
            try {
                await Promise.all(
                    projects.map((project) => operations.deleteProjectAndNamespace(project.id))
                );
            } catch (error) {
                logger.warn('Delete Project And Namespace failed', error);
            }

            try {
                await nodePoolUtils.deleteAllNodePools();
            } catch (error) {
                logger.warn('Delete Node pools failed ', error);
            }

            try {
                await Promise.all([
                        ...runningWorkloads.map((workload: TrainingWorkload) =>
                            workloadUtils.delete(workload)
                        ),
                        ...pendingWorkloads.map((workload: TrainingWorkload) =>
                            workloadUtils.delete(workload)
                        )
                ]);
            } catch (error) {
                logger.warn('Delete Workloads failed ', error);
            }
        });

        test.beforeEach(async function () {
            await startSession();
        });

        test.afterEach(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'overview-' + videoName);
        });

        const validateStatData = async (widget: IWidgetTest): Promise<boolean> => {
            await waitUtils.waitForElementToExist(
                overviewPage,
                `${overviewPage.getWidgetSelector(widget.name)} .stat-data`
            );

            const oldData = Number(initialData.find((data) => data.name === widget.name)?.desiredResult);

            const conditionMet =
                Number(await overviewPage.getStatWidgetText(widget.name)) - oldData ===
                Number(widget.desiredResult);
            if (!conditionMet) {
                await overviewPage.page.reload();
            }
            return conditionMet;
        };

        test('should validate correct number of running workloads', async function () {
            videoName = 'number-of-running-workloads';
            const widget = widgetTests.find((w) => w.name === EWidgetComponentName.RunningWorkloads);
            if (!widget) {
                throw new Error('Running workloads widget test not found');
            }
            await waitUtils.waitForCondition(
                async () => validateStatData(widget),
                `Wait for ${widget.name} widgets to display correct number (${widget.desiredResult})`,
                true,
                60
            );
        });

        test('should validate correct number of pending workloads', async function () {
            videoName = 'number-of-pending-workloads';
            const widget = widgetTests.find((w) => w.name === EWidgetComponentName.PendingWorkloads);
            if (!widget) {
                throw new Error('Pending workloads widget test not found');
            }
            await waitUtils.waitForCondition(
                async () => validateStatData(widget),
                `Wait for ${widget.name} widgets to display correct number (${widget.desiredResult})`,
                true,
                60
            );
        });

        test('should validate correct number of allocated GPUs', async function () {
            videoName = 'number-of-allocated-gpus';
            const widget = widgetTests.find((w) => w.name === EWidgetComponentName.AllocatedGpuWidget);
            if (!widget) {
                throw new Error('Allocated GPU widget test not found');
            }
            await waitUtils.waitForCondition(
                async () => validateStatData(widget),
                `Wait for ${widget.name} widgets to display correct number (${widget.desiredResult})`,
                true,
                60
            );
        });

        test('should filter by node pool', async function () {
            if (!overviewPage.isNodePoolFilterSupported()) {
                logger.info('Node pool filter is not supported in this version, skipping...');
                return;
            }
            await overviewPage.applyNodePoolFilter('default');
            const allocatedGpuText = await overviewPage.getStatWidgetText(
                EWidgetComponentName.AllocatedGpuWidget
            );
            expect(allocatedGpuText).not.toContain('0');
        });

        test('should filter by node pool without workloads', async function () {
            if (!overviewPage.isNodePoolFilterSupported()) {
                logger.info('Node pool filter is not supported in this version, skipping...');
                return;
            }
            await overviewPage.applyNodePoolFilter(nodepools[nodepools.length - 1].nodePoolName);
            const expectedReadyNodesWidget = overviewPage.isReadyNodesFeatureSupported() ? '0/0' : '0';
            const readyNodesText = await overviewPage.getStatWidgetText(
                EWidgetComponentName.ReadyNodesWidget
            );
            expect(readyNodesText).toContain(expectedReadyNodesWidget);

            const expectedReadyGpuWidget = overviewPage.isReadyNodesFeatureSupported() ? '0/0' : '0';
                const readyGpuText = await overviewPage.getStatWidgetText(
                    EWidgetComponentName.ReadyGpuWidget
                );
            expect(readyGpuText).toContain(expectedReadyGpuWidget);

            const allocatedGpuText = await overviewPage.getStatWidgetText(
                EWidgetComponentName.AllocatedGpuWidget
            );
            expect(allocatedGpuText).toContain('0');

            const runningWorkloadsText = await overviewPage.getStatWidgetText(
                EWidgetComponentName.RunningWorkloads
            );
            expect(runningWorkloadsText).toContain('0');

            const pendingWorkloadsText = await overviewPage.getStatWidgetText(
                EWidgetComponentName.PendingWorkloads
            );
            expect(pendingWorkloadsText).toContain('0');

            logger.info('Validated that all widgets are empty filter by node pool without workloads');
            await validateEmptyWidgetsState();
        });
        })
    );
});
