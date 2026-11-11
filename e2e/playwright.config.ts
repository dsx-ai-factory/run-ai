/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as dotenv from 'dotenv';
import { defineConfig } from '@playwright/test';
import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

// ESM equivalent of __dirname
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
// Load `.env` relative to this config so test discovery works regardless of cwd
// (e.g. the Playwright Test extension in VS Code/Cursor spawns from the workspace root).
dotenv.config({ path: path.resolve(__dirname, '.env') });
const stateFile = path.resolve(__dirname, './globalState.json');
const isSelfHosted = process.env.IS_SELF_HOSTED && process.env.IS_SELF_HOSTED !== ''
    ? process.env.IS_SELF_HOSTED === 'true'
    : (process.env.CLUSTER_NAME_PREFIX
        ? !process.env.CLUSTER_NAME_PREFIX.includes('saas-')
        : false);
const isLocalRun = process.env.LOCAL_RUN ? process.env.LOCAL_RUN === 'true' : false;
const retries = process.env.RETRIES ? Number(process.env.RETRIES) : 2;

export const getSharedConfig = (): any => {
    if (!fs.existsSync(stateFile)) {
        return undefined;
    }
    return JSON.parse(fs.readFileSync(stateFile, 'utf-8'));
};

let timer: NodeJS.Timeout | null = null;

export const setTimer = (newTimer: NodeJS.Timeout): void => {
    timer = newTimer;
};

export const clearTimer = (): void => {
    if (timer) {
        clearInterval(timer);
        timer = null;
    }
};

if (!isLocalRun) process.env.NODE_OPTIONS = '--experimental-specifier-resolution=node';

export default defineConfig({
    testDir: 'src/test/suites/playwright/', // Path to your test files - removed this line to run all playwright tests in the 'src' directory
    timeout: isSelfHosted ? 150000 : 240000, // Timeout of 2 minutes for each test / 4 minutes in case of saas env
    expect: {
        timeout: 5000 // Timeout for Playwright assertions
    },
    use: {
        actionTimeout: 35000
    },
    workers: isLocalRun? 1 : (isSelfHosted ? 5 : 3),
    reporter: [
        ['allure-playwright']
    ],
    projects: [
        {
            name: 'parallel-suite',
            testMatch: ['src/test/suites/playwright/parallel/**/*.test.ts'],
            // NV parallel tests live in parallel-nv-suite
            testIgnore: ['src/test/suites/playwright/parallel/nv/**'],
            retries: retries
        },
        {
            name: 'parallel-nv-suite',
            testMatch: ['src/test/suites/playwright/parallel/nv/**/*.test.ts'],
            retries: 0
        },
        {
            name: 'serial-suite', //for-e2e-ci
            testMatch: ['src/test/suites/playwright/serial/**/*.test.ts'],
            ...(isLocalRun && { //in order not to run tests twice locally, when no test package is sent (unlike ci)
                testIgnore: [
                    'src/test/suites/playwright/serial/regression/**',
                    'src/test/suites/playwright/serial/nv/**'
                ]
            }),
            retries: retries
        },
        {
            name: 'serial-regression',
            testMatch: ['src/test/suites/playwright/serial/regression/**/*.test.ts'],
            retries: 0
        },
        {
            name: 'serial-nv-suite',
            testMatch: ['src/test/suites/playwright/serial/nv/**/*.test.ts'],
            retries: 0
        }
    ],

    // ESM equivalent of require.resolve()
    globalSetup: path.resolve(__dirname, './playwright-global-setup.ts'),
    globalTeardown: path.resolve(__dirname, './playwright-global-teardown.ts')
});
