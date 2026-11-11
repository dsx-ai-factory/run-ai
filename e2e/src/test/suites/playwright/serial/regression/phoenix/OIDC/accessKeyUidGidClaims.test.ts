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
    settings,
    tags,
    usersManager
} from '../../../../../../../modules/helpers';
import {
    idpNameOidc,
    keycloakClientName,
    realmName
} from '../../../../../../../modules/helpers/config/const';
import { AccessRule } from '../../../../../../../modules/api/apis/authorizationApi/generated/api';
import { ScopeType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { test } from '@playwright/test';
import { AccessKeyPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { IncomingMessage } from 'http';

// user-application service accounts (personal access keys) must carry
// uid/gid/supplementarygroups claims sourced from the real user who created them
// (via runai.created.by), matching what regular user login tokens already carry.
test.describe(
    `Access key uid/gid/supplementarygroups claims, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.ssoConfiguration}`,
    { tag: [tags.services.identityManager, tags.services.keycloak] },
    () => {
        shouldRunTests(settings.isSelfHosted && !isOCP())(() => {
            const username = 'uidgidaccesskeyuser';
            const password = config.controlPlaneAdminPassword;
            const uid = '1234';
            const gid = '5678';
            const supplementaryGroups = ['100', '200'];
            let createdAccessKey: { response: IncomingMessage; body: AccessKeyPostResponse };
            let accessRule: AccessRule;

            test.beforeAll('Init Keycloak IDP, client and user with uid/gid/supplementarygroups', async function () {
                // Cleanup existing resources from failed previous runs
                try {
                    await idpUtils.deleteIdp(flowRootState.users.system_administrator.accessToken, idpNameOidc);
                } catch (error) {
                    if (!error.message?.includes('404')) {
                        logger.debug(`IDP cleanup: ${error.message}`);
                    }
                }
                try {
                    await Keycloak.deleteRealm(realmName);
                    await new Promise((resolve) => setTimeout(resolve, 1000));
                } catch (error) {
                    logger.debug(`Realm cleanup: ${error.message}`);
                }

                const clientSecret = await Keycloak.setupSSORealmOIDCForUidGidClaims(
                    realmName,
                    keycloakClientName,
                    username,
                    password,
                    uid,
                    gid,
                    supplementaryGroups
                );

                accessRule = await authorizationUtils.createAccessRuleForUserByRoleName(
                    `${username}@example.com`,
                    'L1 researcher',
                    ScopeType.Tenant,
                    flowRootState.cluster.tenantId.toString()
                );

                // Ensure the IDP attribute mappers for UID/GID/SUPPLEMENTARYGROUPS are
                // fully wired in the Run.ai identity manager (createIdp sets basic config;
                // updateIdpOidc adds the mappers that copy claims to user attributes).
                await idpUtils.updateIdpOidc(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameOidc,
                    keycloakClientName,
                    clientSecret,
                    Keycloak.getDiscoveryDocumentURL(realmName),
                    ['openid', 'email', 'profile'],
                    undefined,
                    undefined,
                    {
                        uid: 'UID',
                        gid: 'GID',
                        groups: 'GROUPS',
                        supplementaryGroups: 'SUPPLEMENTARYGROUPS',
                        email: 'email'
                    }
                );
            });

            test.afterAll('delete all uid/gid access key claim test items', async function () {
                try {
                    if (createdAccessKey) {
                        try {
                            await accessKeyUtils.deleteAccessKey(
                                flowRootState.users.system_administrator.accessToken,
                                createdAccessKey.body.id
                            );
                        } catch (error) {
                            logger.debug(`Access key cleanup: ${error.message}`);
                        }
                    }
                    if (accessRule) {
                        try {
                            await authorizationUtils.deleteAccessRule(
                                accessRule,
                                flowRootState.users.system_administrator.accessToken
                            );
                        } catch (error) {
                            logger.debug(`Access rule cleanup: ${error.message}`);
                        }
                    }
                    try {
                        await idpUtils.deleteIdp(flowRootState.users.system_administrator.accessToken, idpNameOidc);
                    } catch (error) {
                        logger.debug(`IDP cleanup: ${error.message}`);
                    }
                    try {
                        await Keycloak.deleteRealm(realmName);
                    } catch (error) {
                        logger.debug(`Realm cleanup: ${error.message}`);
                    }
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });

            test('should propagate the creating user uid/gid/supplementarygroups into the access key token', async function () {
                const userToken = await authUtils.authenticateUser(`${username}@example.com`, password, true);

                createdAccessKey = await accessKeyUtils.createAccessKey(userToken.accessToken, {
                    name: objectGenerator.generateShortNameByPrefix('uidgidaccesskey')
                });

                const accessKeyToken = await authUtils.authenticatePersonalApp(
                    createdAccessKey.body.clientId,
                    createdAccessKey.body.clientSecret
                );

                const isValid = authUtils.validateAllParamsInToken(accessKeyToken.accessToken, {
                    uid: uid,
                    gid: gid,
                    supplementarygroups: supplementaryGroups,
                    client_type: 'user-application'
                });

                expect(
                    isValid,
                    'access key token should contain uid/gid/supplementarygroups claims sourced from the creating user'
                ).to.be.true;
            });
        });
    }
);
