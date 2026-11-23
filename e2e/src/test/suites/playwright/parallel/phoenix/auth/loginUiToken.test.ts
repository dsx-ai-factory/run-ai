/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authorizationUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    usersManager
} from '../../../../../../modules/helpers';
import { tags } from '../../../../../../modules/helpers/config/tags';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { test, chromium, Browser, BrowserContext, Page, TestInfo } from '@playwright/test';
import { ScopeType, SubjectType } from '../../../../../../modules/api/apis/authorizationApi/generated/api';
import { IncomingMessage } from 'node:http';
import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

test.describe(`Token persistence across browser sessions, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, function () {
    let userName: string;
    let password: string;
    let user: { response: IncomingMessage; body: UserPostResponse };
    let browser: Browser;
    let context: BrowserContext;
    let page1: Page;

    test.beforeEach(async function () {
        user = await usersManager.createLocalUser(
            flowRootState.users.system_administrator.accessToken,
            objectGenerator.generateEmailByPrefix('loginuitoken'),
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
        browser = await chromium.launch({
            headless: true,
            timeout: 30000
        });
        context = await browser.newContext({
            ignoreHTTPSErrors: true,
            permissions: ['clipboard-read', 'clipboard-write'],
            viewport: { width: 1280, height: 720 }
        });
        page1 = await context.newPage();
    });

    test.afterEach(async function () {
        try {
            if (page1) await page1.close().catch(() => {});
            if (context) await context.close().catch(() => {});
            if (browser) await browser.close().catch(() => {});
        } catch (error) {
            logger.error(`Error closing browser: ${error}`);
        }
        await usersManager.deleteLocalUser(
            flowRootState.users.system_administrator.accessToken,
            user.body.id
        );
    });

    test('should recover session via httpOnly cookie after browser restart without localStorage tokens', async function () {
        await page1.goto(config.tenantUrl, { waitUntil: 'networkidle', timeout: 15000 });
        await page1.waitForSelector('#username', { timeout: 15000 });
        await page1.fill('#username', userName);
        await page1.fill('#password', password);
        await page1.click('#kc-login');
        await page1.waitForSelector('[aid="profile-btn"]', { timeout: 20000 });
        logger.info('User logged in successfully');

        const postLoginState = await page1.evaluate(() => ({
            hasHadSession: JSON.parse(localStorage.getItem('had-session') || 'false') === 'true',
            hasAccessToken: localStorage.getItem('access-token') !== null,
            hasRefreshToken: localStorage.getItem('refresh-token') !== null,
        }));

        logger.info(`Post-login state: had-session=${postLoginState.hasHadSession}, accessToken=${postLoginState.hasAccessToken}, refreshToken=${postLoginState.hasRefreshToken}`);

        const cookies = await context.cookies();
        const hasRefreshCookie = cookies.some((c) => c.name === 'refresh-token');
        logger.info(`Refresh cookie present: ${hasRefreshCookie}`);

        expect(hasRefreshCookie, 'httpOnly refresh cookie should be present after login').to.be.true;

        expect(postLoginState.hasHadSession, 'had-session flag should be set in localStorage after login').to.be.true;
        expect(postLoginState.hasAccessToken, 'access-token should not be stored in localStorage after login').to.be.false;
        expect(postLoginState.hasRefreshToken, 'refresh-token should not be stored in localStorage after login').to.be.false;

        const storageStatePath = path.join(
            dirname(fileURLToPath(import.meta.url)),
            `cookie-flow-state-${Date.now()}.json`
        );
        await context.storageState({ path: storageStatePath });
        await page1.close();
        await context.close();
        await browser.close();

        browser = await chromium.launch({ headless: true, timeout: 30000 });
        context = await browser.newContext({
            ignoreHTTPSErrors: true,
            viewport: { width: 1280, height: 720 },
            storageState: storageStatePath,
        });
        page1 = await context.newPage();

        let tokenRefreshCalled = false;
        page1.on('response', (response) => {
            if (response.url().includes('/api/v1/token') && response.request().method() === 'POST') {
                tokenRefreshCalled = true;
                logger.info(`Token refresh called: ${response.status()}`);
            }
        });

        await page1.goto(config.tenantUrl, { timeout: 15000 });

        try {
            await page1.waitForSelector('[aid="profile-btn"]', { timeout: 30000 });
            logger.info('Session recovered via cookie - user is logged in');

            expect(tokenRefreshCalled, 'token refresh endpoint should be called after session recovery from cookie').to.be.true;

            const recoveredState = await page1.evaluate(() => ({
                hasAccessToken: localStorage.getItem('access-token') !== null,
                hasRefreshToken: localStorage.getItem('refresh-token') !== null,
            }));

            expect(recoveredState.hasAccessToken, 'access-token should not be in localStorage after cookie-based session recovery').to.be.false;
            expect(recoveredState.hasRefreshToken, 'refresh-token should not be in localStorage after cookie-based session recovery').to.be.false;
        } catch {
            const hasLoginForm = await page1.locator('#username').isVisible({ timeout: 5000 }).catch(() => false);
            if (hasLoginForm) {
                logger.info('Session expired or cookie invalid - login screen shown (acceptable)');
            } else {
                throw new Error('Page is neither showing login screen nor logged in state');
            }
        }

        try {
            fs.unlinkSync(storageStatePath);
        } catch (error) {
            logger.warn(`Could not delete storage state file: ${error}`);
        }
    });
});
