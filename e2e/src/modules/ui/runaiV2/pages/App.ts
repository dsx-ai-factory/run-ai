/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { flowRootState } from '../../../helpers';
import { MENU_SUBJECT_NAMES, MENU_SUB_ITEMS } from '../../utils/sidenavUtils';

export class App extends BaseV2Page {
    settingsDropdown: string;

    settingsDropdownSelectors: string[];

    constructor(page: Page) {
        super(page);
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Admin}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.GeneralSettings}"]`;

        this.settingsDropdown = '[aid="setup-dropdown"]';
        this.settingsDropdownSelectors = [
            '[aid="general-settings-btn"]',
            '[aid="users-btn"]',
            '[aid="applications-btn"]',
            '[aid="access-rules-and-roles-btn"]'
        ];
    }

    async validateNoRolesPageRedirect(): Promise<void> {
        await this.page.waitForNavigation();
        await this.page.waitForSelector('[aid="no-roles"]');
    }

    async redirectToSettings(): Promise<void> {
        await this.page.locator(this.subjectItem).click();
        await this.page.waitForTimeout(300);
        await this.page.locator(this.subMenuItem).click();
    }
}
