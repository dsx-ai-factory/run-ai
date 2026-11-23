/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page, Locator } from 'playwright';
import { logger, waitUtils } from '../../../helpers';
import { WorkloadPage } from './WorkloadPage';

export class WorkloadDetailsDrawer extends WorkloadPage {
    detailsDrawerToggleBtn: string;

    hideDetailsBtn: string;

    workloadLogsBody: string;

    podSelector: string;

    logsTabSelector: string;

    detailsTabSelector: string;

    structureTabSelector: string;

    eventHistoryTabSelector: string;

    metricsTabSelector: string;

    generalInfoSectionSelector: string;

    environmentSectionSelector: string;

    computeSectionSelector: string;

    networkSectionSelector: string;

    cliCommandSectionSelector: string;

    expansionHeaderSelector: string;

    expansionContentSelector: string;

    detailsPanelValueCellSelector: string;

    commandLabel: string;

    argumentsLabel: string;

    imageUrlLabel: string;

    clusterLabel: string;

    departmentLabel: string;

    projectLabel: string;

    categoryLabel: string;

    typeLabel: string;

    priorityLabel: string;

    preemptibleLabel: string;

    creationTimeLabel: string;

    createdByLabel: string;

    nodePoolsLabel: string;

    gpuComputeRequestLabel: string;

    cpuComputeRequestLabel: string;

    cpuMemoryRequestLabel: string;

    numberOfPodsLabel: string;

    structureDetailRowSelector: string;

    runningRequestedPodsLabel: string;

    minPodsLabel: string;

    maxPodsLabel: string;

    connectionUrlLabel: string;

    connectionNameLabel: string;

    connectionPortLabel: string;

    connectionToolLabel: string;

    connectionTypeLabel: string;

    connectionAccessLabel: string;

    multiSelectionPromptMessage: string;

    drawerContentSelector: string;

    constructor(page: Page) {
        super(page);
        this.detailsDrawerToggleBtn = '.runai-action-button:has(.ph-file-text)';
        this.hideDetailsBtn = 'button[aid="show-details-button"]:has-text("hide details")';
        this.workloadLogsBody = '.workload-logs-body';
        this.podSelector = 'label.q-select.runai-select';
        this.logsTabSelector = '.q-tab__label:has-text("logs")';
        this.detailsTabSelector = '.q-tab__label:has-text("DETAILS")';
        this.structureTabSelector = '.q-tab__label:has-text("STRUCTURE")';
        this.eventHistoryTabSelector = '.q-tab__label:has-text("EVENT HISTORY")';
        this.metricsTabSelector = '.q-tab__label:has-text("METRICS")';
        this.generalInfoSectionSelector = '[aid="general information-expansion-toggle"]';
        this.environmentSectionSelector = '[aid="environment-expansion-toggle"]';
        this.computeSectionSelector = '[aid="compute-expansion-toggle"]';
        this.networkSectionSelector = '[aid="network-expansion-toggle"]';
        this.cliCommandSectionSelector = '[aid="cli command-expansion-toggle"]';
        this.expansionHeaderSelector = '.runai-expansion-header';
        this.expansionContentSelector = '.q-expansion-item__content';
        this.detailsPanelValueCellSelector = '.value-cell';
        this.commandLabel = 'command';
        this.argumentsLabel = 'arguments';
        this.imageUrlLabel = 'image-url';
        this.clusterLabel = 'cluster';
        this.departmentLabel = 'department';
        this.projectLabel = 'project';
        this.categoryLabel = 'category';
        this.typeLabel = 'type';
        this.priorityLabel = 'priority';
        this.preemptibleLabel = 'preemptible';
        this.creationTimeLabel = 'creation-time';
        this.createdByLabel = 'created-by';
        this.nodePoolsLabel = 'node-pools';
        this.gpuComputeRequestLabel = 'gpu-compute';
        this.cpuComputeRequestLabel = 'cpu-compute';
        this.cpuMemoryRequestLabel = 'cpu-memory';
        this.numberOfPodsLabel = 'number-of-pods';
        this.structureDetailRowSelector = '.detail-row.text-b-body1';
        this.runningRequestedPodsLabel = 'Running / requested pods';
        this.minPodsLabel = 'min-pods';
        this.maxPodsLabel = 'max-pods';
        this.connectionUrlLabel = 'url';
        this.connectionNameLabel = 'name';
        this.connectionPortLabel = 'port';
        this.connectionToolLabel = 'tool';
        this.connectionTypeLabel = 'connection-type';
        this.connectionAccessLabel = 'access';
        this.multiSelectionPromptMessage = 'Select a single workload to view its details';
        this.drawerContentSelector = '.runai-drawer-content';
    }

