/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    Utils,
    authUtils,
    authorizationUtils,
    config,
    flowRootState,
    idpUtils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    settings,
    tags,
    usersManager
} from '../../../../../../../modules/helpers';
import {
    idpNameOidc,
    keycloakClientName,
    realmName
} from '../../../../../../../modules/helpers/config/const';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { StatusCodes } from 'http-status-codes';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { test, expect as testExpect } from '@playwright/test';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { UsersV2 } from '../../../../../../../modules/ui/runaiV2/pages/UsersV2';

    test.describe(`Group Tests with SSO OIDC, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} `, { tag: [tags.services.identityManager, tags.services.keycloak] }, () => {
            shouldRunTests(settings.isSelfHosted && !isOCP())(
                function () {
            let clientSecret: string;
            test.beforeAll('Init Keycloak IDP, Client, Users', async function () {
                try {
                    logger.debug('Cleaning up existing IDP and realm before test');
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameOidc
                    );
                } catch (error) {
                    if (!error.message?.includes('404')) {
                        logger.debug(`IDP cleanup: ${error.message}`);
                    }
                }

                try {
                    await Keycloak.deleteRealm(realmName);
                    await new Promise((resolve) => setTimeout(resolve, 1000));
                } catch (error) {
                    logger.debug(`Realm cleanup: ${error.message}`);
                }

                await Keycloak.createRealm(realmName);
                const client = await Keycloak.createClient(realmName, keycloakClientName, 'openid-connect');
                clientSecret = await Keycloak.getClientSecret(realmName, client.id);
                await Keycloak.createUsers(
                    realmName,
                    ['researchersso', 'editorsso', 'testsso'],
                    [
                        config.controlPlaneAdminPassword,
                        config.controlPlaneAdminPassword,
                        config.controlPlaneAdminPassword
                    ]
                );
                await Keycloak.changeLoginTheme(realmName, 'runai-theme');
                await Keycloak.createKeycloakGroupsWithUsers(
                    realmName,
                    ['researchergroup', 'editorgroup', 'superadmingroup'],
                    ['researchersso', 'editorsso', 'testsso']
                );
                await Keycloak.addOIDCMapperToClient(
                    realmName,
                    client.id,
                    'groups',
                    'oidc-group-membership-mapper',
                    'String'
                );
                await Keycloak.configLoginWithSSOOIDC(
                    Keycloak.getDiscoveryDocumentURL(realmName),
                    keycloakClientName,
                    clientSecret
                );
                await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'superadmingroup',
                    flowRootState.users.system_administrator.role.id
                );
                await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'researchergroup',
                    flowRootState.users.l1_researcher.role.id
                );
                await authorizationUtils.createAccessRuleForGroup(
                    flowRootState.users.system_administrator.accessToken,
                    'editorgroup',
                    flowRootState.users.editor.role.id
                );
            });

            test.afterAll('delete all oidc group tests items', async function () {
                try {
                    try {
                        await idpUtils.deleteIdp(
                            flowRootState.users.system_administrator.accessToken,
                            idpNameOidc
                        );
                    } catch (error) {
                        logger.error(error);
                    }
                    try {
                        await Keycloak.deleteRealm(realmName);
                    } catch (error) {
                        logger.error(error);
                    }
                    try {
                        await operations.deleteAllProjects();
                    } catch (error) {
                        logger.error(error);
                    }
                } finally {
                    await usersManager.restorePreSeededOidcIdp();
                }
            });

            test('should create role groups and validate permissions', async function () {
                const superadminToken = await authUtils.authenticateUser(
                    'testsso',
                    config.controlPlaneAdminPassword,
                    true
                );
                const superAdminProjectName = objectGenerator.generateNameByPrefix('superadminproject');
                const namespacesuperadmin: string = Utils.namespaceFromProjectName(superAdminProjectName);
                try {
                    await projectResourcesUtils.createProject(
                        namespacesuperadmin,
                        {
                            name: superAdminProjectName,
                            deservedGpus: 2,
                            maxAllowedGpus: 2,
                            permissions: { users: [], groups: ['researchergroup'] }
                        },
                        superadminToken.accessToken
                    );
                } catch (error) {
                    expect.fail(
                        `SSo user superadmin fail to create project ${superAdminProjectName} with error: ${error}`
                    );
                }
                const editorToken = await authUtils.authenticateUser(
                    'editorsso@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const editorProjectName = objectGenerator.generateNameByPrefix('editorproject');
                const namespaceeditor: string = Utils.namespaceFromProjectName(editorProjectName);
                try {
                    await projectResourcesUtils.createProject(
                        namespaceeditor,
                        {
                            name: editorProjectName,
                            deservedGpus: 2,
                            maxAllowedGpus: 2,
                            permissions: { users: [], groups: ['researchergroup'] }
                        },
                        editorToken.accessToken
                    );
                } catch (error) {
                    expect.fail(
                        `SSo user editor fail to create project ${editorProjectName} with error: ${error}`
                    );
                }
                const researcherToken = await authUtils.authenticateUser(
                    'researchersso@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const researcherProjectName = objectGenerator.generateNameByPrefix('researcherproject');
                const namespaceresearcher: string = Utils.namespaceFromProjectName(researcherProjectName);
                try {
                    await projectResourcesUtils.createProject(
                        namespaceresearcher,
                        {
                            name: researcherProjectName,
                            deservedGpus: 2,
                            maxAllowedGpus: 2,
                            permissions: { users: [], groups: ['researchergroup'] }
                        },
                        researcherToken.accessToken
                    );
                } catch (error) {
                    logger.info(
                        `SSo user researcher fail to create project ${researcherProjectName} with error: ${error} as expected`
                    );
                }
            });

            test('should validate user assign to researcher group with usersV2 access rules editor should not submit project', async function () {
                const updatedReasercherToken = await authUtils.authenticateUser(
                    'researchersso@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const researcherProjectName = objectGenerator.generateNameByPrefix('researcherproject');
                const namespaceresearcher: string = Utils.namespaceFromProjectName(researcherProjectName);
                try {
                    await projectResourcesUtils.createProject(
                        namespaceresearcher,
                        {
                            name: researcherProjectName,
                            deservedGpus: 2,
                            maxAllowedGpus: 2,
                            permissions: { users: [], groups: ['researchergroup'] }
                        },
                        updatedReasercherToken.accessToken
                    );
                } catch (error) {
                    expect(error.body.code, 'researcher-role SSO user should be forbidden from creating a project (403)').to.be.equal(StatusCodes.FORBIDDEN);
                }
            });
            test('should add access rule of editor to group with researcher and validate permissions', async function () {
                const updatedReasercherToken = await authUtils.authenticateUser(
                    'researchersso@example.com',
                    config.controlPlaneAdminPassword,
                    true
                );
                const researcherProjectName = objectGenerator.generateNameByPrefix('researcherproject');
                const namespaceresearcher: string = Utils.namespaceFromProjectName(researcherProjectName);
                try {
                    await projectResourcesUtils.createProject(
                        namespaceresearcher,
                        {
                            name: researcherProjectName,
                            deservedGpus: 2,
                            maxAllowedGpus: 2,
                            permissions: { users: [], groups: ['researchergroup'] }
                        },
                        updatedReasercherToken.accessToken
                    );
                } catch (error) {
                    expect(error.body.code, 'researcher-role SSO user should still be forbidden from creating a project after editor access rule is added (403)').to.be.equal(StatusCodes.FORBIDDEN);
                }
            });

            test('should display inherited rule badge in user access rules modal for group member', async function () {
                let browserEngine: BrowserEngine;
                try {
                    browserEngine = new BrowserEngine();
                    await browserEngine.start();
                    const auth = new AuthPage(browserEngine.currentPage);
                    const users = new UsersV2(browserEngine.currentPage);

                    await auth.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
                    await users.navigateToUsers();
                    await users.searchByName('researchersso@example.com');
                    await users.clickOnUser('researchersso@example.com');
                    await users.clickOnAccessRulesBtn();

                    const inheritedRule = browserEngine.currentPage
                        .locator('.card-item')
                        .filter({ hasText: 'Group - researchergroup' })
                        .first();
                    await testExpect(inheritedRule).toBeVisible({ timeout: 15000 });
                    await testExpect(inheritedRule).toContainText('Subject: Group - researchergroup');

                    const deleteBtn = inheritedRule.locator('[aid="card-item-delete-button"]');
                    await testExpect(deleteBtn).toBeDisabled();
                    await deleteBtn.hover();
                    await testExpect(
                        browserEngine.currentPage.getByText(
                            'This rule can\'t be deleted because it is inherited from group researchergroup'
                        )
                    ).toBeVisible({ timeout: 5000 });
                } catch (error) {
                    logger.error('inherited rule badge test failed:', error);
                    throw error;
                } finally {
                    await browserEngine?.stopAndArchive(test.info(), 'inherited-rule-badge');
                }
            });
        });
    }
);
