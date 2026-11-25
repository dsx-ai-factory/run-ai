/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    datavolumeUtils,
    objectGenerator,
    logger,
    projectResourcesUtils,
    operations,
    usersManager,
    flowRootState,
    authorizationUtils,
    authUtils,
    TokenObject,
    departmentUtils,
    tags
} from '../../../../../../../modules/helpers';
import { ScopeType } from '../../../../../../../modules/api/apis/datavolumesApi/generated/model/scopeType';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../../modules/helpers/config/roleMapping';
import { UserPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/api';
import { AccessRule } from '../../../../../../../modules/api/apis/authorizationApi/generated/api';
import { Datavolume } from '../../../../../../../modules/api/apis/datavolumesApi/generated/api';

test.describe(`Negative datavolume tests ${tags.teams.viper}`, { tag: [tags.services.datavolume, tags.runOn.serialCI] }, () => {
    let proj1: Project, proj2: Project;
    let testDatavolume: Datavolume;
    const createdDatavolumes = [];
    let testUser: UserPostResponse, testUserToken: TokenObject, testAccessRule: AccessRule;

    test.beforeAll(async function () {
        proj1 = await projectResourcesUtils.createProjectAndNamespace('project-1');
        proj2 = await projectResourcesUtils.createProjectAndNamespace('project-2');

        const userEmail = objectGenerator.generateEmailByPrefix('datamanager-proj2');
        testUser = (
            await usersManager.createLocalUser(
                flowRootState.users.system_administrator.accessToken,
                userEmail,
                false
            )
        ).body;
        testAccessRule = await authorizationUtils.createAccessRuleForUserByRoleName(
            userEmail,
            PROPERTY_TO_ROLE_MAPPING.data_volume_administrator,
            ScopeType.Project,
            String(proj1.id)
        );
        testUserToken = await authUtils.authenticateUser(userEmail, testUser.tempPassword);
    });

    test.afterAll(async function () {
        // Clean up created datavolumes
        for (const dv of createdDatavolumes) {
            await datavolumeUtils.deleteDatavolume(dv.id);
        }

        // Clean up access rules and user
        if (testAccessRule) await authorizationUtils.deleteAccessRule(testAccessRule);
        if (testUser)
            await usersManager.deleteLocalUser(flowRootState.users.system_administrator.accessToken, testUser.id);

        // Clean up projects
        await operations.deleteProjectAndNamespace(proj1.id);
        await operations.deleteProjectAndNamespace(proj2.id);
    });

    test('should fail to create a datavolume with non-existing PVC', async function () {
        const dvName = objectGenerator.generateNameByPrefix('dv');
        const pvcName = 'unexisting-pvc';

        logger.info(`Creating the datavolume '${dvName}' with un-existing PVC ${pvcName}`);
        testDatavolume = await datavolumeUtils.createDatavolume({
            name: dvName,
            description: 'datavolume description',
            originPvcName: pvcName,
            projectId: String(proj1.id)
        });

        createdDatavolumes.push(testDatavolume);

        await datavolumeUtils.waitForCrdDatavolumeToBeNotReady(testDatavolume.name);
        await datavolumeUtils.waitForApiDatavolumeToBeNotReady(testDatavolume.id);
    });

    test('should fail to share with unauthorized project', async function () {
        const unauthorizedProjectId = String(proj2.id);
        logger.info(`Attempting to share with unauthorized project ${unauthorizedProjectId}`);

        await expect(async () => {
            await datavolumeUtils.patchDatavolumeSharedScopes(
                testDatavolume.id,
                {
                    add: [{ scopeId: unauthorizedProjectId, scopeType: ScopeType.Project }]
                },
                testUserToken.idToken
            );
        }).rejects.toThrow(/403/);
    });

    test('should fail to share with unauthorized department', async function () {
        await expect(async () => {
            const defaultDepartmentId = (
                await departmentUtils.getDefaultDepartmentByCluster(flowRootState.cluster.uuid)
            ).id;

            logger.info(
                `Sharing datavolume ${testDatavolume.id} with department ${defaultDepartmentId} - but without permissions, should fail`
            );

            await datavolumeUtils.patchDatavolumeSharedScopes(
                testDatavolume.id,
                {
                    add: [{ scopeId: String(defaultDepartmentId), scopeType: ScopeType.Department }]
                },
                testUserToken.idToken
            );
        }).rejects.toThrow(/403/);
    });

    test('should fail to share with unauthorized cluster', async function () {
        await expect(async () => {
            logger.info(
                `Sharing datavolume ${testDatavolume.id} with cluster ${flowRootState.cluster.uuid} - but without permissions, should fail`
            );

            await datavolumeUtils.patchDatavolumeSharedScopes(
                testDatavolume.id,
                {
                    add: [{ scopeId: flowRootState.cluster.uuid, scopeType: ScopeType.Cluster }]
                },
                testUserToken.idToken
            );
        }).rejects.toThrow(/403/);
    });
});
