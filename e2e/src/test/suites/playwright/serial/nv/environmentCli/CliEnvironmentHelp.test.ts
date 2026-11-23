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

test.describe(`CLI Environment Help Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        test('T5597641: Verify environment help output', function () {
            test.step('Execute environment help command and verify output', function () {
                logger.info('Test T5597641: Executing environment help command');
                const stdout = runaiCliV2.environmentHelp();

                logger.info('=== Environment Help Output Start ===');
                logger.info(stdout);
                logger.info('=== Environment Help Output End ===');

                // Verify the help output contains all expected content
                const expectedContent = [
                    'Environment asset management',
                    'runai environment [command]',
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

        test('T5597642: Verify environment help flag variants', function () {
            test.step('Execute environment -h and verify output', function () {
                logger.info('Test T5597642: Executing environment -h command');
                const stdout = runaiCliV2.environmentHelp('-h');

                logger.info('=== Environment -h Output Start ===');
                logger.info(stdout);
                logger.info('=== Environment -h Output End ===');

                // Verify output contains help information
                expect(stdout).toContain('Environment asset management');
                expect(stdout).toContain('runai environment [command]');
                expect(stdout).toContain('Available Commands:');

                logger.info('Verified: -h flag displays help information');
            });

            test.step('Verify both -h and --help produce identical output', function () {
                const stdoutShort = runaiCliV2.environmentHelp('-h');
                const stdoutLong = runaiCliV2.environmentHelp('--help');

                // Both outputs should be identical
                expect(stdoutShort).toEqual(stdoutLong);

                logger.info('Verified: -h and --help produce identical output');
            });
        });

        test('T5597643: Execute environment without subcommand', function () {
            test.step('Execute environment without subcommand and verify output', function () {
                logger.info('Test T5597643: Executing environment without subcommand');

                // Execute without throwing on non-zero exit
                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} environment`,
                    false
                );

                const output = result.stdout || result.stderr;

                logger.info('=== Environment No-Subcommand Output Start ===');
                logger.info(output);
                logger.info('=== Environment No-Subcommand Output End ===');

                // Should display help information or error about missing subcommand
                const hasHelpInfo =
                    output.includes('Environment asset management') ||
                    output.includes('Available Commands:') ||
                    output.includes('runai environment [command]');

                const hasErrorInfo = output.includes('required') || output.includes('subcommand');

                // Either shows help or error about missing subcommand
                expect(hasHelpInfo || hasErrorInfo).toBe(true);

                logger.info('Verified: Command without subcommand shows help or appropriate error');
            });
        });

        test('T5597644: Verify environment with invalid subcommand', function () {
            test.step('Execute environment with invalid subcommand "invalid-subcommand"', function () {
                logger.info('Test T5597644: Executing environment with invalid subcommand');

                const stdoutInvalid = runaiCliV2.environmentHelp('invalid-subcommand');
                const stdoutHelp = runaiCliV2.environmentHelp('--help');

                // Verify invalid subcommand displays the same help output as --help flag
                expect(stdoutInvalid).toEqual(stdoutHelp);

                logger.info('Verified: Invalid subcommand "invalid-subcommand" displays help information');
            });

            test.step('Execute environment with invalid subcommand "create"', function () {
                logger.info('Executing environment with subcommand "create" (should be invalid)');

                const stdoutCreate = runaiCliV2.environmentHelp('create');
                const stdoutHelp = runaiCliV2.environmentHelp('--help');

                // Verify "create" subcommand also displays help (since it's not a valid environment subcommand)
                expect(stdoutCreate).toEqual(stdoutHelp);

                logger.info('Verified: Invalid subcommand "create" displays help information');
            });
        });

        test('T5597622: Verify environment describe help output', function () {
            test.step('Execute environment describe help command', function () {
                logger.info('Test T5597622: Verifying environment describe help output');

                const stdout = runaiCliV2.environmentHelp('describe --help');

                logger.info('=== Describe Help Output Start ===');
                logger.info(stdout);
                logger.info('=== Describe Help Output End ===');

                // Verify key sections in help output
                expect(stdout).toContain('describe');
                expect(stdout).toContain('environment');
                expect(stdout).toMatch(/usage|Usage/i);
                expect(stdout).toMatch(/-h|--help/);
                expect(stdout).toMatch(/-o|--output/);
                expect(stdout).toMatch(/-p|--project/);

                logger.info('Verified: Describe help output contains expected information');
            });
        });

        test('T5597640: Verify environment list help output', function () {
            test.step('Execute environment list help command', function () {
                logger.info('Test T5597640: Verifying environment list help output');

                const stdout = runaiCliV2.environmentHelp('list --help');

                logger.info('=== List Help Output Start ===');
                logger.info(stdout);
                logger.info('=== List Help Output End ===');

                // Verify key sections in help output
                expect(stdout).toContain('list');
                expect(stdout).toContain('environment');
                expect(stdout).toMatch(/usage|Usage/i);
                expect(stdout).toMatch(/-h|--help/);
                expect(stdout).toMatch(/--json|--yaml|--table/);
                expect(stdout).toMatch(/-p|--project/);
                expect(stdout).toMatch(/-A|--all/i);

                logger.info('Verified: List help output contains expected information');
            });
        });
    });
});
