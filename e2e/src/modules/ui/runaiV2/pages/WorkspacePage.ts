/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { flowRootState, logger, Utils, waitUtils } from '../../../helpers';
import { Response } from 'playwright-core';
import { WorkloadPage } from './WorkloadPage';
import { AssetKind } from '../../../api/apis/assetsApi';
import { HybridForm } from './HybridSubmissionForm';

export class WorkspacePage extends WorkloadPage {
    // workspace table elements
    addNewWorkspaceBtn: string;

    // create workspace elements
    createWorkspaceNameInput: string;

    createWorkspaceBtn: string;

    createWorkspaceContinueBtn: string;

    newEnvironmentBtn: string;

    newComputeBtn: string;

    customUrlInput: string;

    uidInput: string;

    gidInput: string;

    supplementalGrpInput: string;

    envSectionMoreSettingsButton: string;

    envSectionArgsInput: string;

    envSectionCommandInput: string;

    annotationNameInput: string;

    annotationValueInput: string;

    generalSection: string;

    autoDeletimeTimeSecondsInput: string;

    autoDeletionTimeDaysInput: string;

    terminationGracePeriodSecondsInput: string;

    volumeSection: string;

    addVolumeBtn: string;

    volumeStorageClassSelect: string;

    volumeAccessModeSelect: string;

    volumeSizeInput: string;

    volumeSizeUnitSelect: string;

    volumePathInput: string;

    persistentRadioBtn: string;

    ephemeralRadioBtn: string;

    overQuotaToggle: string;

    templateSection: string;

    runaiGeneralSearch: string;

    jupyterTensorboardCard: string;

    complianceModalCloseButton: string;

    cancelWorkspaceBtn: string;

    leaveModalLeaveBtn: string;

    complianceInfoModalSubtitle: string;

    environmentVarNameInput: string;

    environmentVarValueInput: string;

    submissionAssetsOption: string;

    submissionSection: string;

    ngcImageSelect: string;

    ngcTagSelect: string;

    private readonly hybridForm: HybridForm;

