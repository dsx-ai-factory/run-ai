/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { AssetKind, Scope } from '../../../api/apis/assetsApi';
import { BaseV2Page } from './BaseV2Page';
import { Locator, Page } from 'playwright';
import { credentialUtils, logger, waitUtils } from '../../../helpers';
import { SecretType } from '../../../api/apis/k8sObjectsTrackerApi/generated/model/secretType';
import { AssetsScopeSection } from '../components/AssetsScopeSection';
import { MENU_SUBJECT_NAMES, MENU_SUB_ITEMS } from '../../utils/sidenavUtils';

const SSH_KEY_FORMAT = 'ssh-key';
type TGenericFormats = 'generic' | 'access-key' | 'password' | 'ssh-key';
export class CredentialPage extends BaseV2Page {
    assetsScopeSection: AssetsScopeSection;

    mainPageCredentialsTab: string;

    // Credential table elements
    addNewCredentialBtn: string;

    addNewAccessKeyBtn: string;

    addNewUserPassBtn: string;

    addNewDockerRegistryBtn: string;

    addNewGenericSecretBtn: string;

    addNewNgcApiKeyBtn: string;

    deleteCredentialBtn: string;

    modalDeleteBtn: string;

    newSecretRadioBtn: string;

    credentialSubmitBtn: string;

    sideBar: string;

    renameCredentialsBtn: string;

    renameModalNameInput: string;

    renameModalSaveBtn: string;

    genericPairIdentifierInput: string;

    genericPairSecretInput: string;

    // Docker registry creation elements
    createDockerRegistryNameInput: string;

    dockerRegistryUsernameInput: string;

    dockerRegistryPasswordInput: string;

    dockerRegistryUrlInput: string;

    // Generic secret creation elements

    createGenericSecretNameInput: string;

    genericSecretKeyInput: string;

    genericSecretValueInput: string;

    // NGC API Key creation elements

    createNgcApiKeyNameInput: string;

    ngcApiKeyInput: string;

    ngcApiKeyDescriptionInput: string;

    credentialNameInput: {
        password: string;
        'docker-registry': string;
        'access-key': string;
        generic: string;
    };

    credentialButtons: {
        password: string;
        'docker-registry': string;
        'access-key': string;
        generic: string;
    };

    genericCredentialFormatsLabels: Record<TGenericFormats, string>;

    existingSecretSelect: string;

    genericFormatSelect: string;

    genericFormatRadioOptions: string;

    genericSecretKeyValueSection: string;

    firstProjectCheckboxIndexDepartmentEnabled: number;

    firstProjectCheckboxIndexDepartmentDisabled: number;

    editModalCancelBtn: string;

    genericSecretCreationDescInput: string;

    credentialCreationDescriptionInput: string;

    renameModalDescriptionInput: string;

    updateModalDescriptionInput: string;

    // Table column selectors
    credentialNameCol: string;

    typeCol: string;

    statusCol: string;

    createdByCol: string;

    descriptionCol: string;

