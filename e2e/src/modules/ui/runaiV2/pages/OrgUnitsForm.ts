/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Locator, Page } from 'playwright';
import { BaseV2Page } from './BaseV2Page';
import { RoleDisplayName } from '../../../../modules/helpers/config/roleMapping';
import { logger, waitUtils } from '../../../helpers';
import { expect } from '@playwright/test';
import { QTable } from '../components/QTable';

export enum OrgUnitsFormName {
    Description = 'Description',
    GpuQuota = 'GPU quota',
    GpuOverQuota = 'GPU over quota',
    CpuQuota = 'CPU quota',
    SchedulingRules = 'Scheduling rules',
    AccessRules = 'Access rules'
}

export enum SchedulingRulesIdleGpuTimeoutWorkloadType {
    Training = 'Training',
    PreemptibleWorkspaces = 'Preemptible workspaces',
    NonPreemptibleWorkspaces = 'Non-preemptible workspaces'
}

export enum SchedulingRulesTimeoutWorkloadType {
    Training = 'Training',
    Workspace = 'Workspace'
}

export enum AccessRulesSubjectType {
    User = 'User',
    ServiceAccount = 'Service account'
}

export class OrgUnitsForm extends BaseV2Page {
    primaryBtn: string;

    secondaryBtn: string;

    foldingSectionDiv: string;

    gpuQuotaOrgNameTd: string;

    gpuOverQuotaOrgNameTd: string;

    gpuQuotaOrgGpuValueTd: string;

    gpuQuotaOrgGpuReserveValueTd: string;

    gpuOverQuotaGpuWeightTd: string;

    nodepoolTabDiv: string;

    nodepoolOptionBtn: string;

    addNodepoolBtn: string;

    removeNodepoolOption: string;

    schedulingRulesAddRuleBtn: string;

    schedulingRulesIdleGpuTimeoutOption: string;

    schedulingRulesWorkloadTimeoutLimitOption: string;

    schedulingRulesNodeAffinityOption: string;

    schedulingRulesIdleGpuTimeoutSection: string;

    schedulingRulesWorkloadTimeoutLimitSection: string;

    schedulingRulesNodeAffinitySection: string;

    schedulingRulesWorkloadTypeSelect: string;

    multiSelectInput: string;

    daysInput: string;

    hoursInput: string;

    minutesInput: string;

    accessRulesSubjectSelect: string;

    accessRulesSaveRuleBtn: string;

    modalCreateAndCloseBtn: string;

    modalContinueEditingBtn: string;

    tourModal: string;

    tourModalConfirmLeaveBtn: string;

    quotaMessageDiv: string;

    clusterAndNameStepDiv: string;

    gpuQuotaStepDiv: string;

    gpuOverQuotaStepDiv: string;

    cpuQuotaStepDiv: string;

    schedulingRulesStepDiv: string;

    accessRulesStepDiv: string;

    gpuQuotaCardValueDiv: string;

    gpuOverQuotaToggleDiv: string;

