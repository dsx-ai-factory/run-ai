/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { MENU_SUBJECT_NAMES, MENU_SUB_ITEMS } from '../../utils/sidenavUtils';
import { Scope } from '../../../api/apis/assetsApi';
import { AssetsScopeSection } from '../components/AssetsScopeSection';
import { DISTRIBUTED_TOOLS_LABELS } from './TrainingPage';
import { HybridForm } from './HybridSubmissionForm';
import { WorkloadType } from '../../../api/apis/workloadsApi/generated/api';
import { QTable } from '../components/QTable';
import { Utils } from '../../../helpers';
import { expect } from '@playwright/test';

interface WorkloadConfig {
    image: string;
    runtimeOpts?: {
        cmdArgsOpts?: {
            cmd?: string;
            args?: string;
        };
    };
    existingComputeResName?: string;
}

export class TemplatePageV2 extends BaseV2Page {
    newTemplateBtn: string;

    workspaceMenu: string;

    trainingMenu: string;

    singleNodeOption: string;

    multiNodeOption: string;

    workersOnlyOption: string;

    distributedToolsSelect: string;

    templateName: string;

    templateDescription: string;

    submitBtn: string;

    templateCol: string;

    upadteBtn: string;

    copyBtn: string;

    deleteBtn: string;

    templateNameSection: string;

    modalDeleteBtn: string;

    envAssetLinkBtn: string;

    computeResourceAssetLinkBtn: string;

    envAssetResetBtn: string;

    computeResourceAssetResetBtn: string;

    cantEditTooltip: string;

    assetResetBtnTooltip: string;

    hybridForm: HybridForm;

    assetScopeSection: AssetsScopeSection;

    qTable: QTable;

    constructor(page: Page) {
        super(page);
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Templates}"]`;

        this.newTemplateBtn = '[aid="new-template-dropdown"] button';
        this.workspaceMenu = '[aid="new-template-workspace-option"]';
        this.trainingMenu = '[aid="new-template-training-option"]';
        this.singleNodeOption = '[aid="radio-option-Standard"]';
        this.multiNodeOption = '[aid="radio-option-Distributed"]';
        this.workersOnlyOption = '[aid="radio-option-Workers only"]';
        this.distributedToolsSelect = '[aid="dist-framework-select"]';
        this.templateName = '[aid="workload-template-name-input"]';
        this.templateDescription = '[aid="workload-template-description-input"]';
        this.submitBtn = '[aid="submit-btn"]';
        this.templateCol = '[aid="name"]';
        this.upadteBtn = '[aid="edit-template-btn"]';
        this.copyBtn = '[aid="copy-template-btn"]';
        this.deleteBtn = '[aid="delete-template-btn"]';
        this.templateNameSection = '[aid="template-name-section"]';
        this.modalDeleteBtn = '[aid="delete-modal-delete-button"]';
        this.envAssetLinkBtn = '//*[@aid="hybrid-environment-section"]//*[@aid="link-btn"]';
        this.envAssetResetBtn = '//*[@aid="hybrid-environment-section"]//*[@aid="reset-btn"]';
        this.computeResourceAssetLinkBtn = '//*[@aid="hybrid-compute-resource-section"]//*[@aid="link-btn"]';
        this.computeResourceAssetResetBtn =
            '//*[@aid="hybrid-compute-resource-section"]//*[@aid="reset-btn"]';
        this.cantEditTooltip =
            '[role="tooltip"]:has-text("This can\'t be modified when the Link button is active")';
        this.assetResetBtnTooltip = '[role="tooltip"]:has-text("To reset, first unlink with setup")';

