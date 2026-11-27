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
import { StatusCodes } from 'http-status-codes';

test.describe(`Token Multi-Tab Coordination Web Locks API, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, function () {
    let userName: string;
    let password: string;
    let user: { response: IncomingMessage; body: UserPostResponse };
    let browser: Browser;
    let context: BrowserContext;
    let page1: Page;
    let page2: Page;

    test.beforeEach(async function () {
        user = await usersManager.createLocalUser(
            flowRootState.users.system_administrator.accessToken,
            objectGenerator.generateEmailByPrefix('multitab'),
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
        page2 = await context.newPage();
    });

    test.afterEach(async function () {
        try {
            if (page1) await page1.close().catch(() => {});
            if (page2) await page2.close().catch(() => {});
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

    test('should handle inactive session across multiple tabs (TOKEN_IS_NOT_ACTIVE) Web Locks API (both  errors arrive together)', async function () {
        const tokenRefreshRequests = {
            tab1: [] as any[],
            tab2: [] as any[]
        };
        const tokenErrors = {
            tab1: [] as any[],
            tab2: [] as any[]
        };

        page1.on('request', (request) => {
            if (request.url().includes('/api/v1/token') && request.method() === 'POST') {
                const postData = request.postData();
                if (postData && postData.includes('refresh_token')) {
                    logger.info('[TAB 1] Token refresh request initiated');
                    tokenRefreshRequests.tab1.push({
                        url: request.url(),
                        timestamp: Date.now()
                    });
                }
            }
        });
        page1.on('response', (response) => {
            if (response.url().includes('/api/v1/token') && response.status() === 400) {
                logger.info('[TAB 1] Token refresh failed with 400');
                tokenErrors.tab1.push({ status: response.status(), timestamp: Date.now() });
            }
        });
        page2.on('request', (request) => {
            if (request.url().includes('/api/v1/token') && request.method() === 'POST') {
                const postData = request.postData();
                if (postData && postData.includes('refresh_token')) {
                    logger.info('[TAB 2] Token refresh request initiated');
                    tokenRefreshRequests.tab2.push({
                        url: request.url(),
                        timestamp: Date.now()
                    });
                }
            }
        });
        page2.on('response', (response) => {
            if (response.url().includes('/api/v1/token') && response.status() === 400) {
                logger.info('[TAB 2] Token refresh failed with 400');
                tokenErrors.tab2.push({ status: response.status(), timestamp: Date.now() });
            }
        });
        const startTime = Date.now();
        await page1.clock.install({ time: startTime });
        await page2.clock.install({ time: startTime });
        await page1.goto(config.tenantUrl, { waitUntil: 'networkidle', timeout: 15000 });
        await page1.waitForSelector('#username', { timeout: 15000 });
        await page1.fill('#username', userName);
        await page1.fill('#password', password);
        await page1.click('#kc-login');
        await page1.waitForSelector('[aid="profile-btn"]', { timeout: 20000 });
        await page2.goto(config.tenantUrl, { waitUntil: 'networkidle', timeout: 15000 });
        await page2.waitForSelector('[aid="profile-btn"]', { timeout: 20000 });
        logger.info('Tab 2: Session loaded successfully');
        const logoutResponse = await usersManager.logoutUser(
            flowRootState.users.system_administrator.accessToken,
            user.body.id
        );
        expect(logoutResponse.response.statusCode, 'logout user API should return 200 OK').to.equal(StatusCodes.OK);
        const fastForwardMs = 13 * 60 * 1000;
        await page1.clock.fastForward(fastForwardMs);
        await page2.clock.fastForward(fastForwardMs);
        logger.info('Time fast-forwarded by 13 minutes - will trigger refresh on inactive session');
        const tab1Navigation = page1.goto(`${config.tenantUrl}/projects`).catch((err) => {
            logger.info(`Tab 1 navigation failed (expected): ${err.message}`);
            return null;
        });

        const tab2Navigation = page2.goto(`${config.tenantUrl}/projects`).catch((err) => {
            logger.info(`Tab 2 navigation failed (expected): ${err.message}`);
            return null;
        });
        await Promise.all([tab1Navigation, tab2Navigation]);
        await page1.waitForTimeout(2000);
        const totalErrors = tokenErrors.tab1.length + tokenErrors.tab2.length;
        logger.info(`Total token errors: ${totalErrors}`);
        logger.info(`Tab 1 errors: ${tokenErrors.tab1.length}`);
        logger.info(`Tab 2 errors: ${tokenErrors.tab2.length}`);
        expect(totalErrors, 'at least one tab should have detected a 400 token error after session was deactivated').to.be.greaterThan(0);
        await page1.waitForTimeout(2000);
        const tab1OnLogin = await page1.evaluate(() => {
            return (
                document.querySelector('#username') !== null ||
                window.location.href.includes('/login') ||
                window.location.href.includes('/auth/realms')
            );
        });
        const tab2OnLogin = await page2.evaluate(() => {
            return (
                document.querySelector('#username') !== null ||
                window.location.href.includes('/login') ||
                window.location.href.includes('/auth/realms')
            );
        });
        logger.info(`Tab 1 redirected to login: ${tab1OnLogin}`);
        logger.info(`Tab 2 redirected to login: ${tab2OnLogin}`);
        expect(tab1OnLogin, 'tab 1 should be redirected to login after inactive session error').to.be.true;
        expect(tab2OnLogin, 'tab 2 should be redirected to login after inactive session error').to.be.true;
        const tab1TokensCleared = await page1.evaluate(() => {
            return {
                accessToken: localStorage.getItem('access-token'),
                refreshToken: localStorage.getItem('refresh-token'),
                idToken: localStorage.getItem('id-token')
            };
        });
        const tab2TokensCleared = await page2.evaluate(() => {
            return {
                accessToken: localStorage.getItem('access-token'),
                refreshToken: localStorage.getItem('refresh-token'),
                idToken: localStorage.getItem('id-token')
            };
        });
        logger.info(`Tab 1 tokens: ${JSON.stringify(tab1TokensCleared)}`);
        logger.info(`Tab 2 tokens: ${JSON.stringify(tab2TokensCleared)}`);
        expect(tab1TokensCleared.accessToken, 'tab 1 access-token should be cleared from localStorage after inactive session').to.be.null;
        expect(tab1TokensCleared.refreshToken, 'tab 1 refresh-token should be cleared from localStorage after inactive session').to.be.null;
        expect(tab2TokensCleared.accessToken, 'tab 2 access-token should be cleared from localStorage after inactive session').to.be.null;
        expect(tab2TokensCleared.refreshToken, 'tab 2 refresh-token should be cleared from localStorage after inactive session').to.be.null;
    });

    test('should detect invalid random token and redirect to login', async function () {
        await page1.goto(config.tenantUrl);
        await page1.waitForSelector('#username', { timeout: 10000 });
        logger.info('Injecting random string as access token');
        await page1.evaluate(() => {
            localStorage.setItem('access-token', 'random-garbage-invalid-token-12345');
            localStorage.setItem('refresh-token', 'another-random-garbage-string-67890');
            localStorage.setItem('id-token', 'fake-id-token-abcdef');
        });
        await page1.goto(`${config.tenantUrl}/projects`).catch((err) => {
            logger.info(`Navigation failed (expected): ${err.message}`);
            return null;
        });
        await page1.waitForTimeout(3000);
        const onLoginPage = await page1.evaluate(() => {
            return (
                document.querySelector('#username') !== null ||
                window.location.href.includes('/login') ||
                window.location.href.includes('/auth/realms')
            );
        });
        expect(onLoginPage, 'injecting an invalid token and navigating should redirect to the login page').to.be.true;
    });

});
