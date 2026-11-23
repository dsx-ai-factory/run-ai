/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authorizationUtils,
    authUtils,
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    usersManager,
    versionUtils,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { test, expect } from '@playwright/test';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../../modules/helpers/config/roleMapping';
import { Permission } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/permission';
import { ResourceType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/resourceType';
import { RoleCreationFields } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/roleCreationFields';
import { ResourceTypeGroupId } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/resourceTypeGroupId';
import { Action } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/action';
import { StatusCodes } from 'http-status-codes';
import { KubernetesPermissions } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/kubernetesPermissions';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(
    `Custom roles CRUD, tags: ${tags.tool.api} ${tags.resource.roles} ${tags.teams.viper}`,
    { tag: [tags.services.authorization, tags.runOn.serialCI] },
    () => {
        supportedFrom(SupportedVersions.V2_23)(() => {
            test.describe('Tenant-Level Custom Role Sanity', () => {
                const resourceTypes = [];
                const rolesCreated = [],
                    accessRulesCreated = [];
                let role, accessRule;
                let user, userName, project;

                test.beforeAll(async () => {
                    const res = await authorizationUtils.getRoles();
                    if (res.length === 0) {
                        throw new Error('No roles found');
                    }
                    for (const permission of res[0].permissions) {
                        resourceTypes.push(permission);
                    }
                    userName = objectGenerator.generateEmailByPrefix('user');
                    user = (
                        await usersManager.createLocalUser(
                            flowRootState.users.system_administrator.accessToken,
                            userName
                        )
                    ).body;

                    const projectName = objectGenerator.generateNameByPrefix('p-custom-roles');
                    project = await projectResourcesUtils.createProjectAndNamespace(projectName);
                });

                test.afterAll(async () => {
                    try {
                        await authUtils.deleteUser(
                            user.id,
                            flowRootState.users.system_administrator.accessToken
                        );
                    } catch (error) {
                        logger.warn(`Failed to delete user ${JSON.stringify(user)}: ${error}`);
                    }
                    for (const accessRule of accessRulesCreated) {
                        try {
                            await authorizationUtils.deleteAccessRule(accessRule);
                        } catch (error) {
                            logger.warn(
                                `Failed to delete access rule ${JSON.stringify(accessRule)}: ${error}`
                            );
                        }
                    }
                    for (const role of rolesCreated) {
                        try {
                            await authorizationUtils.deleteRole(
                                role.id,
                                flowRootState.users.system_administrator.accessToken
                            );
                        } catch (error) {
                            logger.warn(`Failed to delete role ${JSON.stringify(role)}: ${error}`);
                        }
                    }
                    try {
                        await operations.deleteProjectAndNamespace(project.id);
                    } catch (error) {
                        logger.warn(`Failed to delete project ${JSON.stringify(project)}: ${error}`);
                    }
                });

                test('Create a tenant-level role as a system admin', async () => {
                    const roleName = objectGenerator.generateNameByPrefix('role');
                    const roleDescription = 'role description';
                    const permissions = [];
                    const rolePermissions = [];
                    const permissionMap = new Map<string, string>();
                    for (const resourceType of resourceTypes) {
                        const actions = [];
                        if (
                            resourceType.resourceType === 'project' ||
                            resourceType.resourceType === 'department'
                        ) {
                            actions.push(Action.Create, Action.Read, Action.Update);
                        }
                        permissions.push({
                            resourceType: resourceType.resourceType,
                            actions: actions,
                            displayName: resourceType.displayName,
                            groupId: resourceType.groupId
                        });
                        rolePermissions.push({
                            resourceType: resourceType.resourceType,
                            actions: actions,
                            displayName: resourceType.displayName
                        });
                        permissionMap[resourceType.resourceType] = resourceType.groupId;
                    }

                    role = await authorizationUtils.createCustomRole(roleName, roleDescription, permissions);
                    const sortByResourceType = (a, b): number => a.resourceType.localeCompare(b.resourceType);

                    expect(role.name, {
                        message: `Role name is: ${role.name}, and not as expected: ${roleName}`
                    }).toBe(roleName);
                    expect(role.description, {
                        message: `Role description is: ${role.description}, and not as expected: ${roleDescription}`
                    }).toBe(roleDescription);
                    expect(role.permissions.sort(sortByResourceType), {
                        message: `Role permissions are: ${JSON.stringify(role.permissions)}, and not as expected: ${JSON.stringify(permissions)}`
                    }).toEqual(rolePermissions.sort(sortByResourceType));
                    expect(role.custom, {
                        message: 'Role created is custom, expected role.custom to be true'
                    }).toBe(true);
                    expect(role.createdBy, {
                        message: `Role createdBy is: ${role.createdBy}, and not as expected: ${flowRootState.users.system_administrator.email}`
                    }).toBe(flowRootState.users.system_administrator.email);
                    expect(role.enabled, {
                        message: 'Role created is enabled, expected role.enabled to be true'
                    }).toBe(true);
                    expect(role.kubernetesPermissions.predefinedRole, {
                        message: `Role k8s permissions are: ${role.kubernetesPermissions.predefinedRole}, and not as expected: undefined`
                    }).toBe(undefined);
                });

                test('Verify the tenant-leval role is functional', async () => {
                    accessRule = await authorizationUtils.createAccessRuleForUser(
                        flowRootState.users.system_administrator.accessToken,
                        userName,
                        role.id
                    );
                    accessRulesCreated.push(accessRule);

                    logger.info(
                        `Created access rule: ${JSON.stringify(accessRule)} with custom role: ${JSON.stringify(role)}`
                    );
                    await authorizationUtils.waitForAccessRule(accessRule.id, true);
                    expect(accessRule.roleId, {
                        message: `Access rule role id is: ${accessRule.roleId}, and not as expected: ${role.id}`
                    }).toBe(role.id);

                    // get user's permissions and validates it is aligned with the role permissions
                    const userToken = await authUtils.authenticateUser(userName, user.tempPassword);
                    const userPermissions = await authorizationUtils.getUserPermissions(userToken.idToken);
                    for (const permission of userPermissions) {
                        if (
                            permission.resourceType === ResourceType.Project ||
                            permission.resourceType === ResourceType.Department
                        ) {
                            expect(permission.actions.length).toBe(3);
                            expect(permission.actions).toContain(Action.Create);
                            expect(permission.actions).toContain(Action.Read);
                            expect(permission.actions).toContain(Action.Update);
                        } else {
                            expect(permission.actions.length).toBe(0);
                        }
                    }
                });

                test('Verify the tenant-leval role is returned is get', async () => {
                    let userRole = (
                        await authorizationUtils.getRoles(
                            flowRootState.users.system_administrator.accessToken
                        )
                    ).find((userRole) => userRole.name === role.name);
                    expect(userRole).toBeDefined();
                    userRole = await authorizationUtils.getRoleV2(role.id);
                    expect(userRole).toBeDefined();
                });

                test('Attempt to update (permissions/disabled) the role associated with the access rule', async () => {
                    const updatedRoleFields = {
                        name: role.name,
                        description: role.description,
                        permissions: [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read],
                                displayName: 'Projects',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ]
                    } as RoleCreationFields;

                    try {
                        await authorizationUtils.updateRole(role, updatedRoleFields);
                        throw new Error(
                            'Expect update role to fail since it is associated with an access rule'
                        );
                    } catch (error) {
                        expect(error.response.statusCode, {
                            message: `Error status is: ${error.response.status}, and not as expected: 400`
                        }).toBe(400);

                        expect(error.response.body.message, {
                            message: `Error message is: ${error.response.body.message}, and not as expected: Can not update or delete a role that is currently in use`
                        }).toBe('Can not update or delete a role that is currently in use');
                    }
                });

                test('Attempt to delete the role associated with the access rule', async () => {
                    try {
                        await authorizationUtils.deleteRole(
                            role.id,
                            flowRootState.users.system_administrator.accessToken
                        );
                        throw new Error(
                            'Expect delete role to fail since it is associated with an access rule'
                        );
                    } catch (error) {
                        expect(error.response.statusCode, {
                            message: `Error status is: ${error.response.status}, and not as expected: 400`
                        }).toBe(400);
                    }
                });

                test('Update the custom role permissions', async () => {
                    // delete the access rule
                    await authorizationUtils.deleteAccessRule(accessRule);
                    await authorizationUtils.waitForAccessRule(accessRule.id, false);

                    // Set the permissions to be only read for the project resource type
                    const newPermission = {
                        resourceType: ResourceType.Project,
                        actions: [Action.Read],
                        displayName: 'Projects',
                        groupId: ResourceTypeGroupId.Organization
                    } as Permission;
                    const updatedRoleFields = {
                        name: role.name,
                        description: role.description,
                        permissions: [newPermission]
                    } as RoleCreationFields;
                    const updatedRole = await authorizationUtils.updateRole(role, updatedRoleFields);

                    expect(updatedRole.name, {
                        message: `Role name is: ${updatedRole.name}, and not as expected: ${role.name}`
                    }).toBe(role.name);
                    for (const permission of updatedRole.permissions) {
                        if (permission.resourceType === ResourceType.Project) {
                            expect(permission.resourceType).toEqual(newPermission.resourceType);
                            expect(permission.actions).toEqual(newPermission.actions);
                            expect(permission.displayName).toEqual(newPermission.displayName);
                        } else {
                            expect(permission.actions.length).toBe(0);
                        }
                    }
                });

                test('Try update custom role predefined cluster role - to a role that violates privilege escalation', async () => {
                    if (
                        !versionUtils.isVersionSupported(
                            flowRootState.cluster.version,
                            SupportedVersions.V2_23
                        )
                    ) {
                        test.skip();
                    }
                    const roles = await authorizationUtils.getRoles(
                        flowRootState.users.system_administrator.accessToken
                    );
                    const departmentAdminRole = roles.find(
                        (role) => role.name === 'Department administrator'
                    );
                    try {
                        const k8sPermissions: KubernetesPermissions = {
                            predefinedRole: departmentAdminRole.id + ''
                        };
                        await authorizationUtils.updateRolesK8sPermissions(role, k8sPermissions);
                        throw new Error('Expect update role to fail since it violates privilege escalation');
                    } catch (error) {
                        expect(error.response.statusCode, {
                            message: `Error status is: ${error.response.status}, and not as expected: 400`
                        }).toBe(400);
                        expect(error.response.body.message, {
                            message: `Error message is: ${error.response.body.message}, and not as expected: custom role invalid since KubernetesPermissions predefined role has read permission on tenant and custom role does not`
                        }).toBe(
                            'custom role invalid since KubernetesPermissions predefined role has read permission on tenant and custom role does not'
                        );
                    }
                });

                test('Update the custom role permissions and predefined cluster role, and verify that a user has permissions in cluster', async () => {
                    if (
                        !versionUtils.isVersionSupported(
                            flowRootState.cluster.version,
                            SupportedVersions.V2_23
                        )
                    ) {
                        test.skip();
                    }
                    // Set the permissions to be only read for the project resource type
                    const researcherRole = flowRootState.users.l1_researcher.role;
                    const researcherPermissions = researcherRole.permissions;

                    const permissions = [];
                    for (const resourceType of resourceTypes) {
                        const actions = [];
                        if (
                            resourceType.resourceType === 'project' ||
                            resourceType.resourceType === 'department'
                        ) {
                            actions.push(Action.Create, Action.Read, Action.Update);
                        } else {
                            actions.push(resourceType.actions);
                        }
                        permissions.push({
                            resourceType: resourceType.resourceType,
                            actions: actions,
                            displayName: resourceType.displayName,
                            groupId: resourceType.groupId
                        });
                    }

                    const updatedRoleFields = {
                        name: role.name,
                        description: role.description,
                        permissions: researcherPermissions,
                        kubernetesPermissions: {
                            predefinedRole: researcherRole.id + ''
                        }
                    } as RoleCreationFields;
                    const updatedRole = await authorizationUtils.updateRole(role, updatedRoleFields);
                    expect(updatedRole.kubernetesPermissions.predefinedRole, {
                        message: `Role k8s permissions are: ${updatedRole.kubernetesPermissions.predefinedRole}, and not as expected: ${researcherRole.id}`
                    }).toBe(researcherRole.id + '');

                    // create access rule
                    accessRule = await authorizationUtils.createAccessRuleForUser(
                        flowRootState.users.system_administrator.accessToken,
                        userName,
                        role.id
                    );
                    accessRulesCreated.push(accessRule);

                    logger.info(
                        `Created access rule: ${JSON.stringify(accessRule)} with custom role: ${JSON.stringify(role)}`
                    );
                    await authorizationUtils.waitForAccessRule(accessRule.id, true);
                    expect(accessRule.roleId, {
                        message: `Access rule role id is: ${accessRule.roleId}, and not as expected: ${role.id}`
                    }).toBe(role.id);

                    const clusterRoleBindingName = 'runai-l1-researcher';
                    await waitUtils.waitForUserInClusterRoleBinding(userName, clusterRoleBindingName);

                    // Validates user actually has the permissions in the cluster
                    const userToken = await authUtils.authenticateUser(userName, user.tempPassword);

                    const workloadName = objectGenerator.generateNameByPrefix('w');
                    const submitRes = await workloadsCrudUtils.createTraining(
                        workloadName,
                        false,
                        project.clusterUuid,
                        project.id.toString(),
                        {
                            image: config.quickStartImage.ubuntu,
                            command: 'sleep infinity'
                        },
                        userToken.idToken
                    );
                    const submittedWorkload = submitRes.body;
                    expect(submittedWorkload, {
                        message:
                            'User with researcher permissions in cluster should be able to create workload through CP backend api'
                    }).toBeDefined();

                    // clean up
                    try {
                        await workloadsUtilsV2.deleteWorkloadByNameAndProjectId(
                            submittedWorkload.name,
                            project.id,
                            WorkloadType.TRAINING,
                            flowRootState.users.system_administrator.accessToken
                        );
                    } catch (error) {
                        logger.warn(
                            `Failed to delete workload ${JSON.stringify(submittedWorkload)}: ${error}`
                        );
                    }
                    try {
                        await authorizationUtils.deleteAccessRule(accessRule);
                    } catch (error) {
                        logger.warn(`Failed to delete access rule ${JSON.stringify(accessRule)}: ${error}`);
                    }
                });

                test('Disable the tenant-level role', async () => {
                    const updatedRoleFields = {
                        name: role.name,
                        description: role.description,
                        permissions: [
                            {
                                resourceType: ResourceType.Project,
                                actions: [Action.Read],
                                displayName: 'Project',
                                groupId: ResourceTypeGroupId.Organization
                            } as Permission
                        ],
                        enabled: false
                    } as RoleCreationFields;

                    const updatedRole = await authorizationUtils.updateRole(role, updatedRoleFields);

                    expect(updatedRole.enabled, {
                        message: 'Role is disabled, expected role.enabled to be false'
                    }).toBe(false);

                    try {
                        const userName = objectGenerator.generateEmailByPrefix('user');
                        const accessRule = await authorizationUtils.createAccessRuleForUser(
                            flowRootState.users.system_administrator.accessToken,
                            userName,
                            role.id
                        );
                        accessRulesCreated.push(accessRule);
                        throw new Error('Expect create access rule to fail since the role is disabled');
                    } catch (error) {
                        expect(error.response.statusCode, {
                            message: `Error status is: ${error.response.status}, and not as expected: 400`
                        }).toBe(StatusCodes.BAD_REQUEST);
                        expect(error.response.body.message, {
                            message: `Error message is: ${error.response.body.message}, and not as expected: Cannot create access rule, role ${role.name} is disabled`
                        }).toBe(`Cannot create access rule, role ${role.name} is disabled`);
                    }
                });

                test('Verify the disabled role is excluded from get roles (if expected behaviour)', async () => {
                    const getRolesRes = await authorizationUtils.getRoles();
                    const disabledRole = getRolesRes.find((r) => r.id === role.id);
                    expect(disabledRole.enabled, {
                        message: 'Role is disabled, expected role.enabled to be false'
                    }).toBe(false);
                });

                test('Delete the tenant-level role', async () => {
                    await authorizationUtils.deleteRole(
                        role.id,
                        flowRootState.users.system_administrator.accessToken
                    );

                    const getRes = await authorizationUtils.getRoles();
                    const deletedRole = getRes.find((r) => r.id === role.id);
                    expect(deletedRole, {
                        message: 'Role is deleted, expected role to be undefined'
                    }).toBeUndefined();
                });
            });

            test.describe('Backwards Compatibility', () => {
                test('Verify that predefined roles cannot be updated', async () => {
                    if (
                        !versionUtils.isVersionSupported(
                            flowRootState.cluster.version,
                            SupportedVersions.V2_23
                        )
                    ) {
                        test.skip();
                    }
                    const predefinedRoles = await authorizationUtils.getRoles();
                    const dataSourceAdminRole = predefinedRoles.find(
                        (r) => r.name === PROPERTY_TO_ROLE_MAPPING.data_source_administrator
                    );

                    try {
                        const roleUpdateFields = {
                            name: dataSourceAdminRole.name,
                            description: 'updated description',
                            permissions: dataSourceAdminRole.permissions,
                            enabled: true
                        };
                        await authorizationUtils.updateRole(dataSourceAdminRole, roleUpdateFields);
                        throw new Error('Expect update role to fail since it is predefined');
                    } catch (error) {
                        expect(error.response.statusCode, {
                            message: `Error status is: ${error.response.status}, and not as expected: 400`
                        }).toBe(StatusCodes.FORBIDDEN);
                        expect(error.response.body.message, {
                            message: `Error message is: ${error.response.body.message}, and not as expected: Insufficient permissions`
                        }).toBe('Insufficient permissions');
                    }
                });

                test('Verify that predefined roles cannot be deleted', async () => {
                    const predefinedRoles = await authorizationUtils.getRoles();
                    const dataSourceAdminRole = predefinedRoles.find(
                        (r) => r.name === PROPERTY_TO_ROLE_MAPPING.data_source_administrator
                    );

                    try {
                        await authorizationUtils.deleteRole(dataSourceAdminRole.id);
                        throw new Error('Expect delete role to fail since it is predefined');
                    } catch (error) {
                        expect(error.response.statusCode, {
                            message: `Error status is: ${error.response.status}, and not as expected: 400`
                        }).toBe(StatusCodes.FORBIDDEN);
                        expect(error.response.body.message, {
                            message: `Error message is: ${error.response.body.message}, and not as expected: Insufficient permissions`
                        }).toBe('Insufficient permissions');
                    }
                });
            });
        });
    }
);
