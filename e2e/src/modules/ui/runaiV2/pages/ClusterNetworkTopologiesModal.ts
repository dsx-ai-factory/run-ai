/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { BaseV2Page } from './BaseV2Page';
import { QTable } from '../components/QTable';

export class ClusterNetworkTopologiesModal extends BaseV2Page {
    searchInput: string;

    topologyName: string;

    labels: string;

    createdBy: string;

    creationTime: string;

    closeBtn: string;

    qTable: QTable;

    constructor(page: Page) {
        super(page);
        this.searchInput = '[aid="runai-list-modal-search"]';
        this.topologyName = '.topologies-list-table [aid="name"]';
        this.labels = '.topologies-list-table [aid="labels"]';
        this.createdBy = '.topologies-list-table [aid="created-by"]';
        this.creationTime = '.topologies-list-table [aid="created-at"]';
        this.closeBtn = '[aid="close-topologies-preview-modal-btn"]';
        this.qTable = new QTable(page, '.topologies-list-table .q-table');
    }

    async getTopoData(topoName: string): Promise<Record<string, string>> {
        return this.qTable.getEntityData('Topology', topoName);
    }

    async closeModal(): Promise<void> {
        await this.page.locator(this.closeBtn).click();
    }
}
