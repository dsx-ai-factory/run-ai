/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    apiUtils,
    config,
    flowRootState,
    idpUtils,
    logger,
    settings,
    tags,
    Utils
} from '../../../../../../../modules/helpers';
import { runaiV1 } from '../../../../../../../modules/ui';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { SettingsKeyEnum } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/settingsKeyEnum';
import { expect } from 'chai';
import { test } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';

test.describe(
    `Login and Logout Functionality: ${tags.teams.phoenix} ${tags.suites.legacySphinx}`,
    { tag: [tags.services.frontend, tags.services.identityManager] },
    () => {
        shouldRunTests(settings.isSelfHosted && !isOCP())(
            function () {
                test.beforeEach('Reset logout Uri', async () => {
                    await idpUtils.updateSecuritySettingByKey(
                        flowRootState.users.system_administrator.accessToken,
                        SettingsKeyEnum.LogoutRedirectUri,
                        { uri: '' }
                    );
                });

                test.afterAll('Cleanup', async () => {
                    try {
                        await idpUtils.updateSecuritySettingByKey(
                            flowRootState.users.system_administrator.accessToken,
                            SettingsKeyEnum.LogoutRedirectUri,
                            { uri: '' }
                        );
                    } catch (error) {
                        logger.error('Failed to reset logout uri back to runai homepage', error);
                    }
                    try {
                        await apiUtils.setInactivityLogoutTimeout(60);
                    } catch (error) {
                        logger.error('Failed to reset logout timeout back to default', error);
                    }
                });

                test.describe('Test login and logout', () => {
                    test.beforeEach('start browser', async () => {
                        await runaiV1.login.start();
                    });

                    test.afterEach('stop browser', async ({}, testInfo) => {
                        await runaiV1.login.stopAndArchive(testInfo.title);
                    });

                    test('should log-in and log-out', async () => {
                        await runaiV1.login.redirectAndSignIn(
                            config.controlPlaneAdminUsername,
                            config.controlPlaneAdminPassword
                        );
                        const logoutValidation = await runaiV1.login.logoutToLoginPage();
                        expect(logoutValidation, 'logout should redirect back to the login page').to.be.true;
                        await runaiV1.login.signIn(
                            config.controlPlaneAdminUsername,
                            config.controlPlaneAdminPassword
                        );
                    });
                });

                test.describe('Test set logout URI', () => {
                    const logoutUri =
                        'https://www.nvidia.com/en-us/software/run-ai/';

                    test.beforeEach('start browser and init logout uri setting', async () => {
                        await idpUtils.updateSecuritySettingByKey(
                            flowRootState.users.system_administrator.accessToken,
                            SettingsKeyEnum.LogoutRedirectUri,
                            { uri: logoutUri }
                        );
                        await runaiV1.login.start();

                        await runaiV1.login.redirectAndSignIn(
                            config.controlPlaneAdminUsername,
                            config.controlPlaneAdminPassword
                        );
                    });

                    test.afterEach('Stop browser', async ({}, testInfo) => {
                        try {
                            await idpUtils.updateSecuritySettingByKey(
                                flowRootState.users.system_administrator.accessToken,
                                SettingsKeyEnum.LogoutRedirectUri,
                                { uri: '' }
                            );
                        } catch (error) {
                            logger.error('Failed to reset logout uri back to runai homepage', error);
                        } finally {
                            await runaiV1.login.stopAndArchive(testInfo.title);
                        }
                    });

                    test('should save logout URI', async () => {
                        const response = await idpUtils.getSecuritySettingByKey(
                            flowRootState.users.system_administrator.accessToken,
                            SettingsKeyEnum.LogoutRedirectUri
                        );
                        expect(response.body.value.uri, 'saved logout URI should match the configured value').to.equal(logoutUri);
                    });

                    test('should change logout URI and log out to the specified URL', async () => {
                        if (settings.isSelfHosted) {
                            const response = await idpUtils.getSecuritySettingByKey(
                                flowRootState.users.system_administrator.accessToken,
                                SettingsKeyEnum.LogoutRedirectUri
                            );
                            expect(response.body.value.uri, 'logout URI setting should match the configured value before logout').to.equal(logoutUri);
                            const uri = await runaiV1.login.logout();
                            logger.info(`Logout redirected to: ${uri}`);
                            logger.info(`Expected logout URI: ${logoutUri}`);
                            expect(uri, 'logout should redirect to the configured logout URI').to.equal(logoutUri);

                            await runaiV1.login.redirectAndSignIn(
                                config.controlPlaneAdminUsername,
                                config.controlPlaneAdminPassword,
                                false,
                                false
                            );
                        }
                    });
                });

                test.describe('Test inactivity logout timeout', () => {
                    const newTimeoutMinutes = 1;
                    const newTimeoutMilliseconds = newTimeoutMinutes * 60 * 1000;

                    test.beforeEach('start browser', async () => {
                        await apiUtils.setInactivityLogoutTimeout(newTimeoutMinutes);
                        await runaiV1.login.start();
                    });

                    test.afterEach('stop browser', async ({}, testInfo) => {
                        await runaiV1.login.stopAndArchive(testInfo.title);
                    });

                    async function validateOnLoginPage(): Promise<void> {
                        await runaiV1.login.page.waitForLoadState('networkidle');
                        const url = runaiV1.login.page.url();
                        await runaiV1.login.page.waitForTimeout(3000);
                        if (settings.isSelfHosted) {
                            expect(url, 'redirect URL after inactivity should contain tenantUrl').to.contain(config.tenantUrl);
                            expect(url, 'redirect URL after inactivity should contain appUrl').to.contain(config.appUrl);
                        } else {
                            expect(url, 'redirect URL after inactivity should be defined').not.undefined;
                        }
                        await runaiV1.login.page.waitForTimeout(3000);
                        await runaiV1.login.signIn(
                            config.controlPlaneAdminUsername,
                            config.controlPlaneAdminPassword
                        );
                    }

                    async function validateStillLoggedIn(): Promise<void> {
                        await runaiV1.login.page.waitForSelector(runaiV1.login.profileIcon, {
                            timeout: 5000
                        });
                        // url should be of tenant - not login
                        const url = runaiV1.login.page.url();
                        expect(Utils.strStartsWith(url, `${config.appUrl}/auth`), 'user should not be on the auth/login page when still logged in').to.eq(false);
                        expect(Utils.strStartsWith(url, `${config.tenantUrl}`), 'user URL should start with tenantUrl when still logged in').to.eq(true);
                    }

                    test('should change and verify inactivity logout time', async () => {
                        await runaiV1.login.redirectAndSignIn(
                            config.controlPlaneAdminUsername,
                            config.controlPlaneAdminPassword
                        );

                        const time = await apiUtils.getLogoutTimeout();
                        expect(time, 'inactivity logout timeout should be updated to the new value').to.equal(newTimeoutMinutes);
                    });

                    test('should log out after inactivity timeout expires', async () => {
                        await runaiV1.login.redirectAndSignIn(
                            config.controlPlaneAdminUsername,
                            config.controlPlaneAdminPassword
                        );

                        await Utils.delayUI(runaiV1.login.page, newTimeoutMilliseconds * 2);
                        await validateOnLoginPage();
                    });

                    test('should validate the user stays logged in before inactivity timeout expires', async () => {
                        await runaiV1.login.redirectAndSignIn(
                            config.controlPlaneAdminUsername,
                            config.controlPlaneAdminPassword
                        );

                        await Utils.delayUI(runaiV1.login.page, newTimeoutMilliseconds / 2);
                        await validateStillLoggedIn();
                    });
                });
            }
        );
    }
);
