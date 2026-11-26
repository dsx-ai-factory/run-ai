/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authorizationUtils,
    clusterUtils,
    departmentUtils,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    operations,
    orgUnitUtils,
    projectResourcesUtils,
    settings,
    tags,
    Utils
} from '../../../../../../../modules/helpers';
import { expect } from '@playwright/test';
import { test } from '@playwright/test';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { Department } from '../../../../../../../modules/api/apis/orgUnitApi';
import { backendClusterApi, orgUnitService } from '../../../../../../../modules/api';
import { StatusCodes } from 'http-status-codes';
import { AccessRule } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/accessRule';
import { SubjectType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { OrgUnit } from '../../../../../../../modules/api/apis/backendApi';
import { ScopeType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { validateTree } from '../../../../../../../modules/helpers/orgTreeUtils';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';
import { TestInfo } from 'playwright/types/test';

test.describe(
    `Org tree tests, tags:  ${tags.resource.orgTree} ${tags.tool.api} ${tags.teams.viper}`,
    { tag: [tags.services.orgUnitService, tags.runOn.serialCI] },
    () => {
        let defaultDepartmentId: number;
        let createdDepartment: Department;
        let createdProject1: Project;
        let createdProject2: Project;
        let createdAccessRule: AccessRule;

        test.beforeAll(async ({}, testInfo: TestInfo) => {
            if (await clusterUtils.isMoreThanOneClusterInEnvironment()) {
                logger.info('More than one cluster is found');
                skipPlaywrightTest(test, testInfo, 'Test only runs with single cluster');
                return; // Exit early, don't do setup
            }

            backendClusterApi.nodePoolsApi.accessToken =
                flowRootState.users.system_administrator.accessToken;
            await operations.forceCleanup();
            defaultDepartmentId = (await departmentUtils.getDefaultDepartmentByCluster()).id;
            createdDepartment = (await departmentUtils.createBackendDepartment()).body;
            let projectName = objectGenerator.generateNameByPrefix('p1');
            let namespace = Utils.namespaceFromProjectName(projectName);
            createdProject1 = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                departmentId: defaultDepartmentId,
                deservedGpus: 0,
                maxAllowedGpus: 0,
                nodePoolsResources: []
            });
            projectName = objectGenerator.generateNameByPrefix('p2');
            namespace = Utils.namespaceFromProjectName(projectName);
            createdProject2 = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                departmentId: Number(createdDepartment.id),
                deservedGpus: 0,
                maxAllowedGpus: 0,
                nodePoolsResources: []
            });
        });

        test.beforeEach(async () => {
            if (createdAccessRule) {
                await authorizationUtils.deleteAccessRule(createdAccessRule);
                createdAccessRule = null;
            }
        });

        test.afterAll(async () => {
            if (createdProject1) {
                await operations.deleteProjectAndNamespace(createdProject1.id);
                createdProject1 = null;
            }
            if (createdProject2) {
                await operations.deleteProjectAndNamespace(createdProject2.id);
                createdProject2 = null;
            }
            if (createdDepartment) {
                await departmentUtils.deleteDepartment(Number(createdDepartment.id));
                createdDepartment = null;
            }
        });

        test('Get in context of system admin on the tenant scope', async () => {
            logger.info('Validate org tree from org unit service');
            orgUnitService.orgUnitApi.accessToken = flowRootState.users.system_administrator.accessToken;
            const orgTree = await orgUnitService.orgUnitApi.getOrgTree();
            expect(orgTree.response.statusCode).toBe(StatusCodes.OK);
            if (settings.isSelfHosted) {
                // skip this test if we run on SaaS
                expect(orgTree.body.orgUnits.length).toBe(6);
            }
            expect(validateTree(orgTree.body.orgUnits as OrgUnit[])).toBe(true);
        });

        test('Get in context of l1 researcher on the tenant scope', async () => {
            const researcher = flowRootState.users.l1_researcher;
            logger.info('Validate org tree from org unit service');
            orgUnitService.orgUnitApi.accessToken = researcher.accessToken;
            const orgTree = await orgUnitService.orgUnitApi.getOrgTree();
            expect(orgTree.response.statusCode).toBe(StatusCodes.OK);
            if (settings.isSelfHosted) {
                // skip this test if we run on SaaS
                expect(orgTree.body.orgUnits.length).toBe(6);
            }
            expect(validateTree(orgTree.body.orgUnits as OrgUnit[])).toBe(true);
        });

        test('Get in context of l1 researcher on a project scope', async () => {
            const researcher = flowRootState.users.l1_researcher;
            createdAccessRule = await authorizationUtils.createAccessRuleWithDefaults(
                researcher.email,
                SubjectType.User,
                researcher.role.id,
                createdProject1.id.toString(),
                ScopeType.Project
            );
            logger.info('Validate org tree from org unit service');
            orgUnitService.orgUnitApi.accessToken = researcher.accessToken;
            const orgTree = await orgUnitService.orgUnitApi.getOrgTree();
            expect(orgTree.response.statusCode).toBe(StatusCodes.OK);
            expect(orgTree.body.orgUnits.length).toBe(6);
            expect(validateTree(orgTree.body.orgUnits as OrgUnit[])).toBe(true);
        });

        test('Validate department operations - System Admin has full access', async () => {
            const adminToken = flowRootState.users.system_administrator.accessToken;

            logger.info('Validate CreateDepartment operation');
            const departmentName = objectGenerator.generateNameByPrefix('dept');
            const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();
            const createRequest = orgUnitUtils.getDepartmentCreateRequestFromV1(
                {
                    name: departmentName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                },
                flowRootState.cluster.uuid,
                defaultNodePool
            );
            const createResponse = await orgUnitUtils.createDepartmentApi(createRequest, adminToken);
            expect(createResponse.response.statusCode).toBe(StatusCodes.CREATED);
            const tempDepartment = createResponse.body;

            logger.info('Validate GetDepartment operation');
            const getResponse = await orgUnitUtils.getDepartmentApi(tempDepartment.id, adminToken);
            expect(getResponse.response.statusCode).toBe(StatusCodes.OK);

            logger.info('Validate UpdateResources operation');
            const updateResponse = await orgUnitUtils.updateDepartmentResourcesApi(
                tempDepartment.id,
                [],
                adminToken
            );
            expect(updateResponse.response.statusCode).toBe(StatusCodes.OK);

            const now = new Date();
            const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000);
            const metricsResponse = await orgUnitUtils.getDepartmentMetricsApi(
                tempDepartment.id,
                ['GPU_ALLOCATION' as any],
                oneHourAgo,
                now,
                undefined,
                adminToken
            );
            expect(metricsResponse.response.statusCode).toBe(StatusCodes.OK);

            // Cleanup
            await departmentUtils.deleteDepartment(Number(tempDepartment.id));
        });
    }
);
