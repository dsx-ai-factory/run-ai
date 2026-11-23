/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState, tags, waitUtils } from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';

test.describe(`Datasource PVC E2E Test, tags: ${tags.tool.ui} ${tags.resource.datasource} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_19)(function () {
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let dataSourcePage: DataSourcePage;
        let videoName: string;

        test.beforeAll(function () {
            browserEngine = new BrowserEngine();
        });

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            dataSourcePage = new DataSourcePage(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.system_administrator.email,
                flowRootState.users.system_administrator.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test.describe('PVC data source creation form validation', function () {
            test('T4060299: validate PVC data source name', async () => {
                videoName = 'pvc-name-validation';
                const dataSet = [
                    {
                        name: 'Abcd',
                        valid: false
                    },
                    {
                        name: '1234',
                        valid: false
                    },
                    {
                        name: '-abc',
                        valid: false
                    },
                    {
                        name: 'abc-',
                        valid: false
                    },
                    {
                        name: 'abc$1',
                        valid: false
                    },
                    {
                        name: 'abc1',
                        valid: true
                    },
                    {
                        name: 'abc-1',
                        valid: true
                    },
                    {
                        name: 'a'.repeat(40),
                        valid: true
                    }
                ];

                await test.step('Open PVC creation form', async () => {
                    await dataSourcePage.goToPage();
                    await dataSourcePage.clickOnAddNewDataSourceBtn();
                    await dataSourcePage.clickOnAddNewPvcBtn();
                    await waitUtils.waitForElementToExist(dataSourcePage, dataSourcePage.PvcNameInput);
                });

                for (const testData of dataSet) {
                    await test.step(`validate PVC name: ${testData.name} is ${testData.valid ? 'valid' : 'invalid'}`, async () => {
                        await dataSourcePage.fillPvcNameInput(testData.name);

                        if (!testData.valid) {
                            const alertShown = await dataSourcePage.isDataSourceNameErrorShown();
                            expect(alertShown).toBeTruthy();
                        } else {
                            await waitUtils.waitForElementToNotExist(
                                dataSourcePage,
                                dataSourcePage.dataSourceNameErrorDiv
                            );
                        }

                        await dataSourcePage.clearPvcNameInput();
                    });
                }
            });
        });
    });
});
