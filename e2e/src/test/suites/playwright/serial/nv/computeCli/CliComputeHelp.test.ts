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

test.describe(`CLI Compute Help Tests, tags: ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        test('T5597682: Verify compute help output', function () {
            test.step('Execute compute help command and verify output', function () {
                logger.info('Test T5597682: Executing compute help command');
                const stdout = runaiCliV2.computeHelp();

                logger.info('=== Compute Help Output Start ===');
                logger.info(stdout);
                logger.info('=== Compute Help Output End ===');

                // Verify the help output contains all expected content
                const expectedContent = [
                    'Compute resource asset management',
                    'runai compute [command]',
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

        test('T5597683: Verify compute help flag variants', function () {
            test.step('Execute compute -h and verify output', function () {
                logger.info('Test T5597683: Executing compute -h command');
                const stdout = runaiCliV2.computeHelp('-h');

                logger.info('=== Compute -h Output Start ===');
                logger.info(stdout);
                logger.info('=== Compute -h Output End ===');

                // Verify output contains help information
                expect(stdout).toContain('Compute resource asset management');
                expect(stdout).toContain('runai compute [command]');
                expect(stdout).toContain('Available Commands:');

                logger.info('Verified: -h flag displays help information');
            });

            test.step('Verify both -h and --help produce identical output', function () {
                const stdoutShort = runaiCliV2.computeHelp('-h');
                const stdoutLong = runaiCliV2.computeHelp('--help');

                // Both outputs should be identical
                expect(stdoutShort).toEqual(stdoutLong);

                logger.info('Verified: -h and --help produce identical output');
            });
        });

        test('T5597684: Execute compute without subcommand', function () {
            test.step('Execute compute without subcommand and verify output', function () {
                logger.info('Test T5597684: Executing compute without subcommand');

                // Execute without throwing on non-zero exit
                const result = runaiCliV2['bashCli'].execWithOutput(
                    `${runaiCliV2['cliPath']} compute`,
                    false
                );

                const output = result.stdout || result.stderr;

                logger.info('=== Compute No-Subcommand Output Start ===');
                logger.info(output);
                logger.info('=== Compute No-Subcommand Output End ===');

                // Should display help information or error about missing subcommand
                const hasHelpInfo =
                    output.includes('Compute resource asset management') ||
                    output.includes('Available Commands:') ||
                    output.includes('runai compute [command]');

                const hasErrorInfo = output.includes('required') || output.includes('subcommand');

                // Either shows help or error about missing subcommand
                expect(hasHelpInfo || hasErrorInfo).toBe(true);

                logger.info('Verified: Command without subcommand shows help or appropriate error');
            });
        });

        test('T5597685: Verify compute with invalid subcommand', function () {
            test.step('Execute compute with invalid subcommand "invalid-command"', function () {
                logger.info('Test T5597685: Executing compute with invalid subcommand');

                const stdoutInvalid = runaiCliV2.computeHelp('invalid-command');
                const stdoutHelp = runaiCliV2.computeHelp('--help');

                // Verify invalid subcommand displays the same help output as --help flag
                expect(stdoutInvalid).toEqual(stdoutHelp);

                logger.info('Verified: Invalid subcommand "invalid-command" displays help information');
            });

            test.step('Execute compute with invalid subcommand "create"', function () {
                logger.info('Executing compute with subcommand "create" (should be invalid)');

                const stdoutCreate = runaiCliV2.computeHelp('create');
                const stdoutHelp = runaiCliV2.computeHelp('--help');

                // Verify "create" subcommand also displays help (since it's not a valid compute subcommand)
                expect(stdoutCreate).toEqual(stdoutHelp);

                logger.info('Verified: Invalid subcommand "create" displays help information');
            });
        });

        test('T5597663: Verify compute describe help output', function () {
            test.step('Execute compute describe help command', function () {
                logger.info('Test T5597663: Verifying compute describe help output');

                const stdout = runaiCliV2.computeHelp('describe --help');

                logger.info('=== Describe Help Output Start ===');
                logger.info(stdout);
                logger.info('=== Describe Help Output End ===');

                // Verify key sections in help output
                expect(stdout).toContain('describe');
                expect(stdout).toContain('compute');
                expect(stdout).toMatch(/usage|Usage/i);
                expect(stdout).toMatch(/-h|--help/);
                expect(stdout).toMatch(/-o|--output/);
                expect(stdout).toMatch(/-p|--project/);

                logger.info('Verified: Describe help output contains expected information');
            });
        });

        test('T5597681: Verify compute list help output', function () {
            test.step('Execute compute list help command', function () {
                logger.info('Test T5597681: Verifying compute list help output');

                const stdout = runaiCliV2.computeHelp('list --help');

                logger.info('=== List Help Output Start ===');
                logger.info(stdout);
                logger.info('=== List Help Output End ===');

                // Verify key sections in help output
                expect(stdout).toContain('list');
                expect(stdout).toContain('compute');
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
