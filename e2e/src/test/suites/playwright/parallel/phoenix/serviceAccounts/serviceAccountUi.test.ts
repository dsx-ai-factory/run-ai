/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    serviceAccountUtils,
    tags
} from '../../../../../../modules/helpers';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { ApplicationsPage } from '../../../../../../modules/ui/runaiV2/pages/ApplicationPage';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { expect } from 'chai';
import { test } from '@playwright/test';
import { ServiceAccount } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';

test.describe(`Service Account UI Tests, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, function () {
    let browserEngine: BrowserEngine;
    let videoName: string;
    let appUiPage: ApplicationsPage;
    let auth: AuthPage;
    let appToDelete: ServiceAccount;
    test.beforeEach(async function () {
        browserEngine = new BrowserEngine();
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
        appUiPage = new ApplicationsPage(browserEngine.page);
    });

    test.afterAll('stop browser', async function ({}, testInfo) {
        await browserEngine.stopAndArchive(testInfo, videoName);
        const serviceAccounts = await serviceAccountUtils.getServiceAccounts(
            flowRootState.users.system_administrator.accessToken
        );
        appToDelete = serviceAccounts.body.find((sa) => sa.name === 'service-account-appui');
        if (appToDelete) {
            await serviceAccountUtils.deleteServiceAccount(
                flowRootState.users.system_administrator.accessToken,
                appToDelete.id
            );
        }
    });

    test('should crud service account', async function () {
        const serviceAccountName = 'service-account-appui';
        videoName = 'service-account-ui';
        await appUiPage.navigateToServiceAccountsPage();
        await appUiPage.createServiceAccount(serviceAccountName);
        const serviceAccount = await appUiPage.validateServiceAccountExist(serviceAccountName);
        expect(serviceAccount, 'service account created via UI should be visible in the list with the correct name').to.equal(serviceAccountName);
        await appUiPage.regenerateServiceAccountSecret(serviceAccountName);
        await appUiPage.deleteServiceAccount(serviceAccountName);
        const appNameLocator = appUiPage.page.locator(
            `[aid="serviceAccount"]:has-text("${serviceAccountName}")`
        );
        const count = await appNameLocator.count();
        expect(count).to.equal(0, 'Service account should not be visible after deletion');
    });
});
