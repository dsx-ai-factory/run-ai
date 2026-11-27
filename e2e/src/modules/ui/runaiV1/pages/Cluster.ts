/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { logger } from '../../../helpers';
import { BaseRunaiPage } from './BaseRunaiPage';

export class Cluster extends BaseRunaiPage {
    clusterIcon: string;

    connectedCluster: string;

    statusHeaderRow: string;

    statusOfCluster: string;

    iframe: string;

    constructor(videoFilePath: string, videoFileType: string) {
        super(videoFilePath, videoFileType);
        this.clusterIcon = '[aid="menu-icon-clusters"]';
        this.connectedCluster = '//*[@aid="status"]';
        this.statusHeaderRow = 'text=Status';
        this.statusOfCluster = '[aid="status"]';
        this.iframe = '#runai-new-app';
    }

    async navigateToCluster(): Promise<void> {
        logger.debug('Navigation to Cluster');
        await this.page.locator(this.clusterIcon).click();
    }

    async clusterConnected(): Promise<void> {
        await this.navigateToCluster();
        await this.page.locator(this.statusHeaderRow).isEnabled({ timeout: 30000 });
        await this.page.locator(this.statusOfCluster).click({ timeout: 30000 });
        await this.page.locator(this.connectedCluster).isEnabled({ timeout: 30000 });
        expect(await this.page.locator(this.connectedCluster).innerText()).to.contains('Connected');
        logger.debug('Cluster connected');
    }

    async clusterConnectedV1(): Promise<void> {
        await this.navigateToCluster();
        await this.page.frameLocator(this.iframe).locator(this.statusHeaderRow).isEnabled({ timeout: 30000 });
        await this.page
            .frameLocator(this.iframe)
            .locator(this.connectedCluster)
            .isEnabled({ timeout: 30000 });
        expect(
            await this.page.frameLocator(this.iframe).locator(this.connectedCluster).innerText()
        ).to.be.equal('Connected');
        logger.debug('Cluster connected');
    }
}