    constructor(page: Page) {
        super(page);

        this.assetsScopeSection = new AssetsScopeSection(page);

        this.sideBar = '.side-nav';
        this.mainPageTab = '[aid="menu-icon-settings"]';
        this.mainPageCredentialsTab = '[aid="menu-icon-credentials"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Credentials}"]`;

        this.addNewCredentialBtn = '.runai-page-actions button[aid="primary-cta-btn"]';
        this.addNewGenericSecretBtn = '[aid="generic-secret-new-btn"]';
        this.addNewDockerRegistryBtn = '[aid="docker-registry-new-btn"]';
        this.addNewNgcApiKeyBtn = '[aid="ngc-api-key-new-btn"]';
        this.deleteCredentialBtn = '[aid="delete-credential-btn"]';
        this.modalDeleteBtn = '[aid="delete-modal-delete-button"]';
        this.newSecretRadioBtn = 'div[role="radio"]:has-text("New secret")';
        this.credentialSubmitBtn = '[aid="submit-btn"]';

        this.renameCredentialsBtn = '[aid="rename-credentials-btn"]';
        this.renameModalNameInput = '[aid="runai-name-validation"]';
        this.renameModalSaveBtn = '[aid="rename-modal-save-button"]';

        this.genericPairIdentifierInput = '[aid="generic-pair-identifier-input"]';
        this.genericPairSecretInput = '[aid="generic-pair-secret-input"]';

        this.createDockerRegistryNameInput = `[aid="${AssetKind.DockerRegistry}-name-input"]`;
        this.dockerRegistryUsernameInput = '[aid="docker-registry-username-input"]';
        this.dockerRegistryPasswordInput = '[aid="docker-registry-password-input"]';
        this.dockerRegistryUrlInput = '[aid="docker-registry-url-input"]';
        this.existingSecretSelect = '[aid="existing-secret-select"]';

        this.createGenericSecretNameInput = `[aid="${AssetKind.GenericSecret}-name-input"]`;
        this.genericSecretKeyInput = '[aid="itemized-name-input"]';
        this.genericSecretValueInput = '[aid="itemized-value-input"]';
        this.editModalCancelBtn = '[aid="rename-modal-cancel-button"]';

        this.createNgcApiKeyNameInput = '[aid="ngcApiKey-name-input"]';
        this.ngcApiKeyInput = '[aid="ngc-api-key-input"]';
        this.ngcApiKeyDescriptionInput = '[aid="ngcApiKey-description-input"]';
        this.updateModalDescriptionInput = 'textarea[aria-label="Description"]';

        // Table column selectors
        this.credentialNameCol = '[aid="credentials"]';
        this.typeCol = '[aid="type"]';
        this.statusCol = '[aid="status"]';
        this.createdByCol = '[aid="created-by"]';
        this.descriptionCol = '[aid="description"]';

        this.credentialButtons = {
            'access-key': this.addNewGenericSecretBtn,
            password: this.addNewGenericSecretBtn,
            'docker-registry': this.addNewDockerRegistryBtn,
            generic: this.addNewGenericSecretBtn
        };

        this.credentialNameInput = {
            'access-key': this.createGenericSecretNameInput,
            password: this.createGenericSecretNameInput,
            'docker-registry': this.createDockerRegistryNameInput,
            generic: this.createGenericSecretNameInput
        };

        this.genericCredentialFormatsLabels = {
            generic: 'Key(s) & Value(s)',
            'access-key': 'Access key',
            password: 'Username & password',
            [SSH_KEY_FORMAT]: 'SSH key'
        };

        this.genericFormatSelect = '[aid="secret-format-select"]';
        this.genericFormatRadioOptions = '[aid="secret-format-radio-options"]';
        this.genericSecretKeyValueSection = '[aid="generic-secret-key-value-section"]';
        this.genericSecretCreationDescInput = '[aid="genericSecret-description-input"]';
        this.credentialCreationDescriptionInput = '[aid="credential-description-input"]';
        this.renameModalDescriptionInput = 'textarea.q-placeholder';
    }

    async goToPage(): Promise<void> {
        await this.page.locator(this.subjectItem).click();
        await this.page.waitForTimeout(300);
        await this.page.locator(this.subMenuItem).click();

        await this.page.locator(this.mainPageTitle).click(); // this is a workaround to close the sidenav bar
        await this.page.waitForTimeout(500); // wait some time for the page to load
    }

    async clickOnAddNewCredentialBtn(): Promise<void> {
        return await this.v2frame.locator(this.addNewCredentialBtn).click();
    }

    async clickOnAddNewGenericSecretBtn(): Promise<void> {
        return await this.v2frame.locator(this.addNewGenericSecretBtn).click();
    }

    async selectCredentialTypeAndClick(credentialType: SecretType): Promise<void> {
        const credentialBtn = this.credentialButtons[credentialType];
        return await this.v2frame.locator(credentialBtn).click();
    }

    async clickOnAddNewDockerRegistryBtn(): Promise<void> {
        return await this.v2frame.locator(this.addNewDockerRegistryBtn).click();
    }

    async clickOnCreateCredentialBtn(): Promise<void> {
        return await this.v2frame.locator(this.credentialSubmitBtn).click();
    }

    async clickOnDeleteCredentialBtn(): Promise<void> {
        return await this.v2frame.locator(this.deleteCredentialBtn).click();
    }

