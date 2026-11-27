/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

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
import {
    keycloakClientName,
    realmName,
    idpNameOidc
} from '../../../../../../../modules/helpers/config/const';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { expect, test } from '@playwright/test';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { UsersFilterSortFields } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';


        test.describe(`OIDC mandatory claim ${tags.teams.phoenix} ${tags.suites.legacySphinx} `, () => {
            shouldRunTests(settings.isSelfHosted && !isOCP())(
                function () {
            let auth: AuthPage;
            let browserEngine: BrowserEngine;
            let videoName: string;

            test.beforeEach(async () => {
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

                // Now proceed with fresh setup
                await Keycloak.setupSSORealmOIDC(realmName, keycloakClientName, {
                    claim: 'claim',
                    values: ['my-value']
                });
            });

            test.afterEach(async () => {
                try {
                    const users = await usersUtils.getUsersFilteredBy(
                        flowRootState.users.system_administrator.accessToken,
                        ['username=@claim-user'],
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
                } finally {
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameOidc
                    );
                    await Keycloak.deleteRealm(realmName);
                }
            });

            test.afterAll(async ({}, testInfo) => {
                try {
                    if (browserEngine) {
                        await browserEngine.stopAndArchive(testInfo, videoName);
                    }
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });

            test('should define claim as mandatory and validate the claim login', async () => {
                videoName = 'shouldDefineClaimAsMandatoryAndValidateTheClaimLogin';
                await Keycloak.addRealmSettingsUserProfileAttribute(realmName, 'claim', 'claim');
                await Keycloak.createUser(realmName, 'claim-user', config.controlPlaneAdminPassword);
                await Keycloak.updateUserDetailsGeneral(realmName, 'claim-user', {
                    username: 'claim-user',
                    email: 'claim-user@example.com',
                    firstName: 'claim',
                    lastName: 'user',
                    enabled: true,
                    attributes: {
                        claim: ['my-value']
                    }
                });
                await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    'claim-user@example.com',
                    flowRootState.users.system_administrator.role.id
                );

                const token = await authUtils.authenticateUser(
                    'claim-user@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                expect(token.idToken, 'SSO user with matching claim value should receive an id token').toBeDefined();
            });

            test('should define idp with claim and validate missing claim login fail ', async () => {
                videoName = 'shouldDefineIdpWithClaimAndValidateMissingClaimLoginFail';
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                auth = new AuthPage(browserEngine.currentPage);
                await Keycloak.createUser(realmName, 'claim-user', config.controlPlaneAdminPassword);
                await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    'claim-user@example.com',
                    flowRootState.users.system_administrator.role.id
                );
                const result = await auth.loginSSOandVerifyMissingClaim(
                    'claim-user@example.com',
                    config.controlPlaneAdminPassword
                );
                expect(result, 'login attempt without the mandatory claim should show the missing-claim error page').toBe(true);
            });

            test('should validate claim with multiple values', async () => {
                videoName = 'shouldValidateClaimWithMultipleValues';
                const idp = await idpUtils.getIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameOidc
                );
                expect(idp.body.oidcData.mandatoryClaim, 'IDP mandatory claim should be defined').toBeDefined();
                expect(idp.body.oidcData.mandatoryClaim.claim, 'mandatory claim key should be "claim"').toBe('claim');
                expect(idp.body.oidcData.mandatoryClaim.values, 'mandatory claim values should be ["my-value"]').toEqual(['my-value']);

                const discoveryURL = Keycloak.getDiscoveryDocumentURL(realmName);
                await idpUtils.updateIdpOidc(
                    flowRootState.users.system_administrator.accessToken,
                    idp.body.alias,
                    idp.body.oidcData.clientId,
                    idp.body.oidcData.clientSecret,
                    discoveryURL,
                    idp.body.oidcData.scopes,
                    {
                        claim: 'claim',
                        values: ['my-value2']
                    }
                );
                const idpAfterUpdate = await idpUtils.getIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameOidc
                );
                expect(idpAfterUpdate.body.oidcData.mandatoryClaim, 'updated IDP mandatory claim should be defined').toBeDefined();
                expect(idpAfterUpdate.body.oidcData.mandatoryClaim.claim, 'updated mandatory claim key should still be "claim"').toBe('claim');
                expect(idpAfterUpdate.body.oidcData.mandatoryClaim.values, 'updated mandatory claim values should be ["my-value2"]').toEqual(['my-value2']);
                await Keycloak.addRealmSettingsUserProfileAttribute(realmName, 'claim', 'claim');
                await Keycloak.createUser(realmName, 'claim-user', config.controlPlaneAdminPassword);
                await Keycloak.updateUserDetailsGeneral(realmName, 'claim-user', {
                    username: 'claim-user',
                    email: 'claim-user@example.com',
                    firstName: 'claim',
                    lastName: 'user',
                    enabled: true,
                    attributes: {
                        claim: ['my-value2']
                    }
                });
                await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    'claim-user@example.com',
                    flowRootState.users.system_administrator.role.id
                );

                const token = await authUtils.authenticateUser(
                    'claim-user@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                expect(token.idToken, 'SSO user with matching "my-value2" claim should receive an id token').toBeDefined();
                await idpUtils.updateIdpOidc(
                    flowRootState.users.system_administrator.accessToken,
                    idp.body.alias,
                    idp.body.oidcData.clientId,
                    idp.body.oidcData.clientSecret,
                    discoveryURL,
                    idp.body.oidcData.scopes,
                    {
                        claim: 'claim',
                        values: ['my-value2', 'my-value3']
                    }
                );
                const idpAfterMultipleValues = await idpUtils.getIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameOidc
                );
                expect(idpAfterMultipleValues.body.oidcData.mandatoryClaim, 'IDP mandatory claim should be defined after multi-value update').toBeDefined();
                expect(idpAfterMultipleValues.body.oidcData.mandatoryClaim.claim, 'mandatory claim key should still be "claim" after multi-value update').toBe('claim');
                expect(idpAfterMultipleValues.body.oidcData.mandatoryClaim.values, 'mandatory claim values should contain both my-value2 and my-value3').toEqual([
                    'my-value2',
                    'my-value3'
                ]);
            });
        });
    }
);
shouldRunTests(settings.isSelfHosted && !isOCP())(
    function () {
        test.describe(`OIDC mandatory claim multiple values ${tags.teams.phoenix} ${tags.suites.legacySphinx} `, () => {
            let browserEngine: BrowserEngine;
            let videoName: string;

            test.beforeEach(async () => {
                await usersManager.deletePreSeededOidcIdp();
                try {
                    await Keycloak.deleteRealm(realmName);
                    await new Promise((resolve) => setTimeout(resolve, 1000));
                } catch (error) {
                    logger.debug(`Realm cleanup: ${error.message}`);
                }
                await Keycloak.setupSSORealmOIDC(realmName, keycloakClientName, {
                    claim: 'claim',
                    values: ['my-value', 'my-value2']
                });
            });

            test.afterEach(async ({}, testInfo) => {
                try {
                    const users = await usersUtils.getUsersFilteredBy(
                        flowRootState.users.system_administrator.accessToken,
                        ['username=@claim-user'],
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
                } finally {
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameOidc
                    );
                    await Keycloak.deleteRealm(realmName);
                }
            });

            test.afterAll(async ({}, testInfo) => {
                try {
                    if (browserEngine) {
                        await browserEngine.stopAndArchive(testInfo, videoName);
                    }
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });

            test('should validate claim with multiple values', async () => {
                videoName = 'shouldValidateClaimWithMultipleValues';
                const idp = await idpUtils.getIdp(
                    flowRootState.users.system_administrator.accessToken,
                    idpNameOidc
                );
                expect(idp.body.oidcData.mandatoryClaim, 'IDP mandatory claim should be defined for multi-value test').toBeDefined();
                expect(idp.body.oidcData.mandatoryClaim.claim, 'mandatory claim key should be "claim" for multi-value test').toBe('claim');
                expect(idp.body.oidcData.mandatoryClaim.values, 'mandatory claim values should contain my-value and my-value2').toEqual(['my-value', 'my-value2']);
                await Keycloak.addRealmSettingsUserProfileAttribute(realmName, 'claim', 'claim');
                await Keycloak.createUser(realmName, 'claim-user', config.controlPlaneAdminPassword);
                await Keycloak.updateUserDetailsGeneral(realmName, 'claim-user', {
                    username: 'claim-user',
                    email: 'claim-user@example.com',
                    firstName: 'claim',
                    lastName: 'user',
                    enabled: true,
                    attributes: {
                        claim: ['my-value', 'my-value2']
                    }
                });
                await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    'claim-user@example.com',
                    flowRootState.users.system_administrator.role.id
                );
                const token = await authUtils.authenticateUser(
                    'claim-user@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                expect(token.idToken, 'SSO user matching any of the multiple claim values should receive an id token').toBeDefined();
            });
        });
    }
);
