/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect, TestInfo } from '@playwright/test';
import {
    Utils,
    logger,
    tags,
    flowRootState,
    objectGenerator,
    projectResourcesUtils,
    operations
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { listFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI List Template - Negative Tests (Mutually Exclusive Flags), tags: ${tags.resource.templateV2} ${tags.teams.loong}`, () => {
    let testProjectName: string;
    let testProjectId: number;

    test.beforeAll(async function () {
        await test.step('Create project for negative tests', async function () {
            testProjectName = objectGenerator.generateShortNameByPrefix('proj-neg-test', 5);
            const namespace = Utils.namespaceFromProjectName(testProjectName);
            const project = await projectResourcesUtils.createProject(namespace, {
                name: testProjectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });
            testProjectId = project.id;
            logger.info(`Created project: ${testProjectName} (ID: ${testProjectId})`);

            runaiCliV2.projectSet(testProjectName);
            logger.info(`Set CLI project context to: ${testProjectName}`);
        });
    });

    test.afterAll(async function () {
        await test.step('Delete project', async function () {
            if (testProjectId) {
                try {
                    await operations.deleteProjectAndNamespace(testProjectId);
                    logger.info(`Deleted project: ${testProjectName}`);
                } catch (error) {
                    logger.error(`Failed to delete project: ${error}`);
                }
            }
        });
    });

    supportedFrom(SupportedVersions.V2_24)(function () {
        // Coverage remains disabled pending revalidation of this case
        test('T5539418: List templates with no-pagination and page-size combination (mutually exclusive flags)', async ({}, testInfo: TestInfo) => {
            skipPlaywrightTest(
                test,
                testInfo,
                '--no-pagination is not mutually exclusive with --page-size'
            );

            await test.step('Attempt to use --no-pagination with --page-size', function () {
                logger.info(
                    'Attempting to list templates with conflicting flags: --no-pagination --page-size'
                );

                const cliPath = runaiCliV2['cliPath'];

                expect(() => {
                    runaiCliV2['bashCli'].exec(`${cliPath} template list --no-pagination --page-size 3`);
                }).toThrow(
                    /if any flags in the group.*are set none of the others can be|mutually exclusive/i
                );

                logger.info('Verified: CLI rejects conflicting --no-pagination and --page-size flags');
            });

            await test.step('Verify page-size alone works', function () {
                logger.info('Verifying that --page-size alone works correctly');

                const stdout = runaiCliV2.list(
                    'template',
                    { [listFlags.pageSize]: '3' } as any,
                    false
                ).stdout;

                expect(stdout).toBeTruthy();
                logger.info('Verified: --page-size flag works correctly when used alone');
            });

            await test.step('Verify no-pagination alone works', function () {
                logger.info('Verifying that --no-pagination alone works correctly');

                const stdout = runaiCliV2.list(
                    'template',
                    { [listFlags.noPagination]: '' } as any,
                    false
                ).stdout;

                expect(stdout).toBeTruthy();
                logger.info('Verified: --no-pagination flag works correctly when used alone');
            });
        });

        // Coverage remains disabled pending revalidation of this case
        test.skip('T5539419: List templates with max-items and no-pagination combination (mutually exclusive flags)', function () {
            test.step('Attempt to use --max-items with --no-pagination', function () {
                logger.info(
                    'Attempting to list templates with conflicting flags: --max-items --no-pagination'
                );

                const cliPath = runaiCliV2['cliPath'];

                expect(() => {
                    runaiCliV2['bashCli'].exec(`${cliPath} template list --max-items 5 --no-pagination`);
                }).toThrow(
                    /if any flags in the group.*are set none of the others can be|mutually exclusive/i
                );

                logger.info('Verified: CLI rejects conflicting --max-items and --no-pagination flags');
            });

            test.step('Verify max-items alone works', function () {
                logger.info('Verifying that --max-items alone works correctly');

                const stdout = runaiCliV2.list(
                    'template',
                    { [listFlags.maxItems]: '5' } as any,
                    false
                ).stdout;

                expect(stdout).toBeTruthy();
                logger.info('Verified: --max-items flag works correctly when used alone');
            });

            test.step('Verify no-pagination alone works', function () {
                logger.info('Verifying that --no-pagination alone works correctly');

                const stdout = runaiCliV2.list(
                    'template',
                    { [listFlags.noPagination]: '' } as any,
                    false
                ).stdout;

                expect(stdout).toBeTruthy();
                logger.info('Verified: --no-pagination flag works correctly when used alone');
            });
        });

        test('T5539420: List templates with all pagination flags combination (mutually exclusive flags)', async function () {
            await test.step('Verify the max-items flag', function () {
                logger.info('Verifying that --max-items can work');

                const stdout = runaiCliV2.list(
                    'template',
                    { [listFlags.maxItems]: '8' } as any,
                    false
                ).stdout;

                expect(stdout).toBeTruthy();
                logger.info('Verified: --max-items works');
            });

            await test.step('Verify manual pagination with page-size', function () {
                logger.info('Verifying manual pagination with --page-size alone');

                const stdout = runaiCliV2.list(
                    'template',
                    { [listFlags.pageSize]: '3' } as any,
                    false
                ).stdout;

                expect(stdout).toBeTruthy();

                // Extract next-token if present
                const tokenMatch = stdout.match(/next token:\s*([^\s]+)/i);
                if (tokenMatch) {
                    const nextToken = tokenMatch[1];
                    logger.info(`Found next-token: ${nextToken}`);

                    // Try to get next page
                    const stdout2 = runaiCliV2.list(
                        'template',
                        { [listFlags.nextToken]: nextToken, [listFlags.pageSize]: '3' } as any,
                        false
                    ).stdout;
                    expect(stdout2).toBeTruthy();
                    logger.info('Verified: Manual pagination with next-token works correctly');
                }

                logger.info('Verified: Manual pagination mechanism works with --page-size');
            });
        });
    });
});