    async clickOnRenameCredentialBtn(): Promise<void> {
        return this.v2frame.locator(this.renameCredentialsBtn).click();
    }

    async fillRenameModalNameInput(newName: string): Promise<void> {
        return this.v2frame.locator(this.renameModalNameInput).fill(newName);
    }

    async fillRenameModalDescriptionInput(desc: string): Promise<void> {
        await this.page.locator(this.renameModalDescriptionInput).fill(desc);
    }

    async fillCredentialCreationDescriptionInput(desc: string): Promise<void> {
        await this.page.locator(this.credentialCreationDescriptionInput).fill(desc);
    }

    async fillUpdateModalDescriptionInput(desc: string): Promise<void> {
        await this.page.locator(this.updateModalDescriptionInput).fill(desc);
    }

    async fillGenericSecretDescInput(desc: string): Promise<void> {
        await this.page.locator(this.genericSecretCreationDescInput).fill(desc);
    }

    async getDescriptionInputText(): Promise<string> {
        return await this.page.locator('[placeholder="Description"]').inputValue();
    }

    async getEditModalDescriptionInputText(): Promise<string> {
        return await this.page.locator(this.renameModalDescriptionInput).inputValue();
    }

    async clickOnEditModalCancelBtn(): Promise<void> {
        await this.page.locator(this.editModalCancelBtn).click();
    }

    async clickOnRenameModalSaveBtn(): Promise<void> {
        return this.v2frame.locator(this.renameModalSaveBtn).click();
    }

    async clickOnDeleteCredentialConfirmBtn(): Promise<void> {
        return await this.v2frame.locator(this.modalDeleteBtn).click();
    }

    async selectCredentialInTableByName(credentialName: string): Promise<void> {
        await this.searchByName(credentialName);
        const credentialCell: Locator = this.findByText('td[aid="credentials"]', credentialName);
        await credentialCell.click();
    }

    async clickOnNewSecretRadioBtn(): Promise<void> {
        return await this.v2frame.locator(this.newSecretRadioBtn).click();
    }

    async fillCredentialName(name: string, credentialType: SecretType): Promise<void> {
        const credentialName = this.credentialNameInput[credentialType];
        await this.fillInputBySelector(credentialName, name);
    }

    async fillGenericPairIdentifierInput(identifier: string): Promise<void> {
        await this.fillInputBySelector(this.genericPairIdentifierInput, identifier);
    }

    async fillGenericPairSecretInput(secret: string): Promise<void> {
        await this.fillInputBySelector(this.genericPairSecretInput, secret);
    }

    async selectGenericFormat(format: TGenericFormats): Promise<void> {
        await this.selectOptionByValue(this.genericFormatSelect, this.genericCredentialFormatsLabels[format]);
    }

    // docker registry
    async fillDockerRegistryName(name: string): Promise<void> {
        await this.fillInputBySelector(this.createDockerRegistryNameInput, name);
    }

    async fillDockerRegistryUsernameInput(username: string): Promise<void> {
        await this.fillInputBySelector(this.dockerRegistryUsernameInput, username);
    }

    async fillDockerRegistryPasswordInput(password: string): Promise<void> {
        await this.fillInputBySelector(this.dockerRegistryPasswordInput, password);
    }

    async fillDockerRegistryUrlInput(url: string): Promise<void> {
        await this.fillInputBySelector(this.dockerRegistryUrlInput, url);
    }

    async fillGenericSecretNameInput(name: string): Promise<void> {
        await this.fillInputBySelector(this.createGenericSecretNameInput, name);
    }

    async fillGenericSecretKeyInput(key: string): Promise<void> {
        await this.fillInputBySelector(this.genericSecretKeyInput, key);
    }

    async fillGenericSecretValueInput(value: string): Promise<void> {
        await this.fillInputBySelector(this.genericSecretValueInput, value);
    }