    constructor(page: Page) {
        super(page);
        this.primaryBtn = '[aid="wizard-primary-btn"]';
        this.secondaryBtn = '[aid="wizard-secondary-btn"]';
        this.foldingSectionDiv = '.runai-folding-sub-section-header';
        this.gpuQuotaOrgNameTd = '[aid="orgUnitName"]';
        this.gpuOverQuotaOrgNameTd = '[aid="name"]';
        this.gpuQuotaOrgGpuValueTd = '[aid="gpuQuota"]';
        this.gpuQuotaOrgGpuReserveValueTd = '[aid="gpuDeserved"]';
        this.gpuOverQuotaGpuWeightTd = '[aid="weight"]';
        this.nodepoolTabDiv = '.node-pool-tabs-container';
        this.nodepoolOptionBtn = '[aid="nodepool-options-btn"]';
        this.addNodepoolBtn = '[aid="add-nodepool-btn"]';
        this.removeNodepoolOption = '[aid="remove-nodepool-btn"]';
        this.schedulingRulesAddRuleBtn = '[aid="add-scheduling-rule-button"]';
        this.schedulingRulesIdleGpuTimeoutOption = '[aid="idleGpuTimeout"]';
        this.schedulingRulesWorkloadTimeoutLimitOption = '[aid="workloadTimeLimit"]';
        this.schedulingRulesNodeAffinityOption = '[aid="nodeAffinity"]';
        this.schedulingRulesIdleGpuTimeoutSection = '[aid="idle-gpu-timeout"]';
        this.schedulingRulesWorkloadTimeoutLimitSection = '[aid="workload-time-limit"]';
        this.schedulingRulesNodeAffinitySection = 'Node type (Affinity)';
        this.multiSelectInput = '[aid="addable-multi-select"]';
        this.schedulingRulesWorkloadTypeSelect = '[aid="gpu-workload-type"]';
        this.daysInput = '[aid="days-input"]';
        this.hoursInput = '[aid="hours-input"]';
        this.minutesInput = '[aid="minutes-input"]';
        this.accessRulesSubjectSelect = '[aid="subject-select"]';
        this.accessRulesSaveRuleBtn = '[aid="save-form-footer-btn"]';
        this.modalCreateAndCloseBtn = '[aid="wizard-warning-create-and-close-btn"]';
        this.modalContinueEditingBtn = '[aid="wizard-warning-continue-editing-btn"]';
        this.tourModal = '.explore-new-experience-dialog';
        this.tourModalConfirmLeaveBtn = '[aid="guided-tour-leave-confirm-btn"]';
        this.quotaMessageDiv = '.quota-message';

        this.clusterAndNameStepDiv = '[aid="wizard-step-item-scope-name"]';
        this.gpuQuotaStepDiv = '[aid="wizard-step-item-gpu-quota"]';
        this.gpuOverQuotaStepDiv = '[aid="wizard-step-item-over-quota"]';
        this.cpuQuotaStepDiv = '[aid="wizard-step-item-cpu-quota"]';
        this.schedulingRulesStepDiv = '[aid="wizard-step-item-scheduling-rules"]';
        this.accessRulesStepDiv = '[aid="wizard-step-item-access-rules"]';
        this.gpuQuotaCardValueDiv = '.gpu-quota-availability-card .card-value';
        this.gpuOverQuotaToggleDiv = '[aid="over-quota-toggle"]';
    }

    async navigateToClusterAndNameStep(): Promise<void> {
        await this.page.locator(this.clusterAndNameStepDiv).click();
        await this.page.waitForLoadState('networkidle');
        await this.page.waitForTimeout(500);
    }

    async navigateToGpuQuotaStep(): Promise<void> {
        await this.page.locator(this.gpuQuotaStepDiv).click();
        await this.page.waitForLoadState('networkidle');
        await this.page.waitForTimeout(500);
    }

    async navigateToGpuOverQuotaStep(): Promise<void> {
        await this.page.locator(this.gpuOverQuotaStepDiv).click();
        await this.page.waitForLoadState('networkidle');
        await this.page.waitForTimeout(500);
    }

    async navigateToCpuQuotaStep(): Promise<void> {
        await this.page.locator(this.cpuQuotaStepDiv).click();
        await this.page.waitForLoadState('networkidle');
        await this.page.waitForTimeout(500);
    }

    async navigateToSchedulingRulesStep(): Promise<void> {
        await this.page.locator(this.schedulingRulesStepDiv).click();
        await this.page.waitForLoadState('networkidle');
        await this.page.waitForTimeout(500);
    }

    async navigateToAccessRulesStep(): Promise<void> {
        await this.page.locator(this.accessRulesStepDiv).click();
        await this.page.waitForLoadState('networkidle');
        await this.page.waitForTimeout(500);
    }

    async clickPrimaryBtn(): Promise<void> {
        await this.page.locator(this.primaryBtn).click();
    }

    async clickSecondaryBtn(): Promise<void> {
        await this.page.locator(this.secondaryBtn).click();
    }

    async clickCloseFormBtn(save: boolean = true): Promise<void> {
        await this.getCloseBtn().click();

        // sometimes the close button is clicked too quickly before save process is complete
        try {
            await this.getSaveChangesBtn().waitFor({ state: 'visible', timeout: 1000 });
        } catch {
            logger.info('the form was closed.');
            return;
        }

        if (save) {
            await this.clickSaveChangesBtn();
            logger.info('form was closed with changes saved.');
        } else {
            await this.clickDiscardBtn();
            logger.info('form was closed with changes discarded.');
        }
    }

    getSaveChangesBtn(): Locator {
        return this.page.getByRole('button', { name: 'Save changes' });
    }

