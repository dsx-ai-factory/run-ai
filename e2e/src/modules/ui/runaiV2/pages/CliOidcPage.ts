/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { BaseV2Page } from './BaseV2Page';
import { config } from '../../../helpers';
// import { expect } from 'chai';

export class CliOidcPage extends BaseV2Page {
    codeBlock: string;

    messageContent: string;

    constructor(page: Page) {
        super(page);
        this.codeBlock = '.copy-command-wrapper.code-block pre';
        this.messageContent = '.oidc-index .content';
    }

    async openPage(code: string): Promise<void> {
        await this.page.goto(`${config.tenantUrl}/oauth-code?code=${code}`);
    }

    async getCodeBlockContent(): Promise<string> {
        return this.page.locator(this.codeBlock).textContent();
    }

    async getMessageContent(): Promise<string> {
        return this.page.locator(this.messageContent).textContent();
    }
}
