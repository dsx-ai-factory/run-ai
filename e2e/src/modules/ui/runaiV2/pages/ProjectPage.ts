/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { flowRootState, versionUtils, waitUtils } from '../../../helpers';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { expect } from 'chai';
import { dateUiUtils } from '../../utils/dateUiiUtils';
import { formUtils } from '../../utils/formUtils';
import { SupportedVersions } from '../../../helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../api/apis/assetsApi';
import { AssetsScopeSection } from '../components/AssetsScopeSection';
import { QuotaManagementSection } from '../components/QuotaManagementSection';
import { QTable } from '../components/QTable';
import { EProjectPhase } from '../../../helpers/utils/ProjectUtils';

enum EResourceType {
    Gpu = 'gpu',
    Cpu = 'cpu',
    Memory = 'memory'
}

type ResourceSelectorsMap = {
    [key in EResourceType]: {
        input: string;
        button: string;
    };
};

export enum ESchedulingRulesSubSectionSelector {
    TrainingMaxIdleDuration = 'training-job-max-idle-duration-secs-workload-duration',
    InteractivePreemptibleMaxIdleDuration = 'interactive-preemptible-job-max-idle-duration-secs-workload-duration',
    InteractiveMaxIdleDuration = 'interactive-job-max-idle-duration-secs-workload-duration',
    TrainingTimeLimit = 'training-job-time-limit-secs-workload-duration',
    InteractiveTimeLimit = 'interactive-job-time-limit-secs-workload-duration'
}

export enum ESchedulingRulesAddSectionSelector {
    IdleGpuTimeout = '[aid="idleGpuTimeout"]',
    WorkloadTimeLimit = '[aid="workloadTimeLimit"]'
}

export enum ESchedulingRulesParentSectionSelector {
    IdleGpuTimeout = '[aid="idle-gpu-timeout"]',
    WorkloadTimeLimit = '[aid="workload-time-limit"]'
}

export enum ESchedulingRulesWorkloadType {
    Training = 'Training',
    Interactive = 'Interactive',
    InteractivePreemptible = 'Interactive Preemptible'
}
export class ProjectPage extends BaseV2Page {
    //project table elements
    clusterFilterChip: string;

    addNewProjectBtn: string;

    emptyStateButton: string;

    projectStatusCol: string;

    deleteProjectBtn: string;

    editProjectBtn: string;

    projectNameCol: string;

    allocatedGpusCol: string;

    allocatedCpuCol: string;

    deleteProjectModalBtn: string;

    gpuQuotaCol: string;

    //create project elements

    projectNameInput: string;

    saveProjectBtn: string;

    gpuQuotaInput: string;

    cpuQuotaInput: string;

    memoryQuotaInput: string;

    gpuQuotaButton: string;

    cpuQuotaButton: string;

    memoryQuotaButton: string;

    saveQuotaButton: string;

    overQuotaPrioritySelect: string;

    overQuotaSelect: string;

    resourceSelectorsMap: ResourceSelectorsMap;

    nodePooRow: string;

    nodePoolSummary: string;

    nodePoolIcon: string;

    schedulingRulesSectionToggle: string;

    cancelBtn: string;

    scopeTreeBtn: string;

    searchInput: string;

    quotaTabContent: string;

    nodePoolNameSpan: string;

    assetsScopeSection: AssetsScopeSection;

    quotaMgmtSect: QuotaManagementSection;

    qTable: QTable;

    tooltip: string;

    projectDescriptionInput: string;

    createAndCloseWizardBtn: string;

    createAndContinueWizardBtn: string;

    warningContinueEditingBtn: string;

    warningCreateAndCloseBtn: string;

