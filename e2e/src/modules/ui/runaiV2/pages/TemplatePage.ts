/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { logger, waitUtils } from '../../../helpers';
import { pageFilterUtils } from '../../utils/pageFilterUtils';
import { MENU_SUBJECT_NAMES, MENU_SUB_ITEMS } from '../../utils/sidenavUtils';
import { Scope } from '../../../api/apis/assetsApi';
import { AssetsScopeSection } from '../components/AssetsScopeSection';

export class TemplatePage extends BaseV2Page {
    addNewWorkspaceTemplateBtn: string;

    createTemplateNameInput: string;

    createTemplateBtn: string;

    newEnvironmentBtn: string;

    newComputeBtn: string;

    deleteTemplateBtn: string;

    deleteTemplateModalDeleteBtn: string;

    environmentSection: string;

    computeResourceSection: string;

    dataSourceSection: string;

    dataSourceExpansion: string;

    workloadFormRunaiGeneralSearch: string;

    skeletonCardsList: string;

    runaiDefaultLoader: string;

    assetsScopeSection: AssetsScopeSection;

    templateCol: string;

    copyTemplateBtn: string;

    renameBtn: string;

    renameInput: string;

    renameModalSaveBtn: string;

    constructor(page: Page) {
        super(page);
        this.mainPageTab = '[aid="menu-icon-workloads"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Templates}"]`;
        this.addNewWorkspaceTemplateBtn = '.runai-page-actions button[aid="primary-cta-btn"]';
        this.createTemplateNameInput = '[aid="workload-template-name-input"]';
        this.createTemplateBtn = '[aid="create-template-btn"]';
        this.newEnvironmentBtn = '[aid="add-new-environment-btn"]';
        this.newComputeBtn = '[aid="add-new-compute-resource-btn"]';
        this.deleteTemplateBtn = '[aid="delete-template-btn"]';
        this.deleteTemplateModalDeleteBtn = '[aid="delete-modal-delete-button"]';
        this.environmentSection = '.environment-section';
        this.computeResourceSection = '.compute-resource-section';
        this.dataSourceSection = '.data-source-section';
        this.dataSourceExpansion = '[aid="data-source-section"]';
        this.workloadFormRunaiGeneralSearch = '.runai-general-search';
        this.skeletonCardsList = '[aid="skeleton-list"]';
        this.runaiDefaultLoader = '.runai-default-loader';
        this.assetsScopeSection = new AssetsScopeSection(page);
        this.templateCol = '[aid="template"]';
        this.copyTemplateBtn = '[aid="copy-template-btn"]';
        this.renameBtn = 'rename';
        this.renameInput = '[aid="runai-name-validation"]';
        this.renameModalSaveBtn = '[aid="rename-modal-save-button"]';
    }

    async navigateToTemplates(): Promise<void> {
        await waitUtils.sleep(3000);
        await this.page.locator(this.subjectItem).click();
        await this.page.locator(this.subMenuItem).click();
        await this.page.locator(this.mainPageTitle).click();
        await pageFilterUtils.selectCluster(this.page);
    }

    async openNewTemplatePage(): Promise<void> {
        await this.page.locator(this.addNewWorkspaceTemplateBtn).click();
    }

    async fillTemplateName(name: string): Promise<void> {
        await this.page.locator(this.createTemplateNameInput).fill(name);
    }

    async clickOnNewEnvironmentBtn(): Promise<void> {
        return this.page.locator(this.newEnvironmentBtn).click();
    }

    async clickOnNewComputeBtn(): Promise<void> {
        return this.page.locator(this.newComputeBtn).click();
    }

    async clickOnCreateTemplateBtn(): Promise<void> {
        return this.page.locator(this.createTemplateBtn).click();
    }

    async clickOnDeleteTemplateBtn(): Promise<void> {
        return this.page.locator(this.deleteTemplateBtn).click();
    }

    async clickOnDeleteTemplateInModalBtn(): Promise<void> {
        return this.page.locator(this.deleteTemplateModalDeleteBtn).click();
    }

    async clickOnRenameBtn(): Promise<void> {
        await this.page.getByRole('button', { name: this.renameBtn }).click();
    }

    async clearNameInRenameModal(): Promise<void> {
        await this.page.locator(this.renameInput).clear();
    }

    async fillNameInRenameModal(name: string): Promise<void> {
        await this.page.locator(this.renameInput).fill(name);
    }

    async clearTemplateNameInput(): Promise<void> {
        await this.page.locator(this.createTemplateNameInput).clear();
    }