    async isDrawerVisible(timeout = 3000): Promise<boolean> {
        try {
            await this.page.locator(this.drawerContentSelector).waitFor({ state: 'visible', timeout });
            return true;
        } catch {
            return false;
        }
    }

    async clickShowDetailsBtn(workloadName?: string): Promise<void> {
        const visible = await this.page.locator(this.detailsDrawerToggleBtn).isVisible();
        if (!visible) {
            if (!workloadName) {
                throw new Error('workloadName is required when show details button is not visible');
            }
            await this.toggleWorkload(workloadName);
        }
        if (!(await this.isDrawerVisible())) {
            await this.page.locator(this.detailsDrawerToggleBtn).click();
        }
        logger.info('Clicked show details button');
    }

    async hideDetailsDrawer(): Promise<void> {
        const isVisible = await this.page.locator(this.hideDetailsBtn).isVisible();

        if (isVisible) {
            await this.page.locator(this.hideDetailsBtn).click();
            logger.info('Clicked to hide details drawer');
        } else {
            logger.info('Details drawer is already hidden, no action needed');
        }
    }

    async verifyMultiSelectionPrompt(): Promise<boolean> {
        const promptLocator = this.page.locator(`text=${this.multiSelectionPromptMessage}`);
        const isVisible = await promptLocator.isVisible({ timeout: 5000 });
        if (isVisible) {
            logger.info('Multi-selection prompt message is displayed');
            return true;
        } else {
            const errorMessage = `Multi-selection prompt message "${this.multiSelectionPromptMessage}" is not visible`;
            logger.error(errorMessage);
            throw new Error(errorMessage);
        }
    }

    async clickOnLogsTab(): Promise<void> {
        await this.page.locator(this.logsTabSelector).click();
        await waitUtils.waitForElementToExist(this, this.workloadLogsBody);
    }

    async getLogContent(): Promise<string> {
        await waitUtils.waitForElementToExist(this, this.workloadLogsBody);
        const logElement = this.page.locator(this.workloadLogsBody);
        return (await logElement.textContent()) || '';
    }

