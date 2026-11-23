/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authUtils,
    authorizationUtils,
    config,
    flowRootState,
    idpUtils,
    logger,
    settings,
    tags,
    usersManager
} from '../../../../../../../modules/helpers';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { idpNameSaml, samlUser } from '../../../../../../../modules/helpers/config/const';
import { expect } from 'chai';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

        test.describe(`SAML, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.ssoConfiguration}`, function () {
            shouldRunTests(settings.isSelfHosted && !isOCP())(
                function () {
            let ssoUser: string;
            test.afterAll(async function () {
                try {
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameSaml
                    );
                    await Keycloak.deleteRealm('autosaml');
                } catch (error) {
                    logger.error(error);
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });
            test.beforeAll(async function () {
                await usersManager.deletePreSeededOidcIdp();
                await Keycloak.setupSSORealmSAML('autosaml');
                ssoUser = `ssouser${Date.now()}`;
                await Keycloak.createUser('autosaml', ssoUser, config.controlPlaneAdminPassword);
            });
            test('should create SAML SSO and login', async function () {
                await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    `${ssoUser}@example.com`,
                    flowRootState.users.system_administrator.role.id
                );
                const token = await authUtils.authenticateUser(
                    `${ssoUser}@example.com`,
                    config.controlPlaneAdminPassword,
                    true
                );
                expect(token, 'SAML SSO user should receive a non-null token after login').to.not.be.null;
            });
        });
    }
);