    async renameTemplate(oldName: string, newName: string): Promise<void> {
        await this.clickOnElementByText('td', oldName);
        await this.clickOnRenameBtn();
        await this.clearNameInRenameModal();
        await this.fillNameInRenameModal(newName);
        await this.page.locator(this.renameModalSaveBtn).click();
    }

    async openCopyAndEditForm(name: string): Promise<void> {
        await this.clickOnElementByText('td', name);
        await this.clickOnCopyTemplate();
    }

    async removeTemplate(name: string): Promise<void> {
        await this.clickOnElementByText('td', name);
        await this.clickOnDeleteTemplateBtn();
        await this.clickOnDeleteTemplateInModalBtn();
    }

    async waitForWorkspaceTemplateNameCount(
        name: string,
        scopeType: string,
        scopeId: string | number
    ): Promise<void> {
        await this.waitForResponse(
            `/api/v1/asset/name-availability?kind=workload-template&assetName=${name}&scopeType=${scopeType}&scopeId=${scopeId}`
        );
    }

    async getTemplateNames(): Promise<string[]> {
        const templateNames = await this.page.locator(this.templateCol).allTextContents();
        return templateNames.map((name) => name.trim());
    }

    async getTemplateRowIndexByName(templateName: string): Promise<number> {
        const templateNames: string[] = await this.getTemplateNames();
        console.log(templateNames);
        const index = templateNames.findIndex((name: string) => name.trim() === templateName);
        if (index === -1) {
            throw new Error(`template row with ${templateName} text not found`);
        } else {
            return index;
        }
    }

    async toggleTemplate(templateName: string): Promise<void> {
        await this.clickOnElementByText(`td${this.templateCol}`, `${templateName}`);
    }

    async deleteTemplate(templateName: string): Promise<void> {
        const visible = await this.isElementVisible(this.deleteTemplateBtn);
        if (!visible) await this.toggleTemplate(templateName);

        await this.clickOnDeleteTemplateBtn();
        await this.clickOnDeleteTemplateInModalBtn();
        await waitUtils.waitForElementToNotExist(this, `td:has-text("${templateName}")`, 90);
        logger.info(`successfully deleted template: ${templateName}`);
    }

    async isElementVisible(query: string): Promise<boolean> {
        return this.page.locator(query).isVisible();
    }

    getCardSelector(name: string): string {
        return `[aid="${name}-card"]`;
    }

    async fillWorkloadFormGeneralSearchInput(sectionSelector: string, name: string): Promise<void> {
        await this.page.locator(`${sectionSelector} ${this.workloadFormRunaiGeneralSearch}`).click();
        await this.page.locator(`${sectionSelector} ${this.workloadFormRunaiGeneralSearch} input`).fill(name);
    }

    async selectEnvironment(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.environmentSection, name);
        return this.page.locator(this.getCardSelector(name)).click();
    }

    async selectComputeResource(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.computeResourceSection, name);
        const computeCard = this.page.locator(this.getCardSelector(name));
        const isSelected = await computeCard.locator('.is-checked').count();
        if (!isSelected) {
            await computeCard.click();
        }
    }

    async selectDataSourceName(name: string): Promise<void> {
        await this.page.locator(this.dataSourceExpansion).click();
        await this.fillWorkloadFormGeneralSearchInput(this.dataSourceSection, name);
        const dataSourceCard = this.page.locator(this.getCardSelector(name));
        const isSelected = await dataSourceCard.locator('.is-checked').count();
        if (!isSelected) {
            await dataSourceCard.click();
        }
    }

    async clickOnCopyTemplate(): Promise<void> {
        await this.page.locator(this.copyTemplateBtn).click();
    }

    async createWorkloadTemplateUI({
        scopeType,
        scopeEntityName,
        templateName,
        environment = 'tensorboard-tensorflow',
        computeResource = 'one-gpu',
        dataSourceName
    }: {
        scopeType: Scope;
        scopeEntityName: string;
        templateName: string;
        environment?: string;
        computeResource?: string;
        dataSourceName?: string;
    }): Promise<void> {
        await this.navigateToTemplates();
        await this.openNewTemplatePage();
        await this.assetsScopeSection.selectScope({
            scope: scopeType,
            scopeEntityName: scopeEntityName
        });
        await this.fillTemplateName(templateName);
        await this.selectEnvironment(environment);
        await this.selectComputeResource(computeResource);
        if (dataSourceName && dataSourceName.trim() !== '') {
            await this.selectDataSourceName(dataSourceName);
        }
        await this.clickOnCreateTemplateBtn();
        logger.info(`successfully created workspace template: ${templateName}`);
    }
}
