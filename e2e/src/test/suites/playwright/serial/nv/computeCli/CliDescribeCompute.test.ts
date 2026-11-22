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
import { describeFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI Describe Compute - Positive Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        const sharedResources = {
            projects: {
                default: null as Project | null,
                projectA: null as Project | null
            },
            departments: {
                dept01: null as any
            },
            computeResources: {
                projectCompute: null as any,
                clusterCompute: null as any,
                deptCompute: null as any,
                defaultCompute: null as any
            }
        };

        test.beforeAll(async function () {
            logger.info('=== Creating shared resources for Describe Positive tests ===');

            await test.step('Create department first', async function () {
                const deptName = objectGenerator.generateShortNameByPrefix('dept-desc-01', 5);
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
                const projectName = objectGenerator.generateShortNameByPrefix('proj-desc-default', 5);
                const namespace = Utils.namespaceFromProjectName(projectName);

                // Create default project UNDER the department so it can access dept-scoped resources
                sharedResources.projects.default = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0,
                    departmentId: sharedResources.departments.dept01.id // Associate with department
                });
                logger.info(
                    `Created default project: ${projectName} (ID: ${sharedResources.projects.default.id})`
                );

                const projectAName = objectGenerator.generateShortNameByPrefix('proj-desc-a', 5);
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

                // Set default project context for CLI
                runaiCliV2.projectSet(projectName);
                logger.info(`Set CLI project context to: ${projectName}`);
            });

            await test.step('Create cluster-scoped compute resource', async function () {
                const computeName = objectGenerator.generateShortNameByPrefix('test-cluster-compute', 5);

                const computeAsset = await computeResourceUtils.create({
                    meta: {
                        name: computeName,
                        scope: Scope.Cluster,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 4,
                        cpuMemoryRequest: '8Gi',
                        gpuDevicesRequest: 2
                    }
                });

                sharedResources.computeResources.clusterCompute = computeAsset;
                logger.info(`Created cluster-scoped compute: ${computeName} (ID: ${computeAsset.meta.id})`);
            });

            await test.step('Create department-scoped compute resource', async function () {
                const computeName = objectGenerator.generateShortNameByPrefix('test-dept-compute', 5);

                const computeAsset = await computeResourceUtils.create({
                    meta: {
                        name: computeName,
                        scope: Scope.Department,
                        departmentId: sharedResources.departments.dept01.id
                    },
                    spec: {
                        cpuCoreRequest: 2,
                        cpuMemoryRequest: '4Gi',
                        gpuDevicesRequest: 1
                    }
                });

                sharedResources.computeResources.deptCompute = computeAsset;
                logger.info(
                    `Created department-scoped compute: ${computeName} (ID: ${computeAsset.meta.id})`
                );
            });

            await test.step('Create project-scoped compute resources', async function () {
                const projectComputeName = objectGenerator.generateShortNameByPrefix(
                    'test-project-compute',
                    5
                );

                const projectComputeAsset = await computeResourceUtils.create({
                    meta: {
                        name: projectComputeName,
                        scope: Scope.Project,
                        projectId: sharedResources.projects.default!.id,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 2,
                        cpuMemoryRequest: '4Gi',
                        gpuDevicesRequest: 1
                    }
                });

                sharedResources.computeResources.projectCompute = projectComputeAsset;
                logger.info(
                    `Created project-scoped compute: ${projectComputeName} (ID: ${projectComputeAsset.meta.id})`
                );

                const defaultComputeName = objectGenerator.generateShortNameByPrefix(
                    'test-default-compute',
                    5
                );

                const defaultComputeAsset = await computeResourceUtils.create({
                    meta: {
                        name: defaultComputeName,
                        scope: Scope.Project,
                        projectId: sharedResources.projects.default!.id,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 1,
                        cpuMemoryRequest: '2Gi',
                        gpuDevicesRequest: 0
                    }
                });

                sharedResources.computeResources.defaultCompute = defaultComputeAsset;
                logger.info(
                    `Created default compute: ${defaultComputeName} (ID: ${defaultComputeAsset.meta.id})`
                );
            });

            logger.info('=== All shared resources created successfully ===');
        });

        test.afterAll(async function () {
            logger.info('=== Cleaning up shared resources ===');

            await test.step('Delete compute resources', async function () {
                const computeIds = [
                    sharedResources.computeResources.projectCompute?.meta.id,
                    sharedResources.computeResources.clusterCompute?.meta.id,
                    sharedResources.computeResources.deptCompute?.meta.id,
                    sharedResources.computeResources.defaultCompute?.meta.id
                ];

                for (const id of computeIds) {
                    if (id) {
                        try {
                            await computeResourceUtils.delete(id);
                            logger.info(`Deleted compute resource: ${id}`);
                        } catch (error) {
                            logger.error(`Error deleting compute resource ${id}:`, error);
                        }
                    }
                }
            });

            await test.step('Delete projects first (before department)', async function () {
                if (sharedResources.projects.default?.id) {
                    await operations.deleteProjectAndNamespace(sharedResources.projects.default.id);
                    logger.info('Deleted default project');
                }
                if (sharedResources.projects.projectA?.id) {
                    await operations.deleteProjectAndNamespace(sharedResources.projects.projectA.id);
                    logger.info('Deleted project-a');
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

        test('T5597649: Describe compute resource with default output', async function () {
            const computeName = sharedResources.computeResources.projectCompute.meta.name;

            await test.step('Describe compute resource', function () {
                logger.info(
                    `Test T5597649: Describing compute resource: ${computeName} (using shared resource)`
                );
                const stdout = runaiCliV2.computeDescribe(computeName);

                logger.info('=== CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout.length).toBeGreaterThan(0);
                logger.info('CLI describe command executed successfully');
            });

            await test.step('Verify output contains compute resource details', function () {
                const stdout = runaiCliV2.computeDescribe(computeName);

                // Verify key fields
                expect(stdout).toContain(computeName);
                expect(stdout).toContain('Project'); // Scope
                expect(stdout).toContain('CPU Cores Request'); // CPU specs (note: plural "Cores")
                expect(stdout).toContain('GPU Devices Request'); // GPU specs

                logger.info('Verified: Output contains expected compute resource information');
            });
        });

        test('T5597650: Describe compute resource in specific project using -p flag', async function () {
            const computeName = sharedResources.computeResources.projectCompute.meta.name;
            const projectName = sharedResources.projects.default!.name;

            await test.step('Describe compute with -p flag', function () {
                logger.info(
                    `Test T5597650: Describing compute in project ${projectName} (using shared resource)`
                );
                const stdout = runaiCliV2.computeDescribe(computeName, {
                    [describeFlags.project]: projectName
                } as any);

                logger.info('=== CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== CLI Output End ===');

                expect(stdout).toBeTruthy();
                expect(stdout).toContain(computeName);
                expect(stdout).toContain('Project');

                logger.info('Verified: Compute resource described with -p flag');
            });
        });

        test('T5597651: Describe compute resource with JSON output format', async function () {
            const computeName = sharedResources.computeResources.projectCompute.meta.name;

            await test.step('Describe compute with JSON output', function () {
                logger.info('Test T5597651: Describing compute with JSON output (using shared resource)');
                const stdout = runaiCliV2.computeDescribe(computeName, { [describeFlags.output]: 'json' });

                logger.info('=== JSON Output Start ===');
                logger.info(stdout);
                logger.info('=== JSON Output End ===');

                // Parse JSON to verify it's valid
                const jsonData = JSON.parse(stdout);

                expect(jsonData).toBeDefined();
                expect(jsonData.meta).toBeDefined();
                expect(jsonData.meta.name).toBe(computeName);
                expect(jsonData.meta.scope).toBe('project');

                logger.info('Verified: Output is valid JSON format');
            });
        });

        test('T5597652: Describe compute resource with YAML output format', async function () {
            const computeName = sharedResources.computeResources.projectCompute.meta.name;

            await test.step('Describe compute with YAML output', function () {
                logger.info('Test T5597652: Describing compute with YAML output (using shared resource)');
                const stdout = runaiCliV2.computeDescribe(computeName, { [describeFlags.output]: 'yaml' });

                logger.info('=== YAML Output Start ===');
                logger.info(stdout);
                logger.info('=== YAML Output End ===');

                // Verify YAML structure
                expect(stdout).toContain('name:');
                expect(stdout).toContain('scope:');
                expect(stdout).toContain(computeName);
                expect(stdout).toContain('cpuCoreRequest:');
                expect(stdout).toContain('gpuDevicesRequest:');

                logger.info('Verified: Output is valid YAML format');
            });
        });

        test('T5597653: Describe compute resource with table output format (explicit)', async function () {
            const computeName = sharedResources.computeResources.projectCompute.meta.name;

            await test.step('Compare table output with default output', function () {
                logger.info('Test T5597653: Comparing table and default output (using shared resource)');

                const stdoutTable = runaiCliV2.computeDescribe(computeName, {
                    [describeFlags.output]: 'table'
                });
                const stdoutDefault = runaiCliV2.computeDescribe(computeName);

                // Both outputs should be identical
                expect(stdoutTable).toEqual(stdoutDefault);

                logger.info('Verified: Table output matches default output - table is the default format');
            });
        });

        test('T5597654: Describe cluster-scoped compute resource from project context', async function () {
            const computeName = sharedResources.computeResources.clusterCompute.meta.name;
            const projectName = sharedResources.projects.default!.name;

            await test.step('Describe cluster-scoped compute from project', function () {
                logger.info(
                    'Test T5597654: Describing cluster-scoped compute from project (using shared resource)'
                );
                const stdout = runaiCliV2.computeDescribe(computeName, {
                    [describeFlags.project]: projectName
                } as any);

                logger.info('=== CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== CLI Output End ===');

                expect(stdout).toContain(computeName);
                expect(stdout).toContain('Cluster'); // Scope should be cluster

                logger.info('Verified: Cluster-scoped compute accessible from project context');
            });
        });

        test('T5597655: Describe department-scoped compute resource from project', async function () {
            const computeName = sharedResources.computeResources.deptCompute.meta.name;
            const projectName = sharedResources.projects.default!.name;

            await test.step('Describe department-scoped compute', function () {
                logger.info('Test T5597655: Describing department-scoped compute (using shared resource)');
                const stdout = runaiCliV2.computeDescribe(computeName, {
                    [describeFlags.project]: projectName
                } as any);

                logger.info('=== CLI Output Start ===');
                logger.info(stdout);
                logger.info('=== CLI Output End ===');

                expect(stdout).toContain(computeName);
                expect(stdout).toContain('Department'); // Scope should be department

                logger.info('Verified: Department-scoped compute accessible from project');
            });
        });

        test('T5597656: Describe compute resource in default project without -p flag', async function () {
            const computeName = sharedResources.computeResources.defaultCompute.meta.name;
            const projectName = sharedResources.projects.default!.name;

            await test.step('Describe using default project context', function () {
                logger.info(
                    'Test T5597656: Describing compute using default project context (using shared resource)'
                );

                // Ensure default project is set
                runaiCliV2.projectSet(projectName);

                const stdoutDefault = runaiCliV2.computeDescribe(computeName);
                const stdoutExplicit = runaiCliV2.computeDescribe(computeName, {
                    [describeFlags.project]: projectName
                } as any);

                // Both outputs should be identical
                expect(stdoutDefault).toEqual(stdoutExplicit);

                logger.info('Verified: Default project context works without -p flag');
            });
        });
    });
});
