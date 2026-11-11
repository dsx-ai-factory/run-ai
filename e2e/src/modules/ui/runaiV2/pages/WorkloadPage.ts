/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Locator, Page } from 'playwright';
import { Utils, logger, waitUtils, settings, flowRootState, apiUtils } from '../../../helpers';
import { Phase } from '../../../api/apis/workloadsApi/generated/model/phase';
import { pageFilterUtils } from '../../utils/pageFilterUtils';
import { MENU_SUBJECT_NAMES, MENU_SUB_ITEMS } from '../../utils/sidenavUtils';
import { Scope, WorkloadType } from '../../../api/apis/assetsApi';
import { AssetsScopeSection } from '../components/AssetsScopeSection';
import { EFeatureFlagsKeys, EFeatureFlagsNames } from '../../../helpers/models/featureFlags';

export enum EWorkloadStatus {
    Running = 'Running',
    Stopped = 'Stopped',
    Completed = 'Completed',
    Failed = 'Failed',
    Pending = 'Pending',
    Deleting = 'Deleting'
}

export enum EToolName {
    Jupyter = 'jupyter',
    TensorBoard = 'tensorboard',

    VSCode = 'vscode'
}

export class WorkloadPage extends BaseV2Page {
    //table elements

    newWorkspaceBtn: string;

    newTrainingBtn: string;

    newInferenceBtn: string;

    addNewWorkloadBtn: string;

    stopWorkloadBtn: string;

    workloadNameCol: string;

    copyAndWorkloadEditBtn: string;

    updateWorkloadEditBtn: string;

    requestedGpuColumn: string;

    deleteWorkloadActionBtn: string;

    deleteWorkloadModalBtn: string;

    connectWorkloadButton: string;

    activateWorkloadBtn: string;

    statusCol: string;

    skeletonCardsList: string;

    runaiDefaultLoader: string;

    workloadFormProjectSection: string;

    workloadFormTemplateSection: string;

    workloadFormTemplateV2Section: string;

    workloadFormTemplateSectionV2: string;

    workloadFormTemplateStartFromScratchCard: string;

    workloadFormRunaiGeneralSearch: string;

    clusterDropDown: string;

    environmentSection: string;

    computeResourceSection: string;

    dataSourceSection: string;

    dataSourceExpansion: string;

    selectedRowActionsContainer: string;

    typeCol: string;

    projectCol: string;

    createdByCol: string;

    podsCol: string;

    selectColumnsButton: string;

    columnsMenu: string;

    allColumnsCheckbox: string;

    bodySelector: string;

    advancedFilterSection: string;

    advancedFilterRemoveButton: string;

    cancelBtn: string;

    elemChecked: string;

    tensorBoardOption: string;

    jupyterOption: string;

    envVarNameInput: string;

    envVarSourceSelect: string;

    envVarValue: string;

    envVarCredentialSelect: string;

    envVarKeySelect: string;

    envVarConfigMapNameSelect: string;

    envVarConfigMapKeyInput: string;

    addEnvVarBtn: string;

    generalExpansionLabel: string;

    envVarBox: string;

    detailsTab: string;

    detailsGeneralInformationExpansionLabel: string;

    detailsEnvironmentExpansionLabel: string;

    detailsComputeExpansionLabel: string;

    eventHistoryTab: string;

    eventHistoryContent: string;

    envImageInput: string;

    runtimeSettingsSection: string;

    runtimeCommandInput: string;

    gpuDeviceInput: string;

    aiApplicationCol: string;

    filterValueInput: string;

    filterSelect: string;

    submitFilterBtn: string;

    metricTab: string;

    metricTypeSelect: string;

    metricPodSelect: string;

    nimMetricOption: string;

    nimWorkloadWidgets: string;

    nimMetricCharts: string[];

