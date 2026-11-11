/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { WorkloadCreateFormPage } from './WorkloadCreateFormPage';
import { waitUtils, settings, logger } from '../../../helpers';
import { MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import {
    EPersistenceOption,
    IInferenceType,
    IServingPortAccessOptions,
    IVolumeOptions
} from '../../../helpers/models/inferenceBaseFlatUI.models';

export class InferencePage extends WorkloadCreateFormPage {
    tryNewWorkloadBtn: string;

    goToNewExperienceBtn: string;

    // workload table elements
    addNewWorkloadBtn: string;

    addNewInferenceBtn: string;

    createInferenceNameInput: string;

    createInferenceContinueBtn: string;

    newEnvironmentBtn: string;

    newComputeBtn: string;

    createInferenceBtn: string;

    submitInferenceMultiBtns: string;

    advancedSetupBtn: string;

    envSectionMoreSettingsBtn: string;

    computeSectionAutoScaleToggleBtn: string;

    minReplicasInput: string;

    maxReplicasInput: string;

    thresholdMetricOptions: string;

    thresholdValueInput: string;

    scaleDownSelect: string;

    inferenceTypeSectionSummary: string;

    titleInferenceTypeAndBackSection: string;

    vllmOption: string;

    tgiOption: string;

    nimOption: string;

    customOption: string;

    hfModelNameSelector: string;

    clusterDropDown: string;

    nimModelSelector: string;

    servingPortProtocol: string;

    servingPortPort: string;

    nimModelStorageToggle: string;

    servingPortAccessToggle: string;

    servingPortAccessEditBtn: string;

    servingPortAccessPublic: string;

    servingPortAccessEveryone: string;

    servingPortAccessGroup: string;

    servingPortAccessUsers: string;

    servingPortAccessSaveBtn: string;

    envVarValue: string;

    envVarName: string;

    envVarBox: string;

    addEnvVarBtn: string;

    constructor(page: Page) {
        super(page);

        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`;

        this.tryNewWorkloadBtn = '[aid="workload-mode-on-btn"]';
        this.goToNewExperienceBtn = '[aid="go-to-new-experience"]';
        this.addNewWorkloadBtn = '.main-actions [aid="primary-cta-btn"]';
        this.addNewInferenceBtn = '[aid="Inference-btn"]';

        // inference create elements
        this.createInferenceNameInput = '[aid="inference-name-input"]';
        this.createInferenceContinueBtn = '[aid="create-inference-continue-btn"]';
        this.newEnvironmentBtn = '[aid="add-new-environment-btn"]';
        this.clusterDropDown = '[aid="cluster-dropdown"]';
        this.newComputeBtn = '[aid="add-new-compute-resource-btn"]';
        this.createInferenceBtn = '.inference-app [aid="submit-btn"]';
        this.envSectionMoreSettingsBtn = '[aid="env-section-more-settings-btn"]';
        this.computeSectionAutoScaleToggleBtn = '[aid="compute-section-auto-scale-btn"]';
        this.minReplicasInput = '[aid="min-rep-input"]';
        this.maxReplicasInput = '[aid="max-rep-input"]';
        this.thresholdMetricOptions = '[aid="threshold-metric-options"]';
        this.thresholdValueInput = '[aid="threshold-value-input"]';
        this.scaleDownSelect = '[aid="scale-down-select"]';
        this.computeResourceSection = '.compute-resource-section';
        this.inferenceTypeSectionSummary = '[aid="inference-type-section-summary"]';
        this.titleInferenceTypeAndBackSection = '[aid="title-inference-type-and-back-section"]';
        this.vllmOption = '[aid="vllm-option"]';
        this.tgiOption = '[aid="tgi-option"]';
        this.hfModelNameSelector = '[aid="hf-autocomplete-select"]';
        this.nimOption = '[aid="nim-option"]';
        this.customOption = '[aid="custom-option"]';
        this.nimModelSelector = '[aid="nim-autocomplete-select"]';
        this.nimModelStorageToggle = '[aid="nim-model-storage"]';
        this.servingPortAccessToggle = '[aid="serving-endpoint-access-btn"]';
        this.servingPortProtocol = '[aid="connection-protocol"]';
        this.servingPortPort = '[aid="container-port"]';
        this.servingPortAccessEditBtn = '.serving-endpoint-access [aid="edit-access-button"]';
        this.servingPortAccessPublic = '[aid="public"]';
        this.servingPortAccessEveryone = '[aid="everyone"]';
        this.servingPortAccessGroup = '[aid="group"]';
        this.servingPortAccessUsers = '[aid="specific-users"]';
        this.servingPortAccessSaveBtn = '[aid="save-connection-access-role"]';

        this.addEnvVarBtn = '[aid="add-environment-variable-button"]';
        this.envVarValue = '[aid="environment-variable-value"]';
        this.envVarName = '[aid="environment-variable-name"]';
        this.envVarBox = '.environment-variable-box';

        this.advancedSetupBtn = '[aid="runai-dropdown-btn-select-advanced-setup-btn"]';
        this.submitInferenceMultiBtns = '.inference-app [aid="submit-btn"] button';
    }


    // create inference elements

    async selectClusterFromDropDown(clusterName: string): Promise<void> {
        await this.page.waitForSelector(this.clusterDropDown, {
            timeout: 5000
        });
        await this.page.click(this.clusterDropDown, { force: true, timeout: 5000 });
        const timeout: number = settings.isSelfHosted ? 1500 : 4000;
        await waitUtils.sleep(timeout);
        await this.page.click(`.runai-select-item:has-text("${clusterName}")`, {
            force: true,
            timeout: 5000
        });
    }

    async fillInferenceName(name: string): Promise<void> {
        return this.page.locator(this.createInferenceNameInput).fill(name);
    }

    async waitForNameCount(name: string): Promise<void> {
        await this.waitForResponseWithDiagnostics({
            urlContains: '/workloads/count',
            label: 'workloads-count',
            predicate: (response) =>
                response.url().includes('/workloads/count') &&
                response.status() === 200 &&
                response.url().includes(name)
        });
    }

    async clickOnContinue(): Promise<void> {
        return await this.page.locator(this.createInferenceContinueBtn).click();
    }

    async clickOnNewEnvironmentBtn(): Promise<void> {
        return this.page.locator(this.newEnvironmentBtn).click();
    }

    async clickOnNewComputeBtn(): Promise<void> {
        return this.page.locator(this.newComputeBtn).click();
    }

    async clickOnCreateInferenceBtn(): Promise<void> {
        return await this.page.locator(this.createInferenceBtn).click();
    }

    async clickOnAdvancedSetupBtn(): Promise<void> {
        // Wait for submit area to be rendered (same as clickOnSubmitInferenceBtn) so that
        // the continue/advanced-setup buttons are present before we interact with them.
        logger.debug('clickOnAdvancedSetupBtn: waiting 3s for submit area to render');
        await this.page.waitForTimeout(3000);
        const submitButtons = this.locator(this.submitInferenceMultiBtns);
        logger.debug('clickOnAdvancedSetupBtn: waiting for at least one submit button to be visible');
        await submitButtons.first().waitFor({ state: 'visible', timeout: 10000 });
        const btns = await submitButtons.all();
        logger.debug(`clickOnAdvancedSetupBtn: found ${btns.length} submit button(s) for selector "${this.submitInferenceMultiBtns}"`);
        if (btns.length === 0) {
            throw new Error(
                `clickOnAdvancedSetupBtn: no submit buttons found for selector "${this.submitInferenceMultiBtns}"`
            );
        }
        if (btns.length > 1) {
            logger.debug('clickOnAdvancedSetupBtn: clicking dropdown (btns[1]) then advanced setup option');
            await btns[1].click();
            await this.locator(this.advancedSetupBtn).click();
            logger.debug('clickOnAdvancedSetupBtn: advanced setup option clicked');
        } else {
            logger.debug('clickOnAdvancedSetupBtn: clicking single continue button (btns[0])');
            await btns[0].click();
            logger.debug('clickOnAdvancedSetupBtn: continue button clicked');
        }
    }

    async clickOnAutoScaleToggleBtn(): Promise<void> {
        await this.page.locator(this.computeSectionAutoScaleToggleBtn).click();
    }

    async clickOnEnvSectionMoreSettingsBtn(): Promise<void> {
        await this.page.locator(this.envSectionMoreSettingsBtn).click();
    }

    async fillMinReplicasInput(num: number): Promise<void> {
        return this.fillInputBySelector(this.minReplicasInput, num.toString());
    }

    async fillMaxReplicasInput(num: number): Promise<void> {
        return this.fillInputBySelector(this.maxReplicasInput, num.toString());
    }

    async fillThresholdValueInput(num: number): Promise<void> {
        return this.fillInputBySelector(this.thresholdValueInput, num.toString());
    }

    async clickOnScaleDownSelect(): Promise<void> {
        return await this.page.locator(this.scaleDownSelect).click();
    }

    async selectAfter1Minute(): Promise<void> {
        await this.clickOnScaleDownSelect();
        return await this.page.locator('div[role="option"]:has-text("After 1 minute of inactivity")').click();
    }

    async selectTemplate(templateName: string): Promise<void> {
        await this.fillGeneralSearchInput(this.templateSection, templateName);
        await this.selectCard(templateName);
    }

    async selectDistributed(): Promise<void> {
        await this.page.locator(this.distributedOption).click();
    }

    async fillTheFirstStep(
        clusterName: string,
        projectName: string,
        workloadName: string,
        inferenceType?: IInferenceType,
        templateName?: string,
        distributed?: boolean
    ): Promise<void> {
        await this.selectClusterFromDropDown(clusterName);
        await this.selectProjectIfNeeded(projectName);

        if (distributed) {
            await this.selectDistributed();
        } else if (templateName) {
            await this.selectTemplate(templateName);
        } else {
            // Fallback case - selects the template "start from scratch" if it exists
            // and no template is provided and there are multiple templates
            await this.selectTemplateIfNeeded();
        }

        if (inferenceType) {
            await this.selectInferenceTypeOption(inferenceType);
        }

        await this.fillInferenceNameAndCount(workloadName);
        await this.clickOnContinue();
    }

    async fillInferenceNameAndCount(workloadName: string): Promise<void> {
        await this.fillInferenceName(workloadName);
        await this.waitForNameCount(workloadName);
    }

    async selectInferenceTypeOption(inferenceType: IInferenceType): Promise<void> {
        switch (inferenceType) {
            case IInferenceType.Custom:
                await this.locator(this.customOption).click();
                break;
            case IInferenceType.VLLM:
                await this.locator(this.vllmOption).click();
                break;
            case IInferenceType.TGI:
                await this.locator(this.tgiOption).click();
                break;
            case IInferenceType.Nim:
                await this.locator(this.nimOption).click();
                break;
        }
    }

    async openServingPortSection(): Promise<void> {
        await this.locator(this.servingPortAccessToggle).click();
        await this.page.waitForTimeout(1000);
    }

    async selectServingPortAccess(servingPortAccessOption: IServingPortAccessOptions): Promise<void> {
        await this.locator(this.servingPortAccessEditBtn).click();
        switch (servingPortAccessOption) {
            case IServingPortAccessOptions.Public:
                await this.locator(this.servingPortAccessPublic).click();
                break;
            case IServingPortAccessOptions.Everyone:
                await this.locator(this.servingPortAccessEveryone).click();
                break;
            case IServingPortAccessOptions.Group:
                await this.locator(this.servingPortAccessGroup).click();
                break;
            case IServingPortAccessOptions.Users:
                await this.locator(this.servingPortAccessUsers).click();
                break;
        }

        await this.locator(this.servingPortAccessSaveBtn).click();
    }

    private async selectStorageClassAndWaitForConfiguration(): Promise<void> {
        try {
            logger.debug(
                'setWorkloadVolume: waiting for storage-class-configuration response after selecting storage class'
            );
            const storageClassResponsePromise = this.page.waitForResponse(
                (response) =>
                    response.url().includes('api/v1/storage-class-configuration') &&
                    response.status() === 200,
                { timeout: 15000 }
            );
            await this.selectFirstOption(this.volumeStorageClassSelect);

            await storageClassResponsePromise;
            logger.debug('setWorkloadVolume: storage-class-configuration response received (200)');
        } catch (error) {
            logger.error(
                `setWorkloadVolume: failed waiting for storage-class-configuration response: ${error?.message ?? error}`
            );
            throw error;
        }
    }

    async setWorkloadVolume(volumeOptions: IVolumeOptions): Promise<void> {
        await this.selectStorageClassAndWaitForConfiguration();
        await this.page.locator(this.volumeAccessModeSelect).waitFor({ state: 'visible' });
        await this.selectOptionByValue(this.volumeAccessModeSelect, volumeOptions.accessMode);
        await this.selectOptionByValue(this.volumeSizeUnitSelect, volumeOptions.sizeUnit);
        await this.fillInputBySelector(this.volumePathInput, volumeOptions.path);
        if (volumeOptions.persistenceOption) {
            const volumePersistentSelector =
                volumeOptions.persistenceOption === EPersistenceOption.Ephemeral
                    ? this.ephemeralRadioBtn
                    : this.persistentRadioBtn;
            await this.locator(volumePersistentSelector).click();
        }
    }

    async fillEnvironmentVariablesValues(values: Array<string>): Promise<void> {
        const elements = await this.locator(this.envVarValue).all();
        if (elements.length !== values.length) {
            logger.warn(
                `Mismatch in environment variable values: Expected ${values.length} elements, but found ${elements.length}.`
            );
        }

        for (const [index, element] of Array.from(elements.entries())) {
            if (index < values.length) {
                await element.fill(values[index]);
            }
        }
    }

    async fillEnvironmentVariablesValuesByName(envVars: Map<string, string>): Promise<void> {
        const elements = await this.locator(this.envVarBox).all();
        if (elements.length === 0) {
            logger.warn('No environment variable boxes found');
        }

        for (const [index, element] of Array.from(elements.entries())) {
            const elementName = await element.locator(this.envVarName).inputValue();
            logger.debug(`elementName: ${elementName}`);
            const elementValue = element.locator(this.envVarValue);
            if (envVars.has(elementName)) {
                await elementValue.fill(envVars.get(elementName));
            }
        }
    }

    async clickOnAddEnvironmentVariable(): Promise<void> {
        await this.page.locator(this.addEnvVarBtn).click();
    }

    async addEnvironmentVariables(
        envVars: Array<{ name: string; source?: string; value: string }>
    ): Promise<void> {
        // Get the count of existing env var boxes before adding
        const existingCount = await this.locator(this.envVarBox).count();

        // Click add button for each new env var
        for (const _ of envVars) {
            await this.clickOnAddEnvironmentVariable();
            await this.page.waitForTimeout(200); // small wait for UI to update
        }

        // Get all env var boxes and fill only the newly added ones
        const allBoxes = await this.locator(this.envVarBox).all();
        const newBoxes = allBoxes.slice(existingCount);

        for (const [index, box] of newBoxes.entries()) {
            const envVar = envVars[index];

            // Fill name
            await box.locator(this.envVarName).fill(envVar.name);

            // Select source if provided (default is "Custom")
            if (envVar.source && envVar.source !== 'Custom') {
                // await box.locator('[aid="environment-variable-source"]').click();
                // await this.selectOptionFromDropdown(envVar.source);
                await this.selectOptionByValue('[aid="environment-variable-source"]', envVar.source);
            }

            // Fill value
            await box.locator(this.envVarValue).fill(envVar.value);
        }
    }
}
