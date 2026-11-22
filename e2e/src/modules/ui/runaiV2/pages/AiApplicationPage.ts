/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Locator, Page } from 'playwright';
import { QTable } from '../components/QTable';
import { MENU_SUBJECT_NAMES, MENU_SUB_ITEMS } from '../../utils/sidenavUtils';
import { logger } from '../../../helpers';

export enum AiApplicationSource {
    NgcCatalog = 'NGC catalog',
    CustomUrl = 'Custom URL',
    UploadFile = 'Upload file'
}

export class AiApplicationPage extends BaseV2Page {
    // table page elements
    newAIpplicationBtn: string;

    clusterFilterChip: string;

    refreshBtn: string;

    // delete elements
    deleteAiApplicationBtn: string;

    deleteModalDeleteBtn: string;

    deleteModalCancelBtn: string;

    closeActionsBtn: string;

    // create form elements
    clusterSelectionSection: string;

    clusterDropdown: string;

    projectSection: string;

    addNewProjectBtn: string;

    projectSearchInput: string;

    workloadNameSection: string;

    aiApplicationNameInput: string;

    sourceSection: string;

    chartFileInput: string;

    cancelBtn: string;

    createBtn: string;

    toggleExpandBtn: string;

    yamlOverridesTexbox: string;

    qTable: QTable;

    constructor(page: Page) {
        super(page);

        this.mainPageTab = '[aid="menu-icon-ai-applications"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.AiApplications}"]`;

        // table page
        this.newAIpplicationBtn = '[aid="primary-cta-btn"]';
        this.clusterFilterChip = '[aid="cluster-filter-chip"]';
        this.refreshBtn = '[aid="refresh-data-button"]';

        // delete
        this.deleteAiApplicationBtn = '[aid="delete-ai-application-btn"]';
        this.deleteModalDeleteBtn = '[aid="delete-modal-delete-button"]';
        this.deleteModalCancelBtn = '[aid="delete-modal-cancel-button"]';
        this.closeActionsBtn = '[aid="close-actions-btn"]';

        // create form
        this.clusterSelectionSection = '[aid="cluster-selection-section"]';
        this.clusterDropdown = '[aid="cluster-dropdown"]';
        this.projectSection = '[aid="project-section"]';
        this.addNewProjectBtn = '[aid="add-new-project-btn"]';
        this.projectSearchInput = '[aid="project-search-card-input"]';
        this.workloadNameSection = '[aid="workload-name-section"]';
        this.aiApplicationNameInput = '[aid="AI application-name-input"]';
        this.sourceSection = '[aid="source-expansion-toggle"]';
        this.chartFileInput = '[aid="chart-file-input"]';
        this.cancelBtn = '[aid="form-cancel-ai-application-btn"]';
        this.createBtn = '[aid="form-create-ai-application-btn"]';
        this.toggleExpandBtn = '[aid="toggle-expand-btn"]';
        this.yamlOverridesTexbox = '.runai-editor-wrapper [role="textbox"]';

        this.qTable = new QTable(this.page);
    }

    // --- Table page actions ---

    async clickNewAiApplicationBtn(): Promise<void> {
        await this.page.locator(this.newAIpplicationBtn).click();
    }

    async clickRefreshBtn(): Promise<void> {
        await this.page.locator(this.refreshBtn).click();
    }

    async clickShowDetailsBtn(): Promise<void> {
        await this.page
            .locator('.selected-rows-actions')
            .getByRole('button', { name: 'show details' })
            .click();
    }

    getDetailsPane(): Locator {
        return this.page.locator('section.application-details');
    }

    getWorkloadCards(): Locator {
        return this.page.locator('.resource-card.workload-card');
    }

    getWorkloadCardStatus(card: Locator): Locator {
        return card.locator('.workload-status-dot .label');
    }

    async getAiApplicationsFromGrid(): Promise<Array<Record<string, string>>> {
        const data = await this.qTable.getTableData();
        logger.info(`table data: \n${JSON.stringify(data)}`);
        return data;
    }

    async getAiApplicationByNameFromGrid(name: string): Promise<Record<string, string>> {
        const data = await this.qTable.getEntityData('AI application', name);
        logger.info(`table data: \n${JSON.stringify(data)}`);
        return data;
    }

    async toggleAiApplication(name: string): Promise<void> {
        await this.qTable.toggleEntity(name);
    }

    getDeleteBtn(): Locator {
        return this.page.locator(this.deleteAiApplicationBtn);
    }

    // --- Create form actions ---

