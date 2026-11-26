/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from '@playwright/test';
import { createPool, editFromGrid, expectTooltip, test } from './wizardFixtures';

const lockedMessage = /^This can’t be modified once a node pool has been created\.?$/;

test.describe('Node pool wizard', () => {
    test('creation keeps later steps locked until the identity is saved', async ({ wizard: page }) => {
        await expect(page.getByTestId('wizard-step-item-identity')).toHaveAttribute('aria-current', 'step');
        for (const step of ['scheduling', 'topology']) {
            const item = page.getByTestId('wizard-step-item-' + step);
            await expect(item).toBeDisabled();
            await expectTooltip(
                page,
                item.locator('..'),
                'Complete the first step to enable the steps that follow'
            );
        }
        await page.getByTestId('create-node-pool-continue').click();
        await expect(page.getByText('Enter a name', { exact: true })).toBeVisible();
        await expect(page.getByTestId('wizard-step-item-scheduling')).toBeDisabled();

        const name = page.getByRole('textbox', { name: 'Name', exact: true });
        await name.fill('unsaved-wizard-draft');
        const exit = page.getByRole('button', { name: 'Leave this page', exact: true });
        await exit.click();
        await page.getByRole('button', { name: 'Stay On Page', exact: true }).click();
        await expect(name).toHaveValue('unsaved-wizard-draft');
        await exit.click();
        await page.getByRole('button', { name: 'Leave', exact: true }).click();
        await expect(page).toHaveURL((url) => url.pathname === '/node-pools');
    });

    test('create and continue saves ranked scheduling and GPU placement', async ({
        wizard: page,
        poolName
    }) => {
        const id = await createPool(page, poolName, 'continue');
        await expect(page).toHaveURL(
            (url) =>
                url.pathname === '/node-pools/edit/' + id && url.searchParams.get('stepName') === 'scheduling'
        );
        await expect(page.getByTestId('mode-proportional')).toBeChecked();
        await page.getByTestId('mode-ranked').click();
        await page.getByTestId('scheduling-save-continue').click();
        await expect(page.getByTestId('wizard-step-item-topology')).toHaveAttribute('aria-current', 'step');
        await page.getByTestId('topology-placement-toggle').click();
        const spread = page
            .getByTestId('topology-gpu-node-placement')
            .getByRole('radio', { name: 'Spread', exact: true });
        await spread.click();
        await page.getByTestId('topology-save-close').click();
        await editFromGrid(page, poolName, 'Topology & placement');
        await page.getByTestId('topology-placement-toggle').click();
        await expect(spread).toBeChecked();
        await page.getByTestId('wizard-step-item-scheduling').click();
        await expect(page.getByTestId('mode-ranked')).toBeChecked();
    });

    test('create and close allows editing the description and preserves guaranteed quota', async ({
        wizard: page,
        poolName
    }) => {
        await createPool(page, poolName, 'close');
        await editFromGrid(page, poolName, 'Description');
        await expect(page.getByTestId('wizard-step-item-identity')).toHaveAttribute('aria-current', 'step');
        for (const label of ['Cluster', 'Name', 'Key', 'Value']) {
            const field = page.getByRole('textbox', { name: label, exact: true });
            await expect(field).toBeDisabled();
            await expectTooltip(page, page.locator('.locked-field').filter({ has: field }), lockedMessage);
        }
        const description = page.getByRole('textbox', { name: 'Description', exact: true });
        await expect(description).toBeEditable();
        await description.fill('Updated through the wizard');
        await page.getByTestId('identity-save-continue').click();
        await page.getByTestId('mode-guaranteed').click();
        await page.getByTestId('scheduling-save-close').click();
        await editFromGrid(page, poolName, 'Scheduling mode');
        await expect(page.getByTestId('mode-guaranteed')).toBeChecked();
        for (const mode of ['proportional', 'ranked']) {
            const choice = page.getByTestId('mode-' + mode);
            await expect(choice).toBeDisabled();
            await expectTooltip(
                page,
                choice.getByText(mode === 'ranked' ? 'Ranked' : 'Proportional', { exact: true }),
                lockedMessage
            );
        }
        await page.getByTestId('wizard-step-item-identity').click();
        await expect(description).toHaveValue('Updated through the wizard');
    });
});