    constructor(page: Page) {
        super(page);
        this.mainPageTab = '[aid="menu-icon-workloads"]';

        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Workloads}"]`;
        this.newWorkspaceBtn = '[aid="Workspace-btn"]';
        this.newTrainingBtn = '[aid="Training-btn"]';
        this.newInferenceBtn = '[aid="Inference-btn"]';
        this.addNewWorkloadBtn = '[aid="primary-cta-btn"] >> nth=0';
        this.workloadNameCol = '[aid="name"]';
        this.copyAndWorkloadEditBtn = '[aid="copy-workload-btn"]';
        this.updateWorkloadEditBtn = '[aid="update-workload-btn"]';
        this.stopWorkloadBtn = '[aid="stop-workload-btn"]';
        this.activateWorkloadBtn = '[aid="activate-workload-btn"]';
        this.requestedGpuColumn = '[aid="requestedGPU"]';
        this.deleteWorkloadActionBtn = '[aid="delete-workload-btn"]';
        this.deleteWorkloadModalBtn = '[aid="delete-modal-delete-button"]';
        this.statusCol = '[aid="phase"]';
        this.connectWorkloadButton = '[aid="connect-workload-btn"]';
        this.skeletonCardsList = '[aid="skeleton-list"]';
        this.runaiDefaultLoader = '.runai-default-loader';
        this.workloadFormProjectSection = '[aid="project-section"]';
        this.workloadFormTemplateSection = '[aid="template-section"]';
        this.workloadFormTemplateV2Section = '[aid="template-section-v2"]';
        this.workloadFormTemplateStartFromScratchCard = '[aid="create-new-card"]';
        this.workloadFormRunaiGeneralSearch = '.runai-general-search';
        this.clusterDropDown = '[aid="cluster-dropdown"]';
        this.environmentSection = '.environment-section';
        this.computeResourceSection = '.compute-resource-section';
        this.dataSourceSection = '.data-source-section';
        this.selectedRowActionsContainer = '.selected-rows-actions';
        this.dataSourceExpansion = '[aid="data-source-section"]';
        this.typeCol = '[aid="type"]';
        this.projectCol = '[aid="projectName"]';
        this.aiApplicationCol = '[aid="aiApplicationName"]';
        this.createdByCol = '[aid="submittedBy"]';
        this.podsCol = '[aid="runningPodsVsRequestedPods"]';
        this.selectColumnsButton = 'button[aid="run-page-action-select-columns-btn"]';
        this.columnsMenu = 'div[role="menu"]';
        this.allColumnsCheckbox = 'div[role="checkbox"][aria-label="All"]';
        this.bodySelector = 'body';
        this.advancedFilterSection = 'section.runai-page-filter.advanced-filters';
        this.advancedFilterRemoveButton = 'i[aria-label="Remove"]';
        this.cancelBtn = '[aid="cancel-btn"]';
        this.elemChecked = 'aria-checked';
        this.tensorBoardOption = '[aid="tensorboard-option"]';
        this.jupyterOption = '[aid="jupyter-option"]';

