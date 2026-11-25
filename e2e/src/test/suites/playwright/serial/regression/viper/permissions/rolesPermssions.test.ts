/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authorizationUtils,
    authUtils,
    clusterUtils,
    departmentUtils,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    operations,
    orgUnitUtils,
    projectResourcesUtils,
    serviceAccountUtils,
    settings,
    tags,
    Utils,
    usersManager,
    versionUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { Department } from '../../../../../../../modules/api/apis/orgUnitApi';
import { ScopeType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { PROPERTY_TO_ROLE_MAPPING, ROLE_NAME_TO_PROPERTY } from '../../../../../../../modules/helpers/config/roleMapping';
import { SubjectType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { AccessRule } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/accessRule';
import {
    backendClusterApi,
    clusterApi,
    identityManagerApi,
    orgUnitService,
    projectsDepartmentsApi
} from '../../../../../../../modules/api';
import { StatusCodes } from 'http-status-codes';
import { NodePool } from '../../../../../../../modules/api/apis/backendClusterApi';
import { ResourceType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/resourceType';
import { Role } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/role';
import { RolePermission } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/rolePermission';
import { PermissionAction } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/permissionAction';
import { test, expect } from '@playwright/test';
import { ProjectUpdateRequest } from '../../../../../../../modules/api/apis/orgUnitApi';

type testRolePermissionsParams = {
    role: string;
    scope: ScopeType;
};

async function validatePermissionOnFunction(
    permission: RolePermission,
    requiredAction: PermissionAction,
    requiredResourceType: ResourceType,
    roleName: string,
    func: () => any
): Promise<any> {
    const hasRequiredAction = permission && permission.actions.some((action) => action === requiredAction);
    try {
        const response = await func();
        if (!hasRequiredAction) {
            throw new Error(
                `role ${roleName} should not have permission to ${requiredAction} ${requiredResourceType}`
            );
        }
        return response;
    } catch (err) {
        if (hasRequiredAction) {
            throw err;
        }
        expect([StatusCodes.FORBIDDEN, StatusCodes.NOT_FOUND]).toContain(err.statusCode);
    }
    return undefined;
}

async function validatePermissionsToServiceAccounts(
    role: Role,
    scopeType: ScopeType,
    scopeId: string,
    userToken: string
): Promise<void> {
    identityManagerApi.serviceAccountsApi.accessToken = userToken;
    const permission = role.permissions.find(
        (permission) => permission.resourceType === ResourceType.ServiceAccount
    );
    const serviceAccountName = objectGenerator.generateShortNameByPrefix('roles-perm-test-sa');

    const createResponse = await validatePermissionOnFunction(
        permission,
        PermissionAction.Create,
        ResourceType.ServiceAccount,
        role.name,
        async () =>
            await serviceAccountUtils.createServiceAccount(userToken, { name: serviceAccountName })
    );

    const sa = createResponse
        ? createResponse.body
        : { name: serviceAccountName, id: 'some-uuid-string' };

    await validatePermissionOnFunction(
        permission,
        PermissionAction.Read,
        ResourceType.ServiceAccount,
        role.name,
        async () => await serviceAccountUtils.getServiceAccounts(userToken)
    );

    await validatePermissionOnFunction(
        permission,
        PermissionAction.Update,
        ResourceType.ServiceAccount,
        role.name,
        async () => await serviceAccountUtils.updateServiceAccountById(userToken, sa.id, sa)
    );

    await validatePermissionOnFunction(
        permission,
        PermissionAction.Delete,
        ResourceType.ServiceAccount,
        role.name,
        async () => await serviceAccountUtils.deleteServiceAccount(userToken, sa.id)
    );
}

async function validatePermissionsToNodePools(
    role: Role,
    scopeType: ScopeType,
    scopeId: string,
    userToken: string
): Promise<void> {
    clusterApi.nodepoolsApi.accessToken = userToken;
    const permission = role.permissions.find(
        (permission) => permission.resourceType === ResourceType.Nodepools
    );
    // If the scope is lower than cluster, we expect the Read permission to apply
    if (permission && scopeType !== ScopeType.Tenant && scopeType !== ScopeType.Cluster) {
        permission.actions = permission.actions.filter((action) => action === PermissionAction.Read);
    }

    const nodePoolName = objectGenerator.generateNameByPrefix('roles-permissions-test-np');
    const nodePoolLabelKey = objectGenerator.generateShortNameByPrefix('k');
    const nodePoolLabelValue = objectGenerator.generateShortNameByPrefix('v');
    logger.info(
        `Testing node pool permissions for role ${role.name} on node pool ${nodePoolName}, label ${nodePoolLabelKey}=${nodePoolLabelValue}`
    );
    const mockedNp: NodePool = {
        name: nodePoolName,
        labelKey: nodePoolLabelKey,
        labelValue: nodePoolLabelValue,
        id: 100000
    };
    const createResponse = await validatePermissionOnFunction(
        permission,
        PermissionAction.Create,
        ResourceType.Nodepools,
        role.name,
        async () => {
            const np = await nodePoolUtils.createClusterNodePool(
                mockedNp.name,
                mockedNp.labelKey,
                mockedNp.labelValue,
                {},
                userToken
            );
            await nodePoolUtils.waitForNodePoolCrdReadyOrEmpty(mockedNp.name);
            return np;
        }
    );
    const np = createResponse ? createResponse.body : mockedNp;
    await validatePermissionOnFunction(
        permission,
        PermissionAction.Read,
        ResourceType.Nodepools,
        role.name,
        async () => await clusterApi.nodepoolsApi.getNodepools()
    );

    // The create-response schedulingConfiguration may carry the new
    // gpuPodsPlacementStrategy.device field (CP defaults them). Cluster < 2.26 rejects
    // this fields, so strip them when round-tripping the response into a PUT.
    const updateSchedulingConfiguration = np.schedulingConfiguration
        ? { ...np.schedulingConfiguration }
        : undefined;
    if (
        updateSchedulingConfiguration?.gpuPodsPlacementStrategy?.device !== undefined
    ) {
        updateSchedulingConfiguration.gpuPodsPlacementStrategy = {
            ...updateSchedulingConfiguration.gpuPodsPlacementStrategy
        };
        delete updateSchedulingConfiguration.gpuPodsPlacementStrategy.device;
    }

    await validatePermissionOnFunction(
        permission,
        PermissionAction.Update,
        ResourceType.Nodepools,
        role.name,
        async () =>
            await clusterApi.nodepoolsApi.updateNodepool(String(np.id), {
                labelKey: np.labelKey,
                labelValue: np.labelValue,
                schedulingConfiguration: updateSchedulingConfiguration
            })
    );

    await validatePermissionOnFunction(
        permission,
        PermissionAction.Delete,
        ResourceType.Nodepools,
        role.name,
        async () => await clusterApi.nodepoolsApi.deleteNodepool(String(np.id))
    );
}

async function validatePermissionsToClustersMinimal(
    role: Role,
    scopeType: ScopeType,
    scopeId: string,
    userToken: string
): Promise<void> {
    const permission = role.permissions.find(
        (permission) => permission.resourceType === ResourceType.ClustersMinimal
    );

    // clusters-minimal is read-only, so we only test Read action
    // Test the list endpoint: /api/v1/clusters/minimal
    await validatePermissionOnFunction(
        permission,
        PermissionAction.Read,
        ResourceType.ClustersMinimal,
        role.name,
        async () => await clusterUtils.getClustersMinimalDataApi(undefined, undefined, undefined, userToken)
    );

    // Test the get-by-id endpoint: /api/v1/clusters/{clusterUuid}/minimal
    await validatePermissionOnFunction(
        permission,
        PermissionAction.Read,
        ResourceType.ClustersMinimal,
        role.name,
        async () => await clusterUtils.getClusterMinimalDataByIdApi(flowRootState.cluster.uuid, userToken)
    );
}

async function validatePermissionsToNodePoolsMinimal(
    role: Role,
    scopeType: ScopeType,
    scopeId: string,
    userToken: string
): Promise<void> {
    const permission = role.permissions.find(
        (permission) => permission.resourceType === ResourceType.NodepoolsMinimal
    );

    // nodepools-minimal is read-only, so we only test Read action
    // Test the list endpoint: /api/v1/node-pools/minimal
    const listResponse = await validatePermissionOnFunction(
        permission,
        PermissionAction.Read,
        ResourceType.NodepoolsMinimal,
        role.name,
        async () =>
            await nodePoolUtils.getNodepoolsMinimalApi(
                undefined,
                undefined,
                undefined,
                undefined,
                undefined,
                userToken
            )
    );

    // Test the get-by-id endpoint: /api/v1/node-pools/{nodepoolId}/minimal
    // Prefer to read default nodepool if exists
    if (listResponse && listResponse.body.nodepools && listResponse.body.nodepools.length > 0) {
        const np =
            listResponse.body.nodepools.find((n) => n.name === 'default') ??
            listResponse.body.nodepools[0];
        const nodepoolId = np.id;
        await validatePermissionOnFunction(
            permission,
            PermissionAction.Read,
            ResourceType.NodepoolsMinimal,
            role.name,
            async () => await nodePoolUtils.getNodepoolMinimalApi(nodepoolId, userToken)
        );
    }
}

test.describe(
    `Roles access permissions tests, tags: ${tags.resource.roles} ${tags.tool.api} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit}`,
    { tag: [tags.services.authorization, tags.runOn.serialCI] },
    () => {
        const departmentsCount = 2,
            projectsPerDepartmentCount = 2,
            departments: Department[] = [],
            projects: Project[] = [];
        let testUserEmail: string, testUserToken: string, createdAccessRule: AccessRule;

        function getScopeId(scope: ScopeType): string {
            switch (scope) {
                case ScopeType.Tenant:
                    return flowRootState.cluster.tenantId.toString();
                case ScopeType.Cluster:
                    return flowRootState.cluster.uuid;
                case ScopeType.Department:
                    return departments[0].id.toString();
                case ScopeType.Project:
                    return projects[0].id.toString();
            }
        }

        async function validatePermissionsToDepartments(
            role: Role,
            scopeType: ScopeType,
            scopeId: string,
            userToken: string
        ): Promise<void> {
            projectsDepartmentsApi.departmentsApi.accessToken = userToken;
            orgUnitService.departmentApi.accessToken = userToken;
            const permission = role.permissions.find(
                (permission) => permission.resourceType === ResourceType.Department
            );
            if (permission) {
                if (scopeType === ScopeType.Department || scopeType === ScopeType.Project) {
                    permission.actions = permission.actions.filter(
                        (action) => action === PermissionAction.Read
                    );
                }
            }

            const getResponse = await validatePermissionOnFunction(
                permission,
                PermissionAction.Read,
                ResourceType.Department,
                role.name,
                async () =>
                    await projectsDepartmentsApi.departmentsApi.getDepartments(
                        flowRootState.cluster.uuid,
                        true
                    )
            );
            if (getResponse) {
                const createdDepartmentIds = new Set(departments.map((d) => Number(d.id)));
                const visibleCreatedDepartments = getResponse.body.filter((d) =>
                    createdDepartmentIds.has(Number(d.id))
                );
                switch (scopeType) {
                    case ScopeType.Tenant:
                    case ScopeType.Cluster:
                        expect(visibleCreatedDepartments.length).toBe(departmentsCount);
                        break;
                    case ScopeType.Department:
                        expect(
                            visibleCreatedDepartments.filter(
                                (d) => Number(d.id) === Number(departments[0].id)
                            ).length
                        ).toBe(1);
                        break;
                    case ScopeType.Project:
                        expect(visibleCreatedDepartments.length).toBe(0);
                        break;
                }
            }

            const createResponse = await validatePermissionOnFunction(
                permission,
                PermissionAction.Create,
                ResourceType.Department,
                role.name,
                async () => await departmentUtils.createBackendDepartment(0, 0, undefined, userToken)
            );
            const dep: Department = createResponse ? createResponse.body : { id: 11111111 };

            await validatePermissionOnFunction(
                permission,
                PermissionAction.Update,
                ResourceType.Department,
                role.name,
                async () => await orgUnitUtils.updateDepartment(dep.id, dep)
            );

            await validatePermissionOnFunction(
                permission,
                PermissionAction.Delete,
                ResourceType.Department,
                role.name,
                async () => await departmentUtils.deleteDepartment(Number(dep.id))
            );
        }

        async function validatePermissionsToProjects(
            role: Role,
            scopeType: ScopeType,
            scopeId: string,
            userToken: string
        ): Promise<void> {
            const permission = role.permissions.find(
                (permission) => permission.resourceType === ResourceType.Project
            );
            if (permission) {
                if (scopeType === ScopeType.Project) {
                    permission.actions = permission.actions.filter(
                        (action) => action === PermissionAction.Read
                    );
                }
            }

            const getResponse = await validatePermissionOnFunction(
                permission,
                PermissionAction.Read,
                ResourceType.Project,
                role.name,
                async () => await orgUnitUtils.getProjectsByClusterId(flowRootState.cluster.uuid, userToken)
            );
            if (getResponse) {
                const departmentIds = new Set(departments.map((d) => Number(d.id)));
                const projectIds = new Set(projects.map((p) => Number(p.id)));
                const visibleCreatedProjects = getResponse.filter((p) =>
                    departmentIds.has(Number(p.parentId))
                );
                const visibleProjectsInFirstDepartment = visibleCreatedProjects.filter(
                    (p) => Number(p.parentId) === Number(departments[0].id)
                );
                const visibleCreatedSpecificProject = getResponse.filter((p) => projectIds.has(Number(p.id)));
                switch (scopeType) {
                    case ScopeType.Tenant:
                    case ScopeType.Cluster:
                        expect(visibleCreatedProjects.length).toBe(
                            departmentsCount * projectsPerDepartmentCount
                        );
                        break;
                    case ScopeType.Department:
                        expect(visibleProjectsInFirstDepartment.length).toBe(projectsPerDepartmentCount);
                        break;
                    case ScopeType.Project:
                        expect(visibleCreatedSpecificProject.length).toBe(1);
                        break;
                }
            }

            const projectName = objectGenerator.generateNameByPrefix('p');
            const namespace = Utils.namespaceFromProjectName(projectName);
            const createResponse = await validatePermissionOnFunction(
                permission,
                PermissionAction.Create,
                ResourceType.Project,
                role.name,
                async () =>
                    await projectResourcesUtils.createProject(
                        namespace,
                        {
                            departmentId: Number(departments[0].id),
                            name: projectName,
                            deservedGpus: 10,
                            maxAllowedGpus: 10
                        },
                        userToken,
                        true,
                        true
                    )
            );
            const proj: Project = createResponse
                ? createResponse
                : { id: 11111111, departmentId: Number(departments[0].id) };

            await validatePermissionOnFunction(
                permission,
                PermissionAction.Update,
                ResourceType.Project,
                role.name,
                async () => {
                    const orgUnitProjects = await orgUnitUtils.getProjectsByClusterId(
                        flowRootState.cluster.uuid,
                        userToken
                    );
                    const orgUnitProject = orgUnitProjects.find((p) => p.name === projectName);
                    const resourcesForUpdate = orgUnitProject?.resources ?? [];
                    await orgUnitUtils.updateProject(
                        proj.id.toString(),
                        {
                            description: 'update',
                            resources: resourcesForUpdate
                        } as ProjectUpdateRequest,
                        userToken
                    );
                }
            );

            await validatePermissionOnFunction(
                permission,
                PermissionAction.Delete,
                ResourceType.Project,
                role.name,
                async () => await orgUnitUtils.deleteProject(proj.id.toString(), userToken)
            );
        }

        async function testRolePermissions(params: testRolePermissionsParams): Promise<void> {
            if (createdAccessRule) {
                try {
                    await authorizationUtils.deleteAccessRule(createdAccessRule);
                    createdAccessRule = null;
                } catch (error) {
                    logger.error(
                        'failed to delete existing access rule, skipping current test. error: ',
                        error
                    );
                    return;
                }
            }

            const role = flowRootState.users[ROLE_NAME_TO_PROPERTY[params.role]].role;
            const scopeId = getScopeId(params.scope);
            createdAccessRule = await authorizationUtils.createAccessRuleWithDefaults(
                testUserEmail,
                SubjectType.User,
                role.id,
                scopeId,
                params.scope
            );
            await authorizationUtils.waitForAccessRule(createdAccessRule.id, true);

            console.log(`testRolePermissions: role :${params.role} scope=${params.scope}`);
            await validatePermissionsToServiceAccounts(role, params.scope, scopeId, testUserToken);
            await validatePermissionsToDepartments(role, params.scope, scopeId, testUserToken);
            await validatePermissionsToProjects(role, params.scope, scopeId, testUserToken);
            await validatePermissionsToClustersMinimal(role, params.scope, scopeId, testUserToken);
            // NodePools must run before NodePoolsMinimal — the minimal test reads a nodepool
            // by ID from the list response, but afterEach deletes all nodepools between iterations.
            // Running NodePools first ensures a freshly created nodepool exists for Minimal to read.
            await validatePermissionsToNodePools(role, params.scope, scopeId, testUserToken);
            await validatePermissionsToNodePoolsMinimal(role, params.scope, scopeId, testUserToken);
        }

        async function createProject(department: Department): Promise<void> {
            const projectName = objectGenerator.generateNameByPrefix('p');
            const namespace = Utils.namespaceFromProjectName(projectName);

            const project = await projectResourcesUtils.createProject(
                namespace,
                {
                    departmentId: Number(department.id),
                    name: projectName,
                    deservedGpus: 10,
                    maxAllowedGpus: 10
                },
                null,
                true,
                true
            );
            projects.push(project);
        }

        async function createDepartmentWithProjects(): Promise<void> {
            const depName = objectGenerator.generateNameByPrefix('d');
            const department = (await departmentUtils.createBackendDepartmentWithName(depName)).body;

            expect(department).toBeDefined();

            const projectsPromises: Promise<any>[] = [];
            for (let i = 0; i < projectsPerDepartmentCount; i++) {
                projectsPromises.push(createProject(department));
            }

            await Promise.all(projectsPromises);

            departments.push(department);
        }

        test.beforeAll(async () => {
            if (!settings.enableAuth) {
                test.skip();
            }

            await departmentUtils.departmentsCleanup();

            projectsDepartmentsApi.departmentsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            orgUnitService.departmentApi.accessToken = flowRootState.users.system_administrator.accessToken;

            const promises: Promise<any>[] = [];
            for (let i = 0; i < departmentsCount; i++) {
                promises.push(createDepartmentWithProjects());
            }
            await Promise.all(promises);

            // Stabilize: ensure truly unique email to avoid 409 on CI re-runs
            testUserEmail = objectGenerator.generateEmailByPrefixWithRandom('u', 12);

            const user = (
                await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    testUserEmail
                )
            ).body;
            expect(user).toBeDefined();
            const token = await authUtils.authenticateKeycloakAndGetToken(user.username, user.tempPassword);
            expect(token).toBeDefined();
            testUserToken = token.idToken;
        });

        test.afterAll(async () => {
            try {
                await nodePoolUtils.deleteAllNodePools();
                await operations.forceCleanup();
            } catch (e) {
                console.warn(`Failed to cleanup. error: ${e}`);
            }
        });

        test.describe('test all roles in all scopes', () => {
            test.afterEach(async () => {
                if (createdAccessRule) {
                    await authorizationUtils.deleteAccessRule(createdAccessRule);
                    createdAccessRule = null;
                }

                backendClusterApi.nodePoolsApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;
                await nodePoolUtils.deleteAllNodePools();
            });

            const tests: { name: string; params: testRolePermissionsParams }[] = [];
            const scopes: ScopeType[] = [
                ScopeType.Tenant,
                ScopeType.Cluster,
                ScopeType.Department,
                ScopeType.Project
            ];
            const roles: string[] = [
                PROPERTY_TO_ROLE_MAPPING.compute_resource_administrator,
                PROPERTY_TO_ROLE_MAPPING.credentials_administrator,
                PROPERTY_TO_ROLE_MAPPING.data_source_administrator,
                PROPERTY_TO_ROLE_MAPPING.data_volume_administrator,
                PROPERTY_TO_ROLE_MAPPING.department_admin,
                PROPERTY_TO_ROLE_MAPPING.department_viewer,
                PROPERTY_TO_ROLE_MAPPING.editor,
                PROPERTY_TO_ROLE_MAPPING.environment_administrator,
                PROPERTY_TO_ROLE_MAPPING.l1_researcher,
                PROPERTY_TO_ROLE_MAPPING.l2_researcher,
                PROPERTY_TO_ROLE_MAPPING.ml_engineer,
                PROPERTY_TO_ROLE_MAPPING.research_manager,
                PROPERTY_TO_ROLE_MAPPING.system_administrator,
                PROPERTY_TO_ROLE_MAPPING.template_administrator,
                PROPERTY_TO_ROLE_MAPPING.viewer
            ];
            roles.forEach((role) =>
                scopes.forEach((scope) => {
                    tests.push({ name: `test ${role} on ${scope}`, params: { role: role, scope: scope } });
                })
            );

            tests.forEach((testCase) => {
                test(testCase.name, async () => {
                    await testRolePermissions(testCase.params);
                });
            });
        });
    }
);
