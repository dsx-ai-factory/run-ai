/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { config, logger, tags } from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { LocalUserSettings } from '../../../../../../modules/ui/runaiV2/pages/localUserSettings';
import { test } from '@playwright/test';

test.describe(`Setting Page UI tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.ui}`, () => {
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let localUserSettings: LocalUserSettings;
    let videoName: string;

    test.beforeAll('init browserEngine', function () {
        browserEngine = new BrowserEngine();
    });

    test.beforeEach('init browsers and classes', async function () {
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
        localUserSettings = new LocalUserSettings(browserEngine.currentPage);
    });

    test.afterEach('stop browserEngine', async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, videoName);
    });

    test('should display the setting page', async () => {
        videoName = 'settings-page';
        logger.info('should display the form');
        await localUserSettings.openLocalSettingsPage();
    });
});
