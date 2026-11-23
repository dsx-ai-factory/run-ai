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
    projectUtils,
    tags,
    Utils,
    waitUtils,
    idpUtils,
    authorizationUtils,
    usersUtils,
    notificationsUtils
} from '../../../../../../modules/helpers';
import { AccessRulesClusterUtils } from '../../../../../../modules/helpers/utils/AccessRulesClusterUtils';

import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { AdminOnboardingPage } from '../../../../../../modules/ui/runaiV2/pages/AdminOnboardingPage';
import { SubjectType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { test, expect } from '@playwright/test';
import { UsersFilterSortFields } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { readFileSync } from 'fs';
import { Keycloak } from '../../../../../../modules/helpers/utils/Keycloak';
import { realmName, keycloakClientName, idpNameOidc, idpNameSaml } from '../../../../../../modules/helpers/config/const';
import { SMTP_CONFIG } from '../../../../../../modules/helpers/utils/NotificationUtils';
import { Idp } from '../../../../../../modules/api/apis/identityManagerApi/generated/model/idp';

test.describe(`Admin Onboarding UI ${tags.teams.loong} ${tags.tool.ui} `, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let browserEngine: BrowserEngine;
        let adminOnboardingPage: AdminOnboardingPage;
        let authPage: AuthPage;
        let projectPage: ProjectPage;
        let workloadPage: WorkloadPage;
        let videoName: string;
        let accessRulesClusterUtils: AccessRulesClusterUtils;
        const samlUrl = Keycloak.getConfigurationProtocolEndpointSAMLUrl(realmName);
        const oidcUrl = Keycloak.getDiscoveryDocumentURL(realmName);
        const oidcClientId = keycloakClientName;
        let oidcClientSecret: string;
        const smtpHost = SMTP_CONFIG.host;
        const authType = 'login';
        const fromEmail = 'runaitest@test.com';
        const fromName = 'Runai Test';
        const teamMemberEmail = `runaitest${Utils.generateShortUuid().slice(0, 5)}@test.com`;

        const deleteAllIdps = async (): Promise<void> => {
            const token = flowRootState.users.system_administrator.accessToken;
            try {
                const idps = await idpUtils.getIdps(token);
                await Promise.all(
                    idps.map(({ alias }) =>
                        idpUtils.deleteIdp(token, alias).catch((err) => {
                            logger.warn(`Failed to delete IDP ${alias}:`, err);
                        })
                    )
                );
            } catch (err) {
                logger.warn('Failed to list IDPs:', err);
            }
        };

        const deleteEmailServerChannel = async (): Promise<void> => {
            const accessToken = flowRootState.users.system_administrator.accessToken;
            try {
                const channel = await notificationsUtils.getEmailChannel(accessToken);
                if (channel) {
                    await notificationsUtils.deleteEmailChannel(accessToken).catch((err) => {
                        logger.warn('Failed to delete email server channel runai-email:', err);
                    });
                }
            } catch (err) {
                logger.debug('No email server channel to delete or failed to get:', err);
            }
        };

        test.beforeAll(async function () {
            logger.info(`SAML metadata URL: ${samlUrl}`);

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

            // Create fresh realm
            await Keycloak.createRealm(realmName);
            const client = await Keycloak.createClient(realmName, oidcClientId, 'openid-connect');
            oidcClientSecret = await Keycloak.getClientSecret(realmName, client.id);
            logger.info(`OIDC discovery URL: ${oidcUrl}`);
            logger.info(`OIDC prepared for clientId=${oidcClientId}`);
        });

        test.afterAll('cleanup OIDC idp and automation realm', async function () {
            logger.info('Starting afterAll cleanup');

            // Delete project and namespace
            if (project && project.id) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch (error) {
                    logger.error(
                        'failed to delete project ' + projectName + ' and namespace ' + namespace,
                        error
                    );
                }
            }

            // Delete OIDC IDP (might not exist if test used SAML)
            try {
                await idpUtils.deleteIdp(flowRootState.users.system_administrator.accessToken, idpNameOidc);
                logger.info('Deleted OIDC IDP');
            } catch (error) {
                logger.debug('Failed to delete OIDC IDP (may not exist):', error.message);
            }

            // Delete SAML IDP (might not exist if test used OIDC)
            try {
                await idpUtils.deleteIdp(flowRootState.users.system_administrator.accessToken, idpNameSaml);
                logger.info('Deleted SAML IDP');
            } catch (error) {
                logger.debug('Failed to delete SAML IDP (may not exist):', error.message);
            }

            // Delete Keycloak realm
            try {
                await Keycloak.deleteRealm(realmName);
                logger.info(`Deleted Keycloak realm: ${realmName}`);
            } catch (error) {
                logger.warn('Failed to delete Keycloak realm:', error.message);
            }
        });

        test.beforeEach(async function () {
            try {
                await operations.deleteAllProjects(flowRootState.cluster.uuid);
            } catch (error) {
                logger.warn('Failed to delete all projects:', error);
            }
            await deleteAllIdps();
            await deleteEmailServerChannel();

            browserEngine = new BrowserEngine();
            await browserEngine.start();
            projectName = objectGenerator.generateNameByPrefix('e2e-project');
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
            adminOnboardingPage = new AdminOnboardingPage(browserEngine.currentPage);
            projectPage = new ProjectPage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);
            accessRulesClusterUtils = new AccessRulesClusterUtils();
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
            if (projectName) {
                try {
                    const projectId = await projectUtils.getProjectIdByProjectName(projectName);
                    await operations.deleteProjectAndNamespace(projectId);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch (error) {
                    logger.error(`failed to delete project ${projectName} and namespace ${namespace}`, error);
                }
            }
            await deleteAllIdps();
            await deleteEmailServerChannel();
        });

        test('T5404930: Complete Standard Flow with sso saml', async function () {
            videoName = 'complete-standard-flow-with-smtp-server-configured';
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

            await test.step('Config email server', async () => {
                await adminOnboardingPage.configureEmailServer(
                    smtpHost,
                    authType,
                    fromEmail,
                    fromName,
                    SMTP_CONFIG.user,
                    SMTP_CONFIG.password
                );
                await waitUtils.waitForCondition(
                    async () => {
                        return await adminOnboardingPage.isEmailServerVerified();
                    },
                    'Wait for email server to be verified',
                    true,
                    3,
                    1
                );
                await adminOnboardingPage.clickContinueButton();
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
                await waitUtils.sleep(3000);
                const projectNames = await projectPage.getProjectNames();
                expect(projectNames.length).toBe(1);
                expect(projectNames[0]).toBe(projectName);
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

            await test.step('Check the email server is configured', async () => {
                const channel = await notificationsUtils.getEmailChannel(
                    flowRootState.users.system_administrator.accessToken
                );
                expect(channel.config.smtpHost).toBe(smtpHost);
                expect(channel.config.fromDisplayName).toBe(fromName);
                expect(channel.config.from).toBe(fromEmail);
            });

            await test.step('Check the new project can be loaded when creating training', async () => {
                await workloadPage.goToPage();
                await workloadPage.openTrainingForm();
                const projectCard = workloadPage.locator(`[aid="${projectName}-card"]`);
                await projectCard.waitFor({ state: 'visible', timeout: 5000 });
            });
        });

        test('T5404931: Complete Standard Flow with sso oidc', async function () {
            videoName = 'select-multiple-workloads-and-delete-them';
            const groupName = `runaigroup${Utils.generateShortUuid().slice(0, 5)}`;
            await test.step('Click let go button', async () => {
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Config saml sso(url) setup', async () => {
                await adminOnboardingPage.configOIDCSso(oidcUrl, oidcClientId, oidcClientSecret);
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Review sso config page', async () => {
                const redirectUrl = await adminOnboardingPage.getRedirectUrl();
                logger.info(`SSO Config - Redirect URL: ${redirectUrl}`);
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Config email server', async () => {
                await adminOnboardingPage.configureEmailServer(
                    smtpHost,
                    authType,
                    fromEmail,
                    fromName,
                    SMTP_CONFIG.user,
                    SMTP_CONFIG.password
                );
                await waitUtils.waitForCondition(
                    async () => {
                        return await adminOnboardingPage.isEmailServerVerified();
                    },
                    'Wait for email server to be verified',
                    true,
                    3,
                    1
                );
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Config team setup', async () => {
                await adminOnboardingPage.enterTeamName(projectName);
                await adminOnboardingPage.selectGroupNameOption();
                await adminOnboardingPage.enterGroupName(groupName);
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
                    5,
                    1
                );
            });

            await test.step('Check the access rule is created', async () => {
                const accessRules = await authorizationUtils.getAccessRules(
                    {
                        subjectType: SubjectType.Group,
                        filterBy: [`subjectIdFilter=@${groupName}`]
                    },
                    flowRootState.users.system_administrator.idToken
                );

                const rulesArray = accessRules.accessRules;

                expect(rulesArray.length).toBe(1);
                const firstRule = accessRules.accessRules[0];
                expect(firstRule.subjectId).toBe(groupName);
                expect(firstRule.subjectType).toBe('group');
                expect(firstRule.roleName).toBe('AI practitioner');
            });

            await test.step('Check sso is configured', async () => {
                const idps = await idpUtils.getIdps(flowRootState.users.system_administrator.idToken);
                const idpResponse = idps.find((i) => i.type === Idp.TypeEnum.Oidc);
                expect(idpResponse).toBeDefined();
                expect(idpResponse!.alias.startsWith('oidc-')).toBe(true);
                expect(idpResponse!.type).toEqual(Idp.TypeEnum.Oidc);
                expect(idpResponse!.redirectUri).toContain('https');
            });

            await test.step('Check the email server is configured', async () => {
                const channel = await notificationsUtils.getEmailChannel(
                    flowRootState.users.system_administrator.accessToken
                );
                expect(channel.config.smtpHost).toBe(smtpHost);
                expect(channel.config.fromDisplayName).toBe(fromName);
                expect(channel.config.from).toBe(fromEmail);
            });
        });

        test('T5404932: Standard Flow -- sso setup skipped', async function () {
            videoName = 'select-multiple-workloads-and-delete-them';
            await test.step('Click let go button', async () => {
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Config saml sso(url) setup', async () => {
                await adminOnboardingPage.clickSkipButton();
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
                logger.info('Credentials file:', credentialsFile);
                logger.info('Credentials file path:', credentialsFilePath);
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

            await test.step('Check the user is created', async () => {
                const user = await usersUtils.getUsersFilteredBy(
                    flowRootState.users.system_administrator.idToken,
                    [`username=@${teamMemberEmail}`],
                    UsersFilterSortFields.Username
                );
                logger.info('User:', user);
                expect(user.length).toBe(1);
                expect(user[0].username).toBe(teamMemberEmail);
                expect(user[0].isLocal).toBe(true);
            });

            await test.step('Check the access rule of user is created', async () => {
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

        test('T5404933: Standard Flow -- team setup skipped', async function () {
            videoName = 'complete-standard-flow-with-smtp-server-configured';
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

            await test.step('Skip email server setup', async () => {
                await adminOnboardingPage.clickSkipButton();
            });

            await test.step('Skip team setup', async () => {
                await adminOnboardingPage.clickSkipButton();
            });

            await test.step('Click continue and it should land on project page', async () => {
                await adminOnboardingPage.clickContinueButton();
                await waitUtils.waitForCondition(
                    async () => {
                        const names = await projectPage.getProjectNames();
                        return names.length === 0;
                    },
                    'Wait for exactly one project with the expected name',
                    true,
                    5,
                    1
                );
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

        test('T5404934: Standard Flow -- sso setup and team setup skipped', async function () {
            videoName = 'complete-standard-flow-with-smtp-server-configured';
            await test.step('Click let go button', async () => {
                await adminOnboardingPage.clickContinueButton();
            });

            await test.step('Skip sso setup', async () => {
                await adminOnboardingPage.clickSkipButton();
            });

            await test.step('Skip email server setup', async () => {
                await adminOnboardingPage.clickSkipButton();
            });
            await test.step('Skip team setup', async () => {
                await adminOnboardingPage.clickSkipButton();
            });

            await test.step('Click continue and it should land on project page', async () => {
                await adminOnboardingPage.clickContinueButton();
                await waitUtils.waitForCondition(
                    async () => {
                        const names = await projectPage.getProjectNames();
                        return names.length === 0;
                    },
                    'Wait for exactly one project with the expected name',
                    true,
                    5,
                    1
                );
            });
        });
    });
});
