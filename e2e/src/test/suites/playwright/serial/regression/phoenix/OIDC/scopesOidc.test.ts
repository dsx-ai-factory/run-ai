/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    flowRootState,
    idpUtils,
    logger,
    operations,
    settings,
    tags,
    usersManager
} from '../../../../../../../modules/helpers';
import {
    idpNameOidc,
    keycloakClientName,
    realmName
} from '../../../../../../../modules/helpers/config/const';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';


        test.describe(`Scopes IDP ${idpNameOidc}, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.ssoConfiguration}`, { tag: [tags.services.keycloak] }, function () {
            shouldRunTests(settings.isSelfHosted && !isOCP())(
                function () {
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
                    await Keycloak.deleteRealm(realmName);
                    await operations.forceCleanup();
                } catch (error) {
                    logger.error(error);
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });

            test('should check scopes sso', async function () {
                const idp = await idpUtils.getIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameOidc
                );
                expect(idp.body.oidcData, 'OIDC IDP data should have a scopes property').to.have.property('scopes');
                expect(idp.body.oidcData.scopes, 'IDP scopes should include openid').to.contain('openid');
                expect(idp.body.oidcData.scopes, 'IDP scopes should include email').to.contain('email');
            });

            test('should update scopes sso', async function () {
                const defaultScope = ['openid', 'email', 'profile', 'offline_access'];
                const discoveryURL = Keycloak.getDiscoveryDocumentURL('runai');
                await idpUtils.updateIdpOidc(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameOidc,
                    'clientId',
                    'clientSecret',
                    discoveryURL,
                    defaultScope
                );
                const idp = await idpUtils.getIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameOidc
                );
                expect(idp.body.oidcData, 'updated OIDC IDP data should have a scopes property').to.have.property('scopes');
                expect(idp.body.oidcData.scopes, 'updated IDP scopes should include openid').to.contain('openid');
                expect(idp.body.oidcData.scopes, 'updated IDP scopes should include email').to.contain('email');
                expect(idp.body.oidcData.scopes, 'updated IDP scopes should include profile').to.contain('profile');
                expect(idp.body.oidcData.scopes, 'updated IDP scopes should include offline_access').to.contain('offline_access');
            });
        });
    }
);
