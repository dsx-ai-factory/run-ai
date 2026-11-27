/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { NodePoolCreateRequest, PlacementStrategy } from '../../../api/apis/backendClusterApi';
import { logger, nodePoolUtils } from '../../../helpers';
import { MENU_SUB_ITEMS, MENU_SUBJECT_NAMES } from '../../utils/sidenavUtils';
import { waitUtils } from '../../../helpers';
import { QTable } from '../components/QTable';

const WAIT_FOR_NODE_POOL_REMOVED_MSG = 'wait for node pool to remove from table';

export class NodePoolPage extends BaseV2Page {
    //tabs
    nodePoolsTab: string;

    //table elements
    addNewNodePoolBtn: string;

    nodePoolStatusCol: string;

    nodePoolNameCol: string;

    nodePoolNodesCol: string;

    deleteNodePoolModalBtn: string;

    //create / edit form elements
    nodePoolNameInput: string;

    nodePoolLabelKeyInput: string;

    nodePoolLabelValueInput: string;

    saveNodePoolBtn: string;

    gpuPlacementRadioBinpack: string;

    gpuPlacementRadioSpread: string;

    cpuPlacementRadioBinpack: string;

    cpuPlacementRadioSpread: string;

    nodePoolLabelKey: string;

    schedulingExpansionLabel: string;

    gpuNetowrkAccelerationExpansionLabel: string;

    networkTopologyComoboxName: string;

    mnnvlIsPresentDiv: string;

    nodePoolTopoCol: string;

    newNetworkTopologyButton: string;

    topoNameInput: string;

    topoLabelKeyInput: string;

    topoLabelBtn: string;

    saveTopoBtn: string;

    nodePoolCancelBtn: string;

    topoModalSearchInput: string;

    topoModalLabels: string;

    networkTopologySection: string;

    tableRow: string;

    qTable: QTable;

    constructor(page: Page) {
        super(page);

        this.nodePoolsTab = 'a[href="/nodes/node-pool"]';
        this.mainPageTab = '[aid="menu-icon-nodes"]';
        this.subjectItem = `[aid="${MENU_SUBJECT_NAMES.Resources}"]`;
        this.subMenuItem = `[aid="${MENU_SUB_ITEMS.NodePools}"]`;

        this.addNewNodePoolBtn = '[aid="primary-cta-btn"]';
        this.nodePoolNameInput = '[aid="node-pool-name-input"]';
        this.nodePoolLabelKeyInput = '[aid="node-pool-label-key-input"]';
        this.nodePoolLabelValueInput = '[aid="node-pool-label-value-input"]';
        this.saveNodePoolBtn = '[aid="save-form-footer-btn"]';
        this.nodePoolNameCol = '[aid="name"]';
        this.nodePoolNodesCol = '[aid="nodes"]';
        this.nodePoolStatusCol = '[aid="status"]';
        this.nodePoolTopoCol = '[aid="topology"]';
        this.deleteNodePoolModalBtn = '[aid="delete-modal-delete-button"]';
        this.gpuPlacementRadioBinpack = '[aid="node-pool-gpu-placement-radio"] [aid="radio-option-Bin-pack"]';
        this.gpuPlacementRadioSpread = '[aid="node-pool-gpu-placement-radio"] [aid="radio-option-Spread"]';
        this.cpuPlacementRadioBinpack = '[aid="node-pool-cpu-placement-radio"] [aid="radio-option-Bin-pack"]';
        this.cpuPlacementRadioSpread = '[aid="node-pool-cpu-placement-radio"] [aid="radio-option-Spread"]';
        this.nodePoolLabelKey = 'nvidia.com/gpu.product';
        this.schedulingExpansionLabel = '[aid="scheduling-expansion-label"]';
        this.gpuNetowrkAccelerationExpansionLabel = '[aid="gpu network acceleration-expansion-toggle"]';
        this.networkTopologyComoboxName = 'Network topology';
        this.mnnvlIsPresentDiv = '[aid="radio-option-MNNVL is present"]';
        this.nodePoolCancelBtn = '[aid="cancel-form-footer-btn"]';
        this.newNetworkTopologyButton = '[aid="runai-select-custom-btn"]';
        this.topoNameInput = '[aid="network-topology-name-input"]';
        this.topoLabelKeyInput = '[aid="network-topology-label-key-input"]';
        this.topoLabelBtn = '[aid="network-topology-add-label-btn"]';
        this.saveTopoBtn = '[aid="network-topologies-modal-save-button"]';
        this.topoModalSearchInput = '[aid="runai-list-modal-search"]';
        this.topoModalLabels = '[aid="labels"]';
        this.tableRow = '.runai-table-row';
        this.networkTopologySection = '[aid="node-pool-network-topology-selection"]';
        this.qTable = new QTable(page);
    }

