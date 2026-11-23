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
import { config, logger, settings, tags, tenantUtils, authUtils } from '../../../../../modules/helpers';
import { Keycloak } from '../../../../../modules/helpers/utils/Keycloak';
import { StatusCodes } from 'http-status-codes';
import { TempBackendSettingsApi } from '../../../../../modules/api/apis/tempBackendSettingsApi';
import { HttpBearerAuth } from '../../../../../modules/api/apis/tempBackendSettingsApi/generated/model/models';
import { EFeatureFlagsKeys } from '../../../../../modules/helpers/models/featureFlags';

test.describe(
    `Settings in Multi-Tenancy: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.tenant}`,
    () => {
        const settingKey = EFeatureFlagsKeys.ALLOW_GPT2_MODEL;
        let tenantName: string;
        let cspToken: TokenResponse;
        let tenant: TenantCreationResponse;
        let tenantUrl: string;
        let childTenantUrl: string;
        let cspSettingsApi: TempBackendSettingsApi;
        let cspConsoleSettingsApi: TempBackendSettingsApi;

        test.beforeAll(async () => {
            settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient(
                'console',
                'default'
            );
            tenantName = `e2e${Date.now()}`;
            const baseUrl = config.tenantUrl.replace('https://', '');
            tenantUrl = `https://console.${baseUrl}`;

            cspToken = await authUtils.authenticateSuperAdminAppAndGetToken(
                'default',
                settings.tenantClientSecret,
                tenantUrl
            );

            logger.info(`Creating tenant: ${tenantName}`);
            tenant = await tenantUtils.createTenant(
                cspToken.accessToken,
                tenantUrl,
                tenantName,
                config.controlPlaneAdminUsername,
                config.controlPlaneAdminPassword,
                TenantCreationRequest.ContractTypeEnum.Normal,
                null,
                null
            );
            logger.info(
                `Tenant created: id:${tenant.tenant.id} name:${tenant.tenant.name} status:${tenant.tenant.status}`
            );

            childTenantUrl = tenant.additionalData['tenant_domain'];

            // both instances use the console URL — CSP operations are scoped to console tenant
            cspSettingsApi = new TempBackendSettingsApi(tenantUrl);
            cspConsoleSettingsApi = new TempBackendSettingsApi(tenantUrl);
        });

        test.afterAll(async () => {
            await tenantUtils.deleteTenant(cspToken.accessToken, tenantUrl, tenant.tenant.id, true);
            logger.info(`Tenant deleted: ${tenant.tenant.name}`);
        });

        test.beforeEach(async () => {
            settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient(
                'console',
                'default'
            );
            const bearerAuth = new HttpBearerAuth();
            bearerAuth.accessToken = cspToken.accessToken;
            cspSettingsApi.tenantApi.setDefaultAuthentication(bearerAuth);
            cspConsoleSettingsApi.tenantApi.setDefaultAuthentication(bearerAuth);
        });

        test('should get console tenant own settings without passing tenant id', async () => {
            const response = await cspSettingsApi.tenantApi.getTenantSettings();
            expect(response.response.statusCode, 'getting console tenant settings should return 200').to.equal(StatusCodes.OK);
            expect(response.body, 'settings response body should be an array').to.be.an('array');
            expect(response.body.length, 'settings array should have at least one entry').to.be.greaterThan(0);
            logger.info(`Got ${response.body.length} settings for console tenant`);
        });

        test('should get console tenant own setting by key without passing tenant id', async () => {
            const response = await cspSettingsApi.tenantApi.getTenantSettingsByKey(settingKey);
            expect(response.response.statusCode, 'getting console tenant setting by key should return 200').to.equal(StatusCodes.OK);
            expect(response.body, 'setting value should be a boolean').to.be.a('boolean');
            logger.info(`Got setting ${settingKey}: ${response.body}`);
        });

        test('should update console tenant own setting without passing tenant id', async () => {
            const getResponse = await cspSettingsApi.tenantApi.getTenantSettingsByKey(settingKey);
            expect(getResponse.response.statusCode, 'getting setting before update should return 200').to.equal(StatusCodes.OK);
            const originalValue = getResponse.body;

            try {
                const updateResponse = await cspSettingsApi.tenantApi.updateTenantSetting(undefined, {
                    key: settingKey,
                    value: !originalValue
                });
                expect(updateResponse.response.statusCode, 'updating console tenant setting should return 200').to.equal(StatusCodes.OK);

                const verifyResponse = await cspSettingsApi.tenantApi.getTenantSettingsByKey(settingKey);
                expect(verifyResponse.response.statusCode, 'verifying updated setting should return 200').to.equal(StatusCodes.OK);
                expect(verifyResponse.body, 'setting value should reflect the toggled value after update').to.equal(!originalValue);
                logger.info(`Updated ${settingKey} from ${originalValue} to ${!originalValue}`);
            } finally {
                const restoreResponse = await cspSettingsApi.tenantApi.updateTenantSetting(undefined, {
                    key: settingKey,
                    value: originalValue
                });
                expect(restoreResponse.response.statusCode, 'restoring original setting value should return 200').to.equal(StatusCodes.OK);
            }
        });

        test('should update child tenant setting by passing tenant id', async () => {
            const tenantId = String(tenant.tenant.id);
            const updateResponse = await cspConsoleSettingsApi.tenantApi.updateTenantSetting(tenantId, {
                key: settingKey,
                value: true
            });
            expect(updateResponse.response.statusCode, 'updating child tenant setting by tenant id should return 200').to.equal(StatusCodes.OK);
            logger.info(`Updated ${settingKey} for child tenant id:${tenantId}`);
        });

        test('should get child tenant settings by passing tenant id', async () => {
            const tenantId = String(tenant.tenant.id);
            const response = await cspConsoleSettingsApi.tenantApi.getTenantSettings(undefined, tenantId);
            expect(response.response.statusCode, 'getting child tenant settings by tenant id should return 200').to.equal(StatusCodes.OK);
            expect(response.body, 'child tenant settings body should be an array').to.be.an('array');
            expect(response.body.length, 'child tenant settings array should have at least one entry').to.be.greaterThan(0);
            logger.info(`Got ${response.body.length} settings for child tenant id:${tenantId}`);
        });

        test('should get child tenant setting by key by passing tenant id', async () => {
            const tenantId = String(tenant.tenant.id);
            const response = await cspConsoleSettingsApi.tenantApi.getTenantSettingsByKey(
                settingKey,
                tenantId
            );
            expect(response.response.statusCode, 'getting child tenant setting by key and tenant id should return 200').to.equal(StatusCodes.OK);
            expect(response.body, 'setting value for child tenant should be a boolean').to.be.a('boolean');
            logger.info(`Got setting ${settingKey} for child tenant id:${tenantId}: ${response.body}`);
        });

        test('should get fewer settings when excludeMigrated is true', async () => {
            const allSettings = await cspSettingsApi.tenantApi.getTenantSettings();
            const nonMigratedSettings = await cspSettingsApi.tenantApi.getTenantSettings(true);
            expect(allSettings.response.statusCode, 'getting all settings should return 200').to.equal(StatusCodes.OK);
            expect(nonMigratedSettings.response.statusCode, 'getting non-migrated settings should return 200').to.equal(StatusCodes.OK);
            expect(nonMigratedSettings.body.length, 'non-migrated settings count should be <= total settings count').to.be.lessThanOrEqual(allSettings.body.length);
            logger.info(
                `All settings: ${allSettings.body.length}, excludeMigrated=true: ${nonMigratedSettings.body.length}`
            );
        });

        test('should not allow non-console tenant to update another tenant settings', async () => {
            // create a second real child tenant as the target
            const targetTenantName = `e2e${Date.now()}`;
            const targetTenant = await tenantUtils.createTenant(
                cspToken.accessToken,
                tenantUrl,
                targetTenantName,
                config.controlPlaneAdminUsername,
                config.controlPlaneAdminPassword,
                TenantCreationRequest.ContractTypeEnum.Normal,
                null,
                null
            );
            logger.info(`Target tenant created: id:${targetTenant.tenant.id} name:${targetTenantName}`);
            const targetTenantUrl = targetTenant.additionalData['tenant_domain'];

            try {
                const targetTenantId = String(targetTenant.tenant.id);

                // wait for target tenant to be ready before reading its settings
                await tenantUtils.waitUntilNewTenantIsReadyAndRetrieveToken(
                    cspToken.accessToken,
                    config.controlPlaneAdminUsername,
                    config.controlPlaneAdminPassword,
                    targetTenantName,
                    targetTenantUrl
                );
                logger.info(`Target tenant is ready: ${targetTenantUrl}`);

                // get target tenant's current setting value via console API before the attempted update
                const beforeResponse = await cspConsoleSettingsApi.tenantApi.getTenantSettingsByKey(
                    settingKey,
                    targetTenantId
                );
                const settingBefore = beforeResponse.body;
                logger.info(`Target tenant ${settingKey} before: ${settingBefore}`);

                // wait for first child tenant to be ready and get its token
                logger.info(`Waiting for child tenant to be ready: ${childTenantUrl}`);
                const childTenantToken = await tenantUtils.waitUntilNewTenantIsReadyAndRetrieveToken(
                    cspToken.accessToken,
                    config.controlPlaneAdminUsername,
                    config.controlPlaneAdminPassword,
                    tenantName,
                    childTenantUrl
                );
                logger.info('Child tenant is ready, token retrieved');

                // non-console tenant tries to update the target tenant's settings — backend returns 200
                // but should not actually apply the change to the target tenant
                const nonConsoleSettingsApi = new TempBackendSettingsApi(childTenantUrl);
                const bearerAuth = new HttpBearerAuth();
                bearerAuth.accessToken = childTenantToken.accessToken;
                nonConsoleSettingsApi.tenantApi.setDefaultAuthentication(bearerAuth);

                const updateResponse = await nonConsoleSettingsApi.tenantApi.updateTenantSetting(targetTenantId, {
                    key: settingKey,
                    value: !settingBefore
                });
                expect(updateResponse.response.statusCode, 'non-console tenant update request should return 200 (but change should not be applied)').to.equal(StatusCodes.OK);
                logger.info(`Non-console tenant update returned ${updateResponse.response.statusCode}`);

                // verify target tenant's setting was NOT changed
                const afterResponse = await cspConsoleSettingsApi.tenantApi.getTenantSettingsByKey(
                    settingKey,
                    targetTenantId
                );
                expect(afterResponse.body, 'target tenant setting should not be changed by a non-console tenant update').to.equal(settingBefore);
                logger.info(`Target tenant setting unchanged: ${afterResponse.body}`);
            } finally {
                await tenantUtils.deleteTenant(cspToken.accessToken, tenantUrl, targetTenant.tenant.id, true);
                logger.info(`Target tenant deleted: ${targetTenantName}`);
            }
        });
    }
);