    constructor(page: Page) {
        super(page);

        this.mainPageTab = '[aid="menu-icon-projects"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Organization}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Projects}"]`;

        this.clusterFilterChip = '[aid="cluster-filter-chip"]';
        this.addNewProjectBtn = '[aid="primary-cta-btn"]';
        this.emptyStateButton = '[aid="empty-state-add-entity-btn"]';
        this.projectNameInput = '[aid="project-name-input"]';
        this.saveProjectBtn = '[aid="save-project-btn"]';
        this.editProjectBtn = '[aid="edit-project-btn"]';
        this.deleteProjectModalBtn = '[aid="delete-modal-delete-button"]';
        this.deleteProjectBtn = '[aid="delete-project-btn"]';
        this.gpuQuotaInput = '[aid="gpu-quota-input"]';
        this.cpuQuotaInput = '[aid="cpu-quota-input"]';
        this.memoryQuotaInput = '[aid="memory-quota-input"]';
        this.gpuQuotaCol = 'td[aid="totalGpuQuota"]';
        this.gpuQuotaButton = 'gpu-quota-button';
        this.cpuQuotaButton = 'cpu-quota-button';
        this.memoryQuotaButton = 'memory-quota-button';
        this.projectNameCol = 'td[aid="name"]';
        this.projectStatusCol = 'td[aid="phase"]';
        this.allocatedGpusCol = 'td[aid="gpuAllocated"]';
        this.allocatedCpuCol = 'td[aid="cpuAllocated"]';
        this.saveQuotaButton = 'save-quota-btn';
        this.overQuotaPrioritySelect = 'over-quota-priority-select';
        this.overQuotaSelect = 'over-quota-select';
        this.nodePooRow = '[aid="node-pool-row"]';
        this.nodePoolSummary = '[aid="node-pools-summary"]';
        this.nodePoolIcon = '[aid="node-pool-icon"]';
        this.schedulingRulesSectionToggle = '[aid="scheduling rules-expansion-toggle"]';
        this.projectDescriptionInput = '[aid="project-description-input"]';
        this.createAndCloseWizardBtn = '[aid="wizard-secondary-btn"]';
        this.createAndContinueWizardBtn = '[aid="wizard-primary-btn"]';
        this.warningContinueEditingBtn = '[aid="wizard-warning-continue-editing-btn"]';
        this.warningCreateAndCloseBtn = '[aid="wizard-warning-create-and-close-btn"]';
        this.resourceSelectorsMap = {
            [EResourceType.Cpu]: {
                input: this.cpuQuotaInput,
                button: this.cpuQuotaButton
            },
            [EResourceType.Gpu]: {
                input: this.gpuQuotaInput,
                button: this.gpuQuotaButton
            },
            [EResourceType.Memory]: {
                input: this.memoryQuotaInput,
                button: this.memoryQuotaButton
            }
        };
        this.cancelBtn = '[aid="cancel-form-footer-btn"]';
        this.scopeTreeBtn = '[class="tree-button-wrapper"]';
        this.searchInput = '[placeholder="Search department"]';
        this.quotaTabContent = '[aid="quota-tab-content"]';
        this.nodePoolNameSpan = '[aid="node-pool-name"] span';
        this.assetsScopeSection = new AssetsScopeSection(page);
        this.quotaMgmtSect = new QuotaManagementSection(page);
        this.qTable = new QTable(page);
        this.tooltip = '.tooltip';
    }

    async clickOnAddNewProjectBtn(): Promise<void> {
        await this.v2frame.locator(this.addNewProjectBtn).click();
    }

    async focusProjectDescriptionInput(): Promise<void> {
        return this.page.locator(this.projectDescriptionInput).focus();
    }

    async pressEnterOnProjectDescriptionInput(): Promise<void> {
        return this.pressKeyOnElement(this.projectDescriptionInput, 'Enter');
    }

    async fillProjectDescriptionInput(value: string): Promise<void> {
        return this.page.locator(this.projectDescriptionInput).fill(value);
    }

    async getProjectDescriptionLineCount(): Promise<number> {
        const value = await this.page.locator(this.projectDescriptionInput).inputValue();
        if (value === '') return 0;
        return value.split('\n').length;
    }

    async hasNewlineAtEnd(): Promise<boolean> {
        const value = await this.page.locator(this.projectDescriptionInput).inputValue();
        return value.endsWith('\n');
    }

    async getProjectDescriptionInputValueWithEscapes(): Promise<string> {
        const value = await this.page.locator(this.projectDescriptionInput).inputValue();
        // show newline characters for debugging and verification
        return value.replace(/\n/g, '\\n').replace(/\r/g, '\\r');
    }

    async submitProjectForm(): Promise<void> {
        return await this.v2frame.locator(this.saveProjectBtn).click();
    }

    /**
     * Clicks the "Create Project & Close" button in the wizard
     * This creates a "default" project with only the first step (name) filled
     */
    async clickOnCreateAndCloseFromWizard(): Promise<void> {
        return await this.v2frame.locator(this.createAndCloseWizardBtn).click();
    }

    /**
     * Clicks the "Create Project & Close" button in the warning dialog
     * This confirms the creation of the project with default settings and closes the wizard
     */
    async clickOnWarningCreateAndClose(): Promise<void> {
        return await this.v2frame.locator(this.warningCreateAndCloseBtn).click();
    }

    async clickIOnEditProject(projectName: string): Promise<void> {
        await this.findByText(this.projectNameCol, projectName).click();
        await this.v2frame.locator(this.editProjectBtn).click();
    }

    async openDeleteProjectDialog(projectName: string): Promise<void> {
        await this.findByText(this.projectNameCol, projectName).click();
        await this.page.locator(this.deleteProjectBtn).click();
        const confirmBtn = this.page.locator(this.deleteProjectModalBtn);
        await confirmBtn.waitFor({ state: 'visible' });
        await waitUtils.waitForCondition(
            async () => await confirmBtn.isEnabled(),
            'Wait for the delete project dialog body to finish loading',
            true,
            30
        );
    }

