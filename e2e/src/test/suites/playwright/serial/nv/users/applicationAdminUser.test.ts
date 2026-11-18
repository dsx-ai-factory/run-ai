/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState, tags } from '../../../../../../modules/helpers';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { Utils } from '../../../../../../modules/helpers';
import {
    shouldRunTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, expect } from '@playwright/test';
import { MENU_SUBJECT_NAMES } from '../../../../../../modules/ui/utils/sidenavUtils';
import { NodePage } from '../../../../../../modules/ui/runaiV2/pages/NodePage';

test.describe(`DGXC Application administrator User - sanity, tags: ${tags.runOn.dgxc} ${tags.teams.loong} ${tags.tool.ui} ${tags.resource.users}`, () => {
    shouldRunTests(Utils.isDGXC())(function () {
        supportedFrom(SupportedVersions.V2_18)(function () {
            let browserEngine: BrowserEngine;
            let authPage: AuthPage;
            let nodePage: NodePage;

            let videoName: string;

            test.beforeAll(function () {
                browserEngine = new BrowserEngine();
            });

            test.beforeEach(async function () {
                await browserEngine.start();
                authPage = new AuthPage(browserEngine.currentPage);
                nodePage = new NodePage(browserEngine.currentPage);

                await authPage.signIn(
                    flowRootState.users.application_administrator.email,
                    flowRootState.users.application_administrator.password
                );
            });

            test.afterEach('stop browser', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, videoName);
            });

            test('T3738321: Application adminstrator user can succeed to login', async () => {
                videoName = 'application-admin-user-login';

                await expect(authPage.page.locator(`[aid="${MENU_SUBJECT_NAMES.Analytics}"]`)).toBeVisible();
            });

            test('T3738322: Application adminstrator user can view node info', async () => {
                videoName = 'application-admin-user-view-node';
                let nodeList = [];

                await nodePage.goToPage();

                await test.step('verify all nodes in table are in ready status', async () => {
                    nodeList = await nodePage.getAllNodesData();
                    expect(nodeList.length).toBeGreaterThan(1);

                    for (const node of nodeList) {
                        expect(node['Status']).toEqual('Ready');
                    }
                });

                await test.step('verify node details page', async () => {
                    for (const node of nodeList) {
                        const isGpuNode = Number(node['GPU type']) > 0;
                        await nodePage.toggleNode(node['Node']);
                        await nodePage.clickShowDetailsBtn();
                        await nodePage.waitForNodeDetailsChartsToExist(isGpuNode);
                        await nodePage.closeDetailsSidePage();
                    }
                });
            });
        });
    });
});
