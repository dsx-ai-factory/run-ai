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
    objectGenerator,
    settings,
    tags,
    usersManager,
    usersUtils
} from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    shouldRunTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { UsersV2 } from '../../../../../../modules/ui/runaiV2/pages/UsersV2';
import { isOCP } from '../../../../../../modules/helpers/utils/EnvironmentUtils';
import { Keycloak } from '../../../../../../modules/helpers/utils/Keycloak';
import { idpNameOidc, keycloakClientName, realmName } from '../../../../../../modules/helpers/config/const';
import { UsersFilterSortFields } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';

test.describe(`Full Name Display SSO, tags: ${tags.teams.loong} ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.ui} ${tags.resource.users} ${tags.resource.ssoConfiguration}`, () => {
    shouldRunTests(settings.isSelfHosted && !isOCP() && process.env.INIT_SSO_USERS_AND_IDP !== 'true')(
        function () {
            supportedFrom(SupportedVersions.V2_26)(function () {
                let browserEngine: BrowserEngine;
                let authPage: AuthPage;
                let usersPage: UsersV2;
                let videoName: string;

                const ssoFirstName = 'Sso';
                const ssoLastName = 'User';
                const ssoFullName = `${ssoFirstName} ${ssoLastName}`;
                const ssoUsername = objectGenerator.generateShortNameByPrefix('fullname-sso-user');
                const ssoUserEmail = `${ssoUsername}@example.com`;

                test.beforeAll(async function () {
                    browserEngine = new BrowserEngine();

                    try {
                        await idpUtils.deleteIdp(
                            flowRootState.users.system_administrator.accessToken,
                            idpNameOidc
                        );
                    } catch (error) {
                        if (!error.message?.includes('404')) {
                            logger.debug(`IDP cleanup before SSO full-name test: ${error.message}`);
                        }
                    }

                    try {
                        await Keycloak.deleteRealm(realmName);
                        await new Promise((resolve) => setTimeout(resolve, 1000));
                    } catch (error) {
                        logger.debug(`Realm cleanup before SSO full-name test: ${error.message}`);
                    }

                    await Keycloak.setupSSORealmOIDC(realmName, keycloakClientName);
                    await Keycloak.createUser(realmName, ssoUsername, config.controlPlaneAdminPassword);
                    await Keycloak.updateUserDetailsGeneral(realmName, ssoUsername, {
                        username: ssoUsername,
                        email: ssoUserEmail,
                        firstName: ssoFirstName,
                        lastName: ssoLastName,
                        enabled: true,
                        emailVerified: true
                    });
                    await authorizationUtils.createAccessRuleForUser(
                        flowRootState.users.system_administrator.accessToken,
                        ssoUserEmail,
                        flowRootState.users.system_administrator.role.id
                    );

                    const token = await authUtils.authenticateUser(
                        ssoUsername,
                        config.controlPlaneAdminPassword,
                        true
                    );
                    await authUtils.getUserV2(ssoUserEmail, token.accessToken);
                });

                test.beforeEach(async function () {
                    await browserEngine.start();
                    authPage = new AuthPage(browserEngine.currentPage);
                    usersPage = new UsersV2(browserEngine.currentPage);

                    await authPage.signIn(
                        flowRootState.users.system_administrator.email,
                        flowRootState.users.system_administrator.password
                    );
                });

                test.afterEach('stop browser', async ({}, testInfo) => {
                    await browserEngine.stopAndArchive(testInfo, videoName);
                });

                test.afterAll(async function () {
                    try {
                        const usersToDelete = await usersUtils.getUsersFilteredBy(
                            flowRootState.users.system_administrator.accessToken,
                            [`username=@${ssoUserEmail}`],
                            UsersFilterSortFields.Username
                        );
                        const ssoUser = usersToDelete.find((user) => user.username === ssoUserEmail);
                        if (ssoUser) {
                            await usersManager.deleteSSOUser(
                                flowRootState.users.system_administrator.accessToken,
                                ssoUserEmail,
                                ssoUser.id
                            );
                        }
                    } catch (error) {
                        logger.error('failed to delete SSO full-name user', JSON.stringify(error));
                    }

                    try {
                        await idpUtils.deleteIdp(
                            flowRootState.users.system_administrator.accessToken,
                            idpNameOidc
                        );
                    } catch (error) {
                        logger.error(`failed to delete SSO IDP ${idpNameOidc}`, JSON.stringify(error));
                    }

                    try {
                        await Keycloak.deleteRealm(realmName);
                    } catch (error) {
                        logger.error(`failed to delete SSO realm ${realmName}`, JSON.stringify(error));
                    }
                });

                test('SSO user names are sourced from IdP and cannot be edited by admin (T6037710)', async () => {
                    videoName = 'sso-user-idp-name-readonly';

                    await test.step('verify the SSO user displays the IdP-provided full name', async () => {
                        await usersPage.verifyUserTableFullName(ssoUserEmail, ssoFullName);
                        await usersPage.verifyUserDetailDoesNotContainInvalidName(ssoUserEmail, ssoFullName);
                    });

                    await test.step('verify edit action is not available for SSO user', async () => {
                        await usersPage.verifyExternalUserEditActionNotAvailable(ssoUserEmail);
                    });
                });
            });
        }
    );
});
