/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import { logger } from './src/modules/helpers';
import path from 'path';
import glob from 'glob';

interface TestCase {
    title: string;
    filePath: string;
    type: 'case' | 'caseGroup';
    templateId: string;
}

class TestCaseCollector {
    private testCases: TestCase[] = [];

    collectTestCases(): void {
        // Find all test files
        const currentDir = process.cwd();
        const testFiles = glob.sync('src/**/*.test.ts', { cwd: currentDir });

        logger.info(`Found ${testFiles.length} test files`);

        for (const file of testFiles) {
            const fullPath = path.resolve(currentDir, file);
            this.parseTestFile(fullPath, file);
        }
    }

    private parseTestFile(fullPath: string, relativePath: string): void {
        try {
            const content = fs.readFileSync(fullPath, 'utf-8');
            const lines = content.split('\n');

            for (let i = 0; i < lines.length; i++) {
                const line = lines[i].trim();

                // Match test() calls
                const testMatch = line.match(/test\s*\(\s*['"`]([^'"`]+)['"`]/);
                if (testMatch) {
                    // using regex to fetch templateId(T + 7 digits + : + should display the setting page) from test('T1234567: should display the setting page')
                    const templateIdMatch = testMatch[1].match(/T\d{7}:/);
                    if (!templateIdMatch) {
                        continue;
                    }
                    const templateId = templateIdMatch[0];
                    // remove : from templateId
                    const templateIdWithoutColon = templateId.replace(/:/, '');
                    this.testCases.push({
                        title: testMatch[1],
                        filePath: relativePath,
                        type: 'case',
                        templateId: templateIdWithoutColon
                    });
                }
            }
        } catch (error) {
            logger.error(`Error parsing file ${relativePath}:`, error);
        }
    }

    printResults(): void {
        logger.info('\n=== Test Case Title and File Path Mapping ===\n');

        // Group by file
        const groupedByFile = this.testCases.reduce(
            (acc, testCase) => {
                if (!acc[testCase.filePath]) {
                    acc[testCase.filePath] = [];
                }
                acc[testCase.filePath].push(testCase);
                return acc;
            },
            {} as Record<string, TestCase[]>
        );

        Object.keys(groupedByFile)
            .sort()
            .forEach((filePath) => {
                logger.info(`\n📁 File: ${filePath}`);
                logger.info('─'.repeat(80));

                groupedByFile[filePath].forEach((testCase) => {
                    const icon = testCase.type === 'case' ? '🧪' : '📋';
                    logger.info(`${icon} [${testCase.type.toUpperCase()}]: ${testCase.title}`);
                });
            });

        logger.info(
            `\n📊 Summary: Found ${this.testCases.length} test cases in ${Object.keys(groupedByFile).length} files`
        );
    }

    exportToJson(outputPath = 'test-cases-mapping.json'): void {
        const output = {
            summary: {
                totalTestCases: this.testCases.length,
                totalFiles: new Set(this.testCases.map((tc) => tc.filePath)).size,
                generatedAt: new Date().toISOString()
            },
            testCases: this.testCases
        };

        fs.writeFileSync(outputPath, JSON.stringify(output, null, 2));
        logger.info(`\n💾 Results exported to: ${outputPath}`);
    }
}

function main(): void {
    const collector = new TestCaseCollector();

    logger.info('🔍 Collecting test cases...');
    collector.collectTestCases();

    collector.printResults();
    collector.exportToJson();
}

// Execute main function directly
main();
