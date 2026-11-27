/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';

const TOUR_SEEN_KEYS = [
    'workload-tour-seen',
    'overview-tour-seen',
    'project-tour-seen',
    'department-tour-seen'
];

async function closeWorkloadTour(page: Page): Promise<void> {
    await page.waitForLoadState('networkidle');
    await page.evaluate((keys) => {
        keys.forEach((key) => localStorage.setItem(key, 'true'));
    }, TOUR_SEEN_KEYS);
}

async function setupTourSkipScript(page: Page): Promise<void> {
    await page.addInitScript((keys) => {
        keys.forEach((key) => localStorage.setItem(key, 'true'));
    }, TOUR_SEEN_KEYS);
}

export const workloadTourUtils = { closeWorkloadTour, setupTourSkipScript };