    async clickSaveChangesBtn(): Promise<void> {
        await this.getSaveChangesBtn().click();
        await this.page.waitForLoadState('networkidle');
        await this.page.waitForTimeout(500);
    }

    async clickDiscardBtn(): Promise<void> {
        await this.page.getByRole('button', { name: 'Discard' }).click();
        await this.page.waitForTimeout(500);
    }

    async fillGpuQuotaInput(value: number): Promise<void> {
        await this.page.locator('.gpu-quota-card').getByLabel('GPUs').fill(value.toString());
    }

    async expandGpuQuotaFoldingOverviewSection(): Promise<void> {
        await this.page.locator(this.foldingSectionDiv).click();
    }

    async getNodepoolTabCount(): Promise<number> {
        const tabs = this.page.locator('.node-pool-tabs .q-tab__label');
        try {
            await tabs.first().waitFor({ state: 'visible', timeout: 3000 });
            return await tabs.count();
        } catch {
            // only has the 'default' node pool
            return 1;
        }
    }

    async clickAddNodepoolBtn(): Promise<void> {
        await this.locator(this.addNodepoolBtn).click();
    }

    async addNodePool(nodepoolName: string): Promise<void> {
        // a new org unit starts with the 'default' node pool only, the rest are added here
        await this.clickAddNodepoolBtn();
        await this.page
            .getByRole('menuitem')
            .filter({ has: this.page.getByText(nodepoolName, { exact: true }) })
            .click();
    }

    async switchToNodepoolTab(nodepoolName: string): Promise<void> {
        // call it only when multiple node pools shown, otherwise it won't work
        await this.page.locator('.node-pool-tabs').getByText(nodepoolName).click();
        await this.page.waitForLoadState('load');
        await this.page.waitForTimeout(1000);
    }

    async clickPopupGpuQuotaApplyBtn(): Promise<void> {
        await this.page.locator('.popup-card').getByRole('button', { name: 'APPLY' }).click();
        await this.page.locator('.popup-card').waitFor({ state: 'hidden', timeout: 5000 });
    }

    async clickSaveBtn(): Promise<void> {
        try {
            await this.page.getByRole('button', { name: 'Save', exact: true }).click({ timeout: 1000 });
            await this.page.waitForTimeout(500);
        } catch {
            logger.info('node pool save button not shown');
        }
    }

    async clickNodepoolOptionsBtn(nodepoolName: string): Promise<void> {
        const tabContent = this.page
            .locator(`${this.nodepoolTabDiv} .q-tab__content`)
            .filter({ has: this.page.locator('span.q-tab__label', { hasText: nodepoolName }) });

        await tabContent.locator(this.nodepoolOptionBtn).click();
    }

    async removeNodePool(nodepoolName: string): Promise<void> {
        await this.clickNodepoolOptionsBtn(nodepoolName);
        await this.locator(this.removeNodepoolOption).click();
    }

    async clickGpuOverQuotaToggle(): Promise<void> {
        await this.page.locator('[role="switch"]').click();
    }

    async fillGpuOverQuotaMaxGpuAllocationInput(value: number): Promise<void> {
        await this.page.getByLabel('Max GPU allocation').fill(value.toString());
    }

    async fillCpusInput(value: number): Promise<void> {
        await this.page.getByLabel('CPUs (Cores)').fill(value.toString());
    }

    async fillCpuMemoryInput(value: number): Promise<void> {
        await this.page.getByLabel('CPU memory').fill(value.toString());
    }

    async clearCpusInput(): Promise<void> {
        await this.page.getByLabel('CPUs (Cores)').clear();
    }

    async clearCpuMemoryInput(): Promise<void> {
        await this.page.getByLabel('CPU memory').clear();
    }

    async clickModalCreateAndCloseBtn(): Promise<void> {
        await this.page.locator(this.modalCreateAndCloseBtn).click();
    }

    async clickModalContinueEditingBtn(): Promise<void> {
        await this.page.locator(this.modalContinueEditingBtn).click();
    }

    getSchedulingRulesIdleGpuTimeoutSection(): Locator {
        return this.page.locator(this.schedulingRulesIdleGpuTimeoutSection);
    }

    getSchedulingRulesWorkloadTimeoutLimitSection(): Locator {
        return this.page.locator(this.schedulingRulesWorkloadTimeoutLimitSection);
    }

    getSchedulingRulesNodeAffinitySection(): Locator {
        return this.page.locator('div.q-card', { hasText: this.schedulingRulesNodeAffinitySection });
    }

