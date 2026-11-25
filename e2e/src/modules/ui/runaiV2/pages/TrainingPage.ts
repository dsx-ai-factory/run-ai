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

    private readonly volumeSizeInput: string;

    private readonly submitButton: string;

    private readonly createTrainingButton: string;

    private readonly advancedSetupButton: string;

    private readonly generalSection: string;

    private readonly multiNodeOption: string;

    private readonly singleNodeOption: string;

    private readonly distributedToolsSelect: string;

    private readonly multiNodeWorkersAndMasterOption: string;

    private readonly multiNodeWorkersOnlyOption: string;

    private readonly templateSection: string;

    jupyterTensorboardCard: string;

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

        // multi node elements
        this.multiNodeOption = '[aid=radio-option-Distributed]';
        this.singleNodeOption = '[aid=radio-option-Standard]';
        this.distributedToolsSelect = '[aid="dist-framework-select"]';
        this.multiNodeWorkersAndMasterOption = 'div[role="radio"]:has-text("Workers & master")';
        this.multiNodeWorkersOnlyOption = 'div[role="radio"]:has-text("Workers only")';

        // data source elements

        // compute resource elements

        // environment elements
        this.jupyterTensorboardCard = '[aid="jupyter-tensorboard-card"]';

        // volume elements
        this.volumeSizeInput = '[aid="size-input"]';

        //general section elements
        this.generalSection = '[aid="general-section"]';

        this.hybridForm = new HybridForm(page);
        this.qTable = new QTable(page);
    }

    async selectComputeResource(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.computeResourceSection, name);
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

    async clickOnSubmitBtn(): Promise<void> {
        return this.v2frame.locator(this.submitButton).click();
    }

    async waitForSubmitButton(): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.submitButton);
        await this.page.waitForTimeout(1000);
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
        await this.fillFlexTrainingFirstPage(projectName, workloadName);
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

    /**
     *
     * @param workloadName
     * @returns an array with all the table header and cell value pairs of the entity
     */

    async getWorkloadData(workloadName: string): Promise<Record<string, string>> {
        return await this.qTable.getEntityData('Workload', workloadName);
    }
}
