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
    waitUtils
} from '../../../../../../../modules/helpers';
import {
    idpNameOidc,
    keycloakClientName,
    realmName
} from '../../../../../../../modules/helpers/config/const';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { StatusCodes } from 'http-status-codes';
import { test } from '@playwright/test';

test.describe(
    'Group name comma handling',
    {
        tag: [
            tags.teams.phoenix, tags.suites.legacySphinx,
            tags.tool.api,
            tags.runOn.serialCI,
            tags.services.identityManager,
            tags.services.keycloak
        ]
    },
    () => {
        shouldRunTests(settings.isSelfHosted && !isOCP())(function () {
            const realAdminGroup = 'runai-admins';
            const negativeCaseCommaGroup = `x,${realAdminGroup}`;
            const negativeCaseCommaGroupUsername = 'commagroupnegativeuser';
            const positiveCaseCommaGroup = `y,${realAdminGroup}`;
            const positiveCaseCommaGroupUsername = 'commagrouppositiveuser';

            test.beforeAll('Init Keycloak IDP, Client, Users, access rules', async function () {
                try {
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameOidc
                    );
                } catch (error) {
                    if (!error.message?.includes('404')) {
                        logger.debug(`IDP cleanup: ${error.message}`);
                    }
                }

                try {
                    await Keycloak.deleteRealm(realmName);
                    await waitUtils.waitForCondition(
                        async () => !(await Keycloak.getRealmByName(realmName)),
                        `realm ${realmName} deleted`,
                        true,
                        15,
                        1
                    );
                } catch (error) {
                    logger.debug(`Realm cleanup: ${error.message}`);
                }

                await Keycloak.createRealm(realmName);
                const client = await Keycloak.createClient(realmName, keycloakClientName, 'openid-connect');
                const clientSecret = await Keycloak.getClientSecret(realmName, client.id);
                await Keycloak.createUsers(
                    realmName,
                    [negativeCaseCommaGroupUsername, positiveCaseCommaGroupUsername],
                    [config.controlPlaneAdminPassword, config.controlPlaneAdminPassword]
                );
                await Keycloak.changeLoginTheme(realmName, 'runai-theme');
                await Keycloak.createKeycloakGroupsWithUsers(
                    realmName,
                    [negativeCaseCommaGroup, positiveCaseCommaGroup],
                    [negativeCaseCommaGroupUsername, positiveCaseCommaGroupUsername]
                );
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
                await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    realAdminGroup,
                    flowRootState.users.system_administrator.role.id
                );
            });

            test.afterAll('delete all comma group test items', async function () {
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

            test('user whose only group name contains a comma must not inherit the real group permissions', async function () {
                const commaUserToken = await authUtils.authenticateUser(
                    `${negativeCaseCommaGroupUsername}@example.com`,
                    config.controlPlaneAdminPassword,
                    true
                );

                try {
                    await authorizationUtils.createAccessRuleForGroup(
                        commaUserToken.accessToken,
                        'someOtherGroup',
                        flowRootState.users.system_administrator.role.id
                    );
                    expect.fail(
                        'user whose only group is "x,runai-admins" must not be able to create access rules'
                    );
                } catch (error) {
                    expect(
                        error.body.code,
                        'user must be forbidden (403), not treated as a member of runai-admins'
                    ).to.be.equal(StatusCodes.FORBIDDEN);
                }
            });

            test('user whose only group name contains a comma inherits permissions granted directly to that group', async function () {
                await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    positiveCaseCommaGroup,
                    flowRootState.users.system_administrator.role.id
                );

                const commaUserToken = await authUtils.authenticateUser(
                    `${positiveCaseCommaGroupUsername}@example.com`,
                    config.controlPlaneAdminPassword,
                    true
                );

                const accessRule = await authorizationUtils.createAccessRuleForGroup(
                    commaUserToken.accessToken,
                    'someOtherGroup2',
                    flowRootState.users.system_administrator.role.id
                );

                expect(
                    accessRule.subjectId,
                    'access rule must be created for the comma-containing group, unescaped'
                ).to.be.equal('someOtherGroup2');
            });
        });
    }
);