    getSaveAndContinueBtn(): Locator {
        return this.page.getByRole('button', { name: /Save & Continue|Continue/ });
    }

    getCloseBtn(): Locator {
        return this.page.getByRole('button', { name: /^close$/i });
    }

    async openSchedulingRulesIdleGpuTimeoutConfigForm(): Promise<void> {
        await this.page.locator(this.schedulingRulesAddRuleBtn).click();
        await this.page.locator(this.schedulingRulesIdleGpuTimeoutOption).click();
    }

    async openSchedulingRulesWorkloadTimeoutConfigForm(): Promise<void> {
        await this.page.locator(this.schedulingRulesAddRuleBtn).click();
        await this.page.locator(this.schedulingRulesWorkloadTimeoutLimitOption).click();
    }

    async openSchedulingRulesNodeAffinityConfigForm(): Promise<void> {
        await this.page.locator(this.schedulingRulesAddRuleBtn).click();
        await this.page.locator(this.schedulingRulesNodeAffinityOption).click();
    }

    async clickAccessRulesAddRuleBtn(): Promise<void> {
        await this.page.getByRole('button', { name: 'ACCESS RULE', exact: true }).click();
    }

    async selectAccessRulesSubject(subjectName: 'User' | 'Service account', userName: string): Promise<void> {
        await this.page.locator(this.accessRulesSubjectSelect).click();
        await this.page.getByRole('option', { name: subjectName, exact: true }).click();
        await this.page.locator(this.multiSelectInput).fill(userName);
        await this.page.getByRole('option', { name: userName }).click();
    }

    async selectAccessRulesRoleByName(roleName: RoleDisplayName): Promise<void> {
        await this.page.getByLabel('Role').click();
        await this.page.getByRole('option', { name: roleName }).click();
    }

    async setSchedulingRulesIdleGpuTimeout(
        workloadType: SchedulingRulesIdleGpuTimeoutWorkloadType,
        days: number,
        hours: number,
        minutes: number
    ): Promise<void> {
        await this.openSchedulingRulesIdleGpuTimeoutConfigForm();
        const section = this.getSchedulingRulesIdleGpuTimeoutSection();

        await section.getByRole('combobox', { name: 'Workload Type' }).click();
        await this.page.getByRole('option', { name: workloadType }).click();
        await section.locator(this.daysInput).fill(days.toString());
        await section.locator(this.hoursInput).fill(hours.toString());
        await section.locator(this.minutesInput).fill(minutes.toString());
    }

    async getSchedulingRulesIdleGpuTimeoutWorkloadType(): Promise<string> {
        const section = this.getSchedulingRulesIdleGpuTimeoutSection();
        return await section.getByRole('combobox', { name: 'Workload Type' }).inputValue();
    }

    async getSchedulingRulesIdleGpuTimeoutDays(): Promise<number> {
        const section = this.getSchedulingRulesIdleGpuTimeoutSection();
        return Number(await section.locator(this.daysInput).inputValue());
    }

    async getSchedulingRulesIdleGpuTimeoutHours(): Promise<number> {
        const section = this.getSchedulingRulesIdleGpuTimeoutSection();
        return Number(await section.locator(this.hoursInput).inputValue());
    }

    async getSchedulingRulesIdleGpuTimeoutMinutes(): Promise<number> {
        const section = this.getSchedulingRulesIdleGpuTimeoutSection();
        return Number(await section.locator(this.minutesInput).inputValue());
    }

    async setSchedulingRulesWorkloadTimeoutLimit(
        workloadType: SchedulingRulesTimeoutWorkloadType,
        days: number,
        hours: number,
        minutes: number
    ): Promise<void> {
        await this.openSchedulingRulesWorkloadTimeoutConfigForm();
        const section = this.getSchedulingRulesWorkloadTimeoutLimitSection();

        await section.getByRole('combobox', { name: 'Workload Type' }).click();
        await this.page.getByRole('option', { name: workloadType }).click();
        await section.locator(this.daysInput).fill(days.toString());
        await section.locator(this.hoursInput).fill(hours.toString());
        await section.locator(this.minutesInput).fill(minutes.toString());
    }

