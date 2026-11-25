/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { BaseV2Page } from './BaseV2Page';

export class NetworkTopologyModal extends BaseV2Page {
    newTopoBtn: string;

    topoNameInput: string;

    newTopoLabelBtn: string;

    saveTopoBtn: string;

    closeTopoModalBtn: string;

    topoDiv: string;

    topoLabelKeyInput: string;

    topoDeleteBtn: string;

    constructor(page: Page) {
        super(page);
        this.newTopoBtn = '[aid="add-network-topology-btn"]';
        this.topoNameInput = '[aid="network-topology-name-input"]';
        this.newTopoLabelBtn = '[aid="network-topology-add-label-btn"]';
        this.saveTopoBtn = '[aid="save-form-footer-btn"]';
        this.closeTopoModalBtn = '[aid="network-topologies-modal-cancel-button"]';
        this.topoDiv = '[aid^="network-topology-card-item-"]';
        this.topoLabelKeyInput = '[aid="network-topology-label-key-input"]';
        this.topoDeleteBtn = '[aid="card-item-delete-button"]';
    }

    async clickNewTopoBtn(): Promise<void> {
        await this.page.locator(this.newTopoBtn).click();
    }

    async fillTopoNameInput(name: string): Promise<void> {
        await this.page.locator(this.topoNameInput).fill(name);
    }

    async clickNewTopoLabelBtn(): Promise<void> {
        await this.page.locator(this.newTopoLabelBtn).click();
    }

    async clickSaveTopoBtn(): Promise<void> {
        await this.page.locator(this.saveTopoBtn).click();
    }

    async closeTopoModal(): Promise<void> {
        await this.page.locator(this.closeTopoModalBtn).click();
    }

    async fillTopoLabelInputByIndex(index: number, text: string): Promise<void> {
        const input = this.page.locator(this.topoLabelKeyInput).nth(index);
        await input.waitFor({ state: 'visible' });
        await input.fill(text);
    }

    async createTopo(name: string, labels: Array<string>): Promise<void> {
        await this.clickNewTopoBtn();
        await this.fillTopoNameInput(name);

        for (let i = 0; i < labels.length; i++) {
            if (i !== 0) {
                await this.clickNewTopoLabelBtn();
            }
            await this.fillTopoLabelInputByIndex(i, labels[i]);
        }

        await this.clickSaveTopoBtn();
        // Wait for the topology to be created before closing the modal, otherwise, the topology will not be created.
        await this.page.waitForResponse(
            (response) => response.url().includes('/api/v1/network-topologies') && response.status() === 201
        );
        await this.closeTopoModal();
    }

    async get_topo_name_list(): Promise<string[]> {
        await this.page.waitForSelector(this.newTopoBtn, { state: 'visible', timeout: 5000 });
        const name_list = await this.page.locator(this.topoDiv).allTextContents();
        return name_list.map((item) => item.split(':')[1].trim());
    }

    async deleteTopoByName(name: string): Promise<void> {
        const topoDivs = this.page.locator(this.topoDiv);
        const topoCnt = await topoDivs.count();
        for (let i = 0; i < topoCnt; i++) {
            const div = topoDivs.nth(i);
            const text = await div.textContent();
            const topoName = text?.split(':')[1].trim();
            if (topoName && topoName === name) {
                await div.locator(this.topoDeleteBtn).click();
                // wait for the delete request to complete
                await this.page.waitForResponse(
                    (response) =>
                        response.url().includes('/api/v1/network-topologies') && response.status() === 204
                );
                break;
            }
        }
    }
}
