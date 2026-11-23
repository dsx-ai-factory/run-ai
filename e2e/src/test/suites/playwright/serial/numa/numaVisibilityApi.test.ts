/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    flowRootState,
    logger,
    nodePoolUtils,
    nodeUtils,
    numaUtils,
    settings,
    tags,
    waitUtils
} from '../../../../../modules/helpers';
import { shouldRunTests } from '../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { clusterApi } from '../../../../../modules/api';
import { kubeCli } from '../../../../../modules/cli';
import {
    Node,
    NodeExtended,
    NodepoolPhase,
    TopologyManagerPolicy
} from '../../../../../modules/api/apis/clusterApi';
import { NumaTestNodePool, NumaUtils } from '../../../../../modules/helpers/utils/NumaUtils';

test.describe(
    `NUMA-aware scheduling visibility API, tags: ${tags.resource.nodepools} ${tags.resource.nodes} ${tags.tool.api} ${tags.runOn.numa} ${tags.teams.viper}`,
    { tag: [tags.services.nodepoolController, tags.services.clusterSync, tags.services.clusterService] },
    () => {
        shouldRunTests(settings.isNumaWorkflow)(() => {
            const expectedPolicy = TopologyManagerPolicy.SingleNumaNode;
            let nodePool: NumaTestNodePool;

            test.beforeAll(() => {
                clusterApi.nodepoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;
            });

            test.afterEach(async () => {
                try {
                    await numaUtils.deleteNumaTestNodePool(nodePool);
                } catch (error) {
                    logger.error(`Failed to clean up node pool ${nodePool?.nodePoolName}`, error);
                }
                nodePool = undefined;
            });

            test('Enabling NUMA-aware scheduling exposes per-node NUMA data', async () => {
                await test.step('Create a node pool with a single GPU node', async () => {
                    nodePool = await numaUtils.createNodePoolForGpuNode();
                    expect(await numaUtils.getNodeNamesWithNrt()).toContain(nodePool.nodeName);
                });

                await test.step('Topology Manager policy is not reported while NUMA is disabled', async () => {
                    const node = await nodeUtils.getNodeByName(nodePool.nodeName);
                    expect(node.numa?.topologyManagerPolicy ?? null).toBeNull();
                });

                await test.step('Enable NUMA-aware scheduling on the node pool', async () => {
                    await nodePoolUtils.setNumaAwareScheduling(
                        nodePool.nodePoolId,
                        nodePool.labelKey,
                        nodePool.labelValue,
                        true
                    );
                    const updated = await clusterApi.nodepoolsApi.getNodepool(nodePool.nodePoolId);
                    expect(updated.body.schedulingConfiguration?.numaAwareScheduling).toBe(true);
                });

                await test.step('Node reports the Topology Manager policy', async () => {
                    await nodeUtils.waitForNodeTopologyManagerPolicy(nodePool.nodeName, expectedPolicy);
                });

                await test.step('Node pool and node stay Ready', async () => {
                    await nodePoolUtils.waitForClusterNodePoolPhase(nodePool.nodePoolId, [
                        NodepoolPhase.Ready
                    ]);
                    await nodeUtils.getNodeStatusAndValidate(nodePool.nodeName, Node.StatusEnum.Ready);
                });

                await test.step('GET node by id returns the per-NUMA-node resources', async () => {
                    const nodeGpus = await kubeCli.getNumOfGpusOnNode(nodePool.nodeName);
                    // cluster-sync publishes the per-NUMA-node data separately from the node itself
                    let node: NodeExtended;
                    await waitUtils.waitForCondition(
                        async () => {
                            node = await nodeUtils.getNodeExtendedByName(nodePool.nodeName);
                            return (node.numa?.nodes ?? []).length > 0;
                        },
                        `wait for per-NUMA-node data on node ${nodePool.nodeName}`,
                        true,
                        120,
                        5
                    );
                    logger.info(`node ${node.name} numa: ${JSON.stringify(node.numa)}`);

                    expect(node.numa?.topologyManagerPolicy).toEqual(expectedPolicy);
                    const numaNodes = node.numa?.nodes ?? [];
                    expect(numaNodes.length).toBeGreaterThan(0);
                    expect(new Set(numaNodes.map((numaNode) => numaNode.name)).size).toEqual(
                        numaNodes.length
                    );

                    const totalGpus = numaNodes.reduce((sum, numaNode) => sum + numaNode.totalGpus, 0);
                    expect(totalGpus).toEqual(nodeGpus);

                    for (const numaNode of numaNodes) {
                        expect(numaNode.totalCpuCores).toBeGreaterThan(0);
                        expect(numaNode.totalCpuMemoryBytes).toBeGreaterThan(0);
                        expect(numaNode.allocatedGpus).toBeLessThanOrEqual(numaNode.totalGpus);
                        expect(numaNode.allocatedCpuCores).toBeLessThanOrEqual(numaNode.totalCpuCores);
                        expect(numaNode.allocatedCpuMemoryBytes).toBeLessThanOrEqual(
                            numaNode.totalCpuMemoryBytes
                        );
                    }
                });

                await test.step('Disabling NUMA-aware scheduling clears the policy', async () => {
                    await nodePoolUtils.setNumaAwareScheduling(
                        nodePool.nodePoolId,
                        nodePool.labelKey,
                        nodePool.labelValue,
                        false
                    );
                    await nodeUtils.waitForNodeTopologyManagerPolicy(nodePool.nodeName, null);
                    await nodePoolUtils.waitForClusterNodePoolPhase(nodePool.nodePoolId, [
                        NodepoolPhase.Ready
                    ]);
                });
            });

            test('Node without NRT data is reported as missing prerequisites', async () => {
                const nodeName = await numaUtils.findNodeWithoutNrt();
                test.skip(!nodeName, 'No Ready node without GPUs and without NRT data in the cluster');

                await test.step(`Create a node pool with node ${nodeName}`, async () => {
                    nodePool = await numaUtils.createNodePoolForNode(nodeName);
                });

                await test.step('Enable NUMA-aware scheduling on the node pool', async () => {
                    await nodePoolUtils.setNumaAwareScheduling(
                        nodePool.nodePoolId,
                        nodePool.labelKey,
                        nodePool.labelValue,
                        true
                    );
                });

                await test.step('Node pool moves to MissingPrerequisites', async () => {
                    const updated = await nodePoolUtils.waitForClusterNodePoolPhase(
                        nodePool.nodePoolId,
                        [NodepoolPhase.MissingPrerequisites],
                        180
                    );
                    logger.info(`node pool status: ${JSON.stringify(updated.status)}`);
                    const condition = updated.status?.conditions?.find(
                        (c) => c.type === NumaUtils.missingNrtConditionType
                    );
                    expect(
                        condition,
                        `node pool should report ${NumaUtils.missingNrtConditionType}`
                    ).toBeDefined();
                    expect(condition.message ?? '').toContain(NumaUtils.missingNrtObjectMessage);
                    expect(updated.status?.nodes ?? []).toContain(nodeName);
                });

                await test.step('Node moves to MissingPrerequisites', async () => {
                    await nodeUtils.waitForNodeStatus(nodeName, Node.StatusEnum.MissingPrerequisites, 180);
                    const node = await nodeUtils.getNodeByName(nodeName);
                    const condition = node.conditions?.find(
                        (c) => c.type === NumaUtils.missingNrtConditionType
                    );
                    expect(
                        condition,
                        `node should report ${NumaUtils.missingNrtConditionType}`
                    ).toBeDefined();
                    expect(condition.message ?? '').toContain(NumaUtils.missingNrtObjectMessage);
                });

                await test.step('Disabling NUMA-aware scheduling returns both to Ready', async () => {
                    await nodePoolUtils.setNumaAwareScheduling(
                        nodePool.nodePoolId,
                        nodePool.labelKey,
                        nodePool.labelValue,
                        false
                    );
                    await nodePoolUtils.waitForClusterNodePoolPhase(
                        nodePool.nodePoolId,
                        [NodepoolPhase.Ready],
                        180
                    );
                    await nodeUtils.waitForNodeStatus(nodeName, Node.StatusEnum.Ready, 180);
                });
            });
        });
    }
);
