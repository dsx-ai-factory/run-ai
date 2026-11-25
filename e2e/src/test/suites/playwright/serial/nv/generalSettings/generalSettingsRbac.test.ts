/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { StatusCodes } from 'http-status-codes';
import { test, expect } from '@playwright/test';
import {
    apiUtils,
    authorizationUtils,
    flowRootState,
    logger,
    tags,
    TokenObject,
    usersManager
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { GeneralSettings } from '../../../../../../modules/ui/runaiV2/pages/GeneralSettings';
import { authorizationApi, bffServiceApi, identityManagerApi } from '../../../../../../modules/api';
import { PermissionAction } from '../../../../../../modules/api/apis/authorizationApi/generated/model/permissionAction';
import { ResourceType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/resourceType';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { SettingsKeyEnum } from '../../../../../../modules/api/apis/identityManagerApi/generated/model/settingsKeyEnum';
import { AuthorizationSettingsKeyEnum } from '../../../../../../modules/api/apis/authorizationApi/generated/model/authorizationSettingsKeyEnum';
import {
    BrandingSettingsKeyEnum,
    GetLogo,
    NotificationsSettingsKeyEnum
} from '../../../../../../modules/api/apis/bffServiceApi';
import { BASELINE_RESOURCES, CRUD_ACTIONS, SETTINGS_RESOURCES } from './generalSettingsRbac.consts';
import { EFeatureFlagsKeys, EFeatureFlagsNames } from '../../../../../../modules/helpers/models/featureFlags';

test.describe(`General Settings per-section RBAC, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.roles}`, () => {
    supportedFrom(SupportedVersions.V2_26)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let generalSettings: GeneralSettings;
        let tenantId: string;
        let videoName: string;
        let viewerToken: TokenObject;
        let viewerUser: UserPostResponse;
        let brandingToken: TokenObject;
        let brandingUser: UserPostResponse;
        let originalBrandingValue: string;
        let mixedToken: TokenObject;
        let mixedUser: UserPostResponse;
        let analyticsOriginal: boolean;
        let brandingNeedsRestore = false;
        let analyticsNeedsRestore = false;
        let analyticsOriginalEnabled: boolean;

        const createdRoleIds: number[] = [];
        const createdUserIds: string[] = [];

        test.beforeAll(function () {
            browserEngine = new BrowserEngine();
            tenantId = flowRootState.cluster.tenantId.toString();
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            generalSettings = new GeneralSettings(browserEngine.currentPage);
        });

        test.afterEach(async function ({}, testInfo) {
            if (analyticsNeedsRestore) {
                if (analyticsOriginal) {
                    await apiUtils.enableLimitOverQuotaSubscription();
                } else {
                    await apiUtils.disableLimitOverQuotaSubscription();
                }
                analyticsNeedsRestore = false;
            }

            await browserEngine.stopAndArchive(testInfo, videoName);

            if (brandingNeedsRestore) {
                try {
                    bffServiceApi.settingsApi.accessToken = brandingToken.accessToken;
                    await bffServiceApi.settingsApi.updateBrandingSettingByKey(BrandingSettingsKeyEnum.Logo, {
                        logo: originalBrandingValue
                    });
                } catch (error) {
                    logger.error(`Failed to restore branding logo: ${error?.message}`);
                }
                brandingNeedsRestore = false;
            }

            while (createdUserIds.length > 0) {
                const userId = createdUserIds.pop();
                try {
                    await usersManager.deleteLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        userId
                    );
                } catch (error) {
                    if (error?.statusCode !== StatusCodes.NOT_FOUND) {
                        logger.error(`Failed to delete user ${userId}: ${error?.message}`);
                    }
                }
            }

            while (createdRoleIds.length > 0) {
                const roleId = createdRoleIds.pop();
                try {
                    await authorizationUtils.deleteRole(
                        roleId,
                        flowRootState.users.system_administrator.accessToken
                    );
                } catch (error) {
                    if (error?.statusCode !== StatusCodes.NOT_FOUND) {
                        logger.error(`Failed to delete role ${roleId}: ${error?.message}`);
                    }
                }
            }
        });

        test('Read-only settings user (READ all settings resources + CRUD registries) — sections render, edits return 403 (T6054220)', async () => {
            videoName = 'settings-rbac-viewer';

            await test.step('create viewer role + user + binding', async () => {
                const availablePermissions = await authorizationUtils.getUserPermissions(
                    flowRootState.users.system_administrator.accessToken
                );
                const writeResources = new Set<ResourceType>([ResourceType.Registries]);
                const readResources = new Set<ResourceType>([...SETTINGS_RESOURCES, ...BASELINE_RESOURCES]);
                const rolePermissions = authorizationUtils.buildRolePermissions(
                    availablePermissions,
                    readResources,
                    writeResources
                );

                expect(rolePermissions.map((permission) => permission.resourceType)).toEqual(
                    expect.arrayContaining([
                        ResourceType.Settings,
                        ResourceType.SecuritySettings,
                        ResourceType.BrandingSettings,
                        ResourceType.NotificationSettings,
                        ResourceType.Registries
                    ])
                );

                const { user, token, roleId } = await authorizationUtils.createUserWithCustomRole({
                    adminToken: flowRootState.users.system_administrator.accessToken,
                    rolePrefix: 'e2e-viewer',
                    description: 'READ all settings resources + CRUD registries',
                    permissions: rolePermissions,
                    tenantId,
                    logLabel: 'Viewer user'
                });
                createdRoleIds.push(roleId);
                createdUserIds.push(user.id);
                viewerUser = user;
                viewerToken = token;
            });

            await test.step('verify viewer effective permissions match read-only settings profile', async () => {
                const viewerPermissions = await authorizationUtils.getUserPermissions(
                    viewerToken.accessToken
                );

                expect(
                    authorizationUtils.getActionsForResource(viewerPermissions, ResourceType.Settings)
                ).toEqual([PermissionAction.Read]);
                expect(
                    authorizationUtils.getActionsForResource(viewerPermissions, ResourceType.SecuritySettings)
                ).toEqual([PermissionAction.Read]);
                expect(
                    authorizationUtils.getActionsForResource(viewerPermissions, ResourceType.BrandingSettings)
                ).toEqual([PermissionAction.Read]);
                expect(
                    authorizationUtils.getActionsForResource(
                        viewerPermissions,
                        ResourceType.NotificationSettings
                    )
                ).toEqual([PermissionAction.Read]);
                expect(
                    authorizationUtils
                        .getActionsForResource(viewerPermissions, ResourceType.Registries)
                        .sort()
                ).toEqual([...CRUD_ACTIONS].sort());
            });

            await test.step('login as viewer through the dashboard', async () => {
                await authPage.signIn(viewerUser.username, viewerUser.tempPassword);
            });

            await test.step('navigate to /settings', async () => {
                await generalSettings.hardNavigateToGeneralSettings();
            });

            await test.step('verify General Settings sections render', async () => {
                for (const selector of generalSettings.allSections) {
                    await expect(authPage.page.locator(selector)).toBeVisible();
                }
            });

            await test.step('verify visible sections are read-only (toggles disabled, no Save)', async () => {
                await authPage.page.locator(generalSettings.analyticsSection).click();
                await expect(
                    authPage.page.locator(generalSettings.limtQuotaOverSubscriptionToggle)
                ).toHaveAttribute('aria-disabled', 'true');

                await authPage.page.locator(generalSettings.resourcesSection).click();
                await expect(authPage.page.locator(generalSettings.cpuQuotaToggle)).toHaveAttribute(
                    'aria-disabled',
                    'true'
                );

                await expect(authPage.page.locator(generalSettings.saveButton)).toHaveCount(0);
            });

            await test.step('viewer cannot mutate security-settings (403)', async () => {
                identityManagerApi.settingsApi.accessToken = viewerToken.accessToken;
                const status = await authorizationUtils.getErrorStatusCode(() =>
                    identityManagerApi.settingsApi.updateSecuritySettingByKey(
                        SettingsKeyEnum.AutoRedirectSso,
                        { enabled: false }
                    )
                );
                expect(status, 'security-settings mutation should return 403').toEqual(StatusCodes.FORBIDDEN);
            });

            await test.step('viewer cannot mutate branding-settings (403)', async () => {
                bffServiceApi.settingsApi.accessToken = viewerToken.accessToken;
                const status = await authorizationUtils.getErrorStatusCode(() =>
                    bffServiceApi.settingsApi.updateBrandingSettingByKey(BrandingSettingsKeyEnum.Logo, {
                        logo: 'rbac-test'
                    })
                );
                expect(status, 'branding-settings mutation should return 403').toEqual(StatusCodes.FORBIDDEN);
            });

            await test.step('viewer cannot mutate settings (403)', async () => {
                authorizationApi.settingsApi.accessToken = viewerToken.accessToken;
                const status = await authorizationUtils.getErrorStatusCode(() =>
                    authorizationApi.settingsApi.updateAuthorizationSettingByKey(
                        AuthorizationSettingsKeyEnum.EnableScopedAuthorization,
                        { enabled: false }
                    )
                );
                expect(status, 'settings mutation should return 403').toEqual(StatusCodes.FORBIDDEN);
            });

            await test.step('viewer cannot update notification settings (403)', async () => {
                bffServiceApi.settingsApi.accessToken = viewerToken.accessToken;
                const status = await authorizationUtils.getErrorStatusCode(() =>
                    bffServiceApi.settingsApi.updateNotificationsSettingByKey(
                        NotificationsSettingsKeyEnum.AdminMessage,
                        { msg: 'rbac-test', enableDismiss: true }
                    )
                );
                expect(status, 'notification-settings mutation should return 403').toEqual(
                    StatusCodes.FORBIDDEN
                );
            });
        });

        test('Branding-only admin (CRUD branding-settings) — only Branding visible + editable, save round-trips (T6054221)', async () => {
            videoName = 'settings-rbac-branding';

            await test.step('create branding admin role + user + binding', async () => {
                const availablePermissions = await authorizationUtils.getUserPermissions(
                    flowRootState.users.system_administrator.accessToken
                );
                const writeResources = new Set<ResourceType>([ResourceType.BrandingSettings]);
                const readResources = new Set<ResourceType>([...SETTINGS_RESOURCES, ...BASELINE_RESOURCES]);
                const rolePermissions = authorizationUtils.buildRolePermissions(
                    availablePermissions,
                    readResources,
                    writeResources
                );

                expect(rolePermissions.map((permission) => permission.resourceType)).toEqual(
                    expect.arrayContaining([ResourceType.BrandingSettings])
                );

                const { user, token, roleId } = await authorizationUtils.createUserWithCustomRole({
                    adminToken: flowRootState.users.system_administrator.accessToken,
                    rolePrefix: 'e2e-branding-admin',
                    description: 'CRUD on branding-settings',
                    permissions: rolePermissions,
                    tenantId,
                    logLabel: 'Branding-only admin'
                });
                createdRoleIds.push(roleId);
                createdUserIds.push(user.id);
                brandingUser = user;
                brandingToken = token;
            });

            await test.step('verify branding admin effective permissions match CRUD branding profile', async () => {
                const brandingPermissions = await authorizationUtils.getUserPermissions(
                    brandingToken.accessToken
                );

                expect(
                    authorizationUtils
                        .getActionsForResource(brandingPermissions, ResourceType.BrandingSettings)
                        .sort()
                ).toEqual([...CRUD_ACTIONS].sort());
                expect(
                    authorizationUtils.getActionsForResource(brandingPermissions, ResourceType.Settings)
                ).toEqual([PermissionAction.Read]);
                expect(
                    authorizationUtils.getActionsForResource(
                        brandingPermissions,
                        ResourceType.SecuritySettings
                    )
                ).toEqual([PermissionAction.Read]);
                expect(
                    authorizationUtils.getActionsForResource(
                        brandingPermissions,
                        ResourceType.NotificationSettings
                    )
                ).toEqual([PermissionAction.Read]);
                expect(
                    authorizationUtils.getActionsForResource(brandingPermissions, ResourceType.Registries)
                ).toEqual([PermissionAction.Read]);
            });

            await test.step('login as branding admin through the dashboard', async () => {
                await authPage.signIn(brandingUser.username, brandingUser.tempPassword);
            });

            await test.step('navigate to /settings and verify all sections render', async () => {
                await generalSettings.hardNavigateToGeneralSettings();
                for (const selector of generalSettings.allSections) {
                    await expect(authPage.page.locator(selector)).toBeVisible();
                }
            });

            await test.step('verify Branding section is editable (logo button present via aid)', async () => {
                await generalSettings.clickonBranding();
                const addLogo = authPage.page.locator(generalSettings.addLogoButton);
                const editLogo = authPage.page.locator(generalSettings.brandingEditLogoButton);
                const buttonCount = (await addLogo.count()) + (await editLogo.count());
                expect(buttonCount).toBeGreaterThan(0);
            });

            await test.step('record original branding logo via API', async () => {
                bffServiceApi.settingsApi.accessToken = brandingToken.accessToken;
                const response = await bffServiceApi.settingsApi.getBrandingSettingByKey(
                    BrandingSettingsKeyEnum.Logo
                );
                expect(response.response.statusCode).toBe(StatusCodes.OK);
                originalBrandingValue = (response.body as GetLogo).value?.logo ?? '';
                brandingNeedsRestore = true;
            });

            await test.step('update branding logo via API as branding admin', async () => {
                bffServiceApi.settingsApi.accessToken = brandingToken.accessToken;
                const response = await bffServiceApi.settingsApi.updateBrandingSettingByKey(
                    BrandingSettingsKeyEnum.Logo,
                    { logo: generalSettings.brandingTestLogoDataUrl }
                );
                expect(response.response.statusCode).toBeLessThan(300);
            });

            await test.step('hard reload /settings and verify logo persisted (UI aid + API value)', async () => {
                await generalSettings.hardNavigateToGeneralSettings();
                await generalSettings.clickonBranding();
                await expect(authPage.page.locator(generalSettings.brandingEditLogoButton)).toBeVisible();

                bffServiceApi.settingsApi.accessToken = brandingToken.accessToken;
                const response = await bffServiceApi.settingsApi.getBrandingSettingByKey(
                    BrandingSettingsKeyEnum.Logo
                );
                expect((response.body as GetLogo).value?.logo).toBe(generalSettings.brandingTestLogoDataUrl);
            });
        });

        test('Mixed role (READ security-settings + CRUD settings) — Security read-only, general sections editable (T6054222)', async () => {
            videoName = 'settings-rbac-mixed';

            await test.step('create mixed role + user + binding', async () => {
                const availablePermissions = await authorizationUtils.getUserPermissions(
                    flowRootState.users.system_administrator.accessToken
                );
                const writeResources = new Set<ResourceType>([ResourceType.Settings]);
                const readResources = new Set<ResourceType>([...SETTINGS_RESOURCES, ...BASELINE_RESOURCES]);
                const rolePermissions = authorizationUtils.buildRolePermissions(
                    availablePermissions,
                    readResources,
                    writeResources
                );

                expect(rolePermissions.map((permission) => permission.resourceType)).toEqual(
                    expect.arrayContaining([ResourceType.Settings, ResourceType.SecuritySettings])
                );

                const { user, token, roleId } = await authorizationUtils.createUserWithCustomRole({
                    adminToken: flowRootState.users.system_administrator.accessToken,
                    rolePrefix: 'e2e-mixed-sec-r-set-crud',
                    description: 'READ security-settings + CRUD settings',
                    permissions: rolePermissions,
                    tenantId,
                    logLabel: 'Mixed-role user'
                });
                createdRoleIds.push(roleId);
                createdUserIds.push(user.id);
                mixedUser = user;
                mixedToken = token;
            });

            await test.step('verify mixed-role effective permissions match security-read + settings-crud profile', async () => {
                const mixedPermissions = await authorizationUtils.getUserPermissions(mixedToken.accessToken);

                expect(
                    authorizationUtils.getActionsForResource(mixedPermissions, ResourceType.SecuritySettings)
                ).toEqual([PermissionAction.Read]);
                expect(
                    authorizationUtils.getActionsForResource(mixedPermissions, ResourceType.Settings).sort()
                ).toEqual([...CRUD_ACTIONS].sort());
                expect(
                    authorizationUtils.getActionsForResource(mixedPermissions, ResourceType.BrandingSettings)
                ).toEqual([PermissionAction.Read]);
                expect(
                    authorizationUtils.getActionsForResource(
                        mixedPermissions,
                        ResourceType.NotificationSettings
                    )
                ).toEqual([PermissionAction.Read]);
                expect(
                    authorizationUtils.getActionsForResource(mixedPermissions, ResourceType.Registries)
                ).toEqual([PermissionAction.Read]);
            });

            await test.step('login as mixed-role user', async () => {
                await authPage.signIn(mixedUser.username, mixedUser.tempPassword);
            });

            await test.step('navigate to /settings', async () => {
                await generalSettings.hardNavigateToGeneralSettings();
            });

            await test.step('verify all General Settings sections render', async () => {
                for (const selector of generalSettings.allSections) {
                    await expect(authPage.page.locator(selector)).toBeVisible();
                }
            });

            await test.step('verify Security section is read-only (add IdP button disabled)', async () => {
                await authPage.page.locator(generalSettings.securityOption).click();
                await expect(authPage.page.locator(generalSettings.addIdentityProvider)).toHaveAttribute(
                    'aria-disabled',
                    'true'
                );
            });

            await test.step('backend reinforcement — security-settings PUT is 403', async () => {
                identityManagerApi.settingsApi.accessToken = mixedToken.accessToken;
                const status = await authorizationUtils.getErrorStatusCode(() =>
                    identityManagerApi.settingsApi.updateSecuritySettingByKey(
                        SettingsKeyEnum.AutoRedirectSso,
                        { enabled: false }
                    )
                );
                expect(status, 'security-settings mutation should return 403').toEqual(StatusCodes.FORBIDDEN);
            });

            await test.step('record current Analytics toggle state', async () => {
                analyticsOriginal = (await apiUtils.getFeatureValue(
                    EFeatureFlagsKeys.ENABLE_LIMIT_OQ_SUBSCRIPTION,
                    EFeatureFlagsNames.ENABLE_LIMIT_OQ_SUBSCRIPTION
                )) as boolean;
            });

            await test.step('flip the Analytics toggle and save', async () => {
                if (analyticsOriginal) {
                    await apiUtils.disableLimitOverQuotaSubscription(mixedToken.accessToken);
                    analyticsOriginalEnabled = false;
                } else {
                    await apiUtils.enableLimitOverQuotaSubscription(mixedToken.accessToken);
                    analyticsOriginalEnabled = true;
                }
                analyticsNeedsRestore = true;
            });

            await test.step('hard reload /settings and verify persisted toggle', async () => {
                if (analyticsOriginalEnabled) {
                    const actualState = await apiUtils.getFeatureValue(
                        EFeatureFlagsKeys.ENABLE_LIMIT_OQ_SUBSCRIPTION,
                        EFeatureFlagsNames.ENABLE_LIMIT_OQ_SUBSCRIPTION
                    );
                    expect(actualState).toBeTruthy();
                } else {
                    const actualState = await apiUtils.getFeatureValue(
                        EFeatureFlagsKeys.ENABLE_LIMIT_OQ_SUBSCRIPTION,
                        EFeatureFlagsNames.ENABLE_LIMIT_OQ_SUBSCRIPTION
                    );
                    expect(actualState).toBeFalsy();
                }
            });

            await test.step('verify persistence through API', async () => {
                authorizationApi.settingsApi.accessToken = mixedToken.accessToken;
                const response = await authorizationApi.settingsApi.getAuthorizationSettings();
                expect(response.response.statusCode).toBe(StatusCodes.OK);
            });
        });
    });
});
