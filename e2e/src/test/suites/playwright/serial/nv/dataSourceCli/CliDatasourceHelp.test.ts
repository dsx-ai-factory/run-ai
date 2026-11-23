/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { logger, tags } from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { runaiCliV2 } from '../../../../../../modules/cli';

test.describe(`CLI Datasource Help Tests, tags: ${tags.resource.datasource} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        test('T5585434: Verify datasource help output', function () {
            test.step('Execute datasource help command and verify output', function () {
                logger.info('Executing datasource help command');
                const stdout = runaiCliV2.datasourceHelp();

                logger.info('=== Datasource Help Output Start ===');
                logger.info(stdout);
                logger.info('=== Datasource Help Output End ===');

                // Verify the entire help output contains all expected content
                const expectedContent = [
                    'Data Source asset management',
                    'runai datasource [command]',
                    'Available Commands:',
                    'describe',
                    'list',
                    'Flags:',
                    '-h, --help'
                ];

                // Check all expected content is present in the output
                expectedContent.forEach((content) => {
                    expect(stdout).toContain(content);
                });

                logger.info('Verified: Help output contains all expected information');
            });
        });

        test('T5585435: Execute datasource without subcommand', function () {
            test.step('Execute datasource without subcommand and verify output', function () {
                logger.info('Executing datasource without subcommand');

                // Execute without throwing on non-zero exit
                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} datasource`,
                    false
                );

                const output = result.stdout || result.stderr;

                logger.info('=== Datasource No-Subcommand Output Start ===');
                logger.info(output);
                logger.info('=== Datasource No-Subcommand Output End ===');

                // Should display help information or error about missing subcommand
                const hasHelpInfo =
                    output.includes('Data Source asset management') ||
                    output.includes('Available Commands:') ||
                    output.includes('runai datasource [command]');

                const hasErrorInfo = output.includes('required') || output.includes('subcommand');

                // Either shows help or error about missing subcommand
                expect(hasHelpInfo || hasErrorInfo).toBe(true);

                logger.info('Verified: Command without subcommand shows help or appropriate error');
            });
        });

        test('T5585436: Verify datasource with invalid subcommand', function () {
            test.step('Execute datasource with invalid subcommand and verify help is displayed', function () {
                logger.info('Executing datasource with invalid subcommand');

                const stdoutInvalid = runaiCliV2.datasourceHelp('invalid-command');
                const stdoutHelp = runaiCliV2.datasourceHelp('--help');

                // Verify invalid subcommand displays the same help output as --help flag
                expect(stdoutInvalid).toEqual(stdoutHelp);

                logger.info('Verified: Invalid subcommand displays help information');
            });
        });

        test('T5585437: Verify datasource help flag variants', function () {
            test.step('Execute datasource -h and verify output', function () {
                logger.info('Executing datasource -h command');
                const stdout = runaiCliV2.datasourceHelp('-h');

                logger.info('=== Datasource -h Output Start ===');
                logger.info(stdout);
                logger.info('=== Datasource -h Output End ===');

                // Verify output contains help information
                expect(stdout).toContain('Data Source asset management');
                expect(stdout).toContain('runai datasource [command]');
                expect(stdout).toContain('Available Commands:');

                logger.info('Verified: -h flag displays help information');
            });

            test.step('Execute datasource --help and verify output', function () {
                logger.info('Executing datasource --help command');
                const stdout = runaiCliV2.datasourceHelp('--help');

                logger.info('=== Datasource --help Output Start ===');
                logger.info(stdout);
                logger.info('=== Datasource --help Output End ===');

                // Verify output contains help information
                expect(stdout).toContain('Data Source asset management');
                expect(stdout).toContain('runai datasource [command]');
                expect(stdout).toContain('Available Commands:');

                logger.info('Verified: --help flag displays help information');
            });
        });
    });
});
