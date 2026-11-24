/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { flowRootState, logger, Utils, waitUtils } from '../../../helpers';
import { Response } from 'playwright-core';
import { WorkloadPage } from './WorkloadPage';
import { HybridForm } from './HybridSubmissionForm';

export class WorkspacePage extends WorkloadPage {
    // workspace table elements
    addNewWorkspaceBtn: string;

    // create workspace elements
    createWorkspaceNameInput: string;

    createWorkspaceBtn: string;

    createWorkspaceContinueBtn: string;

    customUrlInput: string;

    generalSection: string;

    autoDeletimeTimeSecondsInput: string;

    autoDeletionTimeDaysInput: string;

    terminationGracePeriodSecondsInput: string;

    overQuotaToggle: string;

    templateSection: string;

    runaiGeneralSearch: string;

    jupyterTensorboardCard: string;

    ngcImageSelect: string;

    ngcTagSelect: string;

    private readonly hybridForm: HybridForm;

    constructor(page: Page) {
        super(page);

        this.mainPageTab = '[aid="menu-icon-workspaces"]';
        this.addNewWorkspaceBtn = '[aid="primary-cta-btn"]';

        this.runaiGeneralSearch = '.runai-general-search';

        // workspace elements
        this.createWorkspaceNameInput = '[aid="workspace-name-input"]';
        this.templateSection = '.template-section';
        this.createWorkspaceBtn = '[aid="submit-btn"]';
        this.createWorkspaceContinueBtn = '[aid="create-workspace-continue-btn"]';
        this.customUrlInput = '[aid="custom-url-input"]';
        this.generalSection = '[aid="general-section"]';
        this.autoDeletimeTimeSecondsInput = '[aid="seconds-input"]';
        this.autoDeletionTimeDaysInput = '[aid="days-input"]';
        this.terminationGracePeriodSecondsInput = '.termination-grace-period-section [aid="seconds-input"]';
        this.jupyterTensorboardCard = '[aid="jupyter-tensorboard-card"]';
        this.overQuotaToggle = '[aid="allow-over-quota-input"]';

        this.ngcImageSelect = '[aid="ngc-image-select"]';
        this.ngcTagSelect = '[aid="ngc-tag-select"]';
        this.hybridForm = new HybridForm(page);
    }

