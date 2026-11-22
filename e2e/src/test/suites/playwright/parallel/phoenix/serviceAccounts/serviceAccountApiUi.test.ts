/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    serviceAccountUtils,
    tags
} from '../../../../../../modules/helpers';
import { ServiceAccountPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { ApplicationsPage } from '../../../../../../modules/ui/runaiV2/pages/ApplicationPage';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { expect } from 'chai';
import { test } from '@playwright/test';

test.describe(`Service Accounts API & UI Integration Tests, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, function () {
    let serviceAccount: { response: IncomingMessage; body: ServiceAccountPostResponse };
    let browserEngine: BrowserEngine;
    let videoName: string;
    let appUiPage: ApplicationsPage;
    let auth: AuthPage;

    test.beforeEach(async function () {
        serviceAccount = await serviceAccountUtils.createServiceAccount(
            flowRootState.users.system_administrator.accessToken,
            { name: objectGenerator.generateShortNameByPrefix(`uiapisa-${Date.now()}`) }
        );
        browserEngine = new BrowserEngine();
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
        appUiPage = new ApplicationsPage(browserEngine.page);
    });

    test.afterEach(async ({}, testInfo) => {
        try {
            await serviceAccountUtils.deleteServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                serviceAccount.body.id
            );
        } catch (e) {
            logger.error(`Failed to delete service account ${serviceAccount.body.id} ${JSON.stringify(e)}`);
        }
    });

    test.afterAll('stop browser', async function ({}, testInfo) {
        await browserEngine.stopAndArchive(testInfo, videoName);
    });

    test('should validates service account created via API is displayed in the UI', async function () {
        videoName = 'sa-ui-api';
        const validateSaExist = await appUiPage.validateServiceAccountExist(serviceAccount.body.name);
        expect(validateSaExist, 'service account created via API should be visible in the UI with the correct name').to.equal(serviceAccount.body.name);
    });
});
