/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import path from 'path';
import os from 'os';
import { test, expect } from '@playwright/test';
import { tags } from '../../../../../../modules/helpers';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { logger } from '../../../../../../modules/helpers';
import { getTarEntries } from '../../../../../../modules/helpers/utils/TarUtils';
import { collectDiagnosticLogsFlags } from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    CLI_OUTPUT_LOG_FILE_NAME,
    ALLOWED_NAMESPACES,
    extractTarTopLevelFolders,
    assertLogsStructure,
    assertNoPreviousLogs
} from './collectLogsHelpers';

test.describe(
    `Validate collect diagnostic logs: ${tags.teams.phoenix}`,
    { tag: [tags.services.runaiCli] },
    () => {
        let outputDirPath: string;

        test.beforeEach(() => {
            outputDirPath = fs.mkdtempSync(path.join(os.tmpdir(), 'diagnostics-logs-'));
        });

        test.afterEach(() => {
            fs.rmSync(outputDirPath, { recursive: true });
        });

        test('should collect diagnostics logs', async () => {
            runaiCliV2.collectDiagnosticLogs({ [collectDiagnosticLogsFlags.outputDir]: outputDirPath });

            const files = fs.readdirSync(outputDirPath);
            const tarGzFileName = files.find(
                (file) => file.startsWith('runai-logs-') && file.endsWith('.tar.gz')
            );
            expect(tarGzFileName).toBeDefined();

            const tarGzFilePath = path.join(outputDirPath, tarGzFileName);

            const tarEntries = await getTarEntries(tarGzFilePath);
            expect(tarEntries.length).toBeGreaterThan(0);

            const cliOutputLogEntry = tarEntries.find((entry) => entry.path === CLI_OUTPUT_LOG_FILE_NAME);
            expect(cliOutputLogEntry).toBeDefined();
            expect(cliOutputLogEntry.size).toBeGreaterThan(0);

            const collectedNamespaces = extractTarTopLevelFolders(tarEntries);
            logger.info(`Collected namespaces: ${collectedNamespaces.join(', ')}`);
            expect(collectedNamespaces.length).toBeGreaterThan(0);
            expect(collectedNamespaces.every((ns) => ALLOWED_NAMESPACES.includes(ns))).toBe(true);

            assertLogsStructure(tarEntries);
        });

        test('should not collect previous logs if --no-previous flag is provided', async () => {
            runaiCliV2.collectDiagnosticLogs({
                [collectDiagnosticLogsFlags.outputDir]: outputDirPath,
                [collectDiagnosticLogsFlags.noPrevious]: ''
            });

            const files = fs.readdirSync(outputDirPath);
            const tarGzFileName = files.find(
                (file) => file.startsWith('runai-logs-') && file.endsWith('.tar.gz')
            );
            expect(tarGzFileName).toBeDefined();

            const tarGzFilePath = path.join(outputDirPath, tarGzFileName);

            const tarEntries = await getTarEntries(tarGzFilePath);
            expect(tarEntries.length).toBeGreaterThan(0);

            assertNoPreviousLogs(tarEntries);
        });

        test('should fail if namespace is not allowed', () => {
            try {
                runaiCliV2.collectDiagnosticLogs({
                    [collectDiagnosticLogsFlags.outputDir]: outputDirPath,
                    [collectDiagnosticLogsFlags.namespaces]: 'non-allowed-namespace'
                });

                throw new Error(
                    'Expected collectDiagnosticLogs to fail with non-allowed namespace, but it succeeded'
                );
            } catch (error) {
                expect(error.message).toContain(
                    'invalid argument "non-allowed-namespace" for "--namespaces" flag'
                );
            }
        });
    }
);
