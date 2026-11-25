/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import 'dotenv/config';
import { defineConfig } from '@playwright/test';

export default defineConfig({
    testDir: '../src/test/suites/node-pool-wizard',
    timeout: 120_000,
    expect: { timeout: 15_000 },
    workers: 1,
    forbidOnly: !!process.env.CI,
    use: {
        actionTimeout: 20_000,
        baseURL: process.env.APP_URL || process.env.CONTROL_PLANE_URL,
        testIdAttribute: 'aid'
    }
});
