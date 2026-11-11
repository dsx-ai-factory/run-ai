/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags
} from '../../../../../../../modules/helpers';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    EQuotaWidgetComponentName,
    QuotaManagementPage
} from '../../../../../../../modules/ui/runaiV2/pages/QuotaManagementPage';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';

// runOn.serialCI is intentionally not added yet — re-add once the suite has
// proven stable in scheduled regression runs (per code review feedback).
test.describe(
    `Quota Management UI, tags: ${tags.tool.ui} ${tags.resource.overview} ${tags.teams.viper}`,
    { tag: [tags.services.orgUnitService] },
    () => {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let quotaManagementPage: QuotaManagementPage;

        // DeptAllocationRatioWidget is conditional on a non-default department —
        // covered by the multi-dept test in departmentsEnabledByDefault.test.ts.
        const widgetsAlwaysVisible = [
            EQuotaWidgetComponentName.QuotaTotalGpuWidget,
            EQuotaWidgetComponentName.AllocatedQuotaGpuWidget,
            EQuotaWidgetComponentName.PendingWorkloadsWidget,
            EQuotaWidgetComponentName.QuotaByNodepoolWidget,
            EQuotaWidgetComponentName.AllocationByNodepoolWidget,
            EQuotaWidgetComponentName.PendingWorkloadsByNodepoolWidget,
            EQuotaWidgetComponentName.QuotaAllocationTimeRangeWidget,
            EQuotaWidgetComponentName.ProjectAllocationRatioWidget
        ];

        const archiveName = (testInfo: { title: string }): string =>
            'quota-management-' + testInfo.title.replace(/\W+/g, '-').toLowerCase();

        const readQuotaTotalLeadingValue = async (): Promise<number> => {
            const statData = quotaManagementPage.page.locator(
                `${quotaManagementPage.getWidgetSelector(EQuotaWidgetComponentName.QuotaTotalGpuWidget)} .stat-data`
            );
            const text = (await statData.textContent()) ?? '';
            return parseInt(text.match(/^(\d+)/)?.[1] ?? '0', 10);
        };

        test.beforeAll(() => {
            browserEngine = new BrowserEngine();
        });

        const startSession = async (): Promise<void> => {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            quotaManagementPage = new QuotaManagementPage(browserEngine.currentPage);
            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
            await quotaManagementPage.goToPage();
        };

        test.beforeEach(async () => {
            await startSession();
        });

        test.afterEach(async ({}, testInfo) => {
            try {
                await browserEngine.stopAndArchive(testInfo, archiveName(testInfo));
            } catch (e) {
                logger.error(e);
                throw e;
            }
        });

        test('should redirect to quota management page', async () => {
            const title = await quotaManagementPage.getPageTitle();
            expect(title).toContain('Quota management');
        });

        test('should verify widgets presence', async () => {
            for (const widget of widgetsAlwaysVisible) {
                const selector = quotaManagementPage.getWidgetSelector(widget);
                await quotaManagementPage.page.waitForSelector(selector);
                await expect(quotaManagementPage.page.locator(selector)).toBeVisible();
            }
        });

        test('should render numeric values in stat widgets', async () => {
            for (const widget of [
                EQuotaWidgetComponentName.QuotaTotalGpuWidget,
                EQuotaWidgetComponentName.AllocatedQuotaGpuWidget,
                EQuotaWidgetComponentName.PendingWorkloadsWidget
            ]) {
                const statData = quotaManagementPage.page.locator(
                    `${quotaManagementPage.getWidgetSelector(widget)} .stat-data`
                );
                await expect(statData).toBeVisible();
                await expect(statData).toHaveText(/\d/);
            }
        });

        test('should render time-range widget with date picker and chart', async () => {
            const widgetSel = quotaManagementPage.getWidgetSelector(
                EQuotaWidgetComponentName.QuotaAllocationTimeRangeWidget
            );
            await expect(
                quotaManagementPage.page.locator(`${widgetSel} [aid="quota-allocation-date-picker"]`)
            ).toBeVisible();
            await expect(
                quotaManagementPage.page.locator(`${widgetSel} #quota-allocation-time-range-widget-chart-id`)
            ).toBeVisible();
        });

        test('should expose the page-level filter bar', async () => {
            await expect(quotaManagementPage.page.locator('[aid="cluster-filter-chip"]')).toBeVisible();
        });

        test('should reflect a created project deserved GPUs in the quota stat widget', async () => {
            const deservedGpus = 4;
            const statData = quotaManagementPage.page.locator(
                `${quotaManagementPage.getWidgetSelector(EQuotaWidgetComponentName.QuotaTotalGpuWidget)} .stat-data`
            );
            // Wait for the widget to render numeric content before sampling
            // baseline so a race-read of an empty widget can't masquerade as 0.
            await expect(statData).toHaveText(/\d/);
            const baseline = await readQuotaTotalLeadingValue();
            const projectName = objectGenerator.generateShortNameByPrefix('p-qm');
            let project: Project | undefined;
            try {
                project = await orgUnitUtils.createProjectAndValidate({
                    name: projectName,
                    resources: [{ gpu: { deserved: deservedGpus } }],
                    clusterId: flowRootState.cluster.uuid
                });

                await quotaManagementPage.goToPage();

                // Poll up to 90s for metrics aggregation to surface the new project.
                await expect(async () => {
                    expect(await readQuotaTotalLeadingValue()).toBeGreaterThanOrEqual(
                        baseline + deservedGpus
                    );
                }).toPass({ timeout: 90_000, intervals: [2_000, 5_000, 10_000] });
            } finally {
                if (project) {
                    try {
                        await orgUnitUtils.deleteProject(project.id);
                    } catch (e) {
                        logger.error(`Failed to delete project ${project.id}: ${e}`);
                    }
                }
            }
        });
    }
);
