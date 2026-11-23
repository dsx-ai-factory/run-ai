/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    authorizationUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectUtils,
    tags,
    Utils,
    usersManager,
    waitUtils,
    idpUtils,
    usersUtils,
    config
} from '../../../../../../modules/helpers';
import { AccessRulesClusterUtils } from '../../../../../../modules/helpers/utils/AccessRulesClusterUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { AdminOnboardingPage } from '../../../../../../modules/ui/runaiV2/pages/AdminOnboardingPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { test, expect } from '@playwright/test';
import { UsersFilterSortFields } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { Idp } from '../../../../../../modules/api/apis/identityManagerApi/generated/model/idp';
import { readFileSync } from 'fs';
import { ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { Keycloak } from '../../../../../../modules/helpers/utils/Keycloak';
import { realmName, keycloakClientName, idpNameOidc } from '../../../../../../modules/helpers/config/const';

test.describe(`Admin Onboarding UI ${tags.teams.loong} ${tags.tool.ui} `, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let projectName: string;
        let namespace: string;
        let browserEngine: BrowserEngine;
        let deltegateBrowserEngine: BrowserEngine;
        let adminOnboardingPage: AdminOnboardingPage;
        let authPage: AuthPage;
        let projectPage: ProjectPage;
        let videoName: string;
        let accessRulesClusterUtils: AccessRulesClusterUtils;
        const samlUrl = Keycloak.getConfigurationProtocolEndpointSAMLUrl(realmName);
        const oidcUrl = Keycloak.getDiscoveryDocumentURL(realmName);
        const oidcClientId = keycloakClientName;
        let oidcClientSecret: string;
        const delegateUserEmail = `runaidelegate${Utils.generateShortUuid().slice(0, 5)}@example.com`;
        let delegateUserTempPassword: string;
        const delegateSSOUser = 'qatest@example.com';
        const teamMemberEmail = `runaitest${Utils.generateShortUuid().slice(0, 5)}@example.com`;

        const deleteAllIdps = async (): Promise<void> => {
            try {
                const idps = await idpUtils.getIdps(flowRootState.users.system_administrator.idToken);
                await Promise.all(
                    idps.map(({ alias }) =>
                        idpUtils
                            .deleteIdp(flowRootState.users.system_administrator.idToken, alias)
                            .catch((err) => {
                                logger.warn(`Failed to delete IDP ${alias}:`, err);
                            })
                    )
                );
            } catch (err) {
                logger.warn('Failed to list IDPs:', err);
            }
        };
        test.beforeAll(async function () {
            logger.info(`SAML metadata URL: ${samlUrl}`);
            await Keycloak.createRealm(realmName);
            const client = await Keycloak.createClient(realmName, oidcClientId, 'openid-connect');
            oidcClientSecret = await Keycloak.getClientSecret(realmName, client.id);
            await Keycloak.createUser(realmName, 'qatest', config.controlPlaneAdminPassword);
            logger.info(`OIDC discovery URL: ${oidcUrl}`);
            logger.info(`OIDC prepared for clientId=${oidcClientId}`);
        });

        test.beforeEach(async function () {
            try {
                await operations.deleteAllProjects(flowRootState.cluster.uuid);
            } catch (error) {
                logger.warn('Failed to delete all projects:', error);
            }
            await deleteAllIdps();

            browserEngine = new BrowserEngine();
            await browserEngine.start();
            deltegateBrowserEngine = new BrowserEngine();
            await deltegateBrowserEngine.start();
            projectName = objectGenerator.generateNameByPrefix('e2e-project');
            await browserEngine.currentPage.addInitScript(() => {
                try {
                    localStorage.setItem('admin-onboarding-modal-shown', 'false');
                } catch (e) {
                    logger.warn('Failed to preset admin-onboarding-modal-shown in init script', e);
                }
            });
            await deltegateBrowserEngine.currentPage.addInitScript(() => {
                try {
                    localStorage.setItem('admin-onboarding-modal-shown', 'false');
                } catch (e) {
                    logger.warn('Failed to preset admin-onboarding-modal-shown in init script', e);
                }
            });
            authPage = new AuthPage(browserEngine.currentPage);
            await browserEngine.currentPage.goto(config.tenantUrl, { waitUntil: 'domcontentloaded' });
            await authPage.signInValidateLoggedIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
            adminOnboardingPage = new AdminOnboardingPage(browserEngine.currentPage);
            projectPage = new ProjectPage(browserEngine.currentPage);
            accessRulesClusterUtils = new AccessRulesClusterUtils();
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'admin-onboarding-delegation-' + videoName);
            if (deltegateBrowserEngine) {
                try {
                    await deltegateBrowserEngine.stopAndArchive(testInfo, 'delegate-browser-' + videoName);
                } catch (error) {
                    logger.warn('Failed to stop delegate browser:', error);
                }
            }

            await deleteAllIdps();

            if (projectName) {
                try {
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    await operations.deleteProjectAndNamespace(projectId);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch (error) {
                    logger.error(`failed to delete project ${projectName} and namespace ${namespace}`, error);
                }
            }
        });

        test.afterAll('cleanup OIDC idp and automation realm', async function () {
            try {
                await Keycloak.deleteIDPFromRealm('runai', idpNameOidc);
                await Keycloak.deleteRealm(realmName);
            } catch (error) {
                logger.warn('Failed OIDC cleanup:', error);
            }
        });

        test('T5404935: Manual Delegation - SSO Setup & Team Setup ', async function () {
            videoName = 'T5404935-part1-delegate-to-user-';
            await test.step('Click let go button', async () => {
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('config delegation option in sso page', async () => {
                await adminOnboardingPage.selectDelegationOption();
                await adminOnboardingPage.enterDelegateEmail(delegateUserEmail);
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Download credentials', async () => {
                const credentialsFilePath = await adminOnboardingPage.downloadCredentials();
                const credentialsFile = readFileSync(credentialsFilePath, 'utf8');

                const [username, password] = credentialsFile.split('\n')[1].split(',');
                expect(username).toBe(delegateUserEmail);
                expect(password).toBeDefined();
                delegateUserTempPassword = password;
                logger.info(`Extracted password: ${password}`);
            });

            await test.step('Check setup success', async () => {
                const isSuccessful = await adminOnboardingPage.isSetupSuccessful();
                expect(isSuccessful).toBe(true);
            });
            await test.step('Click continue and it should land on project page', async () => {
                await adminOnboardingPage.clickContinueButton();
                const projectNames = await projectPage.getProjectNames();
                expect(projectNames.length).toBe(0);
            });

            await test.step('Check the delegate user is created', async () => {
                const user = await usersUtils.getUsersFilteredBy(
                    flowRootState.users.system_administrator.idToken,
                    [`username=@${delegateUserEmail}`],
                    UsersFilterSortFields.Username
                );
                logger.info('User:', user);
                expect(user.length).toBe(1);
                expect(user[0].username).toBe(delegateUserEmail);
                expect(user[0].isLocal).toBe(true);
            });

            await test.step('Check the access rule of delegate user is created', async () => {
                const accessRulesCount = await accessRulesClusterUtils.getAccessRulesByUser(
                    flowRootState.users.system_administrator.idToken,
                    delegateUserEmail
                );
                const accessRulesDetails = await accessRulesClusterUtils.getAccessRulesDetailsByUser(
                    flowRootState.users.system_administrator.idToken,
                    delegateUserEmail
                );

                logger.info(`Access rules count for ${delegateUserEmail}: ${accessRulesCount}`);
                logger.info('Access rules details:', accessRulesDetails);
                expect(accessRulesCount).toBe(1);

                const firstRule = accessRulesDetails[0];
                expect(firstRule.roleName).toBe('System administrator');

                logger.info(`✓ Scope Name: ${firstRule.scopeName} (expected: ${projectName})`);
                logger.info(`✓ Role Name: ${firstRule.roleName} (expected: System administrator)`);
            });

            await test.step('Open new browser for delegate user and login', async () => {
                authPage = new AuthPage(deltegateBrowserEngine.currentPage);
                adminOnboardingPage = new AdminOnboardingPage(deltegateBrowserEngine.currentPage);
                projectPage = new ProjectPage(deltegateBrowserEngine.currentPage);
                await deltegateBrowserEngine.currentPage.addInitScript(() => {
                    try {
                        localStorage.setItem('admin-onboarding-modal-shown', 'false');
                    } catch (e) {
                        logger.warn('Failed to preset admin-onboarding-modal-shown in init script', e);
                    }
                });
                await deltegateBrowserEngine.currentPage.goto(config.tenantUrl, {
                    waitUntil: 'domcontentloaded'
                });
                await authPage.submitLoginForm(
                    delegateUserEmail,
                    delegateUserTempPassword,
                    config.controlPlaneAdminPassword,
                    false
                );
            });

            await test.step('Config saml sso(url) setup', async () => {
                await adminOnboardingPage.clickContinueButton();
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

            await test.step('Skip email server setup', async () => {
                await adminOnboardingPage.clickSkipButton();
            });

            await test.step('Config team setup', async () => {
                await adminOnboardingPage.enterTeamName(projectName);
                await adminOnboardingPage.selectMemberEmailsOption();
                await adminOnboardingPage.enterTeamMemberEmail(teamMemberEmail);
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Check team setup success', async () => {
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
                    30,
                    1
                );
            });

            await test.step('Check the access rule is created', async () => {
                const accessRulesCount = await accessRulesClusterUtils.getAccessRulesByUser(
                    flowRootState.users.system_administrator.idToken,
                    teamMemberEmail
                );
                const accessRulesDetails = await accessRulesClusterUtils.getAccessRulesDetailsByUser(
                    flowRootState.users.system_administrator.idToken,
                    teamMemberEmail
                );

                logger.info(`Access rules count for ${teamMemberEmail}: ${accessRulesCount}`);
                logger.info('Access rules details:', accessRulesDetails);
                expect(accessRulesCount).toBe(1);

                const firstRule = accessRulesDetails[0];
                expect(firstRule.scopeName).toBe(projectName);
                expect(firstRule.roleName).toBe('AI practitioner');

                logger.info(`✓ Scope Name: ${firstRule.scopeName} (expected: ${projectName})`);
                logger.info(`✓ Role Name: ${firstRule.roleName} (expected: AI practitioner)`);
            });

            await test.step('Check sso is configured', async () => {
                const idps = await idpUtils.getIdps(flowRootState.users.system_administrator.idToken);
                const idp = idps.find((i) => i.type === Idp.TypeEnum.Saml);
                expect(idp).toBeDefined();
                expect(idp!.alias.startsWith('saml-')).toBe(true);
                expect(idp!.type).toEqual(Idp.TypeEnum.Saml);
                expect(idp!.redirectUri).toContain('https');
            });
        });

        test('T5404936: Manual Delegation - Team Setup ', async function () {
            videoName = 'T5404936-delegate-';
            await test.step('Click let go button', async () => {
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Config OIDC sso(url) setup', async () => {
                await adminOnboardingPage.configOIDCSso(oidcUrl, oidcClientId, oidcClientSecret);
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Review sso config page', async () => {
                const redirectUrl = await adminOnboardingPage.getRedirectUrl();
                logger.info(`SSO Config - Redirect URL: ${redirectUrl}`);
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Skip email server setup', async () => {
                await adminOnboardingPage.clickSkipButton();
            });

            await test.step('config delegation option in team setup page', async () => {
                await adminOnboardingPage.selectDelegationOption();
                await adminOnboardingPage.enterDelegateEmail(delegateSSOUser);
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Check setup success', async () => {
                const isSuccessful = await adminOnboardingPage.isSetupSuccessful();
                expect(isSuccessful).toBe(true);
            });
            await test.step('Click continue and it should land on project page', async () => {
                await adminOnboardingPage.clickContinueButton();
                await waitUtils.waitForCondition(
                    async () => {
                        const names = await projectPage.getProjectNames();
                        return names.length === 0;
                    },
                    'Wait for no projects',
                    true,
                    5,
                    1
                );
            });

            await test.step('Check the access rule of delegate user is created', async () => {
                const accessRulesCount = await accessRulesClusterUtils.getAccessRulesByUser(
                    flowRootState.users.system_administrator.idToken,
                    delegateSSOUser
                );
                const accessRulesDetails = await accessRulesClusterUtils.getAccessRulesDetailsByUser(
                    flowRootState.users.system_administrator.idToken,
                    delegateSSOUser
                );

                logger.info(`Access rules count for ${delegateSSOUser}: ${accessRulesCount}`);
                logger.info('Access rules details:', accessRulesDetails);

                expect(accessRulesCount).toBe(1);
                const firstRule = accessRulesDetails[0];
                expect(firstRule.roleName).toBe('System administrator');

                logger.info(`✓ Scope Name: ${firstRule.scopeName} (expected: ${projectName})`);
                logger.info(`✓ Role Name: ${firstRule.roleName} (expected: System administrator)`);
            });

            await test.step('Check sso is configured', async () => {
                const idps = await idpUtils.getIdps(flowRootState.users.system_administrator.idToken);
                const idp = idps.find((i) => i.type === Idp.TypeEnum.Oidc);
                expect(idp).toBeDefined();
                expect(idp!.alias.startsWith('oidc-')).toBe(true);
                expect(idp!.type).toEqual(Idp.TypeEnum.Oidc);
                expect(idp!.redirectUri).toContain('https');
            });

            await test.step('Open new browser for delegate user and login', async () => {
                authPage = new AuthPage(deltegateBrowserEngine.currentPage);
                adminOnboardingPage = new AdminOnboardingPage(deltegateBrowserEngine.currentPage);
                projectPage = new ProjectPage(deltegateBrowserEngine.currentPage);
                await deltegateBrowserEngine.currentPage.goto(config.tenantUrl);
                await deltegateBrowserEngine.currentPage.locator('[id="sso-button-id"]').click();
                await deltegateBrowserEngine.currentPage.locator('[id="username"]').fill(delegateSSOUser);
                await deltegateBrowserEngine.currentPage
                    .locator('[id="password"]')
                    .fill(config.controlPlaneAdminPassword);
                await waitUtils.sleep(3000);
                await deltegateBrowserEngine.currentPage.locator('#kc-login').click();
            });

            await test.step('Config team setup', async () => {
                await adminOnboardingPage.enterTeamName(projectName);
                await adminOnboardingPage.selectMemberEmailsOption();
                await adminOnboardingPage.enterTeamMemberEmail(teamMemberEmail);
                await adminOnboardingPage.clickCompleteSetupButton();
            });

            await test.step('Check team setup success', async () => {
                const isSuccessful = await adminOnboardingPage.isSetupSuccessful();
                expect(isSuccessful).toBe(true);
            });

            await test.step('Click continue and it should land on project page', async () => {
                await adminOnboardingPage.clickCompleteSetupButton();
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

            await test.step('Check the access rule is created', async () => {
                const accessRulesCount = await accessRulesClusterUtils.getAccessRulesByUser(
                    flowRootState.users.system_administrator.idToken,
                    teamMemberEmail
                );
                const accessRulesDetails = await accessRulesClusterUtils.getAccessRulesDetailsByUser(
                    flowRootState.users.system_administrator.idToken,
                    teamMemberEmail
                );

                logger.info(`Access rules count for ${teamMemberEmail}: ${accessRulesCount}`);
                logger.info('Access rules details:', accessRulesDetails);

                expect(accessRulesCount).toBe(1);
                const firstRule = accessRulesDetails[0];
                expect(firstRule.scopeName).toBe(projectName);
                expect(firstRule.roleName).toBe('AI practitioner');

                logger.info(`✓ Scope Name: ${firstRule.scopeName} (expected: ${projectName})`);
                logger.info(`✓ Role Name: ${firstRule.roleName} (expected: AI practitioner)`);
            });
        });

        test('T5404937: Automatic Partial Flow - SSO already Configured ', async function () {
            videoName = 'T5404937-auto-delegation';
            await test.step('Click let go button', async () => {
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Config saml sso(url) setup', async () => {
                await adminOnboardingPage.configSamlSsoURL(samlUrl);
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Review sso config page and close onboarding button', async () => {
                await adminOnboardingPage.clickDownloadXmlButton();
                const redirectUrl = await adminOnboardingPage.getRedirectUrl();
                const entityId = await adminOnboardingPage.getEntityId();
                logger.info(`SSO Config - Redirect URL: ${redirectUrl}, Entity ID: ${entityId}`);
                await adminOnboardingPage.clickContinueButton();
                await adminOnboardingPage.clickSkipButton();
                await adminOnboardingPage.clickCloseOnboardingButton();
            });

            await test.step('Login with another sys admin', async () => {
                const userEmail = objectGenerator.generateEmailByPrefix('system-admin');
                const adminUser = (
                    await usersManager.createLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        userEmail,
                        false
                    )
                ).body;
                await authorizationUtils.bindUserWithRoles(
                    adminUser,
                    ['System administrator'],
                    { scopeId: flowRootState.cluster.tenantId.toString(), scopeType: ScopeType.Tenant },
                    flowRootState.users.system_administrator.accessToken
                );

                authPage = new AuthPage(deltegateBrowserEngine.currentPage);
                adminOnboardingPage = new AdminOnboardingPage(deltegateBrowserEngine.currentPage);
                projectPage = new ProjectPage(deltegateBrowserEngine.currentPage);
                await deltegateBrowserEngine.currentPage.goto(config.tenantUrl, {
                    waitUntil: 'domcontentloaded'
                });
                await authPage.signInValidateLoggedIn(adminUser.username, adminUser.tempPassword);
                await waitUtils.sleep(3000);
            });

            await test.step('Config team setup', async () => {
                await adminOnboardingPage.enterTeamName(projectName);
                await adminOnboardingPage.selectMemberEmailsOption();
                await adminOnboardingPage.enterTeamMemberEmail(teamMemberEmail);
                await adminOnboardingPage.clickCompleteSetupButton();
            });

            await test.step('Check team setup success', async () => {
                const isSuccessful = await adminOnboardingPage.isSetupSuccessful();
                expect(isSuccessful).toBe(true);
            });

            await test.step('Click continue and it should land on project page', async () => {
                await adminOnboardingPage.clickCompleteSetupButton();
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

            await test.step('Check the access rule is created', async () => {
                const accessRulesCount = await accessRulesClusterUtils.getAccessRulesByUser(
                    flowRootState.users.system_administrator.idToken,
                    teamMemberEmail
                );
                const accessRulesDetails = await accessRulesClusterUtils.getAccessRulesDetailsByUser(
                    flowRootState.users.system_administrator.idToken,
                    teamMemberEmail
                );

                logger.info(`Access rules count for ${teamMemberEmail}: ${accessRulesCount}`);
                logger.info('Access rules details:', accessRulesDetails);

                expect(accessRulesCount).toBe(1);
                const firstRule = accessRulesDetails[0];
                expect(firstRule.scopeName).toBe(projectName);
                expect(firstRule.roleName).toBe('AI practitioner');

                logger.info(`✓ Scope Name: ${firstRule.scopeName} (expected: ${projectName})`);
                logger.info(`✓ Role Name: ${firstRule.roleName} (expected: AI practitioner)`);
            });
        });
    });
});
