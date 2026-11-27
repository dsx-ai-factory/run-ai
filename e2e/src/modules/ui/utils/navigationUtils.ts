/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { logger } from '../../helpers/logger/logger';
import { BaseV2Page } from '../runaiV2/pages/BaseV2Page';
import { expect } from 'chai';

export class NavigateUtils {
    static navigateAndValidate = async (page: Page, url: string, selector: string): Promise<void> => {
        logger.debug('Navigate to URL and validate selector');
        await page.waitForLoadState('networkidle');
        await page.goto(url, { timeout: 10000 });
        await page.waitForSelector(selector, { timeout: 10000 });
    };

    static navigateAndValidateError = async (page: Page, url: string): Promise<boolean> => {
        logger.debug('Navigate to URL and error');
        await page.waitForLoadState('networkidle');
        try {
            const res = await page.goto(url, { timeout: 10000 });
            return !res.ok();
        } catch (error) {
            return true;
        }
    };

    async varifyPageTitle(page: BaseV2Page, title: string): Promise<void> {
        const pageTitle = await page.getPageTitle();
        expect(pageTitle).contains(
            title,
            `Page title is not as expected. Expected: ${title}, Actual: ${pageTitle}`
        );
    }
}