    async setSchedulingRulesNodeAffinity(
        workloadType: SchedulingRulesTimeoutWorkloadType,
        nodeAffinityLabel: string
    ): Promise<void> {
        await this.openSchedulingRulesNodeAffinityConfigForm();
        const section = this.getSchedulingRulesNodeAffinitySection();

        await section.getByRole('combobox', { name: 'Workload Type' }).click();
        await this.page.getByRole('option', { name: workloadType }).click();
        await section.locator(this.multiSelectInput).fill(nodeAffinityLabel);
        await section.locator(this.multiSelectInput).getByRole('option', { name: nodeAffinityLabel }).click();
    }

    async clickAccessRulesSaveRuleBtn(): Promise<void> {
        await this.page.locator(this.accessRulesSaveRuleBtn).click();
        await this.page.waitForTimeout(500);
    }

    async setAccessRules(
        subjectName: AccessRulesSubjectType,
        userName: string,
        roleName: RoleDisplayName
    ): Promise<void> {
        await this.clickAccessRulesAddRuleBtn();
        await this.selectAccessRulesSubject(subjectName, userName);
        await this.selectAccessRulesRoleByName(roleName);
        await this.clickAccessRulesSaveRuleBtn();
    }

    async updateOtherOrgUnitGpuQuota(
        nodepoolsGpuQuotaConfig: Array<{ [key: string]: { orgUnitName: string; newGpuQuota: number } }>
    ): Promise<void> {
        const nodepooltabsCnt = await this.getNodepoolTabCount();

        for (const nodepoolConfig of nodepoolsGpuQuotaConfig) {
            for (const [nodepoolName, { orgUnitName, newGpuQuota }] of Object.entries(nodepoolConfig)) {
                if (nodepooltabsCnt > 1) {
                    await this.switchToNodepoolTab(nodepoolName);
                }

                await this.expandGpuQuotaFoldingOverviewSection();

                const targetRow = this.page.locator('.runai-table-row').filter({
                    has: this.page.locator(`${this.gpuQuotaOrgNameTd} span`, { hasText: orgUnitName })
                });

                await targetRow
                    .locator(`${this.gpuQuotaOrgGpuValueTd}, ${this.gpuQuotaOrgGpuReserveValueTd}`)
                    .click();

                await this.page.locator('.popup-card').getByLabel('GPUs').fill(newGpuQuota.toString());

                await this.clickPopupGpuQuotaApplyBtn();

                await this.clickSaveBtn();

                // it has to wait before clicking the 'continue' button, or the 'save changes' dialog will popup
                await this.page.waitForTimeout(2000);
            }
        }
    }

    async getOtherOrgUnitGpuQuotaData(): Promise<Array<Record<string, string>>> {
        const table = new QTable(this.page);
        await this.expandGpuQuotaFoldingOverviewSection();
        return table.getTableData();
    }

    async waitUntilGpuOverWeightCellClickable(
        nodepoolName: string,
        orgUnitName: string,
        maxRetries: number = 6
    ): Promise<void> {
        // after the "Over quota weight" toggle enabled, it takes a few seconds for the change to reflect on the department/project GPU over quota UI
        const nodepooltabsCnt = await this.getNodepoolTabCount();
        if (nodepooltabsCnt > 1) {
            await this.switchToNodepoolTab(nodepoolName);
        }

        for (let attempt = 1; attempt <= maxRetries; attempt++) {
            try {
                await expect(
                    this.page
                        .locator('.runai-table-row')
                        .filter({
                            has: this.page.locator(this.gpuOverQuotaOrgNameTd, { hasText: orgUnitName })
                        })
                        .locator(this.gpuOverQuotaGpuWeightTd)
                ).not.toHaveClass(/cell--disabled/);
                return;
            } catch (e) {
                if (attempt === maxRetries) {
                    throw e;
                } else {
                    await waitUtils.sleep(3000);
                }

                await this.page.reload({ waitUntil: 'networkidle' });
            }
        }
    }

    async updateOrgUnitGpuWeight(
        nodepoolsGpuWeightConfig: Array<{ [key: string]: { orgUnitName: string; newGpuWeight: number } }>
    ): Promise<void> {
        const nodepooltabsCnt = await this.getNodepoolTabCount();

        for (const nodepoolConfig of nodepoolsGpuWeightConfig) {
            for (const [nodepoolName, { orgUnitName, newGpuWeight }] of Object.entries(nodepoolConfig)) {
                if (nodepooltabsCnt > 1) {
                    await this.switchToNodepoolTab(nodepoolName);
                }

                const targetRow = this.page.locator('.runai-table-row').filter({
                    has: this.page.locator(this.gpuOverQuotaOrgNameTd, { hasText: orgUnitName })
                });

                await targetRow.locator(this.gpuOverQuotaGpuWeightTd).click();
                await this.page.locator('.popup-card input[type="number"]').fill(newGpuWeight.toString());

                await this.clickPopupGpuQuotaApplyBtn();
                await this.clickSaveBtn();
            }
        }
    }

