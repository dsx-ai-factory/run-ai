/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authorizationUtils,
    authUtils,
    clusterService,
    config,
    flowRootState,
    idpUtils,
    logger,
    settings,
    tags,
    usersManager,
    usersUtils
} from '../../../../../../../modules/helpers';
import {
    idpNameOidc,
    keycloakClientName,
    realmName
} from '../../../../../../../modules/helpers/config/const';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { UsersFilterSortFields } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

shouldRunTests(settings.isSelfHosted && !isOCP())(
    function () {
        test.describe(`Authenticate External Token Exchange, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.authentication}`, function () {
            test.beforeAll(async function () {
                // Cleanup existing resources from failed previous runs
                try {
                    logger.debug('Cleaning up existing IDP and realm before test suite');
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameOidc
                    );
                } catch (error) {
                    // 404 is expected if IDP doesn't exist
                    if (!error.message?.includes('404')) {
                        logger.debug(`IDP cleanup: ${error.message}`);
                    }
                }

                try {
                    await Keycloak.deleteRealm(realmName);
                    // Wait for Keycloak to finish deletion
                    await new Promise((resolve) => setTimeout(resolve, 1000));
                } catch (error) {
                    // Expected if realm doesn't exist
                    logger.debug(`Realm cleanup: ${error.message}`);
                }

                // Now proceed with fresh setup
                await Keycloak.setupSSORealmOIDC(realmName, keycloakClientName);
            });
            test.afterAll(async function () {
                try {
                    await Keycloak.deleteUserFromRealm('runai', 'exchangetokenuser');
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameOidc
                    );
                    await Keycloak.deleteRealm(realmName);
                } catch (error) {
                    logger.error(`error: ${error}`);
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });
            test.beforeEach(async () => {
                await Keycloak.createUser(realmName, 'exchangetokenuser', config.controlPlaneAdminPassword);
            });
            test.afterEach('stop browser', async () => {
                try {
                    const users = await usersUtils.getUsersFilteredBy(
                        flowRootState.users.system_administrator.accessToken,
                        ['username=@exchangetokenuser'],
                        UsersFilterSortFields.Type
                    );
                    for (const user of users) {
                        await usersManager.deleteLocalUser(
                            flowRootState.users.system_administrator.accessToken,
                            user.id
                        );
                    }
                    await Keycloak.deleteUserFromRealm('runai', 'service-account-service-account-client');
                } catch (error) {
                    logger.error('Failed to delete user from realm', error);
                }
            });
            test('should authenticate external token exchange and get clusters list with access external token for user and service account client', async function () {
                await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    'exchangetokenuser@example.com',
                    flowRootState.users.system_administrator.role.id
                );
                const client = await Keycloak.createClient(
                    realmName,
                    'exchangetokenclient',
                    'openid-connect'
                );
                const clientSecret = await Keycloak.getClientSecret(realmName, client.id);
                const userTokenFromExternalRealm = await Keycloak.getTokenDirect(
                    'exchangetokenuser',
                    config.controlPlaneAdminPassword,
                    realmName,
                    'exchangetokenclient',
                    clientSecret,
                    config.tenantUrl
                );
                const externalTokenExchange = await authUtils.authenticateExternalTokenExchange(
                    userTokenFromExternalRealm.access_token
                );
                expect(externalTokenExchange.refreshToken, 'external token exchange should return a non-empty refresh token').to.be.not.empty;
                expect(externalTokenExchange.accessToken, 'external token exchange should return a non-empty access token').to.be.not.empty;
                const clusters = await clusterService.getClusters(externalTokenExchange.accessToken);
                expect(clusters, 'clusters response should be an array').to.be.an('array');
                expect(clusters.length, 'clusters list should contain at least one cluster').to.be.greaterThan(0);
                const cluster = clusters[0];
                expect(cluster.name, 'first cluster should have a non-empty name').to.be.not.empty;

                // Test service account client
                const serviceAccountClient = await Keycloak.createClient(
                    realmName,
                    'service-account-client',
                    'openid-connect'
                );
                const serviceAccountClientSecret = await Keycloak.getClientSecret(
                    realmName,
                    serviceAccountClient.id
                );
                const serviceAccountToken = await Keycloak.getApplicationDirectToken(
                    realmName,
                    'service-account-client',
                    serviceAccountClientSecret
                );
                const externalTokenExchangeServiceAccount = await authUtils.authenticateExternalTokenExchange(
                    serviceAccountToken.access_token
                );
                expect(externalTokenExchangeServiceAccount.refreshToken, 'service account external token exchange should return a non-empty refresh token').to.be.not.empty;
                expect(externalTokenExchangeServiceAccount.accessToken, 'service account external token exchange should return a non-empty access token').to.be.not.empty;
                try {
                    await clusterService.getClusters(externalTokenExchangeServiceAccount.accessToken);
                } catch (error) {
                    logger.debug('test success the error is according the client have no permissions');
                    expect(error.statusCode, 'service account without permissions should receive 403 when accessing clusters').to.be.equal(403);
                }
            });
        });
    }
);
