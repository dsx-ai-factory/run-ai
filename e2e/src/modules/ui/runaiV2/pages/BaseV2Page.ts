/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page, Locator } from 'playwright';
import { config, flowRootState, logger, settings, waitUtils } from '../../../helpers';
import { Response } from 'playwright-core';
import { tenantName } from '../../../helpers/config/const';
import { readFileSync, existsSync, mkdirSync } from 'fs';
import { join } from 'path';

export abstract class BaseV2Page {
    // global
    page: Page;

    mainPageTab: string;

    mainPageTitle = '[aid="page-title"]';

    subjectItem: string;

    subMenuItem: string;

    addFilterBtn: string;

    inputFilterContains: string;

    applyFilterBtn: string;

    freeTextFilterSelect: string;

    freeTextFilterInput: string;

    modalLeaveBtn: string;

    viewOnlyBadge: string;

    clearSearchBtn: string;

    modalDialog: string;

    noRolesDiv: string;

    headerLeavePageBtn: string;

    downloadCSVOption: string;

    rowsPerPageSelect: string;

    helpMenuButton: string;

    leavePageDialog: string;

    leavePageCancelBtn: string;

    leavePageStayBtn: string;

    helpGuidedStepsMenu: string;

    helpGuidedStepsOverviewDashboardMenu: string;

    baseDialogCloseBtn: string;

    runaiBaseDialogBody: string;

    constructor(page: Page) {
        this.page = page;
        this.addFilterBtn = '//span[contains(text(),"Add filter")]';
        this.inputFilterContains = 'input[aria-label="Contains"]';
        this.applyFilterBtn = '//span[contains(text(),"Apply")]';
        this.freeTextFilterSelect = '[aid="free-text-filter-select"]';
        this.freeTextFilterInput = '[aid="free-text-filter-input"]';
        this.modalLeaveBtn = '[aid="leave-page-modal-cancel-btn"]';
        this.viewOnlyBadge = '[aid="view-only-badge"]';
        this.clearSearchBtn = 'i.search-clear';
        this.modalDialog = '.q-card.dialog-content-container.width-sm';
        this.noRolesDiv = '[aid="no-roles"]';
        this.headerLeavePageBtn = '[aid="header-leave-page-button"]';

        this.downloadCSVOption = 'text=Download as CSV';
        this.rowsPerPageSelect = '.q-table__bottom label.q-table__select';
        this.helpMenuButton = '[aid="help-dropdown"]';
        this.helpGuidedStepsMenu = '[aid="guided-steps-menu"]';
        this.helpGuidedStepsOverviewDashboardMenu = '[aid="overview-guided-steps-option"]';

        this.leavePageDialog = '.runai-base-dialog .dialog-content-container';
        this.leavePageCancelBtn = '[aid="leave-page-modal-cancel-btn"]';
        this.leavePageStayBtn = '[aid="leave-page-modal-stay-btn"]';
        this.baseDialogCloseBtn = '.runai-base-dialog-close-btn';
        this.runaiBaseDialogBody = '.runai-base-dialog .runai-base-dialog-body';
    }

    url(): string {
        return this.page.url();
    }

    /**
     * @deprecated this getter it deprecated, please use the page directly or locator
     */
    get v2frame(): Page {
        return this.page;
    }

    async fillInputBySelector(selector: string, value: string): Promise<void> {
        return await this.locator(selector).fill(value);
    }

    async selectOptionByValue(selector: string, value: string): Promise<void> {
        try {
            await this.locator(selector).click();
            await this.locator(`div[role="option"]:has-text("${value}")`).click();
            return;
        } catch (error) {
            logger.error(`Failed to select option by value: ${value}`, error);
            const options = await this.locator('div[role="option"]').all();
            logger.debug('listing the options:');
            for (const option of options) {
                const text = await option.textContent();
                logger.debug('option text', text);
            }
            logger.debug('listing the options done');
            throw error;
        }
    }

    async selectOptionByExactValue(selector: string, value: string): Promise<void> {
        await this.locator(selector).click();
        return this.page.getByRole('option', { name: value, exact: true }).click();
    }

    async selectLastOptionByValue(selector: string, value: string): Promise<void> {
        await this.locator(selector).last().click();
        return await this.locator(`div[role="option"]:has-text("${value}")`).click();
    }

