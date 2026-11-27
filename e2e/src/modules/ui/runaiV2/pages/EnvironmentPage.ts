/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { AssetKind, Scope } from '../../../api/apis/assetsApi';
import { BaseV2Page } from './BaseV2Page';
import { Locator, Page } from 'playwright';
import { environmentUtils, logger, Utils, waitUtils } from '../../../helpers';
import { AssetsScopeSection } from '../components/AssetsScopeSection';
import { MENU_SUBJECT_NAMES, MENU_SUB_ITEMS } from '../../utils/sidenavUtils';

export class EnvironmentPage extends BaseV2Page {
    assetsScopeSection: AssetsScopeSection;

    // environment table elements
    addNewEnvironmentBtn: string;

    copyEnvironmentBtn: string;

    deleteEnvironmentBtn: string;

    deleteEnvironmentModalDeleteBtn: string;

    // create environment elements
    createEnvironmentNameInput: string;

    createEnvironmentImage: string;

    createEnvironmentBtn: string;

    environmentRunTimeSection: string;

    environmentSecuritySection: string;

    environmentToolsSection: string;

    environmentAddToolButton: string;

    environmentToolTypeSelect: string;

    environmentCmdAndArgsBtn: string;

    addEnvironmentVariableBtn: string;

    environmentCmdInput: string;

    environmentArgsInput: string;

    environmentVariableNameInput: string;

    environmentVariableValueInput: string;

    environmentAutoGenSelect: string;

    environmentUidInput: string;

    environmentGidInput: string;

    environmentOverideToggle: string;

    environmentRadioCustomOption: string;

    environmentToolSelectOptionVsCode: string;

    environmentAutoGenCustomUrlOption: string;

    environmentSection: string;

    runaiGeneralSearch: string;

    containerPort: string;

    environmentStandardArchitecture: string;

    environmentDistributedArchitecture: string;

    distFrameworkSelect: string;

    elemChecked: string;

    editEnvironmentBtn: string;

    environmentNameCol: string;

    constructor(page: Page) {
        super(page);

        this.assetsScopeSection = new AssetsScopeSection(page);
        this.mainPageTab = '[aid="menu-icon-environments"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Environments}"]`;
        this.addNewEnvironmentBtn = '.runai-page-actions button[aid="primary-cta-btn"]';
        this.environmentSection = '.environment-section';
        this.runaiGeneralSearch = '.runai-general-search';

        // environment elements
        this.createEnvironmentNameInput = `[aid="${AssetKind.Environment}-name-input"]`;
        this.createEnvironmentImage = '[aid="environment-image-input"]';
        this.createEnvironmentBtn = '[aid="create-environment-btn"]';
        this.copyEnvironmentBtn = '[aid="copy-environment-btn"]';
        this.deleteEnvironmentBtn = '[aid="delete-environment-btn"]';
        this.deleteEnvironmentModalDeleteBtn = '[aid="delete-modal-delete-button"]';
        this.environmentRunTimeSection = '[aid="form-runtime-section"]';
        this.environmentSecuritySection = '[aid="form-security-section"]';
        this.environmentToolsSection = '[aid="form-tools-section"]';
        this.environmentAddToolButton = '[aid="add-tool-button"]';
        this.environmentToolTypeSelect = '[aid="tool-type-select"]';
        this.environmentCmdAndArgsBtn = '[aid="add-command-and-args-button"]';
        this.environmentCmdInput = '[aid="command-input"]';
        this.environmentArgsInput = '[aid="args-input"]';
        this.addEnvironmentVariableBtn = '[aid="add-environment-variable-button"]';
        this.environmentVariableNameInput = '[aid="environment-variable-name"]';
        this.environmentVariableValueInput = '[aid="environment-variable-value"]';
        this.environmentAutoGenSelect = '[aid="auto-gen-select"]';
        this.environmentUidInput = '[aid="uid-input"]';
        this.environmentGidInput = '[aid="gid-input"]';
        this.environmentOverideToggle = '[aid="override-in-workspace-toggle"]';
        this.environmentRadioCustomOption = 'div[role="radio"]:has-text("Custom")';
        this.environmentToolSelectOptionVsCode = 'div[role="option"]:has-text("VSCode")';
        this.environmentAutoGenCustomUrlOption = 'div[role="option"]:has-text("Custom URL")';
        this.containerPort = '[aid="container-port"]';
        this.environmentStandardArchitecture = '[aid="radio-option-Standard"]';
        this.environmentDistributedArchitecture = '[aid="radio-option-Distributed"]';
        this.distFrameworkSelect = '[aid="dist-framework-select"]';
        this.elemChecked = 'aria-checked';
        this.editEnvironmentBtn = '[aid="edit-environment-btn"]';
        this.environmentNameCol = '[aid="environment"]';
    }

