/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, type Page } from '@playwright/test';

export async function fillNodePoolWizardIdentity(
    page: Page,
    cluster: string,
    name: string,
    labelKey: string,
    labelValue: string
): Promise<void> {
    await page.locator('[aid="cluster-dropdown"]').getByRole('combobox').click();
    const clusterOption = page.getByRole('option').filter({
        hasText: new RegExp('^' + cluster.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + '(?: \\(.*\\))?$')
    });
    await clusterOption.click();
    await page.getByRole('textbox', { name: 'Name', exact: true }).fill(name);
    for (const [label, value] of [
        ['Key', labelKey],
        ['Value', labelValue]
    ]) {
        const input = page.getByRole('combobox', { name: label, exact: true });
        const existing = page.getByRole('option', { name: value, exact: true });
        const create = page.getByRole('option', {
            name: value + ' (new ' + label.toLowerCase() + ')',
            exact: true
        });
        const option = existing.or(create).first();
        await input.click();
        await input.fill(value);
        // Opening the label picker can reset its search while the initial options load.
        await expect(async () => {
            if (await option.isVisible()) return;
            await input.press('ControlOrMeta+a');
            await page.keyboard.insertText(value);
            if ((await input.getAttribute('aria-expanded')) !== 'true') await input.press('ArrowDown');
            await expect(option).toBeVisible({ timeout: 2500 });
        }).toPass({ timeout: 15000 });
        await option.click();
    }
}
