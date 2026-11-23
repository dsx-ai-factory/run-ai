/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
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
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import {
    idpNameOidc,
    keycloakClientName,
    realmName
} from '../../../../../../../modules/helpers/config/const';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

test.describe(`OIDC SSO non-email user, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, function () {
    shouldRunTests(
        settings.isSelfHosted && !isOCP()
    )(function () {
        test.beforeAll(async function () {
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
                await new Promise((resolve) => setTimeout(resolve, 1000));
            } catch (error) {
                logger.debug(`Realm cleanup: ${error.message}`);
            }
            await Keycloak.setupSSORealmOIDCForNonEmailUsers(realmName, keycloakClientName);
        });

        test.afterAll(async function () {
            try {
                await idpUtils.deleteIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameOidc
                );
                await Keycloak.deleteRealm(realmName);
            } catch (error) {
                logger.error(error);
            } finally {
                await usersManager.restorePreSeededOidcIdp();
            }
        });

        test('non-email SSO user with firstName/lastName: login + list + minimal + delete', async function () {
            const username = `noemailnamed${Date.now()}`;
            const firstName = 'NoEmail';
            const lastName = 'WithNames';
            await Keycloak.createUserWithoutEmail(
                realmName,
                username,
                config.controlPlaneAdminPassword,
                firstName,
                lastName
            );
            await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                username,
                flowRootState.users.system_administrator.role.id
            );

            const token = await authUtils.authenticateUser(
                username,
                config.controlPlaneAdminPassword,
                true
            );
            expect(token.accessToken, 'non-email OIDC SSO user with names should receive a string access token').to.be.a('string').and.not.empty;

            const user = await usersUtils.validateNonEmailSSoUser(
                flowRootState.users.system_administrator.accessToken,
                username,
                firstName,
                lastName
            );

            await authUtils.deleteUser(
                user.id,
                flowRootState.users.system_administrator.accessToken
            );
        });

        test('non-email SSO user with username only: login + list + minimal + delete', async function () {
            const username = `noemailbare${Date.now()}`;
            await Keycloak.createUserWithoutEmail(
                realmName,
                username,
                config.controlPlaneAdminPassword
            );
            await authorizationUtils.createAccessRuleForUser(
                flowRootState.users.system_administrator.accessToken,
                username,
                flowRootState.users.system_administrator.role.id
            );

            const token = await authUtils.authenticateUser(
                username,
                config.controlPlaneAdminPassword,
                true
            );
            expect(token.accessToken, 'non-email OIDC SSO user with username only should receive a string access token').to.be.a('string').and.not.empty;

            const user = await usersUtils.validateNonEmailSSoUser(
                flowRootState.users.system_administrator.accessToken,
                username
            );

            await authUtils.deleteUser(
                user.id,
                flowRootState.users.system_administrator.accessToken
            );
        });

    });
});
