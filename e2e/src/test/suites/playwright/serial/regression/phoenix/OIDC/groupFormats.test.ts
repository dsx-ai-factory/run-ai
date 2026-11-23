/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authUtils,
    authorizationUtils,
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
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { test } from '@playwright/test';
import { UsersFilterSortFields } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';

shouldRunTests(settings.isSelfHosted && !isOCP())(
    function () {
        test.describe(`Group tests with formats OIDC, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.ssoConfiguration}`, { tag: [tags.services.identityManager, tags.services.keycloak] }, () => {
            let clientSecret: string;

            test.beforeAll('Init Keycloak IDP, Client, Users', async function () {
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
                await Keycloak.createRealm(realmName);
                const client = await Keycloak.createClient(realmName, keycloakClientName, 'openid-connect');
                clientSecret = await Keycloak.getClientSecret(realmName, client.id);
                await Keycloak.createUsers(
                    realmName,
                    ['formatgroupuser'],
                    [config.controlPlaneAdminPassword]
                );
                await Keycloak.changeLoginTheme(realmName, 'runai-theme');
                await Keycloak.createKeycloakGroupsWithUsers(realmName, ['formatgroup'], ['formatgroupuser']);
                await Keycloak.addOIDCMapperToClient(
                    realmName,
                    client.id,
                    'groupMapper',
                    'oidc-usermodel-attribute-mapper',
                    'JSON',
                    'groupsAttribute'
                );
                await Keycloak.configLoginWithSSOOIDC(
                    Keycloak.getDiscoveryDocumentURL(realmName),
                    keycloakClientName,
                    clientSecret
                );
                await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'formatgroup',
                    flowRootState.users.system_administrator.role.id
                );
                await Keycloak.addRealmSettingsUserProfileAttribute(
                    realmName,
                    'groupsAttribute',
                    'groupAttribute'
                );
                await Keycloak.updateUserDetailsGeneral(realmName, 'formatgroupuser', {
                    username: 'formatgroupuser',
                    email: 'formatgroupuser@example.com',
                    firstName: 'format',
                    lastName: 'format',
                    enabled: true,
                    attributes: {
                        groupsAttribute: '{"groupa": "formatgroup"}'
                    }
                });
            });

            test.afterAll('delete all oidc group tests items', async function () {
                try {
                    try {
                        await idpUtils.deleteIdp(
                            flowRootState.users.system_administrator.accessToken,
                            idpNameOidc
                        );
                    } catch (error) {
                        logger.error(error);
                    }
                    try {
                        await Keycloak.deleteRealm(realmName);
                    } catch (error) {
                        logger.error(error);
                    }
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });

            test('should validate new json group format', async function () {
                await authUtils.authenticateUser(
                    'formatgroupuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const usersCount = await usersUtils.getUsersFilteredBy(
                    flowRootState.users.system_administrator.accessToken,
                    ['username=@formatgroupuser@example.com'],
                    UsersFilterSortFields.Username
                );
                expect(usersCount.length, 'at least one user with formatgroupuser username should exist').to.be.greaterThan(0);
                expect(usersCount[0].groups, 'user should be assigned to the formatgroup group').to.eqls(['formatgroup']);
            });

            test('should validate jwt token with all parameters', async function () {
                const token = await authUtils.authenticateUser(
                    'formatgroupuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );

                const isValid = authUtils.validateAllParamsInToken(token.accessToken, {
                    iss: `${config.tenantUrl}/auth/realms/runai`,
                    aud: ['runai', 'runai-dummy'],
                    sub: 'formatgroupuser@example.com',
                    typ: 'Bearer',
                    azp: 'runai-admin-ui',
                    acr: '1',
                    'allowed-origins': [config.tenantUrl],
                    scope: 'openid profile email',
                    tenant_id: flowRootState.cluster.tenantId,
                    identity_provider: idpNameOidc,
                    email_verified: true,
                    name: 'format format',
                    groups: ['formatgroup'],
                    preferred_username: 'formatgroupuser',
                    given_name: 'format',
                    family_name: 'format',
                    email: 'formatgroupuser@example.com'
                });

                expect(isValid, 'JWT token should contain all expected claims with correct values').to.be.true;
            });
        });
    }
);
