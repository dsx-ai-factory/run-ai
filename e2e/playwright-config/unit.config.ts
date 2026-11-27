/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineConfig } from '@playwright/test';

export default defineConfig({
    testDir: '../src/test/unit',
    timeout: 10000,
    workers: 1,
    reporter: [['line']],
    projects: [
        {
            name: 'unit',
            testMatch: ['**/*.test.ts']
        }
    ]
});
