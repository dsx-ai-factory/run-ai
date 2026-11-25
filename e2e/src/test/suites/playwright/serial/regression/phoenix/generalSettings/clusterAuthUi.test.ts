/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { config, flowRootState, settings, tags, tenantUtils } from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { GeneralSettings } from '../../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { test, expect } from '@playwright/test';

test.describe(`General Settings authentication UI Validations, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, { tag: [tags.services.tempBackendSettings] }, () => {
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let generalSettings: GeneralSettings;
    let videoName: string;

    test.beforeEach('init browsers and classes', async function () {
        browserEngine = new BrowserEngine();
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
        generalSettings = new GeneralSettings(browserEngine.currentPage);
    });

    test.afterEach('stop browserEngine', async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, videoName);
    });

    test('should check if cluster authentication issuer-url is visible and correct', async function () {
        videoName = 'shouldCheckIfClusterAuthenticationIssuerUrlIsVisibleAndCorrect';
        await generalSettings.navigateToGeneralSettingsAuthentication();
        const responseTenantName = await tenantUtils.getTenantConfiguration(
            flowRootState.users.system_administrator.accessToken
        );
        const expectedBaseUrl = config.tenantUrl;
        const expectedTenantName = responseTenantName.tenantName;
        const validation = await generalSettings.validateOidcIssuerUrl(
            expectedBaseUrl,
            expectedTenantName,
            settings.isSelfHosted
        );
        expect(validation.isVisible, 'cluster authentication issuer URL should be visible').toBe(true);
        expect(validation.url, 'cluster authentication issuer URL should be truthy/non-empty').toBeTruthy();
        expect(validation.isValid, 'cluster authentication issuer URL should be valid').toBe(true);
    });
});
