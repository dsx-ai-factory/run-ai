/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Page, TestInfo } from '@playwright/test';
import { logger, objectGenerator, settings } from '../../../helpers';
import fs from 'fs';
import path from 'path';
import { isTestFailed } from '../../../helpers/utils/testing/testStateUtils';

export class ToolsBasePage {
    protected page: Page;

    protected title: string;

    protected name: string;

    protected videoFilePath: string;

    protected videoFileType: string;

    protected browserHarsPath: string;

    protected browserHarsFileType: string;

    constructor(page: Page, title: string, name: string) {
        this.page = page;
        this.title = title;
        this.name = name;
        this.videoFilePath = settings.browserVideosPath;
        this.videoFileType = settings.browserVideosFileType;
        this.browserHarsPath = settings.browserHarsPath;
        this.browserHarsFileType = 'zip';
    }

    async checkForPageUrlAndTitle(expectedUrl: string): Promise<void> {
        await this.page.waitForLoadState('networkidle', { timeout: 30000 });
        await this.page.bringToFront();

        await this.checkForPageUrl(expectedUrl);
        await this.checkForPageTitle();
    }

    checkForPageUrl(expectedUrl: string): Promise<void> {
        const currentUrl = this.page.url();
        logger.info(`Page URL: ${currentUrl}`);

        const matches = currentUrl.includes(expectedUrl);
        if (!matches) {
            logger.info(`URL match check: ${matches ? 'PASS' : 'FAIL'}`);
            logger.info(`Expected URL to contain: ${expectedUrl}`);
            logger.info(`Actual: ${currentUrl}`);
            throw new Error(
                `URL validation failed: Expected URL to contain "${expectedUrl}", but got "${currentUrl}"`
            );
        }
        return Promise.resolve();
    }

    async checkForPageTitle(): Promise<void> {
        const pageTitle = await this.page.title();
        if (!pageTitle.includes(this.title)) {
            throw new Error(`Expected page title to contain '${this.title}', but got '${pageTitle}'`);
        }
    }

    async checkForElementVisible(element: string): Promise<void> {
        const elementVisible = await this.page.locator(element).isVisible();
        if (!elementVisible) {
            throw new Error(`Required element not found: ${element}`);
        }
    }

    async checkForElementIncludeText(element: string, text: string): Promise<void> {
        const elementLocator = this.page.locator(element);
        const elementText = (await elementLocator.innerText()).trim();
        if (!elementText.includes(text)) {
            throw new Error(`Required text not found: ${elementText}`);
        }
    }

    async savePageVideo(testSuite: TestInfo): Promise<void> {
        const videoName = objectGenerator.generateNameByPrefix(this.name);
        const videoFileFullName = `${videoName}${this.videoFileType}`;
        const newVideoFullPath = path.join(this.videoFilePath, videoFileFullName);

        try {
            logger.info(`Saving page video to ${newVideoFullPath}`);
            const videoPath = await this.page.video().path();

            fs.renameSync(videoPath, newVideoFullPath);
            logger.info(`Successfully saved page video to ${newVideoFullPath}`);

            // Upload video to Allure if not local run and test failed
            if (!settings.isLocalRun && isTestFailed(testSuite)) {
                // For Playwright tests, use testInfo.attach()
                if (testSuite && typeof testSuite === 'object' && 'attach' in testSuite) {
                    const testInfo = testSuite as TestInfo;
                    await testInfo.attach(`Video - ${this.name}`, {
                        path: newVideoFullPath,
                        contentType: 'video/webm'
                    });
                    logger.info(`Video uploaded to Allure as attachment: Video - ${this.name}`);
                }
            }
        } catch (error) {
            logger.error(`Failed to save page video: ${error}`);
            throw error;
        }
    }
}
