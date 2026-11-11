/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { AssetKind, Scope } from '../../../api/apis/assetsApi';
import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { computeResourceUtils, logger, waitUtils } from '../../../helpers';
import { tenantName } from '../../../helpers/config/const';
import { AssetsScopeSection } from '../components/AssetsScopeSection';
import { MENU_SUBJECT_NAMES, MENU_SUB_ITEMS } from '../../utils/sidenavUtils';
export class ComputeResourcePage extends BaseV2Page {
    assetsScopeSection: AssetsScopeSection;

    // compute table elements
    addNewComputeBtn: string;

    // create compute elements
    createComputeNameInput: string;

    createGpuRequestInput: string;

    createGpuDevicesInput: string;

    fractionToggle: string;

    createComputeBtn: string;

    deleteComputeBtn: string;

    deleteComputeModalDeleteBtn: string;

    computeResourceSection: string;

    runaiGeneralSearch: string;

    editBtn: string;

    nameAlreadyExistsError: string;

    makeACopyBtn: string;

    constructor(page: Page) {
        super(page);

        this.assetsScopeSection = new AssetsScopeSection(page);

        this.mainPageTab = '[aid="menu-icon-compute-resource"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.ComputeResource}"]`;

        this.addNewComputeBtn = '.runai-page-actions button[aid="primary-cta-btn"]';
        this.computeResourceSection = '.compute-resource-section';
        this.runaiGeneralSearch = '.runai-general-search';

        // compute elements
        this.createComputeNameInput = `[aid="${AssetKind.Compute}-name-input"]`;
        this.createGpuRequestInput = '[aid="gpu-request-input"]';
        this.fractionToggle = '[aid="gpu-fraction-toggle"]';
        this.createGpuDevicesInput = '[aid="gpu-devices-input"]';
        this.createComputeBtn = '[aid="form-create-compute-btn"]';
        this.deleteComputeBtn = '[aid="delete-compute-btn"]';
        this.deleteComputeModalDeleteBtn = '[aid="delete-modal-delete-button"]';
        this.editBtn = '[aid="edit-compute-btn"]';
        this.nameAlreadyExistsError = 'div[role="alert"]:has-text("This name already exists")';
        this.makeACopyBtn = '[aid="copy-compute-btn"]';
    }

    async validateNewComputeResourceExist(): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const computeResourceBtnCount = await this.v2frame.locator(this.addNewComputeBtn).count();
                return computeResourceBtnCount !== 0;
            },
            'is new compute resource button exists',
            true,
            10
        );
    }

    async addComputeResource(): Promise<void> {
        return await this.v2frame.locator(this.addNewComputeBtn).click();
    }

    async createNewComputeResource(
        name: string,
        options?: {
            scope?: Scope;
            scopeEntityName?: string;
        }
    ): Promise<void> {
        await this.selectCurrentCluster();
        await this.assetsScopeSection.selectScope(options);
        await this.fillComputeName(name);
        await this.fillGpuRequestInput();
        await this.clickOnCreateComputeBtn();
    }

    // environment table actions
    async clickOnDeleteComputeBtn(): Promise<void> {
        return await this.v2frame.locator(this.deleteComputeBtn).click();
    }

    async clickOnDeleteComputeInModalBtn(): Promise<void> {
        return await this.v2frame.locator(this.deleteComputeModalDeleteBtn).click();
    }

    async removeCompute(name: string): Promise<void> {
        logger.debug(`deleting compute: ${name}`);
        await this.searchByName(name);
        await this.clickOnElementByText('td', name);
        await this.clickOnDeleteComputeBtn();
        await this.clickOnDeleteComputeInModalBtn();
        await computeResourceUtils.waitForComputeResourceToBeDeleted(name);
        await this.waitForAssetToBeDeletedFromUI(this, name);
    }

    // create compute actions
    async fillComputeName(name: string): Promise<void> {
        await this.v2frame.locator(this.createComputeNameInput).fill(name);
    }

    async fillGpuRequestInput(gpuRequest = 99): Promise<void> {
        await this.v2frame.locator(this.createGpuDevicesInput).fill('1');
        if (gpuRequest < 100) {
            await this.switchFractionToggle(true);
            await this.locator(this.createGpuRequestInput).fill(`${gpuRequest}`);
        }
    }

    async switchFractionToggle(value: boolean): Promise<void> {
        return this.switchToggleBySelector(this.fractionToggle, value);
    }

    async clickOnCreateComputeBtn(): Promise<void> {
        return await this.v2frame.locator(this.createComputeBtn).click();
    }

    async createNewCompute(
        name: string,
        expectedScopeSectionContent: string,
        gpuRequest = 100,
        waitForScopeSectionContent: boolean = true
    ): Promise<void> {
        logger.debug(`creating new compute: ${name}`);
        if (waitForScopeSectionContent) {
            if (expectedScopeSectionContent === tenantName) {
                await this.waitForTenantScopeSection();
            } else {
                await this.waitForScopeSectionContent(expectedScopeSectionContent);
            }
        }
        await this.fillComputeName(name);
        if (gpuRequest !== 0) {
            await this.fillGpuRequestInput(gpuRequest);
        }

        await this.clickOnCreateComputeBtn();
        logger.debug(`compute ${name} created successfully`);
    }

    async selectComputeResource(name: string): Promise<void> {
        await this.v2frame.locator(`${this.computeResourceSection} ${this.runaiGeneralSearch}`).click();
        await this.v2frame
            .locator(`${this.computeResourceSection} ${this.runaiGeneralSearch} input`)
            .fill(name);
        return this.v2frame.locator(`[aid="${name}-card"]`).click();
    }

    async clickOnEditBtn(): Promise<void> {
        await this.page.locator(this.editBtn).click();
    }

    async openEditForm(name: string): Promise<void> {
        await this.clickOnElementByText('td', name);
        await this.clickOnEditBtn();
    }

    async clickOnMakeACopyBtn(): Promise<void> {
        await this.page.locator(this.makeACopyBtn).click();
    }

    async openCopyAndEditForm(name: string): Promise<void> {
        await this.clickOnElementByText('td', name);
        await this.clickOnMakeACopyBtn();
        await waitUtils.waitForElementToExist(this, this.nameAlreadyExistsError);
    }

    async clearComputeResourceNameInput(): Promise<void> {
        await this.page.locator(this.createComputeNameInput).clear();
    }
}
