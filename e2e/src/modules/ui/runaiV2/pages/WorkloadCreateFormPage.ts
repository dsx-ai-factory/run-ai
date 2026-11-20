/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { BaseV2Page } from './BaseV2Page';
import { logger, Utils } from '../../../helpers';

export abstract class WorkloadCreateFormPage extends BaseV2Page {
    projectSection: string;

    templateSection: string;

    distributedOption: string;

    computeResourceSection: string;

    environmentSection: string;

    runaiGeneralSearch: string;

    private readonly submissionSection: string;

    private readonly submissionAssetsOption: string;

    modelSection: string;

    dataSourceNewBtn: string;

    volumeSection: string;

    addVolumeBtn: string;

    volumeStorageClassSelect: string;

    volumeAccessModeSelect: string;

    volumeSizeInput: string;

    volumeSizeUnitSelect: string;

    volumePathInput: string;

    persistentRadioBtn: string;

    ephemeralRadioBtn: string;

    credentialSearchCardInput: string;

    generalSection: string;

    addNewAnnotation: string;

    addNewLabel: string;

    constructor(page: Page) {
        super(page);

        // first step
        this.projectSection = '.project-section';
        this.templateSection = '.template-section';
        this.submissionSection = '[aid="workload-submission-type-section"]';
        this.submissionAssetsOption = '[aid="radio-option-Original"]';
        this.distributedOption = '[aid="radio-option-Distributed"]';

        // second step
        this.computeResourceSection = '.compute-resource-section';
        this.environmentSection = '.environment-section';
        this.modelSection = '.model-catalog-section';
        this.runaiGeneralSearch = '.runai-general-search';
        this.dataSourceNewBtn = '[aid="data-source-gallery-add-new-btn"]';

        // volume section
        this.volumeSection = '[aid="volume-section"]';
        this.addVolumeBtn = '[aid="add-volume-btn"]';
        this.volumeStorageClassSelect = '[aid="storage-class-select"]';
        this.volumeAccessModeSelect = '[aid="access-mode-select"]';
        this.volumeSizeInput = '[aid="size-input"]';
        this.volumeSizeUnitSelect = '[aid="unit-select"]';
        this.volumePathInput = '[aid="path-input"]';
        this.persistentRadioBtn = '[aid="radio-option-Persistent"]';
        this.ephemeralRadioBtn = '[aid="radio-option-Ephemeral"]';

        // credentials
        this.credentialSearchCardInput = '.credential-gallery';

        // general section
        this.generalSection = '[aid="general-section"]';
        this.addNewAnnotation = '[aid="add-item-button"]';
        this.addNewLabel = '[aid="add-item-button"]';
    }

    async selectProjectIfNeeded(projectName: string): Promise<void> {
        // fallback case - when there are more than one project
        await this.fillGeneralSearchInput(this.projectSection, projectName);
        const isProjectSelected = await this.isCardSelected(projectName);
        if (!isProjectSelected) {
            logger.warn(
                `Project ${projectName} is not selected, selecting the project. This can happen when there are more than one project`
            );
            await this.selectCard(projectName);
        }
    }

    async isTemplateExists(): Promise<boolean> {
        return await this.page.locator(this.templateSection).isVisible();
    }

    async selectTemplateIfNeeded(): Promise<void> {
        if (!(await this.isTemplateExists())) return;
        const isTemplateSelected = await this.isCardSelected('create-new');
        if (!isTemplateSelected) {
            logger.info('Template "start from scratch" is not selected, selecting the template');
            await this.selectCard('create-new');
        }
    }

