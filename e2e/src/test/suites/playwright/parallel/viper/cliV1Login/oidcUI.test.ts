/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { CliOidcPage } from '../../../../../../modules/ui/runaiV2/pages/CliOidcPage';
import { logger, tags } from '../../../../../../modules/helpers';
import { test, expect } from '@playwright/test';

test.describe(`CLI login UI, tags: ${tags.tool.ui} ${tags.resource.workspace} ${tags.teams.viper}`, function () {
    let videoName: string;
    let browserEngine: BrowserEngine;
    let cliOidcPage: CliOidcPage;

    test.beforeAll('creating browser engine', function () {
        browserEngine = new BrowserEngine();
    });

    test.beforeEach('start browser', async function () {
        await browserEngine.start();
        cliOidcPage = new CliOidcPage(browserEngine.currentPage);
    });

    test.afterEach('stop browser', async ({}, testInfo) => {
        try {
            await browserEngine.stopAndArchive(testInfo, 'oidc-' + videoName);
        } catch (e) {
            logger.error(e);
        }
    });

    test('should check the "/oauth-code" is working fine when a code is provided in the url', async function () {
        videoName = 'oauth-code-page';
        const code = '12345678';
        await cliOidcPage.openPage(code);
        const codeContent = await cliOidcPage.getCodeBlockContent();
        expect(codeContent).toEqual(code);

        const messageContent = await cliOidcPage.getMessageContent();
        expect(messageContent).toContain('Login Successful');
    });

    test('should check the "/oauth-code" is working fine when no code is provided in the url', async function () {
        videoName = 'oauth-code-page';
        const code = '';
        await cliOidcPage.openPage(code);
        const messageContent = await cliOidcPage.getMessageContent();
        expect(messageContent).toContain('No OIDC auth code detected');
    });
});
