/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    departmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    accessRulesClusterUtils,
    authV2Utils,
    authorizationUtils,
    projectUtils,
    upgradeClusterUtils,
    orgUnitUtils,
    usersManager
} from '../../../../../../../modules/helpers';
import {
    authorizationApi,
    backendClusterApi,
    identityManagerApi,
    orgUnitService,
    projectsDepartmentsApi
} from '../../../../../../../modules/api';
import { Department, Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { Department as DepartmentOU } from '../../../../../../../modules/api/apis/orgUnitApi';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../../modules/helpers/config/roleMapping';
import { Role } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/role';
import { UserPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/userPostResponse';
import { ScopeType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { SubjectType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { AccessRule } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/accessRule';
import { test, expect } from '@playwright/test';
import { namespaceCluster } from '../../../../../../../modules/helpers/config/const';
import { StatusCodes } from 'http-status-codes';
import { Cluster } from '../../../../../../../modules/api/apis/backendClusterApi';

test.describe(
    `AccessRules API tests, tags: ${tags.resource.accessRules} ${tags.teams.viper}`,
    { tag: [tags.services.authorization, tags.services.accessruleController, tags.runOn.serialCI] },
    () => {
        let defaultDepartment: Department;

        let testUsers = new Array<UserPostResponse>(0);
        let roles = new Array<Role>(0);
        const roleNameToUser = new Map<string, UserPostResponse>();
        const roleNameToRole = new Map<string, Role>();

        const randomPrefix = Utils.generateShortUuid().slice(0, 4);

        // enable features and accessTokens
        test.beforeAll(async () => {
            try {
                await operations.forceCleanup();
            } catch (error) {
                logger.error(
                    'Failed force clean before Workload Ownership Protection tests - assuming can continue'
                );
            }

            authorizationApi.accessRulesApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            identityManagerApi.usersApi.accessToken = flowRootState.users.system_administrator.accessToken;
            projectsDepartmentsApi.departmentsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            orgUnitService.departmentApi.accessToken = flowRootState.users.system_administrator.accessToken;

            defaultDepartment = await departmentUtils.getDefaultDepartmentByCluster();

            // create user for each role
            logger.debug('Creating a user for each role');
            roles = await authorizationUtils.getRoles(flowRootState.users.system_administrator.accessToken);

            const promises: Promise<UserPostResponse>[] = [];
            const roleNames = roles.map((role) => {
                roleNameToRole[role.name] = role;

                const roleNameInK8s = authV2Utils.getRoleNameInK8s(role.name);
                const userEmail = objectGenerator.generateEmailByPrefix(`${roleNameInK8s}-${randomPrefix}`);
                promises.push(
                    usersManager
                        .createLocalUser(flowRootState.users.system_administrator.accessToken, userEmail)
                        .then((result) => result.body)
                );
                return role.name;
            });
            testUsers = await Promise.all(promises);

            for (let index = 0; index < roleNames.length; index++) {
                roleNameToUser[roleNames[index]] = testUsers[index];
            }
        });

        test.beforeEach(() => {
            authorizationApi.accessRulesApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            identityManagerApi.usersApi.accessToken = flowRootState.users.system_administrator.accessToken;
            projectsDepartmentsApi.departmentsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            orgUnitService.departmentApi.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        // delete users
        test.afterAll(async () => {
            logger.debug('Deleting users');
            const promises = [];
            testUsers.forEach((user) => {
                promises.push(
                    usersManager.deleteLocalUser(flowRootState.users.system_administrator.accessToken, user.id)
                );
            });
            await Promise.all(promises);

            // cleanup
            try {
                await operations.forceCleanup();
            } catch (error) {
                logger.error(`Failed to delete all projects and departments, error: ${error.toString()}`);
            }
        });

        test.describe('Validate users have accessrules in existing cluster', () => {
            let createdAccessRules = Array<AccessRule>(0);

            let testCreatedDepartment: DepartmentOU;
            const testDepartments = new Array<Department>(0);
            const testCreatedProjects = new Map<string, Project[]>(); // per department id
            const namespacesOfTestCreatedProjects = new Map<string, string[]>(); // per department id

            async function testDep(testDepartment: any): Promise<void> {
                namespacesOfTestCreatedProjects.set(testDepartment.id.toString(), new Array<string>(0));

                const promises = [];
                for (let i = 0; i < 2; i++) {
                    const projectName = objectGenerator.generateNameByPrefix('p');
                    const currentProjNamespace = Utils.namespaceFromProjectName(projectName);
                    namespacesOfTestCreatedProjects
                        .get(testDepartment.id.toString())
                        .push(currentProjNamespace);

                    logger.debug(
                        `Creating project ${projectName} in department ${testDepartment.name}/${testDepartment.id}`
                    );

                    promises.push(
                        projectResourcesUtils.createProject(
                            currentProjNamespace,
                            {
                                name: projectName,
                                maxAllowedGpus: 100,
                                deservedGpus: 10,
                                departmentId: Number(testDepartment.id)
                            },
                            flowRootState.users.editor.accessToken
                        )
                    );
                }
                testCreatedProjects.set(testDepartment.id.toString(), await Promise.all(promises));
            }

            // create department and projects in departments
            test.beforeAll(async () => {
                testCreatedDepartment = (await departmentUtils.createBackendDepartment()).body;
                expect(testCreatedDepartment, 'expected department to be created').toBeDefined();
                testDepartments.push(defaultDepartment);

                for (const testDepartment of testDepartments) {
                    await testDep(testDepartment);
                }
                await testDep(testCreatedDepartment);
            });

            // delete created department, projects and access rules
            test.afterAll(async () => {
                try {
                    const promises = [];
                    for (const projects of Array.from(testCreatedProjects.values())) {
                        for (const proj of projects) {
                            logger.debug(`Deleting created project ${proj.name}`);
                            promises.push(orgUnitUtils.deleteProject(proj.id.toString()));
                        }
                    }
                    await Promise.all(promises);
                } catch {
                    // that's okay..
                }

                try {
                    logger.debug(`Deleting created department ${testCreatedDepartment.name}`);
                    await departmentUtils.deleteDepartment(Number(testCreatedDepartment.id));
                } catch {
                    // that's okay..
                }

                try {
                    logger.debug('Deleting created accessrule');
                    const promises = [];
                    createdAccessRules.forEach((createdAccessRule) => {
                        promises.push(authorizationApi.accessRulesApi.deleteAccessRule(createdAccessRule.id));
                    });
                    await Promise.all(promises);
                } catch {
                    // that's okay..
                }
            });

            test.describe('Validate users have accessrules in tenant scope', () => {
                const rolesToTest = [
                    PROPERTY_TO_ROLE_MAPPING.ml_engineer,
                    PROPERTY_TO_ROLE_MAPPING.department_admin
                ];

                // create access rules in tenant
                test.beforeAll(async () => {
                    logger.debug(
                        `Creating accessrules for users for tenant ${flowRootState.cluster.tenantId}`
                    );

                    const userToRoleMap = new Map<string, number>();
                    rolesToTest.forEach((roleToTest) =>
                        userToRoleMap.set(roleNameToUser[roleToTest].username, roleNameToRole[roleToTest].id)
                    );

                    createdAccessRules = await authorizationUtils.createAccessRulesForUsers(
                        userToRoleMap as Map<string, number>,
                        String(flowRootState.cluster.tenantId),
                        ScopeType.Tenant,
                        SubjectType.User
                    );

                    expect(createdAccessRules.length, 'expected to create accessrules for tenant').toBe(
                        rolesToTest.length
                    );

                    // wait for accessrules to be created in the cluster
                    await accessRulesClusterUtils.waitForAccessRulesCrdsById(
                        createdAccessRules.map((accessRule) => String(accessRule.id))
                    );
                });

                // delete access rules - just in case they weren't already deleted
                test.afterAll(async () => {
                    try {
                        logger.debug('Deleting created accessrule');
                        const promises = [];
                        createdAccessRules.forEach((createdAccessRule) => {
                            promises.push(
                                authorizationApi.accessRulesApi.deleteAccessRule(createdAccessRule.id)
                            );
                        });
                        await Promise.all(promises);
                    } catch (e) {
                        logger.warn(`failed to delete accessrules, error: ${e.toString()}`);
                    }

                    createdAccessRules = [];
                });

                test('1: Validate accessrule + rolebinding + clusterrolebinding match', async () => {
                    logger.debug(
                        'Starting test: 1: Validate accessrule + rolebinding + clusterrolebinding match'
                    );
                    // first, wait for all accessRules to be created in cluster - by spec
                    const promises = [];
                    for (const accessRule of createdAccessRules) {
                        promises.push(
                            accessRulesClusterUtils.waitForAccessRuleCrdBySpec(
                                accessRule.roleName,
                                accessRule.subjectId,
                                String(accessRule.subjectType),
                                String(ScopeType.Cluster)
                            )
                        );
                    }
                    await Promise.all(promises);

                    const allProjectsNamespaces = [].concat(
                        ...Array.from(namespacesOfTestCreatedProjects.values())
                    );
                    await accessRulesClusterUtils.validateAccessRulesRoleBindingInCluster(
                        createdAccessRules,
                        allProjectsNamespaces
                    );
                });
            });

            test.describe('Validate users have accessrules in cluster scope', () => {
                const rolesToTest = [
                    PROPERTY_TO_ROLE_MAPPING.l2_researcher,
                    PROPERTY_TO_ROLE_MAPPING.editor
                ];

                // create access rules in cluster
                test.beforeAll(async () => {
                    logger.debug(`Creating accessrules for users for cluster ${flowRootState.cluster.uuid}`);

                    const userToRoleMap = new Map<string, number>();
                    rolesToTest.forEach((roleToTest) =>
                        userToRoleMap.set(roleNameToUser[roleToTest].username, roleNameToRole[roleToTest].id)
                    );

                    createdAccessRules = await authorizationUtils.createAccessRulesForUsers(
                        userToRoleMap as Map<string, number>,
                        flowRootState.cluster.uuid,
                        ScopeType.Cluster,
                        SubjectType.User
                    );

                    expect(createdAccessRules.length, 'expected to create accessrules for cluster').toBe(
                        rolesToTest.length
                    );

                    // wait for accessrules to be created in the cluster
                    await accessRulesClusterUtils.waitForAccessRulesCrdsById(
                        createdAccessRules.map((accessRule) => String(accessRule.id))
                    );
                });

                // delete access rules - just in case they weren't already deleted
                test.afterAll(async () => {
                    try {
                        logger.debug('Deleting created accessrule');
                        const promises = [];
                        createdAccessRules.forEach((createdAccessRule) => {
                            promises.push(
                                authorizationApi.accessRulesApi.deleteAccessRule(createdAccessRule.id)
                            );
                        });
                        await Promise.all(promises);
                    } catch (e) {
                        logger.warn(`failed to delete accessrules, error: ${e.toString()}`);
                    }

                    createdAccessRules = [];
                });

                test('1: Validate accessrule + rolebinding + clusterrolebinding match', async () => {
                    logger.debug(
                        'Starting test: 1: Validate accessrule + rolebinding + clusterrolebinding match'
                    );
                    // first, wait for all accessRules to be created in cluster - by spec
                    const promises = [];
                    for (const accessRule of createdAccessRules) {
                        promises.push(
                            accessRulesClusterUtils.waitForAccessRuleCrdBySpec(
                                accessRule.roleName,
                                accessRule.subjectId,
                                String(accessRule.subjectType),
                                String(accessRule.scopeType),
                                accessRule.scopeName,
                                60
                            )
                        );
                    }
                    await Promise.all(promises);

                    const allProjectsNamespaces = [].concat(
                        ...Array.from(namespacesOfTestCreatedProjects.values())
                    );
                    await accessRulesClusterUtils.validateAccessRulesRoleBindingInCluster(
                        createdAccessRules,
                        allProjectsNamespaces
                    );
                });
            });

            test.describe('Validate users have accessrules in department scope', () => {
                const rolesToTest = [
                    PROPERTY_TO_ROLE_MAPPING.research_manager,
                    PROPERTY_TO_ROLE_MAPPING.compute_resource_administrator
                ];

                // create access rules in department
                test.beforeAll(async () => {
                    logger.debug(
                        `Creating accessrules for users for department ${testCreatedDepartment.name}`
                    );

                    const userToRoleMap = new Map<string, number>();
                    rolesToTest.forEach((roleToTest) =>
                        userToRoleMap.set(roleNameToUser[roleToTest].username, roleNameToRole[roleToTest].id)
                    );

                    createdAccessRules = await authorizationUtils.createAccessRulesForUsers(
                        userToRoleMap as Map<string, number>,
                        String(testCreatedDepartment.id),
                        ScopeType.Department,
                        SubjectType.User
                    );

                    expect(createdAccessRules.length, 'expected to create accessrules for department').toBe(
                        rolesToTest.length
                    );

                    // wait for accessrules to be created in the cluster
                    await accessRulesClusterUtils.waitForAccessRulesCrdsById(
                        createdAccessRules.map((accessRule) => String(accessRule.id))
                    );
                });

                // delete access rules - just in case they weren't already deleted
                test.afterAll(async () => {
                    try {
                        logger.debug('Deleting created accessrule');
                        const promises = [];
                        createdAccessRules.forEach((createdAccessRule) => {
                            promises.push(
                                authorizationApi.accessRulesApi.deleteAccessRule(createdAccessRule.id)
                            );
                        });
                        await Promise.all(promises);
                    } catch {
                        // that's okay..
                    }

                    createdAccessRules = [];
                });

                test('1: Validate accessrule + rolebinding + clusterrolebinding match', async () => {
                    logger.debug(
                        'Starting test: 1: Validate accessrule + rolebinding + clusterrolebinding match'
                    );
                    // first, wait for all accessRules to be created in cluster - by spec
                    const promises = [];
                    for (const accessRule of createdAccessRules) {
                        promises.push(
                            accessRulesClusterUtils.waitForAccessRuleCrdBySpec(
                                accessRule.roleName,
                                accessRule.subjectId,
                                String(accessRule.subjectType),
                                String(accessRule.scopeType),
                                accessRule.scopeName,
                                60
                            )
                        );
                    }
                    await Promise.all(promises);

                    await accessRulesClusterUtils.validateAccessRulesRoleBindingInCluster(
                        createdAccessRules || [],
                        namespacesOfTestCreatedProjects.get(testCreatedDepartment.id.toString())
                    );
                });

                test('3: Validate accessrules deleted after department deleted', async () => {
                    logger.debug('Starting test: 3: Validate accessrules deleted after department deleted');
                    await departmentUtils.deleteDepartmentAndAllItsProjects(Number(testCreatedDepartment.id));
                    logger.info(
                        `Waiting for accessrules of deleted department ${testCreatedDepartment.name}/${testCreatedDepartment.id} to be deleted`
                    );
                    await departmentUtils.waitForDepartmentCrdToBeDeleted(testCreatedDepartment.name);

                    const promises = [];
                    for (const accessRule of createdAccessRules) {
                        promises.push(
                            accessRulesClusterUtils.waitForAccessRuleToBeDeleted(
                                accessRule.id,
                                accessRule.roleName,
                                accessRule.subjectId,
                                String(accessRule.subjectType),
                                String(accessRule.scopeType),
                                accessRule.scopeName
                            )
                        );
                    }
                    await Promise.all(promises);
                });
            });

            test.describe('Validate users have accessrules in project scope', () => {
                const rolesToTest = [
                    PROPERTY_TO_ROLE_MAPPING.l1_researcher,
                    PROPERTY_TO_ROLE_MAPPING.data_source_administrator
                ];
                let testCreatedProject;
                let namespaceOfTestCreatedProject;

                // create access rules in project
                test.beforeAll(async () => {
                    testCreatedProject = testCreatedProjects.get(defaultDepartment.id.toString())[0];
                    namespaceOfTestCreatedProject = namespacesOfTestCreatedProjects.get(
                        defaultDepartment.id.toString()
                    )[0];

                    logger.debug(`Creating accessrules for users for project ${testCreatedProject.name}`);

                    const userToRoleMap = new Map<string, number>();
                    rolesToTest.forEach((roleToTest) =>
                        userToRoleMap.set(roleNameToUser[roleToTest].username, roleNameToRole[roleToTest].id)
                    );

                    createdAccessRules = await authorizationUtils.createAccessRulesForUsers(
                        userToRoleMap,
                        String(testCreatedProject.id),
                        ScopeType.Project,
                        SubjectType.User
                    );

                    expect(createdAccessRules.length, 'expected to create accessrules for project').toBe(
                        rolesToTest.length
                    );

                    // wait for accessrules to be created in the cluster
                    await accessRulesClusterUtils.waitForAccessRulesCrdsById(
                        createdAccessRules.map((accessRule) => String(accessRule.id))
                    );
                });

                // delete access rules - just in case they weren't already deleted
                test.afterAll(async () => {
                    try {
                        logger.debug('Deleting created accessrule');
                        const promises = [];
                        createdAccessRules.forEach((createdAccessRule) => {
                            promises.push(
                                authorizationApi.accessRulesApi.deleteAccessRule(createdAccessRule.id)
                            );
                        });
                        await Promise.all(promises);
                    } catch {
                        // that's okay..
                    }

                    createdAccessRules = [];
                });

                test('1: Validate accessrule + rolebinding + clusterrolebinding match', async () => {
                    logger.debug(
                        'Starting test: 1: Validate accessrule + rolebinding + clusterrolebinding match'
                    );
                    // first, wait for all accessRules to be created in cluster - by spec
                    const promises = [];
                    for (const accessRule of createdAccessRules) {
                        promises.push(
                            accessRulesClusterUtils.waitForAccessRuleCrdBySpec(
                                accessRule.roleName,
                                accessRule.subjectId,
                                String(accessRule.subjectType),
                                String(accessRule.scopeType),
                                accessRule.scopeName,
                                60
                            )
                        );
                    }
                    await Promise.all(promises);

                    await accessRulesClusterUtils.validateAccessRulesRoleBindingInCluster(
                        createdAccessRules,
                        [namespaceOfTestCreatedProject]
                    );
                });

                test('2: Validate accessrules deleted after project deleted', async () => {
                    logger.debug('Starting test: 2: Validate accessrules deleted after project deleted');
                    projectsDepartmentsApi.projectsApi.accessToken =
                        flowRootState.users.editor.accessToken;
                    await orgUnitUtils.deleteProject(testCreatedProject.id);

                    await projectUtils.waitForProjectCrdToBeDeleted(testCreatedProject.name);

                    logger.info(
                        `Waiting for accessrules of deleted project ${testCreatedProject.name}/${testCreatedProject.id} to be deleted`
                    );
                    const promises = [];
                    for (const accessRule of createdAccessRules) {
                        promises.push(
                            accessRulesClusterUtils.waitForAccessRuleToBeDeleted(
                                accessRule.id,
                                accessRule.roleName,
                                accessRule.subjectId,
                                String(accessRule.subjectType),
                                String(accessRule.scopeType),
                                accessRule.scopeName
                            )
                        );
                    }
                    await Promise.all(promises);
                });
            });
        });

        test.describe('Validate users have accessrules in new cluster', function () {
            const rolesToTest = [
                PROPERTY_TO_ROLE_MAPPING.template_administrator,
                PROPERTY_TO_ROLE_MAPPING.department_viewer
            ];
            let testCluster: Cluster;
            let createdAccessRules = Array<AccessRule>(0);

            // create cluster
            test.beforeAll(async function () {
                testCluster = await upgradeClusterUtils.createClusterByName(
                    objectGenerator.generateNameByPrefix('cl')
                );

                // create access rules for cluster
                logger.debug(`Creating accessrules for users for cluster ${testCluster.uuid}`);

                const userToRoleMap = new Map<string, number>();
                rolesToTest.forEach((roleToTest) =>
                    userToRoleMap.set(roleNameToUser[roleToTest].username, roleNameToRole[roleToTest].id)
                );

                createdAccessRules = await authorizationUtils.createAccessRulesForUsers(
                    userToRoleMap as Map<string, number>,
                    testCluster.uuid,
                    ScopeType.Cluster,
                    SubjectType.User
                );

                expect(createdAccessRules.length, 'expected to create accessrules for cluster').toBe(
                    rolesToTest.length
                );

                // validate accessrules created
                const promises = [];
                for (const accessRule of createdAccessRules) {
                    promises.push(authorizationUtils.waitForAccessRule(accessRule.id, true));
                }
                await Promise.all(promises);
            });

            // delete access rules - just in case they weren't already deleted
            test.afterAll(async function () {
                try {
                    logger.debug('Deleting created accessrule');
                    const promises = [];
                    createdAccessRules.forEach((createdAccessRule) => {
                        promises.push(authorizationApi.accessRulesApi.deleteAccessRule(createdAccessRule.id));
                    });
                    await Promise.all(promises);
                } catch {
                    // that's okay..
                }

                createdAccessRules = [];

                // delete the cluster object
                try {
                    await backendClusterApi.clusterApi.v1K8sClustersUuidDelete(testCluster.uuid);
                } catch (error) {
                    logger.error(`Failed to delete created cluster object, error: ${error.toString()}`);
                }
            });

            test('1: Validate accessrules deleted after cluster deleted', async function () {
                logger.debug('Starting test: 1: Validate accessrules deleted after cluster deleted');
                await backendClusterApi.clusterApi.v1K8sClustersUuidDelete(testCluster.uuid);

                logger.info(
                    `Waiting for accessrules of deleted cluster ${testCluster.name}/${testCluster.uuid} to be deleted`
                );
                const promises = [];
                for (const accessRule of createdAccessRules) {
                    promises.push(authorizationUtils.waitForAccessRule(accessRule.id, false));
                }
                await Promise.all(promises);
            });
        });

        test.describe('Validate accessrules deleted when user deleted', () => {
            let testCreatedUser: UserPostResponse;
            let createdAccessRules = Array<AccessRule>(0);

            let project: Project;
            let namespaceName: string;

            // create project
            test.beforeAll(async () => {
                const projectName = objectGenerator.generateNameByPrefix('p');
                namespaceName = Utils.namespaceFromProjectName(projectName);

                logger.debug(`Creating project ${projectName}`);

                project = await projectResourcesUtils.createProject(
                    namespaceName,
                    {
                        name: projectName,
                        maxAllowedGpus: 100,
                        deservedGpus: 10,
                        departmentId: defaultDepartment.id
                    },
                    flowRootState.users.editor.accessToken
                );

                // create user and accessrules in different scopes
                const userEmail = objectGenerator.generateEmailByPrefix(`user-${randomPrefix}`);
                testCreatedUser = (
                    await usersManager.createLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        userEmail
                    )
                ).body;

                const scopes = [
                    { type: ScopeType.Project, name: project.name, id: project.id },
                    { type: ScopeType.Department, name: defaultDepartment.name, id: defaultDepartment.id },
                    {
                        type: ScopeType.Cluster,
                        name: flowRootState.cluster.name,
                        id: flowRootState.cluster.uuid
                    },
                    { type: ScopeType.Tenant, name: namespaceCluster, id: flowRootState.cluster.tenantId }
                ];

                const promises = [];
                for (const scope of scopes) {
                    logger.debug(
                        `Creating accessrule for user ${testCreatedUser.username} in scope ${scope.type}/${scope.name}`
                    );
                    promises.push(
                        authorizationApi.accessRulesApi.createAccessRule({
                            roleId: roleNameToRole[PROPERTY_TO_ROLE_MAPPING.environment_administrator].id,
                            scopeId: String(scope.id),
                            scopeType: scope.type as ScopeType,
                            subjectId: testCreatedUser.username,
                            subjectType: SubjectType.User
                        })
                    );
                }
                const accessRulesResponses = await Promise.all(promises);
                createdAccessRules = accessRulesResponses.map(
                    (accessRulesResponse) => accessRulesResponse.body
                );

                // wait for accessrules to be created in the cluster
                await accessRulesClusterUtils.waitForAccessRulesCrdsById(
                    createdAccessRules.map((accessRule) => String(accessRule.id))
                );
            });

            // delete access rules - just in case they weren't already deleted
            test.afterAll(async () => {
                try {
                    logger.debug('Deleting created accessrule');
                    const promises = [];
                    createdAccessRules.forEach((createdAccessRule) => {
                        promises.push(authorizationApi.accessRulesApi.deleteAccessRule(createdAccessRule.id));
                    });
                    await Promise.all(promises);
                } catch {
                    // that's okay..
                }

                createdAccessRules = [];

                try {
                    await usersManager.deleteLocalUser(
                        flowRootState.users.system_administrator.accessToken,
                        testCreatedUser.id
                    );
                } catch {
                    // that's okay..
                }
            });

            test('1: delete user and validate accessrules deleted', async () => {
                logger.debug('Starting test: 1: delete user and validate accessrules deleted');
                const response = await usersManager.deleteLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    testCreatedUser.id
                );
                expect(response.response.statusCode).toBe(StatusCodes.NO_CONTENT);

                logger.info(
                    `Waiting for accessrules of deleted user ${testCreatedUser.username} to be deleted`
                );
                const promises = [];
                for (const accessRule of createdAccessRules) {
                    promises.push(
                        accessRulesClusterUtils.waitForAccessRuleToBeDeleted(
                            accessRule.id,
                            accessRule.roleName,
                            accessRule.subjectId,
                            String(accessRule.subjectType),
                            String(accessRule.scopeType),
                            accessRule.scopeName
                        )
                    );
                }
                await Promise.all(promises);
            });
        });
    }
);