    async selectEnvironment(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.environmentSection, name);
        await this.selectCard(name);
    }

    async selectComputeResource(name: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.computeResourceSection, name);
        await this.selectCard(name);
    }

    async selectModelInInferenceCreatePage(name: string): Promise<void> {
        await this.fillGeneralSearchInput(this.modelSection, name);
        await this.selectCard(name);
    }

    async fillGeneralSearchInput(sectionSelector: string, name: string): Promise<void> {
        await this.locator(`${sectionSelector} ${this.runaiGeneralSearch}`).click();
        await this.locator(`${sectionSelector} ${this.runaiGeneralSearch} input`).fill(name);
    }

    async fillWorkloadFormGeneralSearchInput(sectionSelector: string, name: string): Promise<void> {
        await this.locator(`${sectionSelector} ${this.runaiGeneralSearch}`).click();
        await this.locator(`${sectionSelector} ${this.runaiGeneralSearch} input`).fill(name);
    }

    async selectCredentials(credentialsName: string): Promise<void> {
        await this.fillWorkloadFormGeneralSearchInput(this.credentialSearchCardInput, credentialsName);
        await this.selectCard(credentialsName);
    }

    async selectCard(name: string): Promise<void> {
        await this.clickOnElementByAidName(Utils.spaceToDash(name).toLowerCase() + '-card');
    }

    async isCardSelected(name: string): Promise<boolean> {
        const element = this.page.locator(
            `[aid="${Utils.spaceToDash(name).toLowerCase()}-card"] > .card-content`
        );

        // Check for the attribute
        const attributeValue = await element.getAttribute('class');
        return attributeValue.includes('is-checked');
    }

    async clickOnNewDataSourceBtn(): Promise<void> {
        await this.locator(this.dataSourceNewBtn).click();
    }

    async selectAssetsSubmissionOption(): Promise<void> {
        await this.page.waitForSelector(this.submissionSection, { timeout: 3000 });
        await this.page.locator(this.submissionAssetsOption).click();
    }

    async openGeneralSection(): Promise<void> {
        await this.locator(this.generalSection).click();
    }

    async clickOnAddNewAnnotation(): Promise<void> {
        await this.page.locator(this.addNewAnnotation, { hasText: 'annotation' }).click();
    }

    async clickOnAddNewLabel(): Promise<void> {
        await this.page.locator(this.addNewLabel, { hasText: 'label' }).click();
    }

    async fillInitializationTimeoutSeconds(initializationTimeout: {
        minutes?: number;
        seconds?: number;
    }): Promise<void> {
        if (initializationTimeout.minutes) {
            await this.locator(
                '.duration-input-section.initialization-timeout-container [aid="minutes-input"]'
            ).fill(initializationTimeout.minutes.toString());
        }

        if (initializationTimeout.seconds) {
            await this.locator(
                '.duration-input-section.initialization-timeout-container [aid="seconds-input"]'
            ).fill(initializationTimeout.seconds.toString());
        }
    }

    async fillRequestTimeoutSeconds(requestTimeout: { minutes?: number; seconds?: number }): Promise<void> {
        if (requestTimeout.minutes) {
            await this.locator(
                '.duration-input-section.request-timeout-container [aid="minutes-input"]'
            ).fill(requestTimeout.minutes.toString());
        }

        if (requestTimeout.seconds) {
            await this.locator(
                '.duration-input-section.request-timeout-container [aid="seconds-input"]'
            ).fill(requestTimeout.seconds.toString());
        }
    }

    async fillAnnotationValues(annotations: Array<{ name: string; value: string }>): Promise<void> {
        for (const el of annotations.entries()) {
            await this.clickOnAddNewAnnotation();
        }

        await this.fillItemizedValues('.annotations-section', annotations);
    }

    async fillLabelsValues(labels: Array<{ name: string; value: string }>): Promise<void> {
        for (const el of labels.entries()) {
            await this.clickOnAddNewLabel();
        }

        await this.fillItemizedValues('.labels-section', labels);
    }

    async fillItemizedValues(
        locator: string,
        itemizedArr: Array<{ name: string; value: string }>
    ): Promise<void> {
        const nameElements = await this.locator(`${locator} [aid="itemized-name-input"]`).all();
        const valueElements = await this.locator(`${locator} [aid="itemized-value-input"]`).all();

        if (nameElements.length !== valueElements.length) {
            throw new Error(
                `Mismatch in name and values: found ${nameElements.length} name inputs, but found ${valueElements.length} values.`
            );
        }

        if (itemizedArr.length !== valueElements.length) {
            throw new Error(
                `Mismatch in name and values and itemizedArr: Expected ${itemizedArr.length} elements, but found ${valueElements.length} names/values.`
            );
        }

        for (const [index, element] of nameElements.entries()) {
            await element.fill(itemizedArr[index].name);
        }

        for (const [index, element] of valueElements.entries()) {
            await element.fill(itemizedArr[index].value);
        }
    }
}
