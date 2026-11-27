/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { UserPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/userPostResponse';
import {
    authorizationUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    tags,
    usersManager
} from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { UsersV2 } from '../../../../../../../modules/ui/runaiV2/pages/UsersV2';
import { test } from '@playwright/test';
import { ScopeType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { SubjectType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import http from 'node:http';

test.describe(`User Password Reset via UI tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.ui}`, function () {
    let videoName: string;
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let userLocal: { response: http.IncomingMessage; body: UserPostResponse };
    let userEmail: string;
    let usersV2: UsersV2;
    let userResetedPassword: string;

    test.beforeAll('create user', async function () {
        userEmail = objectGenerator.generateEmailByPrefix(`reset-password-${Date.now()}`);
        logger.debug(`Create user ${userEmail}`);
        userLocal = await usersManager.createLocalUser(
            flowRootState.users.system_administrator.accessToken,
            userEmail,
            false
        );
        await authorizationUtils.createAccessRuleWithDefaults(
            userLocal.body.username,
            SubjectType.User,
            flowRootState.users.system_administrator.role.id,
            flowRootState.cluster.tenantId.toString(),
            ScopeType.Tenant,
            flowRootState.users.system_administrator.accessToken
        );
    });

    test.afterAll('delete user', async function () {
        logger.debug(`Delete user ${userEmail}`);
        await usersManager.deleteLocalUser(
            flowRootState.users.system_administrator.accessToken,
            userLocal.body.id
        );
    });

    test.beforeEach('start browser', async function () {
        browserEngine = new BrowserEngine();
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        usersV2 = new UsersV2(browserEngine.currentPage);
    });

    test.afterEach('stop browser', async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, videoName);
    });

    test('should reset the password for a user via the UI', async function () {
        videoName = 'reset-password';
        await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
        await usersV2.navigateToUsers();
        userResetedPassword = await usersV2.returnUserPasswordAfterReset(userLocal.body.username);
        logger.info(`Password reset completed for user ${userLocal.body.username}, new password captured`);
    });

    test('should log in with the reset password', async function () {
        videoName = 'login-reset-password';
        await auth.signInAfterResetPassword(
            userLocal.body.username,
            userResetedPassword,
            config.controlPlaneAdminPassword
        );
    });
});
