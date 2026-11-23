/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { config, flowRootState, tags, tenantUtils } from '../../../../../../modules/helpers';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { LoginUi } from '../../../../../../modules/ui/runaiV2/pages/LoginUiUtils';
import { test } from '@playwright/test';

test.describe(`CLI login: browser-cli-auth, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.runOn.parallelCI}`, function () {
    let browserEngine: BrowserEngine;
    let loginUi: LoginUi;

    test.beforeAll(async function () {
        browserEngine = new BrowserEngine();
        await browserEngine.start();
        loginUi = new LoginUi(browserEngine.currentPage);
    });

    test.afterAll(async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, 'browser-cli-auth');
    });

    test('should be able to login to the CLI with a browser and get code token', async function () {
        const tenant = await tenantUtils.getTenant(flowRootState.users.system_administrator.accessToken);
        const code = await loginUi.browserCliAuthGetToken(config.appUrl, tenant.name, config.tenantUrl);
        expect(code, 'browser CLI auth should return a code token').to.exist;
        expect(code, 'returned code token should not be empty').to.not.be.empty;
        expect(code, 'returned code token should be a string').to.be.a('string');
        expect(code, 'returned code token should be 98 characters long').to.have.lengthOf(98);
    });
});
