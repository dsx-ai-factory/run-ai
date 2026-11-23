/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger, tags, config } from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { EventHistoryPage } from '../../../../../../modules/ui/runaiV2/pages/EventHistoryPage';

test.describe(`Administrator user - Only allowed to view event history, tags: ${tags.teams.loong} ${tags.runOn.dgxc} ${tags.tool.ui} ${tags.resource.users}`, () => {
    supportedFrom(SupportedVersions.V2_19)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let eventHistoryPage: EventHistoryPage;
        let videoName: string;

        test.beforeAll('init browserEngine', function () {
            browserEngine = new BrowserEngine();
        });

        test.beforeEach('init browsers and classes', async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            eventHistoryPage = new EventHistoryPage(browserEngine.currentPage);
            await authPage.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
            await eventHistoryPage.goToPage();
        });

        test.afterEach('stop browserEngine', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test('T3738096: Administrator users can only view the event history', async function () {
            videoName = 'access-event-history-page';
            logger.debug('verify view only badge is visible');
            await expect(eventHistoryPage.page.locator(eventHistoryPage.viewOnlyBadge)).toBeVisible();
        });
    });
});
