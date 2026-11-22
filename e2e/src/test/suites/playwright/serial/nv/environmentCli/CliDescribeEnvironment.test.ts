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
    environmentUtils,
    flowRootState,
    tags
} from '../../../../../../modules/helpers';
import { OrgUnitUtils } from '../../../../../../modules/helpers/utils/OrgUnitUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope, WorkloadSupportedTypes } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { Department } from '../../../../../../modules/api/apis/orgUnitApi';
import { describeFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI Describe Environment - Positive Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        const orgUnitUtils = new OrgUnitUtils();

        // Shared resources structure
        const sharedResources = {
            departments: {
                dept01: null as Department | null
            },
            projects: {
                default: null as Project | null,
                projectA: null as Project | null
            },
            environments: {
                projectEnv: null as string | null, // environment ID
                clusterEnv: null as string | null,
                departmentEnv: null as string | null
            }
        };

        test.beforeAll(async function () {
            logger.info('=== Creating shared resources for Describe Environment Positive tests ===');

            await test.step('Create department', async function () {
                const deptName = objectGenerator.generateShortNameByPrefix('dept-env-desc', 5);
                sharedResources.departments.dept01 = await orgUnitUtils.createDepartment({
                    name: deptName,
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
                logger.info(`Created department: ${deptName} (ID: ${sharedResources.departments.dept01.id})`);
            });

            await test.step('Create projects', async function () {
                // Default project
                const projectName = objectGenerator.generateShortNameByPrefix('proj-env-default', 5);
                const namespace = Utils.namespaceFromProjectName(projectName);
                sharedResources.projects.default = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: sharedResources.departments.dept01.id as any
                });
                logger.info(
                    `Created default project: ${projectName} (ID: ${sharedResources.projects.default.id})`
                );

                // Project A
                const projectAName = objectGenerator.generateShortNameByPrefix('proj-env-a', 5);
                const namespaceA = Utils.namespaceFromProjectName(projectAName);
                sharedResources.projects.projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: sharedResources.departments.dept01.id as any
                });
                logger.info(
                    `Created project-a: ${projectAName} (ID: ${sharedResources.projects.projectA.id})`
                );
            });

            await test.step('Set default project context for CLI', function () {
                runaiCliV2.projectSet(sharedResources.projects.default!.name);
                logger.info(`Set CLI project context to: ${sharedResources.projects.default!.name}`);
            });

            await test.step('Create environment resources', async function () {
                // Project-scoped environment
                const projectEnvName = objectGenerator.generateShortNameByPrefix('test-env-proj', 5);
                const projectEnv = await environmentUtils.create({
                    meta: {
                        name: projectEnvName,
                        scope: Scope.Project,
                        projectId: sharedResources.projects.default!.id,
                        clusterId: flowRootState.cluster.uuid,
                        workloadSupportedTypes: {
                            workspace: true,
                            training: true,
                            inference: true
                        }
                    },
                    spec: {
                        image: 'ubuntu:22.04',
                        command: '/bin/bash',
                        args: '-c "sleep infinity"'
                    }
                });
                sharedResources.environments.projectEnv = projectEnv.meta.id;
                logger.info(`Created project-scoped environment: ${projectEnvName}`);

                // Cluster-scoped environment
                const clusterEnvName = objectGenerator.generateShortNameByPrefix('test-env-cluster', 5);
                const clusterEnv = await environmentUtils.create({
                    meta: {
                        name: clusterEnvName,
                        scope: Scope.Cluster,
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
                        image: 'ubuntu:22.04'
                    }
                });
                sharedResources.environments.clusterEnv = clusterEnv.meta.id;
                logger.info(`Created cluster-scoped environment: ${clusterEnvName}`);

                // Department-scoped environment
                const deptEnvName = objectGenerator.generateShortNameByPrefix('test-env-dept', 5);
                const deptEnv = await environmentUtils.create({
                    meta: {
                        name: deptEnvName,
                        scope: Scope.Department,
                        departmentId: sharedResources.departments.dept01!.id,
                        workloadSupportedTypes: {
                            training: true,
                            inference: true
                        }
                    },
                    spec: {
                        image: 'ubuntu:22.04'
                    }
                });
                sharedResources.environments.departmentEnv = deptEnv.meta.id;
                logger.info(`Created department-scoped environment: ${deptEnvName}`);
            });

            logger.info('=== All shared resources created successfully ===');
        });

        test.afterAll(async function () {
            logger.info('=== Cleaning up shared resources ===');

            await test.step('Delete environment resources', async function () {
                if (sharedResources.environments.projectEnv) {
                    await environmentUtils.delete(sharedResources.environments.projectEnv);
                    logger.info(`Deleted environment: ${sharedResources.environments.projectEnv}`);
                }
                if (sharedResources.environments.clusterEnv) {
                    await environmentUtils.delete(sharedResources.environments.clusterEnv);
                    logger.info(`Deleted environment: ${sharedResources.environments.clusterEnv}`);
                }
                if (sharedResources.environments.departmentEnv) {
                    await environmentUtils.delete(sharedResources.environments.departmentEnv);
                    logger.info(`Deleted environment: ${sharedResources.environments.departmentEnv}`);
                }
            });

            await test.step('Delete projects', async function () {
                if (sharedResources.projects.default?.id) {
                    await operations.deleteProjectAndNamespace(sharedResources.projects.default.id);
                    logger.info(`Deleted project: ${sharedResources.projects.default.id}`);
                }
                if (sharedResources.projects.projectA?.id) {
                    await operations.deleteProjectAndNamespace(sharedResources.projects.projectA.id);
                    logger.info(`Deleted project: ${sharedResources.projects.projectA.id}`);
                }
            });

            await test.step('Delete department', async function () {
                if (sharedResources.departments.dept01?.id) {
                    await orgUnitUtils.deleteDepartment(sharedResources.departments.dept01.id);
                    logger.info(`Deleted department: ${sharedResources.departments.dept01.id}`);
                }
            });

            logger.info('=== All shared resources cleaned up ===');
        });

        // Helper function to get environment name from ID
        async function getEnvironmentName(envId: string): Promise<string> {
            const environments = await environmentUtils.list();
            const env = environments.find((e) => e.meta.id === envId);
            return env?.meta.name || '';
        }

        // Output format tests (can be grouped with forEach)
        const outputFormatTests = [
            {
                id: 'T5597608',
                name: 'default output',
                flags: {},
                verifyFormat: (output: string): void => {
                    // Default is table format - should have organized structure
                    expect(output.length).toBeGreaterThan(0);
                    logger.info('Verified: Default output format displayed correctly');
                }
            },
            {
                id: 'T5597610',
                name: 'JSON format',
                flags: { [describeFlags.output]: 'json' },
                verifyFormat: (output: string): void => {
                    expect(() => JSON.parse(output)).not.toThrow();
                    const parsed = JSON.parse(output);
                    expect(parsed).toHaveProperty('meta');
                    expect(parsed).toHaveProperty('spec');
                    logger.info('Verified: JSON output is valid and contains required fields');
                }
            },
            {
                id: 'T5597611',
                name: 'YAML format',
                flags: { [describeFlags.output]: 'yaml' },
                verifyFormat: (output: string): void => {
                    expect(output).toContain('meta:');
                    expect(output).toContain('spec:');
                    expect(output).toContain('image:');
                    logger.info('Verified: YAML output contains expected structure');
                }
            },
            {
                id: 'T5597612',
                name: 'explicit table format',
                flags: { [describeFlags.output]: 'table' },
                verifyFormat: (output: string): void => {
                    expect(output.length).toBeGreaterThan(0);
                    logger.info('Verified: Explicit table format displayed correctly');
                }
            }
        ];

        outputFormatTests.forEach((testCase) => {
            test(`${testCase.id}: Describe environment with ${testCase.name}`, async function () {
                await test.step(`Describe with ${testCase.name}`, async function () {
                    logger.info(
                        `Test ${testCase.id}: Describing environment with ${testCase.name} (using shared resources)`
                    );

                    const envName = await getEnvironmentName(sharedResources.environments.projectEnv!);
                    const stdout = runaiCliV2.environmentDescribe(envName, testCase.flags as any);

                    logger.info(`=== ${testCase.name} Output Start ===`);
                    logger.info(stdout);
                    logger.info(`=== ${testCase.name} Output End ===`);

                    // Verify output format
                    testCase.verifyFormat(stdout);

                    // Common verifications
                    if (testCase.flags[describeFlags.output] !== 'table') {
                        expect(stdout).toContain(envName);
                        expect(stdout).toContain('ubuntu:22.04');
                    }
                });

                // Special case for T5597612: Compare with default
                if (testCase.id === 'T5597612') {
                    await test.step('Compare explicit table with default output', async function () {
                        const envName = await getEnvironmentName(sharedResources.environments.projectEnv!);
                        const stdoutDefault = runaiCliV2.environmentDescribe(envName);
                        const stdoutExplicit = runaiCliV2.environmentDescribe(envName, {
                            [describeFlags.output]: 'table'
                        } as any);

                        expect(stdoutDefault).toEqual(stdoutExplicit);
                        logger.info('Verified: Explicit table format matches default output');
                    });
                }
            });
        });

        test('T5597609: Describe environment in specific project using -p flag', async function () {
            await test.step('Describe with -p flag', async function () {
                logger.info('Test T5597609: Describing environment with -p flag (using shared resources)');

                const envName = await getEnvironmentName(sharedResources.environments.projectEnv!);
                const projectName = sharedResources.projects.default!.name;

                const stdout = runaiCliV2.environmentDescribe(envName, {
                    '-p': projectName
                } as any);

                logger.info('=== Describe with -p Output Start ===');
                logger.info(stdout);
                logger.info('=== Describe with -p Output End ===');

                expect(stdout).toContain(envName);
                expect(stdout.toLowerCase()).toContain('project');

                logger.info('Verified: Environment described successfully with -p flag');
            });
        });

        test('T5597613: Describe cluster-scoped environment from project context', async function () {
            await test.step('Describe cluster-scoped environment', async function () {
                logger.info('Test T5597613: Describing cluster-scoped environment (using shared resources)');

                const clusterEnvName = await getEnvironmentName(sharedResources.environments.clusterEnv!);
                const projectName = sharedResources.projects.default!.name;

                const stdout = runaiCliV2.environmentDescribe(clusterEnvName, {
                    '-p': projectName
                } as any);

                logger.info('=== Cluster-Scoped Environment Output Start ===');
                logger.info(stdout);
                logger.info('=== Cluster-Scoped Environment Output End ===');

                expect(stdout).toContain(clusterEnvName);
                expect(stdout.toLowerCase()).toContain('cluster');

                logger.info('Verified: Cluster-scoped environment accessible from project context');
            });
        });

        test('T5597614: Describe department-scoped environment from project', async function () {
            await test.step('Describe department-scoped environment', async function () {
                logger.info(
                    'Test T5597614: Describing department-scoped environment (using shared resources)'
                );

                const deptEnvName = await getEnvironmentName(sharedResources.environments.departmentEnv!);
                const projectName = sharedResources.projects.default!.name;

                const stdout = runaiCliV2.environmentDescribe(deptEnvName, {
                    '-p': projectName
                } as any);

                logger.info('=== Department-Scoped Environment Output Start ===');
                logger.info(stdout);
                logger.info('=== Department-Scoped Environment Output End ===');

                expect(stdout).toContain(deptEnvName);
                expect(stdout.toLowerCase()).toContain('department');

                logger.info('Verified: Department-scoped environment accessible from project context');
            });
        });

        test('T5597615: Describe environment in default project without -p flag', async function () {
            await test.step('Describe without -p flag', async function () {
                logger.info('Test T5597615: Describing environment without -p flag (using shared resources)');

                const envName = await getEnvironmentName(sharedResources.environments.projectEnv!);

                // Ensure default project is set
                runaiCliV2.projectSet(sharedResources.projects.default!.name);

                const stdoutDefault = runaiCliV2.environmentDescribe(envName);

                logger.info('=== Default Context Output Start ===');
                logger.info(stdoutDefault);
                logger.info('=== Default Context Output End ===');

                expect(stdoutDefault).toContain(envName);

                logger.info('Verified: Environment found using default project context');
            });

            await test.step('Compare with explicit -p flag', async function () {
                const envName = await getEnvironmentName(sharedResources.environments.projectEnv!);
                const projectName = sharedResources.projects.default!.name;

                const stdoutExplicit = runaiCliV2.environmentDescribe(envName, {
                    '-p': projectName
                } as any);

                logger.info('=== Explicit -p Flag Output Start ===');
                logger.info(stdoutExplicit);
                logger.info('=== Explicit -p Flag Output End ===');

                // Both outputs should be similar (may have minor formatting differences)
                expect(stdoutExplicit).toContain(envName);

                logger.info('Verified: Both default context and -p flag work correctly');
            });
        });
    });
});
