/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    flowRootState,
    computeResourceUtils
} from '../../../../../../modules/helpers';
import { OrgUnitUtils } from '../../../../../../modules/helpers/utils/OrgUnitUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI Submit With Compute - Negative Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-compute-neg', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            logger.info(`Created project: ${projectName} (ID: ${project.id})`);

            // Set project context once for all tests
            runaiCliV2.projectSet(projectName);
            logger.info(`Set project context to: ${projectName}`);
        });

        test.afterAll('Cleanup resources', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`Deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`Failed to delete project ${projectName}:`, error);
            }
        });

        test('T5590863: Submit with non-existent compute resource', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-nonexistent', 5);
            const nonExistentCompute = 'non-existent-compute-xyz';

            await test.step('Attempt to submit workload with non-existent compute resource', function () {
                logger.info(
                    `Attempting to submit workload: ${workloadName} with non-existent compute: ${nonExistentCompute}`
                );

                const submitFlag = {
                    [SubmitFlagsMerged.compute]: nonExistentCompute,
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    [SubmitFlagsMerged.command]: 'sleep 30'
                } as Partial<typeof SubmitFlagsMerged>;

                expect(() => {
                    runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                }).toThrow();

                logger.info('Verified: Command failed as expected - compute resource does not exist');
            });
        });

        test('T5590864: Submit with empty compute resource name', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-empty', 5);

            await test.step('Attempt to submit workload with empty compute name', function () {
                logger.info('Attempting to submit workload with empty compute resource name');

                const submitFlag = {
                    [SubmitFlagsMerged.compute]: '',
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    [SubmitFlagsMerged.command]: 'sleep 30'
                } as Partial<typeof SubmitFlagsMerged>;

                expect(() => {
                    runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                }).toThrow();

                logger.info('Verified: Command failed as expected - empty compute name was rejected');
            });
        });

        test('T5590865: Submit with compute resource from different project (no access)', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-cross-proj', 5);
            const computeName = objectGenerator.generateShortNameByPrefix('compute-isolated', 5);
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-a', 5);
            const projectBName = objectGenerator.generateShortNameByPrefix('proj-b', 5);
            let projectA: Project;
            let projectB: Project;
            let computeId: string;

            await test.step('Create two projects', async function () {
                const namespaceA = Utils.namespaceFromProjectName(projectAName);
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project-a: ${projectAName}`);

                const namespaceB = Utils.namespaceFromProjectName(projectBName);
                projectB = await projectResourcesUtils.createProject(namespaceB, {
                    name: projectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project-b: ${projectBName}`);
            });

            await test.step('Create project-scoped compute resource in project-b', async function () {
                const computeCreationRequest = {
                    meta: {
                        name: computeName,
                        scope: Scope.Project,
                        projectId: projectB.id,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 0.1,
                        cpuMemoryRequest: '100M',
                        gpuDevicesRequest: 0
                    }
                };

                const computeAsset = await computeResourceUtils.create(computeCreationRequest);
                computeId = computeAsset.meta.id;
                logger.info(`Created compute resource in project-b: ${computeName}`);
            });

            await test.step('Attempt to submit from project-a using project-b compute resource', function () {
                runaiCliV2.projectSet(projectA.name);

                logger.info('Attempting to use project-b compute resource from project-a');

                const submitFlag = {
                    [SubmitFlagsMerged.compute]: computeName,
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    [SubmitFlagsMerged.command]: 'sleep 30'
                } as Partial<typeof SubmitFlagsMerged>;

                expect(() => {
                    runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                }).toThrow();

                logger.info('Verified: Project-a cannot access project-b compute resource');
            });

            await test.step('Cleanup: Delete compute resource and projects', async function () {
                try {
                    await computeResourceUtils.delete(computeId);
                    logger.info(`Deleted compute resource: ${computeId}`);

                    await operations.deleteProjectAndNamespace(projectA.id);
                    logger.info(`Deleted project-a: ${projectAName}`);

                    await operations.deleteProjectAndNamespace(projectB.id);
                    logger.info(`Deleted project-b: ${projectBName}`);
                } catch (error) {
                    logger.error('Cleanup error:', error);
                }
            });
        });

        test('T5590866: Submit with compute resource from different department (no access)', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-cross-dept', 5);
            const computeName = objectGenerator.generateShortNameByPrefix('compute-dept-b', 5);
            const deptAName = objectGenerator.generateShortNameByPrefix('dept-a', 5);
            const deptBName = objectGenerator.generateShortNameByPrefix('dept-b', 5);
            const projInDeptAName = objectGenerator.generateShortNameByPrefix('proj-dept-a', 5);
            let deptA: any;
            let deptB: any;
            let projectInDeptA: any;
            let computeId: string;
            const orgUnitUtils = new OrgUnitUtils();

            await test.step('Create two departments and project in dept-a', async function () {
                deptA = await orgUnitUtils.createDepartment({
                    name: deptAName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 },
                            cpu: null,
                            memory: null,
                            nodePool: null
                        }
                    ]
                });
                logger.info(`Created dept-a: ${deptAName} (ID: ${deptA.id})`);

                deptB = await orgUnitUtils.createDepartment({
                    name: deptBName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 },
                            cpu: null,
                            memory: null,
                            nodePool: null
                        }
                    ]
                });
                logger.info(`Created dept-b: ${deptBName} (ID: ${deptB.id})`);

                projectInDeptA = await projectResourcesUtils.createProjectAndNamespace(
                    projInDeptAName,
                    Number(deptA.id),
                    0,
                    0
                );
                logger.info(`Created project in dept-a: ${projectInDeptA.name}`);
            });

            await test.step('Create department-scoped compute resource in dept-b', async function () {
                const computeCreationRequest = {
                    meta: {
                        name: computeName,
                        scope: Scope.Department,
                        departmentId: deptB.id
                    },
                    spec: {
                        cpuCoreRequest: 0.1,
                        cpuMemoryRequest: '100M',
                        gpuDevicesRequest: 0
                    }
                };

                const computeAsset = await computeResourceUtils.create(computeCreationRequest);
                computeId = computeAsset.meta.id;
                logger.info(`Created compute resource in dept-b: ${computeName}`);
            });

            await test.step('Attempt to submit from project in dept-a using dept-b compute resource', function () {
                runaiCliV2.projectSet(projectInDeptA.name);

                logger.info('Attempting to use dept-b compute resource from project in dept-a');

                const submitFlag = {
                    [SubmitFlagsMerged.compute]: computeName,
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    [SubmitFlagsMerged.command]: 'sleep 30'
                } as Partial<typeof SubmitFlagsMerged>;

                expect(() => {
                    runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                }).toThrow();

                logger.info('Verified: Project in dept-a cannot access dept-b compute resource');
            });

            await test.step('Cleanup: Delete compute resource, projects, and departments', async function () {
                try {
                    await computeResourceUtils.delete(computeId);
                    logger.info(`Deleted compute resource: ${computeId}`);

                    await operations.deleteProjectAndNamespace(projectInDeptA.id);
                    logger.info(`Deleted project: ${projectInDeptA.name}`);

                    await orgUnitUtils.deleteDepartment(deptA.id);
                    logger.info(`Deleted dept-a: ${deptAName}`);

                    await orgUnitUtils.deleteDepartment(deptB.id);
                    logger.info(`Deleted dept-b: ${deptBName}`);
                } catch (error) {
                    logger.error('Cleanup error:', error);
                }
            });
        });
    });
});
