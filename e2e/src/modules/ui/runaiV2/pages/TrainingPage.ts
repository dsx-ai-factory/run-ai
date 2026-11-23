/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { logger, objectGenerator, waitUtils, flowRootState } from '../../../helpers';
import { WorkloadPage } from './WorkloadPage';
import { HybridForm } from './HybridSubmissionForm';
import { QTable } from '../components/QTable';

export const enum DISTRIBUTED_TOOLS_LABELS {
    PY_TORCH = 'PyTorch',
    TENSOR_FLOW = 'TensorFlow',
    XG_BOOST = 'XG Boost',
    MPI = 'MPI',
    JAX = 'JAX'
}

interface FillHybridSecondPageOptions {
    command?: string;
    args?: string;
    computeLabel?: string;
    pvcName?: string;
    workersCount?: string;
    toolName?: string;
    imagePullPolicy?: string;
    envVars?: Array<{
        name: string;
        sourcetype: 'Custom' | 'Secret' | 'ConfigMap';
        sourcename?: string;
        sourcevalue: string;
    }>;
}

export class TrainingPage extends WorkloadPage {
    readonly createNameInput: string;

    private readonly createTrainingContinueBtn: string;

    private readonly newEnvironmentBtn: string;

    private readonly newComputeBtn: string;

    private readonly addVolumeBtn: string;

    private readonly volumeStorageClassSelect: string;

    private readonly volumeAccessModeSelect: string;

    private readonly ephemeralRadioBtn: string;

    private readonly volumeSizeInput: string;

    private readonly volumeSizeUnitSelect: string;

    private readonly volumePathInput: string;

    private readonly persistentRadioBtn: string;

    private readonly volumeSection: string;

    private readonly submitButton: string;

    private readonly createTrainingButton: string;

    private readonly advancedSetupButton: string;

    private readonly addNewPvcBtn: string;

    private readonly addNewNFSBtn: string;

    private readonly addNewHostPathBtn: string;

    private readonly containerPathOverrideInput: string;

    private readonly containerPathOverrideSection: string;

    private readonly selectEnvironmentCard: string;

    private readonly computeResourceClass: string;

    private readonly generalSection: string;

    private readonly labelBtn: string;

    private readonly annotationsBtn: string;

    private readonly multiNodeOption: string;

    private readonly singleNodeOption: string;

    private readonly distributedToolsSelect: string;

    private readonly multiNodeWorkersAndMasterOption: string;

    private readonly multiNodeWorkersOnlyOption: string;

    private readonly editMasterToggle: string;

    private readonly templateSection: string;

    private readonly envSectionMoreSettingBtn: string;

    private readonly addEnvarBtn: string;

    private readonly envarNameOption: string;

    private readonly envarSourceSection: string;

    private readonly envarCredentialSection: string;

    private readonly envarKeyOption: string;

    private readonly envVarCredentialKeySelect: string;

    private readonly envVarCredentialKeyInput: string;

    private readonly submissionAssetsOption: string;

    jupyterTensorboardCard: string;

    private readonly submissionSection: string;

    private readonly hybridForm: HybridForm;

    private readonly qTable: QTable;

