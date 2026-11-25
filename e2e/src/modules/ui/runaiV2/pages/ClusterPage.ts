/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Locator, Page } from 'playwright';
import { BaseV2Page } from './BaseV2Page';
import { Utils, clusterUtils, logger, settings } from '../../../helpers';
import { expect } from 'chai';
import { MENU_SUBJECT_NAMES, MENU_SUB_ITEMS } from '../../utils/sidenavUtils';
import { ClusterNetworkTopologiesModal } from './ClusterNetworkTopologiesModal';

export class ClusterPage extends BaseV2Page {
    addNewClusterBtn: string;

    deleteClusterBtn: string;

    deleteModalDeleteBtn: string;

    clusterNameInput: string;

    clusterVersionSelect: string;

    clusterUrlInput: string;

    stepperContinueBtn: string;

    doneBtn: string;

    table: Locator;

    allVersions: string;

    networkTopoBtn: string;

    topologyViewBtn: string;

    tableRow: string;

    clusterName: string;

    networkTopoModal: ClusterNetworkTopologiesModal;

    constructor(page: Page) {
        super(page);
        this.mainPageTab = '[aid="menu-icon-clusters"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Resources}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.Clusters}"]`;

        this.addNewClusterBtn = '.runai-page-actions button[aid="primary-cta-btn"]';
        this.deleteClusterBtn = 'button:has-text("remove")';
        this.deleteModalDeleteBtn = '[aid="remove-modal-remove-button"]';
        this.clusterNameInput = '[aid="runai-name-validation"]';
        this.clusterVersionSelect = '.cluster-version-select';
        this.stepperContinueBtn = 'button:has-text("Continue")';
        this.clusterUrlInput = '[aid=cluster-url-input]';
        this.doneBtn = 'button:has-text("Done")';
        this.table = this.v2frame.locator('.q-table');
        this.allVersions = '[role="listbox"]';
        this.networkTopoBtn = '[aid="network-topology-btn"]';
        this.topologyViewBtn = '[aid="topologies"] > button';
        this.tableRow = '.runai-table-row';
        this.clusterName = '[aid="name"]';

        this.networkTopoModal = new ClusterNetworkTopologiesModal(page);
    }

    async fillClusterName(name: string): Promise<void> {
        return this.v2frame.locator(this.clusterNameInput).fill(name);
    }

    async fillClusterUrl(url: string): Promise<void> {
        return this.v2frame.locator(this.clusterUrlInput).fill(url);
    }

    async clickOnAddNewClusterBtn(): Promise<void> {
        return this.v2frame.locator(this.addNewClusterBtn).click();
    }

    async clickOnDoneBtn(): Promise<void> {
        return this.v2frame.locator(this.doneBtn).click();
    }

    async clickOnDeleteClusterBtn(): Promise<void> {
        return this.v2frame.locator(this.deleteClusterBtn).click();
    }

    async clickOnConfirmDeleteBtn(): Promise<void> {
        return this.v2frame.locator(this.deleteModalDeleteBtn).click();
    }

    async clickOnContinueButton(): Promise<void> {
        return this.v2frame.locator(this.stepperContinueBtn).click();
    }

    async selectClusterFirstVersion(): Promise<void> {
        this.v2frame.locator(this.clusterVersionSelect).click();
        const firstOption = this.v2frame.locator('div[role="option"]').first();
        return firstOption.click();
    }

    async deleteCluster(): Promise<void> {
        await this.clickOnDeleteClusterBtn();
        await this.clickOnConfirmDeleteBtn();
    }

    async selectClusterInTableByName(clusterName: string): Promise<void> {
        const clusterNameCell: Locator = this.findByText('td', clusterName);
        await clusterNameCell.click();
    }

    async selectDistribution(name: string): Promise<void> {
        await this.clickOnElementByAidName(Utils.spaceToDash(`${name}`).toLowerCase() + '-card');
    }

    async getClusterTableData(): Promise<string[][]> {
        await this.table.locator('tbody tr:nth-of-type(2)').waitFor({ timeout: 10000 });
        const rows = await this.table.locator('tbody tr').elementHandles();
        const allRowsData: string[][] = [];
        for (const row of rows) {
            const cells = await row.$$('td');
            const cellTexts: string[] = [];
            for (const cell of cells) {
                const text = await cell.innerText();
                cellTexts.push(text);
            }
            allRowsData.push(cellTexts);
        }
        logger.info('Clusters JSON: ', allRowsData);
        return allRowsData;
    }

    async createClusterByUI(clusterName: string): Promise<void> {
        await this.goToPage();
        await this.clickOnAddNewClusterBtn();
        await this.fillClusterName(clusterName);
        await this.selectClusterFirstVersion();
        if (!settings.isSelfHosted) {
            await this.fillClusterUrl('https://test.com');
        }
        await this.clickOnContinueButton();
        await this.clickOnDoneBtn();
    }

    async getClusterVersions(): Promise<string[]> {
        const clusterVersions = [];
        await this.goToPage();
        await this.clickOnAddNewClusterBtn();
        await this.page.click(this.clusterVersionSelect);
        const versionElements = this.page.locator(this.allVersions);
        await this.page.waitForTimeout(1500);
        const versionHandles = await versionElements.elementHandles();
        for (const versionHandle of versionHandles) {
            const text = await versionHandle.innerText();
            const cleanedText = text
                .trim()
                .split('\n')
                .map((v) => v.trim())
                .filter(Boolean);
            clusterVersions.push(...cleanedText);
        }
        const uniqueVersions = Array.from(new Set(clusterVersions));
        return uniqueVersions;
    }

    async deleteClusterByUI(clusterName: string): Promise<void> {
        const clusterNameCell = this.findByText('td', clusterName);
        await clusterNameCell.click();
        const firstRowStatusColumnText = await this.v2frame
            .locator('tr:nth-child(1) > td[aid="status"]')
            .innerText();
        expect('waiting to connect').to.equal(firstRowStatusColumnText.trim().toLowerCase());
        await this.clickOnDeleteClusterBtn();
        await this.clickOnConfirmDeleteBtn();
        await clusterUtils.waitForClusterToBeDeleted(clusterName);
        logger.info(`Cluster - ${clusterName} was deleted successfully`);
    }

    async clickNeworkTopologyBtn(): Promise<void> {
        await this.page.locator(this.networkTopoBtn).click();
    }

    async openNetworkTopoCreationModal(clusterName: string): Promise<void> {
        await this.page.getByRole('cell', { name: clusterName, exact: true }).click();
        await this.clickNeworkTopologyBtn();
    }

    async openClusterNetworkTopologiesModal(clusterName: string): Promise<void> {
        const rows = this.locator(this.tableRow);
        const rowCount = await rows.count();

        for (let i = 0; i < rowCount; i++) {
            const row = rows.nth(i);
            const name = await row.locator(this.clusterName).textContent();
            if (name === clusterName) {
                await row.locator(this.topologyViewBtn).click();
                await this.networkTopoModal.page.waitForLoadState('load');
                break;
            }
        }
    }
}
