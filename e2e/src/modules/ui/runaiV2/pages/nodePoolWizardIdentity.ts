/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, type Page } from '@playwright/test';

export async function selectNodePoolWizardCluster(page: Page, cluster: string): Promise<void> {
    await page.locator('[aid="cluster-dropdown"]').getByRole('combobox').click();
    const clusterOption = page.getByRole('option').filter({
        hasText: new RegExp('^' + cluster.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + '(?: \\(.*\\))?$')
    });
    await clusterOption.click();
}

export async function fillNodePoolWizardIdentity(
    page: Page,
    cluster: string,
    name: string,
    labelKey: string,
    labelValue: string
): Promise<void> {
    await selectNodePoolWizardCluster(page, cluster);
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
        // Loading options can reset an in-progress search. Retry until the form commits the selection.
        await expect(async () => {
            await input.click();
            await input.press('ControlOrMeta+a');
            await page.keyboard.insertText(value);
            await option.click({ timeout: 3000 });
            if (label === 'Key') {
                await expect(page.getByRole('combobox', { name: 'Value', exact: true })).toBeVisible({
                    timeout: 2000
                });
            } else {
                await input.blur();
                await expect(input).toHaveValue(value);
                await expect(
                    page.getByText('Select a value or press Escape to discard your edit.')
                ).toBeHidden();
            }
        }).toPass({ timeout: 20000 });
    }
}
