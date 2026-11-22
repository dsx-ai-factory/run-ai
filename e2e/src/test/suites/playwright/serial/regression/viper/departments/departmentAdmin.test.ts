/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { orgUnitService, projectsDepartmentsApi } from '../../../../../../../modules/api';
import {
    authUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    orgUnitUtils,
    projectResourcesUtils,
    settings,
    usersManager,
    Utils
} from '../../../../../../../modules/helpers';
import { departmentUtils } from '../../../../../../../modules/helpers';
import { tags } from '../../../../../../../modules/helpers';
import { Department as DepartmentOU } from '../../../../../../../modules/api/apis/orgUnitApi';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { expect } from '@playwright/test';
import { test } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';
import { UserPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/userPostResponse';

test.describe(`Department Admin, tags:  ${tags.resource.department} ${tags.tool.api} ${tags.teams.viper}`, () => {
    test.describe('departments enabled and resources created', () => {
        const numberOfVisibleDepartments = 1;
        const numberOfVisibleProjects = 1;
        let departmentWithAdmin: DepartmentOU = null;
        let departmentWithoutAdmin: DepartmentOU = null;
        let projectWithAdmin: Project = null;
        let projectWithoutAdmin: Project = null;
        let extraProjectCreated: Project = null;
        let departmentAdminUser: UserPostResponse = null;
        let departmentAdminToken: string = null;

        test.beforeAll(async () => {
            try {
                await operations.forceCleanup();
            } catch (error) {
                logger.error('Failed force clean before Department Admin tests - assuming can continue');
            }

            const userEmail = objectGenerator.generateEmailByPrefix('department-admin-for-default');
            departmentAdminUser = (
                await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    userEmail,
                    false
                )
            ).body;

            projectsDepartmentsApi.departmentsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            orgUnitService.departmentApi.accessToken = flowRootState.users.system_administrator.accessToken;
            departmentWithAdmin = await orgUnitUtils.getDefaultDepartment();
            departmentWithoutAdmin = (await departmentUtils.createBackendDepartment()).body;
            const res = await projectsDepartmentsApi.departmentsApi.updateDepartmentAdmins(
                departmentWithAdmin.clusterId,
                Number(departmentWithAdmin.id),
                { departmentAdmins: [departmentAdminUser.username] }
            );
            departmentUtils.validateDepartmentAccessRulesResponse(res, Number(departmentWithAdmin.id), [
                departmentAdminUser.username
            ]);

            departmentAdminToken = (
                await authUtils.authenticateUser(
                    departmentAdminUser.username,
                    departmentAdminUser.tempPassword
                )
            ).idToken;

            const projectName = objectGenerator.generateNameByPrefix('project-without-admin');
            const namespace = Utils.namespaceFromProjectName(projectName);
            projectWithoutAdmin = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 2,
                maxAllowedGpus: 2,
                departmentId: Number(departmentWithoutAdmin.id)
            });
        });

        test.afterAll(async () => {
            try {
                if (projectWithoutAdmin) {
                    await operations.deleteProjectAndNamespace(projectWithoutAdmin.id);
                }
                projectWithoutAdmin = null;
                if (departmentWithoutAdmin) {
                    await departmentUtils.deleteDepartmentCleanup(Number(departmentWithoutAdmin.id));
                }
                departmentWithoutAdmin = null;
            } catch (error) {
                logger.warn('Failed to cleanup', error);
            }
            await authUtils.deleteUser(
                departmentAdminUser.id,
                flowRootState.users.system_administrator.accessToken
            );
        });

        test.describe('Running with department admin context', () => {
            test.beforeAll(async () => {
                projectsDepartmentsApi.departmentsApi.accessToken = departmentAdminToken;
                orgUnitService.departmentApi.accessToken = departmentAdminToken;
                projectsDepartmentsApi.projectsApi.accessToken = departmentAdminToken;

                const projectName = objectGenerator.generateNameByPrefix('project-with-admin');
                const namespace = Utils.namespaceFromProjectName(projectName);
                projectWithAdmin = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2,
                    departmentId: Number(departmentWithAdmin.id)
                });
            });

            test.beforeEach(() => {
                projectsDepartmentsApi.departmentsApi.accessToken = departmentAdminToken;
                orgUnitService.departmentApi.accessToken = departmentAdminToken;
                projectsDepartmentsApi.projectsApi.accessToken = departmentAdminToken;
            });

            test.afterAll(async () => {
                projectsDepartmentsApi.departmentsApi.accessToken =
                    flowRootState.users.editor.accessToken;
                orgUnitService.departmentApi.accessToken = flowRootState.users.editor.accessToken;
                projectsDepartmentsApi.projectsApi.accessToken =
                    flowRootState.users.editor.accessToken;

                if (projectWithAdmin) {
                    await operations.deleteProjectAndNamespace(projectWithAdmin.id);
                }
                projectWithAdmin = null;
                if (extraProjectCreated) {
                    await operations.deleteProjectAndNamespace(extraProjectCreated.id);
                }
                extraProjectCreated = null;
            });

            test('department admin can List only the projects and department under his management - enable auth only', async () => {
                if (settings.enableAuth) {
                    const visibleDepartments = (
                        await projectsDepartmentsApi.departmentsApi.getDepartments(
                            departmentWithAdmin.clusterId
                        )
                    ).body;
                    expect(visibleDepartments.length).toBe(numberOfVisibleDepartments);
                    expect(visibleDepartments[0].id.toString()).toBe(departmentWithAdmin.id);
                    const response = await orgUnitUtils.getDepartment(departmentWithAdmin.id);
                    expect(response?.id).not.toBe(null);

                    const visibleProjects = (
                        await projectsDepartmentsApi.projectsApi.getProjects(departmentWithAdmin.clusterId)
                    ).body;
                    expect(visibleProjects.length).toBe(numberOfVisibleProjects);
                    expect(visibleProjects[0].id).toBe(projectWithAdmin.id);
                    const project = (
                        await projectsDepartmentsApi.projectsApi.getProjects(flowRootState.cluster.uuid)
                    ).body.find((p) => p.id === projectWithAdmin.id);
                    expect(project?.id).toEqual(projectWithAdmin.id);
                } else {
                    test.skip();
                }
            });

            test('department admin cant Get department not under his management - enable auth only', async () => {
                if (settings.enableAuth) {
                    try {
                        await orgUnitService.departmentApi.getDepartment(departmentWithoutAdmin.id);
                        throw new Error('Expected to fail, but succeeded');
                    } catch (e) {
                        expect(e.statusCode).toBe(StatusCodes.FORBIDDEN);
                    }
                } else {
                    test.skip();
                }
            });

            test('department admin can Create and Delete projects in his managed departments', async () => {
                const projectName = objectGenerator.generateNameByPrefix('project');
                const namespace = Utils.namespaceFromProjectName(projectName);
                extraProjectCreated = await projectResourcesUtils.createProject(
                    namespace,
                    {
                        name: projectName,
                        deservedGpus: 2,
                        maxAllowedGpus: 2,
                        departmentId: Number(departmentWithAdmin.id)
                    },
                    departmentAdminToken
                );
                expect(extraProjectCreated.departmentId.toString()).toBe(departmentWithAdmin.id);

                await operations.deleteProjectAndNamespace(extraProjectCreated.id, departmentAdminToken);
                extraProjectCreated = null;
            });

            test('department admin cant Create projects not in his managed departments - enable auth only', async () => {
                if (settings.enableAuth) {
                    const projectNameWillFail = objectGenerator.generateNameByPrefix('project');
                    const namespaceWillFail = Utils.namespaceFromProjectName(projectNameWillFail);
                    try {
                        await projectResourcesUtils.createProject(
                            namespaceWillFail,
                            {
                                name: projectNameWillFail,
                                deservedGpus: 2,
                                maxAllowedGpus: 2,
                                departmentId: Number(departmentWithoutAdmin.id)
                            },
                            departmentAdminToken
                        );
                        throw new Error('Expected to fail, but succeeded');
                    } catch (e) {
                        expect(e.statusCode).toBe(StatusCodes.NOT_FOUND);
                    }
                } else {
                    test.skip();
                }
            });
        });
    });
});
