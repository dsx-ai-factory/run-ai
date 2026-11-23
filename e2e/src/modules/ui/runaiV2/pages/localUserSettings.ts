/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { logger } from '../../../helpers';

export class LocalUserSettings extends BaseV2Page {
    profileIcon: string;

    localSettings: string;

    userDetailsTitle: string;

    createPersonalApp: string;

    PersonalAppNameLabel: string;

    createButton: string;

    personalAppNAme: string;

    doneButton: string;

    deleteButton: string;

    deleteConfirmButton: string;

    regenerateSecret: string;

    regenerateConfirmationButton: string;

    doneRegenerateButton: string;

    newClientSecretValidation: string;

    applicationSubHeader: string;

    credentialsSection: string;

    addNewCredentialBtn: string;

    dockerRegistryOption: string;

    genericSecretOption: string;

    ngcApiKeyOption: string;

    ngcApiKeyNameInput: string;

    ngcApiKeyInput: string;

    ngcApiKeyDescriptionInput: string;

    credentialCard: string;

    credentialType: string;

    credentialName: string;

    credentialDescription: string;

    credentialNameDisplayInput: string;

    credentialTypeDisplayInput: string;

    credentialDescriptionDisplayTextarea: string;

    submitCredentialFormBtn: string;

    deleteCredentialBtn: string;

    deleteModalPersonalCredentialConfirmBtn: string;

