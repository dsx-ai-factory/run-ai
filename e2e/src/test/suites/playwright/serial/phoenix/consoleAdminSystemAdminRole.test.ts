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
    usersUtils
} from '../../../../../modules/helpers';
import { Keycloak } from '../../../../../modules/helpers/utils/Keycloak';
import { StatusCodes } from 'http-status-codes';

test.describe(`Console super admin can create a user in tenant, local admin privileges ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, () => {
    let superAdminToken: TokenResponse;
    let consoleAdminToken: TokenResponse;
    let tenant: TenantCreationResponse;
    let tenantUrl: string;
    let tenantSpecificUrl: string;

    test.beforeAll(async function () {
        settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient('console', 'default');
        const baseUrl = config.tenantUrl.replace('https://', '');
        tenantUrl = `https://console.${baseUrl}`;

        superAdminToken = await authUtils.authenticateSuperAdminAppAndGetToken(
            'default',
            settings.tenantClientSecret,
            tenantUrl
        );

        const tenantName = `e2e-sysadmin-${Date.now()}`;
        tenant = await tenantUtils.createTenant(
            superAdminToken.accessToken,
            tenantUrl,
            tenantName,
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword,
            TenantCreationRequest.ContractTypeEnum.Normal,
            null,
            null
        );
        logger.info(`Tenant created: ${tenant.tenant.name} (id: ${tenant.tenant.id})`);

        tenantSpecificUrl = tenant.additionalData['tenant_domain'];

        consoleAdminToken = await tenantUtils.waitUntilNewTenantIsReadyAndRetrieveToken(
            superAdminToken.accessToken,
            config.controlPlaneAdminUsername,
            config.controlPlaneAdminPassword,
            tenant.tenant.name,
            tenantSpecificUrl
        );
        logger.debug('Test setup complete');
    });

    test.afterAll(async function () {
        if (superAdminToken?.accessToken && tenant?.tenant?.id) {
            await tenantUtils.deleteTenant(superAdminToken.accessToken, tenantUrl, tenant.tenant.id, true);
            logger.info(`Tenant deleted: ${tenant.tenant.id}`);
        }
    });

    test('console super admin should create a local user in tenant', async function () {
        const email = `e2e-super-admin-${Date.now()}@example.com`;

        const result = await usersUtils.createUserInSpecificTenant(
            consoleAdminToken.accessToken,
            tenantSpecificUrl,
            { email, resetPassword: false }
        );

        expect(result.statusCode, 'console super admin creating a user in a tenant should return 201 Created').to.equal(StatusCodes.CREATED);
        logger.info(`User created: ${email}`);
    });
});