    async switchToggleBySelector(selector: string, value: boolean): Promise<void> {
        const toggleHtml: string = await this.locator(selector).first().innerHTML();
        if (
            (toggleHtml.includes('q-toggle__inner--falsy') && value === true) ||
            (toggleHtml.includes('q-toggle__inner--truthy') && value === false)
        ) {
            await this.locator(selector).click();
        }
    }

    async toggleFoldingSubSection(selector: string, value: boolean): Promise<void> {
        const foldingSectionHtml = await this.page.locator(`${selector}`).innerHTML();
        if (
            (foldingSectionHtml.includes('q-expansion-item--collapsed') && value) ||
            (foldingSectionHtml.includes('q-expansion-item--expanded') && !value)
        ) {
            return this.page
                .locator(`${selector} .q-expansion-item .runai-folding-sub-section-header`)
                .click();
        }
    }

    async selectCurrentCluster(): Promise<void> {
        try {
            await this.page.waitForSelector('[aid="cluster-select"]', { timeout: 1000 });
        } catch (_error) {
            logger.info('Cluster selector is not visible');
            return;
        }

        await this.selectOptionByValue('[aid="cluster-select"]', flowRootState.cluster.name);
    }

    async selectFirstOption(selector: string): Promise<void> {
        await this.locator(selector).click();
        return await this.locator('div[role="option"]').first().click();
    }

    async getPageTitle(): Promise<string> {
        return await this.page.locator(this.mainPageTitle).textContent();
    }

    findByText(element: string, text: string): Locator {
        return this.locator(`${element}:has-text("${text}")`);
    }

    async clickOnElementByText(element: string, text: string): Promise<void> {
        return await this.findByText(element, text).click();
    }

    async openPythonConsole(page: Page): Promise<void> {
        logger.debug('Going to open python console');
        await page.locator('.jp-LauncherCard-icon').first().click();
        logger.debug('Python console was open');
    }

    async clickOnElementByAidName(name: string): Promise<void> {
        return this.locator(`[aid='${name}']`).click();
    }

    async goToPage(): Promise<void> {
        await this.page.waitForLoadState('networkidle');

        await this.page.locator(this.subjectItem).click();
        await this.page.waitForTimeout(300);
        await this.page.locator(this.subMenuItem).click();

        await this.page.locator(this.mainPageTitle).click(); // this is a workaround to close the sidenav bar
        await this.page.waitForLoadState('load');
        await this.page.waitForTimeout(500); // wait some time for the page to load
    }

    async waitForResponseByCallback(cb: (response: Response) => boolean): Promise<void> {
        await this.page.waitForResponse(cb);
    }

    async waitForResponse(url: string): Promise<void> {
        await this.page.waitForResponse(`${config.appUrl}${url}`);
    }

    /**
     * Generic `page.waitForResponse` wrapper that, on timeout, reports every response whose URL
     * matched `urlContains` (with status), so the failure log clearly indicates whether no matching
     * request fired at all, fired with an unexpected status, or never satisfied the full predicate.
     *
     * Use this instead of a bare `page.waitForResponse(predicate)` whenever the wait can be flaky
     * and silent timeouts hurt triage.
     */
    async waitForResponseWithDiagnostics({
        predicate,
        urlContains,
        label,
        timeoutMs = 25_000
    }: {
        predicate: (response: Response) => boolean;
        urlContains: string;
        label: string;
        timeoutMs?: number;
    }): Promise<void> {
        const observed: Array<{ url: string; status: number }> = [];

        const responseHandler = (response: Response): void => {
            const url = response.url();
            if (url.includes(urlContains)) {
                observed.push({ url, status: response.status() });
                logger.debug(
                    `[${label}] observed status=${response.status()} url=${url}`
                );
            }
        };
        this.page.on('response', responseHandler);

        try {
            await this.page.waitForResponse(predicate, { timeout: timeoutMs });
        } catch (err) {
            const summary =
                observed.length === 0
                    ? `no responses matching '${urlContains}' observed at all (UI did not fire it)`
                    : `observed ${observed.length} response(s) matching '${urlContains}': ` +
                      observed.map((o) => `${o.status} ${o.url}`).join(' | ');
            logger.error(
                `[${label}] waitForResponseWithDiagnostics timeout after ${timeoutMs}ms — ${summary}`
            );
            throw err;
        } finally {
            this.page.off('response', responseHandler);
        }
    }

