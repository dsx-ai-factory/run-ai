/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page } from 'playwright';

/**
 * Selects a value in a dropdown by its text.
 * @param page - The Playwright page object.
 * @param dropdownSelector - The selector for the dropdown element.
 * @param valueText - The text of the value to select.
 */
export async function selectDropdownValueByText(
    page: Page,
    dropdownSelector: string,
    valueText: string
): Promise<void> {
    await page.waitForSelector(dropdownSelector, { timeout: 10000 });
    await page.click(dropdownSelector, { force: true });
    const optionSelector = `.runai-select-item:has-text("${valueText}")`;
    await page.waitForSelector(optionSelector, { timeout: 10000 });
    await page.click(optionSelector, { force: true });
}

export const formUtils = { selectDropdownValueByText };
