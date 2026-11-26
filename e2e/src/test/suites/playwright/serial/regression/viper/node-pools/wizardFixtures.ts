/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { randomUUID } from 'node:crypto';
import { expect, type Locator, type Page, test as base } from '@playwright/test';
import type { Nodepool } from '../../../../../../../modules/api/apis/clusterApi/generated/model/nodepool';

import { flowRootState } from '../../../../../../../modules/helpers';

const poolsApi = '/api/v1/node-pools';

export const test = base.extend<{ wizard: Page; poolName: string }>({
    wizard: async ({ page }, use) => {
        const { email, password } = flowRootState.users.system_administrator;
        await page.goto('/node-pools');
        await page
            .locator('input[name="email"], input[name="username"], #inputUsername')
            .fill(email, { timeout: 70_000 });
        await page.locator('input[name="password"]').fill(password);
        await page
            .locator(
                '[aid="loginSubmitButton"]:not(#loginWithSSOButton), input[type="submit"], button[type="submit"]'
            )
            .first()
            .click();
        await page.waitForURL(
            (url) => url.pathname === '/node-pools' || /^\/(dashboard|overview)(\/|$)/.test(url.pathname)
        );
        await page.goto('/node-pools');
        await page.addLocatorHandler(
            page.getByRole('button', { name: 'Close onboarding', exact: true }),
            async (button) => button.click()
        );
        await page.getByRole('button', { name: 'New node pool', exact: true }).click();
        await use(page);
    },
    poolName: async ({ request }, use) => {
        const name = 'e2e-wizard-' + randomUUID().slice(0, 18);
        const api = new URL(poolsApi, process.env.CONTROL_PLANE_URL).href;
        const headers = { Authorization: 'Bearer ' + flowRootState.users.system_administrator.accessToken };
        const query = { headers, params: { filterBy: 'name==' + name } };
        try {
            await use(name);
        } finally {
            const result = await request.get(api, query);
            await expect(result).toBeOK();
            const { nodepools = [] }: { nodepools?: Nodepool[] } = await result.json();
            for (const pool of nodepools) {
                expect(pool.name).toBe(name);
                expect(pool.labelKey).toBe('e2e.run.ai/' + name);
                expect(pool.isDefault).toBe(false);
                await expect(
                    await request.delete(api + '/' + encodeURIComponent(pool.id), { headers })
                ).toBeOK();
            }
            await expect
                .poll(
                    async () => {
                        const count = await request.get(api + '/count', query);
                        await expect(count).toBeOK();
                        return (await count.json()).count;
                    },
                    { timeout: 60_000 }
                )
                .toBe(0);
        }
    }
});

export async function createPool(
    page: Page,
    name: string,
    destination: 'close' | 'continue'
): Promise<string> {
    const cluster = flowRootState.cluster.name;
    await page.getByTestId('cluster-dropdown').getByRole('combobox').click();
    const option = page.getByRole('option').filter({
        hasText: new RegExp('^' + cluster.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + '(?: \\(.*\\))?$')
    });
    await option.click();
    await page.getByRole('textbox', { name: 'Name', exact: true }).fill(name);
    await page.getByRole('textbox', { name: 'Description', exact: true }).fill('Wizard E2E');
    const key = 'e2e.run.ai/' + name;
    await page.getByRole('combobox', { name: 'Key', exact: true }).fill(key);
    await page.getByRole('option', { name: key + ' (new key)', exact: true }).click();
    await page.getByRole('combobox', { name: 'Value', exact: true }).fill(name);
    await page.getByRole('option', { name: name + ' (new value)', exact: true }).click();
    const [created] = await Promise.all([
        page.waitForResponse(
            (response) =>
                response.request().method() === 'POST' && new URL(response.url()).pathname === poolsApi
        ),
        page.getByTestId('create-node-pool-' + destination).click()
    ]);
    expect(created.status()).toBe(201);
    const pool = await created.json();
    expect(pool).toMatchObject({ name, labelKey: key, labelValue: name });
    return String(pool.id);
}

export async function editFromGrid(
    page: Page,
    name: string,
    destination: 'Description' | 'Scheduling mode' | 'Topology & placement'
): Promise<void> {
    await expect(page).toHaveURL((url) => url.pathname === '/node-pools');
    const row = page.getByRole('row').filter({ has: page.getByText(name, { exact: true }) });
    await row.getByRole('checkbox').click();
    await page.getByTestId('node-pool-edit-btn').click();
    const editMenu = page.locator('.runai-select-menu');
    await expect(editMenu.getByRole('menuitem')).toHaveText([
        'Description',
        'Scheduling mode',
        'Topology & placement'
    ]);
    await editMenu.getByRole('menuitem', { name: destination, exact: true }).click();
}

export async function expectTooltip(page: Page, target: Locator, message: string | RegExp): Promise<void> {
    await page.mouse.move(0, 0);
    const tooltip = page.getByRole('tooltip').filter({ hasText: message });
    await expect(tooltip).toBeHidden();
    await target.hover();
    await expect(tooltip).toHaveText(message);
}