    async navigateToNodePools(): Promise<void> {
        await this.goToPage();
        logger.debug('redirect to node pools successfully');
    }

    async submitNodePoolForm(): Promise<void> {
        await this.page.locator(this.saveNodePoolBtn).click();
    }

    async openSchedulingSection(): Promise<void> {
        await this.page.locator(this.schedulingExpansionLabel).click();
    }

    async createBasicNodePool(request: NodePoolCreateRequest): Promise<void> {
        await this.page.locator(this.addNewNodePoolBtn).click();
        await this.page.waitForSelector(this.nodePoolNameInput);
        await this.fillInputBySelector(this.nodePoolNameInput, request.name);
        if (request.labelValue && request.labelKey) {
            await this.fillInputBySelector(this.nodePoolLabelKeyInput, request.labelKey);
            await this.fillInputBySelector(this.nodePoolLabelValueInput, request.labelValue);
        }
        await this.submitNodePoolForm();
        logger.debug(`Node pool ${request.name} created successfully`);
        await nodePoolUtils.waitForNodePoolCrdReadyOrEmpty(request.name, 40);
    }

    async createNodePoolWithGpuStrategy(request: NodePoolCreateRequest): Promise<void> {
        await this.page.locator(this.addNewNodePoolBtn).click();
        await this.page.waitForSelector(this.nodePoolNameInput);
        await this.fillInputBySelector(this.nodePoolNameInput, request.name);
        if (request.labelValue && request.labelKey) {
            await this.fillInputBySelector(this.nodePoolLabelKeyInput, request.labelKey);
            await this.fillInputBySelector(this.nodePoolLabelValueInput, request.labelValue);
        }

        if (request.placementStrategy && request.placementStrategy.gpu) {
            await this.openSchedulingSection();
            await this.fillGpuPlacementRadio(request.placementStrategy.gpu);
        }
        await this.submitNodePoolForm();
        await nodePoolUtils.waitForNodePoolCrdReadyOrEmpty(request.name, 40);
    }

    async deleteNodePoolAndWait(nodePoolName: string): Promise<void> {
        await this.navigateToNodePools();
        const nodePoolIndexRow = await this.getNodePoolRowIndexByName(nodePoolName);
        await this.waitForEmptyStatus(nodePoolIndexRow);
        await this.deleteNodePoolByName(nodePoolName);
        await this.waitForDeletingStatus(nodePoolIndexRow);
        const isNodePoolDeleted = async (): Promise<boolean> => this.isTableNotIncludeNodePool(nodePoolName);
        await waitUtils.waitForCondition(isNodePoolDeleted, WAIT_FOR_NODE_POOL_REMOVED_MSG);
    }

    async waitForStatus(nodePoolIndex: number, status: string): Promise<boolean> {
        const nodePoolStatusCell: Array<string> = await this.page
            .locator(this.nodePoolStatusCol)
            .allTextContents();
        return nodePoolStatusCell[nodePoolIndex] === status;
    }

    async waitForReadyStatus(nodePoolIndex: number): Promise<boolean> {
        return await this.waitForStatus(nodePoolIndex, 'Ready');
    }

    async waitForEmptyStatus(nodePoolIndex: number): Promise<boolean> {
        return await this.waitForStatus(nodePoolIndex, 'Empty');
    }

    async waitForDeletingStatus(nodePoolIndex: number): Promise<boolean> {
        return await this.waitForStatus(nodePoolIndex, 'Deleting...');
    }

    async clickOnNodePoolRowByName(name: string): Promise<void> {
        return await this.page
            .locator(`.runai-table-row > td${this.nodePoolNameCol}:has-text("${name}")`)
            .click();
    }

    async clickOnCellByRowIndex(cell: string, nodePoolIndexRow: number): Promise<void> {
        await this.page.locator(`${cell}>> nth=${nodePoolIndexRow}`).click();
    }

    async clickOnEditButton(): Promise<void> {
        await this.page.locator('.runai-action-button:has-text("edit")').click();
    }

    async clickOnDeleteButton(): Promise<void> {
        await this.page.locator('.runai-action-button:has-text("delete")').click();
    }

    async getNodePoolRowIndexByName(nodePoolName: string): Promise<number> {
        const nodePoolsNames: string[] = await this.page.locator(this.nodePoolNameCol).allTextContents();
        return nodePoolsNames.findIndex((name: string) => name === nodePoolName);
    }

    async deleteNodePoolByName(nodePoolName: string): Promise<void> {
        await this.clickOnNodePoolRowByName(nodePoolName);
        await this.clickOnDeleteButton();
        await this.page.locator(this.deleteNodePoolModalBtn).click();
    }

    async getNodePoolsNames(): Promise<string[]> {
        return await this.page.locator(this.nodePoolNameCol).allTextContents();
    }