    async waitForScopeSectionContent(projectName: string): Promise<void> {
        const departmentPath: string = 'default/';
        await waitUtils.waitForElementContent(
            this,
            '[aid="scope-section-subheader"]',
            `${tenantName}/${flowRootState.cluster.name}/${departmentPath}${projectName}`
        );
    }

    async waitForTenantScopeSection(): Promise<void> {
        await waitUtils.waitForElementContent(this, '[aid="scope-section-subheader"]', 'runai');
    }

    async waitForAssetToBeDeletedFromUI(page: BaseV2Page, selector: string, timeout?: number): Promise<void> {
        const assetSelector = `td:has-text("${selector}")`;
        await waitUtils.waitForElementToNotExist(page, assetSelector, timeout);
    }

    async getModalHeader(): Promise<string> {
        return (await this.page.locator('.runai-base-dialog-header').textContent()).trim();
    }

    async clickSearchBtn(): Promise<void> {
        return await this.locator('.runai-page-actions .runai-general-search').click();
    }

    async searchByName(name: string): Promise<void> {
        await this.clickSearchBtn();
        await this.locator('.runai-page-actions .runai-general-search .search-input').fill(name);
    }

    async reload(): Promise<void> {
        await this.page.reload();
    }

    locator(selector: string): Locator {
        return this.page.locator(selector);
    }

    async isVisible(locator: Locator): Promise<boolean> {
        const count = await locator.count();
        return count > 0;
    }

    async waitForVisible(locator: Locator, timeout?: number): Promise<void> {
        await locator.waitFor({ state: 'visible', timeout });
    }

    async ClusterFilterSelect(filterName: string, inputData: string): Promise<string> {
        await this.page.locator(this.addFilterBtn).click();
        const filterOption = this.page.locator(
            `div.q-item__section.column.q-item__section--main:has-text("${filterName}")`
        );
        await filterOption.click();
        const checkbox = this.page.locator(`.q-checkbox[aria-label="${inputData}"]`);
        await checkbox.click();
        await this.page.locator(this.applyFilterBtn).click();
        const filterChip = this.page.locator('.filter-chip:has-text("Status")');
        await filterChip.waitFor({ state: 'visible', timeout: 10000 });
        return await filterChip.innerText();
    }

    async waitForNotificationDisappear(): Promise<void> {
        await this.page.waitForSelector('div.q-notification__message.col', {
            state: 'detached',
            timeout: 10000
        });
    }

    async getNotificationMsg(): Promise<string> {
        const notification = await this.page.waitForSelector('div.q-notification__message.col', {
            timeout: 5000
        });
        const message = await notification.textContent();
        return message;
    }

    async clickModalLeaveBtn(): Promise<void> {
        const exist = await this.isModalDialogVisible();
        if (exist) await this.page.locator(this.modalLeaveBtn).click();
    }

    async viewOnlyBadgeVisible(): Promise<boolean> {
        try {
            await this.page.locator(this.viewOnlyBadge).waitFor({ state: 'visible', timeout: 3000 });
            return true;
        } catch {
            return false;
        }
    }

    async elementEnabled(locator: string, timeout = 2000): Promise<boolean> {
        try {
            await this.page.waitForFunction(
                (selector) => {
                    const el = document.querySelector(selector);
                    return el !== null && !el.hasAttribute('disabled');
                },
                locator,
                { timeout }
            );
            return true;
        } catch {
            return false;
        }
    }

    async clearSearchInput(): Promise<void> {
        await this.page.locator(this.clearSearchBtn).click();
    }

    async isModalDialogVisible(): Promise<boolean> {
        try {
            await this.page.locator(this.modalDialog).waitFor({ state: 'visible', timeout: 3000 });
            return true;
        } catch {
            return false;
        }
    }

    async leavePageWithConfirmation(): Promise<void> {
        await this.page.locator(this.headerLeavePageBtn).click();
        await this.clickModalLeaveBtn();
    }

    async downloadCSV(
        expectedFileName: string,
        downloadDir: string = settings.browserVideosPath
    ): Promise<string> {
        const downloadPromise = this.page.waitForEvent('download');

        await this.page.getByRole('button', { name: 'More', exact: true }).click();
        await this.page.locator(this.downloadCSVOption).click();

        const download = await downloadPromise;

        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        const extension = download.suggestedFilename().split('.').pop() || 'csv';
        const fileName = `${expectedFileName}_${timestamp}.${extension}`;

        if (!existsSync(downloadDir)) {
            mkdirSync(downloadDir, { recursive: true });
        }

        const downloadPath = join(downloadDir, fileName);
        await download.saveAs(downloadPath);

        logger.info(`CSV file downloaded to: ${downloadPath}`);
        return downloadPath;
    }

