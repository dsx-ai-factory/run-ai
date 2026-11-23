/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    accessKeyUtils,
    authorizationUtils,
    authUtils,
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


        test.describe(`Groups at scale feature, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api}`, () => {
            shouldRunTests(settings.isSelfHosted && !isOCP())(
                function () {
            let clientSecret: string;
            test.beforeEach('Init Keycloak IDP, Client, Users', async function () {
                // Cleanup existing resources from failed previous runs
                try {
                    logger.debug('Cleaning up existing IDP and realm before test');
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

                // Now proceed with fresh creation
                await Keycloak.createRealm(realmName);
                const client = await Keycloak.createClient(realmName, keycloakClientName, 'openid-connect');
                clientSecret = await Keycloak.getClientSecret(realmName, client.id);
                await Keycloak.createUser(realmName, 'groupsuser', config.controlPlaneAdminPassword);
                await Keycloak.changeLoginTheme(realmName, 'runai-theme');
                await Keycloak.addOIDCMapperToClient(
                    realmName,
                    client.id,
                    'groups',
                    'oidc-group-membership-mapper',
                    'String'
                );
                await Keycloak.configLoginWithSSOOIDC(
                    Keycloak.getDiscoveryDocumentURL(realmName),
                    keycloakClientName,
                    clientSecret
                );
            });

            test.afterEach('delete all oidc group tests items', async function () {
                try {
                    const groupAccessRules = await authorizationUtils.getAccessRules(
                        { subjectType: 'group' },
                        flowRootState.users.system_administrator.accessToken
                    );
                    if (groupAccessRules.accessRules.length > 0) {
                        logger.debug('Found group access rules during cleanup, deleting them');
                        for (const accessRule of groupAccessRules.accessRules) {
                            try {
                                await authorizationUtils.deleteAccessRule(
                                    accessRule,
                                    flowRootState.users.system_administrator.accessToken
                                );
                            } catch (error) {
                                logger.warn(`Failed to delete group access rule ${accessRule.id}: ${error}`);
                            }
                        }
                    }
                } catch (error) {
                    logger.warn('Failed to clean up group access rules', error);
                }

                try {
                    const users = await usersUtils.getUsersFilteredBy(
                        flowRootState.users.system_administrator.accessToken,
                        ['username=@groupsuser'],
                        UsersFilterSortFields.Type
                    );
                    for (const user of users) {
                        await usersManager.deleteLocalUser(
                            flowRootState.users.system_administrator.accessToken,
                            user.id
                        );
                    }
                } catch (error) {
                    logger.error('Failed to delete user from realm', error);
                }
                try {
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameOidc
                    );
                } catch (error) {
                    logger.error('Failed to delete IDP', error);
                }
                try {
                    await Keycloak.deleteRealm(realmName);
                } catch (error) {
                    logger.error('Failed to delete Keycloak realm', error);
                }
            });

            test.afterAll(async function () {
                await usersManager.restorePreSeededOidcIdp();
            });

            test('should validate token contains all 5 user groups and correct groups count', async function () {
                const groups = Keycloak.generateGroups(5);
                await Keycloak.createKeycloakGroupsWithUsers(realmName, groups, ['groupsuser']);
                const user = await authUtils.authenticateUser(
                    'groupsuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const isValid = authUtils.validateAllParamsInToken(user.accessToken, {
                    groups: groups,
                    groups_count: groups.length
                });
                expect(isValid, 'token should contain all 5 user groups and correct groups_count').to.be.true;
            });

            test('should set access rule for 1 group and validate token contains 1 group and groups count claim to be 12 after delete access rule groups need to be empty and groups count claim to be 12', async function () {
                const groups = Keycloak.generateGroups(12);
                await Keycloak.createKeycloakGroupsWithUsers(realmName, groups, ['groupsuser']);
                const deleteRule = await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'group1',
                    flowRootState.users.system_administrator.role.id
                );
                const user = await authUtils.authenticateUser(
                    'groupsuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const isValid = authUtils.validateAllParamsInToken(user.accessToken, {
                    groups: ['group1'],
                    groups_count: 12
                });
                expect(isValid, 'token should contain only group1 and groups_count=12 when access rule exists for group1').to.be.true;
                await authorizationUtils.deleteAccessRule(
                    deleteRule,
                    flowRootState.users.system_administrator.accessToken
                );
                const user2 = await authUtils.authenticateUser(
                    'groupsuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const isValid2 = authUtils.validateAllParamsInToken(user2.accessToken, {
                    groups: [],
                    groups_count: 12
                });
                expect(isValid2, 'token should have empty groups but groups_count=12 after deleting the access rule').to.be.true;
            });

            test('should validate token groups when access rules are added and removed for multiple groups', async function () {
                const groups = Keycloak.generateGroups(12);
                await Keycloak.createKeycloakGroupsWithUsers(realmName, groups, ['groupsuser']);
                const deleteRule = await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'group1',
                    flowRootState.users.system_administrator.role.id
                );
                const user = await authUtils.authenticateUser(
                    'groupsuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const isValid = authUtils.validateAllParamsInToken(user.accessToken, {
                    groups: ['group1'],
                    groups_count: 12
                });
                expect(isValid, 'token should contain only group1 and groups_count=12 when one access rule exists').to.be.true;
                const deleteRule2 = await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'group2',
                    flowRootState.users.system_administrator.role.id
                );
                const user2 = await authUtils.authenticateUser(
                    'groupsuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const isValid2 = authUtils.validateAllParamsInToken(user2.accessToken, {
                    groups: ['group1', 'group2'],
                    groups_count: 12
                });
                expect(isValid2, 'token should contain group1 and group2 after adding second access rule').to.be.true;

                await authorizationUtils.deleteAccessRule(
                    deleteRule,
                    flowRootState.users.system_administrator.accessToken
                );
                const user3 = await authUtils.authenticateUser(
                    'groupsuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const isValid3 = authUtils.validateAllParamsInToken(user3.accessToken, {
                    groups: ['group2'],
                    groups_count: 12
                });
                expect(isValid3, 'token should contain only group2 after removing group1 access rule').to.be.true;

                await authorizationUtils.deleteAccessRule(
                    deleteRule2,
                    flowRootState.users.system_administrator.accessToken
                );
            });

            test('2 access rules for the same group with 12 groups', async function () {
                const groups = Keycloak.generateGroups(12);
                await Keycloak.createKeycloakGroupsWithUsers(realmName, groups, ['groupsuser']);
                const accessRule1 = await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'group1',
                    flowRootState.users.system_administrator.role.id
                );
                const accessRule2 = await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'group1',
                    flowRootState.users.editor.role.id
                );
                const user = await authUtils.authenticateUser(
                    'groupsuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const isValid = authUtils.validateAllParamsInToken(user.accessToken, {
                    groups: ['group1'],
                    groups_count: 12
                });
                expect(isValid, 'token should contain group1 when 2 access rules exist for the same group').to.be.true;

                await authorizationUtils.deleteAccessRule(
                    accessRule1,
                    flowRootState.users.system_administrator.accessToken
                );
                await authorizationUtils.deleteAccessRule(
                    accessRule2,
                    flowRootState.users.system_administrator.accessToken
                );
            });

            test('should validate token of access key have all 5 groups and correct groups', async function () {
                const groups = Keycloak.generateGroups(5);
                await Keycloak.createKeycloakGroupsWithUsers(realmName, groups, ['groupsuser']);
                const userToken = await authUtils.authenticateUser(
                    'groupsuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );

                const isValidUserToken = authUtils.validateAllParamsInToken(userToken.accessToken, {
                    groups: groups,
                    groups_count: groups.length
                });
                expect(isValidUserToken, 'user token should contain all 5 groups and correct groups_count').to.be.true;
                const accessKey = await accessKeyUtils.createAccessKey(userToken.accessToken, {
                    name: 'groupsuseraccesskey'
                });

                const accessKeyToken = await authUtils.authenticateAppAndGetToken(
                    accessKey.body.clientId,
                    accessKey.body.clientSecret
                );
                const isValidAccessKey = authUtils.validateAllParamsInToken(accessKeyToken.accessToken, {
                    groups: groups
                });
                expect(isValidAccessKey, 'access key token should contain all 5 user groups').to.be.true;
            });

            test('should validate token of access key have all 5 groups id token', async function () {
                const groups = Keycloak.generateGroups(5);
                await Keycloak.createKeycloakGroupsWithUsers(realmName, groups, ['groupsuser']);
                const userToken = await authUtils.authenticateUser(
                    'groupsuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );

                const isValidUserToken = authUtils.validateAllParamsInToken(userToken.accessToken, {
                    groups: groups,
                    groups_count: groups.length
                });
                expect(isValidUserToken, 'user token should contain all 5 groups and correct groups_count for id token test').to.be.true;
                const accessKey = await accessKeyUtils.createAccessKey(userToken.accessToken, {
                    name: 'groupsuseraccesskey'
                });

                const accessKeyToken = await authUtils.authenticateAppAndGetToken(
                    accessKey.body.clientId,
                    accessKey.body.clientSecret
                );
                const isValidAccessKey = authUtils.validateAllParamsInToken(accessKeyToken.accessToken, {
                    groups: groups
                });
                expect(isValidAccessKey, 'access key token id should contain all 5 user groups').to.be.true;
            });

            test('should show 1 group in token of access key with 100 groups', async function () {
                const groups = Keycloak.generateGroups(100);
                await Keycloak.createKeycloakGroupsWithUsers(realmName, groups, ['groupsuser']);
                const accessRule = await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'group1',
                    flowRootState.users.system_administrator.role.id
                );
                const userToken = await authUtils.authenticateUser(
                    'groupsuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const accessKey = await accessKeyUtils.createAccessKey(userToken.accessToken, {
                    name: 'groupsuseraccesskey'
                });

                const accessKeyToken = await authUtils.authenticateAppAndGetToken(
                    accessKey.body.clientId,
                    accessKey.body.clientSecret
                );
                const isValidAccessKey = authUtils.validateAllParamsInToken(accessKeyToken.accessToken, {
                    groups: ['group1']
                });
                expect(isValidAccessKey, 'access key token should contain only group1 when 100 groups exist but 1 access rule').to.be.true;
                await authorizationUtils.deleteAccessRule(
                    accessRule,
                    flowRootState.users.system_administrator.accessToken
                );
            });

            test('should show 2 group in token of access key with 100 groups and after delete access rule groups need to be empty', async function () {
                const groups = Keycloak.generateGroups(100);
                await Keycloak.createKeycloakGroupsWithUsers(realmName, groups, ['groupsuser']);
                const accessRule = await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'group1',
                    flowRootState.users.system_administrator.role.id
                );
                const accessRule2 = await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'group2',
                    flowRootState.users.system_administrator.role.id
                );
                const userToken = await authUtils.authenticateUser(
                    'groupsuser@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const accessKey = await accessKeyUtils.createAccessKey(userToken.accessToken, {
                    name: 'groupsuseraccesskey'
                });

                const accessKeyToken = await authUtils.authenticateAppAndGetToken(
                    accessKey.body.clientId,
                    accessKey.body.clientSecret
                );
                const isValidAccessKey = authUtils.validateAllParamsInToken(accessKeyToken.accessToken, {
                    groups: ['group1', 'group2']
                });
                expect(isValidAccessKey, 'access key token should contain group1 and group2 when 2 access rules exist with 100 groups').to.be.true;
                await authorizationUtils.deleteAccessRule(
                    accessRule,
                    flowRootState.users.system_administrator.accessToken
                );
                await authorizationUtils.deleteAccessRule(
                    accessRule2,
                    flowRootState.users.system_administrator.accessToken
                );
                const accessKeyToken2 = await authUtils.authenticateAppAndGetToken(
                    accessKey.body.clientId,
                    accessKey.body.clientSecret
                );
                const isValidAccessKey2 = authUtils.validateAllParamsInToken(accessKeyToken2.accessToken, {
                    groups: []
                });
                expect(isValidAccessKey2, 'access key token should have empty groups after deleting all access rules').to.be.true;
            });
        });
    }
);
