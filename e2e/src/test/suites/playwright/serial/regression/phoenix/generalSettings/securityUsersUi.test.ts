/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authorizationUtils,
    authV2Utils,
    flowRootState,
    objectGenerator,
    tags,
    logger,
    usersManager
} from '../../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { GeneralSettings } from '../../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { test } from '@playwright/test';
import {
    Action,
    ResourceType,
    ResourceTypeGroupId,
    Role,
    RoleCreationFields
} from '../../../../../../../modules/api/apis/authorizationApi/generated/api';
import { expect } from '@playwright/test';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { settings } from '../../../../../../../modules/helpers/config/config';


        test.describe(`General Security Users Separation, tags: ${tags.tool.ui} ${tags.teams.phoenix} ${tags.suites.legacySphinx}`, () => {
            shouldRunTests(settings.isSelfHosted && !isOCP())(
                function () {
            let browserEngine: BrowserEngine;
            let auth: AuthPage;
            let generalSettings: GeneralSettings;
            let videoName: string;
            let securityAdminRole: Role;
            let generalSettingsAdminRole: Role;
            test.beforeEach('init browsers and classes', async function () {
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                auth = new AuthPage(browserEngine.currentPage);
                generalSettings = new GeneralSettings(browserEngine.currentPage);
            });

            test.afterEach('stop browserEngine', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, videoName);
            });

            test('should validate security admin user can see only security settings', async function () {
                videoName = 'validateSecurityAdminUserCanSeeOnlySecuritySettings';
                securityAdminRole = await authorizationUtils.createCustomRole(
                    `Security-Admin-Role-Test-${Date.now()}`,
                    'Can manage security settings only',
                    [
                        {
                            resourceType: ResourceType.SecuritySettings,
                            actions: [Action.Read, Action.Update, Action.Create, Action.Delete],
                            displayName: 'Security Settings',
                            groupId: ResourceTypeGroupId.Organization
                        }
                    ],
                    undefined,
                    flowRootState.users.system_administrator.accessToken,
                    RoleCreationFields.ScopeTypeEnum.Tenant,
                    undefined,
                    true
                );
                await authorizationUtils.activateRole(
                    securityAdminRole,
                    { enabled: true },
                    flowRootState.users.system_administrator.accessToken
                );
                const createUser = await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    objectGenerator.generateEmailByPrefix('securitySettingsAdminUser'),
                    false
                );
                const securitySettingsAdminUser = await usersManager.getUserById(
                    flowRootState.users.system_administrator.accessToken,
                    createUser.body.id
                );
                logger.debug(`user: ${securitySettingsAdminUser.username} created`);
                const accessRuleViewer = await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    securitySettingsAdminUser.username,
                    flowRootState.users.viewer.role.id
                );
                const accessRuleSecurityAdmin = await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    securitySettingsAdminUser.username,
                    securityAdminRole.id
                );
                await auth.signIn(securitySettingsAdminUser.username, createUser.body.tempPassword);
                await generalSettings.openGeneralSettingsPage();
                expect(await generalSettings.validateOnlySecuritySettingsVisible()).toBe(true);
                try {
                    await authorizationUtils.deleteAccessRule(
                        accessRuleViewer,
                        flowRootState.users.system_administrator.accessToken
                    );
                    await authorizationUtils.deleteAccessRule(
                        accessRuleSecurityAdmin,
                        flowRootState.users.system_administrator.accessToken
                    );
                    await usersManager.deleteLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        securitySettingsAdminUser.id
                    );
                } catch (error) {
                    logger.error(`Error deleting access rules: ${error.message}`);
                }
                await authorizationUtils.deleteRole(
                    securityAdminRole.id,
                    flowRootState.users.system_administrator.accessToken
                );
            });

            test('should validate general settings admin user can see all settings except security settings', async function () {
                videoName = 'validateGeneralSettingsAdminUserCanSeeAllSettings';
                generalSettingsAdminRole = await authorizationUtils.createCustomRole(
                    `General-Settings-Admin-Role-Test-${Date.now()}`,
                    'Can manage general settings',
                    [
                        {
                            resourceType: ResourceType.Settings,
                            actions: [Action.Read, Action.Update, Action.Create, Action.Delete],
                            displayName: 'General Settings',
                            groupId: ResourceTypeGroupId.Organization
                        }
                    ],
                    undefined,
                    flowRootState.users.system_administrator.accessToken,
                    RoleCreationFields.ScopeTypeEnum.Tenant,
                    undefined,
                    true
                );
                await authorizationUtils.activateRole(
                    generalSettingsAdminRole,
                    { enabled: true },
                    flowRootState.users.system_administrator.accessToken
                );
                const createUser = await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    objectGenerator.generateEmailByPrefix('generalSettingsAdminUser'),
                    false
                );
                const generalSettingsAdminUser = await usersManager.getUserById(
                    flowRootState.users.system_administrator.accessToken,
                    createUser.body.id
                );
                logger.debug(`user: ${generalSettingsAdminUser.username} created`);
                const accessRuleViewer = await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    generalSettingsAdminUser.username,
                    flowRootState.users.viewer.role.id
                );
                const accessRuleGeneralSettingsAdmin = await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    generalSettingsAdminUser.username,
                    generalSettingsAdminRole.id
                );
                await auth.signIn(generalSettingsAdminUser.username, createUser.body.tempPassword);
                await generalSettings.openGeneralSettingsPage();
                expect(await generalSettings.ValidateAllSettingsVisibleExceptSecurity()).toBe(true);
                try {
                    await authorizationUtils.deleteAccessRule(
                        accessRuleViewer,
                        flowRootState.users.system_administrator.accessToken
                    );
                    await authorizationUtils.deleteAccessRule(
                        accessRuleGeneralSettingsAdmin,
                        flowRootState.users.system_administrator.accessToken
                    );
                    await usersManager.deleteLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        generalSettingsAdminUser.id
                    );
                } catch (error) {
                    logger.error(`Error deleting access rules: ${error.message}`);
                }
                await authorizationUtils.deleteRole(
                    generalSettingsAdminRole.id,
                    flowRootState.users.system_administrator.accessToken
                );
            });
        });
    }
);
