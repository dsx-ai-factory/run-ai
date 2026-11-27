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
    objectGenerator,
    operations,
    settings,
    tags,
    TokenObject,
    usersManager
} from '../../../../../../../modules/helpers';
import { idpNameSaml, realmName } from '../../../../../../../modules/helpers/config/const';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { User } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { AccessRule } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/accessRule';
import { StatusCodes } from 'http-status-codes';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

shouldRunTests(settings.isSelfHosted && !isOCP())(
    function () {
        test.describe(`Access Keys SSO SAML tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api}`, function () {
            let userToken: TokenObject;
            let userDetails: User;
            let accessRule: AccessRule;
            let ssoUser: string;

            test.beforeAll(async function () {
                await usersManager.deletePreSeededOidcIdp();
                // Cleanup existing resources from failed previous runs
                try {
                    logger.debug('Cleaning up existing IDP and realm before test suite');
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameSaml
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
                await Keycloak.setupSSORealmSAML(realmName);
                ssoUser = `ssouser${Date.now()}`;
                await Keycloak.createUser(realmName, ssoUser, config.controlPlaneAdminPassword);
            });
            test.afterAll(async function () {
                try {
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameSaml
                    );
                    await Keycloak.deleteRealm(realmName);
                    await operations.forceCleanup();
                } catch (error) {
                    logger.error('Error during cleanup:', error);
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });

            test('should SSO users without access rules create access keys', async function () {
                const userName = `nopermission${Date.now()}`;
                await Keycloak.createUser(realmName, userName, config.controlPlaneAdminPassword);
                userToken = await authUtils.authenticateUser(
                    `${userName}@example.com`,
                    config.controlPlaneAdminPassword,
                    true
                );
                userDetails = await authUtils.getUserV2(
                    `${userName}@example.com`,
                    flowRootState.users.system_administrator.accessToken
                );

                logger.debug(`Authenticated SSO SAML user: ${userDetails.username}`);
                logger.debug(`User ID: ${userDetails.id}`);
                const accessKeyName = objectGenerator.generateShortNameByPrefix('ssoadminaccesskey');
                const accessKey = await accessKeyUtils.createAccessKey(userToken.accessToken, {
                    name: accessKeyName
                });

                expect(accessKey.response.statusCode, 'SSO user without access rules should be able to create an access key with 201 Created').eql(StatusCodes.CREATED);
            });

            test('should validate sso users with access rules can create access keys', async function () {
                accessRule = await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    `${ssoUser}@example.com`,
                    flowRootState.users.editor.role.id
                );
                const accessKeyName = objectGenerator.generateShortNameByPrefix('withpermissionaccesskey');
                userToken = await authUtils.authenticateUser(
                    `${ssoUser}@example.com`,
                    config.controlPlaneAdminPassword,
                    true
                );
                const accessKey = await accessKeyUtils.createAccessKey(userToken.accessToken, {
                    name: accessKeyName
                });
                expect(accessKey.response.statusCode, 'SSO user with access rules should be able to create an access key with 201 Created').to.equal(201);
                await authorizationUtils.deleteAccessRule(
                    accessRule,
                    flowRootState.users.system_administrator.accessToken
                );
            });

            test('should validate sso users with access rules can delete access keys', async function () {
                const accessKeyName = objectGenerator.generateShortNameByPrefix('withpermissionaccesskey');
                accessRule = await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    `${ssoUser}@example.com`,
                    flowRootState.users.editor.role.id
                );
                userToken = await authUtils.authenticateUser(
                    `${ssoUser}@example.com`,
                    config.controlPlaneAdminPassword,
                    true
                );
                const accessKey = await accessKeyUtils.createAccessKey(userToken.accessToken, {
                    name: accessKeyName
                });
                expect(accessKey.response.statusCode, 'SSO user with access rules should be able to create an access key with 201 Created for deletion test').to.equal(201);

                const deleteResponse = await accessKeyUtils.deleteAccessKey(
                    userToken.accessToken,
                    accessKey.body.id
                );
                expect(deleteResponse.response.statusCode, 'SSO user with access rules should be able to delete an access key with 204 No Content').to.equal(204);
                await authorizationUtils.deleteAccessRule(
                    accessRule,
                    flowRootState.users.system_administrator.accessToken
                );
            });

            test('should validate sso users with access group without access rules can create access keys', async function () {
                const userNameEditorGroup = objectGenerator.generateShortNameByPrefix(`editor${Date.now()}`);
                const groupName = objectGenerator.generateShortNameByPrefix(`editorgroup${Date.now()}`);
                await Keycloak.createUser(realmName, userNameEditorGroup, config.controlPlaneAdminPassword);
                await Keycloak.createKeycloakGroupsWithUsers(realmName, [groupName], [userNameEditorGroup]);

                const accesGroup = await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    groupName,
                    flowRootState.users.editor.role.id
                );
                logger.debug(`Created group ${groupName} with user ${userNameEditorGroup}`);
                logger.debug(`Access rule created for group ${groupName}`);
                logger.debug(`Access rule: ${JSON.stringify(accesGroup)}`);
                userToken = await authUtils.authenticateUser(
                    `${ssoUser}@example.com`,
                    config.controlPlaneAdminPassword,
                    true
                );
                const accessKeyName = objectGenerator.generateShortNameByPrefix(
                    `withgroupaccesskey${Date.now()}`
                );
                const accessKey = await accessKeyUtils.createAccessKey(userToken.accessToken, {
                    name: accessKeyName
                });
                expect(accessKey.response.statusCode, 'SSO user with access group rule should be able to create an access key with 201 Created').to.equal(201);
            });
        });
    }
);
