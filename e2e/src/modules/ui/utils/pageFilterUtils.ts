/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { config, settings } from '../../helpers';

const selectors = {
    clusterFilterChip: '[aid="cluster-filter-chip"]',
    applyClusterFilterBtn: '[aid="apply-btn"]',
    clusterDropDown: '[aid="cluster-dropdown"]'
};

async function selectClusterFromFilter(page: Page, clusterName: string): Promise<void> {
    await page.waitForSelector(selectors.clusterFilterChip);
    await page.click(selectors.clusterFilterChip);
    await page.click(`[aid="radio-option-${clusterName}"]`);
    await page.click(selectors.applyClusterFilterBtn);
}
async function selectCluster(page: Page, clusterName = config.clusterNamePrefix): Promise<void> {
    if (!settings.isSelfHosted) await selectClusterFromFilter(page, clusterName);
}

export const pageFilterUtils = { selectCluster, selectClusterFromFilter };