    constructor(page: Page) {
        super(page);

        this.mainPageTab = '[aid="menu-icon-workspaces"]';
        this.addNewWorkspaceBtn = '[aid="primary-cta-btn"]';

        this.runaiGeneralSearch = '.runai-general-search';

        // workspace elements
        this.submissionAssetsOption = '[aid="radio-option-Original"]';
        this.createWorkspaceNameInput = '[aid="workspace-name-input"]';
        this.templateSection = '.template-section';
        this.createWorkspaceBtn = '[aid="submit-btn"]';
        this.cancelWorkspaceBtn = '[aid="cancel-btn"]';
        this.leaveModalLeaveBtn = '[aid="leave-page-modal-cancel-btn"]';
        this.createWorkspaceContinueBtn = '[aid="create-workspace-continue-btn"]';
        this.newEnvironmentBtn = '[aid="add-new-environment-btn"]';
        this.newComputeBtn = '[aid="add-new-compute-resource-btn"]';
        this.customUrlInput = '[aid="custom-url-input"]';
        this.uidInput = '[aid="uid-input"]';
        this.gidInput = '[aid="gid-input"]';
        this.supplementalGrpInput = '[aid="supplemental-groups-input"]';
        this.envSectionMoreSettingsButton = '[aid="env-section-more-settings-btn"]';
        this.envSectionArgsInput = '[aid="args-input"]';
        this.envSectionCommandInput = '[aid="command-input"]';
        this.environmentVarNameInput = '[aid="environment-variable-name"]';
        this.environmentVarValueInput = '[aid="environment-variable-value"]';
        this.annotationNameInput = '[aid="itemized-name-input"]';
        this.annotationValueInput = '[aid="itemized-value-input"]';
        this.generalSection = '[aid="general-section"]';
        this.autoDeletimeTimeSecondsInput = '[aid="seconds-input"]';
        this.autoDeletionTimeDaysInput = '[aid="days-input"]';
        this.terminationGracePeriodSecondsInput = '.termination-grace-period-section [aid="seconds-input"]';
        this.volumeSection = '[aid="volume-section"]';
        this.addVolumeBtn = '[aid="add-volume-btn"]';
        this.volumeStorageClassSelect = '[aid="storage-class-select"]';
        this.volumeAccessModeSelect = '[aid="access-mode-select"]';
        this.volumeSizeInput = '[aid="size-input"]';
        this.volumeSizeUnitSelect = '[aid="unit-select"]';
        this.volumePathInput = '[aid="path-input"]';
        this.persistentRadioBtn = '[aid="radio-option-Persistent"]';
        this.ephemeralRadioBtn = '[aid="radio-option-Ephemeral"]';
        this.jupyterTensorboardCard = '[aid="jupyter-tensorboard-card"]';
        this.complianceModalCloseButton = '[aid=compute-policy-info-modal-close-button]';
        this.complianceInfoModalSubtitle = '[aid="compliance-info-modal-subtitle"]';
        this.overQuotaToggle = '[aid="allow-over-quota-input"]';
        this.submissionSection = '[aid="workload-submission-type-section"]';

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

    async clickOnNewEnvironmentBtn(): Promise<void> {
        return this.v2frame.locator(this.newEnvironmentBtn).click();
    }

    async clickOnNewComputeBtn(): Promise<void> {
        return this.v2frame.locator(this.newComputeBtn).click();
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

    async fillUidInput(txt: number): Promise<void> {
        return this.fillInputBySelector(this.uidInput, txt.toString());
    }

    async fillGidInput(txt: number): Promise<void> {
        return this.fillInputBySelector(this.gidInput, txt.toString());
    }

    async fillSupplementalGrpsInput(txt: string): Promise<void> {
        return this.fillInputBySelector(this.supplementalGrpInput, txt);
    }

    async clickEnvSectionMoreSettings(): Promise<void> {
        return this.v2frame.locator(this.envSectionMoreSettingsButton).click();
    }

    async fillEnvSectionArgsInput(txt: string): Promise<void> {
        return this.fillInputBySelector(this.envSectionArgsInput, txt);
    }

    async fillEnvSectionCommandInput(txt: string): Promise<void> {
        return this.fillInputBySelector(this.envSectionCommandInput, txt);
    }

    async fillAnnotationNameInput(txt: string): Promise<void> {
        return this.fillInputBySelector(this.annotationNameInput, txt);
    }

    async fillAnnotationValueInput(txt: string): Promise<void> {
        return this.fillInputBySelector(this.annotationValueInput, txt);
    }

    async openDataSourceSection(): Promise<void> {
        await this.v2frame.locator(this.dataSourceSection).scrollIntoViewIfNeeded();
        await this.v2frame.locator(this.dataSourceSection).click();
    }

    async openGeneralSection(): Promise<void> {
        await this.v2frame.locator(this.generalSection).click();
    }

    async openVolumeSection(): Promise<void> {
        await this.v2frame.locator(this.volumeSection).click();
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

    async selectDataSourceName(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.dataSourceSection, name);
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

    async addWorkloadVolume(isEphemeral?: boolean): Promise<void> {
        await this.v2frame.locator(this.addVolumeBtn).click();
        await this.selectFirstOption(this.volumeStorageClassSelect);
        await this.selectOptionByValue(this.volumeAccessModeSelect, 'Read-write by one node');
        await this.selectOptionByValue(this.volumeSizeUnitSelect, 'MB');
        const volumePersistentSelector = isEphemeral ? this.ephemeralRadioBtn : this.persistentRadioBtn;
        await this.v2frame.locator(volumePersistentSelector).click();
        await this.fillInputBySelector(this.volumePathInput, '/home');
    }

    async fillGeneralSearchInput(sectionSelector: string, name: string): Promise<void> {
        await this.v2frame.locator(`${sectionSelector} ${this.runaiGeneralSearch}`).click();
        await this.v2frame.locator(`${sectionSelector} ${this.runaiGeneralSearch} input`).fill(name);
    }

    async clickOnCardCompliaceButton(assetName: string, assetKind: AssetKind): Promise<void> {
        await this.v2frame.locator(`[aid="${assetKind}-${assetName}-compliance-btn"]`).first().click();
    }

    async closeComplianceModal(): Promise<void> {
        await this.v2frame.locator(this.complianceModalCloseButton).click();
    }

    async closeWorkspaceForm(): Promise<void> {
        await this.v2frame.locator(this.cancelWorkspaceBtn).click();
        await this.v2frame.locator(this.leaveModalLeaveBtn).click();
    }

    async clickOnAllowOverQuotaToggle(): Promise<void> {
        await this.v2frame.locator(this.overQuotaToggle).click();
    }

    async fillTerminationGracePeriodSeconds(seconds: number): Promise<void> {
        await this.fillInputBySelector(this.terminationGracePeriodSecondsInput, seconds.toString());
    }

    async selectAssetsSubmissionOption(): Promise<void> {
        try {
            await this.v2frame.waitForSelector(this.submissionSection, { timeout: 3000 });
        } catch {
            logger.info('submission section is not visible');
            return;
        }
        return this.v2frame.locator(this.submissionAssetsOption).click();
    }

    async createWorkspaceWithDataSource(
        workspaceName: string,
        projectName: string,
        environmentName: string,
        computeName: string,
        dataSourceName: string
    ): Promise<void> {
        await this.goToPage();
        await this.openWorkspaceForm();
        await this.workloadFormSelectProject(projectName);
        await this.fillWorkspaceName(workspaceName);
        await this.waitForWorkspaceNameCount(workspaceName);
        await this.selectAssetsSubmissionOption();
        await this.clickOnContinue();
        await this.selectEnvironment(environmentName);
        await this.selectComputeResource(computeName);
        await this.openDataSourceSection();
        await this.selectDataSourceName(dataSourceName);
        await this.clickOnCreateWorkspaceBtn();
    }

    async submitWorkspaceWorkloadUI(
        projectName: string,
        workspaceName: string,
        options?: {
            environment?: string;
            computeResource?: string;
            cmd?: string;
            templateName?: string;
            backoffLimit?: number;
        }
    ): Promise<void> {
        const {
            environment = 'tensorboard-tensorflow',
            computeResource = 'one-gpu',
            cmd,
            templateName,
            backoffLimit
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
        await this.selectAssetsSubmissionOption();
        await this.clickOnContinue();

        if (!templateName) {
            await this.selectEnvironment(environment);
            await this.selectComputeResource(computeResource);
        }
        if (cmd) {
            await this.clickEnvSectionMoreSettings();
            await this.fillEnvSectionCommandInput(cmd);
        }
        if (backoffLimit) {
            await this.clickOnGeneralExpansionLabel();
            await this.fillAttemptsInput(backoffLimit);
        }
        await this.clickOnCreateWorkspaceBtn();
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
        await this.selectAssetsSubmissionOption();
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
        await hybridForm.clickOnSubmit();
    }
}