    constructor(page: Page) {
        super(page);
        // main training page
        this.mainPageTab = '[aid="menu-icon-trainings"]';

        // create training form
        this.createNameInput = '[aid="training-name-input"]';
        this.createTrainingContinueBtn = '[aid="create-training-continue-btn"]';
        this.submitButton = '[aid="submit-btn"]';
        this.templateSection = '.template-section';
        this.submissionAssetsOption = '[aid="radio-option-Original"]';

        // multi node elements
        this.multiNodeOption = '[aid=radio-option-Distributed]';
        this.singleNodeOption = '[aid=radio-option-Standard]';
        this.distributedToolsSelect = '[aid="dist-framework-select"]';
        this.multiNodeWorkersAndMasterOption = 'div[role="radio"]:has-text("Workers & master")';
        this.multiNodeWorkersOnlyOption = 'div[role="radio"]:has-text("Workers only")';
        this.editMasterToggle = '[aid="edit-master-toggle"]';

        // data source elements
        this.addNewPvcBtn = '[aid=PVC-btn]';
        this.addNewNFSBtn = '[aid=NFS-btn]';
        this.addNewHostPathBtn = '[aid="Host path-btn"]';
        this.containerPathOverrideInput = '[aid="container-path-input"]';
        this.containerPathOverrideSection = '.container-path-override';

        // compute resource elements
        this.newComputeBtn = '[aid="add-new-compute-resource-btn"]';
        this.computeResourceClass = '[class="runai-select-card compute-resource-card"]';

        // environment elements
        this.newEnvironmentBtn = '[aid="add-new-environment-btn"]';
        this.selectEnvironmentCard = '[class="runai-select-card environment-card"]';
        this.jupyterTensorboardCard = '[aid="jupyter-tensorboard-card"]';
        this.envSectionMoreSettingBtn = '[aid="env-section-more-settings-btn"]';
        this.addEnvarBtn = '[aid="add-environment-variable-button"]';
        this.envarSourceSection = '[aid="environment-variable-source"]';
        this.envarNameOption = '[aid="environment-variable-name"]';
        this.envarCredentialSection = '[aid="environment-variable-credential"]';
        this.envVarCredentialKeySelect = '[aid="environment-variable-key-select"]';
        this.envVarCredentialKeyInput = '[aid="environment-variable-key-input"]';

        // volume elements
        this.addVolumeBtn = '[aid="add-volume-btn"]';
        this.persistentRadioBtn = '[aid="radio-option-Persistent"]';
        this.ephemeralRadioBtn = '[aid="radio-option-Ephemeral"]';
        this.volumeStorageClassSelect = '[aid="storage-class-select"]';
        this.volumeAccessModeSelect = '[aid="access-mode-select"]';
        this.volumeSizeInput = '[aid="size-input"]';
        this.volumeSizeUnitSelect = '[aid="unit-select"]';
        this.volumePathInput = '[aid="path-input"]';
        this.volumeSection = '[aid="volume-section"]';

        //general section elements
        this.generalSection = '[aid="general-section"]';
        this.annotationsBtn = 'button:has-text("+ annotation")';
        this.labelBtn = 'button:has-text("+ label")';
        this.submissionSection = '[aid="workload-submission-type-section"]';

        this.hybridForm = new HybridForm(page);
        this.qTable = new QTable(page);
    }

    async selectFirstEnvironment(): Promise<void> {
        return this.v2frame.locator(this.selectEnvironmentCard).first().click();
    }

    async selectFirstComputeResource(): Promise<void> {
        return this.v2frame.locator(this.computeResourceClass).first().click();
    }

