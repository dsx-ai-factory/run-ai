/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { waitUtils } from '../../../helpers';
import { QTable } from '../components/QTable';

export class NodePage extends BaseV2Page {
    nodeNameCol: string;

    showDetailsBtn: string;

    sidePageDiv: string;

    detailsCpuUtilDiv: string;

    detailsCpuMemUsageDiv: string;

    detailsGpuUtilDiv: string;

    detailsGpuMemUtilDiv: string;

    detailsCloseBtn: string;

    qTable: QTable;

    constructor(page: Page) {
        super(page);
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Resources}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Nodes}"]`;

        this.nodeNameCol = '[aid="name"]';
        this.showDetailsBtn = '[aid="show-details-button"]';
        this.sidePageDiv = 'div.drawer-content';
        this.detailsCpuUtilDiv = '[aid="cpu-utilization"]';
        this.detailsCpuMemUsageDiv = '[aid="cpu-memory-usage"]';
        this.detailsGpuUtilDiv = '[aid="gpu-utilization"]';
        this.detailsGpuMemUtilDiv = '[aid="gpu-memory"]';

        this.detailsCloseBtn = 'button.close-button';

        this.qTable = new QTable(page);
    }

    async toggleNode(nodeName: string): Promise<void> {
        const visible = await this.isElementVisible(`td${this.nodeNameCol}:has-text("${nodeName}")`);
        if (!visible) await this.searchByName(nodeName);
        await this.clickOnElementByText(`td${this.nodeNameCol}`, `${nodeName}`);
    }

    async isElementVisible(query: string): Promise<boolean> {
        return await this.page.locator(query).isVisible();
    }

    async clickShowDetailsBtn(): Promise<void> {
        await this.page.locator(this.showDetailsBtn).click();
        await waitUtils.waitForElementToExist(this, this.sidePageDiv);
    }

    async closeDetailsSidePage(): Promise<void> {
        await this.page.locator(this.detailsCloseBtn).click();
        await waitUtils.waitForElementToNotExist(this, this.sidePageDiv);
    }

    async waitForNodeDetailsChartsToExist(gpuNode = false): Promise<void> {
        const chartLocators = [this.detailsCpuUtilDiv, this.detailsCpuMemUsageDiv];
        if (gpuNode) {
            chartLocators.push(this.detailsGpuUtilDiv);
            chartLocators.push(this.detailsGpuMemUtilDiv);
        }

        for (const chartLocator of chartLocators) {
            await waitUtils.waitForElementToExist(this, chartLocator);
        }
    }

    async getAllNodesData(): Promise<Array<Record<string, string>>> {
        return await this.qTable.getTableData();
    }
}
