/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import path from 'path';
import { teams } from './src/modules/helpers/config/tags';
import { logger } from './src/modules/helpers';

const FAILED_STATUSES = ['broken', 'failed'];
const SKIPPED_STATUS = ['skipped'];

const emojiMap: { [key: string]: string } = {
    phoenix: ':phoenix:',
    pegasus: ':pegasus:',
    viper: '🐍',
    loong: ':dragon:'
};

interface TestResult {
    testName: string;
    fullName: string;
    team: string;
    status: string;
    startTime: number;
    labels: string[];
    skipReason: string;
}

function extractTagsFromAllure(fullTestName: string): string[] {
    const labels: string[] = [];
    const tagRegex = /@([\w-]+)/g;
    let match;

    while ((match = tagRegex.exec(fullTestName)) !== null) {
        const foundLabel = match[1];
        if (!Object.keys(teams).includes(foundLabel)) {
            labels.push(foundLabel);
        }
    }
    return labels;
}

//remove tags
function cleanTestTitle(title: string): string {
    return title.replace(/@\w+/g, '').trim();
}

//`@${foundLabel}`
function filterTestResults(allTestResults: TestResult[]): TestResult[] {
    const passedTests = new Set<string>();
    const failedTests = new Set<string>();
    const latestResults = new Map<string, TestResult>();

    // Dedup by fullName so that parameterized describes (same inner test name
    // under different describe blocks) are not collapsed into one bucket.
    allTestResults.forEach((test) => {
        const uniqueKey = test.fullName;

        if (test.status === 'passed') {
            passedTests.add(uniqueKey);
        } else if (test.status === 'failed' || test.status === 'broken') {
            failedTests.add(uniqueKey);
        }

        if (!latestResults.has(uniqueKey) || test.startTime > latestResults.get(uniqueKey)!.startTime) {
            latestResults.set(uniqueKey, test);
        }
    });

    return Array.from(latestResults.values()).filter((test) => {
        const key = test.fullName;

        // If the test passed, remove it
        if (passedTests.has(key)) {
            return false;
        }

        // If the test has both failed and passed statuses, remove both
        return !(failedTests.has(key) && passedTests.has(key));
    });
}

export function fetchTestResults(): TestResult[] {
    const allureResultsDir = './allure-results';
    if (!fs.existsSync(allureResultsDir)) {
        logger.info('allure-results directory not found, skipping report aggregation');
        return [];
    }
    const reportFiles: string[] = fs
        .readdirSync(allureResultsDir)
        .filter((file) => file.endsWith('-result.json'));

    logger.info(`Found ${reportFiles.length} report files in the allure-results directory`);

    const allTestResults: TestResult[] = [];

    for (const file of reportFiles) {
        const filePath = path.join(allureResultsDir, file);
        const fileContent = fs.readFileSync(filePath, 'utf-8');
        const parsedData = JSON.parse(fileContent);

        const isMochaTest = Object.keys(teams).some((team) => parsedData['fullName']?.includes(teams[team]));

        const subSuiteValue: string =
            parsedData.labels?.find((label) => label.name === 'subSuite')?.value ?? '';
        const suiteValue: string =
            parsedData.labels?.find((label) => label.name === 'suite')?.value ?? '';

        const fullTestNameWithTags = isMochaTest ? parsedData['fullName'] : subSuiteValue;

        let teamName: string | undefined = undefined;

        if (isMochaTest) {
            teamName = Object.keys(teams).find((team) => parsedData['fullName']?.includes(teams[team]));
        } else if (parsedData.labels) {
            teamName = parsedData.labels.find(
                (label) => label.name === 'tag' && Object.values(teams).includes(`@${label.value}`)
            )?.value;
        }

        // Fallback: if no team tag came back from Allure labels, look for an
        // @<team> substring in the describe path (subSuite). Handles cases
        // where the tag exists only in the describe title and was not lifted
        // into a separate `tag` label by the Allure reporter.
        if (!teamName && subSuiteValue) {
            teamName = Object.keys(teams).find((team) => subSuiteValue.includes(teams[team]));
        }

        teamName = teamName ? teamName : 'not specified';

        const labels = extractTagsFromAllure(fullTestNameWithTags);

        const testName = cleanTestTitle(parsedData['name']);
        // Build a unique-per-test identifier. Playwright's Allure reporter sets
        // `fullName` to `file:line:col`, which is shared by every iteration of
        // parameterized patterns (`tests.forEach(tc => test(tc.name, ...))` or
        // `describe.forEach(...)`). Composing suite + subSuite + name keeps
        // retries of the same test in one bucket while separating distinct
        // iterations so dedup doesn't drop a failure when a sibling passes.
        // Mocha results already carry the full describe path in `fullName`,
        // so we keep that path for them.
        const playwrightFullName = [suiteValue, subSuiteValue, parsedData['name']]
            .filter(Boolean)
            .join(' > ');
        const fullName: string =
            playwrightFullName ||
            (typeof parsedData['fullName'] === 'string' && parsedData['fullName']) ||
            parsedData['name'] ||
            '';

        allTestResults.push({
            team: teamName,
            status: parsedData.status,
            testName,
            fullName,
            startTime: parsedData.start,
            labels,
            skipReason: parsedData.statusDetails?.message || ''
        });
    }

    // Remove passed tests & tests that passed after retry
    return filterTestResults(allTestResults);
}

export function createReport(statusList: string[], filePath: string, results: TestResult[]): void {
    const testsCountByTeam: { [key: string]: number } = {};

    for (const result of results) {
        const { status, team } = result;

        if (statusList.includes(status)) {
            const teamKey = team || 'tests without a team name'; // Ensure team is not undefined
            if (!testsCountByTeam[teamKey]) {
                testsCountByTeam[teamKey] = 0;
            }
            testsCountByTeam[teamKey]++;
        }
    }

    if (Object.keys(testsCountByTeam).length === 0) {
        logger.info(`There are no tests with status: ${statusList}`);
        return;
    }

    const reportData =
        '\n' +
        Object.entries(testsCountByTeam)
            .map(([team, count]) => {
                const emoji = emojiMap[team] || ''; // Use emojiMap for the emoji if available
                return `${emoji} ${team}: ${count}`;
            })
            .join('\n');

    fs.writeFileSync(filePath, reportData);

    logger.info(`Report generated at: ${filePath}`);
}

export function createFailuresReport(results: TestResult[]): void {
    createReport(FAILED_STATUSES, './failuresCountPerTeam.txt', results);
}

export function createSkippedTestsReport(results: TestResult[]): void {
    createReport(SKIPPED_STATUS, './skippedTestsPerTeam.txt', results);
}

import { fileURLToPath } from 'url';

const isRunDirectly = process.argv[1] === fileURLToPath(import.meta.url);

if (isRunDirectly) {
    try {
        const results = fetchTestResults();
        createFailuresReport(results);
        createSkippedTestsReport(results);
    } catch (error) {
        console.error('Error:', error.message);
    }
}