    async waitForScopeTreeToLoad(clusterId: string): Promise<void> {
        await this.waitForResponse(`/v1/k8s/clusters/${clusterId}/projects`);
    }

    async validateNewEnvironmentButtonExist(): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const environmentBtnCount = await this.v2frame.locator(this.addNewEnvironmentBtn).count();
                return environmentBtnCount !== 0;
            },
            'is new environment button exists',
            true,
            10
        );
    }

    // environment table actions
    async clickOnAddNewEnvironmentBtn(): Promise<void> {
        await this.v2frame.locator(this.addNewEnvironmentBtn).click();
    }

    async clickOnCopyEnvironmentBtn(): Promise<void> {
        await this.v2frame.locator(this.copyEnvironmentBtn).click();
    }

    async clickOnDeleteEnvironmentBtn(): Promise<void> {
        return this.v2frame.locator(this.deleteEnvironmentBtn).click();
    }

    async clickOnDeleteEnvironmentInModalBtn(): Promise<void> {
        return this.v2frame.locator(this.deleteEnvironmentModalDeleteBtn).click();
    }

    async removeEnvironment(name: string): Promise<void> {
        await this.searchByName(name);
        await this.clickOnElementByText('td', name);
        await this.clickOnDeleteEnvironmentBtn();
        await this.clickOnDeleteEnvironmentInModalBtn();
        await environmentUtils.waitForEnvironmentToBeDeleted(name);
        await this.waitForAssetToBeDeletedFromUI(this, name);
    }

    async copyEnvironment(name: string): Promise<void> {
        await this.clickOnElementByText('td', name);
        await this.clickOnCopyEnvironmentBtn();
    }

    async openEditEnvironmentForm(name: string): Promise<void> {
        await this.clickOnElementByText('td', name);
        await this.clickOnEditEnvironmentBtn();
    }

    async clickOnEditEnvironmentBtn(): Promise<void> {
        await this.page.locator(this.editEnvironmentBtn).click();
    }

    async clearImageInput(): Promise<void> {
        await this.page.locator(this.createEnvironmentImage).clear();
    }

    async clearEnvNameInput(): Promise<void> {
        await this.page.locator(this.createEnvironmentNameInput).clear();
    }

    // create environment elements
    async fillEnvironmentName(name: string): Promise<void> {
        await this.v2frame.locator(this.createEnvironmentNameInput).fill(name);
    }

    async selectEnvironmentArchitecture(architecture: 'Standard' | 'Distributed'): Promise<void> {
        if (architecture === 'Standard') {
            await this.page.locator(this.environmentStandardArchitecture).click();
        } else if (architecture === 'Distributed') {
            await this.page.locator(this.environmentDistributedArchitecture).click();
        } else {
            throw new Error(`Invalid architecture: ${architecture}`);
        }
    }

    async selectEnvironmentDistributedFramework(
        framework: 'PyTorch' | 'TensorFlow' | 'XG Boost' | 'MPI' | 'JAX'
    ): Promise<void> {
        await this.page.locator(this.distFrameworkSelect).click();
        await this.page.locator(`div[role="option"]:has-text("${framework}")`).click({ force: true });
        await this.page.waitForTimeout(500);
    }

    async selectWorkloadTypes(workloadTypes: ('workspace' | 'training' | 'inference')[]): Promise<void> {
        for (const type of workloadTypes) {
            const checkbox = this.page.locator(`[role="checkbox"]:has([aid="radio-option-${type}"])`);
            const isChecked = (await checkbox.getAttribute(this.elemChecked)) === 'true';
            if (!isChecked) {
                await checkbox.click();
            }
        }
    }

    async fillEnvironmentImage(name: string): Promise<void> {
        await this.v2frame.locator(this.createEnvironmentImage).fill(name);
    }

    async clickOnCreateEnvironmentBtn(): Promise<void> {
        return this.v2frame.locator(this.createEnvironmentBtn).click();
    }

    async clickOnRunTimeSettingsSection(): Promise<void> {
        return this.v2frame.locator(this.environmentRunTimeSection).click();
    }

    async clickOnSecuritySection(): Promise<void> {
        return this.v2frame.locator(this.environmentSecuritySection).click();
    }

    async openToolsSectionAndAddTool(): Promise<void> {
        await this.v2frame.locator(this.environmentToolsSection).click();
        await this.v2frame.locator(this.environmentAddToolButton).click();
    }

    async selectVsCodeAToolType(): Promise<void> {
        await this.v2frame.locator(this.environmentToolTypeSelect).click();
        return await this.v2frame.locator(this.environmentToolSelectOptionVsCode).click();
    }

    async selectCustomUrlOption(): Promise<void> {
        await this.v2frame.locator(this.environmentAutoGenSelect).click();
        return await this.v2frame.locator(this.environmentAutoGenCustomUrlOption).click();
    }

    async selectEnvironment(name: string): Promise<void> {
        await this.v2frame.locator(`${this.environmentSection} ${this.runaiGeneralSearch}`).click();
        await this.v2frame.locator(`${this.environmentSection} ${this.runaiGeneralSearch} input`).fill(name);
        await this.clickOnElementByAidName(Utils.spaceToDash(`${name}`).toLowerCase() + '-card');
    }

    async openCommandAndArgumentSection(): Promise<void> {
        return await this.v2frame.locator(this.environmentCmdAndArgsBtn).click();
    }

    async setCommand(cmd: string): Promise<void> {
        return await this.v2frame.locator(this.environmentCmdInput).fill(cmd);
    }

    async setArguments(args: string): Promise<void> {
        return this.page.locator(this.environmentArgsInput).fill(args);
    }

    async setEnvironmentVariable(name: string, value: string): Promise<void> {
        await this.v2frame.locator(this.addEnvironmentVariableBtn).click();
        await this.v2frame.locator(this.environmentVariableNameInput).fill(name, { timeout: 3000 });
        return await this.v2frame.locator(this.environmentVariableValueInput).fill(value);
    }

    async clickCustomRadioOption(): Promise<void> {
        await this.v2frame.locator(this.environmentRadioCustomOption).click();
    }

    async setUid(uid: number): Promise<void> {
        await this.v2frame.locator(this.environmentUidInput).fill(uid.toString());
    }

    async setGid(gid: number): Promise<void> {
        await this.v2frame.locator(this.environmentGidInput).fill(gid.toString());
    }

    async setOverideToggle(value: boolean): Promise<void> {
        const checkedAttr = await this.v2frame
            .locator(this.environmentOverideToggle)
            .getAttribute(this.elemChecked);

        const isChecked = checkedAttr === 'true';

        if (isChecked === value) return;

        return await this.v2frame.locator(this.environmentOverideToggle).click();
    }

    async selectEnvironmentInTableByName(environmentName: string): Promise<void> {
        const environmentNameCell: Locator = this.findByText('td', environmentName);
        await environmentNameCell.click();
    }

    async setContainerPort(containerPort: number): Promise<void> {
        await this.v2frame.locator(this.containerPort).fill(containerPort.toString());
    }

    async createNewEnvironment(
        name: string,
        projectName?: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
            image?: string;
            architecture?: 'Standard' | 'Distributed';
            distributedType?: 'PyTorch' | 'TensorFlow' | 'XG Boost' | 'MPI' | 'JAX';
            workloadTypes?: ('workspace' | 'training' | 'inference')[];
            selectProjectScope?: boolean;
            waitForProjectScope?: boolean;
            waitForTenantScope?: boolean;
            command?: string;
            args?: string;
            containerPort?: number;
        }
    ): Promise<void> {
        logger.debug(`creating new environment: ${name}`);
        await this.selectCurrentCluster();
        if (options?.selectProjectScope) {
            await this.assetsScopeSection.selectScope({
                scope: options.scope,
                scopeEntityName: options.scopeEntityName
            });
        } else if (options?.waitForProjectScope) {
            await this.waitForScopeSectionContent(projectName);
        } else if (options?.waitForTenantScope) {
            await this.waitForTenantScopeSection();
        }

        await this.fillEnvironmentName(name);
        if (options?.architecture) {
            await this.selectEnvironmentArchitecture(options?.architecture);
        }
        if (options?.architecture === 'Distributed') {
            if (!options?.distributedType) {
                throw new Error('distributedType must be set when architecture is Distributed');
            }
            await this.selectEnvironmentDistributedFramework(options.distributedType);
        } else if (options?.distributedType) {
            throw new Error('Distributed type can only be selected for Distributed architecture');
        }
        if (options?.workloadTypes) {
            await this.selectWorkloadTypes(options.workloadTypes);
        }
        await this.fillEnvironmentImage(options?.image || 'alpine');

        await this.clickOnRunTimeSettingsSection();
        await this.openCommandAndArgumentSection();
        if (options?.command) {
            await this.setCommand(options?.command);
        }
        if (options?.args) {
            await this.setArguments(options?.args);
        }
        if (options?.containerPort) {
            await this.setContainerPort(options.containerPort);
        }
        await this.clickOnCreateEnvironmentBtn();
        logger.debug(`environment ${name} created successfully`);
    }

    async getEnvironmentNames(): Promise<string[]> {
        return await this.page.locator(this.environmentNameCol).allTextContents();
    }
}