    async fillAiApplicationName(name: string): Promise<void> {
        await this.page.locator(this.aiApplicationNameInput).fill(name);
    }

    async selectProject(projectName: string): Promise<void> {
        await this.page.locator(`[aid="${projectName}-card"]`).click();
    }

    async clickAddNewProjectBtn(): Promise<void> {
        await this.page.locator(this.addNewProjectBtn).click();
    }

    async selectSource(source: AiApplicationSource): Promise<void> {
        await this.page.getByRole('radio', { name: new RegExp(source) }).click();
    }

    async selectRepository(repositoryName: string): Promise<void> {
        await this.page.getByRole('combobox', { name: 'Repository' }).click();
        await this.page.getByRole('option', { name: repositoryName }).click();
    }

    async selectChartName(chartName: string): Promise<void> {
        await this.page.getByRole('combobox', { name: 'Chart name' }).click();
        await this.page.getByRole('option', { name: chartName }).click();
    }

    async selectChartVersion(version: string): Promise<void> {
        await this.page.getByRole('combobox', { name: 'Chart version' }).click();
        await this.page.getByRole('option', { name: version }).click();
    }

    async fillChartUrl(url: string): Promise<void> {
        await this.page.getByRole('textbox', { name: 'Chart URL' }).fill(url);
    }

    async uploadChartFile(filePath: string): Promise<void> {
        const [fileChooser] = await Promise.all([
            this.page.waitForEvent('filechooser'),
            this.page.locator(this.chartFileInput).click()
        ]);
        const [response] = await Promise.all([
            this.page.waitForResponse(
                (res) =>
                    res.url().includes('/api/v1/ai-applications/charts') && res.request().method() === 'POST'
            ),
            fileChooser.setFiles(filePath)
        ]);
        if (!response.ok()) {
            throw new Error(`Chart upload failed with status ${response.status()}`);
        }
    }

    async fillOverrides(yamlContent: string): Promise<void> {
        await this.page.locator(this.yamlOverridesTexbox).fill(yamlContent);
    }

    async clickCancelBtn(): Promise<void> {
        await this.page.locator(this.cancelBtn).click();
    }

    async clickCreateBtn(): Promise<void> {
        await this.page.locator(this.createBtn).click();
    }

    async selectCluster(clusterName: string): Promise<void> {
        await this.page.locator(this.clusterDropdown).click();
        await this.page.getByRole('option', { name: clusterName }).click();
    }

    // --- Composite actions ---

    async createAiApplicationFromNgc(
        name: string,
        projectName: string,
        repositoryName: string,
        chartName: string,
        chartVersion: string,
        overrides?: string
    ): Promise<void> {
        await this.clickNewAiApplicationBtn();
        await this.page.waitForLoadState('load');
        await this.selectProject(projectName);
        await this.fillAiApplicationName(name);
        await this.selectSource(AiApplicationSource.NgcCatalog);
        await this.selectRepository(repositoryName);
        await this.selectChartName(chartName);
        await this.selectChartVersion(chartVersion);
        if (overrides) {
            await this.fillOverrides(overrides);
        }
        await this.clickCreateBtn();
    }

    async createAiApplicationFromCustomUrl(
        name: string,
        projectName: string,
        chartUrl: string,
        overrides?: string
    ): Promise<void> {
        await this.clickNewAiApplicationBtn();
        await this.page.waitForLoadState('load');
        await this.selectProject(projectName);
        await this.fillAiApplicationName(name);
        await this.selectSource(AiApplicationSource.CustomUrl);
        await this.fillChartUrl(chartUrl);
        if (overrides) {
            await this.fillOverrides(overrides);
        }
        await this.clickCreateBtn();
    }

    async createAiApplicationFromUploadFile(
        name: string,
        projectName: string,
        chartFilePath: string,
        overrides?: string
    ): Promise<void> {
        await this.clickNewAiApplicationBtn();
        await this.page.waitForLoadState('load');
        await this.selectProject(projectName);
        await this.fillAiApplicationName(name);
        await this.selectSource(AiApplicationSource.UploadFile);
        await this.uploadChartFile(chartFilePath);
        if (overrides) {
            await this.fillOverrides(overrides);
        }
        await this.clickCreateBtn();
    }

    async deleteAiApplication(name: string): Promise<void> {
        await this.toggleAiApplication(name);
        await this.getDeleteBtn().click();
        await this.page.locator(this.deleteModalDeleteBtn).click();
        await this.waitForAssetToBeDeletedFromUI(this, name, 120);
    }
}
