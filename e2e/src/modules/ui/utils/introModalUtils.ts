/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { config, logger } from '../../helpers';

async function closeIntroModals(page: Page): Promise<void> {
    await page.waitForURL(config.tenantUrl);

    await page.waitForLoadState('networkidle');
    await page.evaluate(() => localStorage.setItem('app-wide-feature-intro-has-been-shown', 'true'));
    const introModalLocator = page.locator('.intro-modal-key');
    const isVisible = (await introModalLocator.count()) > 0;
    logger.debug(`intro modal key visible? ${isVisible}`);

    if (isVisible) {
        const introModalKey = await introModalLocator.getAttribute('aid');
        logger.debug(`intro modal key: ${introModalKey}`);
        await page.evaluate((storageKey: string) => localStorage.setItem(storageKey, 'true'), introModalKey);
        await page.reload();
    }
}

export const introModalUtils = { closeIntroModals };
