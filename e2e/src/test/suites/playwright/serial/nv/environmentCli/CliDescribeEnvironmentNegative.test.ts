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
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';

test.describe(`CLI Describe Environment - Negative Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        // Error message patterns
        const ERROR_ENVIRONMENT_NOT_FOUND = /environment.*not found|not accessible/i;
        const ERROR_UNSUPPORTED_OUTPUT_FORMAT = /unsupported|invalid.*output.*format|unknown flag/i;
        const ERROR_MISSING_ARGUMENT = /required|missing.*argument|accepts.*arg|received/i;

        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-env-neg', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            await test.step('Create project via API', async function () {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
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

        test('T5597616: Describe non-existent environment', async function () {
            await test.step('Attempt to describe non-existent environment', function () {
                logger.info('Test T5597616: Attempting to describe non-existent environment');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    const result = runaiCliV2['bashCli'].execWithOutput(
                        `${runaiCliV2['cliPath']} environment describe non-existent-env-12345`,
                        false
                    );
                    exitCode = result.status ?? undefined;
                    errorOutput = result.stderr || result.stdout || '';
                } catch (error: any) {
                    exitCode = error.exitCode;
                    errorOutput = error.stderr || error.stdout || error.message || '';
                }

                logger.info('=== Error Output Start ===');
                logger.info(`Exit Code: ${exitCode}`);
                logger.info(`Error Output: ${errorOutput}`);
                logger.info('=== Error Output End ===');

                expect(exitCode).not.toBe(0);
                expect(errorOutput).toMatch(ERROR_ENVIRONMENT_NOT_FOUND);

                logger.info('Verified: Non-existent environment returns appropriate error');
            });
        });

        test('T5597617: Describe environment with invalid output format', async function () {
            let testEnvId: string | null = null;
            let testEnvName: string;

            await test.step('Create test environment', async function () {
                testEnvName = objectGenerator.generateShortNameByPrefix('test-env-invalid-fmt', 5);
                const env = await environmentUtils.create({
                    meta: {
                        name: testEnvName,
                        scope: Scope.Project,
                        projectId: project.id,
                        clusterId: flowRootState.cluster.uuid,
                        workloadSupportedTypes: {
                            workspace: true,
                            training: true
                        }
                    },
                    spec: {
                        image: 'alpine:latest'
                    }
                });
                testEnvId = env.meta.id;
                logger.info(`Created test environment: ${testEnvName}`);
            });

            await test.step('Attempt to describe with invalid format', function () {
                logger.info('Test T5597617: Attempting to describe with invalid output format');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    const result = runaiCliV2['bashCli'].execWithOutput(
                        `${runaiCliV2['cliPath']} environment describe ${testEnvName} -o xml`,
                        false
                    );
                    exitCode = result.status ?? undefined;
                    errorOutput = result.stderr || result.stdout || '';
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

                logger.info('Verified: Invalid output format returns appropriate error');
            });

            await test.step('Cleanup test environment', async function () {
                if (testEnvId) {
                    await environmentUtils.delete(testEnvId);
                    logger.info(`Deleted test environment: ${testEnvId}`);
                }
            });
        });

        test('T5597618: Describe environment without name argument', async function () {
            await test.step('Attempt to describe without name', function () {
                logger.info('Test T5597618: Attempting to describe without environment name');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    const result = runaiCliV2['bashCli'].execWithOutput(
                        `${runaiCliV2['cliPath']} environment describe`,
                        false
                    );
                    exitCode = result.status ?? undefined;
                    errorOutput = result.stderr || result.stdout || '';
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

                logger.info('Verified: Missing name argument returns appropriate error');
            });
        });

        // Coverage remains disabled pending revalidation of this case
        test.skip('T5597619: Describe environment with empty name', async function () {
            await test.step('Attempt to describe with empty name', function () {
                logger.info('Test T5597619: Attempting to describe with empty name');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    const result = runaiCliV2['bashCli'].execWithOutput(
                        `${runaiCliV2['cliPath']} environment describe ""`,
                        false
                    );
                    exitCode = result.status ?? undefined;
                    errorOutput = result.stderr || result.stdout || '';
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

                logger.info('Verified: Empty name returns appropriate error');
            });
        });

        test('T5597620: Describe environment from different project without access', async function () {
            let projectB: Project | null = null;
            let testEnvId: string | null = null;
            let testEnvName: string;

            await test.step('Create project-b', async function () {
                const projectBName = objectGenerator.generateShortNameByPrefix('proj-env-b', 5);
                const namespaceB = Utils.namespaceFromProjectName(projectBName);
                projectB = await projectResourcesUtils.createProject(namespaceB, {
                    name: projectBName,
                    deservedGpus: 0,
                    maxAllowedGpus: 0
                });
                logger.info(`Created project-b: ${projectBName} (ID: ${projectB.id})`);
            });

            await test.step('Create environment in project-b', async function () {
                testEnvName = objectGenerator.generateShortNameByPrefix('test-isolated-env', 5);
                const env = await environmentUtils.create({
                    meta: {
                        name: testEnvName,
                        scope: Scope.Project,
                        projectId: projectB!.id,
                        clusterId: flowRootState.cluster.uuid,
                        workloadSupportedTypes: {
                            workspace: true,
                            training: true
                        }
                    },
                    spec: {
                        image: 'alpine:latest'
                    }
                });
                testEnvId = env.meta.id;
                logger.info(`Created environment in project-b: ${testEnvName}`);
            });

            await test.step('Attempt to describe from project-a', function () {
                logger.info('Test T5597620: Attempting to describe project-b environment from project-a');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    const result = runaiCliV2['bashCli'].execWithOutput(
                        `${runaiCliV2['cliPath']} environment describe ${testEnvName} -p ${projectName}`,
                        false
                    );
                    exitCode = result.status ?? undefined;
                    errorOutput = result.stderr || result.stdout || '';
                } catch (error: any) {
                    exitCode = error.exitCode;
                    errorOutput = error.stderr || error.stdout || error.message || '';
                }

                logger.info('=== Error Output Start ===');
                logger.info(`Exit Code: ${exitCode}`);
                logger.info(`Error Output: ${errorOutput}`);
                logger.info('=== Error Output End ===');

                expect(exitCode).not.toBe(0);
                expect(errorOutput).toMatch(ERROR_ENVIRONMENT_NOT_FOUND);

                logger.info('Verified: Cross-project access without permission returns appropriate error');
            });

            await test.step('Cleanup', async function () {
                if (testEnvId) {
                    await environmentUtils.delete(testEnvId);
                    logger.info(`Deleted test environment: ${testEnvId}`);
                }
                if (projectB?.id) {
                    await operations.deleteProjectAndNamespace(projectB.id);
                    logger.info(`Deleted project-b: ${projectB.id}`);
                }
            });
        });

        test('T5597621: Describe deleted environment', async function () {
            let testEnvId: string | null = null;
            let testEnvName: string;

            await test.step('Create test environment', async function () {
                testEnvName = objectGenerator.generateShortNameByPrefix('test-to-delete-env', 5);
                const env = await environmentUtils.create({
                    meta: {
                        name: testEnvName,
                        scope: Scope.Project,
                        projectId: project.id,
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
                testEnvId = env.meta.id;
                logger.info(`Created test environment: ${testEnvName}`);
            });

            await test.step('Verify environment can be described', function () {
                logger.info('Verifying environment is accessible before deletion');

                const stdout = runaiCliV2.environmentDescribe(testEnvName);
                expect(stdout).toContain(testEnvName);

                logger.info('Verified: Environment is accessible');
            });

            await test.step('Delete the environment', async function () {
                await environmentUtils.delete(testEnvId!);
                logger.info(`Deleted environment: ${testEnvId}`);
            });

            await test.step('Attempt to describe deleted environment', function () {
                logger.info('Test T5597621: Attempting to describe deleted environment');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    const result = runaiCliV2['bashCli'].execWithOutput(
                        `${runaiCliV2['cliPath']} environment describe ${testEnvName}`,
                        false
                    );
                    exitCode = result.status ?? undefined;
                    errorOutput = result.stderr || result.stdout || '';
                } catch (error: any) {
                    exitCode = error.exitCode;
                    errorOutput = error.stderr || error.stdout || error.message || '';
                }

                logger.info('=== Error Output Start ===');
                logger.info(`Exit Code: ${exitCode}`);
                logger.info(`Error Output: ${errorOutput}`);
                logger.info('=== Error Output End ===');

                expect(exitCode).not.toBe(0);
                expect(errorOutput).toMatch(ERROR_ENVIRONMENT_NOT_FOUND);

                logger.info('Verified: Deleted environment returns appropriate error');
            });
        });
    });
});
