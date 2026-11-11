/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState, tags, apiUtils } from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';

import { cliExposerApi } from '../../../../../../modules/api';
import { Page } from 'playwright';
import { CliV2Modal } from '../../../../../../modules/ui/runaiV2/pages/CliV2Modal';
import { test, expect } from '@playwright/test';

test.describe(
    `CLI related UI tests ${tags.tool.ui} ${tags.teams.viper}`,
    { tag: [tags.runOn.parallelCI, tags.services.runaiExposer] },
    () => {
        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let cliModal: CliV2Modal;
        let videoName: string;

        const cliDocDescription = 'A tool for managing Run:ai workloads and monitoring available resources';

        async function initBrowserAndLogin(): Promise<void> {
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        }

        async function validateDoc(
            docPage: Page,
            expectedUrl: string,
            expectedText: string,
            altExpectedUrl?: string
        ): Promise<void> {
            if (altExpectedUrl) {
                const urlExist = docPage.url() === expectedUrl || altExpectedUrl === docPage.url();
                expect(
                    urlExist,
                    `expected to navigate to ${expectedUrl} or ${altExpectedUrl}, instead of ${docPage.url()}`
                ).toBeTruthy();
            } else {
                expect(
                    docPage.url(),
                    `expected to navigate to ${expectedUrl}, instead of ${docPage.url()}`
                ).toBe(expectedUrl);
            }
            const containsText = await docPage.evaluate((text) => {
                return document.body.innerText.includes(text);
            }, expectedText);
            expect(containsText, `expected to find "${expectedText}" in the page`).toBeTruthy();
        }

        test.beforeAll(async function () {
            await apiUtils.enableUiCliV2();
        });

        test.beforeEach(async function () {
            await initBrowserAndLogin();
            cliModal = new CliV2Modal(browserEngine);
            await cliModal.openCliDialog();
        });

        test.afterEach(async function () {
            await browserEngine.stopAndArchive(this, 'ff-mixed-' + videoName);
        });
        test('Validate CLI V2 download command', async function () {
            videoName = 'download-command';
            const uiInstallCommand = await cliModal.getLinuxInstallCommand();
            const apiInstallCommand = (await cliExposerApi.cliApi.getInstallerUnixCommands()).body.command;
            expect(apiInstallCommand.slice(0, 73)).toBe(uiInstallCommand.slice(0, 73));
        });
        test('Validate CLI V2 docs', async function () {
            videoName = 'cli-docs';
            const docPage = await cliModal.openCliDoc();
            const apiDocUrl = (await cliExposerApi.cliApi.getManualDocument('runai.md')).response['request']
                .uri.href;
            await validateDoc(docPage, apiDocUrl, cliDocDescription);
        });
    }
);