        this.envVarNameInput = '[aid="environment-variable-name"]';
        this.envVarSourceSelect = '[aid="environment-variable-source"]';
        this.envVarValue = '[aid="environment-variable-value"]';
        this.addEnvVarBtn = '[aid="add-environment-variable-button"]';
        this.envVarCredentialSelect = '[aid="environment-variable-credential"]';
        this.envVarKeySelect = '[aid="environment-variable-key-select"]';
        this.envVarConfigMapNameSelect = '[aid="environment-variable-config-map-name-select"]';
        this.envVarConfigMapKeyInput = '[aid="environment-variable-config-map-key-input"]';
        this.generalExpansionLabel = '[aid="general-expansion-label"]';
        this.envVarBox = '.environment-variable-box';
        this.detailsTab = '.q-tab__label:has-text("DETAILS")';
        this.detailsGeneralInformationExpansionLabel = '[aid="general information-expansion-label"]';
        this.detailsEnvironmentExpansionLabel = '[aid="environment-expansion-label"]';
        this.detailsComputeExpansionLabel = '[aid="compute-expansion-label"]';
        this.eventHistoryTab = '.q-tab__label:has-text("EVENT HISTORY")';
        this.eventHistoryContent = '.runai-event-history';
        this.envImageInput = '[aid="environment-image-input"]';
        this.runtimeSettingsSection = '[aid="runtime-settings-sub-section"]';
        this.runtimeCommandInput = '[aid="command-input"]';
        this.gpuDeviceInput = '[aid="gpu-devices-input"]';
        this.workloadFormTemplateSectionV2 = '[aid="template-section-v2"]';
        this.filterValueInput = '[aid="free-text-filter-input"]';
        this.submitFilterBtn = '[aid="submit-filter-btn"]';
        this.filterSelect = '[aid="free-text-filter-select"]';
        this.metricTab = '.q-tab__label:has-text("Metrics")';
        this.metricTypeSelect = '[aid="metric-type-select"]';
        this.metricPodSelect = '[aid="metric-pod-select"]';
        this.nimMetricOption = '.runai-select-item:has-text("NVIDIA NIM")';
        this.nimWorkloadWidgets = '.nim-workload-widgets';
        this.nimMetricCharts = [
            '[aid="request-concurrency-by-status"]',
            '[aid="request-count-by-status"]',
            '[aid="time-to-first-token-(ttft)"]',
            '[aid="time-to-first-token-(ttft)-by-percentiles"]',
            '[aid="end-to-end-request-latency"]',
            '[aid="end-to-end-request-latency-by-percentiles"]',
            '[aid="gpu-kv-cache-utilization"]'
        ];
    }

    async navigateToWorkloads(): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.subjectItem);

        await this.page.locator(this.subjectItem).click();
        await this.page.locator(this.subMenuItem).click();

        //a workaround to close the side menu
        await this.page.locator(this.mainPageTitle).click();
        await pageFilterUtils.selectCluster(this.page);
    }

    async openWorkloadForm(type: WorkloadType): Promise<void> {
        switch (type) {
            case WorkloadType.Training:
                return this.openTrainingForm();
            case WorkloadType.Workspace:
                return this.openWorkspaceForm();
            case WorkloadType.Inference:
                return this.openInferenceForm();
        }
    }

    async clickOnNewWorkloadButton(): Promise<void> {
        await this.page.locator(this.addNewWorkloadBtn).click();
    }

    // this function assumes the dropdown with the option is already open (at the moment only used for the guided tour context where each step is handled)
    async clickOnNewWorkspaceOption(): Promise<void> {
        await this.page.locator(this.newWorkspaceBtn).click();
    }

    async openWorkspaceForm(): Promise<void> {
        await this.clickOnNewWorkloadButton();
        await this.page.locator(this.newWorkspaceBtn).click();
    }

    async openTrainingForm(): Promise<void> {
        await this.clickOnNewWorkloadButton();
        await this.page.locator(this.newTrainingBtn).click();
    }

    async openInferenceForm(): Promise<void> {
        await this.clickOnNewWorkloadButton();
        const newInferenceBtn = this.page.locator(this.newInferenceBtn);
        const classList = await newInferenceBtn.getAttribute('class');
        const isDisabled = classList?.includes('disabled');
        if (isDisabled) {
            logger.error('Create new inference button is disabled');
            throw new Error('Create new inference button is disabled');
        }

        await this.page.locator(this.newInferenceBtn).click();
    }

    async getWorkloadNames(): Promise<string[]> {
        return await this.page.locator(this.workloadNameCol).allTextContents();
    }

    async getWorkloadAiApplications(): Promise<string[]> {
        return await this.page.locator(this.aiApplicationCol).allTextContents();
    }

    async addFilterOfContainandEquals(
        filterName: string,
        filterType: string,
        filterValue: string
    ): Promise<void> {
        await this.page.locator(this.addFilterBtn).click();
        await this.page
            .getByRole('menuitem')
            .filter({ has: this.page.getByText(filterName, { exact: true }) })
            .click();
        await this.page.locator(this.filterSelect).click();
        if (filterType === 'contains') {
            await this.page.getByRole('option').filter({ hasText: 'contains' }).click();
        } else if (filterType === 'not_contain') {
            await this.page.getByRole('option').filter({ hasText: 'does not contain' }).click();
        } else if (filterType === 'equals') {
            await this.page.getByRole('option').filter({ hasText: 'equals' }).click();
        } else if (filterType === 'not_equal') {
            await this.page.getByRole('option').filter({ hasText: 'does not equal' }).click();
        } else {
            throw new Error(
                `Unsupported filter type: ${filterType}. Supported types are: contains, not_contain, equals, not_equal`
            );
        }
        await this.page.locator(this.filterValueInput).click();
        await this.page.getByRole('option', { name: filterValue }).click();
        await this.page.locator(this.submitFilterBtn).click();
        await this.page.waitForTimeout(1000);
    }

    async getWorkloadRowIndexByName(workloadName: string): Promise<number> {
        const workloadNames: string[] = await this.getWorkloadNames();
        console.log(workloadNames);
        const index = workloadNames.findIndex((name: string) => name.trim() === workloadName);
        if (index === -1) {
            throw new Error(`row with ${workloadName} text not found`);
        } else {
            return index;
        }
    }

    async isColumnContentCorrect(
        columnSelector: string,
        rowIndex: number,
        desiredResult: string
    ): Promise<void> {
        await waitUtils.waitForCondition(async () => {
            const allCells = await this.page.locator(columnSelector).allTextContents();
            if (rowIndex < allCells.length) {
                let cellValue = allCells[rowIndex];
                if (columnSelector === this.statusCol) {
                    cellValue = cellValue.trim().split(' ')[0];
                }
                return cellValue === desiredResult;
            } else {
                return false;
            }
        }, '');
    }

    async isRequestedGpuCorrect(rowIndex: number, desiredResult: string): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.requestedGpuColumn);
        return await this.isColumnContentCorrect(this.requestedGpuColumn, rowIndex, desiredResult);
    }

    async isTypeCorrect(rowIndex: number, desiredResult: WorkloadType): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.typeCol);
        return await this.isColumnContentCorrect(this.typeCol, rowIndex, desiredResult.toString());
    }

    async isStatusCorrect(rowIndex: number, desiredResult: string): Promise<void> {
        const allCells = await this.page.locator(this.statusCol).allTextContents();
        return await this.isColumnContentCorrect(this.statusCol, rowIndex, desiredResult);
    }

    async isProjectCorrect(rowIndex: number, desiredResult: string): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.projectCol);
        return await this.isColumnContentCorrect(this.projectCol, rowIndex, desiredResult);
    }

    async isCreatedByCorrect(rowIndex: number, desiredResult: string): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.createdByCol);
        return await this.isColumnContentCorrect(this.createdByCol, rowIndex, desiredResult);
    }

    async isPodsCorrect(rowIndex: number, desiredResult: string): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.podsCol);
        return await this.isColumnContentCorrect(this.podsCol, rowIndex, desiredResult);
    }

    async clickOnCopyAndEditWorkload(): Promise<void> {
        await this.page.locator(this.copyAndWorkloadEditBtn).click();
    }

    async clickOnUpdateWorkload(): Promise<void> {
        await this.page.locator(this.updateWorkloadEditBtn).click();
    }

    async clickOnStopWorkload(): Promise<void> {
        await this.page.locator(this.stopWorkloadBtn).click();
    }

    async clickOnDeleteWorkloadActionBtn(): Promise<void> {
        await this.page.locator(this.deleteWorkloadActionBtn).click();
    }

    async clickOnDeleteWorkloadModalBtn(): Promise<void> {
        await this.page.locator(this.deleteWorkloadModalBtn).click();
    }

    async clearGridSearchInput(): Promise<void> {
        await this.searchByName('');
    }

    async toggleWorkload(workloadName: string): Promise<void> {
        const visible = await this.isElementVisible(`td${this.workloadNameCol}:has-text("${workloadName}")`);
        if (!visible) await this.searchByName(workloadName);
        await this.clickOnElementByText(`td${this.workloadNameCol}`, `${workloadName}`);
    }

    /**
     * Waits for a successful (200) `/workloads/count` response, optionally requiring the URL to
     * contain a specific substring (typically the workload name being searched for). Logs every
     * `/workloads/count` response observed during the wait so timeouts produce actionable triage
     * data.
     */
    async waitForWorkloadsCountResponse({
        nameInUrl,
        timeoutMs
    }: { nameInUrl?: string; timeoutMs?: number } = {}): Promise<void> {
        await this.waitForResponseWithDiagnostics({
            urlContains: '/workloads/count',
            label: 'workloads-count',
            predicate: (response) =>
                response.url().includes('/workloads/count') &&
                response.status() === 200 &&
                (nameInUrl === undefined || response.url().includes(nameInUrl)),
            timeoutMs
        });
    }

    async deleteWorkload(workloadName: string): Promise<void> {
        const visible = await this.isElementVisible(this.deleteWorkloadActionBtn);
        if (!visible) {
            // Arm the count-response waiter and trigger the search together via Promise.all:
            // - the waiter is the FIRST array element so its response listener registers before
            //   the search request is fired (no register-after-trigger race);
            // - it's scoped by `workloadName` so an unrelated/stale `/workloads/count` cannot
            //   falsely satisfy the predicate;
            // - Promise.all attaches internal handlers to both promises synchronously, so if
            //   `searchByName` throws, the waiter's late rejection isn't an unhandled promise.
            await Promise.all([
                this.waitForWorkloadsCountResponse({ nameInUrl: workloadName }),
                this.searchByName(workloadName)
            ]);
            // UI rendering delay
            await this.page.waitForTimeout(200);
            await this.toggleWorkload(workloadName);
        }

        await this.clickOnDeleteWorkloadActionBtn();
        await this.clickOnDeleteWorkloadModalBtn();
        await waitUtils.waitForElementToNotExist(this, `td:has-text("${workloadName}")`, 90);
        await this.clearGridSearchInput();
        logger.info(`successfully deleted workload: ${workloadName}`);
    }

    async deleteWorkloads(workloadNames: string[]): Promise<void> {
        for (const workloadName of workloadNames) {
            await this.selectWithCmd(workloadName);
        }
        await this.clickOnDeleteWorkloadActionBtn();
        await this.clickOnDeleteWorkloadModalBtn();
    }

    async selectWithCmd(workloadName: string): Promise<void> {
        await this.page.keyboard.down('Meta');
        await this.toggleWorkload(workloadName);
        await this.page.keyboard.up('Meta');
    }

    async selectWorkloadsByNames(workloadNames: string[]): Promise<void> {
        logger.info(`Selecting ${workloadNames.length} workloads: ${workloadNames.join(', ')}`);

        for (const workloadName of workloadNames) {
            const row = this.page.locator(`tr:has(td${this.workloadNameCol}:has-text("${workloadName}"))`);
            const checkbox = row.first().getByRole('checkbox');
            await checkbox.click();
            logger.info(`Selected workload: ${workloadName}`);
        }
        logger.info(`Successfully selected ${workloadNames.length} workloads`);
    }

    async toggleWorkloadDrawerContent(workloadName: string): Promise<void> {
        await this.page.locator('[aid="show-details-button"]').click();
        await this.clickOnElementByText('td', ` ${workloadName}`);
        await waitUtils.waitForElementToExist(this, '.runai-drawer-content');
    }

    async closeWorkloadDrawerContent(): Promise<void> {
        await this.page.locator('.drawer-content >> .ph-x').click();
    }

    async validateEventsHistoryTabContentExists(): Promise<void> {
        await waitUtils.waitForElementToExist(this, '[aid="workload-phase-history-chart"]');
        await waitUtils.waitForElementToExist(this, '.runai-event-history');
    }

    async validateMetricTabAllPodsContentExists(): Promise<void> {
        await this.page.locator('.q-tab__label:has-text("Metric")').click();
        await waitUtils.waitForElementToExist(this, '[aid="gpu-compute-utilization"]');
        await waitUtils.waitForElementToExist(this, '[aid="gpu-memory-usage"]');
        await waitUtils.waitForElementToExist(this, '[aid="cpu-usage"]');
        await waitUtils.waitForElementToExist(this, '[aid="cpu-memory-usage"]');
    }

    async validateNimMetricCharts(): Promise<void> {
        for (const chart of this.nimMetricCharts) {
            await waitUtils.waitForCondition(
                async () => {
                    await waitUtils.waitForElementToExist(this, chart);
                    const wrapper = this.page
                        .locator(chart)
                        .locator('xpath=ancestor::div[contains(@class, "chart-widget-wrapper")]')
                        .first();
                    if ((await wrapper.count()) === 0) {
                        logger.info(`${chart}: waiting for data...`);
                        return false;
                    }
                    const seriesGraphics = wrapper.locator('.echarts svg [clip-path] path');
                    if ((await seriesGraphics.count()) > 0) {
                        logger.info(`${chart}: data found ✓`);
                        return true;
                    }
                    logger.info(`${chart}: waiting for data...`);
                    return false;
                },
                `Wait for chart ${chart} to have data`,
                true,
                360,
                10
            );
        }
    }

    async validateNimMetricsForWorkload(): Promise<void> {
        await this.page.locator(this.metricTab).click();
        await waitUtils.waitForElementToExist(this, this.metricTypeSelect);
        await this.page.locator(this.metricTypeSelect).click();
        await waitUtils.waitForElementToExist(this, this.nimMetricOption);
        await this.page.locator(this.nimMetricOption).click();
        await waitUtils.waitForElementToExist(this, this.nimWorkloadWidgets);
        await this.validateNimMetricCharts();
    }

    async validateNimMetricsForEachPod(): Promise<void> {
        await this.page.locator(this.metricPodSelect).click();
        const podOptions = await this.page.locator('[role="listbox"] .runai-select-item').allTextContents();
        logger.info(`Found ${podOptions.length} pod options: ${podOptions.join(', ')}`);
        await this.page.keyboard.press('Escape');

        for (const podOption of podOptions) {
            logger.info(`Validating NIM metrics for pod: ${podOption.trim()}`);
            await this.page.locator(this.metricPodSelect).click();
            await this.page.getByRole('option', { name: podOption.trim(), exact: true }).click();
            await this.page.waitForTimeout(1000);
            await this.validateNimMetricCharts();
        }
    }

    async validateMetricTabSinglePodContentExists(): Promise<void> {
        await this.page.locator('.q-tab__label:has-text("Metric")').click();
        const elements = [
            '[aid="gpu-compute-utilization-per-gpu"]',
            '[aid="gpu-memory-usage-per-gpu"]',
            '[aid="cpu-usage"]',
            '[aid="cpu-memory-usage"]'
        ];
        for (const element of elements) {
            await waitUtils.waitForElementToExist(this, element);
        }
    }

    async validateDetailsTabContentExists(): Promise<void> {
        await this.page.locator(this.detailsTab).click();
        await this.page
            .locator(this.detailsGeneralInformationExpansionLabel)
            .waitFor({ state: 'visible', timeout: 10000 });
        await this.page
            .locator(this.detailsEnvironmentExpansionLabel)
            .waitFor({ state: 'visible', timeout: 10000 });
        await this.page
            .locator(this.detailsComputeExpansionLabel)
            .waitFor({ state: 'visible', timeout: 10000 });
    }

    async validateLogsTabContentExists(): Promise<void> {
        await this.page.locator('.q-tab__label:has-text("logs")').click();
        await waitUtils.waitForElementToExist(this, '.workload-logs-body');
    }

    async validateCurrentPageIsWorkload(timeoutSeconds = 30): Promise<void> {
        await waitUtils.waitForElementToExist(
            this,
            `${this.mainPageTitle}:has-text("Workloads")`,
            timeoutSeconds
        );
    }

    async waitForPhase(workloadIndex: number, phase: Phase): Promise<boolean> {
        const workloadStatusCells: Array<string> = await this.page.locator(this.statusCol).allTextContents();
        return workloadStatusCells[workloadIndex].includes(phase.toString());
    }

    async waitForRunningPhase(workloadIndex: number): Promise<boolean> {
        return await this.waitForPhase(workloadIndex, Phase.Running);
    }

    async waitForStoppedPhase(workloadIndex: number): Promise<boolean> {
        return await this.waitForPhase(workloadIndex, Phase.Stopped);
    }

    async validateRunningPhase(workloadIndex: number): Promise<void> {
        await waitUtils.waitForCondition(
            async () => await this.waitForRunningPhase(workloadIndex),
            'wait for workload phase to be Ready'
        );
    }

    async isTableNotIncludeWorkload(workloadName: string): Promise<boolean> {
        const workloadNames: string[] = await this.getWorkloadNames();
        return !workloadNames.includes(workloadName);
    }

    /**
     * @deprecated use the much more reliable 'waitForWorkloadUIStatus' to verify status in the UI
     */
    async waitForWorkloadStatus(workloadName: string, status = 'Running'): Promise<void> {
        logger.info(`waiting for workload ${workloadName} status to be: ${status}.`);
        await this.v2frame.waitForSelector(
            `td:has-text("${status}"):right-of(td:has-text("${workloadName}"))`,
            { timeout: 180000 }
        );
        logger.info(`workload ${workloadName} status is: ${status}.`);
    }

    async waitForWorkloadUIStatus(
        workloadName: string,
        status = 'Running',
        timeout = 180000,
        withSearch = true
    ): Promise<void> {
        if (withSearch) await this.searchByName(workloadName);

        logger.info(`waiting for workload ${workloadName} status to be: ${status}.`);
        await this.page.waitForSelector(`[aid="${workloadName}-status"]:has-text("${status}")`, {
            timeout: timeout
        });
        if (withSearch) await this.clearGridSearchInput();
        logger.info(`workload ${workloadName} status is: ${status}.`);
    }

    async clickOnConnectWorkloadButton(): Promise<void> {
        return this.page.locator(this.connectWorkloadButton).click();
    }

    async clickOnConnectTool(toolName: string): Promise<void> {
        await this.clickOnConnectWorkloadButton();

        // Wait briefly for potential menu to appear
        await this.page.waitForTimeout(500);

        // Check if the tools menu appeared (multiple tools scenario)
        const toolsMenu = this.page.locator('[aid="tools-select-menu"]');
        const menuExists = await toolsMenu.isVisible().catch(() => false);

        if (menuExists) {
            // Multiple tools scenario - select the specific tool from the menu
            const toolOptionAid = Utils.spaceToDash(toolName).toLowerCase() + '-option';
            await this.clickOnElementByAidName(toolOptionAid);
        }
        // If no menu exists, single tool scenario - tool should have opened directly
    }

    async isElementVisible(query: string): Promise<boolean> {
        return this.page.locator(query).isVisible();
    }

    async stopWorkloadByName(workloadName: string): Promise<void> {
        const visible = await this.isElementVisible(this.stopWorkloadBtn);
        if (!visible) await this.toggleWorkload(workloadName);

        await this.page.locator(this.stopWorkloadBtn).click();
        await this.toggleWorkload(workloadName);
        await this.clearGridSearchInput();
    }

    async activateWorkloadByName(workloadName: string): Promise<void> {
        const visible = await this.isElementVisible(this.activateWorkloadBtn);
        if (!visible) await this.toggleWorkload(workloadName);

        await this.page.locator(this.activateWorkloadBtn).click();
        await this.toggleWorkload(workloadName);
        await this.clearGridSearchInput();
    }

    async connectWorkloadByName(workloadName: string): Promise<void> {
        const visible = await this.isElementVisible(this.connectWorkloadButton);
        if (!visible) await this.toggleWorkload(workloadName);
        await this.page.locator(this.connectWorkloadButton).click();
    }

    async connectWorkloadToolByName(workloadName: string, toolName: EToolName): Promise<void> {
        await this.connectWorkloadByName(workloadName);
        await this.page.waitForTimeout(500);

        const toolsMenu = this.page.locator('[aid="tools-select-menu"]');
        const menuExists = await toolsMenu.isVisible().catch(() => false);

        if (menuExists) {
            logger.info(`Multiple tools available, selecting ${toolName} from menu`);
            const toolOptionAid = Utils.spaceToDash(toolName).toLowerCase() + '-option';
            await this.clickOnElementByAidName(toolOptionAid);
        } else {
            logger.info(`Single tool scenario detected, ${toolName} should have opened directly`);
        }
    }

    async clickOnSubExpansionByLabel(label: string): Promise<void> {
        return this.clickOnElementByText('.runai-sub-expansion-header', label);
    }

    async clickOnTensorBoard(): Promise<void> {
        await this.page.waitForSelector(this.tensorBoardOption, { timeout: 5000 });
        await this.page.locator(this.tensorBoardOption).click();
    }

    async clickOnJupyter(): Promise<void> {
        await this.page.waitForSelector(this.jupyterOption, { timeout: 5000 });
        await this.page.locator(this.jupyterOption).click();
    }

    async validateNewWorkloadButtonExist(): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const workloadBtnCount = await this.v2frame.locator(this.addNewWorkloadBtn).count();
                return workloadBtnCount !== 0;
            },
            'is new workload button exists',
            true,
            10
        );
    }

    async selectClusterFromDropDown(clusterName: string): Promise<void> {
        await this.page.waitForSelector(this.clusterDropDown, {
            timeout: 5000
        });
        await this.page.click(this.clusterDropDown, { force: true, timeout: 5000 });
        const timeout: number = settings.isSelfHosted ? 1500 : 4000;
        await this.page.waitForTimeout(timeout);
        await this.page.click(`.runai-select-item:has-text("${clusterName}")`, {
            force: true,
            timeout: 5000
        });
    }

    async workloadFormSelectProject(name: string, shouldFilter = true): Promise<void> {
        await waitUtils.waitForElementToNotExist(this, this.runaiDefaultLoader, 10000);
        await waitUtils.waitForElementToNotExist(this, this.skeletonCardsList, 10000);

        if (shouldFilter)
            await this.fillWorkloadFormGeneralSearchInput(this.workloadFormProjectSection, name);

        const projectCard = this.v2frame.locator(`[aid=${name}-card]`);
        await this.page.waitForTimeout(1000);
        const isSelected = await projectCard.locator('.is-checked').count();

        if (!isSelected) {
            await projectCard.click();
        }
    }

    async workloadFormSelectTemplate(name = 'create-new'): Promise<void> {
        await waitUtils.waitForElementToNotExist(this, this.runaiDefaultLoader, 10000);
        await waitUtils.waitForElementToNotExist(this, this.skeletonCardsList, 10000);

        const templateSection = (await apiUtils.isTemplateV2Enabled())
            ? this.workloadFormTemplateV2Section
            : this.workloadFormTemplateSection;

        if (name !== 'create-new') {
            // The search cannot find the "start from scratch" option
            await this.fillWorkloadFormGeneralSearchInput(templateSection, name);
        }

        const templateCard = this.v2frame.locator(`[aid=${name}-card]`);
        const isSelected = await this.isWorkloadFormCardSelected(templateCard);
        if (!isSelected) {
            await templateCard.click();
        }
        await this.waitForWorkloadFormCardSelection(templateCard);
    }

    async workloadFormSelectTemplateV2(name = 'create-new', shouldFilter = true): Promise<void> {
        await waitUtils.waitForElementToNotExist(this, this.runaiDefaultLoader, 10000);
        await waitUtils.waitForElementToNotExist(this, this.skeletonCardsList, 10000);

        if (shouldFilter) {
            await this.fillWorkloadFormGeneralSearchInput(this.workloadFormTemplateSectionV2, name);
        }
        await this.page.waitForTimeout(1000);

        const templateCard = this.page.locator(`[aid=${name}-card]`);
        const isSelected = await this.isWorkloadFormCardSelected(templateCard);
        if (!isSelected) {
            await templateCard.click();
        }
        await this.waitForWorkloadFormCardSelection(templateCard);
    }

    async workloadFromScratchTemplateV2(): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.workloadFormTemplateSectionV2, '');
        const templateCard = this.page.locator(this.workloadFormTemplateStartFromScratchCard);
        await templateCard.click();
        await this.waitForWorkloadFormCardSelection(templateCard);
    }

    private async isWorkloadFormCardSelected(templateCard: Locator): Promise<boolean> {
        return (await templateCard.locator('.card-content.is-checked').count()) > 0;
    }

    private async waitForWorkloadFormCardSelection(templateCard: Locator): Promise<void> {
        await templateCard.locator('.card-content.is-checked').waitFor({ state: 'visible', timeout: 5000 });
    }

    async fillWorkloadFormGeneralSearchInput(sectionSelector: string, name: string): Promise<void> {
        await this.v2frame.locator(`${sectionSelector} ${this.workloadFormRunaiGeneralSearch}`).click();
        await this.v2frame
            .locator(`${sectionSelector} ${this.workloadFormRunaiGeneralSearch} input`)
            .fill(name);
    }

    async selectEnvironment(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.environmentSection, name);
        return this.v2frame.locator(`[aid="${name}-card"]`).click();
    }

    async showAllColumns(): Promise<void> {
        await this.page.locator(this.selectColumnsButton).click();
        await this.page.waitForSelector(this.columnsMenu, { timeout: 5000 });
        const isChecked = await this.page.locator(this.allColumnsCheckbox).getAttribute(this.elemChecked);
        if (isChecked !== 'true') {
            await this.page.locator(this.allColumnsCheckbox).click();
        }
        // close the columns menu
        await this.page.keyboard.press('Escape');
    }

    async checkDisplayedWorkloadDetails({
        jobName,
        requestedGpu,
        type,
        projectName,
        status,
        createdBy,
        pods
    }: {
        jobName: string;
        requestedGpu: string;
        type: WorkloadType;
        projectName: string;
        status: EWorkloadStatus;
        createdBy: string;
        pods: string;
    }): Promise<void> {
        // Arm the count-response waiter and trigger the search together via Promise.all
        // (waiter listed first so its listener registers synchronously before the request fires;
        // Promise.all also keeps the late waiter rejection from going unhandled if searchByName
        // throws). Scoped by `jobName` so a stale count response cannot satisfy the predicate.
        await Promise.all([
            this.waitForWorkloadsCountResponse({ nameInUrl: jobName }),
            this.searchByName(jobName)
        ]);
        // UI rendering delay
        await this.page.waitForTimeout(200);

        const jobRow = await this.getWorkloadRowIndexByName(jobName);
        await this.showAllColumns();
        await this.isRequestedGpuCorrect(jobRow, requestedGpu);
        await this.isTypeCorrect(jobRow, type);
        await this.isProjectCorrect(jobRow, projectName);
        await this.isStatusCorrect(jobRow, status);
        await this.isCreatedByCorrect(jobRow, createdBy);
        await this.isPodsCorrect(jobRow, pods);
    }

    async removeAdvancedFilter(): Promise<void> {
        await this.locator(this.advancedFilterSection).locator(this.advancedFilterRemoveButton).click();
    }

    async closeWorkloadForm(): Promise<void> {
        await this.page.locator(this.cancelBtn).click();
        await this.clickModalLeaveBtn();
    }

    async checkWorkloadUIStatus(
        workloadName: string,
        expectedPhase: string,
        invalidPhase?: string
    ): Promise<string> {
        await this.searchByName(workloadName);
        const promises = [
            this.page
                .waitForSelector(`[aid="${workloadName}-status"]:has-text("${expectedPhase}")`, {
                    timeout: 180000
                })
                .then(() => expectedPhase)
        ];
        if (invalidPhase) {
            promises.push(
                this.page
                    .waitForSelector(`[aid="${workloadName}-status"]:has-text("${invalidPhase}")`, {
                        timeout: 180000
                    })
                    .then(() => invalidPhase)
            );
        }
        const result = await Promise.any(promises);
        await this.clearGridSearchInput();

        if (invalidPhase && result === invalidPhase) {
            throw new Error(`Workload UI should not show ${invalidPhase} state`);
        }

        return result;
    }

    async selectComputeResource(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.computeResourceSection, name);
        const computeCard = this.v2frame.locator(`[aid=${name}-card]`);
        const isSelected = await computeCard.locator('.is-checked').count();
        if (!isSelected) {
            await computeCard.click();
        }
    }

    async selectDataSourceName(name: string): Promise<void> {
        await this.page.locator(this.dataSourceExpansion).click();
        await this.fillWorkloadFormGeneralSearchInput(this.dataSourceSection, name);
        const dataSourceCard = this.v2frame.locator(`[aid=${name}-card]`);
        const isSelected = await dataSourceCard.locator('.is-checked').count();
        if (!isSelected) {
            await dataSourceCard.click();
        }
    }

    async selectOptionFromDropdown(optionText: string): Promise<void> {
        const dropdownMenu = this.page.locator('[role="listbox"]').first();
        await dropdownMenu.locator(`[role="option"]:has-text("${optionText}")`).click();
    }

    async clickOnGeneralExpansionLabel(): Promise<void> {
        await this.page.locator(this.generalExpansionLabel).click();
    }

    async fillAttemptsInput(value: number): Promise<void> {
        await this.page.locator('input[aria-label="Attempts"]').fill(value.toString());
    }

    async fillEnvironmentVariablesByName(
        envVarValues: Array<{
            name: string;
            sourcetype: 'Custom' | 'Credentials' | 'ConfigMap';
            sourcename?: string;
            sourcevalue: string;
        }>
    ): Promise<void> {
        await this.clickOnSubExpansionByLabel('Runtime settings');
        const nameInputs = await this.page.locator(this.envVarNameInput).all();
        for (const envVar of envVarValues) {
            let matchingIndex = -1;
            for (let i = 0; i < nameInputs.length; i++) {
                const currentName = await nameInputs[i].inputValue();
                if (currentName === envVar.name) {
                    matchingIndex = i;
                    break;
                }
            }
            if (matchingIndex !== -1) {
                if (envVar.sourcetype === 'Custom') {
                    await this.page.locator(this.envVarSourceSelect).nth(matchingIndex).click();
                    await this.selectOptionFromDropdown('Custom');
                    await this.page.locator(this.envVarValue).nth(matchingIndex).fill(envVar.sourcevalue);
                } else if (envVar.sourcetype === 'Credentials') {
                    await this.page.locator(this.envVarSourceSelect).nth(matchingIndex).click();
                    await this.selectOptionFromDropdown('Credentials');
                    await this.page.locator(this.envVarCredentialSelect).nth(matchingIndex).click();
                    await this.selectOptionFromDropdown(envVar.sourcename);
                    await this.page.locator(this.envVarKeySelect).nth(matchingIndex).click();
                    await this.selectOptionFromDropdown(envVar.sourcevalue);
                } else if (envVar.sourcetype === 'ConfigMap') {
                    await this.page.locator(this.envVarSourceSelect).nth(matchingIndex).click();
                    await this.selectOptionFromDropdown('ConfigMap');
                    await this.page.locator(this.envVarConfigMapNameSelect).nth(matchingIndex).click();
                    await this.selectOptionFromDropdown(envVar.sourcename);
                    await this.page
                        .locator(this.envVarConfigMapKeyInput)
                        .nth(matchingIndex)
                        .fill(envVar.sourcevalue);
                }
            } else {
                throw new Error(`Environment variable ${envVar.name} not found`);
            }
        }
    }

    async getEventHistoryText(workloadName: string): Promise<string> {
        await this.toggleWorkloadDrawerContent(workloadName);
        await this.page.locator(this.eventHistoryTab).click();
        await waitUtils.waitForElementToExist(this, this.eventHistoryContent);

        const eventHistoryEl = this.page.locator(this.eventHistoryContent);
        const text = await eventHistoryEl.innerText();

        await this.closeWorkloadDrawerContent();

        return text;
    }

    async getEnvImage(): Promise<string> {
        return await this.page.locator(this.envImageInput).inputValue();
    }

    async getRuntimeCommand(): Promise<string> {
        return await this.page.locator(this.runtimeCommandInput).inputValue();
    }

    async getGpuDeivce(): Promise<string> {
        return await this.page.locator(this.gpuDeviceInput).inputValue();
    }

    async templateV2Exists(templateName: string): Promise<boolean> {
        await waitUtils.waitForElementToNotExist(this, this.runaiDefaultLoader, 10000);
        await waitUtils.waitForElementToNotExist(this, this.skeletonCardsList, 10000);

        await this.fillWorkloadFormGeneralSearchInput(this.workloadFormTemplateSectionV2, templateName);
        await this.page.waitForTimeout(2000);

        const template = this.page.locator(`[aid=${templateName}-card]`);
        const count = await template.count();

        if (count === 1) await template.click();

        return count === 1;
    }

    async waitForWorkloadToAppear(workloadName: string, timeout: number = 120000): Promise<void> {
        await this.page.waitForSelector(`td${this.workloadNameCol}:has-text("${workloadName}")`, {
            timeout: timeout
        });
    }
}
