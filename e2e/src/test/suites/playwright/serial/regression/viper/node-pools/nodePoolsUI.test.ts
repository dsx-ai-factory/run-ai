/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// External modules
import { test, expect } from '@playwright/test';

// Helpers
import {
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    operations,
    tags,
    versionUtils,
    waitUtils
} from '../../../../../../../modules/helpers';

// UI Pages
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { NodePoolPage } from '../../../../../../../modules/ui/runaiV2/pages/NodePoolPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';

// Constants and Types
import { DEFAULT_NODE_POOL_NAME } from '../../../../../../../modules/ui/utils/nodePoolUtils';
import {
    NodePoolCreateRequest,
    PlacementStrategy
} from '../../../../../../../modules/api/apis/backendClusterApi';
import { MIN_PLACEMENT_VERSION } from '../../../../../../../modules/helpers/utils/NodePoolUtils';
import { getNodePoolCrdField } from '../../../../../../../modules/helpers/utils/NodePoolCrdFields';
import { getNodePoolsApi } from '../../../../../../../modules/helpers/config/const';
import { clusterApi } from '../../../../../../../modules/api';
import CpuEnum = PlacementStrategy.CpuEnum;
import GpuEnum = PlacementStrategy.GpuEnum;

test.describe(`Node pools UI, tags: ${tags.tool.ui} ${tags.resource.nodepools} ${tags.teams.viper}`, () => {
    let browserEngine: BrowserEngine;
    let authPage: AuthPage;
    let videoName: string;
    let nodePoolPage: NodePoolPage;

    test.beforeAll(async () => {
        getNodePoolsApi().accessToken = flowRootState.users.system_administrator.accessToken;
        clusterApi.nodepoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        browserEngine = new BrowserEngine();

        await operations.reduceSchedulerResources();
    });

    test.afterAll(async () => {
        await nodePoolUtils.deleteAllClusterNodePools();
        await operations.revertSchedulerResourcesToDefault();
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

    test.describe('node pool create / edit form', () => {
        videoName = 'crate-edit-node-pool';
        let nodePoolRequest: NodePoolCreateRequest = {};

        test.beforeAll(async () => {
            await startSession();
            await nodePoolPage.navigateToNodePools();
            nodePoolRequest = {
                name: objectGenerator.generateNameByPrefix('node-pool-a'),
                labelKey: objectGenerator.generateNameByPrefix('nvidia.com/gpu.product'),
                labelValue: objectGenerator.generateNameByPrefix('T4')
            };

            await nodePoolPage.createBasicNodePool(nodePoolRequest);
            await nodePoolPage.v2frame.waitForSelector(nodePoolPage.nodePoolNameCol);
        });

        test.afterAll(async () => {
            await browserEngine.stopAndArchive(undefined, 'nodepools-' + videoName);
        });

        test('should create node pool and validate that node pool has been added to the table', async () => {
            expect(await nodePoolPage.isTableIncludeNodePool(nodePoolRequest.name)).toBe(true);
        });

        test('should create node pool with placementStrategy and edit it', async () => {
            const nodePoolIndexRow = await nodePoolPage.getNodePoolRowIndexByName(nodePoolRequest.name);
            await nodePoolPage.waitForEmptyStatus(nodePoolIndexRow);
            await nodePoolPage.clickOnNodePoolRowByName(nodePoolRequest.name);
            await nodePoolPage.clickOnEditButton();
            if (versionUtils.isVersionSupported(flowRootState.cluster.version, MIN_PLACEMENT_VERSION)) {
                await nodePoolPage.openSchedulingSection();
                await nodePoolPage.fillCpuPlacementRadio(CpuEnum.Binpack);
                await nodePoolPage.fillGpuPlacementRadio(GpuEnum.Spread);
            }

            await nodePoolPage.submitNodePoolForm();

            logger.debug(`Node pool ${nodePoolRequest.name} edited successfully`);

            if (versionUtils.isVersionSupported(flowRootState.cluster.version, MIN_PLACEMENT_VERSION)) {
                // Wait for the CRD to be updated with the new placement strategy values
                await nodePoolUtils.waitForNodePoolCrdProperty(
                    nodePoolRequest.name,
                    'spec.placementStrategy.cpu',
                    (value) => value === PlacementStrategy.CpuEnum.Binpack.toString(),
                    40
                );
                await nodePoolUtils.waitForNodePoolCrdProperty(
                    nodePoolRequest.name,
                    'spec.placementStrategy.gpu',
                    (value) => value === PlacementStrategy.GpuEnum.Spread.toString(),
                    40
                );

                // Now get the final CRD state and validate
                const nodePoolCrd = await nodePoolUtils.waitForNodePoolCrdReadyOrEmpty(
                    nodePoolRequest.name,
                    40
                );
                expect(getNodePoolCrdField(nodePoolCrd, 'spec.placementStrategy.cpu')).toBe(
                    PlacementStrategy.CpuEnum.Binpack.toString()
                );
                expect(getNodePoolCrdField(nodePoolCrd, 'spec.placementStrategy.gpu')).toBe(
                    PlacementStrategy.GpuEnum.Spread.toString()
                );
            }
        });

        test('should create node pool and delete it', async () => {
            const nodePoolIndexRow = await nodePoolPage.getNodePoolRowIndexByName(nodePoolRequest.name);
            await nodePoolPage.waitForEmptyStatus(nodePoolIndexRow);
            await nodePoolPage.deleteNodePoolByName(nodePoolRequest.name);
            await nodePoolPage.waitForDeletingStatus(nodePoolIndexRow);
            const isNodePoolDeleted = async (): Promise<boolean> =>
                nodePoolPage.isTableNotIncludeNodePool(nodePoolRequest.name);

            await waitUtils.waitForCondition(isNodePoolDeleted, 'wait for node pool to remove from table');
        });
    });
});
