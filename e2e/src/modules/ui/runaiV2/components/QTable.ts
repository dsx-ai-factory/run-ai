/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';
import { BaseV2Page } from '../pages/BaseV2Page';
import { waitUtils } from '../../../helpers';

export class QTable extends BaseV2Page {
    nameCol: string;

    table: string;

    tableBottomDiv: string;

    constructor(page: Page, table: string = 'table.q-table') {
        super(page);

        this.nameCol = '[aid="name"]';
        this.table = table;
        this.tableBottomDiv = '.q-table__bottom';
    }

    async getNames(): Promise<string[]> {
        return await this.page.locator(this.nameCol).allTextContents();
    }

    async toggleEntity(entityName: string): Promise<void> {
        await this.clickOnElementByText(`td${this.nameCol}`, `${entityName}`);
    }

    async getTableData(): Promise<Array<Record<string, string>>> {
        await this.page.waitForLoadState('networkidle');
        await waitUtils.waitForElementToExist(this, this.tableBottomDiv);
        await this.page.waitForTimeout(1000);
        const headers = await this.page.locator(this.table).locator('thead tr th').allInnerTexts();

        const rowHandles = await this.page.locator(this.table).locator('tbody tr').elementHandles();
        const data: string[][] = [];

        for (const rowHandle of rowHandles) {
            const cellHandles = await rowHandle.$$('td');
            const rowData: string[] = [];
            for (const cellHandle of cellHandles) {
                rowData.push(await cellHandle.innerText());
            }
            data.push(rowData);
        }

        const tableData = data.map((row) => {
            const rowObj: Record<string, string> = {};
            headers.forEach((header, idx) => {
                rowObj[header] = row[idx] || '';
            });
            return rowObj;
        });

        return tableData;
    }

    async getEntityData(entityName: string, entityValue: string): Promise<Record<string, string> | null> {
        const rowList = await this.getTableData();
        for (const row of rowList) {
            if (row[entityName] === entityValue) {
                return row;
            }
        }
        return null;
    }
}
