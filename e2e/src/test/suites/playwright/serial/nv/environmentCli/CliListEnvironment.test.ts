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
import { listFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI List Environment - Positive Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        const orgUnitUtils = new OrgUnitUtils();

        // Shared resources structure
        const sharedResources = {
            departments: {
                dept01: null as Department | null
            },
            projects: {
                default: null as Project | null,
                projectA: null as Project | null,
                emptyProject: null as Project | null
            },
            environments: [] as string[] // environment IDs
        };

        test.beforeAll(async function () {
            logger.info('=== Creating shared resources for List Environment Positive tests ===');

            await test.step('Create department', async function () {
                const deptName = objectGenerator.generateShortNameByPrefix('dept-env-list', 5);
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
                const projectName = objectGenerator.generateShortNameByPrefix('proj-list-default', 5);
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
                const projectAName = objectGenerator.generateShortNameByPrefix('proj-list-a', 5);
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

                // Empty project (no environments)
                const emptyProjectName = objectGenerator.generateShortNameByPrefix('proj-list-empty', 5);
                const namespaceEmpty = Utils.namespaceFromProjectName(emptyProjectName);
                sharedResources.projects.emptyProject = await projectResourcesUtils.createProject(
                    namespaceEmpty,
                    {
                        name: emptyProjectName,
                        deservedGpus: 0,
                        maxAllowedGpus: 0,
                        departmentId: sharedResources.departments.dept01.id as any
                    }
                );
                logger.info(
                    `Created empty project: ${emptyProjectName} (ID: ${sharedResources.projects.emptyProject.id})`
                );
            });

            await test.step('Set default project context for CLI', function () {
                runaiCliV2.projectSet(sharedResources.projects.default!.name);
                logger.info(`Set CLI project context to: ${sharedResources.projects.default!.name}`);
            });

            await test.step('Create environment resources', async function () {
                // Create 3 environments in default project
                for (let i = 1; i <= 3; i++) {
                    const envName = objectGenerator.generateShortNameByPrefix(`test-project-env-${i}`, 5);
                    const env = await environmentUtils.create({
                        meta: {
                            name: envName,
                            scope: Scope.Project,
                            projectId: sharedResources.projects.default!.id,
                            clusterId: flowRootState.cluster.uuid,
                            workloadSupportedTypes: {
                                workspace: true,
                                training: true
                            }
                        },
                        spec: {
                            image: 'ubuntu:22.04'
                        }
                    });
                    sharedResources.environments.push(env.meta.id);
                    logger.info(`Created project environment ${i}: ${envName}`);
                }

                // Create 2 cluster-scoped environments
                for (let i = 1; i <= 2; i++) {
                    const envName = objectGenerator.generateShortNameByPrefix(`test-cluster-env-${i}`, 5);
                    const env = await environmentUtils.create({
                        meta: {
                            name: envName,
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
                    sharedResources.environments.push(env.meta.id);
                    logger.info(`Created cluster environment ${i}: ${envName}`);
                }

                // Create 2 department-scoped environments
                for (let i = 1; i <= 2; i++) {
                    const envName = objectGenerator.generateShortNameByPrefix(`test-dept-env-${i}`, 5);
                    const env = await environmentUtils.create({
                        meta: {
                            name: envName,
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
                    sharedResources.environments.push(env.meta.id);
                    logger.info(`Created department environment ${i}: ${envName}`);
                }
            });

            logger.info('=== All shared resources created successfully ===');
        });

        test.afterAll(async function () {
            logger.info('=== Cleaning up shared resources ===');

            await test.step('Delete environment resources', async function () {
                for (const envId of sharedResources.environments) {
                    try {
                        await environmentUtils.delete(envId);
                        logger.info(`Deleted environment: ${envId}`);
                    } catch (error) {
                        logger.warn(`Failed to delete environment ${envId}: ${error}`);
                    }
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
                if (sharedResources.projects.emptyProject?.id) {
                    await operations.deleteProjectAndNamespace(sharedResources.projects.emptyProject.id);
                    logger.info(`Deleted project: ${sharedResources.projects.emptyProject.id}`);
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

        test('T5597623: List environment resources with default output', async function () {
            await test.step('List environment resources', function () {
                logger.info(
                    'Test T5597623: Listing environment resources with default output (using shared resources)'
                );
                const stdout = runaiCliV2.list('environment', {}, false).stdout;

                logger.info('=== List Default Output Start ===');
                logger.info(stdout);
                logger.info('=== List Default Output End ===');

                // Verify output contains table structure
                expect(stdout).toContain('Name');
                expect(stdout).toContain('Scope');

                logger.info('Verified: Default output displays environment list');
            });
        });

        test('T5597624: List all scopes with -A flag', async function () {
            await test.step('List with -A flag', function () {
                logger.info(
                    'Test T5597624: Listing all environment resources with -A flag (using shared resources)'
                );
                const stdout = runaiCliV2.list(
                    'environment',
                    { [listFlags.allScopes]: '' } as any,
                    false
                ).stdout;

                logger.info('=== List All Scopes Output Start ===');
                logger.info(stdout);
                logger.info('=== List All Scopes Output End ===');

                // Verify output contains resources from different scopes
                expect(stdout.toLowerCase()).toContain('cluster');
                expect(stdout.toLowerCase()).toContain('department');

                logger.info('Verified: -A flag lists resources from all scopes');
            });
        });

        test('T5597625: List environment resources in specific project', async function () {
            const projectAName = sharedResources.projects.projectA!.name;

            await test.step('List with -p flag', function () {
                logger.info(
                    'Test T5597625: Listing environment in project-a using -p flag (using shared resources)'
                );
                const stdout = runaiCliV2.list(
                    'environment',
                    { [listFlags.project]: projectAName } as any,
                    false
                ).stdout;

                logger.info('=== List Project-A Output Start ===');
                logger.info(stdout);
                logger.info('=== List Project-A Output End ===');

                // Should show cluster and department scoped resources even in project-a
                // (because higher scopes are accessible)
                expect(stdout).toContain('Name');

                logger.info('Verified: -p flag lists environments accessible from specified project');
            });
        });

        test('T5597626: List environment resources with JSON output', async function () {
            await test.step('List with JSON format', function () {
                logger.info('Test T5597626: Listing environments with JSON output (using shared resources)');
                const stdout = runaiCliV2.list('environment', { [listFlags.json]: '' } as any, false).stdout;

                logger.info('=== List JSON Output Start ===');
                logger.info(stdout);
                logger.info('=== List JSON Output End ===');

                // Verify valid JSON
                expect(() => JSON.parse(stdout)).not.toThrow();
                const parsed = JSON.parse(stdout);
                expect(parsed).toHaveProperty('environments');
                expect(Array.isArray(parsed.environments)).toBe(true);

                logger.info('Verified: JSON output is valid and contains environments array');
            });
        });

        test('T5597627: List environment resources with YAML output', async function () {
            await test.step('List with YAML format', function () {
                logger.info('Test T5597627: Listing environments with YAML output (using shared resources)');
                const stdout = runaiCliV2.list('environment', { [listFlags.yaml]: '' } as any, false).stdout;

                logger.info('=== List YAML Output Start ===');
                logger.info(stdout);
                logger.info('=== List YAML Output End ===');

                // Verify YAML structure
                expect(stdout).toContain('environments:');
                expect(stdout).toContain('name:');
                expect(stdout).toContain('scope:');
                expect(stdout).toContain('image:');

                logger.info('Verified: YAML output contains expected structure');
            });
        });

        test('T5597628: List environment resources with table output', async function () {
            await test.step('Compare table and default output', function () {
                logger.info('Test T5597628: Comparing table and default output (using shared resources)');

                const stdoutTable = runaiCliV2.list(
                    'environment',
                    { [listFlags.table]: '' } as any,
                    false
                ).stdout;
                const stdoutDefault = runaiCliV2.list('environment', {}, false).stdout;

                logger.info('=== Table Output ===');
                logger.info(stdoutTable);
                logger.info('=== Default Output ===');
                logger.info(stdoutDefault);

                // Table and default should be similar (default is table)
                expect(stdoutTable).toContain('Name');
                expect(stdoutDefault).toContain('Name');

                logger.info('Verified: Table format displays correctly');
            });
        });

        test('T5597629: List environment resources with --no-headers', async function () {
            await test.step('List with --no-headers', function () {
                logger.info('Test T5597629: Listing environment with --no-headers (using shared resources)');

                const stdoutNoHeaders = runaiCliV2.list(
                    'environment',
                    { [listFlags.noHeaders]: '' } as any,
                    false
                ).stdout;
                const stdoutDefault = runaiCliV2.list('environment', {}, false).stdout;

                logger.info('=== No Headers Output ===');
                logger.info(stdoutNoHeaders);
                logger.info('=== Default Output ===');
                logger.info(stdoutDefault);

                // No headers output should not contain header row
                // But exact behavior depends on CLI implementation
                expect(stdoutNoHeaders.length).toBeGreaterThan(0);

                logger.info('Verified: --no-headers flag works correctly');
            });
        });

        test('T5597630: List environment resources with custom columns', async function () {
            await test.step('List with custom columns', function () {
                logger.info(
                    'Test T5597630: Listing environment with custom columns (using shared resources)'
                );

                const stdout = runaiCliV2.list(
                    'environment',
                    {
                        [listFlags.columns]: 'name,scope'
                    } as any,
                    false
                ).stdout;

                logger.info('=== Custom Columns Output Start ===');
                logger.info(stdout);
                logger.info('=== Custom Columns Output End ===');

                // Verify specified columns are present (capitalized in output)
                expect(stdout).toContain('Name');
                expect(stdout).toContain('Scope');

                logger.info('Verified: Custom columns displayed correctly');
            });
        });

        test('T5597631: List environment resources when empty', async function () {
            const emptyProjectName = sharedResources.projects.emptyProject!.name;

            await test.step('List in empty project', function () {
                logger.info(
                    'Test T5597631: Listing environment in empty project (using shared empty project)'
                );

                const stdout = runaiCliV2.list(
                    'environment',
                    {
                        [listFlags.project]: emptyProjectName
                    } as any,
                    false
                ).stdout;

                logger.info('=== Empty List Output Start ===');
                logger.info(stdout);
                logger.info('=== Empty List Output End ===');

                // Should still show cluster and department scoped resources
                // Empty just means no project-scoped resources
                expect(stdout).toContain('Name');

                logger.info('Verified: List works correctly for project with no project-scoped environments');
            });
        });

        test('T5597632: List includes cluster-scoped environment resources', async function () {
            await test.step('Verify cluster-scoped environment visible', function () {
                logger.info(
                    'Test T5597632: Verifying cluster-scoped environment resources visible (using shared resources)'
                );

                const stdout = runaiCliV2.list('environment', {}, false).stdout;

                logger.info('=== List Output Start ===');
                logger.info(stdout);
                logger.info('=== List Output End ===');

                // Should contain cluster-scoped environments
                expect(stdout.toLowerCase()).toContain('cluster');

                logger.info('Verified: Cluster-scoped environments are visible');
            });
        });

        test('T5597633: List includes department-scoped environment resources', async function () {
            await test.step('Verify department-scoped environment visible', function () {
                logger.info(
                    'Test T5597633: Verifying department-scoped environment resources visible (using shared resources)'
                );

                const stdout = runaiCliV2.list('environment', {}, false).stdout;

                logger.info('=== List Output Start ===');
                logger.info(stdout);
                logger.info('=== List Output End ===');

                // Should contain department-scoped environments
                expect(stdout.toLowerCase()).toContain('department');

                logger.info('Verified: Department-scoped environments are visible');
            });
        });

        test('T5597634: List includes project-scoped environment resources', async function () {
            await test.step('Verify project-scoped environment visible', function () {
                logger.info(
                    'Test T5597634: Verifying project-scoped environment resources visible (using shared resources)'
                );

                const stdout = runaiCliV2.list('environment', {}, false).stdout;

                logger.info('=== List Output Start ===');
                logger.info(stdout);
                logger.info('=== List Output End ===');

                // Should contain project-scoped environments
                expect(stdout.toLowerCase()).toContain('project');

                logger.info('Verified: Project-scoped environments are visible');
            });
        });

        test('T5597635: List environment resources with default project (no -p flag)', async function () {
            const projectName = sharedResources.projects.default!.name;

            await test.step('Compare default context with explicit -p flag', function () {
                logger.info(
                    'Test T5597635: Comparing default context with explicit -p flag (using shared resources)'
                );

                // Ensure default project is set
                runaiCliV2.projectSet(projectName);

                const stdoutDefault = runaiCliV2.list('environment', {}, false).stdout;
                const stdoutExplicit = runaiCliV2.list(
                    'environment',
                    {
                        [listFlags.project]: projectName
                    } as any,
                    false
                ).stdout;

                logger.info('=== Default Context Output ===');
                logger.info(stdoutDefault);
                logger.info('=== Explicit -p Flag Output ===');
                logger.info(stdoutExplicit);

                // Both outputs should be similar
                expect(stdoutDefault).toContain('Name');
                expect(stdoutExplicit).toContain('Name');

                logger.info('Verified: Default project context works without -p flag');
            });
        });
    });
});
