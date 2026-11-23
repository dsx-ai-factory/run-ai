/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { TestInfo } from '@playwright/test';
import { logger } from '../../../helpers';

export function isTestFailed(testSuite: TestInfo | undefined): boolean {
    if (testSuite === undefined) {
        logger.warn('isTestFailed (undefined): defaulting to true');
        return true;
    }
    const isTestFailed = testSuite.status !== 'passed' && testSuite.status !== 'skipped';
    logger.info(
        `isTestFailed (Playwright TestInfo): ${isTestFailed}, status: ${testSuite.status}, expectedStatus: ${testSuite.expectedStatus}`
    );
    return isTestFailed;
}