    async selectComputeResource(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.computeResourceSection, name);
        return this.v2frame.locator(`[aid="${name}-card"]`).click();
    }

    async selectDataSource(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.dataSourceSection, name);
        return this.v2frame.locator(`[aid="${name}-card"]`).click();
    }

    async fillTrainingName(name: string): Promise<void> {
        return this.v2frame.locator(this.createNameInput).fill(name);
    }

    async focusTrainingName(): Promise<void> {
        return this.page.locator(this.createNameInput).focus();
    }

    async pressEnterOnTrainingName(): Promise<void> {
        return this.pressKeyOnElement(this.createNameInput, 'Enter');
    }

    async pressEscapeOnTrainingName(): Promise<void> {
        return this.pressKeyOnElement(this.createNameInput, 'Escape');
    }

    searchProject(name: string): Promise<void> {
        return this.v2frame.locator(this.createNameInput).fill(name);
    }

    async clickOnContinueBtn(options = {}): Promise<void> {
        if (await this.hybridForm.isQuickSubmitDropdownVisible()) {
            return this.hybridForm.continueWithQuickSubmit();
        }

        await waitUtils.waitForElementToExist(this, this.createTrainingContinueBtn);
        return this.v2frame.locator(this.createTrainingContinueBtn).click(options);
    }

    async clickOnNewEnvironmentBtn(): Promise<void> {
        return this.v2frame.locator(this.newEnvironmentBtn).click();
    }

    async clickOnNewComputeBtn(): Promise<void> {
        return this.v2frame.locator(this.newComputeBtn).click();
    }

    async clickOnSubmitBtn(): Promise<void> {
        return this.v2frame.locator(this.submitButton).click();
    }

    async waitForSubmitButton(): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.submitButton);
        await this.page.waitForTimeout(1000);
    }

    async addVolume(isEphemeral?: boolean): Promise<void> {
        await this.v2frame.locator(this.volumeSection).click();
        await this.clickOnVolumeButton();
        await this.selectFirstOption(this.volumeStorageClassSelect);
        await this.selectOptionByValue(this.volumeAccessModeSelect, 'Read-write by one node');
        await this.selectOptionByValue(this.volumeSizeUnitSelect, 'MB');
        const volumePersistentSelector = isEphemeral ? this.ephemeralRadioBtn : this.persistentRadioBtn;
        await this.v2frame.locator(volumePersistentSelector).click();
        await this.fillInputBySelector(
            this.volumePathInput,
            objectGenerator.generateShortNameByPrefix('/test')
        );
    }

    async clickOnAddNewPvcBtn(): Promise<void> {
        return this.v2frame.locator(this.addNewPvcBtn).click();
    }

    async clickOnAddNewNFSBtn(): Promise<void> {
        return this.v2frame.locator(this.addNewNFSBtn).click();
    }

    async clickOnAddNewHostPathBtn(): Promise<void> {
        return this.v2frame.locator(this.addNewHostPathBtn).click();
    }

    async clickOnNewDataSourceButton(): Promise<void> {
        return this.v2frame.locator('[aid="data-source-gallery-add-new-btn"]').click();
    }

    async clickOnVolumeButton(): Promise<void> {
        return this.v2frame.locator(this.addVolumeBtn).click();
    }

    async openDataSourceSection(): Promise<void> {
        await this.v2frame.locator(this.dataSourceSection).scrollIntoViewIfNeeded();
        await this.v2frame.locator(this.dataSourceSection).click();
    }

    async addNewPvc(): Promise<void> {
        await this.openDataSourceSection();
        await this.clickOnNewDataSourceButton();
        await this.clickOnAddNewPvcBtn();
    }

    async addNewNFS(): Promise<void> {
        await this.openDataSourceSection();
        await this.clickOnNewDataSourceButton();
        await this.clickOnAddNewNFSBtn();
    }

    async addNewHostPath(): Promise<void> {
        await this.openDataSourceSection();
        await this.clickOnNewDataSourceButton();
        await this.clickOnAddNewHostPathBtn();
    }

    async clickOnAddAnnotationsBtn(): Promise<void> {
        return this.v2frame.locator(this.annotationsBtn).click();
    }

    async clickOnAddLabelBtn(): Promise<void> {
        return this.v2frame.locator(this.labelBtn).click();
    }

    async clickOnEnvMoreSetting(): Promise<void> {
        return this.v2frame.locator(this.envSectionMoreSettingBtn).click();
    }

    async clickOnAddEnvVar(): Promise<void> {
        return this.v2frame.locator(this.addEnvarBtn).click();
    }

    async fillLastEnvarName(name: string): Promise<void> {
        // To ensure that selecting the new entry in envar use the last
        await this.v2frame.locator(this.envarNameOption).last().fill(name);
    }

    async selectLastEnvarSource(source: string): Promise<void> {
        await this.selectLastOptionByValue(this.envarSourceSection, source);
    }

    async selectLatEnvarCredential(credentialName: string): Promise<void> {
        // After selecting the envar source to be credentail, the envarCredentialSection needs to load the credentails
        // however, there isn't a clean way to make sure that the component is loaded.
        // therefore we sleep for half a sec.
        await waitUtils.sleep(500);
        await this.selectLastOptionByValue(this.envarCredentialSection, credentialName);
    }

    async selectOrFillLastEnVarCredentialKey(key: string): Promise<void> {
        // To ensure that selecting the new entry in envar use the last
        const isSelectVisible = await this.v2frame.locator(this.envVarCredentialKeySelect).isVisible();
        const isInputVisible = await this.v2frame.locator(this.envVarCredentialKeyInput).isVisible();
        if (isSelectVisible) {
            await this.v2frame.locator(this.envVarCredentialKeySelect).last().click();
            await this.v2frame.locator(`div[role="option"]:has-text("${key}")`).click();
        } else if (isInputVisible) {
            await this.v2frame.locator(this.envVarCredentialKeyInput).last().fill(key);
        }
    }

    async fillLastEnvarAsCredential(name: string, credentialName: string, key: string): Promise<void> {
        await this.fillLastEnvarName(name);
        await this.selectLastEnvarSource('Credentials');
        await this.selectLatEnvarCredential(credentialName);
        await this.selectOrFillLastEnVarCredentialKey(key);
    }

    async waitForTrainingNameCount(name: string): Promise<void> {
        await this.waitForWorkloadsCountResponse({ nameInUrl: name });
    }

    async waitForTrainingStatus(workspaceName: string, status = 'Active'): Promise<void> {
        await this.v2frame.waitForSelector(
            `td:has-text("${status}"):right-of(td:has-text("${workspaceName}"))`,
            { timeout: 60000 }
        );
    }

    async selectedMultiNodeOption(): Promise<void> {
        return this.v2frame.locator(this.multiNodeOption).click();
    }

    async selectedSingleNodeOption(): Promise<void> {
        return this.v2frame.locator(this.singleNodeOption).click();
    }

    async selectedDistributedTool(toolName: DISTRIBUTED_TOOLS_LABELS): Promise<void> {
        await this.v2frame.locator(this.distributedToolsSelect).click();
        await this.v2frame.locator(`div[role="option"]:has-text("${toolName}")`).click();
    }

    async selectedWorkersAndMasterOption(): Promise<void> {
        return this.v2frame.locator(this.multiNodeWorkersAndMasterOption).click();
    }

    async selectedWorkersOnlyOption(): Promise<void> {
        return this.v2frame.locator(this.multiNodeWorkersOnlyOption).click();
    }

    async clickOnDistributedContinueBtn(): Promise<void> {
        return this.v2frame.locator(this.submitButton).click();
    }

    async openGeneralSection(): Promise<void> {
        return this.v2frame.locator(this.generalSection).click();
    }

    async clickOnEditMasterToggle(): Promise<void> {
        return this.v2frame.locator(this.editMasterToggle).click();
    }

    async fillContainerPathOverrideInput(containerPathOverride: string): Promise<void> {
        return this.v2frame.locator(this.containerPathOverrideInput).fill(containerPathOverride);
    }

    async isContainerPathOverrideSectionExists(): Promise<boolean> {
        return this.v2frame.locator(this.containerPathOverrideInput).isVisible();
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

    async fillFirstNewTrainingPage(
        projectName: string,
        trainingName: string,
        isMultiNode = false,
        toolName?: DISTRIBUTED_TOOLS_LABELS,
        configureMaster?: boolean,
        flexisbleSubmission?: boolean,
        templateName?: string
    ): Promise<void> {
        await this.navigateToWorkloads();
        await this.openTrainingForm();
        await this.selectClusterFromDropDown(flowRootState.cluster.name);
        await this.workloadFormSelectProject(projectName);
        if (isMultiNode) {
            await waitUtils.sleep(1000);
            await this.selectedMultiNodeOption();
        }
        if (templateName) {
            await this.workloadFormSelectTemplate(templateName);
        } else {
            await this.workloadFormSelectTemplate();
        }
        if (isMultiNode) {
            await this.selectedDistributedTool(toolName);
            if (!configureMaster) await this.selectedWorkersOnlyOption();
        }
        await this.fillTrainingName(trainingName);
        await this.waitForTrainingNameCount(trainingName);

        if (!flexisbleSubmission) {
            await this.selectAssetsSubmissionOption();
        }
        await this.clickOnContinueBtn();
    }

    async fillFlexTrainingFirstPage(
        projectName: string,
        trainingName: string,
        isMultiNode = false,
        toolName?: DISTRIBUTED_TOOLS_LABELS,
        configureMaster?: boolean,
        templateName?: string,
        templateAdvancedSetup: boolean = true
    ): Promise<void> {
        await this.navigateToWorkloads();
        await this.openTrainingForm();
        await this.selectClusterFromDropDown(flowRootState.cluster.name);
        await this.workloadFormSelectProject(projectName);

        if (isMultiNode) {
            await this.selectedMultiNodeOption();
        }

        // need to select template first before doing distributed settings
        if (templateName) {
            await this.workloadFormSelectTemplateV2(templateName);
            await this.page.waitForTimeout(500);
        } else {
            await this.workloadFormSelectTemplateV2(undefined, false);
        }

        // for distributed workload, if 'start from scratch' template is selected, then need to choose framework and choose dist workload config
        // else only need to choose dist workload config (framework dropdown is disabled in this case)
        if (isMultiNode) {
            // distributed template is chosen
            if (templateName) {
                if (
                    (toolName === DISTRIBUTED_TOOLS_LABELS.PY_TORCH ||
                        toolName === DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW) &&
                    !configureMaster
                ) {
                    await this.selectedWorkersOnlyOption();
                }
            } else {
                await this.fillMultiNodeOnTrainingFirstPage(toolName, configureMaster);
            }
        }

        await this.fillTrainingName(trainingName);
        await this.waitForTrainingNameCount(trainingName);

        if (!templateName) {
            await this.clickOnContinueBtn();
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

    async fillMultiNodeOnTrainingFirstPage(
        toolName: DISTRIBUTED_TOOLS_LABELS,
        configureMaster: boolean
    ): Promise<void> {
        await this.selectedDistributedTool(toolName);
        if (
            (toolName === DISTRIBUTED_TOOLS_LABELS.PY_TORCH ||
                toolName === DISTRIBUTED_TOOLS_LABELS.TENSOR_FLOW) &&
            !configureMaster
        ) {
            await this.selectedWorkersOnlyOption();
        }
    }

    async createFlexibleTraining(
        projectName: string,
        workloadName: string,
        dsConfig?: {
            dsName: string;
            containerPath?: string;
        },
        envVars?: Array<{
            name: string;
            sourcetype: 'Custom' | 'Secret' | 'ConfigMap';
            sourcename?: string;
            sourcevalue: string;
        }>,
        environment?: string,
        computeResource?: string
    ): Promise<void> {
        await this.fillFirstNewTrainingPage(projectName, workloadName, false, undefined, undefined, true);
        await this.hybridForm.fillHybridFormByLoadingExistingConfig(
            environment || 'tensorboard-tensorflow',
            computeResource || 'cpu-only',
            dsConfig,
            envVars
        );
        await this.clickOnSubmitBtn();
    }

    async fillHybridDistributedTrainingCommonPartSecondPage(
        image: string,
        opts: FillHybridSecondPageOptions
    ): Promise<void> {
        await this.hybridForm.fillImageSection(image, opts.imagePullPolicy);
        if (opts.toolName) {
            await this.hybridForm.fillToolsSection(opts.toolName);
        }
        if (opts.command || opts.args) {
            await this.hybridForm.fillRuntimeSettings(opts.command, opts.args);
        }
        if (opts.pvcName) {
            await this.hybridForm.fillPvcSection(opts.pvcName);
        }
        if (opts.computeLabel || opts.workersCount) {
            await this.hybridForm.fillComputeSection(opts.computeLabel, Number(opts.workersCount));
        }
        if (opts.envVars?.length) {
            await this.hybridForm.fillEnvironmentVariables(opts.envVars);
        }
    }

    async fillHybridDistributedTrainingSecondPage(
        image: string,
        opts: FillHybridSecondPageOptions,
        masterOpts?: {
            image?: string;
            opts?: FillHybridSecondPageOptions;
        },
        templateName?: string
    ): Promise<void> {
        if (templateName) {
            await this.hybridForm.clickOnSubmit();
        } else {
            await this.fillHybridDistributedTrainingCommonPartSecondPage(image, opts);
            await this.hybridForm.clickOnSubmit();
            if (masterOpts?.opts) {
                await waitUtils.waitForElementToExist(this, this.hybridForm.distributedMasterFormWrapper);
                await this.page.waitForTimeout(1000);
                await this.fillHybridDistributedTrainingCommonPartSecondPage(
                    masterOpts.image || image,
                    masterOpts.opts
                );
            }
            await this.hybridForm.waitForDistributedMasterSubmitButton();
            await this.hybridForm.clickOnSubmitMasterForm();
            await this.waitForSubmitButton();
            await this.hybridForm.clickOnSubmit(); // setup step
        }
    }

    async submitTrainingWorkloadUI(
        projectName: string,
        trainingName: string,
        options?: {
            environment?: string;
            computeResource?: string;
            templateName?: string;
            dataSource?: {
                dsName: string;
                containerPath: string;
            };
        }
    ): Promise<void> {
        const {
            environment = 'tensorboard-tensorflow',
            computeResource = 'one-gpu',
            templateName,
            dataSource
        } = options || {};

        await this.navigateToWorkloads();
        await this.openTrainingForm();
        await this.selectClusterFromDropDown(flowRootState.cluster.name);
        await this.workloadFormSelectProject(projectName);

        if (templateName) {
            await this.workloadFormSelectTemplate(templateName);
        } else {
            await this.workloadFormSelectTemplate();
        }

        await this.fillTrainingName(trainingName);
        await this.selectAssetsSubmissionOption();
        await this.clickOnContinueBtn();

        if (!templateName) {
            await this.selectEnvironment(environment);
            await this.selectComputeResource(computeResource);
        }
        if (dataSource) {
            await this.openDataSourceSection();
            await this.selectDataSource(dataSource.dsName);
            if (dataSource.containerPath) {
                await this.fillContainerPathOverrideInput(dataSource.containerPath);
            }
        }
        await this.clickOnSubmitBtn();
    }

    /**
     *
     * @param workloadName
     * @returns an array with all the table header and cell value pairs of the entity
     */

    async getWorkloadData(workloadName: string): Promise<Record<string, string>> {
        return await this.qTable.getEntityData('Workload', workloadName);
    }
}
