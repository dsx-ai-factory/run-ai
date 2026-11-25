/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { test } from '@playwright/test';
import { Browser, chromium } from 'playwright';
import { TokenResponse } from '../../../../../modules/api/apis/identityManagerApi/generated/api';
import {
    TenantCreationRequest,
    TenantCreationResponse
} from '../../../../../modules/api/apis/tenantsManagerApi/generated/api';
import { authUtils, config, logger, settings, tags, tenantUtils, Utils } from '../../../../../modules/helpers';
import { Keycloak } from '../../../../../modules/helpers/utils/Keycloak';

test.describe(`Phoenix tests tenant alias SH Multi-Tenancy: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.resource.tenant}`, () => {

    let token: TokenResponse;
    let tenant: TenantCreationResponse;
    let tenantUrl: string;
    let baseUrl: string;
    let browser: Browser | null = null;    let clearedAlias: string;

    test.beforeAll(async function () {
        settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient('console', 'default');
        baseUrl = config.tenantUrl.replace('https://', '');
        tenantUrl = `https://console.${baseUrl}`;
        token = await authUtils.authenticateSuperAdminAppAndGetToken(
            'default',
            settings.tenantClientSecret,
            tenantUrl
        );
        const tenantName = `e2e${Date.now()}`;
        tenant = await tenantUtils.createTenant(
            token.accessToken,
            tenantUrl,
            tenantName,
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword,
            TenantCreationRequest.ContractTypeEnum.Normal,
            null,
            null
        );
        logger.info(`Tenant created: name=${tenant.tenant.name} id=${tenant.tenant.id}`);
        await Utils.delay(120000);
        const status = await tenantUtils.getTenantById(token.accessToken, tenantUrl, tenant.tenant.id);
        expect(status.status, 'tenant should reach Ready status before running alias tests').to.equal('Ready');
        logger.info(`Tenant ready: ${tenant.tenant.name}`);

        const initialAlias = `e2ealias${Date.now()}`;
        tenant.tenant = await tenantUtils.updateTenantAlias(token.accessToken, tenantUrl, tenant.tenant.id, initialAlias);
        logger.info(`Initial alias set: ${initialAlias} for tenant: name=${tenant.tenant.name} id=${tenant.tenant.id}`);
    });

    test.afterAll(async function () {
        await tenantUtils.deleteTenant(token.idToken, tenantUrl, tenant.tenant.id, true);
        logger.info(`Tenant deleted: ${tenant.tenant.name}`);
        if (browser) {
            await browser.close();
            browser = null;
        }
    });

    test.beforeEach(async function () {
        settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient('console', 'default');
    });

    test('tenant alias is set and returned in tenant response', function () {
        expect(tenant.tenant.alias, 'tenant alias should be a non-empty string after being set').to.be.a('string').and.not.be.empty;
        logger.info(`Alias verified: ${tenant.tenant.alias} for tenant: name=${tenant.tenant.name} id=${tenant.tenant.id}`);
    });

    test('tenant UI is accessible via alias URL', async function () {
        browser = await chromium.launch({ headless: settings.headlessBrowser });
        const context = await browser.newContext({
            ignoreHTTPSErrors: true,
            recordVideo: { dir: settings.browserVideosPath, size: { width: 1200, height: 700 } }
        });
        const page = await context.newPage();
        const aliasUrl = `https://${tenant.tenant.alias}.${baseUrl}`;

        await page.goto(aliasUrl);
        await page.fill('input[name="email"], input[id="username"], input[id=inputUsername]', config.controlPlaneAdminUsername);
        await page.fill('input[name="password"], input[id="password"]', config.controlPlaneAdminPassword);
        await page.click(':is([aid="loginSubmitButton"]:not([id="loginWithSSOButton"])), [type="submit"]:nth-child(1), [type="submit"]');
        await page.waitForURL('**/clusters/new', { timeout: 30000 });

        expect(page.url(), 'alias URL should redirect to the clusters page after login').to.include('/clusters/new');
        await page.close();
        await browser.close();
        browser = null;
    });

    test('tenant alias is updated and new alias is returned in tenant response', async function () {
        const newAlias = `e2ealias${Date.now()}`;

        const updated = await tenantUtils.updateTenantAlias(token.accessToken, tenantUrl, tenant.tenant.id, newAlias);

        expect(updated.alias, 'tenant alias should be updated to the new alias value').to.equal(newAlias);

        tenant.tenant = updated;
        logger.info(`Alias updated to: ${newAlias} for tenant: name=${tenant.tenant.name} id=${tenant.tenant.id}`);
    });

    test('tenant UI is accessible via tenant name URL after alias update', async function () {
        browser = await chromium.launch({ headless: settings.headlessBrowser });
        const context = await browser.newContext({
            ignoreHTTPSErrors: true,
            recordVideo: { dir: settings.browserVideosPath, size: { width: 1200, height: 700 } }
        });
        const page = await context.newPage();
        const nameUrl = `${tenant.additionalData['tenant_domain']}`;

        await page.goto(nameUrl);
        await page.fill('input[name="email"], input[id="username"], input[id=inputUsername]', config.controlPlaneAdminUsername);
        await page.fill('input[name="password"], input[id="password"]', config.controlPlaneAdminPassword);
        await page.click(':is([aid="loginSubmitButton"]:not([id="loginWithSSOButton"])), [type="submit"]:nth-child(1), [type="submit"]');
        await page.waitForURL(`${nameUrl}/clusters/new`, { timeout: 30000 });

        expect(page.url()).to.equal(`${nameUrl}/clusters/new`);
        await page.close();
        await browser.close();
        browser = null;
    });

    test('tenant UI is accessible via new alias URL after alias update', async function () {
        browser = await chromium.launch({ headless: settings.headlessBrowser });
        const context = await browser.newContext({
            ignoreHTTPSErrors: true,
            recordVideo: { dir: settings.browserVideosPath, size: { width: 1200, height: 700 } }
        });
        const page = await context.newPage();
        const aliasUrl = `https://${tenant.tenant.alias}.${baseUrl}`;

        await page.goto(aliasUrl);
        await page.fill('input[name="email"], input[id="username"], input[id=inputUsername]', config.controlPlaneAdminUsername);
        await page.fill('input[name="password"], input[id="password"]', config.controlPlaneAdminPassword);
        await page.click(':is([aid="loginSubmitButton"]:not([id="loginWithSSOButton"])), [type="submit"]:nth-child(1), [type="submit"]');
        await page.waitForURL('**/clusters/new', { timeout: 30000 });

        expect(page.url(), 'new alias URL should redirect to the clusters page after alias update').to.include('/clusters/new');
        await page.close();
        await browser.close();
        browser = null;
    });

    test('tenant alias is cleared and alias field is absent in tenant response', async function () {
        clearedAlias = tenant.tenant.alias;

        const updated = await tenantUtils.updateTenantAlias(token.accessToken, tenantUrl, tenant.tenant.id, null);

        expect(updated.alias, 'tenant alias should be undefined after being cleared').to.be.undefined;

        tenant.tenant = updated;
        logger.info(`Alias cleared (was: ${clearedAlias}) for tenant: name=${tenant.tenant.name} id=${tenant.tenant.id}`);
    });

    test('tenant UI is still accessible via tenant name URL after alias is cleared', async function () {
        browser = await chromium.launch({ headless: settings.headlessBrowser });
        const context = await browser.newContext({
            ignoreHTTPSErrors: true,
            recordVideo: { dir: settings.browserVideosPath, size: { width: 1200, height: 700 } }
        });
        const page = await context.newPage();
        const nameUrl = `${tenant.additionalData['tenant_domain']}`;

        await page.goto(nameUrl);
        await page.fill('input[name="email"], input[id="username"], input[id=inputUsername]', config.controlPlaneAdminUsername);
        await page.fill('input[name="password"], input[id="password"]', config.controlPlaneAdminPassword);
        await page.click(':is([aid="loginSubmitButton"]:not([id="loginWithSSOButton"])), [type="submit"]:nth-child(1), [type="submit"]');
        await page.waitForURL(`${nameUrl}/clusters/new`, { timeout: 30000 });

        expect(page.url()).to.equal(`${nameUrl}/clusters/new`);
        await page.close();
        await browser.close();
        browser = null;
    });

    test('token API still returns valid token via tenant name URL after alias is cleared', async function () {
        const nameUrl = `${tenant.additionalData['tenant_domain']}`;
        const tokenRes = await authUtils.authenticateTokenApi(
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword,
            nameUrl,
            token.accessToken
        );

        expect(tokenRes.response.statusCode, 'token API via tenant name URL should return 200 after alias is cleared').to.equal(200);
        expect(tokenRes.body.accessToken, 'token API response should include an access token').to.not.be.undefined;
    });

    test('token API returns 404 when called with cleared alias URL', async function () {
        const oldAliasUrl = `https://${clearedAlias}.${baseUrl}`;
        try {
            await authUtils.authenticateTokenApi(
                config.controlPlaneAdminUsername,
                config.controlPlaneAdminPassword,
                oldAliasUrl,
                token.accessToken
            );
            expect.fail('Expected token request to fail for cleared alias URL');
        } catch (error: any) {
            expect(error.statusCode, 'token API with cleared alias URL should return 404').to.equal(404);
        }
    });
});
