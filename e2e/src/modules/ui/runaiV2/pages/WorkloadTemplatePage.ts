/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { BaseV2Page } from './BaseV2Page';
import { pageFilterUtils } from '../../utils/pageFilterUtils';
import { logger, waitUtils } from '../../../helpers';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';

export class WorkloadTemplatePage extends BaseV2Page {
    addNewWorkloadTemplateBtn: string;

    newInferenceBtn: string;

    templateNameInput: string;

    continueBtn: string;

    createInferenceTemplateBtn: string;

    submitInferenceMultiBtns: string;

    advancedSetupBtn: string;

    templateNameCol: string;

    copyTemplateBtn: string;

    editTemplateBtn: string;

    constructor(page: Page) {
        super(page);

        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Templates}"]`;
        this.addNewWorkloadTemplateBtn = '[aid="new-template-dropdown"] .runai-dropdown';
        this.newInferenceBtn = '[aid="new-template-inference-option"]';
        this.templateNameInput = '[aid="workload-template-name-input"]';
        this.continueBtn = '.template-hybrid-new [aid="submit-btn"]'; // TODO: change to the correct selector
        this.createInferenceTemplateBtn = '.inference-template-hybrid [aid="submit-btn"]';

        this.submitInferenceMultiBtns = '.inference-template-hybrid [aid="submit-btn"] button';
        this.advancedSetupBtn = '[aid="runai-dropdown-btn-select-advanced-setup-btn"]';
        this.templateNameCol = '[aid="name"]';
        this.copyTemplateBtn = '[aid="copy-template-btn"]';
        this.editTemplateBtn = '[aid="edit-template-btn"]';
    }

    async navigateToWorkloadTemplates(): Promise<void> {
        await waitUtils.sleep(3000);

        await this.page.locator(this.subjectItem).click();
        await this.page.locator(this.subMenuItem).click();

        //a workaround to close the side menu
        await pageFilterUtils.selectCluster(this.page);
    }

    async openInferenceForm(): Promise<void> {
        await this.page.locator(this.addNewWorkloadTemplateBtn).click();
        await this.page.locator(this.newInferenceBtn).click();
    }

    async fillTemplateName(name: string): Promise<void> {
        await this.page.locator(this.templateNameInput).fill(name);
    }

    async clickOnContinue(): Promise<void> {
        await this.page.locator(this.continueBtn).click();
    }

    async clickOnCreateInferenceTemplateBtn(): Promise<void> {
        await this.page.locator(this.createInferenceTemplateBtn).click();
    }

    async clickOnAdvancedSetupBtn(): Promise<void> {
        const btns = await this.locator(this.submitInferenceMultiBtns).all();
        if (btns.length > 1) {
            await btns[1].click();
            await this.locator(this.advancedSetupBtn).click();
        } else {
            await btns[0].click();
        }
    }

    async isElementVisible(query: string): Promise<boolean> {
        return this.page.locator(query).isVisible();
    }

    async toggleTemplate(templateName: string): Promise<void> {
        // Wait for the template table header to be visible
        await this.page.locator('th:has-text("Template")').waitFor({ state: 'visible' });

        // Narrow the grid via the page-level search so the target row is guaranteed
        // to be rendered regardless of pagination / virtualization / default sort.
        await this.searchByName(templateName);

        const row = this.locator(`td${this.templateNameCol}`)
            .filter({ hasText: templateName })
            .first();
        try {
            await row.waitFor({ state: 'visible', timeout: 15000 });
        } catch (err) {
            // Only translate Playwright timeouts into a "not found" error.
            // Other failures (navigation, closed context, etc.) must surface as-is
            // so real infra issues aren't masked as missing templates.
            if (err instanceof Error && err.name === 'TimeoutError') {
                logger.error(`Template ${templateName} not found`);
                throw new Error(`Template ${templateName} not found`);
            }
            throw err;
        }

        await row.click();
    }

    async clickOnCopyTemplate(): Promise<void> {
        await this.page.locator(this.copyTemplateBtn).click();
    }

    async clickOnEditTemplate(): Promise<void> {
        await this.page.locator(this.editTemplateBtn).click();
    }
}
