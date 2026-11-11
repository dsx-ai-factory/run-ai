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
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { describeFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI Describe Compute - Negative Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        // Error message patterns
        const ERROR_COMPUTE_NOT_FOUND = /Failed to describe compute.*not found/i;
        const ERROR_UNSUPPORTED_OUTPUT_FORMAT = /unsupported output format/i;
        const ERROR_MISSING_ARGUMENT = /accepts 1 arg\(s\), received 0/i;

        const sharedResources = {
            projects: {
                projectA: null as Project | null,
                projectB: null as Project | null
            },
            computeResources: {
                validCompute: null as any,
                isolatedCompute: null as any
            }
        };

        test.beforeAll(async function () {
            logger.info('=== Creating shared resources for Describe Negative tests ===');

            await test.step('Create projects', async function () {
                const projectAName = objectGenerator.generateShortNameByPrefix('proj-desc-neg-a', 5);
                const namespaceA = Utils.namespaceFromProjectName(projectAName);

                sharedResources.projects.projectA = await projectResourcesUtils.createProject(namespaceA, {
                    name: projectAName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(
                    `Created project-a: ${projectAName} (ID: ${sharedResources.projects.projectA.id})`
                );

                const projectBName = objectGenerator.generateShortNameByPrefix('proj-desc-neg-b', 5);
                const namespaceB = Utils.namespaceFromProjectName(projectBName);

                sharedResources.projects.projectB = await projectResourcesUtils.createProject(namespaceB, {
                    name: projectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(
                    `Created project-b: ${projectBName} (ID: ${sharedResources.projects.projectB.id})`
                );

                // Set project-a as default context
                runaiCliV2.projectSet(projectAName);
                logger.info(`Set CLI project context to: ${projectAName}`);
            });

            await test.step('Create compute resources', async function () {
                const validComputeName = objectGenerator.generateShortNameByPrefix('test-neg-compute', 5);

                const validComputeAsset = await computeResourceUtils.create({
                    meta: {
                        name: validComputeName,
                        scope: Scope.Project,
                        projectId: sharedResources.projects.projectA!.id,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 1,
                        cpuMemoryRequest: '2Gi',
                        gpuDevicesRequest: 0
                    }
                });

                sharedResources.computeResources.validCompute = validComputeAsset;
                logger.info(
                    `Created valid compute for negative tests: ${validComputeName} (ID: ${validComputeAsset.meta.id})`
                );

                const isolatedComputeName = objectGenerator.generateShortNameByPrefix(
                    'test-isolated-compute',
                    5
                );

                const isolatedComputeAsset = await computeResourceUtils.create({
                    meta: {
                        name: isolatedComputeName,
                        scope: Scope.Project,
                        projectId: sharedResources.projects.projectB!.id,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 1,
                        cpuMemoryRequest: '2Gi',
                        gpuDevicesRequest: 0
                    }
                });

                sharedResources.computeResources.isolatedCompute = isolatedComputeAsset;
                logger.info(
                    `Created isolated compute in project-b: ${isolatedComputeName} (ID: ${isolatedComputeAsset.meta.id})`
                );
            });

            logger.info('=== Shared resources for negative tests created ===');
        });

        test.afterAll(async function () {
            logger.info('=== Cleaning up shared resources ===');

            await test.step('Delete compute resources', async function () {
                const computeIds = [
                    sharedResources.computeResources.validCompute?.meta.id,
                    sharedResources.computeResources.isolatedCompute?.meta.id
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

            await test.step('Delete projects', async function () {
                if (sharedResources.projects.projectA?.id) {
                    await operations.deleteProjectAndNamespace(sharedResources.projects.projectA.id);
                    logger.info('Deleted project-a');
                }
                if (sharedResources.projects.projectB?.id) {
                    await operations.deleteProjectAndNamespace(sharedResources.projects.projectB.id);
                    logger.info('Deleted project-b');
                }
            });

            logger.info('=== All shared resources cleaned up ===');
        });

        test('T5597657: Describe non-existent compute resource', async function () {
            await test.step('Attempt to describe non-existent compute', function () {
                logger.info('Test T5597657: Attempting to describe non-existent compute resource');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    runaiCliV2.computeDescribe('non-existent-compute-12345');
                } catch (error: any) {
                    exitCode = error.exitCode;
                    errorOutput = error.stderr || error.stdout || error.message || '';
                }

                logger.info('=== Error Output Start ===');
                logger.info(`Exit Code: ${exitCode}`);
                logger.info(`Error Output: ${errorOutput}`);
                logger.info('=== Error Output End ===');

                expect(exitCode).not.toBe(0);
                expect(errorOutput).toMatch(ERROR_COMPUTE_NOT_FOUND);

                logger.info('Verified: Non-existent compute returns error');
            });
        });

        test('T5597658: Describe compute resource with invalid output format', async function () {
            const computeName = sharedResources.computeResources.validCompute.meta.name;

            await test.step('Attempt to describe with invalid format', function () {
                logger.info(
                    `Test T5597658: Attempting invalid output format (using shared resource: ${computeName})`
                );

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    runaiCliV2.computeDescribe(computeName, { '--output': 'xml' } as any);
                } catch (error: any) {
                    exitCode = error.exitCode;
                    errorOutput = error.stderr || error.stdout || error.message || '';
                }

                logger.info('=== Error Output Start ===');
                logger.info(`Exit Code: ${exitCode}`);
                logger.info(`Error Output: ${errorOutput}`);
                logger.info('=== Error Output End ===');

                expect(exitCode).not.toBe(0);
                expect(errorOutput).toMatch(ERROR_UNSUPPORTED_OUTPUT_FORMAT);

                logger.info('Verified: Invalid output format returns error');
            });
        });

        test('T5597659: Describe compute resource without name argument', async function () {
            await test.step('Attempt to describe without name', function () {
                logger.info('Test T5597659: Attempting describe without name argument');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    runaiCliV2.computeDescribe('');
                } catch (error: any) {
                    exitCode = error.exitCode;
                    errorOutput = error.stderr || error.stdout || error.message || '';
                }

                logger.info('=== Error Output Start ===');
                logger.info(`Exit Code: ${exitCode}`);
                logger.info(`Error Output: ${errorOutput}`);
                logger.info('=== Error Output End ===');

                expect(exitCode).not.toBe(0);
                expect(errorOutput).toMatch(ERROR_MISSING_ARGUMENT);

                logger.info('Verified: Missing name argument returns error');
            });
        });

        test('T5597660: Describe compute resource with empty name', async function () {
            await test.step('Attempt to describe with empty name', function () {
                logger.info('Test T5597660: Attempting describe with empty name');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    runaiCliV2.computeDescribe('');
                } catch (error: any) {
                    exitCode = error.exitCode;
                    errorOutput = error.stderr || error.stdout || error.message || '';
                }

                logger.info('=== Error Output Start ===');
                logger.info(`Exit Code: ${exitCode}`);
                logger.info(`Error Output: ${errorOutput}`);
                logger.info('=== Error Output End ===');

                expect(exitCode).not.toBe(0);
                expect(errorOutput).toMatch(ERROR_MISSING_ARGUMENT);

                logger.info('Verified: Empty name returns error');
            });
        });

        test('T5597661: Describe compute resource from different project without access', async function () {
            const isolatedComputeName = sharedResources.computeResources.isolatedCompute.meta.name;
            const projectAName = sharedResources.projects.projectA!.name;

            await test.step('Attempt cross-project access', function () {
                logger.info(
                    'Test T5597661: project-a attempting to access project-b compute (using shared resource)'
                );

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    runaiCliV2.computeDescribe(isolatedComputeName, {
                        [describeFlags.project]: projectAName
                    } as any);
                } catch (error: any) {
                    exitCode = error.exitCode;
                    errorOutput = error.stderr || error.stdout || error.message || '';
                }

                logger.info('=== Error Output Start ===');
                logger.info(`Exit Code: ${exitCode}`);
                logger.info(`Error Output: ${errorOutput}`);
                logger.info('=== Error Output End ===');

                expect(exitCode).not.toBe(0);
                expect(errorOutput).toMatch(ERROR_COMPUTE_NOT_FOUND);

                logger.info('Verified: Cross-project access denied');
            });
        });

        test('T5597662: Describe deleted compute resource', async function () {
            let tempCompute: any = null;
            const projectAId = sharedResources.projects.projectA!.id;

            await test.step('Create temporary compute resource', async function () {
                const tempComputeName = objectGenerator.generateShortNameByPrefix('temp-delete-test', 5);

                tempCompute = await computeResourceUtils.create({
                    meta: {
                        name: tempComputeName,
                        scope: Scope.Project,
                        projectId: projectAId,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 0.5,
                        cpuMemoryRequest: '1Gi',
                        gpuDevicesRequest: 0
                    }
                });

                logger.info(`Created temporary compute: ${tempComputeName} (ID: ${tempCompute.meta.id})`);
            });

            await test.step('Verify compute can be described', function () {
                const stdout = runaiCliV2.computeDescribe(tempCompute.meta.name);
                expect(stdout).toContain(tempCompute.meta.name);
                logger.info('Verified: Compute exists and can be described');
            });

            await test.step('Delete the compute resource', async function () {
                await computeResourceUtils.delete(tempCompute.meta.id);
                logger.info(`Deleted compute: ${tempCompute.meta.id}`);
            });

            await test.step('Verify deleted compute cannot be described', function () {
                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    runaiCliV2.computeDescribe(tempCompute.meta.name);
                } catch (error: any) {
                    exitCode = error.exitCode;
                    errorOutput = error.stderr || error.stdout || error.message || '';
                }

                logger.info('=== Error Output Start ===');
                logger.info(`Exit Code: ${exitCode}`);
                logger.info(`Error Output: ${errorOutput}`);
                logger.info('=== Error Output End ===');

                expect(exitCode).not.toBe(0);
                expect(errorOutput).toMatch(ERROR_COMPUTE_NOT_FOUND);

                logger.info('Verified: Deleted compute returns not found error');
            });
        });
    });
});
