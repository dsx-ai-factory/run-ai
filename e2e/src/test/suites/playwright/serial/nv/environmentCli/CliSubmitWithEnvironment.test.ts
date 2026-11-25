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
    environmentUtils,
    waitUtils,
    orgUnitUtils,
    workloadUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import {
    Scope,
    EnvironmentCreationRequest,
    WorkloadSupportedTypes
} from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliV2WorkloadType,
    cliV2TrainingType,
    SubmitFlagsMerged,
    cliLoginAsSystemAdmin
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI Submit with Environment - Positive Tests, tags: ${tags.resource.environment} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let createdWorkloadIds: string[] = [];
        let createdEnvIds: string[] = [];
        let createdProjectIds: number[] = [];
        let createdDepartmentIds: number[] = [];

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-env-submit', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            await test.step('Create project via API', async function () {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2
                });
                logger.info(`Created project: ${projectName} (ID: ${project.id})`);
            });

            await test.step('Ensure CLI is logged in', function () {
                cliLoginAsSystemAdmin();
                logger.info('Ensured CLI is logged in as system admin');
            });

            await test.step('Set project context for CLI', function () {
                runaiCliV2.projectSet(projectName);
                logger.info(`Set CLI project context to: ${projectName}`);
            });
        });

        test.afterAll(async function () {
            await test.step('Cleanup project', async function () {
                if (project?.id) {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`Deleted project: ${projectName}`);
                }
            });
        });

        test.afterEach(async function () {
            if (createdWorkloadIds.length > 0) {
                try {
                    await workloadUtils.deleteMultipleWorkloads(createdWorkloadIds);
                    logger.info(`Deleted workloads: ${createdWorkloadIds.join(', ')}`);
                } catch (error) {
                    logger.error(`Failed to delete workloads ${createdWorkloadIds.join(', ')}`, error);
                }
            }

            await Promise.all(
                createdEnvIds.map(async (envId) => {
                    try {
                        await environmentUtils.delete(
                            envId,
                            flowRootState.users.system_administrator.idToken
                        );
                        logger.info(`Deleted environment: ${envId}`);
                    } catch (error) {
                        logger.error(`Failed to delete environment ${envId}`, error);
                    }
                })
            );

            await Promise.all(
                createdProjectIds.map(async (projectId) => {
                    try {
                        await operations.deleteProjectAndNamespace(projectId);
                        logger.info(`Deleted project: ${projectId}`);
                    } catch (error) {
                        logger.error(`Failed to delete project ${projectId}`, error);
                    }
                })
            );

            await Promise.all(
                createdDepartmentIds.map(async (departmentId) => {
                    try {
                        await orgUnitUtils.deleteDepartment(departmentId.toString());
                        logger.info(`Deleted department: ${departmentId}`);
                    } catch (error) {
                        logger.error(`Failed to delete department ${departmentId}`, error);
                    }
                })
            );

            createdWorkloadIds = [];
            createdEnvIds = [];
            createdProjectIds = [];
            createdDepartmentIds = [];
        });

        // Helper function: Create environment
        async function createEnvironment(
            envName: string,
            scope: Scope,
            scopeId: any,
            image: string,
            workloadType?: cliV2WorkloadType,
            trainingType?: cliV2TrainingType
        ): Promise<string> {
            // Configure workloadSupportedTypes based on workload type
            const workloadSupportedTypes: any = {
                workspace: true,
                training: true,
                inference: true
            };

            // Distributed training workloads and xgboost/jax workloads need distributed: true and distFramework
            const needsDistributed =
                (workloadType === cliV2WorkloadType.training &&
                    trainingType &&
                    [
                        cliV2TrainingType.pyTorch,
                        cliV2TrainingType.mpi,
                        cliV2TrainingType.tensorFlow,
                        cliV2TrainingType.xgBoost,
                        cliV2TrainingType.jax
                    ].includes(trainingType)) ||
                workloadType === cliV2WorkloadType.xgBoost ||
                workloadType === cliV2WorkloadType.jax;

            if (needsDistributed) {
                workloadSupportedTypes.distributed = true;
                workloadSupportedTypes.distFramework = WorkloadSupportedTypes.DistFrameworkEnum.PyTorch;
            }

            const environmentCreationRequest: EnvironmentCreationRequest = {
                meta: {
                    name: envName,
                    scope: scope,
                    ...(scope === Scope.Project && { projectId: Number(scopeId) }),
                    ...(scope === Scope.Department && { departmentId: String(scopeId) }),
                    ...(scope === Scope.Cluster && { clusterId: String(scopeId) }),
                    workloadSupportedTypes: workloadSupportedTypes
                },
                spec: {
                    image: image
                }
            };
            const envAsset = await environmentUtils.create(environmentCreationRequest);
            createdEnvIds.push(envAsset.meta.id);
            logger.info(`Created environment: ${envName} (ID: ${envAsset.meta.id})`);
            return envAsset.meta.id;
        }

        // Test configuration array for Workload Type Coverage
        const workloadTestCases = [
            {
                id: 'T5590867',
                name: 'pytorch training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.pyTorch,
                prefix: 'pytorch',
                image: 'pytorch/pytorch:latest',
                extraFlags: { '-g': '1' }
            },
            {
                id: 'T5590868',
                name: 'mpi training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.mpi,
                prefix: 'mpi',
                image: 'mpioperator/mpi-pi:latest',
                extraFlags: { '-g': '1', '--workers': '2' }
            },
            {
                id: 'T5590869',
                name: 'tensorflow training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.tensorFlow,
                prefix: 'tf',
                image: 'tensorflow/tensorflow:latest',
                extraFlags: { '-g': '1' }
            },
            {
                id: 'T5590870',
                name: 'xgboost training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.xgBoost,
                prefix: 'xgb-train',
                image: 'xgboost/xgboost:latest',
                extraFlags: { '--workers': '2' }
            },
            {
                id: 'T5590871',
                name: 'jax training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.jax,
                prefix: 'jax-train',
                image: 'google/jax:latest',
                extraFlags: { '-g': '1' }
            },
            {
                id: 'T5590872',
                name: 'standard training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.standard,
                prefix: 'std-train',
                image: 'ubuntu:latest',
                extraFlags: { '-g': '1' }
            },
            {
                id: 'T5590873',
                name: 'standard inference',
                type: cliV2WorkloadType.inference,
                trainingType: null,
                prefix: 'inference',
                image: 'nginx:latest',
                extraFlags: { '-g': '1', '--serving-port': '8080' }
            },
            {
                id: 'T5590874',
                name: 'workspace',
                type: cliV2WorkloadType.workspace,
                trainingType: null,
                prefix: 'workspace',
                image: 'jupyter/base-notebook:latest',
                extraFlags: { '-g': '1' }
            },
            {
                id: 'T5590875',
                name: 'xgboost workload',
                type: cliV2WorkloadType.xgBoost,
                trainingType: null,
                prefix: 'xgboost',
                image: 'xgboost/xgboost:latest',
                extraFlags: { '--workers': '2' }
            },
            {
                id: 'T5590876',
                name: 'jax workload',
                type: cliV2WorkloadType.jax,
                trainingType: null,
                prefix: 'jax',
                image: 'google/jax:latest',
                extraFlags: { '-g': '1' }
            }
        ];

        // Generate test cases from configuration
        workloadTestCases.forEach((testCase) => {
            test(`${testCase.id}: Submit ${testCase.name} with environment`, async function () {
                const envName = objectGenerator.generateShortNameByPrefix(`env-${testCase.prefix}`, 5);
                const workloadName = objectGenerator.generateShortNameByPrefix(`test-${testCase.prefix}`, 5);

                await test.step('Create environment via API', async function () {
                    await createEnvironment(
                        envName,
                        Scope.Project,
                        project.id,
                        testCase.image,
                        testCase.type,
                        testCase.trainingType
                    );
                });

                await test.step(`Submit ${testCase.name} with environment`, function () {
                    logger.info(`Submitting workload: ${workloadName} with environment: ${envName}`);
                    const submitFlag = {
                        [SubmitFlagsMerged.environment]: [envName],
                        [SubmitFlagsMerged.image]: testCase.image,
                        ...testCase.extraFlags
                    } as any;

                    const stdout = testCase.trainingType
                        ? runaiCliV2.submit(submitFlag, workloadName, testCase.type, testCase.trainingType)
                        : runaiCliV2.submit(submitFlag, workloadName, testCase.type);

                    expect(stdout).not.toContain('Error');
                    logger.info('Workload created: ' + workloadName);
                });

                await test.step('Verify workload is created', async function () {
                    const workload = await waitUtils.waitForBackendWorkload(
                        workloadName,
                        String(project.id),
                        60
                    );
                    createdWorkloadIds.push(workload.id);
                    logger.info('Verified: Workload created successfully');
                });
            });
        });

        test('T5590877: Submit with single environment', async function () {
            const envName = objectGenerator.generateShortNameByPrefix('env-single', 5);
            const workloadName = objectGenerator.generateShortNameByPrefix('test-single-env', 5);

            await test.step('Create environment via API', async function () {
                await createEnvironment(
                    envName,
                    Scope.Project,
                    project.id,
                    'pytorch/pytorch:latest',
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
            });

            await test.step('Submit pytorch training with single environment', function () {
                logger.info(`Submitting workload: ${workloadName} with environment: ${envName}`);
                const submitFlag = {
                    [SubmitFlagsMerged.environment]: [envName],
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    '-g': '1'
                } as any;

                const stdout = runaiCliV2.submit(
                    submitFlag,
                    workloadName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );

                expect(stdout).not.toContain('Error');
                logger.info('Workload created: ' + workloadName);
            });

            await test.step('Verify workload is created', async function () {
                const workload = await waitUtils.waitForBackendWorkload(workloadName, String(project.id), 60);
                createdWorkloadIds.push(workload.id);
                logger.info('Verified: Workload created successfully');
            });
        });

        test('T5590878: Submit with multiple environments', async function () {
            const env1Name = objectGenerator.generateShortNameByPrefix('env-base', 5);
            const env2Name = objectGenerator.generateShortNameByPrefix('env-ml-tools', 5);
            const env3Name = objectGenerator.generateShortNameByPrefix('env-monitor', 5);
            const workloadName = objectGenerator.generateShortNameByPrefix('test-multi-env', 5);

            await test.step('Create multiple environments via API', async function () {
                await createEnvironment(
                    env1Name,
                    Scope.Project,
                    project.id,
                    'pytorch/pytorch:latest',
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
                await createEnvironment(
                    env2Name,
                    Scope.Project,
                    project.id,
                    'pytorch/pytorch:latest',
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
                await createEnvironment(
                    env3Name,
                    Scope.Project,
                    project.id,
                    'pytorch/pytorch:latest',
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
            });

            await test.step('Submit pytorch training with multiple environments', function () {
                logger.info(`Submitting workload: ${workloadName} with 3 environments`);
                const submitFlag = {
                    [SubmitFlagsMerged.environment]: [env1Name, env2Name, env3Name],
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    '-g': '1'
                } as Partial<typeof SubmitFlagsMerged>;

                const stdout = runaiCliV2.submit(
                    submitFlag,
                    workloadName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );

                expect(stdout).not.toContain('Error');
                logger.info('Workload created with multiple environments: ' + workloadName);
            });

            await test.step('Verify workload is created', async function () {
                const workload = await waitUtils.waitForBackendWorkload(workloadName, String(project.id), 60);
                createdWorkloadIds.push(workload.id);
                logger.info('Verified: Workload created successfully');
            });
        });

        test('T5590879: Submit with cluster-scoped environment', async function () {
            const envName = objectGenerator.generateShortNameByPrefix('env-cluster', 5);
            const workloadName = objectGenerator.generateShortNameByPrefix('test-cluster-env', 5);

            await test.step('Create cluster-scoped environment via API', async function () {
                await createEnvironment(
                    envName,
                    Scope.Cluster,
                    flowRootState.cluster.uuid,
                    'ubuntu:latest',
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
            });

            await test.step('Submit pytorch training with cluster-scoped environment', function () {
                logger.info(
                    `Submitting workload: ${workloadName} with cluster-scoped environment: ${envName}`
                );
                const submitFlag = {
                    [SubmitFlagsMerged.environment]: [envName],
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    '-g': '1'
                } as any;

                const stdout = runaiCliV2.submit(
                    submitFlag,
                    workloadName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );

                expect(stdout).not.toContain('Error');
                logger.info('Workload created with cluster-scoped environment: ' + workloadName);
            });

            await test.step('Verify workload is created', async function () {
                const workload = await waitUtils.waitForBackendWorkload(workloadName, String(project.id), 60);
                createdWorkloadIds.push(workload.id);
                logger.info('Verified: Workload created successfully');
            });
        });

        test('T5590880: Submit with department-scoped environment', async function () {
            const deptName = objectGenerator.generateShortNameByPrefix('dept-env', 5);
            const projectInDeptName = objectGenerator.generateShortNameByPrefix('proj-in-dept', 5);
            const namespaceInDept = Utils.namespaceFromProjectName(projectInDeptName);
            const envName = objectGenerator.generateShortNameByPrefix('env-dept', 5);
            const workloadName = objectGenerator.generateShortNameByPrefix('test-dept-env', 5);
            let departmentId: number;
            let projectInDept: Project;

            await test.step('Create department', async function () {
                const dept = await orgUnitUtils.createDepartment({
                    name: deptName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            gpu: { deserved: 2, limit: -1 },
                            cpu: null,
                            memory: null,
                            nodePool: null
                        }
                    ]
                });
                departmentId = Number(dept.id);
                createdDepartmentIds.push(departmentId);
                logger.info(`Created department: ${deptName}`);
            });

            await test.step('Create project in department', async function () {
                projectInDept = await projectResourcesUtils.createProject(namespaceInDept, {
                    name: projectInDeptName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2,
                    departmentId: departmentId
                });
                createdProjectIds.push(projectInDept.id);
                logger.info(`Created project in department: ${projectInDeptName}`);
            });

            await test.step('Create department-scoped environment via API', async function () {
                await createEnvironment(
                    envName,
                    Scope.Department,
                    departmentId.toString(),
                    'pytorch/pytorch:latest',
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
            });

            await test.step('Submit pytorch training with department-scoped environment', function () {
                logger.info(
                    `Submitting workload: ${workloadName} with department-scoped environment: ${envName}`
                );

                // Temporarily switch project context
                runaiCliV2.projectSet(projectInDeptName);

                const submitFlag = {
                    [SubmitFlagsMerged.environment]: [envName],
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    '-g': '1'
                } as any;

                const stdout = runaiCliV2.submit(
                    submitFlag,
                    workloadName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );

                expect(stdout).not.toContain('Error');
                logger.info('Workload created with department-scoped environment: ' + workloadName);

                // Restore original project context
                runaiCliV2.projectSet(projectName);
            });

            await test.step('Verify workload is created', async function () {
                const workload = await waitUtils.waitForBackendWorkload(
                    workloadName,
                    String(projectInDept.id),
                    60
                );
                createdWorkloadIds.push(workload.id);
                logger.info('Verified: Workload created successfully');
            });
        });

        test('T5590881: Submit with project-scoped environment', async function () {
            const testProjectName = objectGenerator.generateShortNameByPrefix('proj-test', 5);
            const testNamespace = Utils.namespaceFromProjectName(testProjectName);
            const envName = objectGenerator.generateShortNameByPrefix('env-project', 5);
            const workloadName = objectGenerator.generateShortNameByPrefix('test-proj-env', 5);
            let testProject: Project;

            await test.step('Create test project', async function () {
                testProject = await projectResourcesUtils.createProject(testNamespace, {
                    name: testProjectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2
                });
                createdProjectIds.push(testProject.id);
                logger.info(`Created test project: ${testProjectName}`);
            });

            await test.step('Create project-scoped environment via API', async function () {
                await createEnvironment(
                    envName,
                    Scope.Project,
                    testProject.id,
                    'pytorch/pytorch:latest',
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );
            });

            await test.step('Submit pytorch training with project-scoped environment', function () {
                logger.info(
                    `Submitting workload: ${workloadName} with project-scoped environment: ${envName}`
                );

                // Temporarily switch project context
                runaiCliV2.projectSet(testProjectName);

                const submitFlag = {
                    [SubmitFlagsMerged.environment]: [envName],
                    [SubmitFlagsMerged.image]: 'pytorch/pytorch:latest',
                    '-g': '1'
                } as any;

                const stdout = runaiCliV2.submit(
                    submitFlag,
                    workloadName,
                    cliV2WorkloadType.training,
                    cliV2TrainingType.pyTorch
                );

                expect(stdout).not.toContain('Error');
                logger.info('Workload created with project-scoped environment: ' + workloadName);

                // Restore original project context
                runaiCliV2.projectSet(projectName);
            });

            await test.step('Verify workload is created', async function () {
                const workload = await waitUtils.waitForBackendWorkload(
                    workloadName,
                    String(testProject.id),
                    60
                );
                createdWorkloadIds.push(workload.id);
                logger.info('Verified: Workload created successfully');
            });
        });
    });
});
