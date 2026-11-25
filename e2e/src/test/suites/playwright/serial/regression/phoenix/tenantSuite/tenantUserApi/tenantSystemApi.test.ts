/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { test } from '@playwright/test';
import { tags } from '../../../../../../../../modules/helpers/config/tags';
import {
    bffUtils,
    config,
    flowRootState,
    logger,
    runaiHealthChecker,
    settings,
    tenantUtils
} from '../../../../../../../../modules/helpers';
import { isStagingEnv, saasTenantName } from '../../../../../../../../modules/helpers/config/config';
import { Tenant } from '../../../../../../../../modules/api/apis/tenantsManagerApi/generated/api';
import { AuthConfiguration } from '../../../../../../../../modules/api/apis/identityManagerApi/generated/api';

test.describe(
    `Tenant Systems API ${tags.teams.phoenix} ${tags.suites.legacySphinx}`,
    { tag: [tags.services.tenantsManager, tags.runOn.serialCI] },
    () => {
        let tenant: Tenant;
        let originalLogo: string | undefined;

        test.beforeAll('should get tenant config', async () => {
            tenant = await tenantUtils.getTenant(flowRootState.users.system_administrator.accessToken);
            const body = await bffUtils.getLogo().catch(() => undefined);
            originalLogo = body?.value?.logo;
        });

        test.afterAll('restore original branding logo', async () => {
            await bffUtils.updateLogo(originalLogo ?? '').catch(() => undefined);
        });

        test('should get tenant settings', async () => {
            const settings = await tenantUtils.getTenantSettings(
                flowRootState.users.system_administrator.accessToken
            );
            expect(settings, 'tenant settings should exist').to.exist;
        });

        test('should get tenant by name', async () => {
            const response = await tenantUtils.getTenant(
                flowRootState.users.system_administrator.accessToken,
                tenant.name
            );
            expect(response.contractType, 'tenant retrieved by name should have a contractType').to.exist;
        });

        // -----------------------------------------------------------------------
        // TM logo proxy — original API must continue to work
        // -----------------------------------------------------------------------

        test('should get tenant configuration', async () => {
            const response = await tenantUtils.getTenantConfiguration(
                flowRootState.users.system_administrator.accessToken
            );
            logger.debug(
                ` TENANT CONFIGURATION: response.tenantName: ${response.tenantName}, response.authHost: ${response.authHost}, response.isMultiTenant: ${response.isMultiTenant}, response.authClientID: ${response.authClientID}`
            );
            let expectedConfig: AuthConfiguration;
            if (await runaiHealthChecker.isSelfHostedCluster()) {
                expectedConfig = {
                    tenantName: 'runai',
                    authHost: config.tenantUrl,
                    isMultiTenant: false,
                    authClientID: 'runai-admin-ui',
                    redirectUrl: response.redirectUrl
                };
            } else if (saasTenantName && !(await runaiHealthChecker.isSelfHostedCluster())) {
                expectedConfig = {
                    tenantName: saasTenantName,
                    authHost: config.appUrl,
                    isMultiTenant: true,
                    authClientID: 'runai-admin-ui',
                    redirectUrl: response.redirectUrl
                };
            } else {
                logger.debug('Tenant configuration is not self hosted and not saas tenant');
                expect(false).to.be.true;
            }
            logger.debug(
                ` TENANT CONFIGURATION: expectedConfig.tenantName: ${expectedConfig.tenantName}, expectedConfig.authHost: ${expectedConfig.authHost}, expectedConfig.isMultiTenant: ${expectedConfig.isMultiTenant}, expectedConfig.authClientID: ${expectedConfig.authClientID}`
            );
            logger.debug(`TENANT CONFIGURATION: response: ${JSON.stringify(response)}`);
            logger.debug(`TENANT CONFIGURATION: expectedConfig: ${JSON.stringify(expectedConfig)}`);
            logger.debug(`CONTROL PLANE URL: ${config.tenantUrl}`);
            logger.debug(`APP URL: ${config.appUrl}`);
            logger.debug(`IS SELF HOSTED: ${settings.isSelfHosted}`);
            logger.debug(`IS STAGING ENV: ${isStagingEnv}`);
            logger.debug(`SAAS TENANT NAME: ${saasTenantName}`);
            expect(response, 'tenant configuration response should exist').to.exist;
            expect(response.authClientID, 'auth client ID should be the standard runai-admin-ui').to.equal('runai-admin-ui');
            expect(response.tenantName, 'tenant name should match expected configuration').to.equal(expectedConfig.tenantName);
            expect(response.authHost, 'auth host should match expected configuration').to.equal(expectedConfig.authHost);
            expect(response.isMultiTenant, 'isMultiTenant flag should match expected configuration').to.equal(expectedConfig.isMultiTenant);
        });
    }
);
