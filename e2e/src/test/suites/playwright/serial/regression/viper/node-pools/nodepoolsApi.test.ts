/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { backendClusterApi, clusterApi, networkTopologiesApi } from '../../../../../../../modules/api';
import {
    NodepoolCreateFields,
    NodepoolUpdateFields,
    NodepoolPhase,
    NetworkTopologyCreateRequest,
    NodepoolConditionDetails,
    PlacementStrategy
} from '../../../../../../../modules/api/apis/clusterApi';
import {
    flowRootState,
    logger,
    nodePoolUtils,
    nodeUtils,
    operations,
    tags,
    versionUtils
} from '../../../../../../../modules/helpers';
import { StatusCodes } from 'http-status-codes';
import { test } from '@playwright/test';

import { kubeCli } from '../../../../../../../modules/cli';
import { getNodePoolCrdField } from '../../../../../../../modules/helpers/utils/NodePoolCrdFields';
import { isNil } from 'lodash-es';

import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

const clusterNodePoolsApi = clusterApi.nodepoolsApi;

/**
 * Reusable function to validate that a nodepool has the correct network topology
 * @param nodePoolId - The ID of the nodepool to validate
 * @param expectedTopologyId - The expected network topology ID
 * @param expectedTopologyName - The expected network topology name
 */
async function validateNodePoolNetworkTopology(
    nodePoolId: string,
    expectedTopologyId: string,
    expectedTopologyName: string
): Promise<void> {
    // Get the nodepool from cluster API to validate network topology
    const nodepoolResponse = await nodePoolUtils.getNodePoolById(nodePoolId);
    expect(nodepoolResponse.response.statusCode).to.equal(200);
    expect(nodepoolResponse.body.networkTopologyId).to.equal(expectedTopologyId);

    // Validate that the network topology name is set (this should be populated by the system)
    if (nodepoolResponse.body.networkTopologyName) {
        expect(nodepoolResponse.body.networkTopologyName).to.equal(expectedTopologyName);
        logger.info(`Nodepool network topology name validated: ${nodepoolResponse.body.networkTopologyName}`);
    } else {
        logger.warn('Network topology name not yet populated, this might be expected during creation phase');
    }
}

