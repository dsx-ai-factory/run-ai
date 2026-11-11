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
    orgUnitUtils
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
    cliV2TrainingType
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI Submit with Environment - Negative Tests, tags: ${tags.resource.environment} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        // Error message patterns
        const ERROR_ENVIRONMENT_NOT_FOUND =
            /failed to submit\. the environment .* does not support distributed training workloads/i;
        const ERROR_INVALID_ENVIRONMENT_NAME = /invalid environment name|environment name cannot be empty/i;

        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-env-neg', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            await test.step('Create project via API', async function () {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2
                });
                logger.info(`Created project: ${projectName} (ID: ${project.id})`);
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

        test('T5590882: Submit with non-existent environment', async function () {
            await test.step('Attempt to submit with non-existent environment', function () {
                logger.info('Attempting to submit with non-existent environment');
                const workloadName = objectGenerator.generateShortNameByPrefix('test-nonexist-env', 5);

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} training pytorch submit ${workloadName} ` +
                        '--environment non-existent-environment-xyz ' +
                        '--image pytorch/pytorch:latest ' +
                        '-g 1',
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_ENVIRONMENT_NOT_FOUND);
                logger.info('Verified: Command failed with appropriate error for non-existent environment');
            });
        });

        // Coverage remains disabled pending revalidation of this case
        test.skip('T5590883: Submit with empty environment name', async function () {
            await test.step('Attempt to submit with empty environment name', function () {
                logger.info('Attempting to submit with empty environment name');
                const workloadName = objectGenerator.generateShortNameByPrefix('test-empty-env', 5);

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} training pytorch submit ${workloadName} ` +
                        '--environment "" ' +
                        '--image pytorch/pytorch:latest ' +
                        '-g 1',
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_INVALID_ENVIRONMENT_NAME);
                logger.info('Verified: Command failed with appropriate error for empty environment name');
            });
        });

        test('T5590884: Submit with environment from different project (no access)', async function () {
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-a', 5);
            const projectBName = objectGenerator.generateShortNameByPrefix('proj-b', 5);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            const namespaceB = Utils.namespaceFromProjectName(projectBName);
            const envName = objectGenerator.generateShortNameByPrefix('env-isolated', 5);
            let projectA: Project;
            let projectB: Project;
            let envId: string;

            await test.step('Create project-a and project-b', async function () {
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 2
                });
                logger.info(`Created project-a: ${projectAName}`);

                projectB = await projectResourcesUtils.createProject(namespaceB, {
                    name: projectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 2
                });
                logger.info(`Created project-b: ${projectBName}`);
            });

            await test.step('Create project-scoped environment in project-b', async function () {
                const environmentCreationRequest: EnvironmentCreationRequest = {
                    meta: {
                        name: envName,
                        scope: Scope.Project,
                        projectId: projectB.id,
                        clusterId: flowRootState.cluster.uuid,
                        workloadSupportedTypes: {
                            workspace: true,
                            training: true,
                            inference: true,
                            distributed: true,
                            distFramework: WorkloadSupportedTypes.DistFrameworkEnum.PyTorch
                        }
                    },
                    spec: {
                        image: 'pytorch/pytorch:latest'
                    }
                };
                const envAsset = await environmentUtils.create(environmentCreationRequest);
                envId = envAsset.meta.id;
                logger.info(`Created environment in project-b: ${envName}`);
            });

            test.step('Attempt to submit from project-a using project-b environment', function () {
                logger.info('Attempting to submit from project-a using project-b environment');
                const workloadName = objectGenerator.generateShortNameByPrefix('test-cross-proj', 5);

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} training pytorch submit ${workloadName} ` +
                        `--environment ${envName} ` +
                        '--image pytorch/pytorch:latest ' +
                        '-g 1 ' +
                        `-p ${projectAName}`,
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_ENVIRONMENT_NOT_FOUND);
                logger.info('Verified: project-a cannot access project-b environments');
            });

            await test.step('Cleanup environment and projects', async function () {
                await environmentUtils.delete(envId, flowRootState.users.system_administrator.idToken);
                await operations.deleteProjectAndNamespace(projectA.id);
                await operations.deleteProjectAndNamespace(projectB.id);
                logger.info('Cleaned up all resources');
            });
        });

        test('T5590885: Submit with environment from different department (no access)', async function () {
            const deptAName = objectGenerator.generateShortNameByPrefix('dept-a', 5);
            const deptBName = objectGenerator.generateShortNameByPrefix('dept-b', 5);
            const projectAName = objectGenerator.generateShortNameByPrefix('proj-in-a', 5);
            const namespaceA = Utils.namespaceFromProjectName(projectAName);
            const envName = objectGenerator.generateShortNameByPrefix('env-dept-b', 5);
            let departmentAId: number;
            let departmentBId: number;
            let projectA: Project;
            let projectB: Project;
            let envId: string;

            await test.step('Create dept-a and dept-b', async function () {
                const deptA = await orgUnitUtils.createDepartment({
                    name: deptAName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            gpu: { deserved: 0, limit: -1 },
                            cpu: null,
                            memory: null,
                            nodePool: null
                        }
                    ]
                });
                departmentAId = Number(deptA.id);
                logger.info(`Created dept-a: ${deptAName}`);

                const deptB = await orgUnitUtils.createDepartment({
                    name: deptBName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            gpu: { deserved: 0, limit: -1 },
                            cpu: null,
                            memory: null,
                            nodePool: null
                        }
                    ]
                });
                departmentBId = Number(deptB.id);
                logger.info(`Created dept-b: ${deptBName}`);
            });

            await test.step('Create project-in-dept-a', async function () {
                projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 2,
                    departmentId: departmentAId
                });
                logger.info(`Created project in dept-a: ${projectAName}`);
            });

            await test.step('Create temporary project in dept-b for environment creation', async function () {
                const tempProjectBName = objectGenerator.generateShortNameByPrefix('proj-temp-b', 5);
                const tempNamespaceB = Utils.namespaceFromProjectName(tempProjectBName);
                projectB = await projectResourcesUtils.createProject(tempNamespaceB, {
                    name: tempProjectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: departmentBId
                });
                logger.info(`Created temporary project in dept-b: ${tempProjectBName}`);
            });

            await test.step('Create department-scoped environment in dept-b', async function () {
                const environmentCreationRequest: EnvironmentCreationRequest = {
                    meta: {
                        name: envName,
                        scope: Scope.Department,
                        departmentId: departmentBId.toString(),
                        workloadSupportedTypes: {
                            workspace: true,
                            training: true,
                            inference: true,
                            distributed: true,
                            distFramework: WorkloadSupportedTypes.DistFrameworkEnum.PyTorch
                        }
                    },
                    spec: {
                        image: 'pytorch/pytorch:latest'
                    }
                };
                const envAsset = await environmentUtils.create(environmentCreationRequest);
                envId = envAsset.meta.id;
                logger.info(`Created department-scoped environment in dept-b: ${envName}`);
            });

            test.step('Attempt to submit from project-in-dept-a using dept-b environment', function () {
                logger.info('Attempting to submit from dept-a using dept-b environment');
                const workloadName = objectGenerator.generateShortNameByPrefix('test-cross-dept', 5);

                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} training pytorch submit ${workloadName} ` +
                        `--environment ${envName} ` +
                        '--image pytorch/pytorch:latest ' +
                        '-g 1 ' +
                        `-p ${projectAName}`,
                    false
                );

                expect(result.status).not.toBe(0);
                expect(result.stderr).toMatch(ERROR_ENVIRONMENT_NOT_FOUND);
                logger.info('Verified: project in dept-a cannot access dept-b environments');
            });

            await test.step('Cleanup all resources', async function () {
                await environmentUtils.delete(envId, flowRootState.users.system_administrator.idToken);
                await operations.deleteProjectAndNamespace(projectA.id);
                await operations.deleteProjectAndNamespace(projectB.id);
                await orgUnitUtils.deleteDepartment(departmentAId.toString());
                await orgUnitUtils.deleteDepartment(departmentBId.toString());
                logger.info('Cleaned up all resources');
            });
        });
    });
});
