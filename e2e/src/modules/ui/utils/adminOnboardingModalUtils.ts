/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';

async function closeAdminOnboardingModal(page: Page): Promise<void> {
    await page.waitForLoadState('networkidle');
    await page.evaluate(() => localStorage.setItem('admin-onboarding-modal-shown', 'true'));
}

export const adminOnboardingModalUtils = { closeAdminOnboardingModal };
