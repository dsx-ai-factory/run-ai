/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { chromium, Browser, BrowserContext, Page, Response } from 'playwright';
import { TestInfo } from '@playwright/test';
import { settings, logger, objectGenerator } from '../../../helpers';
import fs from 'fs';
import path from 'path';
import { isTestFailed } from '../../../helpers/utils/testing/testStateUtils';
import { workloadTourUtils } from '../../utils/guidedTourUtil';

export class BrowserEngine {
    browser: Browser;

    context: BrowserContext;

    page: Page;

    videoFilePath: string;

    videoFileType: string;

    browserHarsPath: string;

    browserHarsFileType: string;

    private isTracingActive: boolean;

    constructor() {
        this.videoFilePath = settings.browserVideosPath;
        this.videoFileType = settings.browserVideosFileType;

        this.browserHarsFileType = 'zip';
        const harFileName: string = objectGenerator.generateNameByPrefix('harfile');
        this.browserHarsPath = `${settings.browserHarsPath}${harFileName}.${this.browserHarsFileType}`;

        this.isTracingActive = false;
    }

    async start(headlessBrowser: boolean = null): Promise<Page> {
        logger.debug('launch chromium browser');
        // TODO: Use an existing browser context instead of creating a new one
        const headlessMode = headlessBrowser !== null ? headlessBrowser : settings.headlessBrowser;

        this.browser = await chromium.launch({ headless: headlessMode });

        this.context = await this.browser.newContext({
            ignoreHTTPSErrors: true,
            recordHar: {
                path: this.browserHarsPath,
                content: 'embed'
            },
            recordVideo: {
                dir: settings.browserVideosPath,
                size: { width: 1200, height: 700 }
            },
            permissions: ['clipboard-read', 'clipboard-write']
        });
        this.page = await this.context.newPage();

        await this.setupPageInitScripts();

        return this.page;
    }

    private async setupPageInitScripts(): Promise<void> {
        // Skip guided tour dialogs by setting localStorage flags before Vue mounts
        await workloadTourUtils.setupTourSkipScript(this.page);
    }

    get currentPage(): Page {
        return this.page;
    }

    async pause(): Promise<void> {
        return this.page.pause();
    }

    async refresh(): Promise<Response> {
        return this.page.reload();
    }

    private async stop(): Promise<void> {
        logger.debug('close chromium browser');
        await this.context.close();
        await this.browser.close();
    }

    async stopAndArchive(testSuite: TestInfo, videoNamePrefix: string): Promise<void> {
        const videoName = objectGenerator.generateNameByPrefix(videoNamePrefix);
        const newVideoFullPath = await this.saveVideo(videoName);
        await this.stop();

        const testFailed = isTestFailed(testSuite);
        logger.info(
            `Video recording - Test failed: ${testFailed}, isLocalRun: ${settings.isLocalRun}, Video path: ${newVideoFullPath}`
        );

        if (!settings.isLocalRun && testFailed) {
            logger.info(`Attaching video to Allure report: ${videoNamePrefix}`);

            try {
                // For Playwright tests, use testInfo.attach()
                if (testSuite && typeof testSuite === 'object' && 'attach' in testSuite) {
                    const testInfo = testSuite as TestInfo;
                    await testInfo.attach(`Video - ${videoNamePrefix}`, {
                        path: newVideoFullPath,
                        contentType: 'video/webm'
                    });

                    const newHarFullPath = this.saveHar(videoName);
                    await testInfo.attach(`Network Activity - ${videoNamePrefix}`, {
                        path: newHarFullPath,
                        contentType: 'application/zip'
                    });

                    logger.info('Successfully attached video and HAR to Allure via testInfo');
                } else {
                    logger.warn('testInfo.attach not available, skipping Allure attachment');
                }
            } catch (error) {
                logger.error(`Failed to attach video to Allure: ${error}`);
            }
        } else {
            logger.info(
                `Video not attached to Allure (isLocalRun: ${settings.isLocalRun}, testFailed: ${testFailed})`
            );
        }
    }

    async startTracing(traceName: string): Promise<void> {
        try {
            await this.context.tracing.start({
                screenshots: true,
                snapshots: true,
                sources: true
            });

            this.isTracingActive = true;
            logger.info(`Tracing started for test: ${traceName}`);
        } catch (error) {
            logger.error(`Failed to start tracing for ${traceName}: ${error.message}`);
            this.isTracingActive = false;
        }
    }

    async stopAndArchiveTracing(testSuite: TestInfo, traceName: string): Promise<void> {
        if (!this.isTracingActive) {
            logger.warn(`Tracing was not active for ${traceName}, skipping archive`);
            return;
        }

        try {
            const traceFileName = `${traceName}.zip`;
            const traceFullPath = path.join(settings.browserVideosPath, traceFileName);

            await this.context.tracing.stop({ path: traceFullPath });
            this.isTracingActive = false;
            logger.info(`Tracing stopped and archived to: ${traceFullPath}`);

            const testFailed = isTestFailed(testSuite);
            logger.info(
                `Trace recording - Test failed: ${testFailed}, isLocalRun: ${settings.isLocalRun}, Trace path: ${traceFullPath}`
            );

            if (!settings.isLocalRun && testFailed) {
                logger.info(`Attaching trace to Allure report: ${traceName}`);

                try {
                    // For Playwright tests, use testInfo.attach()
                    if (testSuite && typeof testSuite === 'object' && 'attach' in testSuite) {
                        const testInfo = testSuite as TestInfo;
                        await testInfo.attach(`Trace - ${traceName}`, {
                            path: traceFullPath,
                            contentType: 'application/zip'
                        });

                        logger.info('Successfully attached trace to Allure via testInfo');
                    } else {
                        logger.warn('testInfo.attach not available, skipping Allure attachment');
                    }
                } catch (error) {
                    logger.error(`Failed to attach trace to Allure: ${error}`);
                }
            } else {
                logger.info(
                    `Trace not attached to Allure (isLocalRun: ${settings.isLocalRun}, testFailed: ${testFailed})`
                );
            }
        } catch (error) {
            // If tracing couldn't be stopped, just log and continue
            logger.error(`Could not stop tracing for ${traceName}: ${error.message}`);
            this.isTracingActive = false;
        }
    }

    private async saveVideo(videoName: string): Promise<string> {
        const videoFileFullName = `${videoName}${this.videoFileType}`;
        const newVideoFullPath = path.join(this.videoFilePath, videoFileFullName);
        logger.info(`save and archive browser videos to ${newVideoFullPath}`);

        // Check if video recording is enabled (may not be if using startWithTracing)
        const video = this.page.video();
        if (!video) {
            logger.warn('No video recording active - skipping video save (may be using tracing instead)');
            return newVideoFullPath;
        }

        const videoPath = await video.path();
        fs.renameSync(videoPath, newVideoFullPath);
        return newVideoFullPath;
    }

    private saveHar(videoName: string): string {
        const currentHarsFilename = `${videoName}.${this.browserHarsFileType}`;
        const harDirectory: string = this.browserHarsPath.split('/')[0];
        const newHarFullPath = `${harDirectory}/${currentHarsFilename}`;
        fs.renameSync(this.browserHarsPath, newHarFullPath);
        return newHarFullPath;
    }
}
