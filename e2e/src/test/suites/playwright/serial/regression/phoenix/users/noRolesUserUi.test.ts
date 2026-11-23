/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    flowRootState,
    logger,
    objectGenerator,
    settings,
    tags,
    usersManager
} from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { LoginUi } from '../../../../../../../modules/ui/runaiV2/pages/LoginUiUtils';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { UserPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';

test.describe(`User without roles, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.users}`, function () {
    shouldRunTests(settings.isSelfHosted)(function () {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let loginUi: LoginUi;
        let user: { response: IncomingMessage; body: UserPostResponse };

        test.beforeEach(async () => {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            loginUi = new LoginUi(browserEngine.currentPage);
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'user without roles');
            if (user) {
                await usersManager.deleteLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    user.body.id
                );
            }
        });

        test('should login without permission and validate top bar and logout', async function () {
            const nonPermissionUser = objectGenerator.generateEmailByPrefix('user-without-roles');
            user = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                nonPermissionUser,
                false
            );
            await auth.signIn(user.body.username, user.body.tempPassword);
            expect(await loginUi.nonPermissionLoginValidator()).to.be.true;
            logger.debug(
                `user ${user.body.username}, password:${user.body.tempPassword} created successfully and logged in successfully without roles`
            );
            expect(await loginUi.topBarValidator()).to.be.true;
            expect(await loginUi.logoutToLoginPage()).to.be.true;
        });

        test('user should see brand icons and validate open links', async function () {
            const nonPermissionUser = objectGenerator.generateEmailByPrefix('user-without-roles2');
            user = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                nonPermissionUser,
                false
            );
            await auth.signIn(user.body.username, user.body.tempPassword);
            expect(await loginUi.nonPermissionLoginValidator()).to.be.true;
            expect(await loginUi.tenantIconValidator()).to.be.true;
            await loginUi.openLinks();
        });
    });
});
