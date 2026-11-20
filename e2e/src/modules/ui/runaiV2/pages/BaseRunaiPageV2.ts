/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseRunaiPage } from '../../runaiV1/pages/BaseRunaiPage';
import { FrameLocator, Locator } from 'playwright';

export class BaseRunaiPageV2 extends BaseRunaiPage {
    // global
    v2iframe = '[id=runai-new-app]';

    mainPageTab: string;

    mainPageTitle: string;

    url(): string {
        return this.page.url();
    }

    async pause(): Promise<void> {
        return this.page.pause();
    }

    protected iframePage(): FrameLocator {
        return this.page.frameLocator(this.v2iframe);
    }

    async getPageTitle(): Promise<string> {
        return await this.page.locator(this.mainPageTitle).textContent();
    }

    findByText(element: string, text: string): Locator {
        const v2frame: FrameLocator = this.iframePage();
        return v2frame.locator(`${element}:has-text("${text}")`);
    }

    async clickOnElementByText(element: string, text: string): Promise<void> {
        const v2frame: FrameLocator = this.iframePage();
        return await v2frame.locator(`${element}:has-text("${text}")`).click();
    }

    async navigateToPage(name: string): Promise<void> {
        await this.page.locator(name).click();
        await this.page.locator(this.mainPageTitle).click();
    }
}
