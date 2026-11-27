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
import { authUtils, config, logger, settings, tags, tenantUtils } from '../../../../../modules/helpers';
import { Keycloak } from '../../../../../modules/helpers/utils/Keycloak';

test.describe(`Phoenix tests tenant SH Multi-Tenancy: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.tenant}`, () => {
    let tenantName: string;
    let token: TokenResponse;
    let tenant: TenantCreationResponse;
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
        logger.debug('Phoenix tenant test setup successful');
    });

    test.afterAll(async function () {
        const deleted = await tenantUtils.deleteTenant(token.idToken, tenantUrl, tenant.tenant.id, true);
        logger.info(`Tenant deleted: ${deleted.uid}`);
    });

    test.beforeEach(async function () {
        settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient('console', 'default');
        logger.debug('Before each Phoenix hook successful started');
    });

    test('should create tenant and validate name', function () {
        expect(tenant.tenant.name, 'created tenant name should match the requested tenant name').to.equal(tenantName);
    });

    test('should validate console tenant exist and created', async function () {
        const tenants = await tenantUtils.getTenants(token.accessToken, tenantUrl);
        expect(tenants.find((t) => t && t.name === 'console'), 'console tenant should exist in tenant list').to.not.be.undefined;
    });

    test('should get tenant list', async function () {
        const tenants = await tenantUtils.getTenants(token.accessToken, tenantUrl);
        expect(tenants, 'tenant list should not be empty').to.not.be.empty;
        expect(tenants.length, 'tenant list should contain more than 1 tenant').to.be.greaterThan(1);
    });

    test('should get tenant by id', async function () {
        const tenantById = await tenantUtils.getTenantById(token.accessToken, tenantUrl, tenant.tenant.id);
        expect(tenantById.id, 'tenant retrieved by id should have the correct id').to.equal(tenant.tenant.id);
    });
});
