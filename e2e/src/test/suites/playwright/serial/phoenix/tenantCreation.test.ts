/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { test } from '@playwright/test';
import { TokenResponse } from '../../../../../modules/api/apis/identityManagerApi/generated/api';
import {
    TenantCreationRequest,
    TenantCreationResponse
} from '../../../../../modules/api/apis/tenantsManagerApi/generated/api';
import {
    authUtils,
    config,
    logger,
    settings,
    tags,
    tenantUtils,
    Utils
} from '../../../../../modules/helpers';
import { Browser, chromium } from 'playwright';
import { Keycloak } from '../../../../../modules/helpers/utils/Keycloak';

test.describe(`Phoenix tests creation Tenant SH Multi-Tenancy: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.tenant}`, () => {
    let tenantName: string;
    let token: TokenResponse;
    let tenant: TenantCreationResponse;
    let browser: Browser | null = null;
    let tenantUrl: string;

    test.beforeAll(async function () {
        settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient('console', 'default');
        tenantName = `e2e${Date.now()}`;
        const baseUrl = config.tenantUrl.replace('https://', '');
        tenantUrl = `https://console.${baseUrl}`;

        token = await authUtils.authenticateSuperAdminAppAndGetToken(
            'default',
            settings.tenantClientSecret,
            tenantUrl
        );
        logger.info(`Creating tenant: ${tenantName} with token: ${token.accessToken}`);
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
        logger.info(
            `Tenant created: status:${tenant.tenant.status} and name:${tenant.tenant.name} and id:${tenant.tenant.id}`
        );
        logger.debug('Phoenix tenant creation test setup successful');
    });

    test.afterAll(async function () {
        await tenantUtils.deleteTenant(token.accessToken, tenantUrl, tenant.tenant.id, true);
        if (browser) {
            await browser.close();
            browser = null;
        }
    });

    test.beforeEach(async function () {
        logger.debug('Before each Phoenix hook Started');
        settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient('console', 'default');
        logger.debug('Before each Phoenix hook successful started');
    });

    test('should completely create tenant and validate status is ready', async function () {
        await Utils.delay(120000);
        const getTenantStatus = await tenantUtils.getTenantById(
            token.accessToken,
            tenantUrl,
            tenant.tenant.id
        );
        expect(getTenantStatus.status, 'tenant should reach Ready status after creation').to.equal('Ready');
        logger.info(`Tenant created: status:${getTenantStatus.status} and name:${getTenantStatus.name}`);
    });

    test('should auth new created tenant', async function () {
        browser = await chromium.launch({ headless: settings.headlessBrowser });
        const context = await browser.newContext({
            ignoreHTTPSErrors: true,
            recordVideo: {
                dir: settings.browserVideosPath,
                size: { width: 1200, height: 700 }
            }
        });
        const page = await context.newPage();
        logger.debug(`tenant: ${tenant.additionalData['tenant_domain']}`);
        await page.goto(`${tenant.additionalData['tenant_domain']}`);
        await page.fill(
            'input[name="email"], input[id="username"], input[id=inputUsername]',
            config.controlPlaneAdminUsername
        );
        await page.fill('input[name="password"], input[id="password"]', config.controlPlaneAdminPassword);
        await page.click(
            ':is([aid="loginSubmitButton"]:not([id="loginWithSSOButton"])), [type="submit"]:nth-child(1), [type="submit"]'
        );
        await page.waitForLoadState('networkidle');
        expect(page.url()).to.be.equal(`${tenant.additionalData['tenant_domain']}/clusters/new`);
        await page.close();
    });
});