    async configGpuQuota(config: Array<{ [key: string]: number }>): Promise<void> {
        const nodepooltabsCnt = await this.getNodepoolTabCount();

        for (const npGpuQuotaConfig of config) {
            for (const [nodepoolName, gpuQuota] of Object.entries(npGpuQuotaConfig)) {
                if (nodepooltabsCnt > 1) {
                    await this.switchToNodepoolTab(nodepoolName);
                }
                await this.fillGpuQuotaInput(gpuQuota);
                await this.clickSaveBtn();
            }
        }
    }

    async configGpuOverQuota(
        config: Array<{ [key: string]: { allowExceed: boolean; maxGpuAlloc?: number } }>
    ): Promise<void> {
        const nodepooltabsCnt = await this.getNodepoolTabCount();

        for (const npGpuOverQuotaConfig of config) {
            for (const [nodepoolName, gpuOverQuotaConfig] of Object.entries(npGpuOverQuotaConfig)) {
                if (nodepooltabsCnt > 1) {
                    await this.switchToNodepoolTab(nodepoolName);
                }

                if (gpuOverQuotaConfig.allowExceed) {
                    if (gpuOverQuotaConfig.maxGpuAlloc !== undefined) {
                        await this.fillGpuOverQuotaMaxGpuAllocationInput(gpuOverQuotaConfig.maxGpuAlloc);
                    }
                } else {
                    await this.clickGpuOverQuotaToggle();
                }

                await this.clickSaveBtn();
            }
        }
    }

    async configCpuQuota(
        config: Array<{ [key: string]: { cpus?: number; cpuMemory?: number } }>
    ): Promise<void> {
        const nodepooltabsCnt = await this.getNodepoolTabCount();

        for (const npCpuQuotaConfig of config) {
            for (const [nodepoolName, cpuOverQuotaConfig] of Object.entries(npCpuQuotaConfig)) {
                if (nodepooltabsCnt > 1) {
                    await this.switchToNodepoolTab(nodepoolName);
                }

                if (cpuOverQuotaConfig.cpus) {
                    await this.fillCpusInput(cpuOverQuotaConfig.cpus);
                }

                if (cpuOverQuotaConfig.cpuMemory) {
                    await this.fillCpuMemoryInput(cpuOverQuotaConfig.cpuMemory);
                }

                await this.clickSaveBtn();
            }
        }
    }

    async closeTourModalIfAppear(timeout: number = 2000): Promise<void> {
        try {
            await this.page.locator(this.tourModal).waitFor({ state: 'visible', timeout: timeout });
        } catch {
            logger.info('explore the new experience modal did not appear');
            return;
        }

        await this.page.getByRole('button', { name: 'Skip' }).click();
        await this.page.locator(this.tourModalConfirmLeaveBtn).click();
        logger.info('explore the new experience modal was closed');
    }

    async getQuotaMessage(): Promise<string> {
        return await this.page.locator(this.quotaMessageDiv).innerText({ timeout: 2000 });
    }

    async clickDontShowThisAgainCheckbox(): Promise<void> {
        await this.page.getByRole('checkbox', { name: /Don't show this again/ }).click();
    }

    async clearGpuQuotaInput(): Promise<void> {
        await this.page.locator('.gpu-quota-card').getByLabel('GPUs').clear();
    }

    async getGpuQuotaCardValue(nodepool: string): Promise<{ unassigned: number; total: number }> {
        const nodepooltabsCnt = await this.getNodepoolTabCount();

        if (nodepooltabsCnt > 1) {
            await this.switchToNodepoolTab(nodepool);
            await this.page.waitForLoadState('networkidle');
            await this.page.waitForTimeout(500);
        }

        const cardValue = await this.page.locator(this.gpuQuotaCardValueDiv).innerText();

        const [left, rest] = cardValue.split('/');
        const unassigned = Number(left);
        const total = Number(rest.match(/\d+/)![0]);

        return { unassigned: unassigned, total: total };
    }
}
