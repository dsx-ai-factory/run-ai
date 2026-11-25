/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as dotenv from 'dotenv';
import { defineConfig } from '@playwright/test';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

dotenv.config();

// Get current directory in ESM
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const isSelfHosted = process.env.CLUSTER_NAME_PREFIX
    ? !process.env.CLUSTER_NAME_PREFIX.includes('saas-')
    : false;
const retries = process.env.RETRIES ? Number(process.env.RETRIES) : 1;

export default defineConfig({
    testDir: '../src/test/suites/playwright/serial/envHealthCheck/',
    timeout: isSelfHosted ? 240000 : 360000,
    expect: {
        timeout: 15000
    },
    use: {
        actionTimeout: 25000
    },
    workers: 1,
    reporter: [
        ['allure-playwright']
    ],
    projects: [
        {
            name: 'health-check-suite',
            testMatch: ['**/*.test.ts'],
            retries: retries
        }
    ],
    // Use join instead of require.resolve
    globalSetup: join(__dirname, './envHealthCheck-playwright-setup.ts'),
    globalTeardown: join(__dirname, './envHealthCheck-playwright-teardown.ts')
});