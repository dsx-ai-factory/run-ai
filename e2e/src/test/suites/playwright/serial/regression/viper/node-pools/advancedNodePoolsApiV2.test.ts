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
import { NodePool } from '../../../../../../../modules/api/apis/backendClusterApi';
import { test, expect } from '@playwright/test';
import {
    Node,
    NodepoolGPUNetworkAccelerationDetection,
    NodepoolUpdateFields
} from '../../../../../../../modules/api/apis/clusterApi';
import { GpuNetworkAccelerationDetection } from '../../../../../../../modules/helpers/utils/NodePoolUtils';
import { getNodePoolCrdField } from '../../../../../../../modules/helpers/utils/NodePoolCrdFields';
import { isEmpty, isNil } from 'lodash-es';
import { V1Node } from '@kubernetes/client-node';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

const nodePoolsApi = clusterApi.nodepoolsApi;
const MNNVLLabel = 'nvidia.com/gpu.clique';
const GpuNetworkAccelerationLabelKey = 'gpuNetworkAccelerationLabelKey';

test.describe(
    `Cluster Nodepools - GPU Network Accelerator detection, tags:  ${tags.resource.nodepools} ${tags.tool.api} ${tags.teams.viper}`,
    { tag: [tags.services.nodepoolController, tags.services.clusterService, tags.services.podGroupAssigner, tags.runOn.serialCI] },
    () => {
        supportedFrom(SupportedVersions.V2_23)(() => {
            let testNodePoolCreated: { nodePoolName: string; nodePoolId: string };
            let testNodePoolLabelKey: string;
            let testNodePoolLabelValue: string;
            let testNodeName: string;

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
                nodePoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;

                await operations.reduceSchedulerResources();
            });

            test.describe('Nodepools enabled in backend and cluster', () => {
                test.beforeAll(async () => {
                    await nodePoolUtils.deleteAllNodePools();
                });

                test.afterAll(async () => {
                    try {
                        await nodePoolUtils.deleteAllNodePools();
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
                        await nodePoolUtils.cleanupDeleteNodePool(testNodePoolCreated);
                        testNodePoolCreated = { nodePoolName: '', nodePoolId: '' };

                        if (testNodePoolCreated.nodePoolName) {
                            await kubeCli.unLabelNode(testNodeName, testNodePoolLabelKey);
                        }
                    });

                    const waitAndValidateNodepoolDetected = async (
                        nodePoolName: string,
                        MNNVLLabelOptions: { label: string; value: string },
                        nodePoolId: string,
                        expected: {
                            nvLinkCliqueId: string;
                            nvLinkDomainUid: string;
                            detected: boolean;
                        }
                    ): Promise<void> => {
                        if (MNNVLLabelOptions.label !== '') {
                            testNodeName = await nodePoolUtils.labelAGPUNode(
                                MNNVLLabelOptions.label,
                                MNNVLLabelOptions.value
                            );
                            expect(testNodeName).toBeTruthy();
                        }

                        const isDetectedAsExpected = (detected: boolean): boolean =>
                            !!detected === !!expected.detected;
                        const nodePoolCrdObj = await nodePoolUtils.waitForNodePoolCrdProperty(
                            nodePoolName,
                            'status.gpuNetworkAccelerationDetected',
                            isDetectedAsExpected
                        );
                        const nodesHasLabel = (nodes: Array<V1Node>): V1Node => {
                            return nodes.find(
                                (n) =>
                                    !isNil(n.metadata.annotations) &&
                                    n.metadata.annotations[testNodePoolLabelKey] !== '' &&
                                    isEmpty(
                                        n.metadata.annotations[GpuNetworkAccelerationLabelKey] ===
                                            MNNVLLabelOptions.label
                                    )
                            );
                        };
                        const nodeCrd = await nodePoolUtils.waitForNodeCrdProperty(
                            testNodePoolLabelKey,
                            nodesHasLabel
                        );

                        const nodepool = await nodePoolUtils.waitForClusterNodePoolProperty(
                            nodePoolId,
                            'status.gpuNetworkAccelerationDetected',
                            isDetectedAsExpected
                        );

                        expect(
                            getNodePoolCrdField<boolean>(
                                nodePoolCrdObj,
                                'status.gpuNetworkAccelerationDetected'
                            )
                        ).toBe(expected.detected);

                        const isNodesHasDomainAndClique = (nodes: Node[]): boolean => {
                            const node = nodes.find((n) => n.name === nodeCrd.metadata.name);
                            return (
                                !isNil(node) &&
                                node.nvLinkDomainUid === expected.nvLinkDomainUid &&
                                node.nvLinkCliqueId === expected.nvLinkCliqueId
                            );
                        };
                        await nodePoolUtils.waitForClusterNodesProperty(
                            nodepool.clusterId,
                            nodePoolName,
                            isNodesHasDomainAndClique
                        );
                    };

                    const waitAndValidateNodepoolReady = async (
                        nodePoolName: string,
                        MNNVLLabelOptions: { label: string; value: string },
                        nodePoolId: string,
                        expected: {
                            nvLinkCliqueId: string;
                            nvLinkDomainUid: string;
                            detected: boolean;
                        }
                    ): Promise<void> => {
                        // wait for crd to exist in cluster - with status "Ready"
                        await nodePoolUtils.waitForNodePoolCrdStatus(nodePoolName, [
                            NodePool.StatusEnum.Empty
                        ]);

                        // label a node to be assigned to the nodepool
                        const testNodeName = await nodePoolUtils.labelAGPUNode(
                            testNodePoolLabelKey,
                            testNodePoolLabelValue
                        );
                        expect(testNodeName).not.toBe('');

                        await waitAndValidateNodepoolDetected(
                            nodePoolName,
                            MNNVLLabelOptions,
                            nodePoolId,
                            expected
                        );
                    };

                    const testCreateNodepoolGPUNetworkAcceleration = async (
                        MNNVLLabelOptions: { label: string; value: string },
                        options: GpuNetworkAccelerationDetection,
                        expected: {
                            nvLinkCliqueId: string;
                            nvLinkDomainUid: string;
                            detected: boolean;
                        }
                    ): Promise<void> => {
                        const nodePoolNamePrefix = 'np';
                        // create nodepool
                        const { nodePoolName, nodePoolId } =
                            await nodePoolUtils.createClusterNodePoolAndValidate(
                                testNodePoolLabelKey,
                                testNodePoolLabelValue,
                                nodePoolNamePrefix,
                                options
                            );
                        testNodePoolCreated = { nodePoolName, nodePoolId };

                        await waitAndValidateNodepoolReady(
                            nodePoolName,
                            MNNVLLabelOptions,
                            nodePoolId,
                            expected
                        );
                    };

                    const testUpdateNodepoolGPUNetworkAcceleration = async (
                        MNNVLLabelOptions: { label: string; value: string },
                        options: GpuNetworkAccelerationDetection,
                        expected: {
                            nvLinkCliqueId: string;
                            nvLinkDomainUid: string;
                            detected: boolean;
                        }
                    ): Promise<void> => {
                        // update nodepool
                        const { nodePoolName, nodePoolId } = testNodePoolCreated;
                        const updateFields: NodepoolUpdateFields = {
                            ...options,
                            labelKey: testNodePoolLabelKey,
                            labelValue: testNodePoolLabelValue
                        };
                        await nodePoolUtils.updateClusterNodePoolAndValidate(nodePoolId, updateFields);
                        const isLabelKeySet = (labelValue: string): boolean =>
                            labelValue === options.gpuNetworkAccelerationLabelKey;
                        const nodePoolCrdObj = await nodePoolUtils.waitForNodePoolCrdProperty(
                            nodePoolName,
                            'spec.gpuNetworkAccelerationLabelKey',
                            isLabelKeySet
                        );

                        expect(nodePoolCrdObj).not.toBe(undefined);

                        await waitAndValidateNodepoolDetected(
                            nodePoolName,
                            MNNVLLabelOptions,
                            nodePoolId,
                            expected
                        );
                    };

                    test('Create nodepool - Not detect - validate GPU Network Acceleration Not Detected', async () => {
                        await testCreateNodepoolGPUNetworkAcceleration(
                            { label: MNNVLLabel, value: 'domain.clique' },
                            {
                                gpuNetworkAccelerationDetection:
                                    NodepoolGPUNetworkAccelerationDetection.DontUse,
                                gpuNetworkAccelerationLabelKey: MNNVLLabel
                            },
                            {
                                detected: false,
                                nvLinkDomainUid: undefined,
                                nvLinkCliqueId: undefined
                            }
                        );
                    });

                    test('Create nodepool - - Auto detect - validate GPU Network Acceleration Detected', async () => {
                        await testCreateNodepoolGPUNetworkAcceleration(
                            { label: MNNVLLabel, value: 'domain-1.clique-1' },
                            {
                                gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.Auto,
                                gpuNetworkAccelerationLabelKey: MNNVLLabel
                            },
                            {
                                detected: true,
                                nvLinkDomainUid: 'domain-1',
                                nvLinkCliqueId: 'clique-1'
                            }
                        );
                    });

                    test('Create nodepool - - always detect - validate GPU Network Acceleration Detected', async () => {
                        await testCreateNodepoolGPUNetworkAcceleration(
                            { label: '', value: '' },
                            { gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.Use },
                            {
                                detected: true,
                                nvLinkDomainUid: undefined,
                                nvLinkCliqueId: undefined
                            }
                        );
                    });

                    test('Create/ Update nodepool - - always detect - validate GPU Network Acceleration Detected', async () => {
                        await testCreateNodepoolGPUNetworkAcceleration(
                            { label: 'SOME-LABEL', value: 'only-the-domain' },
                            {
                                gpuNetworkAccelerationDetection:
                                    NodepoolGPUNetworkAccelerationDetection.DontUse,
                                gpuNetworkAccelerationLabelKey: 'OTHER-LABEL'
                            },
                            {
                                detected: false,
                                nvLinkDomainUid: undefined,
                                nvLinkCliqueId: undefined
                            }
                        );

                        await testUpdateNodepoolGPUNetworkAcceleration(
                            { label: 'OTHER-ACCELERATOR-LABEL', value: 'only-the-domain' },
                            {
                                gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.Auto,
                                gpuNetworkAccelerationLabelKey: 'OTHER-ACCELERATOR-LABEL'
                            },
                            {
                                detected: true,
                                nvLinkDomainUid: 'only-the-domain',
                                nvLinkCliqueId: undefined
                            }
                        );

                        await testUpdateNodepoolGPUNetworkAcceleration(
                            { label: 'OTHER-ACCELERATOR-LABEL', value: 'only-the-domain' },
                            {
                                gpuNetworkAccelerationDetection:
                                    NodepoolGPUNetworkAccelerationDetection.DontUse,
                                gpuNetworkAccelerationLabelKey: 'OTHER-ACCELERATOR-LABEL-2'
                            },
                            {
                                detected: false,
                                nvLinkDomainUid: undefined,
                                nvLinkCliqueId: undefined
                            }
                        );

                        await testUpdateNodepoolGPUNetworkAcceleration(
                            { label: 'OTHER-ACCELERATOR-LABEL-2', value: 'only-the-domain' },
                            {
                                gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.Auto,
                                gpuNetworkAccelerationLabelKey: 'OTHER-ACCELERATOR-LABEL-2'
                            },
                            {
                                detected: true,
                                nvLinkDomainUid: 'only-the-domain',
                                nvLinkCliqueId: undefined
                            }
                        );

                        await testUpdateNodepoolGPUNetworkAcceleration(
                            { label: 'OTHER-ACCELERATOR-LABEL-2', value: 'only-the-domain-2' },
                            {
                                gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.Auto,
                                gpuNetworkAccelerationLabelKey: 'OTHER-ACCELERATOR-LABEL-3'
                            },
                            {
                                detected: false,
                                nvLinkDomainUid: undefined,
                                nvLinkCliqueId: undefined
                            }
                        );
                    });
                });
            });
        });
    }
);
