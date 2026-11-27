/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { UsersFilterSortFields } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { flowRootState, logger, tags, usersManager, usersUtils } from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { UsersV2 } from '../../../../../../modules/ui/runaiV2/pages/UsersV2';
import { test } from '@playwright/test';
import { expect } from '@playwright/test';

test.describe(`Duplicate Email Error On User Creation UI, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.ui}`, () => {
    let videoName: string;
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let users: UsersV2;

    test.beforeEach('start browser', async function () {
        browserEngine = new BrowserEngine();
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        users = new UsersV2(browserEngine.currentPage);
    });

    test.afterEach('stop browser', async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, videoName);
    });

    test.afterAll('delete users if test failed', async function () {
        try {
            const user = await usersUtils.getUsersFilteredBy(
                flowRootState.users.system_administrator.accessToken,
                ['username=@e2e-e2e@example.com'],
                UsersFilterSortFields.Username
            );
            if (user.length > 0) {
                await usersManager.deleteLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    user[0].id
                );
            }
            const user2 = await usersUtils.getUsersFilteredBy(
                flowRootState.users.system_administrator.accessToken,
                ['username=@e2e@example.com'],
                UsersFilterSortFields.Username
            );
            if (user2.length > 0) {
                await usersManager.deleteLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    user2[0].id
                );
            }
        } catch (error) {
            logger.error(`Error deleting users: ${error}`);
        }
    });

    test('should create user with duplicate email letters and validate there no errors', async function () {
        videoName = 'duplicate-email-error-on-user-creation';
        await auth.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
        await users.navigateToUsers();
        await users.createUser('e2e-e2e@example.com');
        await users.createUser('e2e@example.com');
        await users.searchByName('e2e@example.com');
        const error = await users.page.locator('.error, .alert').all();
        expect(error, 'no error or alert elements should appear after creating user with similar email').toEqual([]);
        await users.clickOnUser('e2e@example.com');
        await users.clickOnDeleteUserButton();
        await users.clickOnUser('e2e-e2e@example.com');
        await users.clickOnDeleteUserButton();
    });
});