    async createCredentialFromExistingSecret(
        credentialType: SecretType,
        credentialName: string,
        existingSecretName: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
        }
    ): Promise<void> {
        await this.clickOnAddNewCredentialBtn();

        await this.selectCredentialTypeAndClick(credentialType);
        await this.selectCurrentCluster();
        await this.assetsScopeSection.selectScope(options);

        await this.fillCredentialName(credentialName, credentialType);
        await this.selectOptionByValue(this.existingSecretSelect, existingSecretName);

        await this.clickOnCreateCredentialBtn();
    }

    async createAccessKeyCredential(
        name: string,
        id: string,
        secret: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
        }
    ): Promise<void> {
        await this.goToPage();
        await this.clickOnAddNewCredentialBtn();
        await this.clickOnAddNewGenericSecretBtn();
        await this.selectCurrentCluster();
        await this.assetsScopeSection.selectScope(options);
        await this.fillCredentialName(name, SecretType.AccessKey);
        await this.clickOnNewSecretRadioBtn();
        await this.selectGenericFormat('access-key');
        await this.fillGenericPairIdentifierInput(id);
        await this.fillGenericPairSecretInput(secret);
        await this.clickOnCreateCredentialBtn();
    }

    async createDockerRegistryCredential(
        name: string,
        username: string,
        password: string,
        url: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
        }
    ): Promise<void> {
        await this.goToPage();
        await this.clickOnAddNewCredentialBtn();
        await this.clickOnAddNewDockerRegistryBtn();
        await this.selectCurrentCluster();
        await this.assetsScopeSection.selectScope(options);
        await this.fillDockerRegistryName(name);
        await this.clickOnNewSecretRadioBtn();
        await this.fillDockerRegistryUsernameInput(username);
        await this.fillDockerRegistryPasswordInput(password);
        await this.fillDockerRegistryUrlInput(url);
        await this.clickOnCreateCredentialBtn();
    }

    async createUserPassCredential(
        name: string,
        username: string,
        password: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
        },
        description?: string
    ): Promise<void> {
        await this.goToPage();
        await this.clickOnAddNewCredentialBtn();
        await this.clickOnAddNewGenericSecretBtn();
        await this.selectCurrentCluster();
        await this.assetsScopeSection.selectScope(options);
        await this.fillCredentialName(name, SecretType.Password);
        await this.clickOnNewSecretRadioBtn();
        await this.selectGenericFormat('password');
        await this.fillGenericPairIdentifierInput(username);
        await this.fillGenericPairSecretInput(password);
        if (description) {
            await this.fillGenericSecretDescInput(description);
        }
        await this.clickOnCreateCredentialBtn();
    }

    async deleteCredential(name: string): Promise<void> {
        await this.selectCredentialInTableByName(name);
        await this.clickOnDeleteCredentialBtn();
        await this.clickOnDeleteCredentialConfirmBtn();
        await credentialUtils.waitForCredentialsToBeDeleted(name);
        logger.info(`Credential deleted successfully: ${name}`);
    }

    async isElementVisible(query: string): Promise<boolean> {
        return this.page.locator(query).isVisible();
    }

    async toggleCredential(credentialName: string): Promise<void> {
        const visible = await this.isElementVisible(
            `td${this.credentialNameCol}:has-text("${credentialName}")`
        );
        if (!visible) await this.searchByName(credentialName);
        await this.clickOnElementByText(`td${this.credentialNameCol}`, `${credentialName}`);
    }

    async renameCredential(oldName: string, newName: string): Promise<void> {
        // Select the credential
        await this.selectCredentialInTableByName(oldName);
        // Click update button
        await this.clickOnRenameCredentialBtn();
        // Fill the new name
        await this.fillRenameModalNameInput(newName);
        // Click save
        await this.clickOnRenameModalSaveBtn();
        // Wait for the name to change
        await this.selectCredentialInTableByName(newName);
    }

    async updateCredentialDesc(name: string, newDesc: string): Promise<void> {
        // Select the credential
        await this.selectCredentialInTableByName(name);
        // Click update button
        await this.clickOnRenameCredentialBtn();
        // Fill the new description
        await this.fillRenameModalDescriptionInput(newDesc);
        // Click save
        await this.clickOnRenameModalSaveBtn();
        // Wait for the name to change
        await this.selectCredentialInTableByName(name);
    }

    // NGC API Key methods
    async clickOnAddNewNgcApiKeyBtn(): Promise<void> {
        return await this.v2frame.locator(this.addNewNgcApiKeyBtn).click();
    }

    async fillNgcApiKeyNameInput(name: string): Promise<void> {
        await this.fillInputBySelector(this.createNgcApiKeyNameInput, name);
    }

    async fillNgcApiKeyInput(key: string): Promise<void> {
        await this.fillInputBySelector(this.ngcApiKeyInput, key);
    }

    async fillNgcApiKeyDescriptionInput(description: string): Promise<void> {
        await this.fillInputBySelector(this.ngcApiKeyDescriptionInput, description);
    }

    async createNgcApiKeyCredential(
        name: string,
        ngcApiKey: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
        },
        description?: string
    ): Promise<void> {
        await this.goToPage();
        await this.clickOnAddNewCredentialBtn();
        await this.clickOnAddNewNgcApiKeyBtn();
        await this.selectCurrentCluster();
        await this.assetsScopeSection.selectScope(options);
        await this.fillNgcApiKeyNameInput(name);
        await this.fillNgcApiKeyInput(ngcApiKey);
        if (description) {
            await this.fillNgcApiKeyDescriptionInput(description);
        }
        await this.clickOnCreateCredentialBtn();
    }

    // Table verification methods
    async getCredentialNames(): Promise<string[]> {
        return await this.page.locator(this.credentialNameCol).allTextContents();
    }

    async getCredentialRowIndexByName(credentialName: string): Promise<number> {
        const credentialNames: string[] = await this.getCredentialNames();
        const index = credentialNames.findIndex((name: string) => name.trim() === credentialName);
        if (index === -1) {
            throw new Error(`Credential with name ${credentialName} not found in table`);
        }
        return index;
    }

    async isColumnContentCorrect(
        columnSelector: string,
        rowIndex: number,
        expectedValue: string
    ): Promise<void> {
        await waitUtils.waitForCondition(async () => {
            const allCells = await this.page.locator(columnSelector).allTextContents();
            if (rowIndex < allCells.length) {
                const cellValue = allCells[rowIndex].trim();
                return cellValue === expectedValue;
            }
            return false;
        }, `Waiting for column ${columnSelector} at row ${rowIndex} to have value: ${expectedValue}`);
    }

    async isTypeCorrect(rowIndex: number, expectedType: string): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.typeCol);
        await this.isColumnContentCorrect(this.typeCol, rowIndex, expectedType);
    }

    async isStatusCorrect(rowIndex: number, expectedStatus: string): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.statusCol);
        await this.isColumnContentCorrect(this.statusCol, rowIndex, expectedStatus);
    }

    async isCreatedByCorrect(rowIndex: number, expectedCreatedBy: string): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.createdByCol);
        await this.isColumnContentCorrect(this.createdByCol, rowIndex, expectedCreatedBy);
    }

    async isDescriptionCorrect(rowIndex: number, expectedDescription: string): Promise<void> {
        await waitUtils.waitForElementToExist(this, this.descriptionCol);
        await this.isColumnContentCorrect(this.descriptionCol, rowIndex, expectedDescription);
    }

    async verifyCredentialDetails({
        credentialName,
        type,
        status,
        createdBy,
        description
    }: {
        credentialName: string;
        type: string;
        status: string;
        createdBy: string;
        description?: string;
    }): Promise<void> {
        // Check if the credential is visible in the current table view
        const visible = await this.isElementVisible(
            `td${this.credentialNameCol}:has-text("${credentialName}")`
        );

        // Only search if not visible (avoid unnecessary search operations)
        if (!visible) {
            await this.searchByName(credentialName);
        }

        // Don't click to select - we only need to verify the data
        const rowIndex = await this.getCredentialRowIndexByName(credentialName);

        await this.isTypeCorrect(rowIndex, type);
        await this.isStatusCorrect(rowIndex, status);
        await this.isCreatedByCorrect(rowIndex, createdBy);

        // Verify description if provided
        if (description !== undefined) {
            await this.isDescriptionCorrect(rowIndex, description);
        }

        logger.info(`All credential details verified successfully for: ${credentialName}`);
    }
}
