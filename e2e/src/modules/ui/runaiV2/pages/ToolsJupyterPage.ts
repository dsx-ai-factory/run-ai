/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from '@playwright/test';
import { logger } from '../../../helpers';
import { ToolsBasePage } from './ToolsBasePage';

export class JupyterPage extends ToolsBasePage {
    launcherTitle: string;

    notebookSectionTitle: string;

    consoleSectionTitle: string;

    otherSectionTitle: string;

    constructor(page: Page, name: string) {
        super(page, 'JupyterLab', name);
        this.launcherTitle = '[title="New Launcher"]';
        this.notebookSectionTitle = '.jp-Launcher-sectionTitle:has-text("Notebook")';
        this.consoleSectionTitle = '.jp-Launcher-sectionTitle:has-text("Console")';
        this.otherSectionTitle = '.jp-Launcher-sectionTitle:has-text("Other")';
    }

    async checkForJupyterPage(expectedUrl: string): Promise<void> {
        await super.checkForPageUrlAndTitle(expectedUrl);
        await this.checkForJupyterPageElements();
    }

    // Verify the elements on Jupyter page are visible
    async checkForJupyterPageElements(): Promise<void> {
        await this.page.waitForSelector(this.launcherTitle, { state: 'visible', timeout: 30000 });
        await this.checkForElementVisible(this.notebookSectionTitle);
        await this.checkForElementVisible(this.consoleSectionTitle);
        await this.checkForElementVisible(this.otherSectionTitle);
    }
}
