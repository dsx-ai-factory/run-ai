/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as dotenv from 'dotenv';
import { defineConfig } from '@playwright/test';
import * as path from 'path';

// ESM equivalent of __dirname - using process.cwd() for better compatibility
const __dirname = process.cwd() + '/playwright-config';
dotenv.config();

// Point to the e2e root directory where playwright runs
const e2eRoot = path.resolve(__dirname, '..');
const isSelfHosted = process.env.CLUSTER_NAME_PREFIX
    ? !process.env.CLUSTER_NAME_PREFIX.includes('saas-')
    : false;
const isLocalRun = process.env.LOCAL_RUN ? process.env.LOCAL_RUN === 'true' : false;
const retries = process.env.RETRIES ? Number(process.env.RETRIES) : 2;

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

export default defineConfig({
    testDir: path.resolve(e2eRoot, 'src/test/suites/playwright/'),
    timeout: isSelfHosted ? 150000 : 210000,
    expect: {
        timeout: 5000
    },
    use: {
        actionTimeout: 25000
    },
    workers: isLocalRun ? 1 : (isSelfHosted ? 5 : 3),
    reporter: [
        ['allure-playwright', {
            detail: true,
            suiteTitle: false,
            environmentInfo: {}
        }]
    ],
    projects: [
        {
            name: 'parallel-suite',
            testMatch: [path.resolve(e2eRoot, 'src/test/suites/playwright/parallel/**/*.test.ts')],
            retries: retries
        },
        {
            name: 'serial-regression',
            testMatch: [path.resolve(e2eRoot, 'src/test/suites/playwright/serial/regression/**/*.test.ts')],
            retries: 0
        }
    ],
    globalSetup: path.resolve(__dirname, './runai-workloads-kit-global-setup.ts'),
    globalTeardown: path.resolve(__dirname, './runai-workloads-kit-global-teardown.ts')
});
