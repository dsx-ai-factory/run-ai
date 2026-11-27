/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    config,
    flowRootState,
    idpUtils,
    logger,
    objectGenerator,
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
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { LoginUi } from '../../../../../../../modules/ui/runaiV2/pages/LoginUiUtils';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';


        test.describe(
            `Login OIDC test, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.tool.api} ${tags.resource.authentication} ${tags.runOn.serialCI}`,
            { tag: [tags.services.identityManager, tags.services.keycloak, tags.services.tenantsManager , tags.runOn.basicCI] },
            function () {
                shouldRunTests(settings.isSelfHosted && !isOCP())(
                    function () {
                let browserEngine: BrowserEngine;
                let auth: AuthPage;
                let loginUi: LoginUi;
                let nonPermissionUser: string;

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

                    await Keycloak.setupSSORealmOIDC(realmName, keycloakClientName);
                });

                test.afterAll(async () => {
                    try {
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

                test.beforeEach(async () => {
                    browserEngine = new BrowserEngine();
                    await browserEngine.start();
                    auth = new AuthPage(browserEngine.currentPage);
                    loginUi = new LoginUi(browserEngine.currentPage);
                });

                test.afterEach('stop browser', async ({}, testInfo) => {
                    await browserEngine.stopAndArchive(testInfo, 'sso login');
                });

                test('should login without permission sso and validate', async function () {
                    nonPermissionUser = objectGenerator.generateShortNameByPrefix('nopermission');
                    await Keycloak.createUser(realmName, nonPermissionUser, config.controlPlaneAdminPassword);
                    await auth.loginSSO(`${nonPermissionUser}@example.com`, config.controlPlaneAdminPassword);
                    expect(await loginUi.nonPermissionLoginValidator()).to.be.true;
                });
            }
        );
    }
);