    async validateNewWorkspaceButtonExist(): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const workspaceBtnCount = await this.v2frame.locator(this.addNewWorkspaceBtn).count();
                return workspaceBtnCount !== 0;
            },
            'is new workspace button exists',
            true,
            10
        );
    }

    async waitForWorkspaceStatus(workspaceName: string, status = 'Active'): Promise<void> {
        await this.v2frame.waitForSelector(
            `td:has-text("${status}"):right-of(td:has-text("${workspaceName}"))`,
            { timeout: 120000 }
        );
    }

    // create workspace elements
    async fillWorkspaceName(name: string): Promise<void> {
        await this.page.waitForTimeout(500); // wait some time for the page to load
        return this.v2frame.locator(this.createWorkspaceNameInput).fill(name);
    }

    async waitForWorkspaceNameCount(name: string): Promise<void> {
        await this.waitForWorkloadsCountResponse({ nameInUrl: name });
    }

    async waitForImageTagsToLoad(): Promise<void> {
        function isFinished(response: Response): boolean {
            return response.url().includes('/images/tags') && response.status() === 200;
        }

        await this.waitForResponseByCallback(isFinished);
    }

    async clickOnContinue(): Promise<void> {
        if (await this.hybridForm.isQuickSubmitDropdownVisible()) {
            return this.hybridForm.continueWithQuickSubmit();
        }

        return await this.v2frame.locator(this.createWorkspaceContinueBtn).click();
    }

    async isContinueButtonVisible(): Promise<boolean> {
        return await this.v2frame.locator(this.createWorkspaceContinueBtn).isVisible();
    }

    async clickOnCreateWorkspaceBtn(): Promise<void> {
        return await this.v2frame.locator(this.createWorkspaceBtn).click();
    }

    async fillCustomUrlInput(txt: string): Promise<void> {
        return this.fillInputBySelector(this.customUrlInput, txt);
    }

    async openDataSourceSection(): Promise<void> {
        await this.v2frame.locator(this.dataSourceSection).scrollIntoViewIfNeeded();
        await this.v2frame.locator(this.dataSourceSection).click();
    }

    async openGeneralSection(): Promise<void> {
        await this.v2frame.locator(this.generalSection).click();
    }

    async addAutoDeletionTime(seconds: number): Promise<void> {
        await this.fillInputBySelector(this.autoDeletionTimeDaysInput, '0');
        await this.fillInputBySelector(this.autoDeletimeTimeSecondsInput, seconds.toString());
    }

    async selectTemplate(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.templateSection, name);
        await this.clickOnElementByAidName(Utils.spaceToDash(`${name}`).toLowerCase() + '-card');
    }

    async selectComputeResource(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.computeResourceSection, name);
        await this.clickOnElementByAidName(Utils.spaceToDash(`${name}`).toLowerCase() + '-card');
    }

    async selectNgcImage(image?: string): Promise<void> {
        if (!image) {
            await this.selectFirstOption(this.ngcImageSelect);
            return;
        }

        await this.fillInputBySelector(this.ngcImageSelect, image);
        await this.selectOptionByValue(this.ngcImageSelect, image);
    }

    async selectNgcTag(tag?: string): Promise<void> {
        if (!tag) {
            await this.selectFirstOption(this.ngcTagSelect);
            return;
        }

        await this.fillInputBySelector(this.ngcTagSelect, tag);
        await this.selectOptionByValue(this.ngcTagSelect, tag);
    }

    async fillGeneralSearchInput(sectionSelector: string, name: string): Promise<void> {
        await this.v2frame.locator(`${sectionSelector} ${this.runaiGeneralSearch}`).click();
        await this.v2frame.locator(`${sectionSelector} ${this.runaiGeneralSearch} input`).fill(name);
    }

    async clickOnAllowOverQuotaToggle(): Promise<void> {
        await this.v2frame.locator(this.overQuotaToggle).click();
    }

    async fillTerminationGracePeriodSeconds(seconds: number): Promise<void> {
        await this.fillInputBySelector(this.terminationGracePeriodSecondsInput, seconds.toString());
    }

    async fillFlexWorkspaceFirstPage(
        projectName: string,
        workspaceName: string,
        templateName?: string,
        templateAdvancedSetup: boolean = true
    ): Promise<void> {
        await this.navigateToWorkloads();
        await this.openWorkspaceForm();
        await this.selectClusterFromDropDown(flowRootState.cluster.name);
        await this.workloadFormSelectProject(projectName);

        if (templateName) {
            await this.workloadFormSelectTemplateV2(templateName);
            await this.page.waitForTimeout(500);
        } else {
            await this.workloadFormSelectTemplateV2();
        }

        await this.fillWorkspaceName(workspaceName);
        await this.waitForWorkspaceNameCount(workspaceName);

        if (!templateName) {
            await this.clickOnContinue();
        } else {
            if (templateAdvancedSetup) {
                await this.hybridForm.continueWithQuickSubmit();
            } else {
                await this.hybridForm.submitWithQuickSubmit();
            }
        }
        await this.page.waitForLoadState('load');
        await this.page.waitForTimeout(1000);
    }

    async submitWorkspaceWorkloadHybridUI(
        projectName: string,
        workspaceName: string,
        options?: {
            environment?: string;
            computeResource?: string;
            cmd?: string;
            templateName?: string;
            backoffLimit?: number;
            dataSource?: {
                dsName: string;
                containerPath?: string;
            };
        }
    ): Promise<void> {
        const {
            environment = 'tensorboard-tensorflow',
            computeResource = 'one-gpu',
            cmd,
            templateName,
            backoffLimit,
            dataSource
        } = options || {};

        await this.navigateToWorkloads();
        await this.openWorkspaceForm();
        await this.selectClusterFromDropDown(flowRootState.cluster.name);
        await this.workloadFormSelectProject(projectName);

        if (templateName) {
            await this.workloadFormSelectTemplate(templateName);
        } else {
            await this.workloadFormSelectTemplate();
        }
        await this.fillWorkspaceName(workspaceName);
        await this.waitForWorkspaceNameCount(workspaceName);
        await this.clickOnContinue();

        const hybridForm = new HybridForm(this.page);
        if (!templateName) {
            // fill environment
            await hybridForm.openEnvironmentsDrawer();
            await hybridForm.clickOnDrawerItemByLabel(environment);
            // fill compute resource
            await hybridForm.openComputesDrawer();
            await hybridForm.clickOnDrawerItemByLabel(computeResource);
            if (dataSource) {
                await hybridForm.fillPvcSection(dataSource.dsName);
                if (dataSource.containerPath) {
                    await hybridForm.fillContainerPathOverrideInput(dataSource.containerPath);
                }
            }
        }
        if (cmd) {
            await hybridForm.fillRuntimeSettings(cmd);
        }
        if (backoffLimit !== undefined) {
            await this.clickOnGeneralExpansionLabel();
            await this.fillAttemptsInput(backoffLimit);
        }
        await hybridForm.clickOnSubmit();
    }
}
