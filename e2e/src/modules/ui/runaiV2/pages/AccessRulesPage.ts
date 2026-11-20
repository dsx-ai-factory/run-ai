/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { logger } from '../../../helpers';
import { Scope } from '../../../api/apis/assetsApi';
import { tenantName } from '../../../helpers/config/const';
import { expect } from '@playwright/test';

export class AccessRulesPage extends BaseV2Page {
    // Navigation elements
    accessSubjectItem: string;

    accessRulesMenuItem: string;

    // Page elements
    primaryCtaBtn: string;

    pageTitle: string;

    // Access rule creation modal elements
    subjectSelect: string;

    addableMultiSelect: string;

    roleSelect: string;

    scopeInput: string;

    scopeBtn: string;

    saveRuleBtn: string;

    cancelBtn: string;

    // Table elements
    tableRows: string;

    tableRowsAlt: string;

    rowCheckbox: string;

    rowCheckboxAlt: string;

    // Bulk operations elements
    selectedRowsActions: string;

    deleteSelectedBtn: string;

    confirmDeleteBtn: string;

    // Search and filter elements
    searchInput: string;

    searchInputAlt: string;

    searchToggleButton: string;

    searchToggleIcon: string;

    // Notification elements
    successNotification: string;

    errorNotification: string;

    // Modal elements
    modalHeader: string;

    modalDeleteBtn: string;

    // Dropdown options
    userOption: string;

    groupOption: string;

    dropdownMenu: string;

    dropdownOption: string;

    ssoGroupOptionName: RegExp;

    roleOption: string;

    scopeOption: string;

    newAccessRulesBtn: string;