    async waitForLogContentToAppear(): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                try {
                    const logContent = await this.getLogContent();
                    const isReady = !logContent.includes('Getting everything ready');
                    return isReady && logContent && logContent.trim() !== '';
                } catch {
                    return false;
                }
            },
            'wait for log content to appear',
            true,
            60
        );
    }

    async selectWorkerPod(podName: string): Promise<void> {
        await this.page.waitForSelector(this.podSelector, { state: 'visible' });
        await this.page.locator(this.podSelector).first().click();
        const podOption = `div.q-menu div.q-item:has-text("${podName}")`;
        await this.page.waitForSelector(podOption, { state: 'visible' });
        await this.page.locator(podOption).click();
        await this.page.waitForSelector('div.q-menu', { state: 'hidden' });
    }

    async isPodSelectorVisible(): Promise<boolean> {
        return await this.page
            .locator(this.podSelector)
            .first()
            .isVisible({ timeout: 3000 })
            .catch(() => false);
    }

    async getPodSelectorOptions(): Promise<string[]> {
        await this.page.waitForSelector(this.podSelector, { state: 'visible' });
        await this.page.locator(this.podSelector).first().click();
        const podOptions = await this.page.locator('div.q-menu div.q-item').allTextContents();

        await this.page.keyboard.press('Escape');
        await this.page.waitForSelector('div.q-menu', { state: 'hidden' });
        logger.info(`Pod selector options: ${podOptions.join(', ')}`);
        return podOptions.map((option) => option.trim());
    }

    async clickOnDetailsTab(): Promise<void> {
        await this.page.locator(this.detailsTabSelector).click();
        await this.page
            .locator(this.generalInfoSectionSelector)
            .waitFor({ state: 'visible', timeout: 10000 });
        await this.ensureSectionExpanded(this.getGeneralInfoSection());
        await this.ensureSectionExpanded(this.getEnvironmentSection());
        await this.ensureSectionExpanded(this.getComputeSection());
        logger.info('Details tab content loaded successfully');
    }

    async clickOnStructureTab(): Promise<void> {
        await this.page.locator(this.structureTabSelector).click();
        await this.page
            .locator(this.structureDetailRowSelector, { hasText: this.runningRequestedPodsLabel })
            .first()
            .waitFor({ state: 'visible', timeout: 10000 });
        logger.info('Structure tab content loaded successfully');
    }

    async clickOnEventHistoryTab(): Promise<void> {
        await this.page.locator(this.eventHistoryTabSelector).click();
        await this.page.waitForTimeout(500);
    }

    async clickOnMetricsTab(): Promise<void> {
        await this.page.locator(this.metricsTabSelector).click();
        await this.page.waitForTimeout(500);
    }

    async getStructureRunningRequestedPods(): Promise<string | undefined> {
        const row = this.page
            .locator(this.structureDetailRowSelector, { hasText: this.runningRequestedPodsLabel })
            .first();
        try {
            const text = ((await row.textContent({ timeout: 5000 })) || '').trim();
            const match = text.match(/Running\s*\/\s*requested pods:\s*(\d+)\s*\/\s*(\d+)/i);
            if (!match) {
                logger.info(`Could not parse Running / requested pods from: "${text}"`);
                return undefined;
            }
            return `${match[1]}/${match[2]}`;
        } catch {
            logger.info('Running / requested pods field not found on Structure tab');
            return undefined;
        }
    }

    async getStructureRequestedPodsCount(): Promise<string | undefined> {
        const ratio = await this.getStructureRunningRequestedPods();
        if (!ratio) {
            return undefined;
        }
        return ratio.split('/')[1];
    }

    private async ensureSectionExpanded(section: Locator, timeout: number = 5000): Promise<void> {
        const sectionClass = (await section.getAttribute('class')) || '';
        if (sectionClass.includes('q-expansion-item--expanded')) {
            return;
        }
        await section.locator(this.expansionHeaderSelector).click();
        await section.locator(this.expansionContentSelector).waitFor({ state: 'visible', timeout });
    }

    private getDetailsPanelRowSelector(rowAid: string): string {
        return `[aid="details-panel-row-${rowAid}"]`;
    }

    private async getRowValue(section: Locator, rowAid: string, timeout: number = 5000): Promise<string> {
        const row = section.locator(this.getDetailsPanelRowSelector(rowAid));
        const value = row.locator(this.detailsPanelValueCellSelector);
        const text = (await value.textContent({ timeout })) || '';
        return text.trim();
    }

    private async getOptionalRowValue(
        section: Locator,
        rowAid: string,
        fieldName: string,
        timeout: number = 3000
    ): Promise<string | undefined> {
        try {
            const row = section.locator(this.getDetailsPanelRowSelector(rowAid));
            await row.waitFor({ state: 'attached', timeout });
            const text = (await row.locator(this.detailsPanelValueCellSelector).textContent()) || '';
            return text.trim();
        } catch {
            logger.info(`${fieldName} field not found`);
            return undefined;
        }
    }

    // General Information section
    private getGeneralInfoSection(): Locator {
        return this.page.locator(this.generalInfoSectionSelector);
    }

    async getClusterName(): Promise<string> {
        return this.getRowValue(this.getGeneralInfoSection(), this.clusterLabel);
    }

    async getDepartment(): Promise<string> {
        return this.getRowValue(this.getGeneralInfoSection(), this.departmentLabel);
    }

    async getProject(): Promise<string> {
        return this.getRowValue(this.getGeneralInfoSection(), this.projectLabel);
    }

    async getCategory(): Promise<string> {
        return this.getRowValue(this.getGeneralInfoSection(), this.categoryLabel);
    }

    async getType(): Promise<string> {
        return this.getRowValue(this.getGeneralInfoSection(), this.typeLabel);
    }

    async getPriority(): Promise<string> {
        return this.getRowValue(this.getGeneralInfoSection(), this.priorityLabel);
    }

    async getPreemptible(): Promise<string> {
        return this.getRowValue(this.getGeneralInfoSection(), this.preemptibleLabel);
    }

    async getCreationTime(): Promise<string> {
        return this.getRowValue(this.getGeneralInfoSection(), this.creationTimeLabel);
    }

    async getCreatedBy(): Promise<string> {
        return this.getRowValue(this.getGeneralInfoSection(), this.createdByLabel);
    }

    // Environment section
    getEnvironmentSection(): Locator {
        return this.page.locator(this.environmentSectionSelector);
    }

    async getImageUrl(): Promise<string> {
        return this.getRowValue(this.getEnvironmentSection(), this.imageUrlLabel);
    }

    async getCommand(): Promise<string | undefined> {
        return this.getOptionalRowValue(this.getEnvironmentSection(), this.commandLabel, 'Command');
    }

    async getArguments(): Promise<string | undefined> {
        return this.getOptionalRowValue(this.getEnvironmentSection(), this.argumentsLabel, 'Arguments');
    }

    // Compute section
    private getComputeSection(): Locator {
        return this.page.locator(this.computeSectionSelector);
    }

    async getNodePools(): Promise<string> {
        return this.getRowValue(this.getComputeSection(), this.nodePoolsLabel);
    }

    async getGpuComputeRequest(): Promise<string | undefined> {
        return this.getOptionalRowValue(
            this.getComputeSection(),
            this.gpuComputeRequestLabel,
            'GPU compute request'
        );
    }

    async getCpuComputeRequest(): Promise<string | undefined> {
        return this.getOptionalRowValue(
            this.getComputeSection(),
            this.cpuComputeRequestLabel,
            'CPU compute request'
        );
    }

    async getCpuMemoryRequest(): Promise<string | undefined> {
        return this.getOptionalRowValue(this.getComputeSection(), this.cpuMemoryRequestLabel, 'CPU memory');
    }

    async getNumberOfPods(): Promise<string | undefined> {
        return this.getOptionalRowValue(this.getComputeSection(), this.numberOfPodsLabel, 'Number of pods');
    }

    async getMinPods(): Promise<string | undefined> {
        return this.getOptionalRowValue(this.getComputeSection(), this.minPodsLabel, 'Min pods');
    }

    async getMaxPods(): Promise<string | undefined> {
        return this.getOptionalRowValue(this.getComputeSection(), this.maxPodsLabel, 'Max pods');
    }

    // CLI command section
    private getCliCommandSection(): Locator {
        return this.page.locator(this.cliCommandSectionSelector);
    }

    async getCliCommand(): Promise<string> {
        const section = this.getCliCommandSection();
        const codeBlock = section.locator('pre, code').first();
        try {
            const commandText = await codeBlock.textContent({ timeout: 3000 });
            if (commandText && commandText.trim()) {
                return commandText.trim();
            }
        } catch {
            logger.debug('CLI command not found in code block');
        }
        logger.info('No CLI command found - workload may not have been submitted via CLI');
        return '';
    }

    async isCliCommandAvailable(): Promise<boolean> {
        const section = this.getCliCommandSection();
        const codeBlock = section.locator('pre, code').first();
        try {
            const isVisible = await codeBlock.isVisible({ timeout: 3000 });
            if (isVisible) {
                const text = await codeBlock.textContent();
                if (text && text.includes('runai')) {
                    logger.debug('CLI command found and available');
                    return true;
                }
            }
            logger.debug('CLI command element found but no runai command inside');
            return false;
        } catch {
            logger.debug('CLI command not found');
            return false;
        }
    }

    // Network section
    getNetworkSection(): Locator {
        return this.page.locator(this.networkSectionSelector);
    }

    async getConnectionUrl(): Promise<string | undefined> {
        return this.getOptionalRowValue(this.getNetworkSection(), this.connectionUrlLabel, 'Connection URL');
    }

    async getConnectionName(): Promise<string | undefined> {
        return this.getOptionalRowValue(
            this.getNetworkSection(),
            this.connectionNameLabel,
            'Connection name'
        );
    }

    async getConnectionPort(): Promise<string | undefined> {
        return this.getOptionalRowValue(
            this.getNetworkSection(),
            this.connectionPortLabel,
            'Connection port'
        );
    }

    async getConnectionTool(): Promise<string | undefined> {
        return this.getOptionalRowValue(
            this.getNetworkSection(),
            this.connectionToolLabel,
            'Connection tool'
        );
    }

    async getConnectionType(): Promise<string | undefined> {
        return this.getOptionalRowValue(
            this.getNetworkSection(),
            this.connectionTypeLabel,
            'Connection type'
        );
    }

    async getConnectionAccess(): Promise<string | undefined> {
        return this.getOptionalRowValue(
            this.getNetworkSection(),
            this.connectionAccessLabel,
            'Connection access'
        );
    }

    // Get all general information at once
    async getGeneralInformation(): Promise<{
        cluster: string;
        department: string;
        project: string;
        category: string;
        type: string;
        priority: string;
        preemptible: string;
        creationTime: string;
        createdBy: string;
    }> {
        return {
            cluster: await this.getClusterName(),
            department: await this.getDepartment(),
            project: await this.getProject(),
            category: await this.getCategory(),
            type: await this.getType(),
            priority: await this.getPriority(),
            preemptible: await this.getPreemptible(),
            creationTime: await this.getCreationTime(),
            createdBy: await this.getCreatedBy()
        };
    }

    // Get all environment information at once
    async getEnvironmentInfo(): Promise<{
        imageUrl: string;
        command?: string;
        arguments?: string;
    }> {
        return {
            imageUrl: await this.getImageUrl(),
            command: await this.getCommand(),
            arguments: await this.getArguments()
        };
    }

    // Get all compute information at once
    async getComputeInfo(options?: {
        gpuResources?: boolean;
        cpuResources?: boolean;
        pods?: boolean;
    }): Promise<{
        nodePools: string;
        gpuComputeRequest?: string;
        cpuComputeRequest?: string;
        cpuMemoryRequest?: string;
        numberOfPods?: string;
    }> {
        const { gpuResources = true, cpuResources = false, pods = false } = options || {};

        const computeInfo: {
            nodePools: string;
            gpuComputeRequest?: string;
            cpuComputeRequest?: string;
            cpuMemoryRequest?: string;
            numberOfPods?: string;
        } = {
            nodePools: await this.getNodePools()
        };

        if (gpuResources) {
            computeInfo.gpuComputeRequest = await this.getGpuComputeRequest();
        }

        if (cpuResources) {
            computeInfo.cpuComputeRequest = await this.getCpuComputeRequest();
            computeInfo.cpuMemoryRequest = await this.getCpuMemoryRequest();
        }

        if (pods) {
            computeInfo.numberOfPods = await this.getNumberOfPods();
        }

        logger.info(`Retrieved compute info: ${JSON.stringify(computeInfo)}`);
        return computeInfo;
    }

    // get all network section information
    async getNetworkSectionInfo(): Promise<{
        isVisible: boolean;
        connections?: { url: string; name: string; port: string };
    }> {
        try {
            const section = this.getNetworkSection();
            const isVisible = await section.isVisible({ timeout: 3000 });

            if (!isVisible) {
                logger.info('Network section is not visible');
                return { isVisible: false };
            }
            await this.ensureSectionExpanded(section);
            logger.debug('Network section is visible, checking for connection details');
            try {
                const url = await this.getConnectionUrl();
                const name = await this.getConnectionName();
                const port = await this.getConnectionPort();
                logger.info(`Network connections found - URL: ${url}, Name: ${name}, Port: ${port}`);
                return {
                    isVisible: true,
                    connections: {
                        url: url || '',
                        name: name || '',
                        port: port || ''
                    }
                };
            } catch {
                logger.debug('Network section visible but no connection details found');
                return {
                    isVisible: true,
                    connections: {
                        url: '',
                        name: '',
                        port: ''
                    }
                };
            }
        } catch (e) {
            logger.debug(`Network section not found: ${e.message}`);
            return { isVisible: false };
        }
    }

    // Get all CLI command section information
    async getCliCommandSectionInfo(): Promise<{
        isVisible: boolean;
        hasInfoIcon?: boolean;
        messageText?: string;
        docLink?: { text: string; href: string };
        cliCommand?: string;
    }> {
        try {
            const section = this.getCliCommandSection();
            await section.waitFor({ state: 'visible', timeout: 5000 });
            await this.ensureSectionExpanded(section);
            logger.debug('CLI command section is visible, checking details');

            const infoIcon = section.locator('.ph-question').first();
            const hasInfoIcon = await infoIcon.isVisible().catch(() => false);

            const messageText = (await section.textContent()) || '';
            let docLink: { text: string; href: string } | undefined;
            const linkElement = section.locator('a:has-text("CLI Commands Reference")').first();
            try {
                const linkText = (await linkElement.textContent({ timeout: 2000 })) || '';
                const href = (await linkElement.getAttribute('href')) || '';
                if (linkText && href) {
                    docLink = { text: linkText.trim(), href };
                }
            } catch {
                logger.debug('Documentation link not found');
            }
            const cliCommand = await this.getCliCommand();
            return {
                isVisible: true,
                hasInfoIcon,
                messageText: messageText.trim(),
                docLink,
                cliCommand
            };
        } catch (e) {
            logger.debug(`CLI command section not found: ${e.message}`);
            return { isVisible: false };
        }
    }
}
