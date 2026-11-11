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
import { listFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';

test.describe(`CLI List Compute - Negative Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        // Error message patterns
        const ERROR_PROJECT_NOT_FOUND = /Failed to list compute assets.*not found/i;
        const ERROR_UNKNOWN_FLAG = /unknown flag/i;
        const ERROR_CONFLICTING_OUTPUT_FORMATS =
            /if any flags in the group.*are set none of the others can be/i;

        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-list-neg', 5);
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
        test('T5597677: List compute resources with invalid project', async function () {
            await test.step('Attempt to list with invalid project', function () {
                logger.info('Test T5597677: Attempting to list compute with invalid project');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    runaiCliV2.list(
                        'compute',
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

        test('T5597678: List compute resources with invalid output format', async function () {
            await test.step('Attempt to list with invalid format', function () {
                logger.info('Test T5597678: Attempting to list with invalid output format');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    // Manually construct command with invalid --output xml flag
                    const result = runaiCliV2['bashCli'].execWithOutput(
                        `${runaiCliV2['cliPath']} compute list --output xml`,
                        false
                    );
                    exitCode = result.status;
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

        test('T5597679: List compute resources with conflicting output formats', async function () {
            await test.step('Attempt to list with conflicting formats', function () {
                logger.info('Test T5597679: Attempting to list with conflicting output formats');

                let exitCode: number | undefined;
                let errorOutput = '';

                try {
                    runaiCliV2.list(
                        'compute',
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
    });
});
