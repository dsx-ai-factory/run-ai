/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    operations,
    tags
} from '../../../../../../../modules/helpers';
import { clusterApi } from '../../../../../../../modules/api';
import { kubeCli } from '../../../../../../../modules/cli';
import { usesKaiNodePoolCrd } from '../../../../../../../modules/helpers/utils/NodePoolCrdFields';
import { NodePool } from '../../../../../../../modules/api/apis/backendClusterApi';
import { test, expect } from '@playwright/test';

const clusterNodePoolsApi = clusterApi.nodepoolsApi;

test.describe(`Advanced Nodepools With Labels, tags:  ${tags.resource.nodepools} ${tags.tool.api} ${tags.teams.viper}`, () => {
    let testNodePoolCreated;
    let testNodePoolLabelKey;
    let testNodePoolLabelValue;
    let testNodeName;

    test.afterAll(async () => {
        try {
            await operations.forceCleanup();
        } catch (e) {
            logger.error('Failed to cleanup after test: ' + e);
        }

        await operations.revertSchedulerResourcesToDefault();
    });

    test.beforeEach(() => {
        testNodePoolCreated = { nodePoolName: '', nodePoolId: '' };
    });

    test.beforeAll(async () => {
        clusterNodePoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;

        await operations.reduceSchedulerResources();
    });

    test.describe('Nodepools enabled in backend and cluster', () => {
        test.beforeAll(async () => {
            await nodePoolUtils.deleteAllClusterNodePools();
        });

        test.afterAll(async () => {
            try {
                await nodePoolUtils.deleteAllClusterNodePools();
            } catch (error) {
                logger.error(`Failed to delete all nodepools in after all, error: ${error}`);
            }
        });

        test.describe('Label nodes to be assigned to nodepool', () => {
            test.beforeEach(() => {
                testNodePoolLabelKey = objectGenerator.generateNameByPrefix('nodepoolkey');
                testNodePoolLabelValue = objectGenerator.generateNameByPrefix('nodepoolvalue');
            });

            test.afterEach(async () => {
                await nodePoolUtils.cleanupDeleteClusterNodePool(testNodePoolCreated);
                testNodePoolCreated = { nodePoolName: '', nodePoolId: '' };

                if (testNodePoolCreated.nodePoolName) {
                    await kubeCli.unLabelNode(testNodeName, testNodePoolLabelKey);
                }
            });

            test('Create nodepool - then label node - see it assigned to nodepool + remove label - node re-assigned', async () => {
                // create nodepool
                const { nodePoolName, nodePoolId } = await nodePoolUtils.createClusterNodePoolAndValidate(
                    testNodePoolLabelKey,
                    testNodePoolLabelValue
                );
                testNodePoolCreated = { nodePoolName, nodePoolId };

                // wait for crd to exist in cluster - with status "Empty"
                await nodePoolUtils.waitForNodePoolCrdStatus(nodePoolName, [NodePool.StatusEnum.Empty]);

                // label a node to be assigned to the nodepool
                testNodeName = await nodePoolUtils.labelAGPUNode(
                    testNodePoolLabelKey,
                    testNodePoolLabelValue
                );
                expect(testNodeName).not.toBe('');

                // wait for crd to be Ready in cluster - it now has a node assigned to it
                let nodePoolCrd = await nodePoolUtils.waitForNodePoolCrdStatus(nodePoolName, [
                    NodePool.StatusEnum.Ready
                ]);
                await nodePoolUtils.validateNodeLabeledWithNodePool(nodePoolName, testNodeName);
                nodePoolUtils.validateNodesStatusOfNodePoolCrd(nodePoolCrd, [testNodeName]);
                const allNodepools = await nodePoolUtils.getAllNodePoolsByClusterAndValidate();
                expect(
                    allNodepools.some((nodepool) => nodepool.name === nodePoolName),
                    'Nodepools list should contain our created nodepool'
                ).toBeTruthy();
                // unlabel node - see it is removed from nodepool
                await kubeCli.unLabelNode(testNodeName, testNodePoolLabelKey);

                // wait for crd to be 'Empty' in cluster - it now does not have a node assigned to it
                nodePoolCrd = await nodePoolUtils.waitForNodePoolCrdStatus(nodePoolName, [
                    NodePool.StatusEnum.Empty
                ]);
                await nodePoolUtils.validateNodeLabeledWithNodePool(
                    nodePoolUtils.DefaultNodePoolName,
                    testNodeName
                );
                await nodePoolUtils.waitForDesiredNodesStatusOfClusterNodePool(nodePoolCrd);
                nodePoolUtils.validateNodesStatusOfNodePoolCrd(nodePoolCrd);
            });

            test('Label node - then create nodepool - see node assigned to nodepool', async () => {
                const { nodePoolName, nodePoolId, nodeName } =
                    await nodePoolUtils.createClusterNodePoolAndLabelNodeToAssignToNodePool(
                        testNodePoolLabelKey,
                        testNodePoolLabelValue
                    );
                testNodePoolCreated = { nodePoolName, nodePoolId };
                testNodeName = nodeName;
            });

            test('Delete nodepool - see labeled node being re-assigned', async () => {
                const { nodePoolName, nodePoolId, nodeName } =
                    await nodePoolUtils.createClusterNodePoolAndLabelNodeToAssignToNodePool(
                        testNodePoolLabelKey,
                        testNodePoolLabelValue
                    );
                testNodePoolCreated = { nodePoolName, nodePoolId };
                testNodeName = nodeName;

                const nodesList = await kubeCli.getNodes();
                const allNodeNames = nodesList.items.map((node) => node.metadata.name);
                const nodeNames = allNodeNames.filter((currentNodeName) => currentNodeName !== nodeName);
                await nodePoolUtils.waitForDefaultNodePoolNodes(nodeNames);

                await clusterNodePoolsApi.deleteNodepool(nodePoolId);
                await nodePoolUtils.waitForNodePoolCrd(nodePoolName, 25, false);

                // validate our node is labeled with default nodepool after our nodepool is deleted
                await nodePoolUtils.validateNodeLabeledWithNodePool(
                    nodePoolUtils.DefaultNodePoolName,
                    nodeName
                );

                await nodePoolUtils.waitForDefaultNodePoolNodes(allNodeNames);
            });

            test('Update nodepool with different labels - see assigned node being re-assigned', async () => {
                // labelKey/labelValue are immutable on the KAI nodepool CRD - the API server rejects
                // any patch that changes them, so this scenario only applies to the legacy run.ai CRD.
                if (usesKaiNodePoolCrd()) {
                    return;
                }
                const { nodePoolName, nodePoolId, nodeName } =
                    await nodePoolUtils.createClusterNodePoolAndLabelNodeToAssignToNodePool(
                        testNodePoolLabelKey,
                        testNodePoolLabelValue
                    );
                testNodePoolCreated = { nodePoolName, nodePoolId };
                testNodeName = nodeName;

                await nodePoolUtils.updateClusterNodePoolLabelsAndValidate(
                    nodePoolId,
                    testNodePoolLabelKey + '1',
                    testNodePoolLabelValue + '1'
                );
                const nodePoolCrd = await nodePoolUtils.waitForNodePoolCrdStatus(
                    nodePoolName,
                    [NodePool.StatusEnum.Empty],
                    40,
                    0.25
                );

                // validate our node is labeled with default nodepool after our nodepool was updated
                await nodePoolUtils.validateNodeLabeledWithNodePool(
                    nodePoolUtils.DefaultNodePoolName,
                    nodeName
                );

                await nodePoolUtils.waitForDesiredNodesStatusOfClusterNodePool(nodePoolCrd);
            });
        });
    });
});
