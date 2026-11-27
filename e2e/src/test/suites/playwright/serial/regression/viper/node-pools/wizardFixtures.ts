/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, type Locator, type Page, test as base } from '@playwright/test';
import { NodepoolPhase } from '../../../../../../../modules/api/apis/clusterApi/generated/model/nodepoolPhase';
import type { Nodepool } from '../../../../../../../modules/api/apis/clusterApi/generated/model/nodepool';

import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { NodePoolPage } from '../../../../../../../modules/ui/runaiV2/pages/NodePoolPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { flowRootState, objectGenerator } from '../../../../../../../modules/helpers';

const poolsApi = '/api/v1/node-pools';

export const test = base.extend<{ wizard: Page; poolName: string }>({
    wizard: async ({}, use, testInfo) => {
        const browser = new BrowserEngine();
        const page = await browser.start();
        try {
            const { email, password } = flowRootState.users.system_administrator;
            await new AuthPage(page).signIn(email, password);
            const pools = new NodePoolPage(page);
            await pools.navigateToNodePools();
            await page.locator(pools.addNewNodePoolBtn).click();
            await use(page);
        } finally {
            await browser.stopAndArchive(testInfo, 'node-pools-ui-wizard');
        }
    },
    poolName: async ({ request }, use) => {
        const name = objectGenerator.generateNameByPrefix('e2e-wizard');
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
                if (pool.phase === NodepoolPhase.Deleting) continue;
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
