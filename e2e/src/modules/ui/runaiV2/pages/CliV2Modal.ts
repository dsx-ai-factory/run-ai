/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseV2Page } from './BaseV2Page';
import { Page } from 'playwright';
import { BrowserEngine } from './BrowserEngine';

export class CliV2Modal extends BaseV2Page {
    browseEngine: BrowserEngine;

    helpDropdownSelector = 'button[aid="help-dropdown"]';

    cliDropdownSelector = '[aid="researcher-cli-option"]';

    cliDocLinkSelector = 'a[aid="cli-documentation-link"]';

    installDocLinkSelector = 'a[aid="cli-install-doc-link"]';

    linuxDownloadSelector = 'div[aid="radio-option-Linux"]';

    constructor(browseEngine: BrowserEngine) {
        super(browseEngine.page);
        this.browseEngine = browseEngine;
        this.helpDropdownSelector = 'button[aid="help-dropdown"]';
        this.cliDropdownSelector = '[aid="researcher-cli-option"]';
        this.cliDocLinkSelector = 'a[aid="cli-documentation-link"]';
        this.installDocLinkSelector = 'a[aid="cli-install-doc-link"]';
        this.linuxDownloadSelector = 'div[aid="radio-option-Linux"]';
    }

    private async openNewPage(selector: string): Promise<Page> {
        const [newPage] = await Promise.all([
            this.browseEngine.context.waitForEvent('page'),
            this.browseEngine.page.click(selector)
        ]);
        await newPage.waitForLoadState('networkidle');
        return newPage;
    }

    async openCliDoc(): Promise<Page> {
        return await this.openNewPage(this.cliDocLinkSelector);
    }

    async openInstallCliDoc(): Promise<Page> {
        return await this.openNewPage(this.installDocLinkSelector);
    }

    async getLinuxInstallCommand(): Promise<string> {
        await this.page.click(this.linuxDownloadSelector);
        await this.page.waitForSelector('pre');
        const uiInstallCommand = await this.page.evaluate(() => {
            const preElement = document.querySelector('pre');
            return preElement ? preElement.textContent : null;
        });
        return uiInstallCommand;
    }

    async openCliDialog(): Promise<void> {
        await this.page.click(this.helpDropdownSelector);
        await this.page.click(this.cliDropdownSelector);
    }
}