        this.assetScopeSection = new AssetsScopeSection(page);
        this.hybridForm = new HybridForm(page);
        this.qTable = new QTable(page);
    }

    async openCreationForm(workloadType: WorkloadType): Promise<void> {
        switch (workloadType) {
            case WorkloadType.Training:
            case WorkloadType.Distributed:
                await Utils.retry(async () => {
                    await this.page.locator(this.newTemplateBtn).click();
                    await expect(this.page.locator(this.trainingMenu)).toBeVisible({ timeout: 3000 });
                });
                await this.page.locator(this.trainingMenu).click();
                break;
            case WorkloadType.Workspace:
                await Utils.retry(async () => {
                    await this.page.locator(this.newTemplateBtn).click();
                    await expect(this.page.locator(this.workspaceMenu)).toBeVisible({ timeout: 3000 });
                });
                await this.page.locator(this.workspaceMenu).click();
                break;
            default:
                throw new Error(`Unsupported workload template type: ${workloadType}!`);
        }
    }

    async selectMultiNodeOption(): Promise<void> {
        await this.page.locator(this.multiNodeOption).click();
    }

    async selectWorkersOnlyOption(): Promise<void> {
        await this.page.locator(this.workersOnlyOption).click();
    }

    async selectDistributedFramework(frameworkName: DISTRIBUTED_TOOLS_LABELS): Promise<void> {
        await this.page.locator(this.distributedToolsSelect).click();
        await this.page.locator(`div[role="option"]:has-text("${frameworkName}")`).click();
    }

    async fillName(name: string): Promise<void> {
        await this.page.locator(this.templateName).fill(name);
    }

    async createSingleNodeWorkloadTemplate(
        workloadType: WorkloadType,
        templateName: string,
        scopeOpts: {
            scope: Scope;
            scopeEntityName: string;
        },
        workloadConfigOpts: WorkloadConfig
    ): Promise<void> {
        // fill the 1st page
        await this.openCreationForm(workloadType);
        await this.assetScopeSection.selectScope(scopeOpts);
        await this.fillName(templateName);
        await this.clickSubmitBtn();

        // fill the 2nd page
        await this.page.waitForLoadState('load');
        await this.setWorkloadConfig(workloadConfigOpts);

        // submit template
        await Utils.retry(async () => {
            await this.clickSubmitBtn();
            await expect(this.page.locator(this.newTemplateBtn)).toBeVisible({ timeout: 3000 });
        });
    }

    async createDistributedWorkloadTemplate(
        workloadType: WorkloadType,
        templateName: string,
        scopeOpts: {
            scope: Scope;
            scopeEntityName: string;
        },
        distributedOpts: {
            multiNodeFramework: DISTRIBUTED_TOOLS_LABELS;
            workerConfigOpts: WorkloadConfig;
            workersOnly?: boolean;
            masterConfigOpts?: WorkloadConfig;
        }
    ): Promise<void> {
        // fill the 1st page
        await this.openCreationForm(workloadType);
        await this.assetScopeSection.selectScope(scopeOpts);
        await this.fillDisitrbuedFirstPage(distributedOpts.multiNodeFramework, distributedOpts.workersOnly);
        await this.fillName(templateName);
        await this.clickSubmitBtn();

        // fill the 2nd page
        await this.page.waitForLoadState('load');
        await this.setWorkloadConfig(distributedOpts.workerConfigOpts);
        await this.page.waitForTimeout(1000);

        // if distributed framework is specified and it's not 'JAX' and 'workersOnly option is not chosen',
        // then need to go to the 3rd page to do master setup.
        if (
            this.masterSettingsNeededForDistributedTemplate(
                distributedOpts.multiNodeFramework,
                distributedOpts.workersOnly
            )
        ) {
            await this.clickSubmitBtn();
            await this.page.waitForLoadState('load');
            await this.page.waitForTimeout(1000);

            // if masterConfigOpts is not provided, it means user wants to use same config for master as workers,
            // thus do nothing on the 3rd page.
            if (distributedOpts.masterConfigOpts) {
                await this.fillDistributedMasterPage(distributedOpts.masterConfigOpts);
                await this.hybridForm.waitForDistributedMasterSubmitButton();
                await this.hybridForm.clickOnSubmitMasterForm();
            }
        }

        // submit template
        await Utils.retry(async () => {
            await this.clickSubmitBtn();
            await expect(this.page.locator(this.newTemplateBtn)).toBeVisible({ timeout: 3000 });
        });
    }

    async fillDisitrbuedFirstPage(
        multiNodeFramework: DISTRIBUTED_TOOLS_LABELS,
        workersOnly: boolean
    ): Promise<void> {
        await this.selectMultiNodeOption();
        await this.selectDistributedFramework(multiNodeFramework);
        if (
            (multiNodeFramework === DISTRIBUTED_TOOLS_LABELS.PY_TORCH ||
                multiNodeFramework === DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW) &&
            workersOnly
        ) {
            await this.selectWorkersOnlyOption();
        }
    }

    async fillDistributedMasterPage(masterOpts: WorkloadConfig): Promise<void> {
        await this.page.waitForFunction((selector) => {
            const input = document.querySelector<HTMLInputElement>(selector);
            return input && !input.disabled;
        }, this.hybridForm.gpuDeviceInput);
        await this.setWorkloadConfig(masterOpts);
    }

    masterSettingsNeededForDistributedTemplate(
        multiNodeFramework: DISTRIBUTED_TOOLS_LABELS,
        workersOnly: boolean
    ): boolean {
        return (
            multiNodeFramework &&
            multiNodeFramework !== DISTRIBUTED_TOOLS_LABELS.JAX &&
            multiNodeFramework !== DISTRIBUTED_TOOLS_LABELS.MPI &&
            !workersOnly
        );
    }

    async setWorkloadConfig(opts: WorkloadConfig): Promise<void> {
        await this.hybridForm.fillImageInput(opts.image);
        await this.hybridForm.fillRuntimeSettings(
            opts.runtimeOpts?.cmdArgsOpts?.cmd,
            opts.runtimeOpts?.cmdArgsOpts?.args
        );
        if (opts.existingComputeResName) {
            await this.setComputeResources(opts.existingComputeResName);
        }
    }

    async setComputeResources(existingComputeResName: string): Promise<void> {
        await this.hybridForm.selectExistingComputeResource(existingComputeResName);
    }

    async waitForTemplateExist(templateName: string): Promise<void> {
        try {
            await this.searchByName(templateName);

            await this.page.waitForSelector(`td${this.templateCol}:has-text("${templateName}")`, {
                timeout: 10000
            });
        } finally {
            await this.clearSearchInput();
            await this.page.waitForLoadState('load');
            await this.page.waitForTimeout(500);
        }
    }

    async getTemplateData(templateName: string): Promise<Record<string, string>> {
        return await this.qTable.getEntityData('Template', templateName);
    }

    async toggleTemplate(templateName: string): Promise<void> {
        await this.clickOnElementByText(`td${this.templateCol}`, `${templateName}`);
    }

    async clickUpadteTemplateBtn(): Promise<void> {
        await this.page.locator(this.upadteBtn).click();
    }

    async clickCopyTemplateBtn(): Promise<void> {
        await this.page.locator(this.copyBtn).click();
    }

    async clickSubmitBtn(): Promise<void> {
        await this.page.locator(this.submitBtn).scrollIntoViewIfNeeded({ timeout: 3000 });
        // it seems that the event handler is occasionally not attached before clicking the submit button, so I added a wait to address this.
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.submitBtn).click();
    }

    async updateName(newTemplateName: string): Promise<void> {
        await this.page.locator(this.templateNameSection).click();
        await this.page.locator(this.templateName).clear();
        await this.fillName(newTemplateName);
    }

    async delete(templateName: string): Promise<void> {
        await this.toggleTemplate(templateName);
        await this.page.waitForTimeout(500);
        await this.page.locator(this.deleteBtn).click();
        await this.page.locator(this.modalDeleteBtn).click();
        await this.waitForAssetToBeDeletedFromUI(this, templateName);
    }

    async fillSingleNodeWorkloadTemplateFirstPage(
        workloadType: WorkloadType,
        templateName: string,
        scopeOpts: {
            scope: Scope;
            scopeEntityName: string;
        }
    ): Promise<void> {
        await this.openCreationForm(workloadType);
        await this.assetScopeSection.selectScope(scopeOpts);
        await this.fillName(templateName);
        await this.clickSubmitBtn();
        await this.page.waitForLoadState('load');
    }

    async toggleEnvLinkButton(): Promise<void> {
        await this.page.locator(this.envAssetLinkBtn).click();
    }

    async toggleComputeResourceLinkButton(): Promise<void> {
        await this.page.locator(this.computeResourceAssetLinkBtn).click();
    }

    async hoverOnImageInput(): Promise<void> {
        await this.page
            .locator(this.hybridForm.hybridEnvironmentSection.hybridImageInput)
            .hover({ force: true });
    }

    async hoverOnGpuDeviceInput(): Promise<void> {
        await this.page.locator(this.hybridForm.hybridComputeSection.gpuDevicesInput).hover({ force: true });
    }

    async hoverOnEnvAssetResetBtn(): Promise<void> {
        await this.page.locator(this.envAssetResetBtn).hover({ force: true });
    }

    async hoverOnComputeResourceAssetResetBtn(): Promise<void> {
        await this.page.locator(this.computeResourceAssetResetBtn).hover({ force: true });
    }

    async clickEnvAssetResetBtn(): Promise<void> {
        await this.page.locator(this.envAssetResetBtn).click();
    }

    async clickComputeResourceAssetResetBtn(): Promise<void> {
        await this.page.locator(this.computeResourceAssetResetBtn).click();
    }

    async focusTemplateName(): Promise<void> {
        return this.page.locator(this.templateName).focus();
    }

    async pressEnterOnTemplateName(): Promise<void> {
        await this.page.locator(this.templateName).focus();
        return this.page.locator(this.templateName).press('Enter');
    }
}