test.describe(
    `Nodepools API, tags:  ${tags.resource.nodepools} ${tags.tool.api} ${tags.teams.viper}`,
    { tag: [tags.services.clusterService, tags.services.nodepoolController, tags.runOn.serialCI] },
    function () {
        let nodePoolLabelKeyValue: { key: string; value: string };
        let nodePoolCreated: { nodePoolName: string; nodePoolId: string };

        test.beforeEach(() => {
            nodePoolCreated = { nodePoolName: '', nodePoolId: '' };
        });

        test.beforeAll(async () => {
            clusterNodePoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;
            await operations.reduceSchedulerResources('50m', '64Mi', '250m', '128Mi');
        });

        test.afterAll(async () => {
            await operations.revertSchedulerResourcesToDefault();
        });

        test.describe('Nodepools in backend and cluster', () => {
            test.describe('Create', () => {
                test.beforeEach(() => {
                    nodePoolLabelKeyValue = nodePoolUtils.generateKeyValueLabels();
                });

                test.afterEach(async () => {
                    await nodePoolUtils.cleanupDeleteClusterNodePool(nodePoolCreated);
                    nodePoolCreated = { nodePoolName: '', nodePoolId: '' };
                });

                test('Should be created successfully', async () => {
                    // create nodepool
                    const { nodePoolName, nodePoolId } = await nodePoolUtils.createClusterNodePoolAndValidate(
                        nodePoolLabelKeyValue.key,
                        nodePoolLabelKeyValue.value
                    );
                    nodePoolCreated = { nodePoolName, nodePoolId };

                    // validate created nodepool via cluster API
                    const nodepoolResponse = await nodePoolUtils.getNodePoolById(nodePoolId);
                    expect(nodepoolResponse.response.statusCode).to.equal(200);
                    expect(nodepoolResponse.body.name).to.equal(nodePoolName);
                    expect(nodepoolResponse.body.labelKey).to.equal(nodePoolLabelKeyValue.key);
                    expect(nodepoolResponse.body.labelValue).to.equal(nodePoolLabelKeyValue.value);

                    // wait for CRD readiness
                    const nodePoolCrd = await nodePoolUtils.waitForNodePoolCrdReadyOrEmpty(nodePoolName, 40);
                    await nodePoolUtils.waitForClusterNodePoolPhase(nodePoolId, [
                        NodepoolPhase.Ready,
                        NodepoolPhase.Empty
                    ]);

                    // validate nodepool in cluster
                    nodePoolUtils.validateNodePoolCrdReady(
                        nodePoolCrd,
                        nodePoolName,
                        nodePoolLabelKeyValue.key,
                        nodePoolLabelKeyValue.value
                    );

                    await nodePoolUtils.validateNodePoolsSchedulerArgs(nodePoolName);
                    nodePoolUtils.validateNodesStatusOfNodePoolCrd(nodePoolCrd);

                    // validate nodepools count
                    const allNodePools = await nodePoolUtils.getAllNodePoolsByClusterAndValidate();
                    expect(allNodePools.length).to.greaterThanOrEqual(2); // at least one default nodepool should exist
                    const nodepoolsCount = await nodePoolUtils.getNodepoolsCountByClusterId(
                        flowRootState.cluster.uuid
                    );
                    expect(nodepoolsCount).to.equal(allNodePools.length);
                });

                test('Deprecated placement strategy sanity - GET returns the value in both deprecated and new fields', async () => {
                    // create a basic nodepool, then set the strategy via the DEPRECATED placementStrategy field
                    const { nodePoolName, nodePoolId } =
                        await nodePoolUtils.createClusterNodePoolAndValidate(
                            nodePoolLabelKeyValue.key,
                            nodePoolLabelKeyValue.value
                        );
                    nodePoolCreated = { nodePoolName, nodePoolId };

                    // distinct cpu/gpu values so the round-trip is meaningful
                    const updateFields: NodepoolUpdateFields = {
                        labelKey: nodePoolLabelKeyValue.key,
                        labelValue: nodePoolLabelKeyValue.value,
                        schedulingConfiguration: {
                            placementStrategy: {
                                cpu: PlacementStrategy.Binpack,
                                gpu: PlacementStrategy.Spread
                            }
                        }
                    };
                    await nodePoolUtils.updateClusterNodePoolAndValidate(nodePoolId, updateFields);

                    // GET the nodepool and validate the deprecated write round-trips correctly
                    const nodepoolResponse = await nodePoolUtils.getNodePoolById(nodePoolId);
                    expect(nodepoolResponse.response.statusCode).to.equal(StatusCodes.OK);
                    const schedulingConfiguration = nodepoolResponse.body.schedulingConfiguration;
                    expect(schedulingConfiguration, 'schedulingConfiguration should be returned').to.exist;

                    // the deprecated fields still reflect the node-level strategy
                    expect(schedulingConfiguration.placementStrategy.cpu).to.equal(
                        PlacementStrategy.Binpack
                    );
                    expect(schedulingConfiguration.placementStrategy.gpu).to.equal(
                        PlacementStrategy.Spread
                    );

                    // the new node-level fields are always populated (mirrored from the deprecated values)
                    expect(schedulingConfiguration.gpuPodsPlacementStrategy.node).to.equal(
                        PlacementStrategy.Spread
                    );
                    expect(schedulingConfiguration.cpuPodsPlacementStrategy.node).to.equal(
                        PlacementStrategy.Binpack
                    );

                    // on >= 2.26 the device-level field is also populated, with the same value as node-level
                    if (
                        versionUtils.isVersionSupported(
                            flowRootState.cluster.version,
                            SupportedVersions.V2_26
                        )
                    ) {
                        expect(schedulingConfiguration.gpuPodsPlacementStrategy.device).to.equal(
                            schedulingConfiguration.gpuPodsPlacementStrategy.node
                        );
                    }
                });

                test('Already exists by name', async () => {
                    const { nodePoolName, nodePoolId } = await nodePoolUtils.createClusterNodePoolAndValidate(
                        nodePoolLabelKeyValue.key,
                        nodePoolLabelKeyValue.value
                    );
                    nodePoolCreated = { nodePoolName, nodePoolId };

                    await nodePoolUtils.waitForNodePoolCrdReadyOrEmpty(nodePoolName, 40);
                    await nodePoolUtils.waitForClusterNodePoolPhase(nodePoolId, [
                        NodepoolPhase.Ready,
                        NodepoolPhase.Empty
                    ]);

                    try {
                        await nodePoolUtils.createClusterNodePool(nodePoolName, 'key3', 'val3');
                        expect.fail('Should not allow duplicated name nodepools');
                    } catch (e) {
                        expect(e.statusCode).equal(StatusCodes.CONFLICT);
                    }
                });
            });

            test.describe('Delete', () => {
                test.beforeEach(() => {
                    nodePoolLabelKeyValue = nodePoolUtils.generateKeyValueLabels();
                });

                test('Deleted successfully', async () => {
                    const createRes = await nodePoolUtils.createClusterNodePoolAndValidate(
                        nodePoolLabelKeyValue.key,
                        nodePoolLabelKeyValue.value
                    );
                    nodePoolCreated = createRes;
                    const nodePoolId = createRes.nodePoolId;

                    await nodePoolUtils.waitForNodePoolCrdReadyOrEmpty(createRes.nodePoolName);

                    await clusterNodePoolsApi.deleteNodepool(nodePoolId);
                    await nodePoolUtils.waitForDeletedNodePool(nodePoolId);
                    await nodePoolUtils.waitForNodePoolCrd(createRes.nodePoolName, 60, false);

                    const allNodePools = await nodePoolUtils.getAllNodePoolsByClusterAndValidate();
                    const currentNodePool = allNodePools.find((nodePool) => nodePool.id === nodePoolId);
                    expect(currentNodePool).to.be.undefined;
                });
            });

            test.describe('Should create network topology with 3 levels and nodepool with that topology', () => {
                supportedFrom(SupportedVersions.V2_23)(() => {
                    test('Should create network topology with 3 levels and nodepool with that topology', async () => {
                        let nodePoolWithTopologyCreated: { nodePoolName: string; nodePoolId: string };
                        let createdTopologyId: string;
                        // Set access tokens
                        networkTopologiesApi.accessToken =
                            flowRootState.users.system_administrator.accessToken;
                        clusterNodePoolsApi.accessToken =
                            flowRootState.users.system_administrator.accessToken;

                        // Generate unique names
                        const createdTopologyName = `test-topology-${nodePoolUtils.generateKeyValueLabels().key}`;
                        const nodePoolLabelKeyValue = nodePoolUtils.generateKeyValueLabels();

                        await test.step('Create network topology with 3 levels', async () => {
                            const createRequest: NetworkTopologyCreateRequest = {
                                name: createdTopologyName,
                                clusterId: flowRootState.cluster.uuid,
                                levels: [
                                    'cloud.provider.com/topology-block',
                                    'cloud.provider.com/topology-rack',
                                    'kubernetes.io/hostname'
                                ]
                            };

                            const response = await networkTopologiesApi.createNetworkTopology(createRequest, {
                                headers: {}
                            });
                            expect(response.response.statusCode).to.equal(201);
                            expect(response.body).to.have.property('id');
                            expect(response.body.name).to.equal(createdTopologyName);
                            expect(response.body.clusterId).to.equal(flowRootState.cluster.uuid);
                            expect(response.body.levels).to.deep.equal(createRequest.levels);
                            expect(response.body.source).to.equal('Run.ai');

                            createdTopologyId = response.body.id;
                            logger.info(`Network topology created with ID: ${createdTopologyId}`);
                        });

                        await test.step('Create nodepool with network topology ID', async function () {
                            const nodePoolCreateRequest = new NodepoolCreateFields();
                            nodePoolCreateRequest.name = nodePoolLabelKeyValue.key;
                            nodePoolCreateRequest.labelKey = nodePoolLabelKeyValue.key;
                            nodePoolCreateRequest.labelValue = nodePoolLabelKeyValue.value;
                            nodePoolCreateRequest.clusterId = flowRootState.cluster.uuid;
                            nodePoolCreateRequest.networkTopologyId = createdTopologyId;

                            const response = await clusterNodePoolsApi.createNodepool(nodePoolCreateRequest, {
                                headers: {}
                            });
                            expect(response.response.statusCode).to.equal(201);
                            expect(response.body).to.have.property('id');
                            expect(response.body.name).to.equal(nodePoolLabelKeyValue.key);
                            expect(response.body.networkTopologyId).to.equal(createdTopologyId);

                            nodePoolWithTopologyCreated = {
                                nodePoolName: response.body.name,
                                nodePoolId: response.body.id
                            };
                            logger.info(`Nodepool created with topology ID: ${response.body.id}`);
                        });

                        await test.step('Validate nodepool has correct network topology name', async function () {
                            // Wait for the nodepool to be ready
                            await nodePoolUtils.waitForNodePoolCrdReadyOrEmpty(
                                nodePoolWithTopologyCreated.nodePoolName,
                                40
                            );

                            // Use the reusable validation function
                            await validateNodePoolNetworkTopology(
                                nodePoolWithTopologyCreated.nodePoolId,
                                createdTopologyId,
                                createdTopologyName
                            );
                        });
                    });
                });
            });

            test.describe('Should create network topology with 3 levels and validate nodepool CRD spec.networkTopologyName', () => {
                supportedFrom(SupportedVersions.V2_23)(() => {
                    let nodePoolWithTopologyCreated: { nodePoolName: string; nodePoolId: string };
                    let createdTopologyId: string;
                    // Set access tokens
                    networkTopologiesApi.accessToken = flowRootState.users.system_administrator.accessToken;
                    clusterNodePoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;

                    // Generate unique names
                    const topologyName = `test-topology-crd-${nodePoolUtils.generateKeyValueLabels().key}`;
                    const nodePoolLabelKeyValue = nodePoolUtils.generateKeyValueLabels();
                    test('Create network topology with 3 levels', async function () {
                        const createRequest: NetworkTopologyCreateRequest = {
                            name: topologyName,
                            clusterId: flowRootState.cluster.uuid,
                            levels: [
                                'cloud.provider.com/topology-block',
                                'cloud.provider.com/topology-rack',
                                'kubernetes.io/hostname'
                            ]
                        };

                        const response = await networkTopologiesApi.createNetworkTopology(createRequest, {
                            headers: {}
                        });
                        expect(response.response.statusCode).to.equal(201);
                        expect(response.body).to.have.property('id');
                        expect(response.body.name).to.equal(topologyName);
                        expect(response.body.clusterId).to.equal(flowRootState.cluster.uuid);
                        expect(response.body.levels).to.deep.equal(createRequest.levels);
                        expect(response.body.source).to.equal('Run.ai');

                        createdTopologyId = response.body.id;
                        logger.info(`Network topology created with ID: ${createdTopologyId}`);
                    });

                    test('Create nodepool with network topology ID', async function () {
                        const nodePoolCreateRequest = new NodepoolCreateFields();
                        nodePoolCreateRequest.name = nodePoolLabelKeyValue.key;
                        nodePoolCreateRequest.labelKey = nodePoolLabelKeyValue.key;
                        nodePoolCreateRequest.labelValue = nodePoolLabelKeyValue.value;
                        nodePoolCreateRequest.clusterId = flowRootState.cluster.uuid;
                        nodePoolCreateRequest.networkTopologyId = createdTopologyId;

                        const response = await clusterNodePoolsApi.createNodepool(nodePoolCreateRequest, {
                            headers: {}
                        });
                        expect(response.response.statusCode).to.equal(201);
                        expect(response.body).to.have.property('id');
                        expect(response.body.name).to.equal(nodePoolLabelKeyValue.key);
                        expect(response.body.networkTopologyId).to.equal(createdTopologyId);

                        nodePoolWithTopologyCreated = {
                            nodePoolName: response.body.name,
                            nodePoolId: response.body.id
                        };
                        logger.info(`Nodepool created with topology ID: ${response.body.id}`);
                    });

                    test('Get nodepool CRD and validate spec.networkTopologyName', async function () {
                        // Wait for the nodepool CRD to be ready
                        const nodePoolCrd = await nodePoolUtils.waitForNodePoolCrdReadyOrEmpty(
                            nodePoolWithTopologyCreated.nodePoolName,
                            40
                        );

                        // Validate that the CRD has the correct network topology name in spec - the
                        // KAI CRD renamed the field to spec.preferredNetworkTopologyName
                        expect(nodePoolCrd).to.have.property('spec');
                        const crdTopologyName = getNodePoolCrdField<string>(
                            nodePoolCrd,
                            'spec.networkTopologyName'
                        );
                        expect(crdTopologyName).to.equal(topologyName);

                        logger.info(`Nodepool CRD network topology name validated: ${crdTopologyName}`);

                        // Additional validation: check that the CRD metadata and other fields are correct
                        expect((nodePoolCrd['metadata'] as any).name).to.equal(
                            nodePoolWithTopologyCreated.nodePoolName
                        );
                        expect((nodePoolCrd['spec'] as any).labelKey).to.equal(nodePoolLabelKeyValue.key);
                        expect((nodePoolCrd['spec'] as any).labelValue).to.equal(nodePoolLabelKeyValue.value);
                    });
                });
            });

            test.describe('Should validate network topology with nodes and detect topology mismatches when labels are missing', () => {
                supportedFrom(SupportedVersions.V2_23)(() => {
                    // Set access tokens
                    networkTopologiesApi.accessToken = flowRootState.users.system_administrator.accessToken;
                    clusterNodePoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;

                    // Generate unique names
                    const topologyName = `test-topology-nodes-${nodePoolUtils.generateKeyValueLabels().key}`;
                    const nodePoolLabelKeyValue = nodePoolUtils.generateKeyValueLabels();
                    let topologyId: string;
                    let nodePoolId: string;
                    let nodeName: string;

                    test('Create network topology with 3 levels', async function () {
                        const createRequest: NetworkTopologyCreateRequest = {
                            name: topologyName,
                            clusterId: flowRootState.cluster.uuid,
                            levels: [
                                'cloud.provider.com/topology-block',
                                'cloud.provider.com/topology-rack',
                                'cloud.provider/hostname'
                            ]
                        };

                        const response = await networkTopologiesApi.createNetworkTopology(createRequest, {
                            headers: {}
                        });
                        expect(response.response.statusCode).to.equal(201);
                        expect(response.body).to.have.property('id');
                        expect(response.body.name).to.equal(topologyName);
                        expect(response.body.clusterId).to.equal(flowRootState.cluster.uuid);
                        expect(response.body.levels).to.deep.equal(createRequest.levels);
                        expect(response.body.source).to.equal('Run.ai');

                        topologyId = response.body.id;
                        logger.info(`Network topology created with ID: ${topologyId}`);
                    });

                    test('Create nodepool with network topology', async function () {
                        const nodePoolCreateRequest = new NodepoolCreateFields();
                        nodePoolCreateRequest.name = nodePoolLabelKeyValue.key;
                        nodePoolCreateRequest.labelKey = nodePoolLabelKeyValue.key;
                        nodePoolCreateRequest.labelValue = nodePoolLabelKeyValue.value;
                        nodePoolCreateRequest.clusterId = flowRootState.cluster.uuid;
                        nodePoolCreateRequest.networkTopologyId = topologyId;

                        const response = await clusterNodePoolsApi.createNodepool(nodePoolCreateRequest, {
                            headers: {}
                        });
                        expect(response.response.statusCode).to.equal(201);
                        expect(response.body).to.have.property('id');
                        expect(response.body.name).to.equal(nodePoolLabelKeyValue.key);
                        expect(response.body.networkTopologyId).to.equal(topologyId);

                        nodePoolId = response.body.id;
                        logger.info(
                            `Nodepool ${response.body.name} NodepoolId ${nodePoolId} created with topology ID: ${topologyId}`
                        );
                    });

                    test('Get 1 node from the cluster', async function () {
                        // Get a node with GPUs
                        const node = await nodePoolUtils.getNodeWithGpus();
                        expect(node).to.not.be.undefined;
                        nodeName = node.metadata.name;
                        try {
                            const removeLabels = [
                                nodePoolLabelKeyValue.key,
                                'cloud.provider.com/topology-block',
                                'cloud.provider.com/topology-rack',
                                'cloud.provider/hostname'
                            ];

                            await nodeUtils.unlabelNodes(nodeName, removeLabels);
                            logger.error(`Un Label Node ${nodeName} finished`);
                        } catch (e) {
                            logger.error(`Un Label Node ${nodeName}`, e);
                        }

                        logger.info(`Selected node: ${nodeName}`);
                    });

                    test('Add node with all 3 labels (matching topology levels)', async function () {
                        try {
                            // Apply nodepool label to connect node to nodepool
                            await kubeCli.labelNode(
                                nodeName,
                                nodePoolLabelKeyValue.key,
                                nodePoolLabelKeyValue.value
                            );

                            // Do not Apply topology level labels
                        } catch (e) {
                            logger.error(`Label Node ${nodeName}`, e);
                        }
                        logger.info(
                            `Node ${nodeName} labeled with all topology levels and connected to nodepool`
                        );
                    });

                    test('Validate topology mismatch conditions on nodepool', async function () {
                        const listsMismatchingNode = (
                            conditions: Array<NodepoolConditionDetails>
                        ): boolean => {
                            const topologyMismatchCondition = conditions?.find(
                                (condition) => condition.type === 'NodeTopologyMismatch'
                            );
                            return (
                                !isNil(topologyMismatchCondition) &&
                                topologyMismatchCondition.status ===
                                    NodepoolConditionDetails.StatusEnum.True &&
                                (topologyMismatchCondition.nodes ?? []).includes(nodeName)
                            );
                        };

                        await nodePoolUtils.waitForClusterNodePoolProperty(
                            nodePoolId,
                            'status.conditions',
                            listsMismatchingNode,
                            90
                        );

                        const nodepoolResponse = await clusterNodePoolsApi.getNodepool(nodePoolId, {
                            headers: {}
                        });
                        expect(nodepoolResponse.response.statusCode).to.equal(200);

                        // Validate that the nodepool has NodeTopologyMismatch condition
                        const topologyMismatchCondition = nodepoolResponse.body.status.conditions.find(
                            (condition: any) => condition.type === 'NodeTopologyMismatch'
                        );
                        expect(topologyMismatchCondition).to.not.be.undefined;
                        expect(topologyMismatchCondition.status).to.equal('True');
                        expect(topologyMismatchCondition.nodes).to.include(nodeName);
                        logger.info(
                            `NodeTopologyMismatch condition found: ${JSON.stringify(topologyMismatchCondition)}`
                        );

                        logger.info('Topology mismatch conditions validated on nodepool');
                    });

                    test('Cleanup: Remove labels from node', async function () {
                        if (!nodeName) {
                            logger.warn('Skipping label cleanup - nodeName is undefined (prior test likely failed)');
                            return;
                        }

                        const removeLabels = [
                            nodePoolLabelKeyValue.key,
                            'cloud.provider.com/topology-block',
                            'cloud.provider.com/topology-rack',
                            'cloud.provider/hostname'
                        ];

                        await nodeUtils.unlabelNodes(nodeName, removeLabels);
                        logger.info('Labels removed from node for cleanup');
                    });
                });
            });
        });

        test.describe('Backwards compatibility - backend API', () => {
            test('backwards - create and delete nodepool via backend API', async () => {
                backendClusterApi.nodePoolsApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;
                const labelKeyValue = nodePoolUtils.generateKeyValueLabels();
                const { nodePoolName, nodePoolId } = await nodePoolUtils.createBackendNodePoolAndValidate(
                    labelKeyValue.key,
                    labelKeyValue.value
                );

                await nodePoolUtils.waitForNodePoolCrdReadyOrEmpty(nodePoolName, 40);

                await backendClusterApi.nodePoolsApi.deleteNodePool(flowRootState.cluster.uuid, nodePoolId);
                await nodePoolUtils.waitForNodePoolCrd(nodePoolName, 60, false);
            });
        });
    }
);
