/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils,
    usersManager,
    idpUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { AdminOnboardingPage } from '../../../../../../modules/ui/runaiV2/pages/AdminOnboardingPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { test, expect } from '@playwright/test';
import { Keycloak } from '../../../../../../modules/helpers/utils/Keycloak';
import { realmName } from '../../../../../../modules/helpers/config/const';
import { readFileSync } from 'fs';

test.describe(`Admin Onboarding UI ${tags.teams.loong} ${tags.tool.ui} `, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let project: Project;
        let projectName: string;
        let newProjectName: string;
        let namespace: string;
        let browserEngine: BrowserEngine;
        let adminOnboardingPage: AdminOnboardingPage;
        let authPage: AuthPage;
        let projectPage: ProjectPage;
        let videoName: string;
        const unreachableIdp = new URL(config.tenantUrl);
        unreachableIdp.hostname = `invalid-${unreachableIdp.hostname}`;
        const invalidSamlUrl = `${unreachableIdp.origin}/auth/realms/${realmName}/protocol/saml/descriptor`;
        const invalidOidcUrl = `${unreachableIdp.origin}/auth/realms/workAI/.well-known/openid-configuration`;
        const samlUrl = Keycloak.getConfigurationProtocolEndpointSAMLUrl(realmName);
        const oidcClientId = 'invalid-client-id';
        const oidcClientSecret = 'invalid-secret';

        const deleteAllIdps = async (): Promise<void> => {
            try {
                const idps = await idpUtils.getIdps(flowRootState.users.system_administrator.accessToken);
                await Promise.all(
                    idps.map(({ alias }) =>
                        idpUtils
                            .deleteIdp(flowRootState.users.system_administrator.accessToken, alias)
                            .catch((err) => {
                                logger.warn(`Failed to delete IDP ${alias}:`, err);
                            })
                    )
                );
            } catch (err) {
                logger.warn('Failed to list IDPs:', err);
            }
        };

        test.beforeEach('delete idps', async function () {
            try {
                await operations.deleteAllProjects(flowRootState.cluster.uuid);
            } catch (error) {
                logger.warn('Failed to delete all projects:', error);
            }
            await deleteAllIdps();
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
            await deleteAllIdps();
        });

        test.describe('Test different types of negative scenarios', function () {
            test.beforeEach(async function () {
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                await browserEngine.currentPage.addInitScript(() => {
                    try {
                        localStorage.setItem('admin-onboarding-modal-shown', 'false');
                    } catch (e) {
                        logger.warn('Failed to preset admin-onboarding-modal-shown in init script', e);
                    }
                });
                authPage = new AuthPage(browserEngine.currentPage);
                await browserEngine.currentPage.goto(config.tenantUrl);
                await authPage.signInValidateLoggedIn(
                    flowRootState.users.system_administrator.email,
                    flowRootState.users.system_administrator.password
                );

                // Wait for page to fully load after authentication
                await browserEngine.currentPage.waitForLoadState('networkidle', { timeout: 30000 });
                await browserEngine.currentPage.waitForTimeout(2000);

                adminOnboardingPage = new AdminOnboardingPage(browserEngine.currentPage);
                projectPage = new ProjectPage(browserEngine.currentPage);
                projectName = objectGenerator.generateNameByPrefix('e2e-project');
            });

            test('T5404938: SSO IDP Creation Failure -- saml', async function () {
                const teamMemberEmail = `runaitest${Utils.generateShortUuid().slice(0, 5)}@test.com`;
                videoName = 'saml-idp-creation-failure';
                await test.step('Click let go button', async () => {
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Config saml sso(url) setup', async () => {
                    await adminOnboardingPage.configSamlSsoURL(invalidSamlUrl);
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Review sso config page', async () => {
                    await adminOnboardingPage.clickSkipSsoSetupButton();
                });

                await test.step('Skip email server setup', async () => {
                    await adminOnboardingPage.clickSkipButton();
                });

                await test.step('Config team setup', async () => {
                    await adminOnboardingPage.enterTeamName(projectName);
                    await adminOnboardingPage.enterTeamMemberEmail(teamMemberEmail);
                    await adminOnboardingPage.clickContinueButton();
                });
                await test.step('Download credentials', async () => {
                    const credentialsFilePath = await adminOnboardingPage.downloadCredentials();
                    const contents = readFileSync(credentialsFilePath, 'utf8');
                    expect(contents && contents.includes(teamMemberEmail)).toBe(true);
                });

                await test.step('Check team setup success', async () => {
                    await adminOnboardingPage.clickContinueButton();
                    const isSuccessful = await adminOnboardingPage.isSetupSuccessful();
                    expect(isSuccessful).toBe(true);
                });

                await test.step('Click continue and it should land on project page', async () => {
                    await adminOnboardingPage.clickContinueButton();
                    await waitUtils.waitForCondition(
                        async () => {
                            const names = await projectPage.getProjectNames();
                            return names.length === 1 && names[0] === projectName;
                        },
                        'Wait for exactly one project with the expected name',
                        true,
                        5,
                        1
                    );
                });
            });

            test('T5404940: project creation failure', async function () {
                videoName = 'project-creation-failure';
                newProjectName = `new-project-${Utils.generateShortUuid()}`;
                const teamMemberEmail = `runaitest${Utils.generateShortUuid().slice(0, 5)}@test.com`;
                await test.step('Click let go button', async () => {
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Skip sso setup', async () => {
                    await adminOnboardingPage.clickSkipButton();
                });

                await test.step('Skip email server setup', async () => {
                    await adminOnboardingPage.clickSkipButton();
                });

                await test.step('Create a project by api', async () => {
                    namespace = Utils.namespaceFromProjectName(projectName);
                    project = await projectResourcesUtils.createProject(namespace, {
                        name: projectName,
                        deservedGpus: 0,
                        maxAllowedGpus: 0
                    });
                });
                await test.step('Config team setup', async () => {
                    await adminOnboardingPage.enterTeamName(projectName);
                    await adminOnboardingPage.enterTeamMemberEmail(teamMemberEmail);
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Check project creation failure', async () => {
                    await adminOnboardingPage.clickCreatingTeamAgainSpan();
                });

                await test.step('Config team setup again', async () => {
                    await adminOnboardingPage.enterTeamName(newProjectName);
                    await adminOnboardingPage.enterTeamMemberEmail(teamMemberEmail);
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Click continue and it should land on project page', async () => {
                    await adminOnboardingPage.clickContinueButton();
                    await adminOnboardingPage.clickContinueButton();
                    await waitUtils.waitForCondition(
                        async () => {
                            const names = await projectPage.getProjectNames();
                            return names.length === 1 && names[0] === newProjectName;
                        },
                        'Wait for exactly one project with the expected name',
                        true,
                        5,
                        1
                    );
                });
            });

            test('T5404941: user creation failure', async function () {
                videoName = 'user-creation-failure';
                const teamMemberEmail = `runaitest${Utils.generateShortUuid().slice(0, 5)}@test.com`;
                await test.step('Click let go button', async () => {
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Skip sso setup', async () => {
                    await adminOnboardingPage.clickSkipButton();
                });

                await test.step('Skip email server setup', async () => {
                    await adminOnboardingPage.clickSkipButton();
                });

                await test.step('create a user by api', async () => {
                    await usersManager.createLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        teamMemberEmail,
                        false
                    );
                });
                await test.step('Config team setup', async () => {
                    await adminOnboardingPage.enterTeamName(projectName);
                    await adminOnboardingPage.enterTeamMemberEmail(teamMemberEmail);
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Check user creation failure', async () => {
                    await waitUtils.sleep(3500);
                    const contentSubtitle = await adminOnboardingPage.getContentSubtitleText();
                    logger.info(`Content subtitle: ${contentSubtitle}`);
                    expect(contentSubtitle).toContain(
                        'We ran into some trouble adding the following users to the team:'
                    );
                    expect(contentSubtitle).toContain(teamMemberEmail);
                });

                await test.step('Click continue and it should land on project page', async () => {
                    await adminOnboardingPage.clickContinueButton();
                    await waitUtils.waitForCondition(
                        async () => {
                            const names = await projectPage.getProjectNames();
                            return names.length === 1 && names[0] === projectName;
                        },
                        'Wait for exactly one project with the expected name',
                        true,
                        5,
                        1
                    );
                });
            });

            test('T5404939: SSO IDP Creation Failure -- oidc', async function () {
                videoName = 'oidc-idp-creation-failure';
                const teamMemberEmail = `runaitest${Utils.generateShortUuid().slice(0, 5)}@test.com`;
                await test.step('Click let go button', async () => {
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Config saml sso(url) setup', async () => {
                    await adminOnboardingPage.configOIDCSso(invalidOidcUrl, oidcClientId, oidcClientSecret);
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Review sso config page', async () => {
                    await adminOnboardingPage.clickSkipSsoSetupButton();
                });

                await test.step('Skip email server setup', async () => {
                    await adminOnboardingPage.clickSkipButton();
                });

                await test.step('Config team setup', async () => {
                    await adminOnboardingPage.enterTeamName(projectName);
                    await adminOnboardingPage.enterTeamMemberEmail(teamMemberEmail);
                    await adminOnboardingPage.clickContinueButton();
                });
                await test.step('Download credentials', async () => {
                    const credentialsFilePath = await adminOnboardingPage.downloadCredentials();
                    const credentialsFile = readFileSync(credentialsFilePath, 'utf8');
                    expect(credentialsFile).toBeDefined();
                    expect(credentialsFile.includes(teamMemberEmail)).toBe(true);
                });

                await test.step('Check team setup success', async () => {
                    await adminOnboardingPage.clickContinueButton();
                    const isSuccessful = await adminOnboardingPage.isSetupSuccessful();
                    expect(isSuccessful).toBe(true);
                });

                await test.step('Click continue and it should land on project page', async () => {
                    await adminOnboardingPage.clickContinueButton();
                    await waitUtils.waitForCondition(
                        async () => {
                            const names = await projectPage.getProjectNames();
                            return names.length === 1 && names[0] === projectName;
                        },
                        'Wait for exactly one project with the expected name',
                        true,
                        5,
                        1
                    );
                });
            });
        });

        test.describe('backward and forward navigation', function () {
            test.beforeEach(async function () {
                // Clean up any existing realm from previous failed runs
                try {
                    const existingRealm = await Keycloak.getRealmByName(realmName);
                    if (existingRealm) {
                        logger.info(`Realm ${realmName} already exists, deleting it first`);
                        await Keycloak.deleteRealm(realmName);
                        // Wait for deletion to complete
                        await new Promise((resolve) => setTimeout(resolve, 2000));
                    }
                } catch (error) {
                    logger.debug(`Realm ${realmName} does not exist or error checking:`, error);
                }

                await Keycloak.createRealm(realmName);
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                await browserEngine.currentPage.addInitScript(() => {
                    try {
                        localStorage.setItem('admin-onboarding-modal-shown', 'false');
                    } catch (e) {
                        logger.warn('Failed to preset admin-onboarding-modal-shown in init script', e);
                    }
                });
                authPage = new AuthPage(browserEngine.currentPage);
                await browserEngine.currentPage.goto(config.tenantUrl);
                await authPage.signInValidateLoggedIn(
                    flowRootState.users.system_administrator.email,
                    flowRootState.users.system_administrator.password
                );

                // Wait for page to fully load after authentication
                await browserEngine.currentPage.waitForLoadState('networkidle', { timeout: 30000 });
                await browserEngine.currentPage.waitForTimeout(2000);

                adminOnboardingPage = new AdminOnboardingPage(browserEngine.currentPage);
                const onboardingContainer = browserEngine.currentPage.locator('.onboarding-container');

                // Increased timeout for Cloud environments and added retry logic
                await onboardingContainer.waitFor({ state: 'visible', timeout: 90000 });

                // Wait for modal animations to complete
                await browserEngine.currentPage.waitForTimeout(1000);

                logger.info('Admin onboarding modal is visible and ready');
            });
            test.afterEach('delete realm', async () => {
                await Keycloak.deleteRealm(realmName);
                logger.info('deleted realm ' + realmName);
            });
            test('T5404952: Backward Navigation - Review Completed Steps', async function () {
                videoName = 'backward-navigation-review-completed-steps';
                await test.step('Click let go button', async () => {
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Config saml sso(url) setup', async () => {
                    await adminOnboardingPage.configSamlSsoURL(samlUrl);
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Review sso config page', async () => {
                    await adminOnboardingPage.clickDownloadXmlButton();
                    const redirectUrl = await adminOnboardingPage.getRedirectUrl();
                    const entityId = await adminOnboardingPage.getEntityId();
                    logger.info(`SSO Config - Redirect URL: ${redirectUrl}, Entity ID: ${entityId}`);
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Now on team setup page, Check if sso setup step is clickable', async () => {
                    const isStepClickable = await adminOnboardingPage.isStepClickable(1);
                    expect(isStepClickable).toBe(true);
                });
            });

            test('T5404953: Forward Navigation Prevention', async function () {
                videoName = 'complete-standard-flow-with-smtp-server-configured';
                await test.step('Click let go button', async () => {
                    await adminOnboardingPage.clickContinueButton();
                });

                await test.step('Now on sso setup page, Check if team setup step is clickable', async () => {
                    const isStepClickable = await adminOnboardingPage.isStepClickable(2);
                    expect(isStepClickable).toBe(false);
                });
            });
        });

        test.describe('onboarding diag popup flag', function () {
            test('T5404954: Previously Dismissed Modal', async function () {
                videoName = 'previously-dismissed-modal';
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                await browserEngine.currentPage.addInitScript(() => {
                    try {
                        localStorage.setItem('admin-onboarding-modal-shown', 'true');
                    } catch (e) {
                        logger.warn('Failed to preset admin-onboarding-modal-shown in init script', e);
                    }
                });
                authPage = new AuthPage(browserEngine.currentPage);
                adminOnboardingPage = new AdminOnboardingPage(browserEngine.currentPage);
                await test.step('Login UI', async () => {
                    await browserEngine.currentPage.goto(config.tenantUrl);
                    await authPage.signInValidateLoggedIn(
                        flowRootState.users.system_administrator.email,
                        flowRootState.users.system_administrator.password
                    );

                    // Wait for page to fully load after authentication
                    await browserEngine.currentPage.waitForLoadState('networkidle', { timeout: 30000 });
                    await browserEngine.currentPage.waitForTimeout(2000);
                });

                await test.step('check if onboarding modal is not shown', async () => {
                    const isOnboardingModalShown = await adminOnboardingPage.isOnboardingPageExists();
                    expect(isOnboardingModalShown).toBe(false);
                });
            });

            test('T5404955: Login with a project already configured', async function () {
                videoName = 'login-with-a-project-already-configured';
                projectName = objectGenerator.generateNameByPrefix('e2e-project');
                namespace = Utils.namespaceFromProjectName(projectName);
                await test.step('create a project by api', async () => {
                    project = await projectResourcesUtils.createProject(namespace, {
                        name: projectName,
                        deservedGpus: 0,
                        maxAllowedGpus: 0
                    });
                });

                await test.step('Login with another new browser', async () => {
                    browserEngine = new BrowserEngine();
                    await browserEngine.start();
                    await browserEngine.currentPage.addInitScript(() => {
                        try {
                            localStorage.setItem('admin-onboarding-modal-shown', 'false');
                        } catch (e) {
                            logger.warn('Failed to preset admin-onboarding-modal-shown in init script', e);
                        }
                    });
                    authPage = new AuthPage(browserEngine.currentPage);
                    adminOnboardingPage = new AdminOnboardingPage(browserEngine.currentPage);
                    await browserEngine.currentPage.goto(config.tenantUrl);
                    await authPage.signInValidateLoggedIn(
                        flowRootState.users.system_administrator.email,
                        flowRootState.users.system_administrator.password
                    );

                    // Wait for page to fully load after authentication
                    await browserEngine.currentPage.waitForLoadState('networkidle', { timeout: 30000 });
                    await browserEngine.currentPage.waitForTimeout(2000);

                    const isOnboardingModalShown = await adminOnboardingPage.isOnboardingPageExists();
                    expect(isOnboardingModalShown).toBe(false);
                });
            });

            test('T5404956: Login with a L1 researcher user', async function () {
                videoName = 'login-with-a-l1-researcher-user';
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                await browserEngine.currentPage.addInitScript(() => {
                    try {
                        localStorage.setItem('admin-onboarding-modal-shown', 'false');
                    } catch (e) {
                        logger.warn('Failed to preset admin-onboarding-modal-shown in init script', e);
                    }
                });

                authPage = new AuthPage(browserEngine.currentPage);
                adminOnboardingPage = new AdminOnboardingPage(browserEngine.currentPage);
                await test.step('Login UI', async () => {
                    await browserEngine.currentPage.goto(config.tenantUrl);
                    await authPage.signInValidateLoggedIn(
                        flowRootState.users.l1_researcher.email,
                        flowRootState.users.l1_researcher.password
                    );
                });

                await test.step('Login again and check if onboarding modal is not shown', async () => {
                    const isOnboardingModalShown = await adminOnboardingPage.isOnboardingPageExists();
                    expect(isOnboardingModalShown).toBe(false);
                });
            });
        });
    });
});
