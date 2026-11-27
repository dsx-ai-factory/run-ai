/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { config, flowRootState, idpUtils, settings, tags } from '../../../../../../../modules/helpers';
import { idpNameOidc, idpNameSaml } from '../../../../../../../modules/helpers/config/const';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { GeneralSettings } from '../../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

   test.describe(`General Settings Security UI Validations, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, { tag: [tags.services.tempBackendSettings] }, () => {
    shouldRunTests(settings.isSelfHosted && !isOCP())(
        function () {        
    let browserEngine: BrowserEngine;
            let auth: AuthPage;
            let generalSettings: GeneralSettings;
            let videoName: string;

            test.beforeEach('init browsers and classes', async function () {
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                auth = new AuthPage(browserEngine.currentPage);
                await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
                generalSettings = new GeneralSettings(browserEngine.currentPage);
                await idpUtils.autoRedirectSSO(
                    flowRootState.users.system_administrator.accessToken,
                    false,
                    idpNameOidc
                );
                await idpUtils.autoRedirectSSO(
                    flowRootState.users.system_administrator.accessToken,
                    false,
                    idpNameSaml
                );
            });

            test.afterEach('stop browserEngine', async ({}, testInfo) => {
                try {
                    await browserEngine.stopAndArchive(testInfo, videoName);
                } finally {
                    browserEngine = new BrowserEngine();
                    await browserEngine.start();
                    auth = new AuthPage(browserEngine.currentPage);
                    await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
                    generalSettings = new GeneralSettings(browserEngine.currentPage);
                    await generalSettings.generalSecurityForceCleaner();
                    await browserEngine.stopAndArchive(testInfo, videoName);
                }
            });

            test('should check if tooltips are visible', async function () {
                videoName = 'shouldCheckIfTooltipsAreVisible';
                await generalSettings.navigateToGeneralSettingsSecurity();
                const tooltip = await generalSettings.isSecurityTooltipsVisible();
                expect(tooltip, 'security settings tooltips should be visible').to.be.equal(true);
            });

            test('should set SAML 2.0 Identity Provider and delete it', async function () {
                videoName = 'shouldSetSAML20IdentityProviderAndDeleteIt';
                await generalSettings.navigateToGeneralSettingsSecurity();
                await generalSettings.setIdentityProviderUI('SAML 2.0');
                expect(await generalSettings.isValidatorSAMLCreated(), 'SAML 2.0 IDP should be created successfully via UI').to.be.equal(true);
                await generalSettings.removeIdentityProvider();
            });

            test('should set OpenID Connect Identity Provider and delete it successfully', async function () {
                videoName = 'shouldSetOpenIDConnectIdentityProviderAndDeleteIt';
                await generalSettings.navigateToGeneralSettingsSecurity();
                await generalSettings.setIdentityProviderUI('OpenID Connect');
                expect(await generalSettings.isValidatorOIDCCreated(), 'OpenID Connect IDP should be created successfully via UI').to.be.equal(true);
                await generalSettings.removeIdentityProvider();
            });

            test('should set OpenID and validate scopes', async function () {
                videoName = 'shouldSetOpenIDAndValidateScopes';
                await generalSettings.navigateToGeneralSettingsSecurity();
                const validate = await generalSettings.validateOIDCScopes();
                expect(validate, 'OpenID Connect scopes should be valid').to.be.equal(true);
            });

            test('should set and remove logout redirect URL', async function () {
                videoName = 'shouldSetAndRemoveLogoutRedirectURL';
                await generalSettings.navigateToGeneralSettingsSecurity();
                await generalSettings.setLogoutRedirectURL('https://www.nvidia.com/en-eu/');
                await generalSettings.removeLogoutUrl();
            });

            test('should verify session ends after the timeout', async function () {
                videoName = 'shouldVerifySessionEndsAfterTheTimeout';
                await generalSettings.navigateToGeneralSettingsSecurity();
                await generalSettings.setSessionTimeout('1');
                await generalSettings.verifySessionEndAfterTimePass();
            });

            test('should set and verify logout redirect URL validation', async function () {
                videoName = 'shouldSetAndVerifyLogoutRedirectURLValidation';
                await generalSettings.navigateToGeneralSettingsSecurity();
                await generalSettings.setLogoutRedirectURL('https://www.nvidia.com/en-eu/');
                await generalSettings.logoutAndVerifyUrl();
            });

            test('should validate all settings options are clickable without errors', async function () {
                videoName = 'shouldValidateAllSettingsOptionsAreClickableWithoutErrors';
                await generalSettings.navigateToGeneralSettingsSecurity();
                await generalSettings.validateAllOptionsClickable();
            });

            test('should validate that all settings menus have names', async function () {
                videoName = 'shouldValidateThatAllSettingsMenusHaveNames';
                await generalSettings.navigateToGeneralSettingsSecurity();
                await generalSettings.validateAllSectionsHaveName();
            });

            test('should success update idp when the DiscoveryUrl empty successfully', async function () {
                videoName = 'shouldSuccessUpdateIdpWhenTheDiscoveryUrlEmpty';
                await generalSettings.navigateToGeneralSettingsSecurity();
                await generalSettings.setIdentityProviderUI('OpenID Connect');
                const isUpdated = await generalSettings.updateIdpWhenTheDiscoveryUrlEmpty();
                expect(isUpdated, 'IDP update should succeed when discovery URL is empty').to.be.equal(true);
            });

            test('should validate logo upload, edit and remove by ui', async function () {
                videoName = 'shouldValidateLogo';
                await generalSettings.openGeneralSettingsPage();
                await generalSettings.clickonBranding();
                await generalSettings.removeExistingLogoIfExists();
                await generalSettings.clickonBranding();
                await generalSettings.uploadLogo();
                expect(await generalSettings.isLogoUploaded(), 'logo should be uploaded successfully').to.be.equal(true);
                await generalSettings.editLogo('./logo.png');
                expect(await generalSettings.isLogoUploaded(), 'logo should still be present after editing').to.be.equal(true);
                await generalSettings.removeExistingLogoIfExists();
                expect(await generalSettings.isLogoRemoved(), 'logo should be removed successfully').to.be.equal(true);
            });
        });
    }
);