    async isTableIncludeNodePool(nodePoolName: string): Promise<boolean> {
        const nodePoolsNames: string[] = await this.getNodePoolsNames();
        return nodePoolsNames.includes(nodePoolName);
    }

    async isTableNotIncludeNodePool(nodePoolName: string): Promise<boolean> {
        const nodePoolsNames: string[] = await this.getNodePoolsNames();
        return !nodePoolsNames.includes(nodePoolName);
    }

    async fillCpuPlacementRadio(input: PlacementStrategy.CpuEnum): Promise<void> {
        switch (input) {
            case PlacementStrategy.CpuEnum.Binpack:
                await this.page.locator(this.cpuPlacementRadioBinpack).click();
                break;
            case PlacementStrategy.CpuEnum.Spread:
                await this.page.locator(this.cpuPlacementRadioSpread).click();
                break;
        }
    }

    async fillGpuPlacementRadio(input: PlacementStrategy.GpuEnum): Promise<void> {
        switch (input) {
            case PlacementStrategy.GpuEnum.Binpack:
                await this.page.locator(this.gpuPlacementRadioBinpack).click();
                break;
            case PlacementStrategy.GpuEnum.Spread:
                await this.page.locator(this.gpuPlacementRadioSpread).click();
                break;
        }
    }

    async openEditForm(name: string): Promise<void> {
        await this.clickOnNodePoolRowByName(name);
        await this.clickOnEditButton();
    }

    async clickGpuNetworkAccelerationLabel(): Promise<void> {
        await this.page.locator(this.gpuNetowrkAccelerationExpansionLabel).click();
    }

    async attachTopology(topoName: string): Promise<void> {
        await this.clickGpuNetworkAccelerationLabel();
        await this.page.getByRole('combobox', { name: this.networkTopologyComoboxName }).click();
        await this.page.getByRole('option', { name: topoName }).click();
    }

    async detachTopology(): Promise<void> {
        await this.clickGpuNetworkAccelerationLabel();
        await this.page.locator(this.networkTopologySection).getByRole('button', { name: 'Clear' }).click();
    }

    async clickMnnvlIsPresent(): Promise<void> {
        await this.page.locator(this.mnnvlIsPresentDiv).click();
    }

    async getNodePoolData(name: string): Promise<Record<string, string>> {
        return await this.qTable.getEntityData('Node pool', name);
    }

    async openNewNetworkTopologyForm(): Promise<void> {
        await this.clickGpuNetworkAccelerationLabel();
        await this.page.getByRole('combobox', { name: this.networkTopologyComoboxName }).click();
        await this.page.locator(this.newNetworkTopologyButton).click();
    }

    async clickNewTopoLabelBtn(): Promise<void> {
        await this.page.locator(this.topoLabelBtn).click();
    }

    async clickSaveTopoBtn(): Promise<void> {
        await this.page.locator(this.saveTopoBtn).click();
    }

    async fillTopoLabelInputByIndex(index: number, text: string): Promise<void> {
        const input = this.page.locator(this.topoLabelKeyInput).nth(index);
        await input.waitFor({ state: 'visible' });
        await input.fill(text);
    }

    async fillTopoNameInput(name: string): Promise<void> {
        await this.page.locator(this.topoNameInput).fill(name);
    }

    async createTopo(name: string, labels: Array<string>): Promise<void> {
        await this.openNewNetworkTopologyForm();
        await this.fillTopoNameInput(name);

        for (let i = 0; i < labels.length; i++) {
            if (i !== 0) {
                await this.clickNewTopoLabelBtn();
            }
            await this.fillTopoLabelInputByIndex(i, labels[i]);
        }

        await this.clickSaveTopoBtn();
        await this.page.waitForResponse(
            (response) => response.url().includes('/api/v1/network-topologies') && response.status() === 201
        );
    }

    async clickCancelBtn(): Promise<void> {
        await this.page.locator(this.nodePoolCancelBtn).click();
        await this.clickModalLeaveBtn();
    }

    async openNodePoolTopoModal(nodePoolName: string): Promise<void> {
        const rows = this.locator(this.tableRow);
        const rowCount = await rows.count();

        for (let i = 0; i < rowCount; i++) {
            const row = rows.nth(i);
            const name = await row.locator(this.nodePoolNameCol).textContent();
            if (name === nodePoolName) {
                await row.locator(this.nodePoolTopoCol).click();
                await this.page.waitForLoadState('load');
                await this.page.waitForTimeout(200);
                break;
            }
        }
    }

    async getTopoLabels(): Promise<string[]> {
        await this.page.waitForSelector(this.topoModalLabels, { state: 'visible', timeout: 3000 });
        return await this.page.locator(this.topoModalLabels).allTextContents();
    }
}
