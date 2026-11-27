/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    authorizationUtils,
    authUtils,
    settings,
    config,
    usersManager
} from '../../../../../../modules/helpers';
import { AuthorizationUtils } from '../../../../../../modules/helpers/utils/AuthorizationUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { authorizationApi } from '../../../../../../modules/api';
import { PermissionSet } from '../../../../../../modules/api/apis/authorizationApi/generated/model/permissionSet';
import { StatusCodes } from 'http-status-codes';
import { expect as chaiExpect } from 'chai';
import { Role } from '../../../../../../modules/api/apis/authorizationApi/generated/model/role';
import { RoleCreationFieldsInternal } from '../../../../../../modules/api/apis/authorizationApi/generated/model/roleCreationFieldsInternal';
import { RolePermissionSet } from '../../../../../../modules/api/apis/authorizationApi/generated/model/rolePermissionSet';
import { randomUUID } from 'crypto';

test.describe(`Test Group & organize RBAC permissions, tags: ${tags.tool.api} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let user: UserPostResponse;
        let testPermissionSetId: string;
        let testPermissionSetName: string;
        let secondPermissionSetId: string;
        const createdRoleIds: number[] = [];
        let permissionSets: PermissionSet[] = [];
        let permissionSet: PermissionSet;
        let adminToken: string;
        let superAdminToken: string | null = null;
        let tenantId: string;
        let createdRole: Role;
        let error: any;
        let response: any;

        const VALID_ACTIONS = ['create', 'read', 'update', 'delete'];

        test.beforeAll(async function () {
            project = await projectResourcesUtils.createProjectAndNamespace(
                'group-permissions-test',
                null,
                0,
                0
            );
            projectName = project.name;
            namespace = project.namespace;

            adminToken = flowRootState.users.system_administrator.accessToken;
            tenantId = flowRootState.cluster.tenantId.toString();

            // Try to get super admin token for system scope role testing (self-hosted only)
            if (settings.isSelfHosted && config.tenantUrl) {
                try {
                    const { Keycloak } = await import('../../../../../../modules/helpers/utils/Keycloak');
                    settings.tenantClientSecret = await Keycloak.getClientSecretFromSpecificClient(
                        'console',
                        'default'
                    );
                    const token = await authUtils.authenticateSuperAdminAppAndGetToken(
                        'default',
                        settings.tenantClientSecret,
                        config.tenantUrl
                    );
                    superAdminToken = token.idToken;
                    logger.info('Super admin token obtained successfully');
                } catch (error) {
                    logger.debug(
                        `Super admin token not available (${error.message}), T5591170 will be skipped`
                    );
                }
            }

            authorizationApi.permissionsApi.accessToken = adminToken;
            authorizationApi.rolesApi.accessToken = adminToken;

            const response = await authorizationApi.permissionsApi.getPermissionSets();
            const permissionSets = response.body.permissionSets;
            if (permissionSets && permissionSets.length > 0) {
                testPermissionSetId = permissionSets[0].id;
                testPermissionSetName = permissionSets[0].name;
                if (permissionSets.length > 1) {
                    secondPermissionSetId = permissionSets[1].id;
                }
                logger.info(`Initialized permission sets: ${testPermissionSetId} (${testPermissionSetName})`);
            }
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        test.beforeEach(function () {
            authorizationApi.permissionsApi.accessToken = adminToken;
            authorizationApi.rolesApi.accessToken = adminToken;
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            if (createdRoleIds.length > 0) {
                authorizationApi.rolesApi.accessToken = flowRootState.users.system_administrator.accessToken;
                for (const roleId of createdRoleIds) {
                    try {
                        await authorizationApi.rolesApi.deleteRole(roleId);
                        logger.info(`Deleted role with ID: ${roleId}`);
                    } catch (error) {
                        if (error.statusCode !== StatusCodes.NOT_FOUND) {
                            logger.error(`Failed to delete role ${roleId}: ${error.message}`);
                        } else {
                            logger.debug(
                                `Role ${roleId} not found (may have been already deleted or never created)`
                            );
                        }
                    }
                }
                createdRoleIds.length = 0;
            }

            if (user && user.id) {
                await usersManager.deleteLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    user.id
                );
            }
        });

        test('T5591155: Get All Permission Sets', async function () {
            await test.step('Call GET API: /v1/api/permission-sets', async () => {
                response = await authorizationApi.permissionsApi.getPermissionSets();
                chaiExpect(response.response.statusCode).to.equal(StatusCodes.OK);
                logger.info('Successfully called GET /v1/api/permission-sets');
            });

            await test.step('Verify response is a list of permission sets', () => {
                chaiExpect(response.body).to.have.property('permissionSets');
                chaiExpect(response.body.permissionSets).to.be.an('array');
                permissionSets = response.body.permissionSets;
                chaiExpect(permissionSets.length).to.be.greaterThan(
                    0,
                    'Expected at least one permission set in the response'
                );
                logger.info(`Verified response contains ${permissionSets.length} permission sets`);
            });

            await test.step('For each permission set, verify structure contains required fields', () => {
                permissionSets.forEach((permissionSet, index) => {
                    chaiExpect(permissionSet, `Permission set ${index} should have id`).to.have.property(
                        'id'
                    );
                    chaiExpect(permissionSet.id, `Permission set ${index} id should be a string`).to.be.a(
                        'string'
                    );
                    chaiExpect(
                        permissionSet.id,
                        `Permission set ${index} id should be a valid UUID`
                    ).to.match(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i);

                    chaiExpect(permissionSet, `Permission set ${index} should have name`).to.have.property(
                        'name'
                    );
                    chaiExpect(permissionSet.name, `Permission set ${index} name should be a string`).to.be.a(
                        'string'
                    );

                    chaiExpect(
                        permissionSet,
                        `Permission set ${index} should have description`
                    ).to.have.property('description');
                    if (permissionSet.description !== undefined) {
                        chaiExpect(
                            permissionSet.description,
                            `Permission set ${index} description should be a string`
                        ).to.be.a('string');
                    }

                    chaiExpect(
                        permissionSet,
                        `Permission set ${index} should have permissions array`
                    ).to.have.property('permissions');
                    chaiExpect(
                        permissionSet.permissions,
                        `Permission set ${index} permissions should be an array`
                    ).to.be.an('array');
                });
                logger.info(
                    'Verified all permission sets have required structure (id, name, description, permissions)'
                );
            });

            await test.step('For each permission object, verify it contains required fields', () => {
                let totalPermissionsChecked = 0;
                permissionSets.forEach((permissionSet, psIndex) => {
                    permissionSet.permissions.forEach((permission, permIndex) => {
                        chaiExpect(
                            permission,
                            `Permission ${permIndex} in set ${psIndex} should have resourceType`
                        ).to.have.property('resourceType');
                        chaiExpect(
                            permission.resourceType,
                            `Permission ${permIndex} in set ${psIndex} resourceType should be a string`
                        ).to.be.a('string');

                        chaiExpect(
                            permission,
                            `Permission ${permIndex} in set ${psIndex} should have displayName`
                        ).to.have.property('displayName');
                        if (permission.displayName !== undefined) {
                            chaiExpect(
                                permission.displayName,
                                `Permission ${permIndex} in set ${psIndex} displayName should be a string`
                            ).to.be.a('string');
                        }

                        chaiExpect(
                            permission,
                            `Permission ${permIndex} in set ${psIndex} should have actions`
                        ).to.have.property('actions');
                        chaiExpect(
                            permission.actions,
                            `Permission ${permIndex} in set ${psIndex} actions should be an array`
                        ).to.be.an('array');

                        permission.actions.forEach((action, actionIndex) => {
                            chaiExpect(
                                action,
                                `Action ${actionIndex} in permission ${permIndex} of set ${psIndex} should be one of: ${VALID_ACTIONS.join(', ')}`
                            ).to.be.oneOf(VALID_ACTIONS);
                        });

                        totalPermissionsChecked++;
                    });
                });
                logger.info(
                    `Verified ${totalPermissionsChecked} permissions across ${permissionSets.length} permission sets have required structure (resourceType, displayName, actions with create/read/update/delete)`
                );
            });

            await test.step('Verify no deprecated permission sets are included in the response', () => {
                permissionSets.forEach((permissionSet) => {
                    chaiExpect(permissionSet.name.toLowerCase()).to.not.include('deprecated');
                    if (permissionSet.description) {
                        chaiExpect(permissionSet.description.toLowerCase()).to.not.include('deprecated');
                    }
                });
                logger.info('Verified no deprecated permission sets are included in the response');
            });
        });

        test('T5591164: Get Specific Permission Set by ID', async function () {
            await test.step(`Call GET API: /v1/api/permission-sets/${testPermissionSetId}`, async () => {
                chaiExpect(testPermissionSetId).to.exist;
                response = await authorizationApi.permissionsApi.getPermissionSet(testPermissionSetId);
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.OK);
                logger.info(`Successfully got permission set ${testPermissionSetId}`);
            });

            await test.step('Verify response contains the requested permission set with correct ID', () => {
                chaiExpect(response.body).to.exist;
                permissionSet = response.body;
                chaiExpect(permissionSet).to.be.an('object');
                chaiExpect(permissionSet, 'Permission set should have id').to.have.property('id');
                chaiExpect(permissionSet.id, 'Permission set id should match the requested ID').to.equal(
                    testPermissionSetId
                );
                logger.info(
                    `Verified response contains permission set with correct ID: ${testPermissionSetId}`
                );
            });
        });

        test('T5591165: Get Permission Sets Response Structure Validation', async function () {
            let firstPermissionSet: PermissionSet;

            await test.step('Call GET API and select first permission set', async () => {
                authorizationApi.permissionsApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;
                response = await authorizationApi.permissionsApi.getPermissionSets();

                chaiExpect(response.body).to.have.property('permissionSets').that.is.an('array');
                permissionSets = response.body.permissionSets;
                chaiExpect(permissionSets.length).to.be.greaterThan(0);
                firstPermissionSet = permissionSets[0];
                logger.info(
                    `Got ${permissionSets.length} permission set(s), validating first: ${firstPermissionSet.id}`
                );
            });

            await test.step('Verify permission set structure', () => {
                chaiExpect(firstPermissionSet.id).to.match(
                    /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
                    'Permission set id should be a valid UUID'
                );

                chaiExpect(firstPermissionSet.name).to.be.a('string').with.length.greaterThan(0);

                if (firstPermissionSet.description) {
                    chaiExpect(firstPermissionSet.description).to.be.a('string');
                }

                chaiExpect(firstPermissionSet.permissions).to.be.an('array').with.length.greaterThan(0);

                logger.info(
                    `Verified structure: name="${firstPermissionSet.name}", ${firstPermissionSet.permissions.length} permission(s)`
                );
            });

            await test.step('Verify permissions array details', () => {
                firstPermissionSet.permissions.forEach((permission, index) => {
                    const resourceTypeStr = String(permission.resourceType);
                    chaiExpect(
                        resourceTypeStr.length,
                        `Permission ${index} resourceType should not be empty`
                    ).to.be.greaterThan(0);

                    if (permission.displayName) {
                        chaiExpect(
                            permission.displayName,
                            `Permission ${index} displayName should be a string`
                        ).to.be.a('string');
                    }

                    chaiExpect(permission.actions, `Permission ${index} should have non-empty actions array`)
                        .to.be.an('array')
                        .with.length.greaterThan(0);

                    permission.actions.forEach((action) => {
                        chaiExpect(action).to.be.oneOf(VALID_ACTIONS);
                    });
                });
                logger.info(
                    `Verified all ${firstPermissionSet.permissions.length} permissions have valid structure`
                );
            });

            await test.step('Verify no internal fields exposed', () => {
                ['_id', 'internalId', 'createdBy', 'updatedBy', 'deletedAt'].forEach((field) => {
                    chaiExpect(firstPermissionSet).to.not.have.property(field);
                });
                logger.info('Verified no internal fields exposed');
            });
        });

        test('T5591166: Get Non-Existent Permission Set', async function () {
            const nonExistentId = randomUUID();

            await test.step('Call GET API with non-existent ID and expect 404 error', async () => {
                authorizationApi.permissionsApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;
                try {
                    response = await authorizationApi.permissionsApi.getPermissionSet(nonExistentId);
                    chaiExpect.fail('API should have returned 404 error for non-existent permission set');
                } catch (err) {
                    error = err;
                    logger.info(`API returned 404 as expected for ID: ${nonExistentId}`);
                }
            });

            await test.step('Verify 404 error response', () => {
                chaiExpect(error).to.exist;
                chaiExpect(error.statusCode || error.status).to.equal(StatusCodes.NOT_FOUND);

                const errorBody = error.body || error.response?.body;
                chaiExpect(errorBody).to.exist;
                chaiExpect(errorBody.code).to.equal(404);
                chaiExpect(errorBody.message).to.be.a('string').that.includes('Failed to get permission set');

                logger.info(`Verified 404 error: ${errorBody.message}`);
            });
        });

        test('T5591167: Get Permission Sets Without Authentication', async function () {
            await test.step('GET API without authentication and verify response', async () => {
                authorizationApi.permissionsApi.accessToken = '';

                try {
                    await authorizationApi.permissionsApi.getPermissionSets();
                    chaiExpect.fail('API call should have failed without authentication');
                } catch (err) {
                    error = err;
                } finally {
                    authorizationApi.permissionsApi.accessToken = adminToken;
                }
            });

            await test.step('Verify error response', () => {
                chaiExpect(error.statusCode).to.equal(StatusCodes.UNAUTHORIZED);
                const errorMessage = AuthorizationUtils.extractErrorMessage(error.body);
                chaiExpect(errorMessage).to.include('authorization header is missing or invalid');
                logger.info(
                    'Verified HTTP 401 and error message contains "authorization header is missing or invalid"'
                );
            });
        });

        test('T5591168: Create Role with Single Permission Set', async function () {
            const roleName = objectGenerator.generateShortNameByPrefix('test-role-single-ps');

            await test.step('Create role with single permission set', async () => {
                createdRole = await authorizationUtils.createRoleWithPermissionSets(
                    roleName,
                    'Test role with single permission set',
                    [testPermissionSetId],
                    RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    tenantId,
                    true,
                    adminToken
                );
                createdRoleIds.push(createdRole.id);
                logger.info(`Created role ${createdRole.id} with permission set ${testPermissionSetId}`);
            });

            await test.step('Verify response contains required fields', () => {
                chaiExpect(createdRole.id).to.be.a('number');
                chaiExpect(createdRole.name).to.equal(roleName);
                chaiExpect(createdRole.permissionSets).to.be.an('array');
                chaiExpect(createdRole.permissions).to.be.an('array');
                logger.info('Verified role structure');
            });

            await test.step('Verify permission sets and permissions', () => {
                chaiExpect(createdRole.permissionSets.length).to.equal(1);
                chaiExpect(createdRole.permissionSets[0].id).to.equal(testPermissionSetId);
                chaiExpect(createdRole.permissionSets[0].name).to.equal(testPermissionSetName);
                chaiExpect(createdRole.permissions).to.be.an('array');
                chaiExpect(createdRole.permissions.length).to.be.greaterThan(0);

                const permissionsWithActions = createdRole.permissions.filter(
                    (p) => p.actions && p.actions.length > 0
                );
                chaiExpect(permissionsWithActions.length).to.be.greaterThan(0);

                createdRole.permissions.forEach((permission) => {
                    chaiExpect(permission).to.have.property('resourceType');
                    chaiExpect(permission).to.have.property('actions');
                    chaiExpect(permission.actions).to.be.an('array');
                });

                logger.info(
                    `Verified: 1 permission set, ${createdRole.permissions.length} total permissions, ${permissionsWithActions.length} with actions`
                );
            });
        });

        test('T5591169: Create Role with Multiple Permission Sets', async function () {
            const roleName = objectGenerator.generateShortNameByPrefix('test-role-multi-ps');

            await test.step('Create role with multiple permission sets', async () => {
                chaiExpect(secondPermissionSetId).to.exist;
                createdRole = await authorizationUtils.createRoleWithPermissionSets(
                    roleName,
                    'Test role with multiple permission sets',
                    [testPermissionSetId, secondPermissionSetId],
                    RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    tenantId,
                    true,
                    adminToken
                );
                createdRoleIds.push(createdRole.id);
                logger.info(`Created role ${createdRole.id} with 2 permission sets`);
            });

            await test.step('Verify permissionSets contains both permission sets', () => {
                chaiExpect(createdRole.permissionSets).to.be.an('array').with.lengthOf(2);
                const permissionSetIds = createdRole.permissionSets.map((ps) => ps.id);
                chaiExpect(permissionSetIds).to.include.members([testPermissionSetId, secondPermissionSetId]);

                createdRole.permissionSets.forEach((ps) => {
                    chaiExpect(ps).to.have.property('id');
                    chaiExpect(ps).to.have.property('name');
                });
                logger.info('Verified both permission sets are present');
            });

            await test.step('Verify permissions flat list contains merged permissions', () => {
                chaiExpect(createdRole.permissions).to.be.an('array');
                chaiExpect(createdRole.permissions.length).to.be.greaterThan(0);

                const permissionsWithActions = createdRole.permissions.filter(
                    (p) => p.actions && p.actions.length > 0
                );
                chaiExpect(permissionsWithActions.length).to.be.greaterThan(0);

                createdRole.permissions.forEach((permission) => {
                    chaiExpect(permission).to.have.property('resourceType');
                    chaiExpect(permission).to.have.property('actions');
                    chaiExpect(permission.actions).to.be.an('array');
                });

                logger.info(
                    `Verified ${createdRole.permissions.length} total permissions, ${permissionsWithActions.length} with actions`
                );
            });
        });

        test('T5591170: Create Role with System Scope', async function () {
            if (!superAdminToken) {
                test.skip(
                    settings.isSelfHosted,
                    'Skipping: Super admin token not available. This test requires super admin privileges.'
                );
                return;
            }
            const roleName = objectGenerator.generateShortNameByPrefix('test-role-system-scope');
            await test.step('Create system-scoped role with super admin', async () => {
                createdRole = await authorizationUtils.createRoleWithPermissionSets(
                    roleName,
                    'Test system-scoped role',
                    [testPermissionSetId],
                    RoleCreationFieldsInternal.ScopeTypeEnum.System,
                    tenantId,
                    true,
                    superAdminToken
                );
                createdRoleIds.push(createdRole.id);
                logger.info(`Created system-scoped role: ${createdRole.name}`);
            });

            await test.step('Verify system scope role properties', () => {
                chaiExpect(createdRole).to.have.property('id');
                chaiExpect(createdRole.name).to.equal(roleName);
                chaiExpect(createdRole.description).to.equal('Test system-scoped role');
                chaiExpect(createdRole.scopeType).to.equal(RoleCreationFieldsInternal.ScopeTypeEnum.System);
                chaiExpect(createdRole.scopeId).to.equal(tenantId);
                chaiExpect(createdRole.enabled).to.be.true;
                logger.info(`System-scoped role created successfully with scopeId=${tenantId}`);
            });

            await test.step('Verify permission set is attached', () => {
                chaiExpect(createdRole.permissionSets).to.be.an('array').with.lengthOf(1);
                chaiExpect(createdRole.permissionSets[0].id).to.equal(testPermissionSetId);
                logger.info('Permission set attached to system-scoped role');
            });
        });

        test('T5591171: Create Role with Enabled/Disabled State', async function () {
            let enabledRole: Role;
            let disabledRole: Role;

            await test.step('Create role with enabled: true', async () => {
                const roleName = objectGenerator.generateShortNameByPrefix('test-role-enabled');
                enabledRole = await authorizationUtils.createRoleWithPermissionSets(
                    roleName,
                    'Enabled test role',
                    [testPermissionSetId],
                    RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    tenantId,
                    true,
                    adminToken
                );
                createdRoleIds.push(enabledRole.id);
                logger.info(`Created enabled role: ${enabledRole.id}`);
            });

            await test.step('Verify enabled role state', () => {
                chaiExpect(enabledRole.enabled).to.be.true;
                chaiExpect(enabledRole.effectiveEnabled).to.be.true;
                chaiExpect(enabledRole.permissionSets).to.be.an('array').with.lengthOf(1);
                chaiExpect(enabledRole.permissions).to.be.an('array').and.not.be.empty;
                logger.info('Verified enabled role: enabled=true, effectiveEnabled=true');
            });

            await test.step('Create role with enabled: false', async () => {
                const roleName = objectGenerator.generateShortNameByPrefix('test-role-disabled');
                const roleCreationFields: RoleCreationFieldsInternal = {
                    name: roleName,
                    description: 'Disabled test role',
                    scopeType: RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    scopeId: tenantId,
                    enabled: false,
                    permissionSets: [{ id: testPermissionSetId } as RolePermissionSet]
                };
                const response = await authorizationApi.rolesApi.createRole(roleCreationFields);
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.CREATED);
                disabledRole = response.body;
                createdRoleIds.push(disabledRole.id);
                logger.info(`Created disabled role: ${disabledRole.id}`);
            });

            await test.step('Verify disabled role state', () => {
                chaiExpect(disabledRole.enabled).to.be.false;
                chaiExpect(disabledRole.effectiveEnabled).to.be.false;
                chaiExpect(disabledRole.permissionSets).to.be.an('array').with.lengthOf(1);
                chaiExpect(disabledRole.permissions).to.be.an('array').and.not.be.empty;
                logger.info('Verified disabled role: enabled=false, effectiveEnabled=false');
            });
        });

        test('T5591172: Create Role Response Includes Both Permission Sets and Flat List', async function () {
            const roleName = objectGenerator.generateShortNameByPrefix('test-role-both-structures');

            await test.step('Create role with permission set', async () => {
                createdRole = await authorizationUtils.createRoleWithPermissionSets(
                    roleName,
                    'Test role to verify both structures',
                    [testPermissionSetId],
                    RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    tenantId,
                    true,
                    adminToken
                );
                createdRoleIds.push(createdRole.id);
                logger.info(`Created role: ${createdRole.id}`);
            });

            await test.step('Verify response contains both permissionSets and permissions arrays', () => {
                chaiExpect(createdRole).to.have.property('permissionSets');
                chaiExpect(createdRole.permissionSets).to.be.an('array').with.lengthOf(1);

                chaiExpect(createdRole).to.have.property('permissions');
                chaiExpect(createdRole.permissions).to.be.an('array').and.not.be.empty;

                logger.info('Verified both permissionSets and permissions arrays exist');
            });

            await test.step('Verify permissionSets array structure', () => {
                const permissionSet = createdRole.permissionSets[0];
                chaiExpect(permissionSet).to.have.property('id');
                chaiExpect(permissionSet.id).to.equal(testPermissionSetId);
                chaiExpect(permissionSet).to.have.property('name');
                chaiExpect(permissionSet.name).to.equal(testPermissionSetName);
                logger.info(
                    `Verified permissionSets: id=${testPermissionSetId}, name=${testPermissionSetName}`
                );
            });

            await test.step('Verify permissions flat list structure', () => {
                const permissionsWithActions = createdRole.permissions.filter(
                    (p) => p.actions && p.actions.length > 0
                );
                chaiExpect(permissionsWithActions.length).to.be.greaterThan(0);

                createdRole.permissions.forEach((permission) => {
                    chaiExpect(permission).to.have.property('resourceType');
                    chaiExpect(permission).to.have.property('actions');
                    chaiExpect(permission.actions).to.be.an('array');
                });

                logger.info(
                    `Verified permissions flat list: ${createdRole.permissions.length} total, ${permissionsWithActions.length} with actions`
                );
            });

            await test.step('Verify both structures are consistent', () => {
                chaiExpect(createdRole.permissionSets[0].id).to.equal(testPermissionSetId);
                chaiExpect(createdRole.permissions.length).to.be.greaterThan(0);
                const uniqueResourceTypes = new Set(createdRole.permissions.map((p) => p.resourceType));
                chaiExpect(uniqueResourceTypes.size).to.equal(createdRole.permissions.length);
                logger.info('Verified both structures are consistent');
            });
        });

        test('T5591173: Create Role with Minimal Required Fields', async function () {
            const roleName = objectGenerator.generateShortNameByPrefix('test-role-minimal');

            await test.step('Create role with minimal fields (omit enabled field)', async () => {
                const roleCreationFields: RoleCreationFieldsInternal = {
                    name: roleName,
                    description: 'Minimal test role',
                    scopeType: RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    scopeId: tenantId,
                    permissionSets: [{ id: testPermissionSetId } as RolePermissionSet]
                };

                response = await authorizationApi.rolesApi.createRole(roleCreationFields);
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.CREATED);
                createdRole = response.body;
                createdRoleIds.push(createdRole.id);
                logger.info(`Created role ${createdRole.id} with minimal fields (enabled omitted)`);
            });

            await test.step('Verify enabled defaults to true', () => {
                chaiExpect(createdRole.enabled).to.be.true;
                chaiExpect(createdRole.effectiveEnabled).to.be.true;
                logger.info('Verified enabled defaults to true when not specified');
            });

            await test.step('Verify basic role properties', () => {
                chaiExpect(createdRole.name).to.equal(roleName);
                chaiExpect(createdRole.scopeType).to.equal(RoleCreationFieldsInternal.ScopeTypeEnum.Tenant);
                chaiExpect(createdRole.scopeId).to.equal(tenantId);
                logger.info('Verified basic role properties');
            });

            await test.step('Verify permission sets and permissions are populated', () => {
                chaiExpect(createdRole.permissionSets).to.be.an('array').with.lengthOf(1);
                chaiExpect(createdRole.permissionSets[0].id).to.equal(testPermissionSetId);
                chaiExpect(createdRole.permissions).to.be.an('array').and.not.be.empty;
                logger.info('Verified permission sets and permissions are populated');
            });
        });

        test('T5591174: Get All Roles with Permission Sets', async function () {
            let roles: Role[] = [];

            await test.step('Get all roles', async () => {
                const response = await authorizationApi.rolesApi.getRoles();
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.OK);
                if (Array.isArray(response.body)) {
                    roles = response.body;
                } else if (response.body.roles) {
                    roles = response.body.roles;
                }
                chaiExpect(roles).to.be.an('array').and.not.be.empty;
                logger.info(`Retrieved ${roles.length} roles`);
            });

            await test.step('Verify all roles have basic structure', () => {
                roles.forEach((role, index) => {
                    chaiExpect(role, `Role ${index}`).to.have.property('id');
                    chaiExpect(role, `Role ${index}`).to.have.property('name');
                    chaiExpect(role, `Role ${index}`).to.have.property('permissions');
                    chaiExpect(role.permissions, `Role ${index} permissions`).to.be.an('array');
                });
                logger.info('Verified all roles have basic structure (id, name, permissions)');
            });

            await test.step('Verify roles with permission sets', () => {
                const rolesWithPermissionSets = roles.filter(
                    (r) => r.permissionSets && r.permissionSets.length > 0
                );
                const legacyRoles = roles.filter((r) => !r.permissionSets || r.permissionSets.length === 0);

                logger.info(`Found ${rolesWithPermissionSets.length} roles with permission sets (new roles)`);
                logger.info(`Found ${legacyRoles.length} legacy roles (without permission sets)`);

                rolesWithPermissionSets.forEach((role) => {
                    chaiExpect(role.permissionSets).to.be.an('array').and.not.be.empty;
                    chaiExpect(role.permissions).to.be.an('array').and.not.be.empty;
                });

                logger.info(
                    'Verified roles with permission sets have both permissionSets and permissions arrays'
                );
            });
        });

        test('T5591175: Get Specific Role by ID', async function () {
            const roleName = objectGenerator.generateShortNameByPrefix('test-role-get-by-id');
            let roleId: number;

            await test.step('Create a test role', async () => {
                createdRole = await authorizationUtils.createRoleWithPermissionSets(
                    roleName,
                    'Test role for get by ID',
                    [testPermissionSetId],
                    RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    tenantId,
                    true,
                    adminToken
                );
                createdRoleIds.push(createdRole.id);
                roleId = createdRole.id;
                logger.info(`Created test role with ID: ${roleId}`);
            });

            await test.step('Get role by ID and verify structure', async () => {
                const response = await authorizationApi.rolesApi.getRole(roleId);
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.OK);
                const roleDetails: any = response.body;

                chaiExpect(roleDetails.id).to.equal(roleId);
                chaiExpect(roleDetails.name).to.equal(roleName);

                if (roleDetails.permissionSets) {
                    chaiExpect(roleDetails.permissionSets).to.be.an('array').with.lengthOf(1);
                    chaiExpect(roleDetails.permissionSets[0].id).to.equal(testPermissionSetId);
                }

                chaiExpect(roleDetails.permissions).to.be.an('array').and.not.be.empty;

                roleDetails.permissions.forEach((permission) => {
                    chaiExpect(permission).to.have.property('resourceType');
                    chaiExpect(permission).to.have.property('actions');
                });

                logger.info(
                    `Verified role ${roleId} structure with ${roleDetails.permissions.length} permissions`
                );
            });
        });

        test('T5591176: Get Roles Filters by Scope Type', async function () {
            const tenantRoleName = objectGenerator.generateShortNameByPrefix('test-role-tenant-scope');
            let createdTenantRoleId: number;
            let allRoles: Role[];

            await test.step('Create tenant-scoped role for testing', async () => {
                createdRole = await authorizationUtils.createRoleWithPermissionSets(
                    tenantRoleName,
                    'Test tenant-scoped role',
                    [testPermissionSetId],
                    RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    tenantId,
                    true,
                    adminToken
                );
                createdRoleIds.push(createdRole.id);
                createdTenantRoleId = createdRole.id;
                logger.info(`Created tenant-scoped role ${createdTenantRoleId}`);
            });

            await test.step('Get all roles', async () => {
                const response = await authorizationApi.rolesApi.getRoles();
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.OK);
                if (Array.isArray(response.body)) {
                    allRoles = response.body;
                } else if (response.body.roles) {
                    allRoles = response.body.roles;
                }
                chaiExpect(allRoles).to.be.an('array').and.not.be.empty;
                logger.info(`Retrieved ${allRoles.length} roles`);
            });

            await test.step('Verify roles can be filtered by scope type', () => {
                const rolesWithScopeType = allRoles.filter((r) => r.scopeType !== undefined);

                if (rolesWithScopeType.length === 0) {
                    logger.info('No roles have scopeType field - may be using legacy role structure');
                    return;
                }

                const tenantRoles = allRoles.filter((r) => r.scopeType === Role.ScopeTypeEnum.Tenant);
                const systemRoles = allRoles.filter((r) => r.scopeType === Role.ScopeTypeEnum.System);

                logger.info(
                    `Found ${tenantRoles.length} tenant-scoped roles and ${systemRoles.length} system-scoped roles`
                );

                const createdRoleInList = tenantRoles.find((r) => r.id === createdTenantRoleId);
                chaiExpect(createdRoleInList).to.exist;
                chaiExpect(createdRoleInList.scopeType).to.equal(Role.ScopeTypeEnum.Tenant);
                chaiExpect(createdRoleInList.scopeId).to.equal(tenantId);
                logger.info('Verified created role appears in tenant-scoped roles list');

                tenantRoles.forEach((role) => {
                    chaiExpect(role.scopeId).to.exist;
                    chaiExpect(role.scopeId).to.not.be.null;
                });

                if (systemRoles.length > 0) {
                    systemRoles.forEach((role) => {
                        chaiExpect(role.scopeId === null || role.scopeId === undefined).to.be.true;
                    });
                    logger.info('Verified system roles have null or undefined scopeId');
                }
                logger.info('Verified all roles have correct scope structure');
            });
        });

        test('T5591177: Verify Permission Resolution from Single Permission Set', async function () {
            const roleName = objectGenerator.generateShortNameByPrefix('test-role-perm-resolution');

            await test.step('Get permission set details', async () => {
                const response = await authorizationApi.permissionsApi.getPermissionSet(testPermissionSetId);
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.OK);
                permissionSet = response.body;
                logger.info(
                    `Retrieved permission set ${testPermissionSetId} with ${permissionSet.permissions.length} permissions`
                );
            });

            await test.step('Create role with this permission set', async () => {
                createdRole = await authorizationUtils.createRoleWithPermissionSets(
                    roleName,
                    'Test permission resolution',
                    [testPermissionSetId],
                    RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    tenantId,
                    true,
                    adminToken
                );
                createdRoleIds.push(createdRole.id);
                logger.info(`Created role ${createdRole.id}`);
            });

            await test.step('Verify permissions flat list matches permission set', () => {
                const setResourceTypes = permissionSet.permissions.map((p) => p.resourceType);
                const roleResourceTypes = createdRole.permissions.map((p) => p.resourceType);

                setResourceTypes.forEach((resourceType) => {
                    chaiExpect(roleResourceTypes).to.include(resourceType);
                });

                permissionSet.permissions.forEach((setPermission) => {
                    const rolePermission = createdRole.permissions.find(
                        (p) => p.resourceType === setPermission.resourceType
                    );
                    chaiExpect(
                        rolePermission,
                        `Permission for ${setPermission.resourceType} should exist in role`
                    ).to.exist;

                    VALID_ACTIONS.forEach((action) => {
                        const setHasAction = setPermission.actions.includes(action as any);
                        const roleHasAction = rolePermission.actions.includes(action as any);
                        if (setHasAction) {
                            chaiExpect(
                                roleHasAction,
                                `Action ${action} for ${setPermission.resourceType} should match`
                            ).to.be.true;
                        } else {
                            chaiExpect(
                                roleHasAction,
                                `Action ${action} for ${setPermission.resourceType} should not exist in role when not in permission set`
                            ).to.be.false;
                        }
                    });
                });

                logger.info('Verified permissions flat list matches permission set');
            });
        });

        test('T5591178: Verify Permission Resolution from Multiple Permission Sets', async function () {
            const roleName = objectGenerator.generateShortNameByPrefix('test-role-multi-perm-resolution');
            let permissionSet1: PermissionSet;
            let permissionSet2: PermissionSet;

            await test.step('Get both permission sets', async () => {
                chaiExpect(secondPermissionSetId).to.exist;

                const response1 = await authorizationApi.permissionsApi.getPermissionSet(testPermissionSetId);
                AuthorizationUtils.verifyStatusCode(response1, StatusCodes.OK);
                permissionSet1 = response1.body;

                const response2 =
                    await authorizationApi.permissionsApi.getPermissionSet(secondPermissionSetId);
                AuthorizationUtils.verifyStatusCode(response2, StatusCodes.OK);
                permissionSet2 = response2.body;

                logger.info(
                    `Retrieved 2 permission sets with ${permissionSet1.permissions.length} and ${permissionSet2.permissions.length} permissions`
                );
            });

            await test.step('Create role with both permission sets', async () => {
                createdRole = await authorizationUtils.createRoleWithPermissionSets(
                    roleName,
                    'Test multi permission resolution',
                    [testPermissionSetId, secondPermissionSetId],
                    RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    tenantId,
                    true,
                    adminToken
                );
                createdRoleIds.push(createdRole.id);
                logger.info(`Created role ${createdRole.id}`);
            });

            await test.step('Verify permissions from both sets are present', () => {
                const allSetResourceTypes = [
                    ...permissionSet1.permissions.map((p) => p.resourceType),
                    ...permissionSet2.permissions.map((p) => p.resourceType)
                ];
                const uniqueSetResourceTypes = [...new Set(allSetResourceTypes)];
                const roleResourceTypes = createdRole.permissions.map((p) => p.resourceType);

                uniqueSetResourceTypes.forEach((resourceType) => {
                    chaiExpect(
                        roleResourceTypes,
                        `Role should contain permission for ${resourceType}`
                    ).to.include(resourceType);
                });

                logger.info(
                    `Verified all ${uniqueSetResourceTypes.length} unique permissions from both sets are present in role`
                );
            });

            await test.step('Verify no duplicate resource type entries', () => {
                const roleResourceTypes = createdRole.permissions.map((p) => p.resourceType);
                const uniqueRoleResourceTypes = [...new Set(roleResourceTypes)];

                chaiExpect(roleResourceTypes.length).to.equal(uniqueRoleResourceTypes.length);
                logger.info('Verified no duplicate resource type entries');
            });
        });

        test('T5591179: Verify Pre-Defined Roles Structure', async function () {
            let allRoles: Role[];

            await test.step('Get all roles', async () => {
                const response = await authorizationApi.rolesApi.getRoles();
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.OK);
                if (Array.isArray(response.body)) {
                    allRoles = response.body;
                } else if (response.body.roles) {
                    allRoles = response.body.roles;
                }
                logger.info(`Retrieved ${allRoles.length} roles`);
            });

            await test.step('Filter and verify pre-defined roles', () => {
                const preDefinedRoles = allRoles.filter((r) => r.custom === false);
                const preDefinedWithSets = preDefinedRoles.filter(
                    (r) => r.permissionSets && r.permissionSets.length > 0
                );
                const legacyPreDefined = preDefinedRoles.filter(
                    (r) => !r.permissionSets || r.permissionSets.length === 0
                );

                chaiExpect(preDefinedRoles.length).to.be.greaterThan(
                    0,
                    'Should have at least one pre-defined role'
                );

                logger.info(`Found ${preDefinedRoles.length} pre-defined roles total`);
                logger.info(`  - ${preDefinedWithSets.length} with permission sets (new roles)`);
                logger.info(`  - ${legacyPreDefined.length} legacy roles (without permission sets)`);

                logger.info('\n=== All Pre-Defined Roles Summary ===');
                preDefinedRoles.forEach((role, index) => {
                    const summary = {
                        index: index + 1,
                        id: role.id,
                        name: role.name,
                        custom: role.custom,
                        scopeType: role.scopeType,
                        scopeId: role.scopeId,
                        enabled: role.enabled,
                        effectiveEnabled: role.effectiveEnabled,
                        hasPermissionSets: role.permissionSets && role.permissionSets.length > 0,
                        permissionSetsCount: role.permissionSets ? role.permissionSets.length : 0,
                        permissionsCount: role.permissions ? role.permissions.length : 0
                    };
                    logger.info(`${index + 1}. ${JSON.stringify(summary)}`);
                });

                preDefinedRoles.forEach((role) => {
                    chaiExpect(role.custom).to.be.false;
                    chaiExpect(role.id).to.be.a('number');
                    chaiExpect(role.permissions).to.be.an('array').and.not.be.empty;
                    chaiExpect(role.effectiveEnabled).to.be.a('boolean');
                });

                if (preDefinedWithSets.length > 0) {
                    preDefinedWithSets.forEach((role) => {
                        chaiExpect(role.permissionSets).to.be.an('array').and.not.be.empty;
                        role.permissionSets.forEach((ps) => {
                            chaiExpect(ps.id).to.be.a('string');
                            chaiExpect(ps.name).to.be.a('string');
                        });
                    });
                    logger.info(
                        `Verified ${preDefinedWithSets.length} new pre-defined roles with permission sets`
                    );
                }

                if (legacyPreDefined.length > 0) {
                    logger.info(
                        `Verified ${legacyPreDefined.length} legacy pre-defined roles (without permission sets)`
                    );
                }

                logger.info('Verified all pre-defined roles have required structure');
            });
        });

        test('T5591180: Create Role Without Permission Sets or Permissions', async function () {
            const roleName = objectGenerator.generateShortNameByPrefix('test-role-no-perms');

            await test.step('Attempt to create role without permissionSets or permissions', async () => {
                const roleCreationFields: any = {
                    name: roleName,
                    description: 'Test role without permissions',
                    scopeType: RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    scopeId: tenantId,
                    enabled: true
                };
                try {
                    const response = await authorizationApi.rolesApi.createRole(roleCreationFields);
                    createdRole = response.body;
                    if (createdRole.id) {
                        createdRoleIds.push(createdRole.id);
                    }
                    logger.info(`Unexpected: Role created with ID ${createdRole.id}`);
                } catch (err) {
                    error = err;
                }
            });

            await test.step('Verify API rejects role without permissions', () => {
                chaiExpect(error).to.exist;
                chaiExpect(error.statusCode).to.equal(StatusCodes.BAD_REQUEST);

                const errorMessage = AuthorizationUtils.extractErrorMessage(error.body);
                chaiExpect(errorMessage).to.include('permission');

                logger.info(`API correctly rejected role: ${errorMessage}`);
            });
        });

        test('T5591181: Create Role with Permissions Field (Backward Compatibility)', async function () {
            const roleName = objectGenerator.generateShortNameByPrefix('test-role-legacy-permissions');

            await test.step('Create role using permissions field (without permissionSets)', async () => {
                const roleCreationFields: any = {
                    name: roleName,
                    description: 'Role created with permissions field',
                    scopeType: RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    scopeId: tenantId,
                    enabled: true,
                    permissions: [
                        {
                            resourceType: 'workloads',
                            actions: ['create', 'read', 'update']
                        },
                        {
                            resourceType: 'project',
                            actions: ['read']
                        }
                    ]
                };

                const response = await authorizationApi.rolesApi.createRole(roleCreationFields);
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.CREATED);
                createdRole = response.body;
                createdRoleIds.push(createdRole.id);
                logger.info(`Created role ${createdRole.id} using legacy permissions field`);
            });

            await test.step('Verify role structure', () => {
                chaiExpect(createdRole.permissions).to.be.an('array').and.not.be.empty;

                if (createdRole.permissionSets !== undefined) {
                    chaiExpect(createdRole.permissionSets).to.be.an('array').and.be.empty;
                } else {
                    logger.info('permissionSets field is undefined (legacy role)');
                }

                const workloadPerm = createdRole.permissions.find(
                    (p) => p.resourceType === ('workloads' as any)
                );
                chaiExpect(workloadPerm).to.exist;
                chaiExpect(workloadPerm.actions).to.include.members(['create', 'read', 'update']);

                const projectPerm = createdRole.permissions.find(
                    (p) => p.resourceType === ('project' as any)
                );
                chaiExpect(projectPerm).to.exist;
                chaiExpect(projectPerm.actions).to.include('read');

                logger.info('Verified backward compatibility with permissions field');
            });
        });

        test('T5591182: Create Role with Invalid Permission Set ID', async () => {
            const roleName = objectGenerator.generateShortNameByPrefix('test-role-invalid-ps');
            const invalidPermissionSetId = randomUUID();

            await test.step('Attempt to create role with invalid permission set ID', async () => {
                try {
                    createdRole = await authorizationUtils.createRoleWithPermissionSets(
                        roleName,
                        'Test role with invalid PS',
                        [invalidPermissionSetId],
                        RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                        tenantId,
                        true,
                        adminToken
                    );
                    createdRoleIds.push(createdRole.id);
                } catch (err) {
                    error = err;
                    logger.info(`API rejected invalid permission set: ${err.statusCode}`);
                }
            });

            await test.step('Verify error response', () => {
                chaiExpect(error).to.exist;
                const expectedStatusCodes = [StatusCodes.BAD_REQUEST, StatusCodes.NOT_FOUND];

                chaiExpect(
                    error.statusCode,
                    `Expected status code to be 400 or 404 for invalid Permission Set ID, but got ${error.statusCode}`
                ).to.be.oneOf(expectedStatusCodes);

                const errorMessage = AuthorizationUtils.extractErrorMessage(error.body);
                chaiExpect(errorMessage).to.include('permission sets not found');
            });
        });

        test('T5591184: Create Role with Missing Required Fields', async () => {
            let createdSuccessfully = false;
            await test.step('Attempt to create role without name', async () => {
                const uniqueDescription = `Test role without name - ${Date.now()}`;
                const roleCreationFields: any = {
                    description: uniqueDescription,
                    scopeType: RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    scopeId: tenantId,
                    permissionSets: [{ id: testPermissionSetId }]
                };

                try {
                    response = await authorizationApi.rolesApi.createRole(roleCreationFields);
                    createdSuccessfully = true;
                    if (response.body?.id) {
                        createdRoleIds.push(response.body.id);
                    }
                } catch (err) {
                    error = err;
                }

                if (createdSuccessfully) {
                    chaiExpect.fail(
                        `Role creation should have failed for missing name field, but succeeded with status ${response.response.statusCode}`
                    );
                } else {
                    chaiExpect(error).to.exist;
                    chaiExpect(error.statusCode).to.equal(StatusCodes.BAD_REQUEST);
                }
            });

            await test.step('Attempt to create role without description', async () => {
                const roleCreationFields: any = {
                    name: objectGenerator.generateShortNameByPrefix('test-role-no-description'),
                    scopeType: RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    scopeId: tenantId,
                    permissionSets: [{ id: testPermissionSetId }]
                };

                try {
                    await authorizationApi.rolesApi.createRole(roleCreationFields);
                } catch (err) {
                    error = err;
                }

                chaiExpect(error).to.exist;
                chaiExpect(error.statusCode).to.be.oneOf([StatusCodes.BAD_REQUEST, StatusCodes.CONFLICT]);
                logger.info(`Verified description is required (API returned ${error.statusCode})`);
            });
        });

        test('T5591185: Get Permission Set Without Proper Authorization', async function () {
            await test.step('Attempt to get permission sets without authentication', async () => {
                const originalToken = authorizationApi.permissionsApi.accessToken;
                try {
                    authorizationApi.permissionsApi.accessToken = '';
                    await authorizationApi.permissionsApi.getPermissionSets();
                } catch (err) {
                    error = err;
                } finally {
                    authorizationApi.permissionsApi.accessToken = originalToken;
                }
            });

            await test.step('Verify error response', () => {
                chaiExpect(error).to.exist;
                chaiExpect(error.statusCode).to.equal(StatusCodes.UNAUTHORIZED);
                const errorMessage = AuthorizationUtils.extractErrorMessage(error.body);
                chaiExpect(errorMessage).to.include('authorization header is missing or invalid');
                logger.info(
                    `Verified GET permission sets without auth: ${error.statusCode} - ${errorMessage}`
                );
            });

            await test.step('Attempt to get specific permission set without authentication', async () => {
                const originalToken = authorizationApi.permissionsApi.accessToken;
                try {
                    authorizationApi.permissionsApi.accessToken = '';
                    await authorizationApi.permissionsApi.getPermissionSet(testPermissionSetId);
                } catch (err) {
                    error = err;
                } finally {
                    authorizationApi.permissionsApi.accessToken = originalToken;
                }

                chaiExpect(error).to.exist;
                chaiExpect(error.statusCode).to.equal(StatusCodes.UNAUTHORIZED);
                const specificErrorMessage = AuthorizationUtils.extractErrorMessage(error.body);
                chaiExpect(specificErrorMessage).to.include('authorization header is missing or invalid');
                logger.info(
                    `Verified GET specific permission set without auth: ${error.statusCode} - ${specificErrorMessage}`
                );
            });
        });

        test('T5591186: Permission Set with Minimal Permissions', async function () {
            let allPermissionSets: PermissionSet[];
            let minimalPermissionSet: PermissionSet;

            await test.step('Find permission set with minimal actions', async () => {
                response = await authorizationApi.permissionsApi.getPermissionSets();
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.OK);
                allPermissionSets = response.body.permissionSets;

                minimalPermissionSet = allPermissionSets.reduce((min, ps: any) => {
                    const countActions = (permSet: any): number => {
                        if (!permSet.permissions) return 0;
                        return permSet.permissions.reduce((count: number, perm: any) => {
                            return count + (perm.actions?.length || 0);
                        }, 0);
                    };
                    return countActions(ps) < countActions(min) ? ps : min;
                }, allPermissionSets[0]);

                const actionCount =
                    minimalPermissionSet.permissions?.reduce(
                        (count, perm: any) => count + (perm.actions?.length || 0),
                        0
                    ) || 0;

                logger.info(
                    `Using permission set with minimal actions: "${minimalPermissionSet.name}", ${actionCount} total actions`
                );
            });

            await test.step('Create role with minimal permission set', async () => {
                const roleName = objectGenerator.generateShortNameByPrefix('test-role-minimal-ps');
                createdRole = await authorizationUtils.createRoleWithPermissionSets(
                    roleName,
                    'Test role with minimal permission set',
                    [minimalPermissionSet.id],
                    RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    tenantId,
                    true,
                    adminToken
                );
                createdRoleIds.push(createdRole.id);
                logger.info(`Created role ${createdRole.id} with minimal permission set`);
            });

            await test.step('Verify role structure with minimal permissions', () => {
                chaiExpect(createdRole.permissionSets).to.be.an('array').with.length(1);
                chaiExpect(createdRole.permissionSets[0].id).to.equal(minimalPermissionSet.id);
                chaiExpect(createdRole.permissionSets[0].name).to.equal(minimalPermissionSet.name);

                // System populates all resource types
                chaiExpect(createdRole.permissions).to.be.an('array').and.not.be.empty;

                const permissionsWithActions = createdRole.permissions.filter(
                    (perm: any) => perm.actions?.length > 0
                );

                // Minimal permission set: most permissions should have empty actions
                const totalPermissions = createdRole.permissions.length;
                chaiExpect(permissionsWithActions.length).to.be.lessThan(totalPermissions);

                logger.info(
                    `Verified: ${createdRole.permissions.length} resource types, only ${permissionsWithActions.length} with actions`
                );
            });
        });

        test('T5591187: Role Name and Description with Special Characters', async function () {
            const specialCharsName = objectGenerator.generateShortNameByPrefix('test-role-!@#$%');
            const specialCharsDescription = 'Test role with special characters: !@#$%^&*()_+-=[]{}|;:,.<>?';
            let specialCharsRoleId: number;
            let longNameRoleId: number;
            let longName: string;

            await test.step('Create role with special characters', async () => {
                createdRole = await authorizationUtils.createRoleWithPermissionSets(
                    specialCharsName,
                    specialCharsDescription,
                    [testPermissionSetId],
                    RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    tenantId,
                    true,
                    adminToken
                );
                createdRoleIds.push(createdRole.id);
                specialCharsRoleId = createdRole.id;
                logger.info(`Created role ${createdRole.id} with special characters`);
            });

            await test.step('Verify special characters are preserved', async () => {
                const response = await authorizationApi.rolesApi.getRole(specialCharsRoleId);
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.OK);
                const retrievedRole: any = response.body;
                chaiExpect(retrievedRole.name).to.equal(specialCharsName);
                chaiExpect(retrievedRole.description).to.equal(specialCharsDescription);
                logger.info('Verified special characters preserved in name and description');
            });

            await test.step('Create role with very long name', async () => {
                longName = objectGenerator.generateShortNameByPrefix('test-role-') + 'a'.repeat(500);
                createdRole = await authorizationUtils.createRoleWithPermissionSets(
                    longName,
                    'Test very long name',
                    [testPermissionSetId],
                    RoleCreationFieldsInternal.ScopeTypeEnum.Tenant,
                    tenantId,
                    true,
                    adminToken
                );
                createdRoleIds.push(createdRole.id);
                longNameRoleId = createdRole.id;
                logger.info(`Created role with long name (${longName.length} chars): ${createdRole.id}`);
                logger.info(`Long name used for creation: ${longName.substring(0, 50)}...`);
            });

            await test.step('Verify long name is preserved', async () => {
                const response = await authorizationApi.rolesApi.getRole(longNameRoleId);
                AuthorizationUtils.verifyStatusCode(response, StatusCodes.OK);
                const retrievedRole: any = response.body;
                logger.info(`Expected name: ${longName.substring(0, 50)}...`);
                logger.info(`Retrieved name: ${retrievedRole.name.substring(0, 50)}...`);
                logger.info(
                    `Expected length: ${longName.length}, Retrieved length: ${retrievedRole.name.length}`
                );
                chaiExpect(retrievedRole.name).to.equal(longName);
                logger.info(`Verified long name preserved (${retrievedRole.name.length} chars)`);
            });
        });
    });
});
