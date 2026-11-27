/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authorizationUtils,
    config,
    flowRootState,
    idpUtils,
    logger,
    settings,
    tags,
    usersManager
} from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { GeneralSettings } from '../../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { test } from '@playwright/test';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { idpNameSaml, samlUser } from '../../../../../../../modules/helpers/config/const';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

test.describe(`Link to SAML, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, { tag: [tags.services.tempBackendSettings] }, () => {
shouldRunTests(settings.isSelfHosted && !isOCP())(
    function () {
            let browserEngine: BrowserEngine;
            let auth: AuthPage;
            let generalSettings: GeneralSettings;
            let videoName: string;

            test.beforeAll('create saml idp', async function () {
                await usersManager.deletePreSeededOidcIdp();
                await Keycloak.setupSSORealmSAML('linktosaml');
                await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    samlUser,
                    flowRootState.users.system_administrator.role.id
                );
            });

            test.beforeEach('init browsers and classes', async function () {
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                auth = new AuthPage(browserEngine.currentPage);
                await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
                generalSettings = new GeneralSettings(browserEngine.currentPage);
            });

            test.afterAll('delete saml idp', async function () {
                try {
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameSaml
                    );

                    await Keycloak.deleteRealm('linktosaml');
                } catch (error) {
                    logger.error(error);
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });

            test.afterEach('stop browserEngine', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, videoName);
            });

            test('should validate link to SAML descriptor', async function () {
                videoName = 'shouldValidateLinkToSamlDescriptor';
                await generalSettings.navigateToGeneralSettingsSecurity();
                const url = await generalSettings.validateLinkToSamlDescriptor();
                logger.debug(`url to saml descriptor get from general settings: ${url}`);
                const expectedUrl = `${config.tenantUrl}/auth/realms/runai/broker/saml/endpoint/descriptor`;
                logger.debug(`expected url to saml descriptor: ${expectedUrl}`);
                expect(url, 'SAML descriptor link should point to the correct Keycloak broker endpoint URL').to.eql(expectedUrl);
            });
        });
    }
);