    constructor(page: Page) {
        super(page);

        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Access}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.AccessRules}"]`;
        this.newAccessRulesBtn = '.runai-page-actions button[aid="primary-cta-btn"]';

        // Navigation
        this.accessSubjectItem = `[aid="${MENU_SUBJECT_NAMES.Access}"]`;
        this.accessRulesMenuItem = `[aid="${MENU_SUB_ITEMS.AccessRules}"]`;
        this.subjectItem = this.accessSubjectItem;
        this.subMenuItem = this.accessRulesMenuItem;

        // Page elements
        this.primaryCtaBtn = '[aid="primary-cta-btn"]';
        this.pageTitle = '[aid="page-title"]';

        // Access rule creation modal elements
        this.subjectSelect = '[aid="subject-select"]';
        this.addableMultiSelect = '[aid="addable-multi-select"]';
        this.roleSelect = 'input[aria-label="Role"], .q-select__focus-target[aria-label="Role"]';
        this.scopeInput = '[aid="scope-input"], .runai-multi-scope-input input';
        this.scopeBtn = '.scope-btn';
        this.saveRuleBtn =
            '[aid="save-rule-btn"], button:has-text("SAVE RULE"), button:has-text("save rule")';
        this.cancelBtn = '[aid="cancel-btn"], button:has-text("Cancel"), button:has-text("CANCEL")';

        // Table elements
        this.tableRows = 'table tbody tr';
        this.tableRowsAlt = '[role="row"]';
        this.rowCheckbox = 'input[type="checkbox"]';
        this.rowCheckboxAlt = '[role="checkbox"]';

        // Bulk operations elements
        this.selectedRowsActions = '.selected-rows-actions, [aid="selected-rows-actions"]';
        this.deleteSelectedBtn =
            '[aid="delete-selected-btn"], .selected-rows-actions button:has-text("Delete")';
        this.confirmDeleteBtn =
            '[aid="confirm-delete-btn"], button:has-text("DELETE"), button:has-text("delete")';

        // Search and filter elements
        this.searchInput = 'input[placeholder*="Search"], input[type="search"], .search-input';
        this.searchInputAlt = 'input[aria-label*="search"], input[name*="search"]';
        this.searchToggleButton = 'button:has-text("SEARCH")';
        this.searchToggleIcon = '.ph-bold.ph-magnifying-glass';
        this.clearSearchBtn = 'i.search-clear, .search-clear';

        // Notification elements
        this.successNotification =
            '.q-notification:has-text("success"), .notification:has-text("created"), .alert:has-text("created"), .notification:has-text("deleted"), .alert:has-text("deleted")';
        this.errorNotification =
            '.q-notification:has-text("error"), .notification:has-text("error"), .alert:has-text("error")';

        // Modal elements
        this.modalDialog = '.q-card.dialog-content-container, .modal-dialog, .dialog-content-container';

        this.modalDeleteBtn = '[aid="delete-modal-delete-button"], .modal-footer button:has-text("Delete")';

        // Dropdown options
        this.userOption = '.q-menu .q-item:has-text("User"), [role="option"]:has-text("User")';
        this.groupOption = '.q-menu .q-item:has-text("Group"), [role="option"]:has-text("Group")';
        this.dropdownMenu = '.q-menu';
        this.dropdownOption = '[role="option"]';
        this.ssoGroupOptionName = /^(SSO )?Group$/;
        this.roleOption = '.q-menu .q-item, [role="option"]';
        this.scopeOption = '.q-menu .q-item, [role="option"]';
    }

    /**
     * Navigate to the access rules page
     */
    async goToPage(): Promise<void> {
        logger.info('Navigating to access rules page');
        await this.selectCurrentCluster();
        await this.page.locator(this.accessSubjectItem).click();
        await this.page.waitForTimeout(500);
        await this.page.locator(this.accessRulesMenuItem).click();
        await this.page.waitForTimeout(2000);
    }

    /**
     * Check if we're on the access rules page
     */
    async isOnAccessRulesPage(): Promise<boolean> {
        try {
            await this.page.waitForSelector(this.primaryCtaBtn, { timeout: 5000 });
            return true;
        } catch {
            return false;
        }
    }

    /**
     * Click the primary CTA button (New Access Rule)
     */
    async clickNewAccessRule(): Promise<void> {
        logger.info('Clicking new access rule button');
        await this.page.locator(this.primaryCtaBtn).click();
        await this.page.waitForTimeout(1000);
    }

    /**
     * Select subject type (User, Group, or Service account)
     */
    async selectSubjectType(subjectType: 'User' | 'Group' | 'Service account'): Promise<void> {
        logger.info(`Selecting subject type: ${subjectType}`);

        // First click on the dropdown to open it
        await this.page.locator(this.subjectSelect).click();
        await this.page.waitForTimeout(500);

        // Wait for dropdown menu to appear
        await this.page.waitForSelector(this.dropdownMenu, { state: 'visible' });

        const subjectTypeMenu = this.page.locator(this.dropdownMenu).last();
        const subjectTypeOption =
            subjectType === 'Group'
                ? subjectTypeMenu.locator(this.dropdownOption).getByText(this.ssoGroupOptionName)
                : subjectTypeMenu.locator(this.dropdownOption).getByText(subjectType, { exact: true });
        await subjectTypeOption.click();
        await this.page.waitForTimeout(500);
    }

    /**
     * Add a subject (user or group) to the access rule
     */
    async addSubject(subjectName: string): Promise<void> {
        logger.info(`Adding subject: ${subjectName}`);

        const subjectInput = this.page.locator(this.addableMultiSelect);
        await subjectInput.click();
        await subjectInput.fill(subjectName);
        await this.page.waitForTimeout(500);

        // Try to select from dropdown
        const userOption = this.page.locator(`.q-menu .q-item:has-text("${subjectName}")`);
        if (await userOption.isVisible({ timeout: 2000 })) {
            await userOption.click();
        } else {
            // Try alternative selector
            const altUserOption = this.page.locator(`[role="option"]:has-text("${subjectName}")`);
            if (await altUserOption.isVisible({ timeout: 1000 })) {
                await altUserOption.click();
            } else {
                // Press Enter if dropdown option not found
                await subjectInput.press('Enter');
            }
        }
        await this.page.waitForTimeout(300);
    }

    /**
     * Add multiple subjects to the access rule
     */
    async addMultipleSubjects(subjects: string[]): Promise<void> {
        logger.info(`Adding multiple subjects: ${subjects.join(', ')}`);

        for (const subject of subjects) {
            await this.addSubject(subject);
        }
    }

    /**
     * Select a role for the access rule
     */
    async selectRole(roleName: string): Promise<void> {
        logger.info(`Selecting role: ${roleName}`);

        try {
            // First click on the role dropdown to open it
            await this.page.locator(this.roleSelect).click({ timeout: 10000 });
            await this.page.waitForTimeout(500);

            // Wait for dropdown menu to appear
            await this.page.waitForSelector('.q-menu', { state: 'visible', timeout: 10000 });

            // Then click on the specific role option from the dropdown (exact match to avoid ambiguity)
            await this.page
                .locator('.q-menu [role="option"]')
                .getByText(roleName, { exact: true })
                .click({ timeout: 5000 });
            await this.page.waitForTimeout(500);
        } catch (error) {
            logger.error(`Failed to select role ${roleName}: ${error}`);
            // Try alternative selector if primary fails - click the container
            await this.page
                .locator('.q-field__control-container:has(.q-field__label:has-text("Role"))')
                .click({ timeout: 10000 });
            await this.page.waitForTimeout(500);
            await this.page.waitForSelector('.q-menu', { state: 'visible', timeout: 10000 });
            await this.page
                .locator('.q-menu [role="option"]')
                .getByText(roleName, { exact: true })
                .click({ timeout: 5000 });
        }
    }

    /**
     * Set the scope for the access rule using the scope tree modal (similar to AssetsScopeSection)
     */
    async setScope(scopeType: Scope, scopeEntityName?: string): Promise<void> {
        logger.info(`Setting scope: ${scopeType}${scopeEntityName ? ` - ${scopeEntityName}` : ''}`);

        switch (scopeType) {
            case Scope.Tenant:
                await this.selectTenantScope();
                break;
            case Scope.Project:
                await this.selectProjectScope(scopeEntityName!);
                break;
            case Scope.Department:
                await this.selectDepartmentScope(scopeEntityName);
                break;
            case Scope.Cluster:
                await this.selectClusterScope(scopeEntityName);
                break;
            default:
                throw new Error(`Unsupported scope type: ${scopeType}`);
        }
    }

    private async selectProjectScope(projectName: string): Promise<void> {
        await this.page.locator(this.scopeBtn).click();

        // Wait for modal to be fully loaded
        await this.page.waitForSelector('.q-dialog', { state: 'visible' });
        await this.page.waitForSelector('[aid="org-tree-modal-select-btn"]', { state: 'visible' });
        await this.page.waitForTimeout(1500); // Give time for modal to stabilize

        // Search for the project
        await this.page
            .locator('[placeholder="Search for a cluster, department, or project"]')
            .fill(projectName);

        // Look for the tree node with the project name and get its sibling checkbox
        const projectTreeNode = this.page.locator(`div[aid*="${projectName}"]`).first();
        await this.page.waitForTimeout(500);
        const projectCheckbox = projectTreeNode.locator('xpath=../../div[@role="checkbox"]').first();
        await projectCheckbox.click({ force: true });
        await this.page.waitForTimeout(500);

        await this.page.locator('[aid="org-tree-modal-select-btn"]').click();
    }

    private async selectTenantScope(): Promise<void> {
        await this.page.locator(this.scopeBtn).click();

        // Wait for modal to be fully loaded
        await this.page.waitForSelector('.q-dialog', { state: 'visible' });
        await this.page.waitForSelector('[aid="org-tree-modal-select-btn"]', { state: 'visible' });
        await this.page.waitForTimeout(1500); // Give time for modal to stabilize

        const tenantCheckbox = this.page.locator(`.q-checkbox:left-of([aid="${tenantName}"])`).first();
        await tenantCheckbox.click({ force: true });
        logger.info('Tenant scope selected using left-of approach');

        await this.page.waitForTimeout(500);
        await this.page.locator('[aid="org-tree-modal-select-btn"]').click();
    }

    private async selectDepartmentScope(departmentName?: string): Promise<void> {
        await this.page.locator(this.scopeBtn).click();

        // Wait for modal to be fully loaded
        await this.page.waitForSelector('.q-dialog', { state: 'visible' });
        await this.page.waitForSelector('[aid="org-tree-modal-select-btn"]', { state: 'visible' });
        await this.page.waitForTimeout(1500); // Give time for modal to stabilize

        if (departmentName) {
            // Search for the department
            await this.page
                .locator('[placeholder="Search for a cluster, department, or project"]')
                .fill(departmentName);
            await this.page.waitForTimeout(1000);

            // Look for the tree node with the department name and get its sibling checkbox
            const departmentTreeNode = this.page.locator(`div[aid*="${departmentName}"]`).first();
            const departmentCheckbox = departmentTreeNode
                .locator('xpath=../../../div[@role="checkbox"]')
                .first();
            await departmentCheckbox.click({ force: true });
        } else {
            // Click a general department checkbox
            await this.page.locator('.q-checkbox__svg').nth(2).click({ force: true });
        }
        await this.page.waitForTimeout(500);

        await this.page.locator('[aid="org-tree-modal-select-btn"]').click();
    }

    private async selectClusterScope(clusterName?: string): Promise<void> {
        await this.page.locator(this.scopeBtn).click();

        // Wait for modal to be fully loaded
        await this.page.waitForSelector('.q-dialog', { state: 'visible' });
        await this.page.waitForSelector('[aid="org-tree-modal-select-btn"]', { state: 'visible' });
        await this.page.waitForTimeout(1500); // Give time for modal to stabilize

        if (clusterName) {
            // Search for the cluster
            await this.page
                .locator('[placeholder="Search for a cluster, department, or project"]')
                .fill(clusterName);
            await this.page.waitForTimeout(1000);

            // Look for the tree node with the cluster name and get its sibling checkbox
            const clusterTreeNode = this.page.locator(`div[aid*="${clusterName}"]`).first();
            const clusterCheckbox = clusterTreeNode.locator('xpath=../../../div[@role="checkbox"]').first();
            await clusterCheckbox.click({ force: true });
        } else {
            // Click a general cluster checkbox
            await this.page.locator('.q-checkbox__svg').nth(1).click({ force: true });
        }
        await this.page.waitForTimeout(500);

        await this.page.locator('[aid="org-tree-modal-select-btn"]').click();
    }

    /**
     * Save the access rule
     */
    async saveAccessRule(): Promise<void> {
        logger.info('Saving access rule');

        await this.page.locator(this.saveRuleBtn).click();
        await this.page.waitForTimeout(3000);
    }

    /**
     * Create a single access rule with all steps
     */
    async createAccessRule(
        subjectType: 'User' | 'Group' | 'Service account',
        subjects: string[],
        roleName: string,
        scopeType: Scope = Scope.Tenant,
        scopeEntityName?: string
    ): Promise<void> {
        logger.info(
            `Creating access rule for ${subjectType}: ${subjects.join(', ')}, role: ${roleName}, scope: ${scopeType}${scopeEntityName ? ` - ${scopeEntityName}` : ''}`
        );

        await this.clickNewAccessRule();
        await this.selectSubjectType(subjectType);
        await this.addMultipleSubjects(subjects);
        await this.selectRole(roleName);
        await this.setScope(scopeType, scopeEntityName);
        await this.saveAccessRule();
    }

    /**
     * Search for access rules
     */
    async searchAccessRules(searchTerm: string): Promise<void> {
        logger.info(`Searching for access rules: ${searchTerm}`);

        const searchBox = this.page.getByRole('searchbox', { name: 'Search' }).first();
        if (!(await searchBox.isVisible({ timeout: 1000 }).catch(() => false))) {
            const searchToggleButton = this.page.locator(this.searchToggleButton).first();
            if (await searchToggleButton.isVisible({ timeout: 1000 }).catch(() => false)) {
                await searchToggleButton.click();
            } else {
                await this.page.locator(this.searchToggleIcon).first().click({ timeout: 5000 });
            }
        }

        if (await searchBox.isVisible({ timeout: 1000 }).catch(() => false)) {
            await searchBox.fill(searchTerm, { timeout: 5000 });
        } else {
            const searchInput = this.page.locator(this.searchInput).first();
            if (await searchInput.isVisible({ timeout: 1000 }).catch(() => false)) {
                await searchInput.fill(searchTerm, { timeout: 5000 });
            } else {
                await this.page.locator(this.searchInputAlt).first().fill(searchTerm, { timeout: 5000 });
            }
        }

        await this.page.waitForTimeout(1000);
    }

    async verifyAccessRuleSubject(subjectSearch: string, expectedSubjectDisplay: string): Promise<void> {
        await this.goToPage();
        await this.searchByName(subjectSearch);
        const subjectRow = this.page
            .locator(this.tableRows)
            .filter({ hasText: expectedSubjectDisplay })
            .first();
        await expect(subjectRow).toBeVisible();
        await expect(subjectRow).not.toContainText(/undefined|null/);
    }

    async verifyUserPickerResult(
        searchTerm: string,
        userEmail: string,
        userFirstName: string,
        userLastName: string
    ): Promise<void> {
        const subjectInput = this.page.locator(this.addableMultiSelect);
        await subjectInput.click();
        await subjectInput.fill(searchTerm);
        const option = this.page
            .locator(this.dropdownMenu)
            .last()
            .locator(this.dropdownOption)
            .filter({ hasText: userEmail })
            .filter({ hasText: userFirstName })
            .filter({ hasText: userLastName })
            .first();
        await expect(option).toBeVisible();
    }

    /**
     * Clear the search filter
     */
    async clearSearch(): Promise<void> {
        logger.info('Clearing search filter');

        try {
            await this.page.locator(this.clearSearchBtn).click();
        } catch {
            // Alternative: clear the input field directly
            await this.page.locator(this.searchInput).clear();
        }

        await this.page.waitForTimeout(1000);
    }

    /**
     * Select multiple access rules for bulk operations
     */
    async selectMultipleAccessRules(numberOfRules: number = 2): Promise<void> {
        logger.info(`Selecting ${numberOfRules} access rules for bulk operations`);

        // Wait for table to load
        await this.page.waitForSelector(this.tableRows, { timeout: 10000 });

        // Get all checkboxes in the table rows
        const checkboxes = this.page.locator(
            `${this.tableRows} ${this.rowCheckbox}, ${this.tableRowsAlt} ${this.rowCheckboxAlt}`
        );

        const count = await checkboxes.count();
        const rulesToSelect = Math.min(numberOfRules, count);

        for (let i = 0; i < rulesToSelect; i++) {
            await checkboxes.nth(i).click();
            await this.page.waitForTimeout(300);
        }

        logger.info(`Selected ${rulesToSelect} access rules`);
    }

    /**
     * Sort the table by creation time (newest first) and select the newest access rules
     */
    async sortByCreationTimeAndSelectNewest(numberOfAccessRule?: number): Promise<void> {
        logger.info('Sorting by creation time from newest to oldest and');
        if (numberOfAccessRule) logger.info(`selecting top ${numberOfAccessRule} newest access rules`);

        // Wait for table to load
        await this.page.waitForSelector('table', { timeout: 10000 });

        // Click on the "Creation time" column header to sort by creation time
        // First click sorts oldest-to-newest (ascending)
        const creationTimeHeader = this.page.locator(
            'th:has-text("Creation time"), th:has-text("creation time"), th:has-text("Created"), th:has-text("Date created")'
        );
        await creationTimeHeader.click();

        // Wait for the table to reload with sorted data
        await this.page.waitForTimeout(2000);
        await this.page.waitForSelector('table', { timeout: 10000 });

        // Click again to reverse the sort to newest-to-oldest (descending)
        await creationTimeHeader.click();

        // Wait for the table to reload with the reversed sort
        await this.page.waitForTimeout(2000);
        await this.page.waitForSelector('table', { timeout: 10000 });

        // Now select the top N access rules using role-based selectors
        if (numberOfAccessRule) {
            const allRows = this.page.getByRole('row');
            const dataRows = allRows.filter({ hasNot: this.page.getByRole('columnheader') }); // Exclude header row

            const count = await dataRows.count();
            const rulesToSelect = Math.min(numberOfAccessRule, count);

            logger.info(`Found ${count} access rules, selecting top ${rulesToSelect} newest ones`);

            for (let i = 0; i < rulesToSelect; i++) {
                try {
                    const row = dataRows.nth(i);
                    const checkbox = row.getByRole('checkbox');

                    const isVisible = await checkbox.isVisible();
                    if (isVisible) {
                        const isChecked = await checkbox.isChecked();
                        if (!isChecked) {
                            await checkbox.click();
                            await this.page.waitForTimeout(300);
                            logger.info(`Selected access rule ${i + 1} of ${rulesToSelect}`);
                        }
                    }
                } catch (error) {
                    logger.warn(`Failed to select access rule ${i + 1}: ${error}`);
                }
            }
            logger.info(`Selected ${rulesToSelect} newest access rules after sorting by creation time`);
        }
    }

    /**
     * Select access rules by their subject strings (both users and groups)
     * This method directly finds checkboxes in rows containing the subject strings
     */
    async selectAccessRulesBySubjects(subjects: string[]): Promise<void> {
        logger.info(`Selecting access rules for subjects: ${subjects.join(', ')}`);

        // Wait for table to load
        await this.page.waitForSelector('table', { timeout: 10000 });

        let selectedCount = 0;

        for (const subject of subjects) {
            try {
                // Use Playwright's built-in role selectors to find rows containing the subject
                // This approach is more reliable than trying to click hidden input elements
                const row = this.page.getByRole('row', { name: new RegExp(subject, 'i') });
                const rowCount = await row.count();

                if (rowCount > 0) {
                    logger.info(`Found ${rowCount} access rules for subject: ${subject}`);

                    // For each row containing this subject, click its checkbox
                    for (let i = 0; i < rowCount; i++) {
                        const currentRow = row.nth(i);
                        const checkbox = currentRow.getByRole('checkbox');

                        // Check if checkbox exists and is not already checked
                        const isVisible = await checkbox.isVisible();
                        if (isVisible) {
                            const isChecked = await checkbox.isChecked();
                            if (!isChecked) {
                                await checkbox.click();
                                await this.page.waitForTimeout(300);
                                selectedCount++;
                                logger.info(`Selected access rule checkbox for subject: ${subject}`);
                            } else {
                                logger.info(`Checkbox for subject ${subject} is already checked`);
                            }
                        }
                    }
                } else {
                    logger.warn(`No access rules found for subject: ${subject}`);
                }
            } catch (error) {
                logger.error(`Error selecting checkbox for subject ${subject}: ${error}`);
            }
        }

        logger.info(`Selected ${selectedCount} access rules by subject strings`);
    }

    /**
     * Select access rules by user email addresses
     * This method specifically targets user emails in the subject column
     */
    async selectAccessRulesByUserEmails(userEmails: string[]): Promise<void> {
        logger.info(`Selecting access rules for user emails: ${userEmails.join(', ')}`);

        // Wait for table to load
        await this.page.waitForSelector('table', { timeout: 10000 });

        let selectedCount = 0;

        for (const userEmail of userEmails) {
            try {
                // Find all rows that contain this user email in the subject column
                // Use a more specific pattern to match email addresses
                const rows = this.page.getByRole('row').filter({ hasText: userEmail });
                const rowCount = await rows.count();

                if (rowCount > 0) {
                    logger.info(`Found ${rowCount} access rules for user email: ${userEmail}`);

                    // For each row containing this user email, click its checkbox
                    for (let i = 0; i < rowCount; i++) {
                        const currentRow = rows.nth(i);

                        // Double-check this row actually contains the email (not just partial match)
                        const rowText = await currentRow.textContent();
                        if (rowText && rowText.includes(userEmail)) {
                            const checkbox = currentRow.getByRole('checkbox');

                            const isVisible = await checkbox.isVisible();
                            if (isVisible) {
                                const isChecked = await checkbox.isChecked();
                                if (!isChecked) {
                                    await checkbox.click();
                                    await this.page.waitForTimeout(300);
                                    selectedCount++;
                                    logger.info(`Selected access rule checkbox for user: ${userEmail}`);
                                } else {
                                    logger.info(`Checkbox for user ${userEmail} is already checked`);
                                }
                            }
                        }
                    }
                } else {
                    logger.warn(`No access rules found for user email: ${userEmail}`);
                }
            } catch (error) {
                logger.error(`Error selecting checkbox for user ${userEmail}: ${error}`);
            }
        }

        logger.info(`Selected ${selectedCount} access rules by user emails`);
    }

    /**
     * Delete selected access rules
     */
    async deleteSelectedAccessRules(): Promise<void> {
        logger.info('Deleting selected access rules');

        // Step 1: Wait for the selection bar to appear with delete button
        await this.page.waitForSelector(':has-text("selected")', { timeout: 1000 });

        // Step 2: Click the FIRST delete button in the selection bar
        // Be specific to avoid the "strict mode violation" with the modal delete button
        // Find the selection bar first, then the delete button within it
        await this.page.locator('button.runai-action-button:has-text("delete")').click();

        // Step 3: Wait for the confirmation modal to appear
        // await this.page.waitForSelector('dialog', { timeout: 1000 });

        // Step 4: Click the confirmation button in the modal
        // The button text is "Delete X access rules" where X is the number of selected items
        // Target the button specifically within the dialog to avoid strict mode violation
        await this.page.locator('[aid="delete-modal-delete-button"]').click();

        // Step 5: Wait for the deletion to complete and modal to disappear
        await this.page.waitForSelector('[aid="delete-modal-delete-button"]');
        await this.page.waitForSelector('dialog', { state: 'detached', timeout: 10000 });

        logger.info('Access rules deletion completed');
    }

    /**
     * Wait for success notification
     */
    async waitForSuccessNotification(): Promise<void> {
        logger.info('Waiting for success notification');
        await this.page.waitForSelector(this.successNotification, { timeout: 10000 });
    }

    /**
     * Wait for error notification
     */
    async waitForErrorNotification(): Promise<void> {
        logger.info('Waiting for error notification');
        await this.page.waitForSelector(this.errorNotification, { timeout: 10000 });
    }

    /**
     * Check if access rule exists in table
     */
    async accessRuleExists(subjectName: string, roleName?: string): Promise<boolean> {
        try {
            if (roleName) {
                return await this.page
                    .locator(`td:has-text("${subjectName}"), [role="cell"]:has-text("${subjectName}")`)
                    .and(
                        this.page.locator(`td:has-text("${roleName}"), [role="cell"]:has-text("${roleName}")`)
                    )
                    .isVisible();
            } else {
                return await this.page
                    .locator(`td:has-text("${subjectName}"), [role="cell"]:has-text("${subjectName}")`)
                    .isVisible();
            }
        } catch {
            return false;
        }
    }

    /**
     * Get the number of access rules in the table
     */
    async getAccessRulesCount(): Promise<number> {
        try {
            await this.page.waitForSelector(this.tableRows, { timeout: 5000 });
            return await this.page.locator(this.tableRows).count();
        } catch {
            return 0;
        }
    }
}