    constructor(page: Page) {
        super(page);
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Admin}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.GeneralSettings}"]`;
        this.profileIcon = '[aid="profile-btn"]';
        this.localSettings = 'text="Settings"';
        this.userDetailsTitle = '[aid="user details-expansion-label"]';
        this.createPersonalApp = '[aid="aidAddApplication"]';
        this.PersonalAppNameLabel = '[aid="runai-name-validation"]';
        this.createButton = '[aid="next-btn"]';
        this.personalAppNAme = '[aid="user-app-name"]';
        this.doneButton = '[aid="done-btn"]';
        this.deleteButton = '[aid="remove-user-app-button"]';
        this.deleteConfirmButton = '[aid="delete-modal-delete-button"]';
        this.regenerateSecret = '.q-icon.ph-fill.ph-lock';
        this.regenerateConfirmationButton = '(//span[contains(text(),"regenerate")])[1]';
        this.doneRegenerateButton = '//span[contains(text(),"Done")]';
        this.newClientSecretValidation = '//span[normalize-space()="New client secret:"]';
        this.applicationSubHeader = 'text="Access keys"';
        this.credentialsSection = '[aid="credentials-expansion-toggle"]';
        this.addNewCredentialBtn = '[aid="add-new-credential"]';
        this.dockerRegistryOption = '[aid="docker-registry-new-btn"]';
        this.genericSecretOption = '[aid="generic-secret-new-btn"]';
        this.ngcApiKeyOption = '[aid="ngc-api-key-new-btn"]';
        this.ngcApiKeyNameInput = '[aid="ngcApiKey-name-input"]';
        this.ngcApiKeyInput = '[aid="ngc-api-key-input"]';
        this.ngcApiKeyDescriptionInput = '[aid="credential-description-input"]';
        this.credentialCard = '[aid="credential-card"]';
        this.credentialType = '[aid="credential-type"]';
        this.credentialName = '[aid="credential-name"]';
        this.credentialDescription = '[aid="credential-description"]';
        this.credentialNameDisplayInput = 'input[type="text"][disabled][aria-label="Credential name"]';
        this.credentialTypeDisplayInput = 'input[type="text"][disabled][aria-label="Credential type"]';
        this.credentialDescriptionDisplayTextarea = 'textarea[disabled][aria-label="Description"]';
        this.submitCredentialFormBtn = '[aid="credential-modal-create-button"]';
        this.deleteCredentialBtn = '[aid="delete-credential-${name}"]';
        this.deleteModalPersonalCredentialConfirmBtn = '[aid="delete-modal-delete-button"]';
    }

    async openLocalSettingsPage(): Promise<void> {
        await this.page.locator(this.profileIcon).click({ noWaitAfter: true });
        await this.page.locator(this.localSettings).click();
        await this.page.waitForSelector(this.userDetailsTitle, { timeout: 10000 });
    }

    async createAccessKeyUi(accessKeyName: string): Promise<string> {
        await this.page.locator(this.applicationSubHeader).click();
        await this.page.locator(this.createPersonalApp).click();
        await this.page.waitForSelector(this.PersonalAppNameLabel, { timeout: 10000 });
        await this.page.locator(this.PersonalAppNameLabel).fill(accessKeyName);
        await this.page.locator(this.createButton).click({ timeout: 10000 });
        await this.page.waitForTimeout(1000);
        await this.page.locator(this.createButton).click({ timeout: 10000 });
        await this.page.waitForTimeout(1000);
        const accessKeyNameElements = this.page.locator(this.personalAppNAme);
        const count = await accessKeyNameElements.count();

        for (let i = 0; i < count; i++) {
            const currentAccessKeyName = await accessKeyNameElements.nth(i).inputValue();
            if (currentAccessKeyName?.trim() === accessKeyName) {
                return currentAccessKeyName;
            }
        }
        throw new Error(`Access key name "${accessKeyName}" not found`);
    }

    async deleteAccessKeyUi(): Promise<void> {
        await this.page.locator(this.deleteButton).click();
        await this.page.waitForTimeout(1500);
        await this.page.locator(this.deleteConfirmButton).click();
        try {
            await this.page.waitForResponse((response) => response.status() === 204, { timeout: 10000 });
        } catch (error) {
            logger.error('Error occurred while waiting for the delete response:', error);
        }
    }

    async validateAccessKeyUiNoExist(accessKeyName: string): Promise<boolean> {
        const accessKeyNameElements = this.page.locator(this.personalAppNAme);
        let count = await accessKeyNameElements.count();
        let attempts = 0;

        while (count > 0 && attempts < 5) {
            await this.page.waitForTimeout(1000);
            count = await accessKeyNameElements.count();
            attempts++;
        }
        for (let i = 0; i < count; i++) {
            const currentAccessKeyName = await accessKeyNameElements.nth(i).inputValue();
            if (currentAccessKeyName?.trim() === accessKeyName) {
                return true;
            }
        }
        return false;
    }

    async regenerateClientSecret(): Promise<boolean> {
        try {
            await this.page.locator(this.regenerateSecret).click();
            await this.page.waitForTimeout(1500);
            await this.page.locator(this.regenerateConfirmationButton).click();
            await this.page.waitForTimeout(1500);
            const isValid = await this.validateNewClientSecret();
            if (isValid) {
                await this.page.locator(this.doneRegenerateButton).click();
                logger.info('Client secret regenerated successfully.');
                return true;
            } else {
                logger.error('Failed to regenerate client secret. Validation failed.');
                return false;
            }
        } catch (error) {
            logger.error('Error occurred while regenerating client secret:', error);
            return false;
        }
    }

    async validateNewClientSecret(): Promise<boolean> {
        try {
            await this.page.waitForSelector(this.newClientSecretValidation, {
                timeout: 10000
            });
            logger.info('New client secret validated successfully.');
            return true;
        } catch (error) {
            logger.error('Failed to validate new client secret:', error);
            return false;
        }
    }

    async toggleCredentialsSection(): Promise<void> {
        await this.page.locator(this.credentialsSection).click();
    }

    async clickOnAddNewCredentialBtn(): Promise<void> {
        await this.page.locator(this.addNewCredentialBtn).click();
    }

    async openNewPersonalCredentialModal(credentialType: string): Promise<void> {
        await this.clickOnAddNewCredentialBtn();
        if (credentialType === 'docker-registry') {
            await this.page.locator(this.dockerRegistryOption).click();
        } else if (credentialType === 'generic-secret') {
            await this.page.locator(this.genericSecretOption).click();
        } else if (credentialType === 'ngc-api-key') {
            await this.page.locator(this.ngcApiKeyOption).click();
        }
    }

    async deletePersonalCredential(name: string): Promise<void> {
        await this.toggleCredentialsSection();
        await this.page.locator(this.deleteCredentialBtn.replace('${name}', name)).click();
        await this.page.locator(this.deleteModalPersonalCredentialConfirmBtn).click();
    }

    async clickOnSubmitCredentialFormBtn(): Promise<void> {
        await this.page.locator(this.submitCredentialFormBtn).click();
    }

    async createPersonalNgcApiKeyCredential(
        name: string,
        ngcApiKey: string,
        description?: string
    ): Promise<void> {
        await this.toggleCredentialsSection();
        await this.openNewPersonalCredentialModal('ngc-api-key');
        await this.page.locator(this.ngcApiKeyNameInput).fill(name);
        await this.page.locator(this.ngcApiKeyInput).fill(ngcApiKey);
        if (description) {
            await this.page.locator(this.ngcApiKeyDescriptionInput).fill(description);
        }
        await this.clickOnSubmitCredentialFormBtn();
    }

    async verifyPersonalNgcApiKeyCredential(
        name: string,
        description?: string
    ): Promise<{ nameMatch: boolean; typeMatch: boolean; descriptionMatch: boolean }> {
        await this.page.waitForTimeout(2000);

        const credentialNameInput = this.page.locator(this.credentialNameDisplayInput);
        const nameValue = await credentialNameInput.inputValue();
        const nameMatch = nameValue === name;

        const credentialTypeInput = this.page.locator(this.credentialTypeDisplayInput);
        const typeValue = await credentialTypeInput.inputValue();
        const typeMatch = typeValue === 'NGC API key';

        let descriptionMatch = true;
        if (description) {
            const descriptionTextarea = this.page.locator(this.credentialDescriptionDisplayTextarea);
            const descValue = await descriptionTextarea.inputValue();
            descriptionMatch = descValue === description;
        }

        return {
            nameMatch: nameMatch,
            typeMatch: typeMatch,
            descriptionMatch: descriptionMatch
        };
    }
}
