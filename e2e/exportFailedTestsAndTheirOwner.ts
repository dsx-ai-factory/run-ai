/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import { logger } from './src/modules/helpers';
import { fetchTestResults } from './aggregateResultsPerTeam';

const FAILED_STATUSES = ['broken', 'failed'];

function generateFailedTestsJson(
    testResults: { testName: string; team: string; status: string }[]
): Record<string, string[]> {
    const failedTestsByTeam: Record<string, string[]> = {};
    testResults.forEach(({ testName, team, status }) => {
        if (FAILED_STATUSES.includes(status)) {
            const teamKey = `@${team}`;
            if (!failedTestsByTeam[teamKey]) {
                failedTestsByTeam[teamKey] = [];
            }
            failedTestsByTeam[teamKey].push(testName);
        }
    });
    return failedTestsByTeam;
}

((): void => {
    const results = fetchTestResults();

    const failedTestsJson = generateFailedTestsJson(results);
    const filePath = './failedTestsJsonForCi.json';
    fs.writeFileSync(filePath, JSON.stringify(failedTestsJson, null, 2));
    logger.info(`Failed tests JSON written to: ${filePath}`);
})();
