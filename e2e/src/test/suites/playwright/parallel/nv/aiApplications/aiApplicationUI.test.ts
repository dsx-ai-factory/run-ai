/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, TestInfo } from '@playwright/test';
import {
    flowRootState,
    objectGenerator,
    projectResourcesUtils,
    tags,
    logger,
    operations,
    waitUtils,
    helmReleaseUtils,
    workloadsUtilsV2,
    Utils,
    crds
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import {
    AiApplicationPage,
    AiApplicationSource
} from '../../../../../../modules/ui/runaiV2/pages/AiApplicationPage';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { workloadsService } from '../../../../../../modules/api';
import { LoginUi } from '../../../../../../modules/ui/runaiV2/pages/LoginUiUtils';
import * as yaml from 'js-yaml';
import { User } from '../../../../../../modules/helpers/config/types';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';
import { execSync } from 'child_process';
import * as os from 'os';
import * as path from 'path';

const CUSTOM_URL_OLD_CHART =
    'https://github.com/prometheus-community/helm-charts/releases/download/alertmanager-1.33.0/alertmanager-1.33.0.tgz';

const CUSTOM_URL_CHART =
    'https://github.com/prometheus-community/helm-charts/releases/download/alertmanager-1.33.1/alertmanager-1.33.1.tgz';
const INVALID_CHART_URL = 'https://example.com/invalid-chart.tgz';

// Chart archive used by the "Upload file" source test. Fetched locally once in beforeAll so the UI
// can upload real .tgz bytes rather than reference a URL.
const UPLOAD_CHART_SOURCE_URL = CUSTOM_URL_CHART;

interface CreateAndDeleteTestCase {
    testName: string;
    videoName: string;
    user: User;
    source: AiApplicationSource;
    repository?: string;
    chartName?: string;
    chartVersion?: string;
    chartUrl?: string;
    overrides?: string;
}

test.describe(`AI Application UI E2E Tests, tags: ${tags.tool.ui} ${tags.resource.aiApplications} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_25)(function () {
        test.setTimeout(5 * 60 * 1000);

        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let aiApplicationPage: AiApplicationPage;
        let loginUIPage: LoginUi;
        let videoName: string;
        let appName: string;

        // Shared resources
        let projectNameList: string[];
        let namespaceList: string[];
        let projectIdList: number[];

        // Per-test resource tracking for cleanup
        let aiApplicationIds: string[];
        let helmReleasesToCleanup: { name: string; namespace: string }[];

        // Local .tgz used by the "Upload file" source test.
        let uploadedChartFilePath: string;

        test.beforeAll(async () => {
            browserEngine = new BrowserEngine();

            projectNameList = Array.from({ length: 2 }, () =>
                objectGenerator.generateShortNameByPrefix('ai-app', 5)
            );
            namespaceList = projectNameList.map((name) => Utils.namespaceFromProjectName(name));

            const projectsToCreate = [
                { namespace: namespaceList[0], name: projectNameList[0] },
                { namespace: namespaceList[1], name: projectNameList[1] }
            ];
            const projResponse = await Promise.all(
                projectsToCreate.map((config) =>
                    projectResourcesUtils.createProject(config.namespace, {
                        name: config.name,
                        deservedGpus: 8,
                        maxAllowedGpus: -1
                    })
                )
            );
            projectIdList = projResponse.map((proj) => proj.id);

            uploadedChartFilePath = path.join(os.tmpdir(), `ai-app-upload-${Date.now()}.tgz`);
            execSync(`curl -sSL -o "${uploadedChartFilePath}" "${UPLOAD_CHART_SOURCE_URL}"`);
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            aiApplicationPage = new AiApplicationPage(browserEngine.currentPage);
            loginUIPage = new LoginUi(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );

            appName = objectGenerator.generateShortNameByPrefix('ai-app', 5);
            aiApplicationIds = [];
            helmReleasesToCleanup = [];
        });

        test.afterEach('stop browser and cleanup', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
            workloadsService.aiApplicationsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;

            // Delete AI applications via API
            await Promise.all(
                aiApplicationIds.map(async (appId) => {
                    try {
                        await workloadsService.aiApplicationsApi.deleteApplication(appId);
                    } catch (e) {
                        logger.error(`Failed to delete AI application ${appId}`, e);
                    }
                })
            );

            // Uninstall helm releases
            await Promise.all(
                helmReleasesToCleanup.map(async (release) => {
                    try {
                        await helmReleaseUtils.helmUninstall(release.name, release.namespace);
                    } catch (e) {
                        logger.error(
                            `Failed to uninstall helm release ${release.name} from ${release.namespace}`,
                            e
                        );
                    }
                })
            );
        });

        test.afterAll('cleanup', async () => {
            await Promise.all(
                projectIdList.map(async (projectId, i) => {
                    try {
                        await operations.deleteProjectAndNamespace(projectId);
                        logger.info(
                            `deleted project ${projectNameList[i]} and namespace ${namespaceList[i]}`
                        );
                    } catch (error) {
                        logger.error(
                            `failed to delete project ${projectNameList[i]} and namespace ${namespaceList[i]}`,
                            error
                        );
                    }
                })
            );
        });

        test.describe('Create and delete AI application tests', () => {
            const testCases: CreateAndDeleteTestCase[] = [
                {
                    testName: 'T5857335: Create and delete an NGC AI application using System administrator',
                    videoName: 'create-delete-ngc-ai-app-sys-admin',
                    user: flowRootState.users.system_administrator,
                    source: AiApplicationSource.NgcCatalog,
                    repository: 'Nvidia Blueprints',
                    chartName: 'Nvidia-blueprint-rag',
                    chartVersion: 'v2.3.0'
                },
                {
                    testName: 'T5857336: Create and delete an NGC AI application using AI practitioner',
                    videoName: 'create-delete-ngc-ai-app-ai-practitioner',
                    user: flowRootState.users.ai_practitioner,
                    source: AiApplicationSource.NgcCatalog,
                    repository: 'Nvidia Blueprints',
                    chartName: 'Nvidia-blueprint-rag',
                    chartVersion: 'v2.3.0'
                },
                {
                    testName:
                        'T5857337: Create and delete a Non-NGC(Custom URL) AI application using System administrator',
                    videoName: 'create-delete-custom-url-ai-app-sys-admin',
                    user: flowRootState.users.system_administrator,
                    source: AiApplicationSource.CustomUrl,
                    chartUrl: CUSTOM_URL_CHART
                },
                {
                    testName: 'T5857339: Create AI Application with custom values override',
                    videoName: 'create-delete-ai-app-with-overrides',
                    user: flowRootState.users.system_administrator,
                    source: AiApplicationSource.CustomUrl,
                    chartUrl: CUSTOM_URL_CHART,
                    overrides: 'replicaCount: 2'
                },
                {
                    // TODO: replace with the real TestRail case id once allocated.
                    testName: 'T5860001: Create and delete an AI application from an uploaded chart file',
                    videoName: 'create-delete-upload-file-ai-app-sys-admin',
                    user: flowRootState.users.system_administrator,
                    source: AiApplicationSource.UploadFile
                }
            ];

            for (const [index, tc] of testCases.entries()) {
                test(tc.testName, async () => {
                    videoName = tc.videoName;
                    const projectName = projectNameList[index % projectNameList.length];
                    const namespace = namespaceList[index % namespaceList.length];

                    if (tc.user !== flowRootState.users.system_administrator) {
                        await loginUIPage.logoutToLoginPage();
                        await authPage.signIn(tc.user.email, tc.user.password);
                    }

                    await test.step('Navigate to AI Applications page', async () => {
                        await aiApplicationPage.goToPage();
                    });

                    await test.step(`Create AI application ${appName} from ${tc.source}`, async () => {
                        if (tc.source === AiApplicationSource.NgcCatalog) {
                            await aiApplicationPage.createAiApplicationFromNgc(
                                appName,
                                projectName,
                                tc.repository,
                                tc.chartName,
                                tc.chartVersion,
                                tc.overrides
                            );
                        } else if (tc.source === AiApplicationSource.UploadFile) {
                            await aiApplicationPage.createAiApplicationFromUploadFile(
                                appName,
                                projectName,
                                uploadedChartFilePath,
                                tc.overrides
                            );
                        } else {
                            await aiApplicationPage.createAiApplicationFromCustomUrl(
                                appName,
                                projectName,
                                tc.chartUrl,
                                tc.overrides
                            );
                        }
                    });

                    await test.step('Verify the application appears in the grid with correct details', async () => {
                        const appId = await waitUtils.waitForAiApplicationToAppear(appName);
                        aiApplicationIds.push(appId);

                        const appData = await aiApplicationPage.getAiApplicationByNameFromGrid(appName);
                        expect(appData).toBeDefined();
                        expect(appData['AI application']).toBe(appName);
                        expect(appData['Project']).toBe(projectName);
                    });

                    await test.step('Verify the application CR is created in the cluster', async () => {
                        await waitUtils.waitForApplicationCrdToAppear(appName, namespace);

                        if (tc.source !== AiApplicationSource.NgcCatalog) {
                            await waitUtils.waitForApplicationCRStatus(
                                appName,
                                namespace,
                                'completed',
                                'deployed'
                            );
                        }
                    });

                    if (tc.overrides) {
                        await test.step('Verify the overrides are applied in the CR', async () => {
                            const cr = await kubeCli.getCrd(crds.applicationsCrd, appName, namespace);
                            const crOverrides = (cr as any)?.spec?.helm?.values;
                            expect(crOverrides).toBeDefined();
                            const expected = yaml.load(tc.overrides) as Record<string, any>;
                            expect(crOverrides).toMatchObject(expected);
                        });
                    }

                    await test.step('Delete the application from UI', async () => {
                        await aiApplicationPage.deleteAiApplication(appName);
                        aiApplicationIds.pop();
                    });

                    await test.step('Verify the application disappears from the grid', async () => {
                        const apps = await aiApplicationPage.getAiApplicationsFromGrid();
                        const deletedApp = apps.find((a) => a['AI application'] === appName);
                        expect(deletedApp).toBeUndefined();
                    });

                    await test.step('Verify the application CR disappears from the cluster', async () => {
                        await kubeCli.waitForCrdToBeDeleted(crds.applicationsCrd, appName, namespace);
                    });
                });
            }

            test('T5857338: Verify workloads from cluster-installed chart are associated with AI Application and cannot be deleted from UI', async () => {
                videoName = 'helm-installed-app-no-ui-delete';
                const namespace = namespaceList[0];
                const helmReleaseName = objectGenerator.generateShortNameByPrefix('helm-app', 5);

                await test.step(`Helm install ${helmReleaseName} into ${namespace}`, async () => {
                    await helmReleaseUtils.helmInstall(helmReleaseName, namespace, CUSTOM_URL_CHART);
                    helmReleasesToCleanup.push({ name: helmReleaseName, namespace });
                });

                await test.step('Verify the application CR is created', async () => {
                    await waitUtils.waitForApplicationCrdToAppear(helmReleaseName, namespace);
                });

                await test.step('Verify the application appears in AI Applications grid', async () => {
                    await aiApplicationPage.goToPage();
                    await waitUtils.waitForCondition(
                        async () => {
                            const appData =
                                await aiApplicationPage.getAiApplicationByNameFromGrid(helmReleaseName);
                            return appData !== undefined;
                        },
                        `Wait for helm app ${helmReleaseName} to appear in grid`,
                        true,
                        60,
                        5
                    );
                });

                await test.step('Verify the application cannot be deleted from UI', async () => {
                    await aiApplicationPage.toggleAiApplication(helmReleaseName);
                    await expect(aiApplicationPage.getDeleteBtn()).toBeDisabled({ timeout: 1000 });
                });

                await test.step(`Helm uninstall ${helmReleaseName}`, async () => {
                    await helmReleaseUtils.helmUninstall(helmReleaseName, namespace);
                    helmReleasesToCleanup.pop();
                });
            });

            test('T5857341: Deploy a chart that fails and verify Failed status with error message', async () => {
                videoName = 'deploy-failing-chart';
                const projectName = projectNameList[0];
                const namespace = namespaceList[0];

                await test.step('Navigate to AI Applications page', async () => {
                    await aiApplicationPage.goToPage();
                });

                await test.step('Create an AI application with an invalid chart URL', async () => {
                    await aiApplicationPage.createAiApplicationFromCustomUrl(
                        appName,
                        projectName,
                        INVALID_CHART_URL
                    );

                    const appId = await waitUtils.waitForAiApplicationToAppear(appName);
                    aiApplicationIds.push(appId);
                });

                await test.step('Verify status shows Failed', async () => {
                    await waitUtils.waitForCondition(
                        async () => {
                            const appData = await aiApplicationPage.getAiApplicationByNameFromGrid(appName);
                            return appData?.['Status'].includes('Failed');
                        },
                        `Wait for AI application ${appName} status to become Failed`,
                        true,
                        120,
                        5
                    );
                });

                await test.step('Verify the application CR status contains error details', async () => {
                    const cr = await kubeCli.getCrd(crds.applicationsCrd, appName, namespace);
                    const actionError = (cr as any)?.status?.action?.error;
                    expect(actionError).toBeDefined();
                    expect(actionError).toContain('failed');
                });
            });

            test('T5857346: Helm install and UI install AI applications coexist in the same project', async () => {
                videoName = 'helm-and-ui-apps-coexist';
                const projectName = projectNameList[0];
                const namespace = namespaceList[0];
                const helmReleaseName = objectGenerator.generateShortNameByPrefix('cluster-app', 5);
                const uiAppName = appName;

                await test.step(`Helm install ${helmReleaseName} into ${namespace}`, async () => {
                    await helmReleaseUtils.helmInstall(helmReleaseName, namespace, CUSTOM_URL_CHART);
                    helmReleasesToCleanup.push({ name: helmReleaseName, namespace });
                });

                await test.step(`Create ${uiAppName} via Run:ai platform in the same project`, async () => {
                    await aiApplicationPage.goToPage();
                    await aiApplicationPage.createAiApplicationFromCustomUrl(
                        uiAppName,
                        projectName,
                        CUSTOM_URL_CHART
                    );
                    const appId = await waitUtils.waitForAiApplicationToAppear(uiAppName);
                    aiApplicationIds.push(appId);
                });

                await test.step('Verify both applications appear in the AI Applications grid and UI app is deployed', async () => {
                    await waitUtils.waitForCondition(
                        async () => {
                            const apps = await aiApplicationPage.getAiApplicationsFromGrid();
                            const helmApp = apps.find((a) => a['AI application'] === helmReleaseName);
                            const uiApp = apps.find((a) => a['AI application'] === uiAppName);
                            return (
                                helmApp !== undefined && uiApp !== undefined && uiApp['Status'] !== 'Creating'
                            );
                        },
                        `Wait for both apps to appear and ${uiAppName} to finish creating`,
                        true,
                        120,
                        5
                    );
                });

                await test.step('Verify UI-created app has delete action available', async () => {
                    await aiApplicationPage.toggleAiApplication(uiAppName);
                    await expect(aiApplicationPage.getDeleteBtn()).toBeEnabled({ timeout: 3000 });
                    // Deselect
                    await aiApplicationPage.toggleAiApplication(uiAppName);
                });

                await test.step('Verify helm-installed app does NOT have delete action', async () => {
                    await aiApplicationPage.toggleAiApplication(helmReleaseName);
                    await expect(aiApplicationPage.getDeleteBtn()).toBeDisabled({ timeout: 1000 });
                });

                await test.step('Clean up UI-created app', async () => {
                    await aiApplicationPage.goToPage();
                    await aiApplicationPage.deleteAiApplication(uiAppName);
                    aiApplicationIds.pop();
                });

                await test.step(`Helm uninstall ${helmReleaseName}`, async () => {
                    await helmReleaseUtils.helmUninstall(helmReleaseName, namespace);
                    helmReleasesToCleanup.pop();
                });
            });

            test('T5858800: Helm install and upgrade AI application successfully', async () => {
                videoName = 'helm-install-upgrade-success';
                const namespace = namespaceList[0];
                const helmReleaseName = objectGenerator.generateShortNameByPrefix('helm-upg', 5);

                await test.step(`Helm install ${helmReleaseName}`, async () => {
                    const stdout = await helmReleaseUtils.helmInstall(
                        helmReleaseName,
                        namespace,
                        CUSTOM_URL_OLD_CHART
                    );
                    helmReleasesToCleanup.push({ name: helmReleaseName, namespace });
                    expect(stdout).toContain('STATUS: deployed');
                });

                await test.step(`Helm upgrade ${helmReleaseName}`, async () => {
                    const stdout = await helmReleaseUtils.helmUpgrade(
                        helmReleaseName,
                        namespace,
                        CUSTOM_URL_CHART
                    );
                    expect(stdout).toContain('STATUS: deployed');
                });

                await test.step('Verify upgrade CR k8sPhase is deployed', async () => {
                    await waitUtils.waitForApplicationCRK8sPhase(helmReleaseName, namespace, 'deployed');
                });

                await test.step('Verify AI application status in UI', async () => {
                    await aiApplicationPage.goToPage();
                    await waitUtils.waitForCondition(
                        async () => {
                            const appData =
                                await aiApplicationPage.getAiApplicationByNameFromGrid(helmReleaseName);
                            return appData?.['Status'].includes('Deployed');
                        },
                        `Wait for ${helmReleaseName} to show Deployed in UI`,
                        true,
                        60,
                        5
                    );
                });

                await test.step(`Helm uninstall ${helmReleaseName}`, async () => {
                    await helmReleaseUtils.helmUninstall(helmReleaseName, namespace);
                    helmReleasesToCleanup.pop();
                });
            });

            test('T5858801: Helm install and upgrade AI application with failure', async () => {
                videoName = 'helm-install-upgrade-fail';
                const namespace = namespaceList[0];
                const helmReleaseName = objectGenerator.generateShortNameByPrefix('helm-fail', 5);

                await test.step(`Helm install ${helmReleaseName}`, async () => {
                    const stdout = await helmReleaseUtils.helmInstall(
                        helmReleaseName,
                        namespace,
                        CUSTOM_URL_CHART
                    );
                    helmReleasesToCleanup.push({ name: helmReleaseName, namespace });
                    expect(stdout).toContain('STATUS: deployed');
                });

                await test.step(`Helm upgrade ${helmReleaseName} with invalid image (expect failure)`, () => {
                    // --wait (not --atomic): --atomic rolls a failed upgrade back to the previous healthy
                    // revision, so the release returns to "deployed" and the CR k8sPhase never reaches
                    // "failed". --wait still fails the upgrade on the bad image but leaves it failed.
                    const stderr = helmReleaseUtils.helmUpgradeExpectFailure(
                        helmReleaseName,
                        namespace,
                        CUSTOM_URL_CHART,
                        '--set image.repository=nonexistent --set image.tag=fake --wait --timeout 5s'
                    );
                    expect(stderr).toBeDefined();
                    logger.info(`Helm upgrade failed as expected: ${stderr}`);
                });

                await test.step('Verify CR k8sPhase reflects failure', async () => {
                    await waitUtils.waitForApplicationCRK8sPhase(helmReleaseName, namespace, 'failed', 60);
                });

                await test.step('Verify AI application shows failure status in UI', async () => {
                    await aiApplicationPage.goToPage();
                    await waitUtils.waitForCondition(
                        async () => {
                            const appData =
                                await aiApplicationPage.getAiApplicationByNameFromGrid(helmReleaseName);
                            return appData?.['Status'].includes('Failed');
                        },
                        `Wait for ${helmReleaseName} to show Failed in UI`,
                        true,
                        60,
                        5
                    );
                });

                await test.step(`Helm uninstall ${helmReleaseName}`, async () => {
                    await helmReleaseUtils.helmUninstall(helmReleaseName, namespace);
                    helmReleasesToCleanup.pop();
                });
            });

            test('T5857348: Create AI Application with duplicate name in the same project', async () => {
                videoName = 'duplicate-name-same-project';
                const projectName = projectNameList[0];

                await test.step('Navigate to AI Applications page', async () => {
                    await aiApplicationPage.goToPage();
                });

                await test.step(`Create AI application ${appName}`, async () => {
                    await aiApplicationPage.createAiApplicationFromCustomUrl(
                        appName,
                        projectName,
                        CUSTOM_URL_CHART
                    );
                    const appId = await waitUtils.waitForAiApplicationToAppear(appName);
                    aiApplicationIds.push(appId);

                    await aiApplicationPage.waitForNotificationDisappear();
                });

                await test.step('Create another AI application with the same name in the same project - should fail', async () => {
                    await aiApplicationPage.goToPage();
                    await aiApplicationPage.createAiApplicationFromCustomUrl(
                        appName,
                        projectName,
                        CUSTOM_URL_CHART
                    );
                    const msg = await aiApplicationPage.getNotificationMsg();
                    expect(msg.toLowerCase()).toContain('failed to create ai application');
                    await aiApplicationPage.clickCancelBtn();
                });

                await test.step('Delete the AI application', async () => {
                    await aiApplicationPage.deleteAiApplication(appName);
                    aiApplicationIds.pop();
                });
            });

            test('T5857349: Create AI Application with same name in different projects', async () => {
                videoName = 'same-name-different-projects';
                const projectName1 = projectNameList[0];
                const projectName2 = projectNameList[1];

                await test.step(`Create AI application ${appName} in ${projectName1}`, async () => {
                    await aiApplicationPage.goToPage();
                    await aiApplicationPage.createAiApplicationFromCustomUrl(
                        appName,
                        projectName1,
                        CUSTOM_URL_CHART
                    );
                    const appId = await waitUtils.waitForAiApplicationToAppear(appName);
                    aiApplicationIds.push(appId);
                });

                await test.step(`Create AI application with same name ${appName} in ${projectName2} - should succeed`, async () => {
                    await aiApplicationPage.goToPage();
                    await aiApplicationPage.createAiApplicationFromCustomUrl(
                        appName,
                        projectName2,
                        CUSTOM_URL_CHART
                    );
                    // Wait for second app to appear — list by project to find it
                    await waitUtils.waitForCondition(
                        async () => {
                            const apps = await aiApplicationPage.getAiApplicationsFromGrid();
                            const matchingApps = apps.filter((a) => a['AI application'] === appName);
                            return matchingApps.length === 2;
                        },
                        `Wait for two AI applications named ${appName} to appear in grid`,
                        true,
                        60,
                        5
                    );
                });

                await test.step('Verify both applications exist with correct projects', async () => {
                    const apps = await aiApplicationPage.getAiApplicationsFromGrid();
                    const matchingApps = apps.filter((a) => a['AI application'] === appName);
                    expect(matchingApps.length).toBe(2);

                    const projects = matchingApps.map((a) => a['Project']).sort();
                    expect(projects).toEqual([projectName1, projectName2].sort());
                });

                await test.step('Track second application for cleanup', async () => {
                    const listResponse = await workloadsUtilsV2.listApplications();
                    const matchingApps = listResponse.body.applications?.filter(
                        (a: any) => a.name === appName
                    );
                    const secondAppId = matchingApps?.find((a: any) => a.id !== aiApplicationIds[0])?.id;
                    if (secondAppId) {
                        aiApplicationIds.push(secondAppId);
                    }
                });
            });
        });

        test.describe('Miscellaneous AI application tests', () => {
            // Pre-created apps: 2 in project 1 (1 valid + 1 invalid), 1 in project 2
            let validApp1Name: string;
            let validApp2Name: string;
            let failedAppName: string;
            const miscAppIds: string[] = [];

            test.beforeAll(async () => {
                validApp1Name = objectGenerator.generateShortNameByPrefix('misc-valid', 5);
                failedAppName = objectGenerator.generateShortNameByPrefix('misc-fail', 5);
                validApp2Name = objectGenerator.generateShortNameByPrefix('misc-ok', 5);

                // Create 3 apps via API-triggered UI flow using workloads service
                workloadsService.aiApplicationsApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;

                const [resp1, resp2, resp3] = await Promise.all([
                    workloadsService.aiApplicationsApi.createApplication({
                        name: validApp1Name,
                        projectId: projectIdList[0].toString(),
                        clusterId: flowRootState.cluster.uuid,
                        spec: { chartRef: { chartURL: CUSTOM_URL_CHART } }
                    }),
                    workloadsService.aiApplicationsApi.createApplication({
                        name: failedAppName,
                        projectId: projectIdList[0].toString(),
                        clusterId: flowRootState.cluster.uuid,
                        spec: { chartRef: { chartURL: INVALID_CHART_URL } }
                    }),
                    workloadsService.aiApplicationsApi.createApplication({
                        name: validApp2Name,
                        projectId: projectIdList[1].toString(),
                        clusterId: flowRootState.cluster.uuid,
                        spec: { chartRef: { chartURL: CUSTOM_URL_CHART } }
                    })
                ]);
                miscAppIds.push(resp1.body.id, resp2.body.id, resp3.body.id);

                // Wait for all apps to reach expected status in parallel
                await Promise.all([
                    waitUtils.waitForApplicationCRStatus(
                        validApp1Name,
                        namespaceList[0],
                        'completed',
                        'deployed',
                        120
                    ),
                    waitUtils.waitForApplicationCRStatus(
                        validApp2Name,
                        namespaceList[1],
                        'completed',
                        'deployed',
                        120
                    ),
                    waitUtils.waitForApplicationCRStatus(
                        failedAppName,
                        namespaceList[0],
                        'failed',
                        undefined,
                        120
                    )
                ]);
            });

            test.afterAll(async () => {
                workloadsService.aiApplicationsApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;
                await Promise.all(
                    miscAppIds.map(async (appId) => {
                        try {
                            await workloadsService.aiApplicationsApi.deleteApplication(appId);
                            logger.info(`Deleted misc AI application ${appId}`);
                        } catch (e) {
                            logger.error(`Failed to delete misc AI application ${appId}`, e);
                        }
                    })
                );
            });

            test('T5857340: View deployed AI Application details in drawer', async () => {
                videoName = 'view-ai-app-details-drawer';

                await test.step('Navigate to AI Applications page', async () => {
                    await aiApplicationPage.goToPage();
                });

                await test.step('Click the deployed application row and open details drawer', async () => {
                    await aiApplicationPage.toggleAiApplication(validApp1Name);
                    await aiApplicationPage.clickShowDetailsBtn();
                });

                await test.step('Verify details pane is visible', async () => {
                    await expect(aiApplicationPage.getDetailsPane()).toBeVisible({ timeout: 5000 });
                });

                await test.step('Verify workload cards show Running status', async () => {
                    const workloadCards = aiApplicationPage.getWorkloadCards();
                    const count = await workloadCards.count();
                    expect(count).toBeGreaterThan(0);

                    await Promise.all(
                        Array.from({ length: count }, (_, i) =>
                            expect(aiApplicationPage.getWorkloadCardStatus(workloadCards.nth(i))).toHaveText(
                                'Running',
                                { timeout: 60000 }
                            )
                        )
                    );
                    logger.info(`Verified ${count} workload cards with Running status`);
                });
            });

            test('T5857342: AI Applications grid - filter', async () => {
                videoName = 'ai-app-grid-filter';

                await test.step('Navigate to AI Applications page', async () => {
                    await aiApplicationPage.goToPage();
                });

                await test.step('Filter by AI application name contains partial name', async () => {
                    const partialName = validApp1Name.substring(0, 10);
                    await aiApplicationPage.addFilter('AI application', partialName, 'contains');

                    try {
                        const apps = await aiApplicationPage.getAiApplicationsFromGrid();
                        const matchingApps = apps.filter((a) => a['AI application'].includes(partialName));
                        expect(matchingApps.length).toBeGreaterThan(0);
                        for (const app of apps) {
                            expect(app['AI application']).toContain(partialName);
                        }
                    } finally {
                        await aiApplicationPage.closeFilter();
                    }
                });

                await test.step('Apply project filter', async () => {
                    await aiApplicationPage.addFilter('Project', projectNameList[0], 'equals');

                    try {
                        const apps = await aiApplicationPage.getAiApplicationsFromGrid();
                        for (const app of apps) {
                            expect(app['Project']).toBe(projectNameList[0]);
                        }
                    } finally {
                        await aiApplicationPage.closeFilter();
                    }
                });
            });

            test('T5857343: AI Applications grid - search', async ({}, testInfo: TestInfo) => {
                skipPlaywrightTest(test, testInfo, 'the search function is not working');

                videoName = 'ai-app-grid-search';

                await test.step('Navigate to AI Applications page', async () => {
                    await aiApplicationPage.goToPage();
                });

                await test.step('Search for a specific application name', async () => {
                    await aiApplicationPage.searchByName(validApp2Name);

                    await waitUtils.waitForCondition(
                        async () => {
                            const apps = await aiApplicationPage.getAiApplicationsFromGrid();
                            return apps.length === 1 && apps[0]['AI application'] === validApp2Name;
                        },
                        `Wait for search to return only ${validApp2Name}`,
                        true,
                        10,
                        2
                    );
                });
            });

            test('T5857344: AI Applications grid - sort', async ({}, testInfo: TestInfo) => {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'descending sort by ai application column is not working'
                );

                videoName = 'ai-app-grid-sort';

                await test.step('Navigate to AI Applications page', async () => {
                    await aiApplicationPage.goToPage();
                });

                await test.step('Sort by AI application column and verify order', async () => {
                    // Click the AI application column header to sort
                    const columnHeader = aiApplicationPage.page.locator('th:has-text("AI application")');
                    await columnHeader.click();
                    await aiApplicationPage.page.waitForTimeout(1000);

                    const apps = await aiApplicationPage.getAiApplicationsFromGrid();
                    const appNames = apps.map((a) => a['AI application']);
                    const sortedAsc = [...appNames].sort((a, b) => a.localeCompare(b));
                    expect(appNames).toEqual(sortedAsc);

                    // Click again to reverse sort
                    await columnHeader.click();
                    await aiApplicationPage.page.waitForTimeout(1000);

                    const appsDesc = await aiApplicationPage.getAiApplicationsFromGrid();
                    const appNamesDesc = appsDesc.map((a) => a['AI application']);
                    const sortedDesc = [...appNamesDesc].sort((a, b) => b.localeCompare(a));
                    expect(appNamesDesc).toEqual(sortedDesc);
                });
            });

            test('T5857345: AI Applications grid - column customization', async ({}, testInfo: TestInfo) => {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'the department column displays - for all ai applications'
                );

                videoName = 'ai-app-grid-column-customization';

                await test.step('Navigate to AI Applications page', async () => {
                    await aiApplicationPage.goToPage();
                });

                await test.step('Enable the Department column', async () => {
                    await aiApplicationPage.setColumn('Department', true);
                });

                await test.step('Verify the Department column shows correct value', async () => {
                    const appData = await aiApplicationPage.getAiApplicationByNameFromGrid(validApp1Name);
                    expect(appData).toBeDefined();
                    expect(appData['Department']).toBe('default');
                });

                await test.step('Disable the Department column', async () => {
                    await aiApplicationPage.setColumn('Department', false);
                });
            });

            test('T5857347: Export AI Applications list as CSV', async ({}, testInfo: TestInfo) => {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'app statused in downloaded csv file differ from the statuses in the UI'
                );

                videoName = 'ai-app-export-csv';

                await test.step('Navigate to AI Applications page', async () => {
                    await aiApplicationPage.goToPage();
                });

                let gridData: Array<Record<string, string>>;
                await test.step('Get current grid data', async () => {
                    gridData = await aiApplicationPage.getAiApplicationsFromGrid();
                    expect(gridData.length).toBeGreaterThan(0);
                });

                let csvData: Array<Record<string, string>>;
                await test.step('Download CSV file', async () => {
                    const csvFilePath = await aiApplicationPage.downloadCSV('ai-applications-export.csv');
                    csvData = Utils.readCSVFile(csvFilePath);
                    logger.info(`CSV data: ${JSON.stringify(csvData, null, 2)}`);
                });

                await test.step('Verify CSV data matches grid data', () => {
                    expect(csvData.length).toBeGreaterThanOrEqual(gridData.length);

                    // Verify each app from the grid exists in CSV
                    for (const gridRow of gridData) {
                        const appName = gridRow['AI application'];
                        if (!appName) continue;
                        const csvRow = csvData.find((row) => row['AI application'] === appName);
                        expect(csvRow).toBeDefined();
                        // Verify key fields match
                        expect(csvRow['AI application']).toBe(gridRow['AI application']);
                        expect(csvRow['Status']).toBe(gridRow['Status']);
                        expect(csvRow['Project']).toBe(gridRow['Project']);
                        expect(csvRow['Type']).toBe(gridRow['Type']);
                    }
                });
            });
        });
    });
});
