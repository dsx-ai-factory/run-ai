/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { chromium, Browser, BrowserContext, Page } from 'playwright';
import { settings, logger, objectGenerator } from '../helpers';
import fs from 'fs';
import path from 'path';
import { attachment } from 'allure-js-commons';
import { test as playwrightTest } from '@playwright/test';
import { workloadTourUtils } from './utils/guidedTourUtil';

export class BrowserEngine {
    browser: Browser;

    context: BrowserContext;

    page: Page;

    videoFilePath: string;

    videoFileType: string;

    constructor(videoFilePath: string, videoFileType: string) {
        this.videoFilePath = videoFilePath;
        this.videoFileType = videoFileType;
    }

    async start(): Promise<void> {
        logger.debug('launch chromium browser');
        // TODO: Use an existing browser context instead of creating a new one
        this.browser = await chromium.launch({ headless: settings.headlessBrowser });

        this.context = await this.browser.newContext({
            ignoreHTTPSErrors: true,
            recordVideo: {
                dir: settings.browserVideosPath,
                size: { width: 1200, height: 700 }
            },
            permissions: ['clipboard-read', 'clipboard-write']
        });
        this.page = await this.context.newPage();

        await this.setupPageInitScripts();
    }

    private async setupPageInitScripts(): Promise<void> {
        // Skip guided tour dialogs by setting localStorage flags before Vue mounts
        await workloadTourUtils.setupTourSkipScript(this.page);
    }

    async startNonHeadless(): Promise<void> {
        logger.debug('launch chromium browser');
        // TODO: Use an existing browser context instead of creating a new one
        this.browser = await chromium.launch({ headless: false });

        this.context = await this.browser.newContext({
            ignoreHTTPSErrors: true,
            recordVideo: {
                dir: settings.browserVideosPath,
                size: { width: 1200, height: 700 }
            },
            permissions: ['clipboard-read', 'clipboard-write']
        });
        this.page = await this.context.newPage();

        await this.setupPageInitScripts();
    }

    private async stop(): Promise<void> {
        logger.debug('close chromium browser');
        await this.context.close();
        await this.browser.close();
    }

    async stopAndArchive(videoNamePrefix?: string): Promise<void> {
        videoNamePrefix = videoNamePrefix ? videoNamePrefix : this.constructor.name;
        const videoName = objectGenerator.generateNameByPrefix(videoNamePrefix);
        const newVideoFullPath = await this.saveVideo(videoName);
        await this.stop();

        if (!settings.isLocalRun) {
            try {
                playwrightTest.info();
                await attachment(
                    `Video Attachment - ${videoNamePrefix}`,
                    Buffer.from(fs.readFileSync(newVideoFullPath)),
                    'video/webm'
                );
            } catch (error) {
                logger.info(
                    `SSO authentication video saved: ${newVideoFullPath} (not in test context, likely global setup) ${error}`
                );
            }
        }
    }

    private async saveVideo(videoName: string): Promise<string> {
        const videoFileFullName = `${videoName}${this.videoFileType}`;
        const newVideoFullPath = path.join(this.videoFilePath, videoFileFullName);
        logger.info(`save and archive browser videos to ${newVideoFullPath}`);
        const videoPath = await this.page.video().path();
        fs.renameSync(videoPath, newVideoFullPath);
        return newVideoFullPath;
    }
}