    async setRowsPerPage(rowsPerPage: number): Promise<void> {
        await this.page.locator(this.rowsPerPageSelect).click();
        await this.page.getByRole('option', { name: rowsPerPage.toString(), exact: true }).click();
        await this.page.waitForTimeout(1000);
        logger.info(`Set rows per page to: ${rowsPerPage}`);
    }

    async clickOnHelpMenuButton(): Promise<void> {
        await this.page.locator(this.helpMenuButton).click();
    }

    async selectHelpMenuOptionByText(option: string): Promise<void> {
        await this.clickOnHelpMenuButton();
        await this.page
            .locator('.help-dropdown-menu-container')
            .locator('.q-list')
            .locator(`div[role="listitem"]:has-text("${option}")`)
            .click();
    }

    async pressKeyOnElement(selector: string, key: string): Promise<void> {
        await this.page.locator(selector).focus();
        return this.page.locator(selector).press(key);
    }

    async pressEscKey(): Promise<void> {
        return this.page.keyboard.press('Escape');
    }

    async pressEnterKey(): Promise<void> {
        return this.page.keyboard.press('Enter');
    }

    async verifyLeavePageDialogAppears(timeout: number = 5000): Promise<void> {
        logger.info('Verifying "Leave This Page?" dialog appears');
        const dialogContainer = this.page.locator(this.leavePageDialog);
        await dialogContainer.waitFor({ state: 'visible', timeout });

        await dialogContainer.locator('h6 b:has-text("Leave This Page?")').waitFor({ state: 'visible' });
        await dialogContainer.locator(this.leavePageCancelBtn).waitFor({ state: 'visible' });
        await dialogContainer.locator(this.leavePageStayBtn).waitFor({ state: 'visible' });
    }

    async clickOnColumnsBtn(): Promise<void> {
        await this.page.getByRole('button', { name: 'Columns' }).click();
    }

    async setColumn(columnName: string, checked = true): Promise<void> {
        await this.clickOnColumnsBtn();
        const isChecked = await this.page.getByRole('checkbox', { name: columnName }).isChecked();
        if (checked !== isChecked) {
            await this.page.getByLabel(columnName).click();
        }

        await this.clickOnColumnsBtn();
    }

    async isElementVisibleWithWait(locator: string, timeout = 3000): Promise<boolean> {
        try {
            await this.page.locator(locator).waitFor({ state: 'visible', timeout });
            return true;
        } catch {
            return false;
        }
    }

    async clickAddFilterButton(): Promise<void> {
        await this.page.locator(this.addFilterBtn).click();
    }

    async clickApplyFilterButton(): Promise<void> {
        await this.page.locator(this.applyFilterBtn).click();
    }

    async addFilter(filterName: string, filterValue: string, operator: string = 'contains'): Promise<void> {
        await this.clickAddFilterButton();
        await this.page
            .getByRole('menuitem')
            .filter({ has: this.page.getByText(filterName, { exact: true }) })
            .click();
        if (operator !== 'contains') {
            await this.page.locator(this.freeTextFilterSelect).click();
            await this.page.getByRole('option', { name: operator, exact: true }).click();
        }
        await this.page.locator(this.freeTextFilterInput).fill(filterValue);
        await this.page.waitForTimeout(1000); // the filter input autocomplete is debounced by 600ms
        await this.closeFilterAutocompleteDropdown();
        await this.clickApplyFilterButton();
    }

    async closeFilterAutocompleteDropdown(): Promise<void> {
        const filterInput = this.locator(this.freeTextFilterInput);

        if ((await filterInput.getAttribute('aria-expanded')) === 'true') {
            await filterInput.press('Escape');
        }
    }

    async closeFilter(): Promise<void> {
        await this.page.locator('.filter-chip').getByRole('button', { name: 'Remove', exact: true }).click();
    }

    async clickBaseDialogCloseBtn(): Promise<void> {
        await this.page.locator(this.baseDialogCloseBtn).click();
    }
}
