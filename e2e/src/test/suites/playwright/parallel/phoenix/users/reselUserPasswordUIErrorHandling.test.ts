/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/model/userPostResponse';
import {
    authorizationUtils,
    flowRootState,
    logger,
    objectGenerator,
    tags,
    usersManager
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { UsersV2 } from '../../../../../../modules/ui/runaiV2/pages/UsersV2';
import { expect, test } from '@playwright/test';
import { IncomingMessage } from 'http';


test.describe(`Reset Password with Invalid Current Password, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.ui} ${tags.resource.users}`, () => {
    let videoName: string;
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let user: { response: IncomingMessage; body: UserPostResponse };
    let userEmail: string;
    let users: UsersV2;

    const NEW_USER_PASSWORD = objectGenerator.generatePassword();

    test.beforeAll('Create user', async function () {
        userEmail = objectGenerator.generateEmailByPrefix(`reretpass-${Date.now()}`);
        logger.debug('Creating user');
        user = await usersManager.createLocalUser(
            flowRootState.users.system_administrator.accessToken,
            userEmail,
            false
        );
        await authorizationUtils.createAccessRuleForUser(
            flowRootState.users.system_administrator.accessToken,
            user.body.username,
            flowRootState.users.system_administrator.role.id
        );
    });

    test.beforeEach('Start browser', async function () {
        browserEngine = new BrowserEngine();
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        users = new UsersV2(browserEngine.currentPage);
    });

    test.afterEach('Stop browser', async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, 'users' + videoName);
        await usersManager.deleteLocalUser(
            flowRootState.users.system_administrator.accessToken,
            user.body.id
        );
    });

    test('should reset the password with invalid current password and verify user not logged out', async function () {
        videoName = 'reset-password-invalid-current-password';
        await auth.signIn(user.body.username, user.body.tempPassword);
        await users.resetLocalUserPasswordWithInvalidCurrentPassword(
            'invalidpassword@!',
            NEW_USER_PASSWORD
        );
        const errorMessage = await users.page.locator('.q-notification__message').textContent();
        expect(errorMessage, 'submitting wrong current password should show an error notification').toContain('Current password is incorrect');
        const userNotLoggedOut = await users.verifyUserNotLoggedOut();
        expect(userNotLoggedOut, 'user should remain logged in after a failed password reset attempt').toBe(true);
    });
});