    async getDeleteProjectDialogText(): Promise<string> {
        return ((await this.page.locator(this.runaiBaseDialogBody).textContent()) || '').trim();
    }

    async confirmDeleteProject(): Promise<void> {
        await this.page.locator(this.deleteProjectModalBtn).click();
    }

    async deleteProject(projectName: string): Promise<void> {
        await this.openDeleteProjectDialog(projectName);
        await this.confirmDeleteProject();
    }

    async waitForProjectDeletion(projectName: string): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const projectNames = await this.page.locator(this.projectNameCol).allTextContents();
                return !projectNames.some((name: string) => name.trim() === projectName);
            },
            `Wait for project ${projectName} to be deleted from UI`,
            true,
            60,
            1
        );
    }

    async getProjectNames(): Promise<string[]> {
        return await this.page.locator(this.projectNameCol).allTextContents();
    }

    async toggleNamespaceSection(): Promise<void> {
        await this.v2frame.locator('[aid="namespace-section-toggle"]').click();
    }

    async fillNodePoolQuota(type: EResourceType, value: string, nodePoolName: string): Promise<void> {
        if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_20)) {
            await this.v2frame.locator('[aid="quota-tab"]').click();
        }
        await this.v2frame
            .locator(`[data-testid='${this.resourceSelectorsMap[type].button}-${nodePoolName}']`)
            .click();
        await this.fillInputBySelector(this.resourceSelectorsMap[type].input, value);
        await this.clickOnElementByAidName(this.saveQuotaButton);
        await waitUtils.waitForElementToNotExist(this, this.resourceSelectorsMap[type].input);
    }

    async selectOverQuotaPriority(text: string, nodePoolName: string): Promise<void> {
        await this.v2frame.locator(`[data-testid="${this.overQuotaPrioritySelect}-${nodePoolName}"]`).click();
        await this.v2frame.locator(`div[role="option"]:has-text("${text}")`).click();
    }

    async selectOverQuota(text: string, nodePoolName: string): Promise<void> {
        if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_20)) {
            await this.v2frame.locator('[aid="scheduling-tab"]').click();
        }
        await this.v2frame.locator(`[data-testid="${this.overQuotaSelect}-${nodePoolName}"]`).click();
        await this.v2frame.locator(`div[role="option"]:has-text("${text}")`).click();
    }

    async toggleScheduleRulesSection(): Promise<void> {
        await this.v2frame.locator(this.schedulingRulesSectionToggle).click();
    }

    async addSchedulingRuleParentSection(btnSelector: ESchedulingRulesAddSectionSelector): Promise<void> {
        await this.page.locator('[aid="add-scheduling-rule-button"]').click();
        await this.page.locator(`${btnSelector}`).click();
    }

    async validateSchedulingRuleInputValue(
        rule: ESchedulingRulesSubSectionSelector,
        valueInMs: number
    ): Promise<void> {
        const sectionSelector = `[aid="${rule}"]`;

        const daysInput = this.locator(`${sectionSelector} [aid="days-input"]`);
        const hoursInput = this.locator(`${sectionSelector} [aid="hours-input"]`);
        const minutesInput = this.locator(`${sectionSelector} [aid="minutes-input"]`);

        const days = parseInt(await daysInput.inputValue(), 10) || 0;
        const hours = parseInt(await hoursInput.inputValue(), 10) || 0;
        const minutes = parseInt(await minutesInput.inputValue(), 10) || 0;

        const { daysInSeconds, hoursInSeconds, minutesInSeconds } = dateUiUtils.convertTimeUnitsToSeconds(
            days,
            hours,
            minutes
        );

        const totalDurationInSeconds = daysInSeconds + hoursInSeconds + minutesInSeconds;

        expect(totalDurationInSeconds).to.eql(valueInMs);
    }

    async setSchedulingRuleDuration(
        rule: ESchedulingRulesSubSectionSelector,
        durationInSec: number
    ): Promise<void> {
        const sectionSelector = `[aid="${rule}"]`;
        const daysInput = this.locator(`${sectionSelector} [aid="days-input"]`);
        const hoursInput = this.locator(`${sectionSelector} [aid="hours-input"]`);
        const minutesInput = this.locator(`${sectionSelector} [aid="minutes-input"]`);

        const { days, hours, minutes } = dateUiUtils.convertSecondsToTimeUnits(durationInSec);

        if (days > 0) {
            await daysInput.fill(days.toString());
        }
        if (hours > 0) {
            await hoursInput.fill(hours.toString());
        }
        if (minutes > 0) {
            await minutesInput.fill(minutes.toString());
        }
    }

    async setWorkloadTypeSchedulingRule(
        parentSection: ESchedulingRulesParentSectionSelector,
        workloadType: ESchedulingRulesWorkloadType
    ): Promise<void> {
        await formUtils.selectDropdownValueByText(
            this.page,
            `${parentSection} [aid="gpu-workload-type"]`,
            workloadType
        );
    }

    async validateProjectExceedsInheritedDuration(
        rule: ESchedulingRulesSubSectionSelector,
        validDuration: number
    ): Promise<void> {
        const sectionSelector = `[aid="${rule}"]`;
        const formattedDuration = dateUiUtils.formatDuration(validDuration);
        expect(
            await this.locator(
                `${sectionSelector} div[role="alert"]:has-text("Enter a limit lower than ${formattedDuration}")`
            ).isVisible()
        ).eql(true);
    }

    async resetSchedulingRuleDuration(rule: ESchedulingRulesSubSectionSelector): Promise<void> {
        const sectionSelector = `[aid="${rule}"]`;
        await this.locator(`${sectionSelector} [aid="reset-button"]`).click();
    }

    async getAllocatedResourceCount(
        projectName: string,
        resourceType: EResourceType.Gpu | EResourceType.Cpu
    ): Promise<number> {
        await this.goToPage();
        const colSelector = resourceType === EResourceType.Gpu ? this.allocatedGpusCol : this.allocatedCpuCol;
        await waitUtils.waitForCondition(
            async () => {
                const projectNames = await this.page.locator(this.projectNameCol).allTextContents();
                return projectNames.findIndex((name: string) => name.trim() === projectName) !== -1;
            },
            `Wait for project ${projectName} to appear in UI list`,
            true,
            10,
            0.5
        );
        const projectNames = await this.page.locator(this.projectNameCol).allTextContents();
        const rowIndex = projectNames.findIndex((name: string) => name.trim() === projectName);
        const allocatedCells = await this.page.locator(colSelector).allTextContents();
        const value = allocatedCells[rowIndex] || '0';
        return resourceType === EResourceType.Gpu ? parseInt(value, 10) : parseFloat(value);
    }

    async getAllocatedGpuCount(projectName: string): Promise<number> {
        return this.getAllocatedResourceCount(projectName, EResourceType.Gpu);
    }

    async getAllocatedCpuCount(projectName: string): Promise<number> {
        return this.getAllocatedResourceCount(projectName, EResourceType.Cpu);
    }

    async createProject(
        projectName: string,
        gpuQuota: string,
        priority?: string,
        overQuota?: boolean,
        maxGpuLimit?: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
            nodePool?: string;
        }
    ): Promise<void> {
        await this.clickOnAddNewProjectBtn();
        await this.selectScope(options);
        await this.page.locator(this.projectNameInput).fill(projectName);
        await this.quotaMgmtSect.setQuota(gpuQuota, priority, overQuota, maxGpuLimit, options.nodePool);
        await this.page.locator(this.saveProjectBtn).click();
    }

    async closeProjectForm(): Promise<void> {
        await this.page.locator(this.cancelBtn).click();
        await this.clickModalLeaveBtn();
    }

    async selectScope(options?: { scope?: Scope; scopeEntityName?: string }): Promise<void> {
        this.assetsScopeSection.scopeTreeBtn = this.scopeTreeBtn;
        this.assetsScopeSection.searchElement = this.searchInput;
        await this.assetsScopeSection.selectScope(options);
    }

    async checkNodePoolAtTop(nodePoolName: string): Promise<boolean> {
        let allNodePoolNames = await this.page.locator(this.nodePoolNameSpan).allTextContents();

        if (allNodePoolNames.length === 0) {
            await this.page.waitForTimeout(2000);
            allNodePoolNames = await this.page.locator(this.nodePoolNameSpan).allTextContents();
        }

        if (allNodePoolNames.length === 0) {
            throw new Error('No node pools found in the quota tab content');
        }

        let targetIndex = allNodePoolNames.findIndex((name) => name.includes(nodePoolName));

        if (targetIndex === -1) {
            const prefix = nodePoolName.split('-').slice(0, 2).join('-');
            targetIndex = allNodePoolNames.findIndex((name) => name.includes(prefix));
        }

        if (targetIndex === -1) {
            throw new Error(
                `Node pool ${nodePoolName} not found. Available: [${allNodePoolNames.join(', ')}]`
            );
        }

        const isAtTop = targetIndex === 0;

        return isAtTop;
    }

    async fillProjectName(name: string): Promise<void> {
        await this.page.locator(this.projectNameInput).fill(name);
    }

    async waitForDialogToClose(): Promise<void> {
        await waitUtils.waitForElementToNotExist(this, '.dialog-content-container', 30);
    }
}
