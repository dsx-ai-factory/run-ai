/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    authorizationUtils,
    authUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    usersManager,
    versionUtils
} from '../../../../../../../modules/helpers';
import { SupportedVersions } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { UserPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { identityManagerApi } from '../../../../../../../modules/api';
import { test } from '@playwright/test';
import {
    Department,
    DepartmentCreationRequest,
    Project,
    ProjectCreationRequest
} from '../../../../../../../modules/api/apis/orgUnitApi';
import { orgUnitService } from '../../../../../../../modules/api';
import { StatusCodes } from 'http-status-codes';
import { SubjectType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { ScopeType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';

test.describe(`Authorization API Permissions, tags: ${tags.tool.api} ${tags.teams.viper}`, function () {
    let userName: string;
    let password: string;
    let user: UserPostResponse;
    let targetSystemAdminUser: UserPostResponse;
    let department1: Department;
    let department2: Department;
    let project: Project;

    test.beforeAll(async function () {
        logger.info('create departments');
        const departmentName = objectGenerator.generateNameByPrefix('d');
        department1 = await orgUnitUtils.createDepartment({
            name: departmentName,
            clusterId: flowRootState.cluster.uuid,
            resources: [
                {
                    gpu: { deserved: 10, limit: -1, overQuotaWeight: 10 }
                }
            ]
        });
        const department2Name = objectGenerator.generateNameByPrefix('d');
        department2 = await orgUnitUtils.createDepartment({
            name: department2Name,
            clusterId: flowRootState.cluster.uuid,
            resources: [
                {
                    gpu: { deserved: 10, limit: -1, overQuotaWeight: 10 }
                }
            ]
        });

        logger.info('create project');
        const projectName = objectGenerator.generateNameByPrefix('p');
        project = await orgUnitUtils.createProject({
            name: projectName,
            clusterId: flowRootState.cluster.uuid,
            parentId: department1.id,
            resources: [
                {
                    gpu: { deserved: 2, limit: -1, overQuotaWeight: 2 }
                }
            ]
        });

        logger.info('create user');
        user = (await usersManager.createLocalUser(
            flowRootState.users.system_administrator.accessToken,
            objectGenerator.generateEmailByPrefix('permissiontest'),
            false
        )).body;
        userName = user.username;
        password = user.tempPassword;

        logger.info('create access rule with role editor and scope project');
        await authorizationUtils.createAccessRuleForUser(
            flowRootState.users.system_administrator.accessToken,
            userName,
            flowRootState.users.editor.role.id,
            SubjectType.User,
            project.id,
            ScopeType.Project
        );

        logger.info('create access rule with role l2 researcher and scope department2');
        await authorizationUtils.createAccessRuleForUser(
            flowRootState.users.system_administrator.accessToken,
            userName,
            flowRootState.users.l2_researcher.role.id,
            SubjectType.User,
            flowRootState.cluster.uuid,
            ScopeType.Cluster
        );

        logger.info('create target system admin user for privilege escalation test');
        targetSystemAdminUser = (await usersManager.createLocalUser(
            flowRootState.users.system_administrator.accessToken,
            objectGenerator.generateEmailByPrefix('target-sysadmin'),
            false
        )).body;
        await authorizationUtils.createAccessRuleForUser(
            flowRootState.users.system_administrator.accessToken,
            targetSystemAdminUser.username,
            flowRootState.users.system_administrator.role.id,
            SubjectType.User
        );
    });

    test.afterAll(async function () {
        try {
            await orgUnitUtils.deleteProject(project.id);
        } catch (e) {
            logger.warn(`failed to delete project ${project.id}`);
        }

        try {
            await orgUnitUtils.deleteDepartment(department1.id);
        } catch (e) {
            logger.warn(`failed to delete department ${department1.id}`);
        }
        try {
            await orgUnitUtils.deleteDepartment(department2.id);
        } catch (e) {
            logger.warn(`failed to delete department ${department2.id}`);
        }

        await usersManager.deleteLocalUser(flowRootState.users.system_administrator.accessToken, user.id);

        try {
            await usersManager.deleteLocalUser(
                flowRootState.users.system_administrator.accessToken,
                targetSystemAdminUser.id
            );
        } catch (e) {
            logger.warn(`failed to delete target system admin user ${targetSystemAdminUser.id}`);
        }
    });

    test('department admin should not be able to delete a system admin user', async function () {
        identityManagerApi.usersApi.accessToken = flowRootState.users.department_admin.accessToken;
        try {
            await identityManagerApi.usersApi.deleteUserById(targetSystemAdminUser.id);
            expect.fail('delete should have been rejected with 403');
        } catch (e) {
            expect(e.statusCode).to.equal(StatusCodes.FORBIDDEN);
            // The dedicated message was added in 3.0; older control planes reject with the
            // generic identity-manager delete error.
            const expectedMessage = versionUtils.isVersionSupported(
                flowRootState.controlPlaneVersion,
                SupportedVersions.V3_0
            )
                ? 'You\'re not authorized to delete the selected user'
                : `failed deleting user by id: ${targetSystemAdminUser.id}`;
            expect(e.body?.message).to.equal(expectedMessage);
        }
    });

    test('should not be able to create department on cluster with l2 researcher role', async function () {
        logger.info('create department on cluster should fail');
        const token = await authUtils.authenticateTokenApi(userName, password);
        orgUnitService.departmentApi.accessToken = token.body.accessToken;
        const depReq = {
            name: objectGenerator.generateNameByPrefix('d'),
            clusterId: flowRootState.cluster.uuid,
            resources: [
                {
                    gpu: { deserved: 10, limit: -1, overQuotaWeight: 10 }
                }
            ]
        } as DepartmentCreationRequest;
        try {
            await orgUnitService.departmentApi.createDepartment(depReq);
            expect.fail('department creation should fail');
        } catch (e) {
            expect(e.statusCode).to.equal(StatusCodes.NOT_FOUND);
        }
    });

    test('should not be able to create project on department with l2 researcher role', async function () {
        logger.info('create project on department2 should fail');
        const token = await authUtils.authenticateTokenApi(userName, password);
        orgUnitService.projectApi.accessToken = token.body.accessToken;
        const projReq = {
            name: objectGenerator.generateNameByPrefix('p'),
            clusterId: flowRootState.cluster.uuid,
            parentId: department2.id,
            resources: [
                {
                    gpu: { deserved: 2, limit: -1, overQuotaWeight: 2 }
                }
            ]
        } as ProjectCreationRequest;
        try {
            await orgUnitService.projectApi.createProject(projReq);
            expect.fail('project creation should fail');
        } catch (e) {
            expect(e.statusCode).to.equal(StatusCodes.NOT_FOUND);
        }
    });
});
