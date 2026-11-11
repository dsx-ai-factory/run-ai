/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authorizationUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    tags,
    Utils,
    orgUnitUtils,
    nodePoolUtils,
    settings,
    usersManager
} from '../../../../../../modules/helpers';
import { authorizationApi, orgUnitService } from '../../../../../../modules/api';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { App } from '../../../../../../modules/ui/runaiV2/pages/App';
import { AccessRulesPage } from '../../../../../../modules/ui/runaiV2/pages/AccessRulesPage';
import { test, expect } from '@playwright/test';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/model/userPostResponse';
import { AccessRule } from '../../../../../../modules/api/apis/authorizationApi/generated/model/accessRule';
import { Department, Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { Role } from '../../../../../../modules/api/apis/authorizationApi/generated/model/role';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { shouldRunTests } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';

test.describe(
    `Access Rules Bulk Create and Delete UI Tests, tags: ${tags.teams.viper} ${tags.tool.ui} ${tags.resource.accessRules}`,
    { tag: [tags.services.authorization, tags.services.accessruleController, tags.runOn.parallelCI] },
    () => {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let appPage: App;
        let accessRulesPage: AccessRulesPage;
        let videoName: string;

        // Test data
        const testUsers = new Array<UserPostResponse>();
        const testGroups = new Array<string>();
        const createdAccessRules = new Array<AccessRule>();
        const createdProjects = new Array<Project>();
        const randomPrefix = Utils.generateShortUuid().slice(0, 6);

        let roleForTest: Role;
        let defaultDepartment: Department;
        let testProject: Project;

        test.beforeAll(async function () {
            // Set up API tokens
            const systemAdminToken = flowRootState.users.system_administrator.accessToken;
            authorizationApi.accessRulesApi.accessToken = systemAdminToken;
            orgUnitService.projectApi.accessToken = systemAdminToken;

            // Get roles and default department
            roleForTest = flowRootState.users.department_admin.role;
            defaultDepartment = await orgUnitUtils.getDefaultDepartment();
            const defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);

            // Create test users
            logger.info('Creating test users');
            for (let i = 0; i < 3; i++) {
                const userEmail = objectGenerator.generateEmailByPrefix(`bulk-user-${i}-${randomPrefix}`);
                const user = (await usersManager.createLocalUser(systemAdminToken, userEmail, false)).body;
                testUsers.push(user);
            }

            // Create test SSO groups
            logger.info('Creating test SSO groups');
            for (let i = 0; i < 2; i++) {
                const groupName = `bulk-group-${i}-${randomPrefix}`;
                testGroups.push(groupName);
            }

            // Create test project
            logger.info('Creating test project');
            const projectName = objectGenerator.generateNameByPrefix(`bulk-project-${randomPrefix}`);
            const namespace = Utils.namespaceFromProjectName(projectName);
            testProject = await orgUnitUtils.createProject({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                parentId: defaultDepartment.id,
                requestedNamespace: namespace,
                resources: [
                    {
                        nodePool: { name: 'default', id: defaultNodePoolId },
                        gpu: { deserved: 1 },
                        cpu: null,
                        memory: null
                    }
                ]
            });
            createdProjects.push(testProject);
        });

        test.afterAll(async function () {
            // Clean up access rules
            logger.info('Cleaning up access rules');
            for (const accessRule of createdAccessRules) {
                try {
                    await authorizationApi.accessRulesApi.deleteAccessRule(accessRule.id);
                } catch (error) {
                    logger.warn(`Failed to delete access rule ${accessRule.id}: ${error}`);
                }
            }

            // Clean up users
            logger.info('Cleaning up users');
            const systemAdminToken = flowRootState.users.system_administrator.accessToken;
            for (const user of testUsers) {
                try {
                    await usersManager.deleteLocalUser(systemAdminToken, user.id);
                } catch (error) {
                    logger.warn(`Failed to delete user ${user.id}: ${error}`);
                }
            }

            // Clean up projects
            logger.info('Cleaning up projects');
            for (const project of createdProjects) {
                try {
                    await orgUnitUtils.deleteProject(project.id);
                } catch (error) {
                    logger.warn(`Failed to delete project ${project.id}: ${error}`);
                }
            }
        });

        test.beforeEach(async function () {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            appPage = new App(browserEngine.currentPage);
            accessRulesPage = new AccessRulesPage(browserEngine.currentPage);
            videoName = `access-rules-bulk-${Date.now()}`;
        });

        test.afterEach(async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test('should create and delete multiple access rules via UI', async () => {
            logger.info('Testing access rules bulk create and delete via UI');

            await authPage.signIn(config.controlPlaneAdminUsername, config.controlPlaneAdminPassword);
            await appPage.page.waitForLoadState('networkidle');

            await accessRulesPage.goToPage();
            await expect(accessRulesPage.page.locator(accessRulesPage.pageTitle)).toBeVisible();

            await test.step('Create access rules for users with project scope', async () => {
                logger.info('Creating access rules for users with project scope');
                await accessRulesPage.clickNewAccessRule();
                await accessRulesPage.selectSubjectType('User');

                const userEmails = testUsers.map((user) => user.username);
                await accessRulesPage.addMultipleSubjects(userEmails);
                await accessRulesPage.selectRole(roleForTest.name);
                await accessRulesPage.setScope(Scope.Project, testProject.name);
                await accessRulesPage.saveAccessRule();

                // Verify access rules were created
                for (const email of userEmails) {
                    const userAccessRules = await authorizationUtils.getAccessRulesByUserName(email);
                    expect(userAccessRules).toHaveLength(1);
                    expect(userAccessRules[0].scopeType).toBe(ScopeType.Project);
                }
            });

            await test.step('Create access rules for groups with department scope', function () {
                shouldRunTests(settings.enableSSO)(async function () {
                    logger.info('Creating access rules for groups with department scope');
                    await accessRulesPage.clickNewAccessRule();
                    await accessRulesPage.selectSubjectType('Group');
                    await accessRulesPage.addMultipleSubjects(testGroups);
                    await accessRulesPage.selectRole(roleForTest.name);
                    await accessRulesPage.setScope(Scope.Department, defaultDepartment.name);
                    await accessRulesPage.saveAccessRule();

                    // Verify group access rules were created
                    for (const groupName of testGroups) {
                        await expect(
                            accessRulesPage.page.locator(`td:has-text("${groupName}")`)
                        ).toBeVisible();
                    }
                });
            });

            await test.step('Create access rules for user with tenant scope', async () => {
                logger.info('Creating access rules for mixed subjects with tenant scope');
                await accessRulesPage.clickNewAccessRule();
                await accessRulesPage.selectSubjectType('User');
                await accessRulesPage.addSubject(testUsers[0].username);
                await accessRulesPage.selectRole(roleForTest.name);
                await accessRulesPage.setScope(Scope.Tenant);
                await accessRulesPage.saveAccessRule();
                // Verify access rule for tenant user was created
                const userAccessRules = await authorizationUtils.getAccessRulesByUserName(
                    testUsers[0].username
                );
                // validate that one access rule was created for the tenant scope with scopeType= 'Tenant'
                const tenantScopeRule = userAccessRules.find((rule) => rule.scopeType === ScopeType.Tenant);
                expect(tenantScopeRule).toBeDefined();
            });

            await test.step('Bulk delete access rules', async () => {
                logger.info('Testing bulk delete of access rules by selecting specific user emails');

                await accessRulesPage.sortByCreationTimeAndSelectNewest();

                const userEmails = testUsers.map((user) => user.username);
                await accessRulesPage.selectAccessRulesByUserEmails(userEmails);

                // If SSO is enabled, also select group-based access rules
                if (settings.enableSSO && testGroups.length > 0) {
                    await accessRulesPage.selectAccessRulesByUserEmails(testGroups);
                }

                // Verify some access rules are selected using role-based selectors
                const selectedCheckboxes = accessRulesPage.page.getByRole('checkbox', { checked: true });
                const selectedCount = await selectedCheckboxes.count();
                expect(selectedCount).toBeGreaterThan(0);
                logger.info(`Selected ${selectedCount} access rules for deletion`);

                // Delete the selected access rules using the updated method that handles the confirmation modal correctly
                await accessRulesPage.deleteSelectedAccessRules();

                logger.info('Bulk delete completed successfully');
            });

            logger.info('Successfully completed bulk create and delete test');
        });
    }
);
