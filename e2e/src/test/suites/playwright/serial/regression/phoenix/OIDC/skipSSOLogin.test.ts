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
    operations,
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
import { LoginUi } from '../../../../../../../modules/ui/runaiV2/pages/LoginUiUtils';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { GeneralSettings } from '../../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { UsersFilterSortFields } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { workloadTourUtils } from '../../../../../../../modules/ui/utils/guidedTourUtil';


        test.describe(`Skip SSO Login test, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.authentication}`, { tag: [tags.services.identityManager, tags.services.keycloak] }, function () {
            shouldRunTests(settings.isSelfHosted && !isOCP())(
                function () {
            let loginUi: LoginUi;
            let auth: AuthPage;
            let generalSettings: GeneralSettings;
            let uniqueUsername: string;

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
                    await idpUtils.autoRedirectSSO(
                        flowRootState.users.system_administrator.accessToken,
                        false,
                        idpNameOidc
                    );
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameOidc
                    );
                    await Keycloak.deleteRealm(realmName);
                    await operations.forceCleanup();
                } catch (error) {
                    logger.error(error);
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });

            test.beforeEach(async ({ page }) => {
                // Setup tour skip script before any navigation to prevent guided tour popups
                await workloadTourUtils.setupTourSkipScript(page);

                const timestamp = Date.now();
                uniqueUsername = `automationauskipsso_${timestamp}`;

                loginUi = new LoginUi(page);
                auth = new AuthPage(page);
                generalSettings = new GeneralSettings(page);
                await Keycloak.createUser(realmName, uniqueUsername, config.controlPlaneAdminPassword);
                await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    `${uniqueUsername}@example.com`,
                    flowRootState.users.system_administrator.role.id
                );
            });

            test.afterEach('cleanup', async () => {
                await idpUtils.autoRedirectSSO(
                    flowRootState.users.system_administrator.accessToken,
                    false,
                    idpNameOidc
                );
                try {
                    const users = await usersUtils.getUsersFilteredBy(
                        flowRootState.users.system_administrator.accessToken,
                        [`username=@${uniqueUsername}`],
                        UsersFilterSortFields.Type
                    );
                    for (const user of users) {
                        await usersManager.deleteLocalUser(
                            flowRootState.users.system_administrator.accessToken,
                            user.id
                        );
                    }

                    await Keycloak.deleteUserFromRealm(realmName, uniqueUsername);
                    logger.debug(
                        `Cleaned up user ${uniqueUsername} from both local system and Keycloak realm ${realmName}`
                    );
                } catch (error) {
                    logger.error('Failed to delete user from realm', error);
                }
            });

            test('should show standard sso login page', async function () {
                await auth.loginPageDisplayed();
                const ssoUserToken = await authUtils.authenticateUser(
                    uniqueUsername,
                    config.controlPlaneAdminPassword,
                    true
                );
                await idpUtils.autoRedirectSSO(ssoUserToken.idToken, false, idpNameOidc);
                const expectResult = await loginUi.ssoLoginPageShown();
                expect(expectResult, 'standard SSO login page should be shown when autoRedirectSSO is disabled').to.be.true;
            });

            test('should enable skipSSO login page', async function () {
                await auth.loginPageDisplayed();
                const ssoUserToken = await authUtils.authenticateUser(
                    uniqueUsername,
                    config.controlPlaneAdminPassword,
                    true
                );
                await idpUtils.autoRedirectSSO(ssoUserToken.idToken, true, idpNameOidc);
                const expectResult = await loginUi.enableSkipSSOLoginPageShown();
                expect(expectResult, 'skip SSO login option should be shown when autoRedirectSSO is enabled').to.be.eq(true);
            });

            test('should set incorrect IdP Alias via API -> Login Page Still Usable', async function () {
                await auth.loginPageDisplayed();
                const ssoUserToken = await authUtils.authenticateUser(
                    uniqueUsername,
                    config.controlPlaneAdminPassword,
                    true
                );
                const responseSSOLoginPage = await idpUtils.autoRedirectSSO(
                    ssoUserToken.idToken,
                    true,
                    'incorrect_idp_alias'
                );
                expect(responseSSOLoginPage.response.statusCode, 'setting incorrect IDP alias should return 200 OK').to.be.eq(200);
                expect(responseSSOLoginPage.response.statusMessage, 'setting incorrect IDP alias response message should be OK').to.be.eq('OK');
                const expectResult = await loginUi.ssoLoginPageShown();
                expect(expectResult, 'login page should remain usable with standard SSO page when IDP alias is incorrect').to.be.eq(true);
            });

            test('should enable and disable via UI, Logout Check', async function () {
                await auth.loginPageDisplayed();
                await auth.loginSSO(uniqueUsername, config.controlPlaneAdminPassword);
                expect(await loginUi.loggedUiValidator(), 'permitted SSO user should see the logged-in UI').to.be.eq(true);
                await generalSettings.navigateToGeneralSettingsSecurity();
                await generalSettings.toggleAutoRedirectSSO();
                await loginUi.logoutToLoginPage();
                const expectResult = await loginUi.enableSkipSSOLoginPageShown();
                expect(expectResult, 'skip SSO login option should be shown after enabling autoRedirectSSO via UI').to.be.eq(true);
                await auth.loginSSOWithAutoRedirectOn(uniqueUsername, config.controlPlaneAdminPassword);
                await generalSettings.navigateToGeneralSettingsSecurity();
                await generalSettings.disableAutoRedirectSSO();
                await loginUi.logoutToLoginPage();
                const expectResultDisable = await loginUi.ssoLoginPageShown();
                expect(expectResultDisable, 'standard SSO login page should be shown after disabling autoRedirectSSO via UI').to.be.eq(true);
            });

            test('should enable skipSSO login page and delete idp validate login page back to normal', async function () {
                await auth.loginPageDisplayed();
                const ssoUserToken = await authUtils.authenticateUser(
                    uniqueUsername,
                    config.controlPlaneAdminPassword,
                    true
                );
                await idpUtils.autoRedirectSSO(ssoUserToken.idToken, true, idpNameOidc);
                await idpUtils.deleteIdp(flowRootState.users.system_administrator.accessToken, idpNameOidc);
                await auth.loginPageDisplayed();
                const expectResult = await loginUi.normalLoginPageShown();
                expect(expectResult, 'normal login page should be restored after deleting the IDP').to.be.eq(true);
                await auth.signInValidateLoggedIn(
                    config.controlPlaneAdminUsername,
                    config.controlPlaneAdminPassword
                );
            });
        });
    }
);
