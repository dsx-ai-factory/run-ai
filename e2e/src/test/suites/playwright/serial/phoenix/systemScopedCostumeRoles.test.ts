/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import { expect } from 'chai';
import {
    authorizationUtils,
    authUtils,
    config,
    logger,
    objectGenerator,
    settings,
    tags,
    tenantUtils
} from '../../../../../modules/helpers';
import {
    shouldRunTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../modules/helpers/config/roleMapping';
import { Keycloak } from '../../../../../modules/helpers/utils/Keycloak';
import { TenantCreationResponse } from '../../../../../modules/api/apis/tenantsManagerApi/generated/model/tenantCreationResponse';
import { Tenant } from '../../../../../modules/api/apis/tenantsManagerApi/generated/model/tenant';
import { TokenResponse } from '../../../../../modules/api/apis/identityManagerApi/generated/model/tokenResponse';
import { ResourceType } from '../../../../../modules/api/apis/authorizationApi/generated/model/resourceType';
import { Action } from '../../../../../modules/api/apis/authorizationApi/generated/model/action';
import { ResourceTypeGroupId } from '../../../../../modules/api/apis/authorizationApi/generated/model/resourceTypeGroupId';
import { Permission } from '../../../../../modules/api/apis/authorizationApi/generated/model/permission';
import { Role } from '../../../../../modules/api/apis/authorizationApi/generated/model/role';
import { RoleCreationFields } from '../../../../../modules/api/apis/authorizationApi/generated/model/roleCreationFields';
import { RolesApi } from '../../../../../modules/api/apis/authorizationApi/generated/api/rolesApi';
import { TenantApi } from '../../../../../modules/api/apis/backendApi/generated/api/tenantApi';
import { AccessRule } from '../../../../../modules/api/apis/authorizationApi/generated/model/accessRule';
import { AccessRulesApi } from '../../../../../modules/api/apis/authorizationApi/generated/api/accessRulesApi';
import { SubjectType } from '../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { RoleActiveFields } from '../../../../../modules/api/apis/authorizationApi/generated/model/roleActiveFields';
import { ScopeType } from '../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';

test.describe(
    `Costume Roles In System Scope Tests, tags: ${tags.teams.viper} ${tags.resource.roles}`,
    { tag: [tags.teams.viper, tags.resource.roles, tags.runOn.serialCI] },
    function () {
        supportedFrom(SupportedVersions.V2_23)(function () {
            shouldRunTests(settings.isSelfHosted)(function () {
                interface TenantInfo {
                    tenant: TenantCreationResponse;
                    tenantName: string;
                    tenantAdminEmail: string;
                    tenantAdminToken: TokenResponse;
                    tenantUrl: string;
                }

                const systemTenantName = 'console';
                let consoleToken: TokenResponse;
                let consoleTenantUrl: string;
                let systemTenant: Tenant;

                const tenant1Info: TenantInfo = {} as TenantInfo;
                const tenant2Info: TenantInfo = {} as TenantInfo;

                let adminPassword: string;

                let systemRoles: Role[] = new Array<Role>();
                let tenant1Roles: Role[] = new Array<Role>();
                let tenant1AccessRules: AccessRule[] = new Array<AccessRule>();

                let consoleRolesApi: RolesApi;
                let tenant1RolesApi: RolesApi;
                let tenant2RolesApi: RolesApi;

                let tenant1AccessRulesApi: AccessRulesApi;

                let consoleTenantsApi: TenantApi;
                let tenant1TenantsApi: TenantApi;
                let tenant2TenantsApi: TenantApi;

                let accessRuleUserName: string;

                const roleDescription = 'role description';

                async function deleteTenant(tenantInfo: TenantInfo): Promise<void> {
                    await tenantUtils.deleteTenantAndValidate(
                        consoleToken.accessToken,
                        consoleTenantUrl,
                        tenantInfo.tenant.tenant.id,
                        true,
                        tenantInfo.tenant.tenant.name
                    );
                }

                test.beforeAll(async function () {
                    // phoenix-playwright-setup runs in playwright's globalSetup
                    // process; workers re-import config/settings fresh and
                    // don't see the mutations to config.tenantUrl /
                    // settings.tenantClientSecret. Re-derive both here, matching
                    // the pattern in sibling phoenix playwright tests.
                    const baseUrl = config.tenantUrl.replace('https://', '');
                    consoleTenantUrl = `https://console.${baseUrl}`;
                    settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient(
                        'console',
                        'default'
                    );
                    adminPassword = config.controlPlaneAdminPassword;

                    tenant1Info.tenantAdminEmail = objectGenerator.generateEmailByPrefix('admin1');
                    tenant2Info.tenantAdminEmail = objectGenerator.generateEmailByPrefix('admin2');

                    const tenant1Name = objectGenerator.generateShortNameByPrefix('tenant1');
                    const tenant2Name = objectGenerator.generateShortNameByPrefix('tenant2');
                    tenant1Info.tenantName = tenant1Name;
                    tenant2Info.tenantName = tenant2Name;

                    consoleToken = await authUtils.authenticateSuperAdminAppAndGetToken(
                        'default',
                        settings.tenantClientSecret,
                        consoleTenantUrl
                    );

                    const foundSystemTenant = (
                        await tenantUtils.getTenants(consoleToken.accessToken, consoleTenantUrl)
                    ).find((tenant) => tenant.name === systemTenantName);
                    if (!foundSystemTenant) {
                        throw new Error(`System tenant "${systemTenantName}" was not found`);
                    }
                    systemTenant = foundSystemTenant;

                    consoleRolesApi = new RolesApi(consoleTenantUrl);
                    consoleRolesApi.accessToken = consoleToken.accessToken;
                    consoleTenantsApi = new TenantApi(consoleTenantUrl);
                    consoleTenantsApi.accessToken = consoleToken.accessToken;

                    const [tenant1Data, tenant2Data] = await Promise.all([
                        tenantUtils.createTenantAndRetrieveToken(
                            consoleToken.accessToken,
                            consoleTenantUrl,
                            config.appUrl,
                            tenant1Name,
                            tenant1Info.tenantAdminEmail,
                            adminPassword
                        ),
                        tenantUtils.createTenantAndRetrieveToken(
                            consoleToken.accessToken,
                            consoleTenantUrl,
                            config.appUrl,
                            tenant2Name,
                            tenant2Info.tenantAdminEmail,
                            adminPassword
                        )
                    ]);

                    tenant1Info.tenant = tenant1Data.tenant;
                    tenant1Info.tenantAdminToken = tenant1Data.adminToken;
                    tenant1Info.tenantUrl = tenant1Data.tenantSpecificUrl;

                    tenant2Info.tenant = tenant2Data.tenant;
                    tenant2Info.tenantAdminToken = tenant2Data.adminToken;
                    tenant2Info.tenantUrl = tenant2Data.tenantSpecificUrl;

                    tenant1RolesApi = new RolesApi(tenant1Info.tenantUrl);
                    tenant1RolesApi.accessToken = tenant1Info.tenantAdminToken.accessToken;
                    tenant1AccessRulesApi = new AccessRulesApi(tenant1Info.tenantUrl);
                    tenant1AccessRulesApi.accessToken = tenant1Info.tenantAdminToken.accessToken;
                    tenant1TenantsApi = new TenantApi(tenant1Info.tenantUrl);
                    tenant1TenantsApi.accessToken = tenant1Info.tenantAdminToken.accessToken;

                    tenant2RolesApi = new RolesApi(tenant2Info.tenantUrl);
                    tenant2RolesApi.accessToken = tenant2Info.tenantAdminToken.accessToken;
                    tenant2TenantsApi = new TenantApi(tenant2Info.tenantUrl);
                    tenant2TenantsApi.accessToken = tenant2Info.tenantAdminToken.accessToken;

                    accessRuleUserName = objectGenerator.generateEmailByPrefix('access-rule-user');
                });

                test.beforeEach(async function () {
                    const [tenant1TokenRes, tenant2TokenRes] = await Promise.all([
                        authUtils.authenticateTokenApi(
                            tenant1Info.tenantAdminEmail,
                            adminPassword,
                            tenant1Info.tenantUrl,
                            consoleToken.accessToken
                        ),
                        authUtils.authenticateTokenApi(
                            tenant2Info.tenantAdminEmail,
                            adminPassword,
                            tenant2Info.tenantUrl,
                            consoleToken.accessToken
                        )
                    ]);
                    tenant1Info.tenantAdminToken = tenant1TokenRes.body;
                    tenant2Info.tenantAdminToken = tenant2TokenRes.body;

                    tenant1RolesApi.accessToken = tenant1Info.tenantAdminToken.accessToken;
                    tenant1AccessRulesApi.accessToken = tenant1Info.tenantAdminToken.accessToken;
                    tenant1TenantsApi.accessToken = tenant1Info.tenantAdminToken.accessToken;
                    tenant2RolesApi.accessToken = tenant2Info.tenantAdminToken.accessToken;
                    tenant2TenantsApi.accessToken = tenant2Info.tenantAdminToken.accessToken;
                });

                test.afterAll(async function () {
                    const errors = [];
                    if (tenant1AccessRules && tenant1AccessRules.length > 0) {
                        for (const accessRule of tenant1AccessRules) {
                            try {
                                await authorizationUtils.deleteAccessRule(
                                    accessRule,
                                    tenant1Info.tenantAdminToken.accessToken,
                                    tenant1AccessRulesApi
                                );
                            } catch (error) {
                                logger.error(`Failed to delete accessRule ${accessRule.id}`);
                                errors.push(error);
                            }
                        }
                        tenant1AccessRules = new Array<AccessRule>();
                    }
                    if (systemRoles && systemRoles.length > 0) {
                        for (const role of systemRoles) {
                            try {
                                await authorizationUtils.deleteRole(
                                    role.id,
                                    consoleToken.accessToken,
                                    consoleRolesApi
                                );
                            } catch (error) {
                                logger.error(`Failed to delete role ${role.name}`);
                                errors.push(error);
                            }
                        }
                        systemRoles = new Array<Role>();
                    }
                    if (tenant1Roles && tenant1Roles.length > 0) {
                        for (const role of tenant1Roles) {
                            try {
                                await authorizationUtils.deleteRole(
                                    role.id,
                                    tenant1Info.tenantAdminToken.accessToken,
                                    tenant1RolesApi
                                );
                            } catch (error) {
                                logger.error(`Failed to delete role ${role.name}`);
                                errors.push(error);
                            }
                        }
                        tenant1Roles = new Array<Role>();
                    }
                    if (tenant1Info.tenant) {
                        try {
                            await deleteTenant(tenant1Info);
                        } catch (error) {
                            logger.error(`Failed to delete tenant ${tenant1Info.tenant.tenant.name}`);
                            errors.push(error);
                        }
                    }
                    if (tenant2Info.tenant) {
                        try {
                            await deleteTenant(tenant2Info);
                        } catch (error) {
                            logger.error(`Failed to delete tenant ${tenant2Info.tenant.tenant.name}`);
                            errors.push(error);
                        }
                    }
                    if (errors.length > 0) {
                        throw new Error(errors.join('\n'));
                    }
                });

                test.describe('role creation', function () {
                    function validateCreatePermissions(role: Role): void {
                        for (const permission of role.permissions) {
                            if (permission.resourceType === ResourceType.Project) {
                                expect(permission.actions.length, 'permission should have exactly 1 action').to.equal(1);
                                expect(permission.actions, 'permission actions should include Read').to.include(Action.Read);
                            } else {
                                expect(permission.actions.length, 'non-applicable resource permission should have no actions').to.equal(0);
                            }
                        }
                    }

                    test('Create a system-scoped role as system admin in console', async function () {
                        const roleName = objectGenerator.generateShortNameByPrefix('system-scope-role');
                        const systemTenantId = systemTenant.id + '';
                        const permissions = [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ];

                        const role = await authorizationUtils.createCustomRole(
                            roleName,
                            roleDescription,
                            permissions,
                            undefined,
                            consoleToken.accessToken,
                            RoleCreationFields.ScopeTypeEnum.System,
                            systemTenantId,
                            true,
                            consoleRolesApi
                        );
                        expect(role, 'created role should not be null').to.not.be.null;
                        systemRoles.push(role);

                        const systemRole = (
                            await authorizationUtils.getRoles(consoleToken.accessToken, consoleRolesApi)
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(systemRole, 'system role should be visible in console').to.not.be.null;
                        expect(systemRole.enabled, 'system role should be enabled').equal(true);
                        expect(systemRole.effectiveEnabled, 'system role should be effectively enabled in all tenants').equal(true);
                        validateCreatePermissions(systemRole);

                        let tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant1Info.tenantAdminToken.accessToken,
                                tenant1RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'system role should be visible in tenant').to.not.be.null;
                        expect(tenantRole.enabled, 'tenant role should be enabled').equal(true);
                        expect(tenantRole.effectiveEnabled, 'tenant role should be effectively enabled').equal(true);
                        validateCreatePermissions(tenantRole);

                        tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant2Info.tenantAdminToken.accessToken,
                                tenant2RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'system role should be visible in tenant').to.not.be.null;
                        expect(tenantRole.enabled, 'tenant role should be enabled').equal(true);
                        expect(tenantRole.effectiveEnabled, 'tenant role should be effectively enabled').equal(true);
                        validateCreatePermissions(tenantRole);
                    });

                    test('Create a tenant-scoped role as system admin in console', async function () {
                        const roleName = objectGenerator.generateShortNameByPrefix('system-role');

                        const permissions = [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ];

                        const role = await authorizationUtils.createCustomRole(
                            roleName,
                            roleDescription,
                            permissions,
                            undefined,
                            consoleToken.accessToken,
                            RoleCreationFields.ScopeTypeEnum.Tenant,
                            systemTenant.id + '',
                            true,
                            consoleRolesApi
                        );
                        expect(role, 'created role should not be null').to.not.be.null;
                        systemRoles.push(role);

                        const systemRole = (
                            await authorizationUtils.getRoles(consoleToken.accessToken, consoleRolesApi)
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(systemRole, 'system role should be visible in console').to.not.be.null;
                        expect(systemRole.enabled, 'system role should be enabled').equal(true);
                        expect(systemRole.effectiveEnabled, 'system role should be effectively enabled in all tenants').equal(true);
                        validateCreatePermissions(systemRole);

                        let tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant1Info.tenantAdminToken.accessToken,
                                tenant1RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'tenant-scoped role should not be visible cross-tenant').to.be.undefined;

                        tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant2Info.tenantAdminToken.accessToken,
                                tenant2RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'tenant-scoped role should not be visible cross-tenant').to.be.undefined;
                    });

                    test('Create a tenant-scoped role as tenant', async function () {
                        const roleName = objectGenerator.generateShortNameByPrefix('tenant1-role');

                        const permissions = [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ];

                        const role = await authorizationUtils.createCustomRole(
                            roleName,
                            roleDescription,
                            permissions,
                            undefined,
                            tenant1Info.tenantAdminToken.accessToken,
                            RoleCreationFields.ScopeTypeEnum.Tenant,
                            tenant1Info.tenant.tenant.id + '',
                            true,
                            tenant1RolesApi
                        );
                        expect(role, 'created role should not be null').to.not.be.null;
                        tenant1Roles.push(role);

                        const systemRole = (
                            await authorizationUtils.getRoles(consoleToken.accessToken, consoleRolesApi)
                        ).find((sysRole) => sysRole.id === role.id);
                        expect(systemRole, 'tenant-only role should not be visible in console').to.be.undefined;

                        let tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant1Info.tenantAdminToken.accessToken,
                                tenant1RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'system role should be visible in tenant').to.not.be.null;
                        expect(tenantRole.enabled, 'tenant role should be enabled').equal(true);
                        expect(tenantRole.effectiveEnabled, 'tenant role should be effectively enabled').equal(true);
                        validateCreatePermissions(tenantRole);

                        tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant2Info.tenantAdminToken.accessToken,
                                tenant2RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'tenant-scoped role should not be visible cross-tenant').to.be.undefined;
                    });

                    test('Create a system-scoped role as tenant', async function () {
                        const roleName = objectGenerator.generateShortNameByPrefix('sys-role-bad-request');

                        const permissions = [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ];

                        try {
                            const role = await authorizationUtils.createCustomRole(
                                roleName,
                                roleDescription,
                                permissions,
                                undefined,
                                tenant2Info.tenantAdminToken.accessToken,
                                RoleCreationFields.ScopeTypeEnum.Tenant,
                                tenant1Info.tenant.tenant.id + '',
                                true,
                                tenant1RolesApi
                            );
                            tenant1Roles.push(role);
                            expect.fail('Expect system-scoped role as tenant creation to fail.');
                        } catch (error) {
                            expect(error.response.statusCode).equal(
                                400,
                                `Error status is: ${error.response.status}, and not as expected: 400`
                            );

                            expect(error.response.body.message).equal(
                                `ScopeId ${tenant1Info.tenant.tenant.id} is different than subject tenant id`,
                                `Error message is: ${error.response.body.message}, and not as expected: ScopeId ${tenant1Info.tenant.tenant.id} is different than subject tenant id`
                            );
                        }
                    });
                });

                test.describe('role update', function () {
                    function validateUpdatePermissions(role: Role): void {
                        for (const permission of role.permissions) {
                            if (permission.resourceType === ResourceType.Department) {
                                expect(permission.actions.length, 'permission should have exactly 1 action').to.equal(1);
                                expect(permission.actions[0], 'Department permission first action should be Read').to.equal(Action.Read);
                            } else if (permission.resourceType === ResourceType.Project) {
                                expect(permission.actions.length, 'Project permission should have 4 actions after update').to.equal(4);
                                expect(permission.actions, 'permission actions should include Read').to.include(Action.Read);
                                expect(permission.actions, 'Project permission should include Update action').to.include(Action.Update);
                                expect(permission.actions, 'Project permission should include Create action').to.include(Action.Create);
                                expect(permission.actions, 'Project permission should include Delete action').to.include(Action.Delete);
                            } else {
                                expect(permission.actions.length, 'non-applicable resource permission should have no actions').to.equal(0);
                            }
                        }
                    }

                    test('Update a system-scoped role as system admin in console', async function () {
                        const roleName = objectGenerator.generateShortNameByPrefix('sys-role-update');

                        const permissions = [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ];
                        const role = await authorizationUtils.createCustomRole(
                            roleName,
                            roleDescription,
                            permissions,
                            undefined,
                            consoleToken.accessToken,
                            RoleCreationFields.ScopeTypeEnum.System,
                            systemTenant.id + '',
                            true,
                            consoleRolesApi
                        );
                        expect(role, 'created role should not be null').to.not.be.null;
                        systemRoles.push(role);

                        const permissionsUpdate = [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read, Action.Update, Action.Delete, Action.Create],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission,
                            {
                                resourceType: ResourceType.Department,
                                actions: [Action.Read],
                                displayName: 'Departments',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ];

                        const roleUpdate = await authorizationUtils.updateRole(
                            role,
                            {
                                name: role.name,
                                description: role.description,
                                permissions: permissionsUpdate,
                                kubernetesPermissions: role.kubernetesPermissions,
                                scopeType: role.scopeType,
                                scopeId: role.scopeId,
                                enabled: role.enabled
                            } as RoleCreationFields,
                            consoleToken.accessToken,
                            consoleRolesApi
                        );
                        expect(roleUpdate, 'role update should return the updated role').to.not.be.null;

                        const systemRole = (
                            await authorizationUtils.getRoles(consoleToken.accessToken, consoleRolesApi)
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(systemRole, 'system role should be visible in console').to.not.be.null;
                        expect(systemRole.enabled, 'system role should be enabled').equal(true);
                        expect(systemRole.effectiveEnabled, 'system role should be effectively enabled in all tenants').equal(true);
                        validateUpdatePermissions(systemRole);

                        let tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant1Info.tenantAdminToken.accessToken,
                                tenant1RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'system role should be visible in tenant').to.not.be.null;
                        expect(tenantRole.enabled, 'tenant role should be enabled').equal(true);
                        expect(tenantRole.effectiveEnabled, 'tenant role should be effectively enabled').equal(true);
                        validateUpdatePermissions(tenantRole);

                        tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant2Info.tenantAdminToken.accessToken,
                                tenant2RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'system role should be visible in tenant').to.not.be.null;
                        expect(tenantRole.enabled, 'tenant role should be enabled').equal(true);
                        expect(tenantRole.effectiveEnabled, 'tenant role should be effectively enabled').equal(true);
                        validateUpdatePermissions(tenantRole);
                    });

                    test('Update a tenant-scoped role with access rule as system admin in console', async function () {
                        const roleName = objectGenerator.generateShortNameByPrefix('sys-role-update-fail');

                        const permissions = [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ];
                        const role = await authorizationUtils.createCustomRole(
                            roleName,
                            roleDescription,
                            permissions,
                            undefined,
                            consoleToken.accessToken,
                            RoleCreationFields.ScopeTypeEnum.System,
                            systemTenant.id + '',
                            true,
                            consoleRolesApi
                        );
                        expect(role, 'created role should not be null').to.not.be.null;
                        systemRoles.push(role);

                        const accessRule = await authorizationUtils.createAccessRuleForUser(
                            tenant1Info.tenantAdminToken.accessToken,
                            accessRuleUserName,
                            role.id,
                            SubjectType.User,
                            tenant1Info.tenant.tenant.id.toString(),
                            ScopeType.Tenant,
                            tenant1AccessRulesApi
                        );
                        tenant1AccessRules.push(accessRule);

                        const permissionsUpdate = [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read, Action.Update, Action.Delete, Action.Create],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission,
                            {
                                resourceType: ResourceType.Department,
                                actions: [Action.Read],
                                displayName: 'Departments',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ];

                        try {
                            await authorizationUtils.updateRole(
                                role,
                                {
                                    name: role.name,
                                    description: role.description,
                                    permissions: permissionsUpdate,
                                    kubernetesPermissions: role.kubernetesPermissions,
                                    scopeType: role.scopeType,
                                    scopeId: role.scopeId,
                                    enabled: role.enabled
                                } as RoleCreationFields,
                                consoleToken.accessToken,
                                consoleRolesApi
                            );

                            expect.fail('Expect update role with access rule to fail.');
                        } catch (error) {
                            expect(error.response.statusCode).equal(
                                400,
                                `Error status is: ${error.response.status}, and not as expected: 400`
                            );

                            expect(error.response.body.message).equal(
                                'Can not update or delete a role that is currently in use',
                                `Error message is: ${error.response.body.message}, and not as expected: Can not update or delete a role that is currently in use`
                            );
                        }
                    });
                });

                test.describe('role deletion', function () {
                    test('Delete a system-scoped role as system admin', async function () {
                        const roleName = objectGenerator.generateShortNameByPrefix('sys-role-delete');

                        const permissions = [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ];
                        const role = await authorizationUtils.createCustomRole(
                            roleName,
                            roleDescription,
                            permissions,
                            undefined,
                            consoleToken.accessToken,
                            RoleCreationFields.ScopeTypeEnum.System,
                            systemTenant.id + '',
                            true,
                            consoleRolesApi
                        );
                        expect(role, 'created role should not be null').to.not.be.null;

                        let systemRole = (
                            await authorizationUtils.getRoles(consoleToken.accessToken, consoleRolesApi)
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(systemRole, 'system role should be visible in console').to.not.be.null;
                        expect(systemRole.enabled, 'system role should be enabled').equal(true);
                        expect(systemRole.effectiveEnabled, 'system role should be effectively enabled in all tenants').equal(true);

                        let tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant1Info.tenantAdminToken.accessToken,
                                tenant1RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'system role should be visible in tenant').to.not.be.null;
                        expect(tenantRole.enabled, 'tenant role should be enabled').equal(true);
                        expect(tenantRole.effectiveEnabled, 'tenant role should be effectively enabled').equal(true);

                        tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant2Info.tenantAdminToken.accessToken,
                                tenant2RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'system role should be visible in tenant').to.not.be.null;
                        expect(tenantRole.enabled, 'tenant role should be enabled').equal(true);
                        expect(tenantRole.effectiveEnabled, 'tenant role should be effectively enabled').equal(true);

                        const accessRule = await authorizationUtils.createAccessRuleForUser(
                            tenant1Info.tenantAdminToken.accessToken,
                            accessRuleUserName,
                            role.id,
                            SubjectType.User,
                            tenant1Info.tenant.tenant.id.toString(),
                            ScopeType.Tenant,
                            tenant1AccessRulesApi
                        );

                        try {
                            await authorizationUtils.deleteRole(
                                role.id,
                                consoleToken.accessToken,
                                consoleRolesApi
                            );
                            tenant1AccessRules.push(accessRule);
                            systemRoles.push(role);

                            expect.fail('Expect delete role with access rule to fail.');
                        } catch (error) {
                            expect(error.response.statusCode).equal(
                                400,
                                `Error status is: ${error.response.status}, and not as expected: 400`
                            );

                            expect(error.response.body.message).equal(
                                'Can not update or delete a role that is currently in use',
                                `Error message is: ${error.response.body.message}, and not as expected: Can not update or delete a role that is currently in use`
                            );
                        }

                        try {
                            await authorizationUtils.deleteAccessRule(
                                accessRule,
                                tenant1Info.tenantAdminToken.accessToken,
                                tenant1AccessRulesApi
                            );
                        } catch (error) {
                            tenant1AccessRules.push(accessRule);
                            systemRoles.push(role);

                            throw error;
                        }

                        try {
                            await authorizationUtils.deleteRole(
                                role.id,
                                consoleToken.accessToken,
                                consoleRolesApi
                            );
                        } catch (error) {
                            systemRoles.push(role);

                            throw error;
                        }

                        systemRole = (
                            await authorizationUtils.getRoles(consoleToken.accessToken, consoleRolesApi)
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(systemRole, 'deleted role should not be visible in console').to.be.undefined;

                        tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant1Info.tenantAdminToken.accessToken,
                                tenant1RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'deleted role should not be visible in tenant1').to.be.undefined;

                        tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant2Info.tenantAdminToken.accessToken,
                                tenant2RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'deleted role should not be visible in tenant2').to.be.undefined;
                    });
                });

                test.describe('disable roles', function () {
                    test.describe.configure({ mode: 'serial' });

                    let role: Role;
                    let accessRule: AccessRule;

                    test.beforeAll(async function () {
                        const roleName = objectGenerator.generateShortNameByPrefix('sys-role-update');

                        const permissions = [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ];
                        role = await authorizationUtils.createCustomRole(
                            roleName,
                            roleDescription,
                            permissions,
                            undefined,
                            consoleToken.accessToken,
                            RoleCreationFields.ScopeTypeEnum.System,
                            systemTenant.id + '',
                            true,
                            consoleRolesApi
                        );
                        expect(role, 'created role should not be null').to.not.be.null;
                    });

                    test.afterAll(async function () {
                        try {
                            await authorizationUtils.deleteRole(
                                role.id,
                                consoleToken.accessToken,
                                consoleRolesApi
                            );
                        } catch (error) {
                            logger.error(`Failed to delete role ${role.name}: ${error}`);
                        }
                    });

                    test('create access-rule for tenant1 and disable - fail', async function () {
                        accessRule = await authorizationUtils.createAccessRuleForUser(
                            tenant1Info.tenantAdminToken.accessToken,
                            accessRuleUserName,
                            role.id,
                            SubjectType.User,
                            tenant1Info.tenant.tenant.id.toString(),
                            ScopeType.Tenant,
                            tenant1AccessRulesApi
                        );
                        expect(accessRule, 'access rule should be created successfully').to.not.be.undefined;

                        const roleActiveFields = {
                            scopeType: RoleActiveFields.ScopeTypeEnum.Tenant,
                            enabled: false
                        };
                        try {
                            await authorizationUtils.activateRole(
                                role,
                                roleActiveFields,
                                tenant1Info.tenantAdminToken.accessToken,
                                tenant1RolesApi
                            );
                            expect.fail('Expect disable of role with access-rule to fail.');
                        } catch (error) {
                            expect(error.response.statusCode).equal(
                                400,
                                `Error status is: ${error.response.status}, and not as expected: 400`
                            );

                            expect(error.response.body.message).equal(
                                'Can not disable a role that is currently in use',
                                `Error message is: ${error.response.body.message}, and not as expected: Can not disable a role that is currently in use`
                            );
                        }
                    });

                    test('delete access-rule for tenant1, disable role - success, create access-rule - fail', async function () {
                        await authorizationUtils.deleteAccessRule(
                            accessRule,
                            tenant1Info.tenantAdminToken.accessToken,
                            tenant1AccessRulesApi
                        );

                        const roleActiveFields = {
                            scopeType: RoleActiveFields.ScopeTypeEnum.Tenant,
                            enabled: false
                        };
                        await authorizationUtils.activateRole(
                            role,
                            roleActiveFields,
                            tenant1Info.tenantAdminToken.accessToken,
                            tenant1RolesApi
                        );

                        const systemRole = (
                            await authorizationUtils.getRoles(consoleToken.accessToken, consoleRolesApi)
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(systemRole, 'system role should be visible in console').to.not.be.null;
                        expect(systemRole.enabled, 'system role should be enabled').equal(true);
                        expect(systemRole.effectiveEnabled, 'system role should be effectively enabled in all tenants').equal(true);

                        let tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant1Info.tenantAdminToken.accessToken,
                                tenant1RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'system role should be visible in tenant').to.not.be.null;
                        expect(tenantRole.enabled, 'tenant role should be enabled').equal(true);
                        expect(tenantRole.effectiveEnabled, 'tenant role should be effectively disabled after tenant disable').equal(false);

                        tenantRole = (
                            await authorizationUtils.getRoles(
                                tenant2Info.tenantAdminToken.accessToken,
                                tenant2RolesApi
                            )
                        ).find((sysRole) => sysRole.id === role.id);

                        expect(tenantRole, 'system role should be visible in tenant').to.not.be.null;
                        expect(tenantRole.enabled, 'tenant role should be enabled').equal(true);
                        expect(tenantRole.effectiveEnabled, 'tenant role should be effectively enabled').equal(true);

                        try {
                            await authorizationUtils.createAccessRuleForUser(
                                tenant1Info.tenantAdminToken.accessToken,
                                accessRuleUserName,
                                role.id,
                                SubjectType.User,
                                tenant1Info.tenant.tenant.id.toString(),
                                ScopeType.Tenant,
                                tenant1AccessRulesApi
                            );
                            expect.fail('Expect disables of role access-rule to fail.');
                        } catch (error) {
                            expect(error.response.statusCode).equal(
                                400,
                                `Error status is: ${error.response.status}, and not as expected: 400`
                            );

                            expect(error.response.body.message).equal(
                                `Cannot create access rule, role ${role.name} is disabled`,
                                `Error message is: ${error.response.body.message}, and not as expected: Cannot create access rule, role ${role.name} is disabled`
                            );
                        }
                    });
                });

                test.describe('get roles', function () {
                    test('test get roles', async function () {
                        const rolesPrefix = 'get-roles-';
                        const permissions = [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ];
                        const role1 = await authorizationUtils.createCustomRole(
                            rolesPrefix + 'sys-role-1',
                            roleDescription,
                            permissions,
                            undefined,
                            consoleToken.accessToken,
                            RoleCreationFields.ScopeTypeEnum.System,
                            systemTenant.id + '',
                            true,
                            consoleRolesApi
                        );
                        expect(role1, 'sys-role-1 should be created').to.not.be.null;
                        systemRoles.push(role1);
                        const role2 = await authorizationUtils.createCustomRole(
                            rolesPrefix + 'sys-role-2',
                            roleDescription,
                            permissions,
                            undefined,
                            consoleToken.accessToken,
                            RoleCreationFields.ScopeTypeEnum.System,
                            systemTenant.id + '',
                            true,
                            consoleRolesApi
                        );
                        expect(role2, 'sys-role-2 should be created').to.not.be.null;
                        systemRoles.push(role2);
                        const role3 = await authorizationUtils.createCustomRole(
                            rolesPrefix + 'sys-role-3',
                            roleDescription,
                            permissions,
                            undefined,
                            consoleToken.accessToken,
                            RoleCreationFields.ScopeTypeEnum.System,
                            systemTenant.id + '',
                            false,
                            consoleRolesApi
                        );
                        expect(role3, 'sys-role-3 (disabled) should be created').to.not.be.null;
                        systemRoles.push(role3);
                        const roleActiveFields = {
                            scopeType: RoleActiveFields.ScopeTypeEnum.Tenant,
                            enabled: false
                        };
                        await authorizationUtils.activateRole(
                            role1,
                            roleActiveFields,
                            tenant1Info.tenantAdminToken.accessToken,
                            tenant1RolesApi
                        );
                        const role = await authorizationUtils.createCustomRole(
                            rolesPrefix + 'tenant-role',
                            roleDescription,
                            permissions,
                            undefined,
                            tenant1Info.tenantAdminToken.accessToken,
                            RoleCreationFields.ScopeTypeEnum.Tenant,
                            tenant1Info.tenant.tenant.id + '',
                            true,
                            tenant1RolesApi
                        );
                        expect(role, 'created role should not be null').to.not.be.null;
                        tenant1Roles.push(role);
                        const sysRoles = await authorizationUtils.getRolesV2(
                            consoleToken.accessToken,
                            consoleRolesApi,
                            new Array<string>('name=^' + rolesPrefix)
                        );
                        expect(sysRoles.length, 'system roles query should return 3 roles matching prefix').to.equal(3);
                        for (const sysRole of sysRoles) {
                            if (sysRole.name === rolesPrefix + 'sys-role-3') {
                                expect(sysRole.enabled, 'sys-role-3 should have enabled=false').to.equal(false);
                                expect(sysRole.effectiveEnabled, 'sys-role-3 should be effectively disabled').to.equal(false);
                            } else {
                                expect(sysRole.enabled, 'non-disabled system role should be enabled').to.equal(true);
                                expect(sysRole.effectiveEnabled, 'non-disabled system role should be effectively enabled').to.equal(true);
                            }
                        }
                        const t1Roles = await authorizationUtils.getRolesV2(
                            tenant1Info.tenantAdminToken.accessToken,
                            tenant1RolesApi,
                            new Array<string>('name=^' + rolesPrefix)
                        );
                        expect(t1Roles.length, 'tenant1 roles query should return 4 roles matching prefix (3 sys + 1 tenant)').to.equal(4);
                        for (const t1Role of t1Roles) {
                            if (t1Role.name === rolesPrefix + 'sys-role-3') {
                                expect(t1Role.enabled, 'sys-role-3 should have enabled=false').to.equal(false);
                                expect(t1Role.effectiveEnabled, 'sys-role-3 should be effectively disabled').to.equal(false);
                            } else {
                                expect(t1Role.enabled, 'non-disabled system role should be enabled').to.equal(true);
                                if (t1Role.name === rolesPrefix + 'sys-role-1') {
                                    expect(t1Role.effectiveEnabled, 'tenant role should be effectively disabled after tenant disable').to.equal(false);
                                } else {
                                    expect(t1Role.effectiveEnabled, 'non-disabled system role should be effectively enabled').to.equal(true);
                                }
                            }
                        }
                        const t2Roles = await authorizationUtils.getRolesV2(
                            tenant2Info.tenantAdminToken.accessToken,
                            tenant2RolesApi,
                            new Array<string>('name=^' + rolesPrefix)
                        );
                        expect(t2Roles.length, 'tenant2 roles query should return 3 system roles matching prefix').to.equal(3);
                        for (const t2Role of t2Roles) {
                            if (t2Role.name === rolesPrefix + 'sys-role-3') {
                                expect(t2Role.enabled, 'sys-role-3 should have enabled=false').to.equal(false);
                                expect(t2Role.effectiveEnabled, 'sys-role-3 should be effectively disabled').to.equal(false);
                            } else {
                                expect(t2Role.enabled, 'non-disabled system role should be enabled').to.equal(true);
                                expect(t2Role.effectiveEnabled, 'non-disabled system role should be effectively enabled').to.equal(true);
                            }
                        }
                    });
                });

                test.describe('predefined roles', function () {
                    test('Verify predefined roles can be disabled in tenant level', async function () {
                        const predefinedRoles = await authorizationUtils.getRoles(
                            tenant1Info.tenantAdminToken.accessToken,
                            tenant1RolesApi
                        );
                        const dataSourceAdminRole = predefinedRoles.find(
                            (r) => r.name === PROPERTY_TO_ROLE_MAPPING.data_source_administrator
                        );

                        const roleDeActiveFields = {
                            scopeType: RoleActiveFields.ScopeTypeEnum.Tenant,
                            enabled: false
                        };
                        await authorizationUtils.activateRole(
                            dataSourceAdminRole,
                            roleDeActiveFields,
                            tenant1Info.tenantAdminToken.accessToken,
                            tenant1RolesApi
                        );

                        const getRolesRes = await authorizationUtils.getRoles(
                            tenant1Info.tenantAdminToken.accessToken,
                            tenant1RolesApi
                        );
                        const disabledRole = getRolesRes.find((r) => r.id === dataSourceAdminRole.id);
                        expect(disabledRole.enabled, 'Role is disabled, expected role.enabled to be true').to.be
                            .true;
                        expect(
                            disabledRole.effectiveEnabled,
                            'Role is disabled, expected role.effectiveEnabled to be false'
                        ).to.be.false;

                        const roleActiveFields = {
                            scopeType: RoleActiveFields.ScopeTypeEnum.Tenant,
                            enabled: true
                        };
                        await authorizationUtils.activateRole(
                            dataSourceAdminRole,
                            roleActiveFields,
                            tenant1Info.tenantAdminToken.accessToken,
                            tenant1RolesApi
                        );

                        const getRolesResEnabled = await authorizationUtils.getRoles(
                            tenant1Info.tenantAdminToken.accessToken,
                            tenant1RolesApi
                        );
                        const enabledRole = getRolesResEnabled.find((r) => r.id === dataSourceAdminRole.id);
                        expect(enabledRole.enabled, 'Role is enabled, expected role.enabled to be true').to.be
                            .true;
                        expect(
                            enabledRole.effectiveEnabled,
                            'Role is enabled, expected role.effectiveEnabled to be true'
                        ).to.be.true;
                    });
                });
            });
        });
    }
);
