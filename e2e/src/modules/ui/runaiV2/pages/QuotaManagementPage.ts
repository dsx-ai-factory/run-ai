/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page, Locator } from 'playwright';
import { stringUtil } from '../../utils/stringUtils';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';

// Mirror values of EWidgetComponent in
// control-plane/frontend/frontend-v2/src/models/chart.model.ts; aid is the
// kebab-case of the value.
export enum EQuotaWidgetComponentName {
    QuotaTotalGpuWidget = 'QuotaTotalGpuWidget',
    AllocatedQuotaGpuWidget = 'AllocatedQuotaGpuWidget',
    PendingWorkloadsWidget = 'PendingWorkloadsWidget',
    QuotaByNodepoolWidget = 'QuotaByNodepoolWidget',
    AllocationByNodepoolWidget = 'AllocationByNodepoolWidget',
    PendingWorkloadsByNodepoolWidget = 'PendingWorkloadsByNodepoolWidget',
    QuotaAllocationTimeRangeWidget = 'QuotaAllocationTimeRangeWidget',
    DeptAllocationRatioWidget = 'DeptAllocationRatioWidget',
    ProjectAllocationRatioWidget = 'ProjectAllocationRatioWidget'
}

export class QuotaManagementPage extends BaseV2Page {
    constructor(page: Page) {
        super(page);
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Analytics}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Resources}"]`;
    }

    getWidgetSelector(widgetName: EQuotaWidgetComponentName): string {
        return `[aid="${stringUtil.camelCaseToHyphenated(widgetName)}"]`;
    }

    async getWidgetXLabel(widgetName: EQuotaWidgetComponentName): Promise<string> {
        const labels = await this.page
            .locator(this.getWidgetCategoryLabelLocator(widgetName))
            .allTextContents();
        return labels.join(' ');
    }

    getWidgetXLabelLocator(widgetName: EQuotaWidgetComponentName): string {
        return `${this.getWidgetSelector(widgetName)} [aria-label="Explore chart data"]`;
    }

    // Drillable categories are buttons; terminal categories remain as ECharts SVG labels.
    getWidgetCategoryLabelLocator(widgetName: EQuotaWidgetComponentName): string {
        const widgetSelector = this.getWidgetSelector(widgetName);
        return `${widgetSelector} [aria-label="Explore chart data"], ${widgetSelector} svg text`;
    }

    // Match either label form; the text filter keeps chart titles and axis ticks out.
    drillableCategoryLocator(widgetName: EQuotaWidgetComponentName, text: string): Locator {
        const widgetSelector = this.getWidgetSelector(widgetName);
        return this.page
            .locator(
                `${widgetSelector} [aria-label="Explore chart data"] button, ` +
                    `${widgetSelector} svg text`
            )
            .filter({ hasText: text });
    }

    async isDrillableCategoryVisible(widgetName: EQuotaWidgetComponentName, text: string): Promise<boolean> {
        return (await this.drillableCategoryLocator(widgetName, text).count()) > 0;
    }

    async clickWidgetXLabel(widgetName: EQuotaWidgetComponentName, text: string): Promise<void> {
        return await this.clickWidgetXLabelWithText(widgetName, text);
    }

    async clickWidgetXLabelWithText(widgetName: EQuotaWidgetComponentName, text: string): Promise<void> {
        // SVG labels are not hit-testable like buttons.
        return await this.drillableCategoryLocator(widgetName, text).first().click({ force: true });
    }
}
