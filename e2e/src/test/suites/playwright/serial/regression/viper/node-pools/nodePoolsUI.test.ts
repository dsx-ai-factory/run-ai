/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// External modules
import { test, expect } from '@playwright/test';

// Helpers
import { flowRootState, tags } from '../../../../../../../modules/helpers';

// UI Pages
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { NodePoolPage } from '../../../../../../../modules/ui/runaiV2/pages/NodePoolPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';

// Constants and Types
import { DEFAULT_NODE_POOL_NAME } from '../../../../../../../modules/ui/utils/nodePoolUtils';

test.describe(`Node pools UI, tags: ${tags.tool.ui} ${tags.resource.nodepools} ${tags.teams.viper}`, () => {
    let browserEngine: BrowserEngine;
    let authPage: AuthPage;
    let videoName: string;
    let nodePoolPage: NodePoolPage;

    test.beforeAll(() => {
        browserEngine = new BrowserEngine();
    });

    const startSession = async (): Promise<void> => {
        await browserEngine.start();
        authPage = new AuthPage(browserEngine.currentPage);
        nodePoolPage = new NodePoolPage(browserEngine.currentPage);
        await authPage.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
    };

    test.describe('node pool table', () => {
        videoName = 'node pool table';

        test.beforeAll(async () => {
            await startSession();
            await nodePoolPage.navigateToNodePools();
        });

        test.afterAll(async () => {
            await browserEngine.stopAndArchive(undefined, 'nodepools-' + videoName);
        });

        test('should redirect to node pool page', async () => {
            const title = await nodePoolPage.getPageTitle();
            expect(title).toContain('Node pools');
        });

        test('should validate that default node pool exists', async () => {
            await nodePoolPage.v2frame.waitForSelector(nodePoolPage.nodePoolNameCol);
            expect(await nodePoolPage.isTableIncludeNodePool(DEFAULT_NODE_POOL_NAME)).toBe(true);
        });

        test('should toggle default node pool nodes modal', async () => {
            const nodePoolIndexRow = await nodePoolPage.getNodePoolRowIndexByName(DEFAULT_NODE_POOL_NAME);
            await nodePoolPage.clickOnCellByRowIndex(nodePoolPage.nodePoolNodesCol, nodePoolIndexRow);
            expect(await nodePoolPage.getModalHeader()).toBe('Nodes Associated with Node pool default');
        });
    });
});
