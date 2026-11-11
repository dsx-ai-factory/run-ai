/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { flowRootState, waitUtils } from '../../../helpers';
import { AssetsScopeSection } from '../components/AssetsScopeSection';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { pageFilterUtils } from '../../utils/pageFilterUtils';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';

// ESM equivalent of __dirname
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export class PolicyPage extends BaseV2Page {
    // components
    assetsScopeSection: AssetsScopeSection;

    toolSettingsIcon: string;

    policyIcon: string;

    addNewPolicyBtn: string;

    policyScopeBtn: string;

    policySelectScopeBtn: string;

    workloadSelect: string;

    policyAddYamlBtn: string;

    policyApplyYamlBtn: string;

    policySubmitBtn: string;

    policyDeleteBtn: string;

    policyModalDeleteBtn: string;

    headerBackButton: string;

    async selectFirstOption(selector: string): Promise<void> {
        return super.selectFirstOption(selector);
    }

    constructor(page: Page) {
        super(page);
        this.assetsScopeSection = new AssetsScopeSection(page);
        this.toolSettingsIcon = '[aid="setup-dropdown"]';
        this.policyIcon = '[aid="policies-btn"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Policies}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Policies}"]`;
        this.addNewPolicyBtn = '[aid="primary-cta-btn"]';
        this.policyScopeBtn = '[aid="scope-tree-btn"]';
        this.policySelectScopeBtn = '[aid="org-tree-modal-select-btn"]';
        this.workloadSelect = '[aid="workload-type-select"]';
        this.policyAddYamlBtn = '[aid="add-policy-btn"]';
        this.policyApplyYamlBtn = '[aid="editor-modal-apply-button"]';
        this.policySubmitBtn = '[aid="submit-btn"]';
        this.policyDeleteBtn = '[aid="delete-policy-btn"]';
        this.policyModalDeleteBtn = '[aid="delete-modal-delete-button"]';
        this.headerBackButton = '[aid=header-leave-page-button]';
    }

    async navigateToPolicyPage(): Promise<void> {
        await this.page.locator(this.subjectItem).click();
        await this.page.waitForTimeout(300);
        await this.page.locator(this.subMenuItem).click();

        await pageFilterUtils.selectCluster(this.page);
    }

    async clickOnPolicyApplyYamlBtn(): Promise<void> {
        await this.v2frame.locator(this.policyApplyYamlBtn).click();
    }

    async clickOnPolicySubmitBtn(): Promise<void> {
        await this.v2frame.locator(this.policySubmitBtn).click();
        await waitUtils.waitForElementToNotExist(this, this.policySubmitBtn, 40); //submission terminated
    }

    async waitForOnPolicySubmitBtnToAppear(): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.policySubmitBtn);
    }

    async clickOnPolicyDeleteBtn(): Promise<void> {
        await this.v2frame.locator(this.policyDeleteBtn).click();
    }

    async clickOnPolicyModalDeleteBtn(): Promise<void> {
        await this.v2frame.locator(this.policyModalDeleteBtn).click();
    }

    async clickOnAddNewPolicyBtn(): Promise<void> {
        await this.v2frame.locator(this.addNewPolicyBtn).click();
    }

    async clickOnPolicyYamlAddBtn(): Promise<void> {
        await this.v2frame.locator(this.policyAddYamlBtn).click();
    }

    async selectWorkloadByText(workloadText: string): Promise<void> {
        await this.selectOptionByValue(this.workloadSelect, workloadText);
    }

    async InjectPolicyToMonacoYamlEditor(yamlFileName: string): Promise<void> {
        await this.v2frame.locator('.monaco-editor').nth(0).click();
        const selectAllKey = this.isMac() ? 'Meta+A' : 'Control+A';
        await this.page.keyboard.press(selectAllKey);
        await this.page.keyboard.press('Backspace');

        await this.readFromFileAndCopyToClipboard(yamlFileName);
        await this.pasteFromClipboard();
    }

    async selectPolicyInTableByName(policyName: string): Promise<void> {
        await this.findByText('td', policyName).click();
    }

    async getPolicyFromTableByName(policyName: string): Promise<string> {
        return (await this.findByText('td', policyName).textContent()).trim();
    }

    async verifyPolicyDeletedSuccessfully(policyName: string): Promise<void> {
        await waitUtils.waitForElementToNotExist(this, `td:has-text(" ${policyName}")`);
    }

    isMac(): boolean {
        return process.platform === 'darwin';
    }

    async indentCodeInEditor(): Promise<void> {
        await this.page.keyboard.press('Alt+Shift+F');
    }

    // Note: this function only looking for files in the following directory: src/modules/ui/utils/predefinedPolicyYamls
    async readFromFileAndCopyToClipboard(fileName: string): Promise<void> {
        const filePath = path.join(__dirname, '../../utils/predefinedPolicyYamls', fileName);
        const content = fs.readFileSync(filePath, 'utf8');

        await this.page.evaluate((data) => {
            navigator.clipboard.writeText(data);
        }, content);

        await this.page.evaluateHandle(() => navigator.clipboard.readText());
    }

    async pasteFromClipboard(): Promise<void> {
        const key = this.isMac() ? 'Meta' : 'Control';
        await this.page.keyboard.press(`${key}+V`);
    }
}
