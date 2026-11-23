/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';

export class EventHistoryPage extends BaseV2Page {
    constructor(page: Page) {
        super(page);
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Admin}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Audit}"]`;
    }

    async filterByStatus(status: 'Succeeded' | 'Failed'): Promise<string> {
        await this.page.waitForTimeout(2000);
        const addFilterBtn = this.page.locator('.add-filter-btn');
        await addFilterBtn.waitFor({ state: 'visible', timeout: 5000 });
        await addFilterBtn.click();
        await this.page.waitForTimeout(1000);
        const statusOption = this.page.locator('.add-filter-item:has-text("Status")');
        await statusOption.waitFor({ state: 'visible', timeout: 5000 });
        await statusOption.click();
        await this.page.waitForTimeout(1500);
        await this.page.locator('text=Matches any:').waitFor({ state: 'visible', timeout: 5000 });
        const statusCheckbox = this.page.locator(`.q-checkbox:has-text("${status}") .q-checkbox__inner`);
        await statusCheckbox.waitFor({ state: 'visible', timeout: 5000 });
        await statusCheckbox.click();
        await this.page.waitForTimeout(1000);
        const applyBtn = this.page.locator('button:has-text("Apply")').first();
        await applyBtn.waitFor({ state: 'attached', timeout: 5000 });
        await applyBtn.click({ force: true });
        await this.page.waitForTimeout(2000);
        const filterChip = this.page.locator('.filter-chip');
        await filterChip.waitFor({ state: 'visible', timeout: 10000 });
        return await filterChip.textContent();
    }

    async verifyFilterChip(...expectedTexts: string[]): Promise<boolean> {
        const filterChip = this.page.locator('.filter-chip');
        await filterChip.waitFor({ state: 'visible', timeout: 10000 });
        const chipText = await filterChip.textContent();
        return expectedTexts.every((text) => chipText?.includes(text));
    }
}
