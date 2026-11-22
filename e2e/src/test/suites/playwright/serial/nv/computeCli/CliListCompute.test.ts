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
import { listFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI List Compute - Positive Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        const sharedResources = {
            projects: {
                default: null as Project | null,
                projectA: null as Project | null,
                emptyProject: null as Project | null
            },
            departments: {
                dept01: null as any
            },
            computeResources: {
                // Cluster scope
                clusterCompute1: null as any,
                clusterCompute2: null as any,
                // Department scope
                deptCompute1: null as any,
                deptCompute2: null as any,
                // Project scope (default)
                projectCompute1: null as any,
                projectCompute2: null as any,
                projectCompute3: null as any,
                // Project scope (project-a)
                projectACompute1: null as any,
                projectACompute2: null as any
            }
        };

        test.beforeAll(async function () {
            logger.info('=== Creating shared resources for List Positive tests ===');

            await test.step('Create department first', async function () {
                const deptName = objectGenerator.generateShortNameByPrefix('dept-list-01', 5);
                const orgUnitUtils = new OrgUnitUtils();

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

            await test.step('Create projects under department', async function () {
                const projectName = objectGenerator.generateShortNameByPrefix('proj-list-default', 5);
                const namespace = Utils.namespaceFromProjectName(projectName);

                // Create default project UNDER the department so it can see dept-scoped resources
                sharedResources.projects.default = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: sharedResources.departments.dept01.id // Associate with department
                });
                logger.info(
                    `Created default project: ${projectName} (ID: ${sharedResources.projects.default.id})`
                );

                const projectAName = objectGenerator.generateShortNameByPrefix('proj-list-a', 5);
                const namespaceA = Utils.namespaceFromProjectName(projectAName);

                sharedResources.projects.projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: sharedResources.departments.dept01.id // Associate with department
                });
                logger.info(
                    `Created project-a: ${projectAName} (ID: ${sharedResources.projects.projectA.id})`
                );

                const emptyProjectName = objectGenerator.generateShortNameByPrefix('proj-list-empty', 5);
                const namespaceEmpty = Utils.namespaceFromProjectName(emptyProjectName);

                sharedResources.projects.emptyProject = await projectResourcesUtils.createProject(
                    namespaceEmpty,
                    {
                        name: emptyProjectName,
                        deservedGpus: 0,
                        maxAllowedGpus: 0,
                        departmentId: sharedResources.departments.dept01.id // Associate with department
                    }
                );
                logger.info(
                    `Created empty project: ${emptyProjectName} (ID: ${sharedResources.projects.emptyProject.id})`
                );

                // Set default project context for CLI
                runaiCliV2.projectSet(projectName);
                logger.info(`Set CLI project context to: ${projectName}`);
            });

            await test.step('Create cluster-scoped compute resources', async function () {
                const compute1Name = objectGenerator.generateShortNameByPrefix('test-cluster-compute-1', 5);
                sharedResources.computeResources.clusterCompute1 = await computeResourceUtils.create({
                    meta: {
                        name: compute1Name,
                        scope: Scope.Cluster,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 4,
                        cpuMemoryRequest: '8Gi',
                        gpuDevicesRequest: 2
                    }
                });
                logger.info(`Created cluster compute 1: ${compute1Name}`);

                const compute2Name = objectGenerator.generateShortNameByPrefix('test-cluster-compute-2', 5);
                sharedResources.computeResources.clusterCompute2 = await computeResourceUtils.create({
                    meta: {
                        name: compute2Name,
                        scope: Scope.Cluster,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 2,
                        cpuMemoryRequest: '4Gi',
                        gpuDevicesRequest: 1
                    }
                });
                logger.info(`Created cluster compute 2: ${compute2Name}`);
            });

            await test.step('Create department-scoped compute resources', async function () {
                const deptCompute1Name = objectGenerator.generateShortNameByPrefix('test-dept-compute-1', 5);
                sharedResources.computeResources.deptCompute1 = await computeResourceUtils.create({
                    meta: {
                        name: deptCompute1Name,
                        scope: Scope.Department,
                        departmentId: sharedResources.departments.dept01.id
                        // Note: clusterId should NOT be provided for department-scoped assets
                    },
                    spec: {
                        cpuCoreRequest: 2,
                        cpuMemoryRequest: '4Gi',
                        gpuDevicesRequest: 1
                    }
                });
                logger.info(`Created department compute 1: ${deptCompute1Name}`);

                const deptCompute2Name = objectGenerator.generateShortNameByPrefix('test-dept-compute-2', 5);
                sharedResources.computeResources.deptCompute2 = await computeResourceUtils.create({
                    meta: {
                        name: deptCompute2Name,
                        scope: Scope.Department,
                        departmentId: sharedResources.departments.dept01.id
                        // Note: clusterId should NOT be provided for department-scoped assets
                    },
                    spec: {
                        cpuCoreRequest: 1,
                        cpuMemoryRequest: '2Gi',
                        gpuDevicesRequest: 0
                    }
                });
                logger.info(`Created department compute 2: ${deptCompute2Name}`);
            });

            await test.step('Create project-scoped compute resources (default)', async function () {
                const projectId = sharedResources.projects.default!.id;

                const compute1Name = objectGenerator.generateShortNameByPrefix('test-project-compute-1', 5);
                sharedResources.computeResources.projectCompute1 = await computeResourceUtils.create({
                    meta: {
                        name: compute1Name,
                        scope: Scope.Project,
                        projectId: projectId,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 2,
                        cpuMemoryRequest: '4Gi',
                        gpuDevicesRequest: 1
                    }
                });
                logger.info(`Created project compute 1: ${compute1Name}`);

                const compute2Name = objectGenerator.generateShortNameByPrefix('test-project-compute-2', 5);
                sharedResources.computeResources.projectCompute2 = await computeResourceUtils.create({
                    meta: {
                        name: compute2Name,
                        scope: Scope.Project,
                        projectId: projectId,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 4,
                        cpuMemoryRequest: '8Gi',
                        gpuDevicesRequest: 0
                    }
                });
                logger.info(`Created project compute 2: ${compute2Name}`);

                const compute3Name = objectGenerator.generateShortNameByPrefix('test-project-compute-3', 5);
                sharedResources.computeResources.projectCompute3 = await computeResourceUtils.create({
                    meta: {
                        name: compute3Name,
                        scope: Scope.Project,
                        projectId: projectId,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 1,
                        cpuMemoryRequest: '2Gi',
                        gpuDevicesRequest: 2
                    }
                });
                logger.info(`Created project compute 3: ${compute3Name}`);
            });

            await test.step('Create project-scoped compute resources (project-a)', async function () {
                const projectId = sharedResources.projects.projectA!.id;

                const computeA1Name = objectGenerator.generateShortNameByPrefix('test-compute-proj-a-1', 5);
                sharedResources.computeResources.projectACompute1 = await computeResourceUtils.create({
                    meta: {
                        name: computeA1Name,
                        scope: Scope.Project,
                        projectId: projectId,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 1,
                        cpuMemoryRequest: '2Gi',
                        gpuDevicesRequest: 1
                    }
                });
                logger.info(`Created project-a compute 1: ${computeA1Name}`);

                const computeA2Name = objectGenerator.generateShortNameByPrefix('test-compute-proj-a-2', 5);
                sharedResources.computeResources.projectACompute2 = await computeResourceUtils.create({
                    meta: {
                        name: computeA2Name,
                        scope: Scope.Project,
                        projectId: projectId,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 2,
                        cpuMemoryRequest: '4Gi',
                        gpuDevicesRequest: 0
                    }
                });
                logger.info(`Created project-a compute 2: ${computeA2Name}`);
            });

            logger.info('=== All shared resources created (9 compute resources) ===');
        });

        test.afterAll(async function () {
            logger.info('=== Cleaning up shared resources ===');

            await test.step('Delete compute resources', async function () {
                const computeIds = Object.values(sharedResources.computeResources)
                    .filter(Boolean)
                    .map((compute: any) => compute.meta.id);

                for (const id of computeIds) {
                    try {
                        await computeResourceUtils.delete(id);
                        logger.info(`Deleted compute resource: ${id}`);
                    } catch (error) {
                        logger.error(`Error deleting compute resource ${id}:`, error);
                    }
                }
            });

            await test.step('Delete projects first (before department)', async function () {
                const projects = Object.values(sharedResources.projects).filter(Boolean);
                for (const project of projects) {
                    if ((project as Project).id) {
                        await operations.deleteProjectAndNamespace((project as Project).id);
                        logger.info(`Deleted project: ${(project as Project).id}`);
                    }
                }
            });

            await test.step('Delete department', async function () {
                if (sharedResources.departments.dept01?.id) {
                    const orgUnitUtils = new OrgUnitUtils();
                    await orgUnitUtils.deleteDepartment(sharedResources.departments.dept01.id.toString());
                    logger.info(`Deleted department: ${sharedResources.departments.dept01.id}`);
                }
            });

            logger.info('=== All shared resources cleaned up ===');
        });

        test('T5597664: List compute resources with default output', async function () {
            await test.step('List compute resources', function () {
                logger.info(
                    'Test T5597664: Listing compute resources with default output (using shared resources)'
                );
                const stdout = runaiCliV2.list('compute', {}, false).stdout;

                logger.info('=== List Output Start ===');
                logger.info(stdout);
                logger.info('=== List Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI list command executed successfully');
            });

            await test.step('Verify output contains project compute resources', function () {
                const stdout = runaiCliV2.list('compute', {}, false).stdout;

                // Should contain project-scoped resources
                expect(stdout).toContain('test-project-compute');

                logger.info('Verified: Output contains project compute resources');
            });
        });

        test('T5597665: List all scopes with -A flag', async function () {
            await test.step('List with -A flag', function () {
                logger.info(
                    'Test T5597665: Listing all compute resources with -A flag (using shared resources)'
                );
                const stdout = runaiCliV2.list('compute', { [listFlags.allScopes]: '' } as any, false).stdout;

                logger.info('=== List All Output Start ===');
                logger.info(stdout);
                logger.info('=== List All Output End ===');

                // Should contain resources from all scopes
                expect(stdout).toContain('test-cluster-compute');
                expect(stdout).toContain('test-dept-compute');
                expect(stdout).toContain('test-project-compute');

                logger.info('Verified: Output contains resources from all scopes');
            });
        });

        test('T5597666: List compute resources in specific project with -p flag', async function () {
            const projectAName = sharedResources.projects.projectA!.name;

            await test.step('List with -p flag', function () {
                logger.info(
                    'Test T5597666: Listing compute in project-a using -p flag (using shared resources)'
                );
                const stdout = runaiCliV2.list(
                    'compute',
                    { [listFlags.project]: projectAName } as any,
                    false
                ).stdout;

                logger.info('=== List Project-A Output Start ===');
                logger.info(stdout);
                logger.info('=== List Project-A Output End ===');

                // Should contain project-a resources
                expect(stdout).toContain('test-compute-proj-a');
                // Should NOT contain default project-specific resources
                expect(stdout).not.toContain('test-project-compute-1');

                logger.info('Verified: Output contains only project-a resources (and higher scopes)');
            });
        });

        test('T5597667: List compute resources with JSON output', async function () {
            await test.step('List with JSON output', function () {
                logger.info('Test T5597667: Listing compute with JSON output (using shared resources)');
                const stdout = runaiCliV2.list('compute', { [listFlags.json]: '' } as any, false).stdout;

                logger.info('=== JSON Output Start ===');
                logger.info(stdout);
                logger.info('=== JSON Output End ===');

                // Parse JSON to verify it's valid
                const jsonData = JSON.parse(stdout);

                expect(jsonData.computes || jsonData).toBeDefined();
                expect(Array.isArray(jsonData.computes || jsonData)).toBe(true);

                logger.info('Verified: Output is valid JSON format');
            });
        });

        test('T5597668: List compute resources with YAML output', async function () {
            await test.step('List with YAML output', function () {
                logger.info('Test T5597668: Listing compute with YAML output (using shared resources)');
                const stdout = runaiCliV2.list('compute', { [listFlags.yaml]: '' } as any, false).stdout;

                logger.info('=== YAML Output Start ===');
                logger.info(stdout);
                logger.info('=== YAML Output End ===');

                // Verify YAML structure
                expect(stdout).toContain('computes:');
                expect(stdout).toContain('name:');
                expect(stdout).toContain('scope:'); // Field is "scope", not "scopeType"

                logger.info('Verified: Output is valid YAML format');
            });
        });

        test('T5597669: List compute resources with table output (explicit)', async function () {
            await test.step('Compare table output with default', function () {
                logger.info('Test T5597669: Comparing table and default output (using shared resources)');

                const stdoutTable = runaiCliV2.list(
                    'compute',
                    { [listFlags.table]: '' } as any,
                    false
                ).stdout;
                const stdoutDefault = runaiCliV2.list('compute', {}, false).stdout;

                // Both outputs should be identical
                expect(stdoutTable).toEqual(stdoutDefault);

                logger.info('Verified: Table output matches default - table is the default format');
            });
        });

        test('T5597670: List compute resources with --no-headers', async function () {
            await test.step('List with --no-headers', function () {
                logger.info('Test T5597670: Listing compute with --no-headers (using shared resources)');

                const stdoutNoHeaders = runaiCliV2.list(
                    'compute',
                    { [listFlags.noHeaders]: '' } as any,
                    false
                ).stdout;
                const stdoutDefault = runaiCliV2.list('compute', {}, false).stdout;

                logger.info('=== No-Headers Output Start ===');
                logger.info(stdoutNoHeaders);
                logger.info('=== No-Headers Output End ===');

                // No-headers should be shorter (no header row)
                expect(stdoutNoHeaders.split('\n').length).toBeLessThan(stdoutDefault.split('\n').length);

                logger.info('Verified: No-headers output displays without header row');
            });
        });

        test('T5597671: List compute resources with custom columns', async function () {
            await test.step('List with custom columns', function () {
                logger.info('Test T5597671: Listing compute with custom columns (using shared resources)');

                const stdout = runaiCliV2.list(
                    'compute',
                    {
                        [listFlags.columns]: 'name,scope'
                    } as any,
                    false
                ).stdout;

                logger.info('=== Custom Columns Output Start ===');
                logger.info(stdout);
                logger.info('=== Custom Columns Output End ===');

                expect(stdout).toContain('Name');
                expect(stdout).toContain('Scope');

                logger.info('Verified: Custom columns displayed correctly');
            });
        });

        test('T5597672: List compute resources when empty', async function () {
            const emptyProjectName = sharedResources.projects.emptyProject!.name;

            await test.step('List in empty project', function () {
                logger.info('Test T5597672: Listing compute in empty project (using shared empty project)');

                const stdout = runaiCliV2.list(
                    'compute',
                    {
                        [listFlags.project]: emptyProjectName
                    } as any,
                    false
                ).stdout;

                logger.info('=== Empty List Output Start ===');
                logger.info(stdout);
                logger.info('=== Empty List Output End ===');

                expect(stdout).toBeTruthy();

                logger.info('Verified: Empty list returns successfully');
            });
        });

        test('T5597673: List includes cluster-scoped compute resources', async function () {
            await test.step('Verify cluster-scoped compute visible', function () {
                logger.info(
                    'Test T5597673: Verifying cluster-scoped compute resources visible (using shared resources)'
                );

                const stdout = runaiCliV2.list('compute', {}, false).stdout;

                // Should contain cluster-scoped resources
                expect(stdout).toContain('test-cluster-compute');
                expect(stdout).toMatch(/cluster/i);

                logger.info('Verified: Cluster-scoped compute resources visible');
            });
        });

        test('T5597674: List includes department-scoped compute resources', async function () {
            await test.step('Verify department-scoped compute visible', function () {
                logger.info(
                    'Test T5597674: Verifying department-scoped compute resources visible (using shared resources)'
                );

                const stdout = runaiCliV2.list('compute', {}, false).stdout;

                // Should contain department-scoped resources
                expect(stdout).toContain('test-dept-compute');
                expect(stdout).toMatch(/department/i);

                logger.info('Verified: Department-scoped compute resources visible');
            });
        });

        test('T5597675: List includes project-scoped compute resources', async function () {
            await test.step('Verify project-scoped compute visible', function () {
                logger.info(
                    'Test T5597675: Verifying project-scoped compute resources visible (using shared resources)'
                );

                const stdout = runaiCliV2.list('compute', {}, false).stdout;

                // Should contain project-scoped resources
                expect(stdout).toContain('test-project-compute');
                expect(stdout).toMatch(/project/i);

                logger.info('Verified: Project-scoped compute resources visible');
            });
        });

        test('T5597676: List compute resources with default project (no -p flag)', async function () {
            const projectName = sharedResources.projects.default!.name;

            await test.step('Compare default context with explicit -p flag', function () {
                logger.info(
                    'Test T5597676: Comparing default context with explicit -p flag (using shared resources)'
                );

                // Ensure default project is set
                runaiCliV2.projectSet(projectName);

                const stdoutDefault = runaiCliV2.list('compute', {}, false).stdout;
                const stdoutExplicit = runaiCliV2.list(
                    'compute',
                    {
                        [listFlags.project]: projectName
                    } as any,
                    false
                ).stdout;

                // Both outputs should be identical
                expect(stdoutDefault).toEqual(stdoutExplicit);

                logger.info('Verified: Default project context works without -p flag');
            });
        });
    });
});
