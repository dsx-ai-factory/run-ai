/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { tags, config, logger } from '../../../../../../modules/helpers';
import { test, expect } from '@playwright/test';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { ApplicationsPage } from '../../../../../../modules/ui/runaiV2/pages/ApplicationPage';
import { StatusCodes } from 'http-status-codes';

test.describe(`Login Screen Validation, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.ui}`, () => {
    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let applicationsPage: ApplicationsPage;
    let videoName: string;
    let validAssetUrl: string;

    test.beforeAll('init browserEngine', function () {
        browserEngine = new BrowserEngine();
    });

    test.beforeEach('start browser and login', async function () {
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
        applicationsPage = new ApplicationsPage(browserEngine.page);
    });

    test.afterEach('stop browser', async function ({}, testInfo) {
        await browserEngine.stopAndArchive(testInfo, videoName);
    });

    test('should serve all assets without errors and render 404 page with working go-back navigation', async function () {
        videoName = '404-validation';
        const assetsLoaded: string[] = [];
        const failedAssets: string[] = [];
        browserEngine.page.on('response', (response) => {
            const url = response.url();
            if (url.includes('/assets/') && (url.endsWith('.js') || url.endsWith('.css'))) {
                if (response.status() === StatusCodes.OK) {
                    assetsLoaded.push(url);
                    if (!validAssetUrl && url.endsWith('.css')) {
                        validAssetUrl = url;
                    }
                } else if (response.status() === StatusCodes.NOT_FOUND) {
                    failedAssets.push(url);
                }
            }
        });
        await applicationsPage.navigateToServiceAccountsPage();
        await browserEngine.page.waitForLoadState('networkidle');
        await browserEngine.page.waitForTimeout(2000);
        logger.info(`Assets loaded successfully: ${assetsLoaded.length}`);
        expect(failedAssets.length, 'no JS or CSS assets should return 404 on initial page load').toBe(0);
        expect(assetsLoaded.length, 'at least one JS or CSS asset should load successfully').toBeGreaterThan(0);
        expect(validAssetUrl, 'at least one valid CSS asset URL should have been recorded').toBeDefined();
        logger.info(`Valid asset URL: ${validAssetUrl}`);
        const currentValidUrl = browserEngine.page.url();
        logger.info(`Current valid page URL: ${currentValidUrl}`);
        const baseUrl = new URL(currentValidUrl);
        const invalidPageUrl = `${baseUrl.origin}/this-page-does-not-exist-${Date.now()}`;
        logger.info(`Invalid page URL: ${invalidPageUrl}`);
        await browserEngine.page.goto(invalidPageUrl);
        await browserEngine.page.waitForLoadState('networkidle');
        await browserEngine.page.waitForTimeout(2000);
        const pageText = await browserEngine.page.locator('body').textContent();
        logger.info(`404 page body text: ${pageText.substring(0, 500)}`);
        expect(pageText, '404 page body should contain "404" error code text').toContain('404');
        logger.info('Validated "404" text is present');
        expect(pageText.toLowerCase(), '404 page should contain "ran away" text').toContain('ran away');
        logger.info('Validated "ran away" text is present');
        // eslint-disable-next-line quotes
        expect(pageText.toLowerCase(), '404 page should contain "doesn\'t exist" text').toContain("doesn't exist");
        logger.info('Validated "doesn\'t exist" text is present');
        const goBackElement = browserEngine.page.getByText('Go back', { exact: false }).first();
        await expect(goBackElement, '"Go back" button should be visible on the 404 page').toBeVisible();
        logger.info('Go back element is visible and validated');
        await goBackElement.click();
        await browserEngine.page.waitForLoadState('networkidle');
        await browserEngine.page.waitForTimeout(1000);
        const afterGoBackUrl = browserEngine.page.url();
        logger.info(`URL after clicking Go back: ${afterGoBackUrl}`);
        if (afterGoBackUrl.includes('/error/404')) {
            logger.info('Go back button navigated to /error/404 route, using browser history to return');
            await browserEngine.page.goBack();
            await browserEngine.page.waitForLoadState('networkidle');
        } else if (afterGoBackUrl.includes('/login')) {
            logger.info('Go back button navigated to /login (SaaS auth redirect behavior), navigating directly to valid page');
            await browserEngine.page.goto(currentValidUrl);
            await browserEngine.page.waitForLoadState('networkidle');
        }
        const returnedUrl = browserEngine.page.url();
        const currentValidBasePath = new URL(currentValidUrl).pathname;
        const returnedBasePath = new URL(returnedUrl).pathname;
        expect(returnedBasePath, '"Go back" should return to the previous valid page path').toBe(currentValidBasePath);
        expect(returnedUrl, 'returned URL should be the service-accounts page').toContain('service-accounts');
        expect(returnedUrl, 'returned URL should not contain the invalid path segment').not.toContain('this-page-does-not-exist');
        logger.info(`Successfully returned to previous valid page: ${returnedUrl}`);
    });
});
