/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    auditLogsUtils,
    flowRootState,
    logger,
    notificationsUtils,
    tags,
    usersManager,
    waitUtils
} from '../../../../../../../modules/helpers';
import { StatusCodes } from 'http-status-codes';
import { Page, test, TestInfo } from '@playwright/test';
import { expect as expectPlaywright } from '@playwright/test';
import { config, settings } from '../../../../../../../modules/helpers/config/config';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { SMTP_CONFIG } from '../../../../../../../modules/helpers/utils/NotificationUtils';
import { AuditLog } from '../../../../../../../modules/api/apis/auditApi';

    test.describe(`Users Send Mail Notification, tags: ${tags.resource.users} ${tags.tool.api} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`,
        { tag: [tags.runOn.serialCI, tags.services.notificationsProxy, tags.services.notificationsService, tags.services.keycloak] },
        () => {
            shouldRunTests(!isOCP())(function () {
        let accessToken: string;

        async function configureEmailServer(): Promise<void> {
            const conf = notificationsUtils.getSmtpChannel([]);

            try {
                if (settings.isSelfHosted) {
                    conf.config['from'] = 'test@example.com';
                    await notificationsUtils.createOrPatchEmailChannel(accessToken, conf);
                } else {
                    await notificationsUtils.patchEmailChannel(accessToken, conf);
                }
                logger.info('Email channel configured successfully');
                
                logger.info('Getting email channel for validation');
                const channel = await notificationsUtils.getEmailChannel(accessToken);
                logger.info('Email channel retrieved successfully');
                expectPlaywright(channel.config.smtpHost).toEqual(SMTP_CONFIG.host);
                logger.info('SMTP host validated successfully');
                expectPlaywright(channel.config.smtpPort).toEqual(587);
                logger.info('SMTP port validated successfully');
                expectPlaywright(channel.config.username).toEqual(SMTP_CONFIG.user);
                logger.info('Username validated successfully');
                expectPlaywright(channel.config.smtpTlsEnabled).toEqual(true);
                logger.info('SMTP TLS enabled validated successfully');
            } catch (error) {
                logger.error('Error getting email channel:', error);
                throw error;
            }
        }

        test.beforeAll(async function () {
            if (settings.isSelfHosted) {
                accessToken = flowRootState.users.system_administrator.accessToken;
            } else {
                const systemScopedToken = await notificationsUtils.getSystemScopedSuperAdminToken();
                accessToken = systemScopedToken.accessToken;
            }
            await configureEmailServer();
            await notificationsUtils.enableNotifications(accessToken);
        });

        test.afterAll(async function () {
            try {
                await notificationsUtils.disableNotifications(accessToken);
            } finally {
                await notificationsUtils.restoreSaasEmailChannel();
            }
        });

        test('should validate send mail notification by creating a user', async function ({}, testInfo: TestInfo) {
            const testStartTime = new Date();
            const email = `e2esmtp${Date.now()}@example.com`;
            const user = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                email,
                false,
                true
            );
            expectPlaywright(user.response.statusCode).toEqual(StatusCodes.CREATED);
            const emailReceived = await notificationsUtils.checkEmailReceivedWithDelay(
                (subject: string) =>
                    subject.includes('Welcome') || subject.includes('Welcome to NVIDIA Run:ai'),
                email
            );
            expectPlaywright(emailReceived).toBeTruthy();
            const emailLinksResult = await notificationsUtils.getEmailLinksAndValidate(
                SMTP_CONFIG.inboxId,
                SMTP_CONFIG.apiToken,
                email,
                (subject: string) =>
                    subject.includes('Welcome') || subject.includes('Welcome to NVIDIA Run:ai'),
                false
            );
            expectPlaywright(emailLinksResult).not.toBeNull();
            expectPlaywright(emailLinksResult!.emailReceived).toBeTruthy();
            expectPlaywright(emailLinksResult!.links.length).toBeGreaterThan(0);
            expectPlaywright(emailLinksResult!.linkValidation.validLinks.length).toEqual(5);
            expectPlaywright(emailLinksResult!.linkValidation.invalidLinks.length).toBe(0);
            const browserEngine = new BrowserEngine();
            let page: Page;
            try {
                await browserEngine.start();
                page = browserEngine.currentPage;
                logger.debug('Browser started successfully for test');

                const signInLink = await notificationsUtils.getSignInLinkFromEmailWithDelay(
                    SMTP_CONFIG.inboxId,
                    SMTP_CONFIG.apiToken,
                    email,
                    (subject: string) =>
                        subject.includes('Welcome') || subject.includes('Welcome to NVIDIA Run:ai'),
                    25
                );

                expectPlaywright(signInLink).not.toBeNull();
                if (!signInLink) {
                    throw new Error('Sign-in link not found');
                }

                logger.debug('Navigating to sign-in link');
                await page.goto(signInLink);
                await page.waitForLoadState('networkidle');
                const currentUrl = page.url();
                logger.debug('Current URL after navigation:', currentUrl);
                expectPlaywright(currentUrl).toContain('auth');
                await page.click('text=» Click here to proceed');
                await page.waitForLoadState('networkidle');
                await page.fill('#password-new', config.controlPlaneAdminPassword);
                await page.fill('#password-confirm', config.controlPlaneAdminPassword);
                logger.debug('Submitting password reset');
                await page.locator('#kc-passwd-update-form input[type="submit"]').click();
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(2000);
                logger.debug('Checking for success message');
                try {
                    const successMessage = await page.textContent('text=Your account has been updated', {
                        timeout: 5000
                    });
                    logger.debug('Success message found:', successMessage);
                    expectPlaywright(successMessage).toContain('Your account has been updated');
                    logger.debug('Password reset completed successfully!');
                } catch (messageError) {
                    logger.error('Could not find success message:', messageError);
                    throw messageError;
                }
                const filters = ['action=@Password set', `subject==${email}`];
                await waitUtils.waitForCondition(
                    async () =>
                        ((
                            await auditLogsUtils.getAuditLogs(
                                flowRootState.users.system_administrator.accessToken,
                                testStartTime,
                                new Date(),
                                filters
                            )
                        ).body.auditLogs?.length ?? 0) > 0,
                    'waiting for the "Password set" audit log to appear',
                    true,
                    60,
                    5
                );

                const auditLogs = await auditLogsUtils.getAuditLogs(
                    flowRootState.users.system_administrator.accessToken,
                    testStartTime,
                    new Date(),
                    filters
                );
                expectPlaywright(auditLogs.response.statusCode).toEqual(StatusCodes.OK);
                const setEntries = auditLogs.body.auditLogs ?? [];
                expectPlaywright(
                    setEntries.length,
                    'expected exactly one "Password set" audit entry for the invited user'
                ).toEqual(1);
                const setEntry = setEntries[0];
                expectPlaywright(setEntry.subject).toEqual(email);
                expectPlaywright(String(setEntry.subjectType).toLowerCase()).toEqual('user');
                expectPlaywright(setEntry.result).toEqual(AuditLog.ResultEnum.Succeeded);
                expectPlaywright(String(setEntry.action)).toContain('Password set');
                logger.debug('Password set audit validated: single entry, correct subject/result');
            } catch (error) {
                logger.error('Test failed with error:', error);
                throw error;
            } finally {
                if (browserEngine && browserEngine.browser) {
                    try {
                        if (browserEngine.browser.isConnected()) {
                            if (browserEngine.context) {
                                await browserEngine.context.close();
                            }
                            await browserEngine.browser.close();
                            logger.debug('Browser closed successfully');
                        } else {
                            logger.debug('Browser was already closed');
                        }
                    } catch (closeError) {
                        logger.error('Error closing browser:', closeError);
                    }
                }
            }
        });

        test('should validate send mail notification by creating a bulk of users', async function () {
            const usersEmails = [`e2ebulk${Date.now()}@example.com`, `bulk${Date.now()}@example.com`];
            const users = await usersManager.createBulkLocalUsers(
                flowRootState.users.system_administrator.accessToken,
                usersEmails,
                false,
                true
            );
            expectPlaywright(users.response.statusCode).toEqual(StatusCodes.OK);

            for (const userEmail of usersEmails) {
                const userEmailReceived = await notificationsUtils.checkEmailReceivedWithDelay(
                    (subject: string) =>
                        subject.includes('Welcome') || subject.includes('Welcome to NVIDIA Run:ai'),
                    userEmail
                );
                expectPlaywright(userEmailReceived).toBeTruthy();
            }
        });

        test('should validate forgot password notification and verify audit log', async function () {
            const testStartTime = new Date();
            const email = `e2eforgotp${Date.now()}@example.com`;
            const user = await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                email,
                false,
                false
            );
            const browserEngine = new BrowserEngine();
            let page: Page;
            try {
                await browserEngine.start();
                page = browserEngine.currentPage;
                logger.debug('Browser started successfully for test');
                await page.goto(config.tenantUrl);
                await page.waitForLoadState('networkidle');
                await page.locator('a[href*="reset-credentials"]').click();
                await page.waitForLoadState('networkidle');
                await page.fill('#username', email);
                await page.locator('#kc-reset-password-form input[type="submit"]').click();
                await page.waitForLoadState('networkidle');
                await page.waitForTimeout(2000);
                const emailReceived = await notificationsUtils.checkEmailReceivedWithDelay(
                    (subject: string) => subject.includes('Create New NVIDIA Run:ai password'),
                    email,
                    25
                );
                expectPlaywright(user.response.statusCode).toEqual(StatusCodes.CREATED);
                expectPlaywright(emailReceived).toBeTruthy();
                const signInLink = await notificationsUtils.getSignInLinkFromEmailWithDelay(
                    SMTP_CONFIG.inboxId,
                    SMTP_CONFIG.apiToken,
                    email,
                    (subject: string) => subject.includes('Create New NVIDIA Run:ai password'),
                    25
                );
                expectPlaywright(signInLink).not.toBeNull();
                if (!signInLink) {
                    throw new Error('Sign-in link not found in password reset email');
                }

                expectPlaywright(signInLink!).toContain('auth');
                await page.goto(signInLink!);
                await page.waitForLoadState('networkidle');
                await page.fill('#password-new', config.controlPlaneAdminPassword);
                await page.fill('#password-confirm', config.controlPlaneAdminPassword);
                logger.debug('Submitting password reset form');
                await page.locator('#kc-passwd-update-form input[type="submit"]').click();
                await page.waitForLoadState('networkidle');
                const filters = ['action=@Password reset', `subject==${email}`];
                await waitUtils.waitForCondition(
                    async () =>
                        ((
                            await auditLogsUtils.getAuditLogs(
                                flowRootState.users.system_administrator.accessToken,
                                testStartTime,
                                new Date(),
                                filters
                            )
                        ).body.auditLogs?.length ?? 0) > 0,
                    'waiting for the "Password reset" audit log to appear',
                    true,
                    60,
                    5
                );

                const auditLogs = await auditLogsUtils.getAuditLogs(
                    flowRootState.users.system_administrator.accessToken,
                    testStartTime,
                    new Date(),
                    filters
                );
                expectPlaywright(auditLogs.response.statusCode).toEqual(StatusCodes.OK);
                const resetEntries = auditLogs.body.auditLogs ?? [];
                expectPlaywright(
                    resetEntries.length,
                    'expected exactly one "Password reset" audit entry for the user'
                ).toEqual(1);
                const resetEntry = resetEntries[0];
                expectPlaywright(resetEntry.subject).toEqual(email);
                expectPlaywright(String(resetEntry.subjectType).toLowerCase()).toEqual('user');
                expectPlaywright(resetEntry.result).toEqual(AuditLog.ResultEnum.Succeeded);
                expectPlaywright(String(resetEntry.action)).toContain('Password reset');
                logger.debug('Password reset audit validated: single entry, correct subject/result');
            } catch (error) {
                logger.error('Test failed with error:', error);
                throw error;
            } finally {
                if (browserEngine && browserEngine.browser) {
                    try {
                        if (browserEngine.browser.isConnected()) {
                            if (browserEngine.context) {
                                await browserEngine.context.close();
                            }
                            await browserEngine.browser.close();
                            logger.debug('Browser closed successfully');
                        } else {
                            logger.debug('Browser was already closed');
                        }
                    } catch (closeError) {
                        logger.error('Error closing browser:', closeError);
                    }
                }
            }
        });
    });
});
