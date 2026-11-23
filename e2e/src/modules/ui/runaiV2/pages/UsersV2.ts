/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import * as playwright from 'playwright';
import { logger, waitUtils } from '../../../helpers';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { expect } from '@playwright/test';
import { QTable } from '../components/QTable';
export class UsersV2 extends BaseV2Page {
    toolSettingsIcon: string;

    usersIcon: string;

    creationTime: string;

    resetPassword: string;

    resetButton: string;

    doneButton: string;

    profileIcon: string;

    settings: string;

    emailNotificationsSettingsAll: string;

    emailChannelCheckbox: string;

    changePassword: string;

    currentPassword: string;

    newPassword: string;

    confirmNewPassword: string;

    resetPasswordButton: string;

    addNewUserButton: string;

    userEmail: string;

    userFirstName: string;

    userLastName: string;

    createAndDoneButton: string;

    deleteUserButton: string;

    deleteModalDeleteButton: string;

    accessRulesBtn: string;

    usernameSelector: string;

    createUserButton: string;

    editUserButton: string;

    userDetailRow: string;

    searchIcon: string;

    searchInput: string;

    constructor(page: Page) {
        super(page);
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Access}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Users}"]`;

        this.toolSettingsIcon = '[aid="setup-dropdown"]';
        this.usersIcon = '[aid="users-btn"]';
        this.creationTime = 'text="Creation time"';
        this.resetPassword = 'text="reset password"';
        this.resetButton = '.q-btn.q-btn--actionable.save';
        this.doneButton = '[aid="stepper-navigation-save"]';
        this.profileIcon = '[aid="profile-btn"]';
        this.settings = 'text="Settings"';
        this.emailNotificationsSettingsAll = '[aria-label="All"]';
        this.emailChannelCheckbox = '.q-checkbox__label:has-text("Email")';
        this.changePassword = 'text="Change password"';
        this.currentPassword = 'input[aid="current-password"]';
        this.newPassword = 'input[aid="new-password"]';
        this.confirmNewPassword = 'input[aid="verify-password"]';
        this.resetPasswordButton = '[aid="reset-password-button"]';
        this.addNewUserButton = '[aid="primary-cta-btn"]';
        this.userEmail = '[aid="runai-email-validation"]';
        this.userFirstName = '[aid="user-creation-first-name"]';
        this.userLastName = '[aid="user-creation-last-name"]';
        this.createAndDoneButton = '[aid="stepper-navigation-done"]';
        this.deleteUserButton = '.runai-page-actions button:has-text("DELETE")';
        this.accessRulesBtn = '.runai-page-actions button:has-text("ACCESS RULES")';
        this.deleteModalDeleteButton = '[aid="delete-modal-delete-button"]';
        this.usernameSelector = '[aid="username"]';
        this.createUserButton = '[aid="stepper-navigation-save"]';
        this.editUserButton = '.runai-page-actions button:has-text("EDIT")';
        this.userDetailRow = 'table tbody tr';
        this.searchIcon = '.ph-bold.ph-magnifying-glass';
        this.searchInput = 'input[type="search"]';
    }

    async navigateToUsers(): Promise<void> {
        await this.page.locator(this.subjectItem).click();
        await this.page.waitForTimeout(300);
        await this.page.locator(this.subMenuItem).click();
    }

    async clickOnCreationTime(user: string): Promise<void> {
        let attempts = 0;
        await this.page.waitForTimeout(3000);
        while (attempts < 5) {
            await this.page.locator(this.creationTime).isEnabled({ timeout: 15000 });
            await this.page.locator(this.creationTime).click();
            try {
                await this.page.locator(`text="${user}"`).isEnabled({ timeout: 15000 });
                return;
            } catch {
                logger.error(`User ${user} not found, clicking again...`);
            }
            attempts++;
        }
        logger.error(`User ${user} not found after ${attempts} attempts`);
    }

    async searchByName(name: string): Promise<void> {
        if (!(await this.page.locator(this.searchInput).isVisible({ timeout: 1000 }))) {
            await this.page.locator(this.searchIcon).click();
        }
        await this.page.locator(this.searchInput).fill(name);
        await this.page.waitForTimeout(2500);
    }

    async clickOnUser(username: string): Promise<void> {
        await this.page
            .locator('table [aid="username"]')
            .getByText(username, { exact: true })
            .click({ timeout: 15000 });
    }

    async clickOnResetPassword(): Promise<void> {
        await this.page.locator(this.resetPassword).isEnabled({ timeout: 10000 });
        await this.page.locator(this.resetPassword).click();
    }

    async clickOnResetButton(): Promise<void> {
        await this.page.locator(this.resetButton).isEnabled({ timeout: 10000 });
        await this.page.locator(this.resetButton).click();
    }

    async clickOnDoneButton(): Promise<void> {
        await this.page.locator(this.doneButton).isEnabled({ timeout: 10000 });
        await this.page.locator(this.doneButton).click();
    }

    async returnUserPasswordAfterReset(username: string): Promise<string> {
        let password: string;
        let capturedUserId: string;

        const responseHandler = async (response: playwright.Response): Promise<void> => {
            const url = response.url();
            if (
                url.includes('/users/') &&
                url.includes('/password') &&
                response.request().method() === 'POST' &&
                (response.status() === 200 || response.status() === 201)
            ) {
                try {
                    const responseData = await response.json();
                    // Extract user ID from URL to verify we're resetting the correct user
                    const userIdMatch = url.match(/\/users\/([^/]+)\/password/);
                    if (userIdMatch) {
                        capturedUserId = userIdMatch[1];
                        logger.info(`Captured password reset for user ID: ${capturedUserId}`);
                    }
                    password = responseData;
                } catch (error) {
                    logger.error('Failed to parse response data as JSON:', error);
                }
            }
        };

        this.page.on('response', responseHandler);

        try {
            await this.searchByName(username);
            await this.clickOnUser(username);
            await this.clickOnResetPassword();
            await this.clickOnResetButton();
            await this.clickOnDoneButton();
            // Wait for password reset to propagate in SaaS environments
            await this.page.waitForTimeout(2000);

            if (!password) {
                throw new Error(`Failed to capture password reset response for user: ${username}`);
            }

            logger.info(`Successfully captured password for user: ${username} (ID: ${capturedUserId})`);
            return password;
        } finally {
            this.page.off('response', responseHandler);
        }
    }

    async resetLocalUserPassword(currentPassword: string, newPassword: string): Promise<void> {
        await this.page.locator(this.profileIcon).click();
        await this.page.locator(this.changePassword).click();
        await this.page.locator(this.currentPassword).fill(currentPassword);
        await this.page.locator(this.newPassword).fill(newPassword);
        await this.page.locator(this.confirmNewPassword).fill(newPassword);
        await this.page.locator(this.resetPasswordButton).click();
        await this.page.waitForResponse((response) => response.url().includes('/api/v1/me/password'));
    }

    async resetLocalUserPasswordWithInvalidCurrentPassword(
        currentPassword: string,
        newPassword: string
    ): Promise<void> {
        await this.resetLocalUserPassword(currentPassword, newPassword);
        await this.page.locator(this.currentPassword).fill(currentPassword);
        await this.page.locator(this.newPassword).fill(newPassword);
        await this.page.locator(this.confirmNewPassword).fill(newPassword);
        await this.page.locator(this.resetPasswordButton).click();
    }

    async validateEula(): Promise<boolean> {
        try {
            await this.page.waitForTimeout(5000);
            await this.page.waitForSelector('.eula-contract-content', { timeout: 10000 });
            await this.page.waitForSelector('text="Run:ai - End User License Agreement"', { timeout: 10000 });
            await this.page.waitForTimeout(3000);
            await this.page.waitForSelector('.eula-contract-content embed.contract', { timeout: 10000 });
            const frames = this.page.frames();
            const pdfFrame = frames.length > 1 ? frames[1] : null;
            if (pdfFrame) {
                const errorCodes = ['404', '500', '503', '502', '403', '401', '400'];
                for (const code of errorCodes) {
                    const hasError = await pdfFrame.locator(`text="${code}"`).count();
                    if (hasError > 0) {
                        logger.error(`HTTP ${code} error found in EULA PDF`);
                        return false;
                    }
                }
                const errorMessages = [
                    'Looks like the page you were looking for ran away',
                    'Not Found',
                    'Internal Server Error',
                    'Service Unavailable',
                    'Bad Gateway',
                    'Forbidden',
                    'Unauthorized',
                    'Bad Request'
                ];
                for (const message of errorMessages) {
                    const hasErrorMsg = await pdfFrame.locator(`text="${message}"`).count();
                    if (hasErrorMsg > 0) {
                        logger.error(`EULA PDF error: ${message}`);
                        return false;
                    }
                }
            }
            return true;
        } catch (error) {
            logger.error('Failed to validate EULA:', error.message);
            return false;
        }
    }

    async acceptEula(): Promise<boolean> {
        try {
            await this.page.waitForTimeout(1000);
            await this.page.waitForSelector('[aid="agree-button"]', { state: 'visible', timeout: 10000 });
            await this.page.click('[aid="agree-button"]');
            await this.page.waitForTimeout(1000);
            return true;
        } catch (error) {
            logger.error('Failed to click on accept EULA:', error);
            return false;
        }
    }

    async clickOnSettings(): Promise<void> {
        await this.page.locator(this.profileIcon).click();
        await this.page.locator(this.settings).click();
    }

    async deselectAllNotifications(): Promise<void> {
        await this.expandNotificationsSection();
        await this.uncheckAllNotifications();

        await this.page.locator('text="save"').click();
        await waitUtils.sleep(1000);
    }

    async selectAllNotificationsNoDeleting(): Promise<void> {
        await this.expandNotificationsSection();
        await this.checkAllNotifications();

        await this.page.locator('[aria-label="Deleting"]').click();
        await this.page.locator('text="save"').click();
        await waitUtils.sleep(1000);
    }

    async selectDeletingStatusNotifications(): Promise<void> {
        await this.expandNotificationsSection();
        await this.uncheckAllNotifications();

        await this.page.locator('[aria-label="Deleting"]').click();

        await this.page.locator('text="save"').click();
        await waitUtils.sleep(1000);
    }

    async expandNotificationsSection(): Promise<void> {
        await waitUtils.waitToLoadNewPageOrSaveChanges();
        if (!(await this.page.locator(this.emailNotificationsSettingsAll).first().isVisible())) {
            await this.page
                .locator(
                    '.runai-expansion-header[aria-label*="Email notifications"], .runai-expansion-header[aria-label*="Notifications"]'
                )
                .first()
                .click();
        }
        await this.selectEmailChannelIfAvailable(); // Check if the Email channel checkbox exists (relevant for newer versions- v2.23 and above)
    }

    private async selectEmailChannelIfAvailable(): Promise<void> {
        try {
            const emailCheckbox = this.page.locator(this.emailChannelCheckbox);
            const isEmailCheckboxVisible = await emailCheckbox.isVisible({ timeout: 2000 });

            if (!isEmailCheckboxVisible) {
                logger.info(
                    'Email channel checkbox not found - using older UI version without channel selection'
                );
                return;
            }
            logger.info('Email channel checkbox found - selecting it for newer UI version');
            const isEmailChecked = await emailCheckbox.isChecked();
            if (!isEmailChecked) {
                await emailCheckbox.click();
                await waitUtils.sleep(500);
            }
        } catch {
            logger.info('Email channel checkbox not available - continuing with older UI version');
        }
    }

    private async uncheckAllNotifications(): Promise<void> {
        const checkboxes = this.page.locator(this.emailNotificationsSettingsAll);
        const count = await checkboxes.count();

        for (let i = 0; i < count; i++) {
            const checkbox = checkboxes.nth(i);
            if (await checkbox.isChecked()) {
                await checkbox.click();
                await waitUtils.sleep(200);
            }
        }
    }

    private async checkAllNotifications(): Promise<void> {
        const checkboxes = this.page.locator(this.emailNotificationsSettingsAll);
        const count = await checkboxes.count();

        for (let i = 0; i < count; i++) {
            const checkbox = checkboxes.nth(i);
            if (!(await checkbox.isChecked())) {
                await checkbox.click();
                await waitUtils.sleep(200);
            }
        }
    }

    async validateUserPage(): Promise<boolean> {
        try {
            await this.page.waitForSelector('//table[@class="q-table"]', { timeout: 5000 });
            await this.page.waitForSelector('[aid="createdBy"]', { timeout: 5000 });
            await this.page.waitForSelector('[aid="creationTime"]', { timeout: 5000 });
            const usernames = await this.page.locator('[aid="username"]').all();
            if (usernames.length === 0) {
                logger.error('No username elements found on the page.');
                return false;
            }
            const errorElements = await this.page.locator('.error, .alert').all();
            if (errorElements.length > 0) {
                logger.error('Errors or alerts found on the page.');
                return false;
            }
            return true;
        } catch (error) {
            logger.error('Failed to validate user page:', error);
            return false;
        }
    }

    async clickOnAddNewUserButton(): Promise<void> {
        await this.page.locator(this.addNewUserButton).click();
    }

    async openCreateUserForm(): Promise<void> {
        await this.navigateToUsers();
        await this.clickOnAddNewUserButton();
        await this.page.locator(this.userFirstName).waitFor({ state: 'visible', timeout: 10000 });
        await this.page.locator(this.userLastName).waitFor({ state: 'visible', timeout: 10000 });
    }

    async submitCreateUserForm(email: string, firstName?: string, lastName?: string): Promise<void> {
        await this.page.locator(this.userEmail).fill(email);
        if (firstName) {
            await this.page.locator(this.userFirstName).fill(firstName);
        }
        if (lastName) {
            await this.page.locator(this.userLastName).fill(lastName);
        }
        await this.page.locator(this.createUserButton).click({ timeout: 10000 });
        await this.page.locator(this.createAndDoneButton).click({ timeout: 10000 });
    }

    async createUser(email: string, firstName?: string, lastName?: string): Promise<void> {
        await this.clickOnAddNewUserButton();
        await this.submitCreateUserForm(email, firstName, lastName);
    }

    async createLocalUser(email: string, firstName?: string, lastName?: string): Promise<void> {
        await this.navigateToUsers();
        await this.createUser(email, firstName, lastName);
    }

    async verifyUserTableFullName(userEmail: string, expectedFullName: string): Promise<void> {
        await this.navigateToUsers();
        await this.searchByName(userEmail);
        const qTable = new QTable(this.page);
        const row = await qTable.getEntityData('User', userEmail);
        expect(row).not.toBeNull();
        expect(row!['Full name']).toBe(expectedFullName);
        expect(row!['Full name']).not.toMatch(/undefined|null/);
    }

    private async openLocalUserEditForm(userEmail: string): Promise<playwright.Locator> {
        await this.navigateToUsers();
        await this.searchByName(userEmail);
        await this.clickOnUser(userEmail);
        await this.page.locator(this.editUserButton).click({ timeout: 5000 });
        const editUserDialog = this.page.getByRole('dialog').filter({ hasText: 'Edit user' });
        await editUserDialog.getByRole('textbox', { name: 'First name' }).waitFor({
            state: 'visible',
            timeout: 10000
        });
        await editUserDialog.getByRole('textbox', { name: 'Last name' }).waitFor({
            state: 'visible',
            timeout: 10000
        });
        return editUserDialog;
    }

    async updateLocalUserFullName(
        userEmail: string,
        updatedFirstName: string,
        updatedLastName: string
    ): Promise<void> {
        const editUserDialog = await this.openLocalUserEditForm(userEmail);
        await editUserDialog.getByRole('textbox', { name: 'First name' }).fill(updatedFirstName);
        await editUserDialog.getByRole('textbox', { name: 'Last name' }).fill(updatedLastName);
        await editUserDialog.getByRole('button', { name: 'Save' }).click({ timeout: 10000 });
        await editUserDialog.waitFor({ state: 'hidden', timeout: 10000 });
        await this.page.waitForTimeout(1500);
        const closeSelectionButton = this.page.getByRole('button', { name: 'Close selection' });
        if (await closeSelectionButton.isVisible({ timeout: 1000 }).catch(() => false)) {
            await closeSelectionButton.click();
        }
    }

    async verifyUserDetailDoesNotContainInvalidName(
        userEmail: string,
        expectedFullName?: string
    ): Promise<void> {
        await this.navigateToUsers();
        await this.searchByName(userEmail);
        await this.clickOnUser(userEmail);
        const userDetailSurface = this.page
            .locator(this.userDetailRow)
            .filter({ hasText: userEmail })
            .first();
        await expect(userDetailSurface).toBeVisible({ timeout: 5000 });
        await expect(userDetailSurface).toContainText(userEmail);
        if (expectedFullName && expectedFullName !== '-') {
            await expect(userDetailSurface).toContainText(expectedFullName);
        }
        await expect(userDetailSurface).not.toContainText(/undefined|null/);
    }

    async verifyExternalUserEditActionNotAvailable(userEmail: string): Promise<void> {
        await this.navigateToUsers();
        await this.searchByName(userEmail);
        await this.clickOnUser(userEmail);

        await expect(this.page.locator(this.accessRulesBtn)).toBeVisible({ timeout: 5000 });
        await expect(this.page.locator(this.editUserButton)).toBeHidden({ timeout: 5000 });
    }

    async clickOnDeleteUserButton(): Promise<void> {
        await this.page.locator(this.deleteUserButton).click({ timeout: 10000 });
        await this.page.locator(this.deleteModalDeleteButton).click({ timeout: 10000 });
    }

    async clickOnAccessRulesBtn(): Promise<void> {
        await this.page.locator(this.accessRulesBtn).click();
    }

    async getUsersCount(): Promise<number> {
        const userCount = await this.page.locator(this.usernameSelector).count();
        return userCount;
    }

    async verifyUserNotLoggedOut(): Promise<boolean> {
        try {
            await this.page.waitForSelector(this.profileIcon, { timeout: 10000 });
            return true;
        } catch (error) {
            logger.error('User logged out:', error);
            return false;
        }
    }
}
