/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, Page } from '@playwright/test';

export class ElementValidationUtils {
    constructor(private page: Page) {}

    async getTextContent(locator: string): Promise<string> {
        return await this.page.locator(locator).textContent();
    }

    async getInputValue(locator: string): Promise<string> {
        return await this.page.locator(locator).inputValue();
    }

    async getCheckedRadioText(locator: string): Promise<string> {
        const checkedRadio = this.page.locator(locator).getByRole('radio', { checked: true });
        const checkedElText = await checkedRadio.textContent();
        return checkedElText.trim();
    }

    async expectInputValueToEqual(locator: string, value: string): Promise<void> {
        const input: string = await this.getInputValue(locator);
        expect(input).toBe(value);
    }

    async expectInputValueToContain(locator: string, value: string): Promise<void> {
        const input: string = await this.getInputValue(locator);
        expect(input).toContain(value);
    }

    async expectElementValueToContain(locator: string, value: string): Promise<void> {
        const input: string = await this.getTextContent(locator);
        expect(input).toContain(value);
    }

    async expectElementValueToEqual(locator: string, value: string): Promise<void> {
        const input: string = await this.getTextContent(locator);
        expect(input).toBe(value);
    }

    async expectAuthorizedValueToContain(locator: string, value: string): Promise<void> {
        const checkedRadioText = await this.getCheckedRadioText(locator);
        expect(checkedRadioText).toContain(value);
    }
}
