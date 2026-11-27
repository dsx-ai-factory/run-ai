/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Locator, Page } from 'playwright';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { BaseV2Page } from './BaseV2Page';
import { ScopeType } from '../../../../modules/api/apis/authorizationApi/generated/model/scopeType';

export class DataVolumePage extends BaseV2Page {
    // Table page
    addNewDataVolumeBtn: string;

    actionsDeleteBtn: string;

    actionsEditBtn: string;

    deleteModalConfirmBtn: string;

    nameColumn: string;

    // Data volume form
    dataVolumeNameInput: string;

    dataVolumeDescriptionInput: string;

    dataVolumeOriginPvcTreeBtn: string;

    dataVolumeScopeTreeSearchInput: string;

    tenantScopeCheckbox: string;

    clusterScopeCheckbox: string;

    departmentScopeCheckbox: string;

    projectScopeCheckboxDepartmentsEnabled: string;

    projectScopeCheckboxDepartmentsDisabled: string;

    applyScopeBtn: string;

    sharedScopesScopeBtn: string;

    dataVolumeOriginPvcSelect: string;

    submitDataVolumeBtn: string;

    dataVolumeNameCol: string;

    deleteBtn: string;

    modalDeleteBtn: string;

    constructor(page: Page) {
        super(page);
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.WorkloadsManager}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.DataVolumes}"]`;

        this.addNewDataVolumeBtn = '[aid="primary-cta-btn"]';
        this.actionsDeleteBtn = '[aid="delete-data-volume-btn"]';
        this.actionsEditBtn = '[aid="edit-data-volume-btn"]';
        this.deleteModalConfirmBtn = '[aid="delete-modal-delete-button"]';
        this.nameColumn = 'td[aid="name"]';

        this.dataVolumeNameInput = '[aid="data-volume-name-input"]';
        this.dataVolumeDescriptionInput = '[aid="data-volume-description-input"]';
        this.dataVolumeOriginPvcTreeBtn = '.data-origin-section .tree-button';
        this.dataVolumeScopeTreeSearchInput = '[aid="org-tree-search-input"]';
        this.tenantScopeCheckbox = 'div[role="checkbox"] >> nth=0';
        this.clusterScopeCheckbox = 'div[role="checkbox"] >> nth=1';
        this.departmentScopeCheckbox = 'div[role="checkbox"] >> nth=2';
        this.projectScopeCheckboxDepartmentsEnabled = 'div[role="checkbox"] >> nth=3';
        this.projectScopeCheckboxDepartmentsDisabled = 'div[role="checkbox"] >> nth=2';
        this.applyScopeBtn = '[aid="org-tree-modal-select-btn"]';
        this.sharedScopesScopeBtn = '.scope-btn';
        this.dataVolumeOriginPvcSelect = '.origin-pvc-select';
        this.submitDataVolumeBtn = '[aid="data-volume-submit-btn"]';
        this.dataVolumeNameCol = '[aid="name"]';
        this.deleteBtn = '.runai-action-button:has-text("delete")';
        this.modalDeleteBtn = '[aid="delete-modal-delete-button"]';
    }

    async goToPage(): Promise<void> {
        await this.page.locator(this.subjectItem).click();
        await this.page.waitForTimeout(300);
        await this.page.locator(this.subMenuItem).click();
        await this.page.locator(this.mainPageTitle).click(); // this is a workaround to close the sidenav bar
        await this.page.waitForTimeout(500); // wait some time for the page to load
    }

    async fillNameInput(name: string): Promise<void> {
        await this.page.fill(this.dataVolumeNameInput, name);
    }

    async fillDescriptionInput(description: string): Promise<void> {
        await this.page.fill(this.dataVolumeDescriptionInput, description);
    }

    async clickOnOriginPvcTreeBtn(): Promise<void> {
        await this.page.locator(this.dataVolumeOriginPvcTreeBtn).click();
    }

    async clickOnNewDataVolumeBtn(): Promise<void> {
        await this.page.locator(this.addNewDataVolumeBtn).click();
    }

    async searchScopeTree(searchText: string): Promise<void> {
        await this.page.fill(this.dataVolumeScopeTreeSearchInput, searchText);
    }

    async selectOriginProjectScope(projectName: string): Promise<void> {
        await this.clickOnOriginPvcTreeBtn();
        await this.searchScopeTree(projectName);
        const projectCheckbox = this.projectScopeCheckboxDepartmentsEnabled;

        await this.page.locator(projectCheckbox).click();
        await this.page.locator(this.applyScopeBtn).click();
    }

    async selectOriginPvc(pvcName: string): Promise<void> {
        await this.selectOptionByValue(this.dataVolumeOriginPvcSelect, pvcName);
    }

    async selectSharedScope(scopeType: ScopeType, scopeName: string): Promise<void> {
        await this.locator(this.sharedScopesScopeBtn).click();
        await this.fillInputBySelector(this.dataVolumeScopeTreeSearchInput, scopeName);

        if (scopeType === ScopeType.Cluster) {
            await this.page.locator(this.clusterScopeCheckbox).click();
        }

        if (scopeType === ScopeType.Department) {
            await this.page.locator(this.departmentScopeCheckbox).click();
        }

        if (scopeType === ScopeType.Project) {
            const projectCheckbox = this.projectScopeCheckboxDepartmentsEnabled;

            await this.page.locator(projectCheckbox).click();
        }

        await this.page.locator(this.applyScopeBtn).click();
    }

    async clickOnSubmitDataVolumeBtn(): Promise<void> {
        await this.page.locator(this.submitDataVolumeBtn).click();
    }

    async selectDataVolumeInTableByName(dvName: string): Promise<void> {
        await this.searchByName(dvName);
        const dvCell: Locator = this.findByText(this.nameColumn, dvName);
        await dvCell.click();
    }

    async clickOnDeleteDataVolumeBtn(): Promise<void> {
        return await this.locator(this.actionsDeleteBtn).click();
    }

    async clickOnDeleteDataVolumeConfirmBtn(): Promise<void> {
        return await this.locator(this.deleteModalConfirmBtn).click();
    }

    async clickOnEditDataVolumeBtn(): Promise<void> {
        await this.locator(this.actionsEditBtn).click();
    }

    async createDataVolume(
        dataVolumeName: string,
        dataOrigin: { projectName: string; pvcName: string },
        sharedScope: { scopeType: ScopeType; scopeNameList: Array<string> }
    ): Promise<void> {
        await this.clickOnNewDataVolumeBtn();
        await this.fillNameInput(dataVolumeName);
        await this.selectOriginProjectScope(dataOrigin.projectName);
        await this.selectOriginPvc(dataOrigin.pvcName);
        for (const scopeName of sharedScope.scopeNameList) {
            await this.selectSharedScope(sharedScope.scopeType, scopeName);
        }
        await this.clickOnSubmitDataVolumeBtn();
    }

    async toggleDataVolume(dataVolumeName: string): Promise<void> {
        await this.clickOnElementByText(`td${this.dataVolumeNameCol}`, `${dataVolumeName}`);
    }

    async deleteDataVolume(dataVolumeName: string): Promise<void> {
        await this.toggleDataVolume(dataVolumeName);
        await this.page.locator(this.deleteBtn).click();
        await this.page.locator(this.modalDeleteBtn).click();
    }
}
