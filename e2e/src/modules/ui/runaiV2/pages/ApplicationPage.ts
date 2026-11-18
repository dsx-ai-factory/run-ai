/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { logger } from '../../../helpers';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';

export class ApplicationsPage extends BaseV2Page {
    toolSettingsIcon: string;

    applicationIcon: string;

    generalSearch: string;

    generalSearchInput: string;

    newApplicationBtn: string;

    newServiceAccountBtn: string;

    serviceAccountNameInput: string;

    createButton: string;

    doneButton: string;

    deleteButton: string;

    deleteConfirmButton: string;

    regenerateSecretButton: string;

    newSecretValidation: string;

    constructor(page: Page) {
        super(page);
        this.toolSettingsIcon = '[aid="setup-dropdown"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Access}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.ServiceAccounts}"]`;
        this.applicationIcon = '[aid="applications-btn"]';
        this.generalSearch = '.runai-general-search';
        this.generalSearchInput = '.search-input';
        this.newApplicationBtn = '.runai-page-actions button[aid="primary-cta-btn"]';
        this.newServiceAccountBtn = '[aid="primary-cta-btn"]';
        this.serviceAccountNameInput = '[aid="runai-name-validation"]';
        this.createButton = '[aid="next-btn"]';
        this.doneButton = 'text=done';
        this.deleteButton = '[aid="delete-service-account-btn"]';
        this.deleteConfirmButton = '[aid="delete-modal-delete-button"]';
        this.regenerateSecretButton = '[aid="regenerate-service-account-btn"]';
        this.newSecretValidation = 'text=New client secret: ';
    }

    async navigateToServiceAccountsPage(): Promise<void> {
        await this.page.locator(this.subjectItem).click();
        await this.page.waitForTimeout(300);
        await this.page.locator(this.subMenuItem).click();
    }

    async createServiceAccount(serviceAccountName: string): Promise<void> {
        await this.page.locator(this.newServiceAccountBtn).click();
        await this.page.locator(this.serviceAccountNameInput).fill(serviceAccountName);
        await this.page.waitForTimeout(2000);
        await this.page.locator(this.createButton).click();
        await this.page.waitForTimeout(2000);
        await this.page.locator(this.doneButton).click();
        await this.page.waitForTimeout(2000);
    }

    async validateServiceAccountExist(serviceAccountName: string): Promise<boolean | string> {
        await this.navigateToServiceAccountsPage();
        await this.page.waitForTimeout(2000);
        await this.page.waitForSelector(this.generalSearch);
        await this.page.locator(this.generalSearch).click();
        await this.page.locator(this.generalSearchInput).fill(serviceAccountName);
        const appNameLocator = this.page.locator(`[aid="serviceAccount"]:has-text("${serviceAccountName}")`);
        const appName = await appNameLocator.textContent();
        logger.debug(`UI Service Account name: ${serviceAccountName}`);
        return serviceAccountName;
    }

    async deleteServiceAccount(serviceAccountName: string): Promise<void> {
        await this.navigateToServiceAccountsPage();
        await this.page.waitForTimeout(2000);
        await this.page.locator(this.generalSearch).click();
        await this.page.locator(this.generalSearchInput).fill(serviceAccountName);
        const serviceAccountLocator = this.page.locator(
            `[aid="serviceAccount"]:has-text("${serviceAccountName}")`
        );
        await serviceAccountLocator.click();
        await this.page.locator(this.deleteButton).click();
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.deleteConfirmButton).click();
        await this.page.waitForTimeout(1000);
    }

    async regenerateServiceAccountSecret(serviceAccountName: string): Promise<void> {
        await this.navigateToServiceAccountsPage();
        await this.page.waitForTimeout(2000);
        await this.page.locator(this.generalSearch).click();
        await this.page.locator(this.generalSearchInput).fill(serviceAccountName);
        const serviceAccountLocator = this.page.locator(
            `[aid="serviceAccount"]:has-text("${serviceAccountName}")`
        );
        await serviceAccountLocator.click();
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.regenerateSecretButton).click();
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.createButton).click();
        await this.page.waitForTimeout(1000);
        await this.validateNewServiceAccountSecret();
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.createButton).click();
        await this.page.waitForTimeout(1000);
    }

    async validateNewServiceAccountSecret(): Promise<void> {
        await this.page.waitForSelector(this.newSecretValidation, {
            timeout: 10000
        });
    }

    async filterByServiceAccount(serviceAccountName: string): Promise<string> {
        await this.page.waitForTimeout(2000);
        const addFilterBtn = this.page.locator('.add-filter-btn');
        await addFilterBtn.waitFor({ state: 'visible', timeout: 5000 });
        await addFilterBtn.click();
        await this.page.waitForTimeout(1000);
        const serviceAccountOption = this.page.locator(
            '.q-item__section.column.q-item__section--main:has-text("Service account")'
        );
        await serviceAccountOption.waitFor({ state: 'visible', timeout: 5000 });
        await serviceAccountOption.click();
        await this.page.waitForTimeout(1000);
        const filterInput = this.page.locator('input[aria-label="Contains"]');
        await filterInput.waitFor({ state: 'visible', timeout: 5000 });
        await filterInput.fill(serviceAccountName);
        await filterInput.press('Enter');
        await this.page.waitForTimeout(500);
        const applyBtn = this.page.locator(this.applyFilterBtn);
        await applyBtn.waitFor({ state: 'visible', timeout: 5000 });
        await applyBtn.click();
        await this.page.waitForTimeout(2000);
        const filterChip = this.page.locator('[aid="filter-chip-service-account"]');
        await filterChip.waitFor({ state: 'visible', timeout: 10000 });
        return await filterChip.textContent();
    }

    async verifyServiceAccountFilterChip(...expectedTexts: string[]): Promise<boolean> {
        const filterChip = this.page.locator('[aid="filter-chip-service-account"]');
        await filterChip.waitFor({ state: 'visible', timeout: 10000 });
        const chipText = await filterChip.textContent();
        return expectedTexts.every((text) => chipText?.includes(text));
    }
}
