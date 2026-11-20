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
    computeResourceUtils,
    waitUtils,
    workloadsUtilsV2
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

test.describe(`CLI Submit With Compute - Workload Type Coverage Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        // Helper function: Create compute resource
        async function createComputeResource(computeName: string): Promise<string> {
            const computeCreationRequest = {
                meta: {
                    name: computeName,
                    scope: Scope.Project,
                    projectId: project.id,
                    clusterId: flowRootState.cluster.uuid
                },
                spec: {
                    cpuCoreRequest: 0.1,
                    cpuMemoryRequest: '100M',
                    gpuDevicesRequest: 0
                }
            };

            const computeAsset = await computeResourceUtils.create(computeCreationRequest);
            logger.info(`Created compute resource: ${computeName} (ID: ${computeAsset.meta.id})`);
            return computeAsset.meta.id;
        }

        // Helper function: Verify compute resource is attached
        async function verifyComputeResourceAttached(workloadName: string): Promise<void> {
            const workloadId = await workloadsUtilsV2.getWorkloadIdByName(workloadName, project.id);
            const workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;

            const requestedCpu = workload.workloadRequestedResources?.cpu?.request ?? 0;
            expect(requestedCpu).toBeCloseTo(0.1, 0.01);
            logger.info('Verified: Compute resource configuration applied to workload');
        }

        // Helper function: Cleanup workload and compute resource
        async function cleanupWorkloadAndCompute(
            workloadName: string,
            workloadType: any,
            computeId: string
        ): Promise<void> {
            try {
                runaiCliV2.delete({}, workloadName, workloadType);
                await computeResourceUtils.delete(computeId);
                logger.info(`Cleaned up workload: ${workloadName} and compute resource: ${computeId}`);
            } catch (error) {
                logger.error('Cleanup error:', error);
            }
        }

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-compute-cli', 5);
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

        // Test configuration array
        const workloadTestCases = [
            {
                id: 'T5590849',
                name: 'pytorch training',
                type: cliV2WorkloadType.training,
                prefix: 'pytorch',
                image: 'pytorch/pytorch:latest',
                extraFlags: { [SubmitFlagsMerged.command]: '-- sleep 30' }
            },
            {
                id: 'T5590850',
                name: 'mpi training',
                type: cliV2WorkloadType.mpi,
                prefix: 'mpi',
                image: 'mpioperator/mpi-pi:latest',
                extraFlags: {
                    [SubmitFlagsMerged.command]: '-- sleep 30',
                    [SubmitFlagsMerged.workers]: 2
                }
            },
            {
                id: 'T5590851',
                name: 'tensorflow training',
                type: cliV2WorkloadType.tensorFlow,
                prefix: 'tf',
                image: 'tensorflow/tensorflow:latest',
                extraFlags: { [SubmitFlagsMerged.command]: '-- sleep 30' }
            },
            {
                id: 'T5590852',
                name: 'xgboost training',
                type: cliV2WorkloadType.xgBoost,
                prefix: 'xgb-train',
                image: 'xgboost/xgboost:latest',
                extraFlags: {
                    [SubmitFlagsMerged.command]: '-- sleep 30',
                    [SubmitFlagsMerged.workers]: 2
                }
            },
            {
                id: 'T5590853',
                name: 'jax training',
                type: cliV2WorkloadType.jax,
                prefix: 'jax-train',
                image: 'google/jax:latest',
                extraFlags: { [SubmitFlagsMerged.command]: '-- sleep 30' }
            },
            {
                id: 'T5590854',
                name: 'standard training',
                type: cliV2WorkloadType.training,
                prefix: 'std-train',
                image: 'ubuntu:latest',
                extraFlags: { [SubmitFlagsMerged.command]: '-- sleep 30' }
            },
            {
                id: 'T5590855',
                name: 'standard inference',
                type: cliV2WorkloadType.inference,
                prefix: 'inference',
                image: 'nginx:latest',
                extraFlags: { [SubmitFlagsMerged.servingPort]: '8080' }
            },
            {
                id: 'T5590856',
                name: 'workspace',
                type: cliV2WorkloadType.workspace,
                prefix: 'workspace',
                image: 'jupyter/base-notebook:latest',
                extraFlags: {}
            },
            {
                id: 'T5590857',
                name: 'xgboost workload',
                type: cliV2WorkloadType.xgBoost,
                prefix: 'xgb',
                image: 'xgboost/xgboost:latest',
                extraFlags: {
                    [SubmitFlagsMerged.command]: '-- sleep 30',
                    [SubmitFlagsMerged.workers]: 2
                }
            },
            {
                id: 'T5590858',
                name: 'jax workload',
                type: cliV2WorkloadType.jax,
                prefix: 'jax',
                image: 'google/jax:latest',
                extraFlags: { [SubmitFlagsMerged.command]: '-- sleep 30' }
            }
        ];

        // Generate test cases from configuration
        workloadTestCases.forEach((testCase) => {
            test(`${testCase.id}: Submit ${testCase.name} with compute resource`, async function () {
                const workloadName = objectGenerator.generateShortNameByPrefix(`test-${testCase.prefix}`, 5);
                const computeName = objectGenerator.generateShortNameByPrefix(
                    `compute-${testCase.prefix}`,
                    5
                );
                let computeId: string;

                await test.step('Create compute resource via API', async function () {
                    computeId = await createComputeResource(computeName);
                });

                await test.step(`Submit ${testCase.name} with compute resource`, async function () {
                    logger.info(`Submitting workload: ${workloadName} with compute: ${computeName}`);

                    const submitFlag = {
                        [SubmitFlagsMerged.compute]: computeName,
                        [SubmitFlagsMerged.image]: testCase.image,
                        ...testCase.extraFlags
                    } as any;

                    const stdout = runaiCliV2.submit(submitFlag, workloadName, testCase.type);
                    expect(stdout).not.toContain('Error');

                    await waitUtils.waitForBackendWorkload(workloadName, String(project.id), 60);
                    logger.info(`Workload created: ${workloadName}`);
                });

                await test.step('Verify compute resource is attached to workload', async function () {
                    await verifyComputeResourceAttached(workloadName);
                });

                await test.step('Cleanup', async function () {
                    await cleanupWorkloadAndCompute(workloadName, testCase.type, computeId);
                });
            });
        });
    });
});

test.describe(`CLI Submit With Compute - Scope Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-compute-scope', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            logger.info(`Created project: ${projectName} (ID: ${project.id})`);
            runaiCliV2.projectSet(projectName);
        });

        test.afterAll('Cleanup resources', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`Deleted project ${projectName}`);
            } catch (error) {
                logger.error('Failed to delete project:', error);
            }
        });

        test('T5590859: Submit with single compute resource', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-single', 5);
            const computeName = objectGenerator.generateShortNameByPrefix('compute-single', 5);
            let computeId: string;

            await test.step('Create compute resource', async function () {
                const computeCreationRequest = {
                    meta: {
                        name: computeName,
                        scope: Scope.Project,
                        projectId: project.id,
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
                logger.info(`Created compute resource: ${computeName}`);
            });

            await test.step('Submit workload with single compute resource', async function () {
                const submitFlag = {
                    [SubmitFlagsMerged.compute]: computeName,
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    [SubmitFlagsMerged.command]: '-- sleep 30'
                } as Partial<typeof SubmitFlagsMerged>;

                const stdout = runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                expect(stdout).not.toContain('Error');

                await waitUtils.waitForBackendWorkload(workloadName, String(project.id), 60);
                logger.info('Workload created successfully with single compute resource');
            });

            await test.step('Cleanup', async function () {
                try {
                    runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);
                    await computeResourceUtils.delete(computeId);
                } catch (error) {
                    logger.error('Cleanup error:', error);
                }
            });
        });

        test('T5590860: Submit with cluster-scoped compute resource', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-cluster', 5);
            const computeName = objectGenerator.generateShortNameByPrefix('compute-cluster', 5);
            let computeId: string;

            await test.step('Create cluster-scoped compute resource', async function () {
                const computeCreationRequest = {
                    meta: {
                        name: computeName,
                        scope: Scope.Cluster,
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
                logger.info(`Created cluster-scoped compute resource: ${computeName}`);
            });

            await test.step('Submit workload with cluster-scoped compute', async function () {
                const submitFlag = {
                    [SubmitFlagsMerged.compute]: computeName,
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    [SubmitFlagsMerged.command]: '-- sleep 30'
                } as Partial<typeof SubmitFlagsMerged>;

                const stdout = runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                expect(stdout).not.toContain('Error');

                await waitUtils.waitForBackendWorkload(workloadName, String(project.id), 60);
                logger.info('Project-level workload can access cluster-scoped compute resource');
            });

            await test.step('Cleanup', async function () {
                try {
                    runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);
                    await computeResourceUtils.delete(computeId);
                } catch (error) {
                    logger.error('Cleanup error:', error);
                }
            });
        });

        test('T5590862: Submit with project-scoped compute resource', async function () {
            const workloadName = objectGenerator.generateShortNameByPrefix('test-project', 5);
            const computeName = objectGenerator.generateShortNameByPrefix('compute-project', 5);
            let computeId: string;

            await test.step('Create project-scoped compute resource', async function () {
                const computeCreationRequest = {
                    meta: {
                        name: computeName,
                        scope: Scope.Project,
                        projectId: project.id,
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
                logger.info(`Created project-scoped compute resource: ${computeName}`);
            });

            await test.step('Submit workload with project-scoped compute', async function () {
                const submitFlag = {
                    [SubmitFlagsMerged.compute]: computeName,
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    [SubmitFlagsMerged.command]: '-- sleep 30'
                } as Partial<typeof SubmitFlagsMerged>;

                const stdout = runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                expect(stdout).not.toContain('Error');

                await waitUtils.waitForBackendWorkload(workloadName, String(project.id), 60);
                logger.info('Workload uses project-scoped compute resource');
            });

            await test.step('Cleanup', async function () {
                try {
                    runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);
                    await computeResourceUtils.delete(computeId);
                } catch (error) {
                    logger.error('Cleanup error:', error);
                }
            });
        });

        test.describe('Department-scoped compute resource test', () => {
            let department: any;
            let projectInDept: any;
            const orgUnitUtils = new OrgUnitUtils();

            test.afterAll('Cleanup department and project', async () => {
                try {
                    if (projectInDept?.id) {
                        await operations.deleteProjectAndNamespace(projectInDept.id);
                        logger.info(`Deleted project: ${projectInDept.name}`);
                    }
                    if (department?.id) {
                        await orgUnitUtils.deleteDepartment(department.id);
                        logger.info(`Deleted department: ${department.name}`);
                    }
                    // Reset project context
                    runaiCliV2.projectSet(projectName);
                } catch (error) {
                    logger.error('Failed to cleanup department/project:', error);
                }
            });

            test('T5590861: Submit with department-scoped compute resource', async function () {
                const workloadName = objectGenerator.generateShortNameByPrefix('test-dept', 5);
                const computeName = objectGenerator.generateShortNameByPrefix('compute-dept', 5);
                const departmentName = objectGenerator.generateShortNameByPrefix('dept', 5);
                const deptProjectName = objectGenerator.generateShortNameByPrefix('proj-dept', 5);
                let computeId: string;

                await test.step('Create department and project within department', async function () {
                    department = await orgUnitUtils.createDepartment({
                        name: departmentName,
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
                    logger.info(`Created department: ${departmentName} (ID: ${department.id})`);

                    projectInDept = await projectResourcesUtils.createProjectAndNamespace(
                        deptProjectName,
                        Number(department.id),
                        0,
                        0
                    );
                    logger.info(`Created project: ${projectInDept.name} in department: ${departmentName}`);
                });

                await test.step('Create department-scoped compute resource', async function () {
                    const computeCreationRequest = {
                        meta: {
                            name: computeName,
                            scope: Scope.Department,
                            departmentId: department.id
                        },
                        spec: {
                            cpuCoreRequest: 0.1,
                            cpuMemoryRequest: '100M',
                            gpuDevicesRequest: 0
                        }
                    };

                    const computeAsset = await computeResourceUtils.create(computeCreationRequest);
                    computeId = computeAsset.meta.id;
                    logger.info(`Created department-scoped compute resource: ${computeName}`);
                });

                await test.step('Submit workload from project in department', async function () {
                    runaiCliV2.projectSet(projectInDept.name);

                    const submitFlag = {
                        [SubmitFlagsMerged.compute]: computeName,
                        [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                        [SubmitFlagsMerged.command]: '-- sleep 30'
                    } as Partial<typeof SubmitFlagsMerged>;

                    const stdout = runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.training);
                    expect(stdout).not.toContain('Error');

                    await waitUtils.waitForBackendWorkload(workloadName, String(projectInDept.id), 60);
                    logger.info('Project-level workload can access department-scoped compute resource');
                });

                await test.step('Cleanup', async function () {
                    try {
                        runaiCliV2.delete({}, workloadName, cliV2WorkloadType.training);
                        await computeResourceUtils.delete(computeId);
                    } catch (error) {
                        logger.error('Cleanup error:', error);
                    }
                });
            });
        });
    });
});
