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
import { idpNameSaml, samlRealmName } from '../../../../../../../modules/helpers/config/const';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

test.describe(`SAML SSO non-email user by minimal api, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, function () {
    shouldRunTests(
        settings.isSelfHosted && !isOCP()
    )(function () {
        test.beforeAll(async function () {
            await usersManager.deletePreSeededOidcIdp();
            try {
                await idpUtils.deleteIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameSaml
                );
            } catch (error) {
                if (!error.message?.includes('404')) {
                    logger.debug(`IDP cleanup: ${error.message}`);
                }
            }
            try {
                await Keycloak.deleteRealm(samlRealmName);
                await new Promise((resolve) => setTimeout(resolve, 1000));
            } catch (error) {
                logger.debug(`Realm cleanup: ${error.message}`);
            }
            await Keycloak.setupSSORealmSAMLForNonEmailUsers(samlRealmName);
        });

        test.afterAll(async function () {
            try {
                await idpUtils.deleteIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameSaml
                );
                await Keycloak.deleteRealm(samlRealmName);
            } catch (error) {
                logger.error(error);
            } finally {
                await usersManager.restorePreSeededOidcIdp();
            }
        });

        test('non-email SAML SSO user with firstName/lastName: login + list + minimal + delete', async function () {
            const username = `samlnoemailnamed${Date.now()}`;
            const firstName = 'NoEmail';
            const lastName = 'SamlNames';
            await Keycloak.createUserWithoutEmail(
                samlRealmName,
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
            expect(token.accessToken, 'non-email SAML SSO user with names should receive a string access token').to.be.a('string').and.not.empty;

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

        test('non-email SAML SSO user with username only: login + list + minimal + delete', async function () {
            const username = `samlnoemailbare${Date.now()}`;
            await Keycloak.createUserWithoutEmail(
                samlRealmName,
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
            expect(token.accessToken, 'non-email SAML SSO user with username only should receive a string access token').to.be.a('string').and.not.empty;

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
