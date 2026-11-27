/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test, TestInfo } from '@playwright/test';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { WorkloadPage } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import {
    flowRootState,
    config,
    Utils,
    waitUtils,
    workloadsUtilsV2,
    logger,
    crds,
    settings,
    tags,
    notificationsUtils,
    authUtils
} from '../../../../../../../modules/helpers';
import { GeneralSettings } from '../../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { UsersV2 } from '../../../../../../../modules/ui/runaiV2/pages/UsersV2';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import { AllureUtils } from '../../../../../../../modules/helpers/utils/AllureUtils';
import { submitTrainWorkload } from '../../../../../../../modules/helpers/models/workloads';
import { createProject } from '../../../../../../../modules/helpers/models/projects';
import { TokenResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { saasConsoleUrl, tenantClientSecret } from '../../../../../../../modules/helpers/config/config';
import { shouldRunNativeWorkloadCrdTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(
    `Notifications Service Tests - Email Notifications for Workload Events, tags: ${tags.tool.ui} ${tags.tool.api} ${tags.teams.phoenix}`,
    {
        tag: [
            tags.runOn.serialCI,
            tags.services.notificationsProxy,
            tags.services.notificationsService,
            tags.services.frontend
        ]
    },
    () => {
        const EMAIL_CHECK_INTERVAL_MS = 20_000;
        const MAX_EMAIL_CHECK_ATTEMPTS = 25;
        const NO_EMAIL_CHECK_ATTEMPTS = 16;
        const WORKLOAD_WAIT_TIMEOUT_SECONDS = 120;
        const TEST_TIMEOUT_MS = 600_000; // 10 minutes. It takes time for the email to be received. Currently the workloads notifications are sent in interval every 5 minutes.

        type NotificationTestContext = {
            browserEngine: BrowserEngine;
            authPage: AuthPage;
            workloadPage: WorkloadPage;
            generalSettingsPage: GeneralSettings;
            usersV2Page: UsersV2;
            project: Project;
            namespace: string;
            videoName: string;
        };

        let ctx: NotificationTestContext;
        let systemScopedToken: TokenResponse;
        const allureUtils = new AllureUtils();

        async function initProject(): Promise<void> {
            const options = { maxAllowedGpus: -1, deservedGpus: 1 };
            ctx.project = await createProject('e2e-notifications-service', options);
            ctx.namespace = Utils.namespaceFromProjectName(ctx.project.name);
        }

        async function configSMTPServer(): Promise<void> {
            await ctx.generalSettingsPage.navigateToGeneralSettingsNotifications();
            const channel = notificationsUtils.getSmtpChannel([
                flowRootState.users.system_administrator.email
            ]);

            if (settings.isSelfHosted) {
                await ctx.generalSettingsPage.configureEmailServerSh(channel.config);
            } else {
                await notificationsUtils.patchEmailChannel(systemScopedToken.accessToken, channel);
                await ctx.generalSettingsPage.enableNotificationsSaas();
            }
        }

        async function disableNotifications(): Promise<void> {
            await ctx.generalSettingsPage.navigateToGeneralSettingsNotifications();
            if (settings.isSelfHosted) {
                await ctx.generalSettingsPage.deleteEmailServerSh();
            } else {
                await ctx.generalSettingsPage.disableNotificationsSaas();
            }
        }

        async function createTrainingWorkload(): Promise<{ workloadName: string; workloadId: string }> {
            const workload = await submitTrainWorkload(
                'notifications-test',
                ctx.project.name,
                undefined, // imagePath (uses default)
                undefined, // command (uses default)
                undefined, // numGpus (uses default)
                undefined, // podStatus (uses default)
                undefined, // nodePoolName
                undefined, // cpu
                undefined, // memory
                undefined, // cpuLimit
                undefined, // memoryLimit
                flowRootState.users.system_administrator.email
            );

            const workloadId = await workloadsUtilsV2.getWorkloadIdByWorkloadNameAndNamespace(
                crds.runaiTrainingWorkloadCrd,
                workload.name,
                ctx.namespace
            );

            logger.info(`Created workload: ${workload.name}`);

            return { workloadName: workload.name, workloadId };
        }

        async function checkEmailReceivedWithDelay(
            validSubjectHandler: (str: string) => boolean,
            attempts = MAX_EMAIL_CHECK_ATTEMPTS
        ): Promise<boolean> {
            return await notificationsUtils.checkEmailReceivedWithDelay(
                validSubjectHandler,
                flowRootState.users.system_administrator.email,
                attempts,
                EMAIL_CHECK_INTERVAL_MS
            );
        }

        async function getSystemScopedToken(): Promise<TokenResponse> {
            return await authUtils.authenticateSuperAdminAppAndGetToken(
                'default_e2e',
                tenantClientSecret,
                saasConsoleUrl
            );
        }

        async function attachNotificationsServiceLogs(): Promise<void> {
            try {
                logger.info('Attaching notifications service logs for debugging');
                await allureUtils.attachSystemComponentLogs(
                    'notifications-service',
                    'Notifications Service Logs',
                    'runai-backend',
                    1000
                );
            } catch (error) {
                logger.error('Failed to attach notifications service logs:', error);
            }
        }

        async function attachNotificationsProxyLogs(): Promise<void> {
            try {
                logger.info('Attaching notifications proxy logs for debugging');
                await allureUtils.attachSystemComponentLogs(
                    'notifications-proxy',
                    'Notifications Proxy Logs',
                    'runai-backend',
                    1000
                );
            } catch (error) {
                logger.error('Failed to attach notifications proxy logs:', error);
            }
        }

        async function attachWorkloadsHelperLogs(): Promise<void> {
            try {
                logger.info('Attaching workloads helper logs for debugging');
                await allureUtils.attachSystemComponentLogs(
                    'workloads-helper',
                    'Workloads Helper Logs',
                    'runai-backend',
                    1000
                );
            } catch (error) {
                logger.error('Failed to attach workloads helper logs:', error);
            }
        }

        async function attachRelevantK8sLogs(): Promise<void> {
            await attachNotificationsServiceLogs();
            await attachNotificationsProxyLogs();
            await attachWorkloadsHelperLogs();
        }

        async function goToHomePage(): Promise<void> {
            await ctx.browserEngine.currentPage.goto(config.tenantUrl);
        }

        async function configureEmailServerAndResetNotifications(): Promise<void> {
            // Configure email server and enable notifications
            await configSMTPServer();

            // Reset notifications events selection
            await ctx.usersV2Page.clickOnSettings();
            await ctx.usersV2Page.deselectAllNotifications();
            await goToHomePage();
        }

        test.beforeAll(async function () {
            if (settings.isSelfHosted) return;
            systemScopedToken = await getSystemScopedToken();
        });

        test.beforeEach(async function () {
            test.setTimeout(TEST_TIMEOUT_MS);

            // 1. Initialize browser and pages
            ctx = {} as NotificationTestContext;
            ctx.browserEngine = new BrowserEngine();
            await ctx.browserEngine.start();

            ctx.authPage = new AuthPage(ctx.browserEngine.currentPage);
            ctx.generalSettingsPage = new GeneralSettings(ctx.browserEngine.currentPage);
            ctx.usersV2Page = new UsersV2(ctx.browserEngine.currentPage);
            ctx.workloadPage = new WorkloadPage(ctx.browserEngine.currentPage);

            // 2. Authenticate user
            const { email, password } = flowRootState.users.system_administrator;
            await ctx.authPage.signIn(email, password);

            // 3. Initialize project
            await initProject();
            await waitUtils.waitToLoadNewPageOrSaveChanges();
        });

        test.afterEach('disable notifications and stop browser', async ({}, testInfo) => {
            try {
                await disableNotifications();
                await notificationsUtils.restoreSaasEmailChannel();
            } finally {
                const [archiveResult, logsResult] = await Promise.allSettled([
                    ctx.browserEngine.stopAndArchive(testInfo, ctx.videoName),
                    attachRelevantK8sLogs()
                ]);

                if (archiveResult.status === 'rejected') {
                    logger.error('Failed to stop and archive browser:', archiveResult.reason);
                }

                if (logsResult.status === 'rejected') {
                    logger.error('Failed to attach relevant K8s logs:', logsResult.reason);
                }
            }
        });

        shouldRunNativeWorkloadCrdTests(() =>
        test('should send email notification when workload starts running', async ({}, testInfo: TestInfo) => {
            await configureEmailServerAndResetNotifications();
            ctx.videoName = 'notifications-service-should-send-notification';

            await ctx.usersV2Page.clickOnSettings();
            await ctx.usersV2Page.selectAllNotificationsNoDeleting();
            await goToHomePage();

            const { workloadName, workloadId } = await createTrainingWorkload();
            await ctx.workloadPage.goToPage();
            await workloadsUtilsV2.waitForApiWorkloadWithPhase(
                workloadId,
                Phase.Running,
                workloadName,
                true,
                WORKLOAD_WAIT_TIMEOUT_SECONDS
            );

            logger.info(
                `Checking email was sent to user ${flowRootState.users.system_administrator.email} for workload ${workloadName}`
            );
            const emailReceived = await checkEmailReceivedWithDelay((subject: string) =>
                subject.includes(workloadName)
            );

            expect(emailReceived).toBeTruthy();

            // Cleanup
            await workloadsUtilsV2.deleteWorkload(workloadId, WorkloadType.TRAINING);
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test('should send notification only for specific events selected', async ({}, testInfo: TestInfo) => {
            await configureEmailServerAndResetNotifications();
            ctx.videoName = 'should send notification only for specific events selected';

            await ctx.usersV2Page.clickOnSettings();
            await ctx.usersV2Page.selectDeletingStatusNotifications();
            await goToHomePage();

            const { workloadName, workloadId } = await createTrainingWorkload();
            await ctx.workloadPage.goToPage();
            await workloadsUtilsV2.waitForApiWorkloadWithPhase(
                workloadId,
                Phase.Running,
                workloadName,
                true,
                WORKLOAD_WAIT_TIMEOUT_SECONDS
            );
            // Wait for workload to stabilize in Running state before deletion
            // This prevents race conditions where pod status updates from the cluster interfere with deletion
            await waitUtils.sleep(12000);
            await workloadsUtilsV2.deleteWorkload(workloadId, WorkloadType.TRAINING);
            await workloadsUtilsV2.waitForApiWorkloadWithPhase(
                workloadId,
                Phase.Deleting,
                workloadName,
                true,
                WORKLOAD_WAIT_TIMEOUT_SECONDS
            );
            logger.info(`Deleted workload ${workloadName} with id ${workloadId}`);

            logger.info(
                `Checking email was sent to user ${flowRootState.users.system_administrator.email} for workload ${workloadName} with status Deleting`
            );
            const emailReceived = await checkEmailReceivedWithDelay((subject: string) => {
                if (!subject.includes(workloadName)) {
                    return false;
                }
                if (subject.includes('Deleting')) {
                    return true;
                }
                logger.error(
                    `Email received for workload ${workloadName} but includes non requested event: ${subject}`
                );
                throw new Error(
                    `Email received for workload ${workloadName} but includes non requested event`
                );
            });

            expect(emailReceived).toBeTruthy();
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test('should not send email notification when notifications are disabled after being enabled', async ({}, testInfo: TestInfo) => {
            await configureEmailServerAndResetNotifications();
            ctx.videoName =
                'notifications-service-should-not-send-notification-when-notifications-are-disabled-after-being-enabled';

            await disableNotifications();
            const { workloadName, workloadId } = await createTrainingWorkload();
            await ctx.workloadPage.goToPage();
            await workloadsUtilsV2.waitForApiWorkloadWithPhase(
                workloadId,
                Phase.Running,
                workloadName,
                true,
                WORKLOAD_WAIT_TIMEOUT_SECONDS
            );

            logger.info(`
                    Checking email was sent to user ${flowRootState.users.system_administrator.email} for workload ${workloadName}.
                    No email should be sent because notifications are disabled.
                `);
            const emailReceived = await checkEmailReceivedWithDelay(
                (subject: string) => subject.includes(workloadName),
                NO_EMAIL_CHECK_ATTEMPTS
            );

            expect(emailReceived).toBeFalsy();

            // Cleanup
            await workloadsUtilsV2.deleteWorkload(workloadId, WorkloadType.TRAINING);
            })
        );
    }
);
