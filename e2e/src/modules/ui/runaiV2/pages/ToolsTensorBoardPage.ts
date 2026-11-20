/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from '@playwright/test';
import { ToolsBasePage } from './ToolsBasePage';

export class TensorBoardPage extends ToolsBasePage {
    brandTitle: string;

    pluginsComponentElement: string;

    noActivePluginElement: string;

    noActivePluginText: string;

    constructor(page: Page, name: string) {
        super(page, 'TensorBoard', name);
        this.pluginsComponentElement = 'plugins-component';
        this.brandTitle = '.brand:has-text("TensorBoard")';
        this.noActivePluginElement = 'h3.no-active-plugin';
        this.noActivePluginText = 'No dashboards are active for the current data set.';
    }

    async checkForTensorBoardPage(expectedUrl: string): Promise<void> {
        await super.checkForPageUrlAndTitle(expectedUrl);
        await this.checkForTensorBoardPageElements();
    }

    // Verify the elements on Jupyter page are visible
    async checkForTensorBoardPageElements(): Promise<void> {
        await this.page.waitForSelector(this.pluginsComponentElement, { state: 'visible', timeout: 50000 });
        await super.checkForElementVisible(this.brandTitle);
        await super.checkForElementIncludeText(this.noActivePluginElement, this.noActivePluginText);
    }
}
