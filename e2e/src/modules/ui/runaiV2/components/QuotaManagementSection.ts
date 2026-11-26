/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { BaseV2Page } from '../pages/BaseV2Page';
import { flowRootState, versionUtils } from '../../../helpers';
import { SupportedVersions } from '../../../helpers/utils/SupporetedEnvsUtils';

export class QuotaManagementSection extends BaseV2Page {
    gpuQuotaButton: string;

    gpuQuotaInput: string;

    quotaApplyButton: string;

    schedulingTab: string;

    prioritySelect: string;

    overQuotaSelect: string;

    maxGpuLimitButton: string;

    maxGpuLimitInput: string;

    gpuQuotaExceedsAlertDiv: string;

    constructor(page: Page) {
        super(page);

        this.gpuQuotaButton = 'gpu-quota-button';
        this.gpuQuotaInput = '[aid="gpu-quota-input"]';
        this.schedulingTab = '[aid="scheduling-tab"]';
        this.prioritySelect = '[aid="priority-select"]';
        this.overQuotaSelect = 'over-quota-select';
        this.maxGpuLimitButton = 'limit-button';
        this.maxGpuLimitInput = '[aid="limit-quota-input"]';
        this.quotaApplyButton = '[aid="save-quota-btn"]';
        this.gpuQuotaExceedsAlertDiv = 'div.quota-exceeds-alert.text-negative';
    }

    async setGpuQuota(gpuQuota: string, nodePool = 'default'): Promise<void> {
        await this.page.locator(`[data-testid="${this.gpuQuotaButton}-${nodePool}"]`).click();
        await this.page.locator(this.gpuQuotaInput).fill(gpuQuota);
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.quotaApplyButton).click();
    }

    async switchToSchedulingTab(): Promise<void> {
        await this.page.locator(this.schedulingTab).click();
    }

    async selectPorioty(priority: string): Promise<void> {
        await this.page.locator(this.prioritySelect).click();
        await this.page.locator(`div[role="option"]:has-text("${priority}")`).click();
    }

    async enableOverQuota(nodePool = 'default'): Promise<void> {
        await this.page.locator(`[data-testid="${this.overQuotaSelect}-${nodePool}"]`).click();
        await this.page.locator('div.q-item__label:has-text("Enabled")').click();
    }

    async disableOverQuota(nodePool = 'default'): Promise<void> {
        await this.page.locator(`[data-testid="${this.overQuotaSelect}-${nodePool}"]`).click();
        await this.page.locator('div[.q-item__label:has-text("Disabled")').click();
    }

    async setMaxGpuLimit(maxGpu: string, nodePool = 'default'): Promise<void> {
        await this.page.locator(`[aid="${this.maxGpuLimitButton}-${nodePool}"]`).click();
        await this.page.locator(this.maxGpuLimitInput).fill(maxGpu);
        await this.page.waitForTimeout(300);
        await this.page.locator(this.quotaApplyButton).click();
    }

    async setQuota(
        gpuQuota: string,
        priority?: string,
        overQuota?: boolean,
        maxGpuLimit?: string,
        nodePool: string = 'default'
    ): Promise<void> {
        await this.setGpuQuota(gpuQuota, nodePool);

        if (priority || overQuota !== undefined || maxGpuLimit) {
            if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_20)) {
                this.switchToSchedulingTab();
            }

            if (priority) {
                if (nodePool === 'default') {
                    await this.selectPorioty(priority);
                } else {
                    await this.selectPriorityForNodePool(priority, nodePool);
                }
            }

            if (overQuota !== undefined) {
                if (overQuota) {
                    await this.enableOverQuota(nodePool);
                } else {
                    await this.disableOverQuota(nodePool);
                }
            }

            if (maxGpuLimit) await this.setMaxGpuLimit(maxGpuLimit, nodePool);
        }
    }

    async selectPriorityForNodePool(priority: string, nodePool: string): Promise<void> {
        await this.page.locator(`[aid="priority-select-${nodePool}"]`).click();
        await this.page.locator(`div[role="option"]:has-text("${priority}")`).click();
    }

    async getGpuQuotaExcceedsAlertContent(): Promise<string> {
        return this.page.locator(this.gpuQuotaExceedsAlertDiv).textContent();
    }
}
