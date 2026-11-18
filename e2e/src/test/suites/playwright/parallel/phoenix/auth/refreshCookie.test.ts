/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authorizationUtils,
    flowRootState,
    logger,
    objectGenerator,
    usersManager
} from '../../../../../../modules/helpers';
import { tags } from '../../../../../../modules/helpers';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { test } from '@playwright/test';
import { ScopeType, SubjectType } from '../../../../../../modules/api/apis/authorizationApi/generated/api';
import { IncomingMessage } from 'node:http';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';

test.describe(
    `Refresh cookie lifecycle on login and logout, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`,
    { tag: [tags.services.frontend, tags.services.identityManager] },
    function () {
        let userName: string;
        let password: string;
        let user: { response: IncomingMessage; body: UserPostResponse };
        let browserEngine: BrowserEngine;
        let auth: AuthPage;

        test.beforeEach(async function () {
            user = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                objectGenerator.generateEmailByPrefix(`refreshcookie-${Date.now()}`),
                false
            );
            userName = user.body.username;
            password = user.body.tempPassword;
            await authorizationUtils.createAccessRuleWithDefaults(
                userName,
                SubjectType.User,
                flowRootState.users.system_administrator.role.id,
                flowRootState.cluster.tenantId.toString(),
                ScopeType.Tenant,
                flowRootState.users.system_administrator.accessToken
            );
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
        });

        test.afterEach(async function () {
            try {
                await browserEngine.browser.close().catch(() => {});
            } catch (error) {
                logger.error(`Error closing browser: ${error}`);
            }
            await usersManager.deleteLocalUser(
                flowRootState.users.system_administrator.accessToken,
                user.body.id
            );
        });

        test('should set refresh-token cookie on login', async function () {
            await auth.signIn(userName, password);
            logger.info('User logged in successfully');

            const cookies = await browserEngine.context.cookies();
            const refreshCookie = cookies.find((c) => c.name === 'refresh-token');
            logger.info(`Cookies after login: ${JSON.stringify(cookies.map((c) => c.name))}`);

            expect(refreshCookie, 'refresh-token cookie should be present after login').to.not.be.undefined;
            expect(refreshCookie?.httpOnly, 'refresh-token cookie should be httpOnly').to.be.true;
            expect(refreshCookie?.secure, 'refresh-token cookie should be secure').to.be.true;
            expect(refreshCookie?.sameSite, 'refresh-token cookie should be SameSite=Strict').to.equal('Strict');

            const now = Date.now() / 1000;
            const expiresIn = refreshCookie.expires - now;
            const twentyFourHoursInSeconds = 24 * 60 * 60;
            const toleranceInSeconds = 5 * 60; // 5 minutes tolerance

            logger.info(`refresh-token cookie expires in ${Math.round((expiresIn / 3600) * 100) / 100} hours`);
            expect(expiresIn, 'refresh-token cookie should expire in ~24 hours').to.be.within(
                twentyFourHoursInSeconds - toleranceInSeconds,
                twentyFourHoursInSeconds + toleranceInSeconds
            );
        });

        test('should clear refresh-token cookie on logout', async function () {
            await auth.signIn(userName, password);
            logger.info('User logged in successfully');

            const cookiesBeforeLogout = await browserEngine.context.cookies();
            const refreshCookieBefore = cookiesBeforeLogout.find((c) => c.name === 'refresh-token');
            expect(refreshCookieBefore, 'refresh-token cookie should be present before logout').to.not.be.undefined;
            logger.info(`refresh-token cookie confirmed present before logout (domain: ${refreshCookieBefore?.domain}, path: ${refreshCookieBefore?.path})`);

            await browserEngine.currentPage.locator('[aid="profile-btn"]').click();
            await browserEngine.currentPage.waitForSelector('[aid="sign-out-btn"]', { timeout: 15000 });
            const navigationPromise = browserEngine.currentPage.waitForNavigation({
                waitUntil: 'networkidle',
                timeout: 30000
            });
            await browserEngine.currentPage.click('[aid="sign-out-btn"]');
            await navigationPromise;
            await browserEngine.currentPage.waitForSelector('#username', { timeout: 20000 });
            logger.info('User logged out successfully');

            const appUrl = browserEngine.currentPage.url().replace(/\/auth\/.*/, '');
            const cookiesAfterLogout = await browserEngine.context.cookies(appUrl);
            const refreshCookieAfter = cookiesAfterLogout.find((c) => c.name === 'refresh-token');
            logger.info(`Cookies after logout (app domain): ${JSON.stringify(cookiesAfterLogout.map((c) => c.name))}`);

            expect(refreshCookieAfter, 'refresh-token cookie should be cleared after logout').to.be.undefined;
        });
    }
);
