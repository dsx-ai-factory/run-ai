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
    tenantUtils
} from '../../../../../modules/helpers';
import { StatusCodes } from 'http-status-codes';
import { Keycloak } from '../../../../../modules/helpers/utils/Keycloak';

test.describe(`Tenant creation status codes SH Multi-Tenancy: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.tenant}`, () => {
    let tenantName: string;
    let token: TokenResponse;
    let tenant: TenantCreationResponse;

    test.beforeEach(async function () {
        settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient('console', 'default');
        logger.debug('Before each Phoenix hook successful started');
    });

    test.afterEach(async function () {
        try {
            if (tenant && tenant.tenant && tenant.tenant.id) {
                const baseUrl = config.tenantUrl.replace('https://', '');
                const tenantUrl = `https://console.${baseUrl}`;

                const deleted = await tenantUtils.deleteTenant(
                    token.accessToken,
                    tenantUrl,
                    tenant.tenant.id,
                    true
                );
                logger.info(`Tenant deleted: ${deleted.uid}`);
            }
        } catch (error) {
            logger.debug(`tenant cant be deleted in after each because: ${error.message}`);
        }
    });

    test('should not create tenant without authentication', async function () {
        const baseUrl = config.tenantUrl.replace('https://', '');
        const tenantUrl = `https://console.${baseUrl}`;

        tenantName = `e2e${Date.now()}`;
        try {
            await tenantUtils.createTenant(
                '',
                tenantUrl,
                tenantName,
                config.controlPlaneAdminUsername,
                config.controlPlaneAdminPassword,
                TenantCreationRequest.ContractTypeEnum.Normal,
                null,
                null
            );
            throw new Error('Tenant creation without authentication should have failed but succeeded');
        } catch (error) {
            logger.info(`Expected error occurred: ${error.message}`);
            expect(error.statusCode, 'creating a tenant without authentication should return 401 Unauthorized').to.eql(StatusCodes.UNAUTHORIZED);
        }
    });

    test('should return 400 when deleting a non-existent tenant value out of range', async function () {
        const baseUrl = config.tenantUrl.replace('https://', '');
        const tenantUrl = `https://console.${baseUrl}`;

        token = await authUtils.authenticateSuperAdminAppAndGetToken(
            'default',
            settings.tenantClientSecret,
            tenantUrl
        );

        const nonExistentTenantId = 9999999999999;
        try {
            await tenantUtils.deleteTenant(token.accessToken, tenantUrl, nonExistentTenantId, true);
        } catch (error) {
            logger.info(`Expected error occurred: ${error.message}`);
            expect(error.statusCode, 'deleting a tenant with an out-of-range id should return 400 Bad Request').to.eql(StatusCodes.BAD_REQUEST);
            expect(error.message, 'error message should indicate value out of range').to.contain('value out of range');
        }
    });

    test('should return 401 when deleting a tenant without authentication', async function () {
        const baseUrl = config.tenantUrl.replace('https://', '');
        const tenantUrl = `https://console.${baseUrl}`;

        tenantName = `e2e${Date.now()}`;
        token = await authUtils.authenticateSuperAdminAppAndGetToken(
            'default',
            settings.tenantClientSecret,
            tenantUrl
        );

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

        try {
            await tenantUtils.deleteTenant('', tenantUrl, tenant.tenant.id, true);
        } catch (error) {
            logger.info(`Expected error occurred: ${error.message}`);
            expect(error.statusCode, 'deleting a tenant without authentication should return 401 Unauthorized').to.eql(StatusCodes.UNAUTHORIZED);
        }
    });
});
