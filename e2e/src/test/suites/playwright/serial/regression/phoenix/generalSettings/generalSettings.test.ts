/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { flowRootState, idpUtils, logger, settings, tags } from '../../../../../../../modules/helpers';
import { SettingsKeyEnum } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { idpNameOidc } from '../../../../../../../modules/helpers/config/const';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';


        test.describe(`Security General Settings API, tags: ${tags.tool.api} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, { tag: [tags.services.tempBackendSettings] }, () => {
            shouldRunTests(settings.isSelfHosted && !isOCP())(
                function () {
            test.beforeAll(async function () {
                await idpUtils.updateSecuritySettingByKey(
                    flowRootState.users.system_administrator.accessToken,
                    SettingsKeyEnum.BrowserSessionTimeout,
                    {
                        idleTimeout: 60
                    }
                );
            });
            test.afterAll(async function () {
                try {
                    await idpUtils.updateSecuritySettingByKey(
                        flowRootState.users.system_administrator.accessToken,
                        SettingsKeyEnum.BrowserSessionTimeout,
                        {
                            idleTimeout: 60
                        }
                    );

                    await idpUtils.updateSecuritySettingByKey(
                        flowRootState.users.system_administrator.accessToken,
                        SettingsKeyEnum.LogoutRedirectUri,
                        {
                            uri: ''
                        }
                    );
                    await idpUtils.updateSecuritySettingByKey(
                        flowRootState.users.system_administrator.accessToken,
                        SettingsKeyEnum.AutoRedirectSso,
                        {
                            enabled: false,
                            idpAlias: idpNameOidc
                        }
                    );
                } catch (error) {
                    logger.error(`Error resetting security settings: ${JSON.stringify(error)}`);
                }
            });

            test('should validate get security settings api successfully', async function () {
                const securitySettings = await idpUtils.getSecuritySettings(
                    flowRootState.users.system_administrator.accessToken
                );
                expect(securitySettings.body, 'security settings response body should be defined').toBeDefined();
                expect(securitySettings.body.autoRedirectSSO, 'security settings should include autoRedirectSSO field').toBeDefined();
                expect(securitySettings.body.browserSessionTimeout, 'security settings should include browserSessionTimeout field').toBeDefined();
                expect(securitySettings.body.logoutRedirectUri, 'security settings should include logoutRedirectUri field').toBeDefined();
            });

            test('should validate get auto redirect SSO setting by key', async () => {
                const autoRedirectSSO = await idpUtils.getSecuritySettingByKey(
                    flowRootState.users.system_administrator.accessToken,
                    SettingsKeyEnum.AutoRedirectSso
                );
                expect(autoRedirectSSO.body, 'autoRedirectSSO setting response body should be defined').toBeDefined();
                expect(autoRedirectSSO.body.value.enabled, 'autoRedirectSSO should be disabled by default').toBe(false);
                expect(autoRedirectSSO.body.value.idpAlias, 'autoRedirectSSO setting should have an idpAlias field').toBeDefined();
            });

            test('should validate get browser session timeout setting by key', async () => {
                const browserSessionTimeout = await idpUtils.getSecuritySettingByKey(
                    flowRootState.users.system_administrator.accessToken,
                    SettingsKeyEnum.BrowserSessionTimeout
                );
                expect(browserSessionTimeout.body, 'browserSessionTimeout setting response body should be defined').toBeDefined();
                expect(browserSessionTimeout.body.value.idleTimeout, 'default browser session idle timeout should be 60').toBe(60);
            });

            test('should validate get logout redirect URI setting by key', async () => {
                const logoutRedirectUri = await idpUtils.getSecuritySettingByKey(
                    flowRootState.users.system_administrator.accessToken,
                    SettingsKeyEnum.LogoutRedirectUri
                );
                expect(logoutRedirectUri.body, 'logoutRedirectUri setting response body should be defined').toBeDefined();
                expect(logoutRedirectUri.body.value?.uri, 'default logout redirect URI should be empty string').toBe('');
            });

            test('should validate update auto redirect SSO setting by key', async () => {
                const updateAutoRedirectSSO = await idpUtils.updateSecuritySettingByKey(
                    flowRootState.users.system_administrator.accessToken,
                    SettingsKeyEnum.AutoRedirectSso,
                    {
                        enabled: true,
                        idpAlias: idpNameOidc
                    }
                );
                expect(updateAutoRedirectSSO.body, 'update autoRedirectSSO response body should be defined').toBeDefined();
                expect(updateAutoRedirectSSO.body.enabled, 'autoRedirectSSO should be enabled after update').toBe(true);
                expect(updateAutoRedirectSSO.body.idpAlias, 'autoRedirectSSO idpAlias should match the configured OIDC IDP').toBe(idpNameOidc);
            });

            test('should validate update browser session timeout setting by key', async () => {
                const updateBrowserSessionTimeout = await idpUtils.updateSecuritySettingByKey(
                    flowRootState.users.system_administrator.accessToken,
                    SettingsKeyEnum.BrowserSessionTimeout,
                    {
                        idleTimeout: 120
                    }
                );
                expect(updateBrowserSessionTimeout.body, 'update browserSessionTimeout response body should be defined').toBeDefined();
                expect(updateBrowserSessionTimeout.body.idleTimeout, 'browser session idle timeout should be 120 after update').toBe(120);
            });

            test('should validate update logout redirect URI setting by key', async () => {
                const updateLogoutRedirectUri = await idpUtils.updateSecuritySettingByKey(
                    flowRootState.users.system_administrator.accessToken,
                    SettingsKeyEnum.LogoutRedirectUri,
                    {
                        uri: 'https://example.com'
                    }
                );
                expect(updateLogoutRedirectUri.body, 'update logoutRedirectUri response body should be defined').toBeDefined();
                expect(updateLogoutRedirectUri.body.uri, 'logout redirect URI should be set to example.com after update').toBe('https://example.com');
                const updateLogoutRedirectUri2 = await idpUtils.updateSecuritySettingByKey(
                    flowRootState.users.system_administrator.accessToken,
                    SettingsKeyEnum.LogoutRedirectUri,
                    {
                        uri: 'https://exampleee.com'
                    }
                );
                expect(updateLogoutRedirectUri2.body, 'second update logoutRedirectUri response body should be defined').toBeDefined();
                expect(updateLogoutRedirectUri2.body.uri, 'logout redirect URI should be updated to exampleee.com on second update').toBe('https://exampleee.com');
            });
        });
    }
);
