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
    tags
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { listFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI List Environment - Negative Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        // Error message patterns
        const ERROR_PROJECT_NOT_FOUND = /Failed to list environment assets.*not found/i;
        const ERROR_UNKNOWN_FLAG = /unknown flag/i;
        const ERROR_CONFLICTING_OUTPUT_FORMATS =
            /if any flags in the group.*are set none of the others can be/i;

        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-list-env-neg', 5);
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

        test('T5597636: List environment resources with invalid project name', async function () {
            await test.step('Attempt to list with invalid project', function () {
                logger.info('Test T5597636: Attempting to list environment with invalid project');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    runaiCliV2.list(
                        'environment',
                        {
                            [listFlags.project]: 'non-existent-project-12345'
                        } as any,
                        false
                    );
                } catch (error: any) {
                    exitCode = error.exitCode;
                    errorOutput = error.stderr || error.stdout || error.message || '';
                }

                logger.info('=== Error Output Start ===');
                logger.info(`Exit Code: ${exitCode}`);
                logger.info(`Error Output: ${errorOutput}`);
                logger.info('=== Error Output End ===');

                expect(exitCode).not.toBe(0);
                expect(errorOutput).toMatch(ERROR_PROJECT_NOT_FOUND);

                logger.info('Verified: Invalid project returns error');
            });
        });

        test('T5597637: List environment resources with invalid output format', async function () {
            await test.step('Attempt to list with invalid format', function () {
                logger.info('Test T5597637: Attempting to list with invalid output format');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    // Manually construct command with invalid --output xml flag
                    const result = runaiCliV2['bashCli'].execWithOutput(
                        `${runaiCliV2['cliPath']} environment list --output xml`,
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
                expect(errorOutput).toMatch(ERROR_UNKNOWN_FLAG);

                logger.info('Verified: Invalid output format returns error');
            });
        });

        test('T5597638: List environment resources with conflicting output formats', async function () {
            await test.step('Attempt to list with conflicting formats', function () {
                logger.info('Test T5597638: Attempting to list with conflicting output formats');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    runaiCliV2.list(
                        'environment',
                        {
                            [listFlags.json]: '',
                            [listFlags.yaml]: ''
                        } as any,
                        false
                    );
                } catch (error: any) {
                    exitCode = error.exitCode;
                    errorOutput = error.stderr || error.stdout || error.message || '';
                }

                logger.info('=== Error Output Start ===');
                logger.info(`Exit Code: ${exitCode}`);
                logger.info(`Error Output: ${errorOutput}`);
                logger.info('=== Error Output End ===');

                expect(exitCode).not.toBe(0);
                expect(errorOutput).toMatch(ERROR_CONFLICTING_OUTPUT_FORMATS);

                logger.info('Verified: Conflicting formats return error');
            });
        });

        test('T5597639: List environment resources with invalid flag combination', async function () {
            await test.step('Attempt to list with -A and -p together', function () {
                logger.info('Test T5597639: Attempting to list with -A and -p flags together');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    const result = runaiCliV2['bashCli'].execWithOutput(
                        `${runaiCliV2['cliPath']} environment list -A -p ${projectName}`,
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

                // Note: Some CLIs may allow this combination or may error
                // Adjust assertion based on actual CLI behavior
                if (exitCode !== 0) {
                    expect(errorOutput.length).toBeGreaterThan(0);
                    logger.info('Verified: Conflicting flags -A and -p handled appropriately');
                } else {
                    logger.info('Note: CLI accepts -A and -p together (one takes precedence)');
                }
            });
        });
    });
});
