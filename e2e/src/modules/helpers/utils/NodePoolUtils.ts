/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import http from 'http';
import {
    crds,
    CustomResourceDefinition,
    flowRootState,
    logger,
    versionUtils,
    networkTopologiesUtils,
    nodeUtils,
    objectGenerator,
    orgUnitUtils,
    projectResourcesUtils,
    Utils,
    waitUtils,
    workloadsUtilsV2
} from '../index';

import {
    DisplayedJob,
    NodePool,
    NodePoolCreateRequest,
    NodePoolLabelsRequest
} from '../../api/apis/backendClusterApi';
import {
    NodeAffinity,
    Project as ProjectFromProjectsDepartments,
    ProjectCreateRequest
} from '../../api/apis/projectsDepartmentsApi';
import { Project as ProjectFromOrgUnit } from '../../api/apis/orgUnitApi';
import { backendClusterApi, clusterApi, orgUnitService } from '../../api';
import { expect } from 'chai';
import { kubeCli, runaiCliV2 } from '../../cli';
import { StatusCodes } from 'http-status-codes';
import { KubernetesObject } from '@kubernetes/client-node/dist/types';

import { NodeType } from '../../api/apis/orgUnitApi';
import {
    GetNodepoolsMinimal200Response,
    GpuPodsPlacementStrategy,
    NetworkTopologyCreateRequest,
    Node,
    Nodepool,
    NodepoolCreateFields,
    NodepoolGPUNetworkAccelerationDetection,
    NodepoolMinimal,
    NodepoolPhase,
    NodepoolUpdateFields,
    PlacementStrategy,
    SchedulingConfiguration,
    SyncNodepoolsStatusRequest
} from '../../api/apis/clusterApi';
import { get, isEmpty, isNil } from 'lodash-es';
import { SupportedVersions } from './SupporetedEnvsUtils';
import { Workload } from '../../api/apis/workloadsApi/generated/model/workload';
import { Phase } from '../../api/apis/workloadsApi/generated/model/phase';
import { Department } from '../../api/apis/orgUnitApi';
import { V1Affinity, V1Deployment, V1Node, V1NodeSelectorRequirement } from '@kubernetes/client-node';
import { NodepoolConfig } from '../models/externalWorkload';
import { DEFAULT_NODE_POOL_NAME } from '../../ui/utils/nodePoolUtils';
import { cliV2WorkloadType, SubmitFlagsMerged } from '../../cli/tools/RunaiCliV2.constants';
import { getNodePoolCrdField, resolveNodePoolCrd } from './NodePoolCrdFields';

const runaiSchedulerName = 'runai-scheduler';
export const MIN_PLACEMENT_VERSION = '2.13.0';

export interface CreateNodeGroupOptions {
    nodeLabelKey: string;
    nodeLabelVal1: string;
    nodeLabelVal2?: string;
    nodeGroupSize?: number;
    createTopology?: boolean;
    attachTopology?: boolean;
    topoLevels?: string[];
    createTopoRequest?: NetworkTopologyCreateRequest;
    topologyIds?: string[];
    topoRegionKey?: string;
    topoCityKey?: string;
    region?: string;
    cityName?: string;
    nodePoolCreateRequest: NodepoolCreateFields;
    nodePoolCreateRequest2?: NodepoolCreateFields;
    nodePools: Array<{ name: string; id: string }>;
    npName1: string;
    npName2?: string;
    npProjectName: string;
    gpuDeserved?: number;
    topologyReadyTimeout?: number;
    workerNodeNames?: string[];
    nodeGroup1?: string[];
    nodeGroup2?: string[];
    midIndex?: number;
}

export type NodePoolParamsForCreateProjectWithResources = {
    nodePoolName: string;
    nodePoolId: number | string;
    deservedGPUs: number;
};

export type GpuNetworkAccelerationDetection = {
    gpuNetworkAccelerationLabelKey?: string;
    gpuNetworkAccelerationDetection?: NodepoolGPUNetworkAccelerationDetection;
};

export class NodePoolUtils {
    RunaiNodeTypeLabelKey = 'run.ai/type';

    DefaultNodePoolName = 'default';

    NodePoolNameLabel = 'runai/node-pool';

    InfiniteSchedulingBackoff = -1;

    SingleSchedulingBackoff = 1;

    UnschedulableOnNodePool = 'UnschedulableOnNodePool';

    async getNodePoolById(
        nodePoolId: string,
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<any> {
        clusterApi.nodepoolsApi.accessToken = accessToken;
        return await clusterApi.nodepoolsApi.getNodepool(nodePoolId);
    }

    async getNodeWithGpus(excludeNodeNames = []): Promise<V1Node> {
        const nodesList = await kubeCli.getNodes('nvidia.com/gpu.count');
        expect(nodesList.items.length).greaterThan(0);
        for (const node of nodesList.items) {
            if (!excludeNodeNames.includes(node.metadata.name)) {
                return node;
            }
        }
        logger.warn(`Failed to find a node with GPUs that's not in excluded list: ${excludeNodeNames}`);
        return undefined;
    }

    async validateNodeLabeledWithNodePool(nodePoolName: string, nodeName: string): Promise<void> {
        const nodeLabels = await kubeCli.getNodeLabels(nodeName);
        if (nodePoolName === this.DefaultNodePoolName) {
            expect(nodeLabels).not.to.haveOwnProperty(this.NodePoolNameLabel);
        } else {
            expect(nodeLabels[this.NodePoolNameLabel]).to.equal(nodePoolName);
        }
    }

    async labelAGPUNode(
        nodePoolLabelKey: string,
        nodePoolLabelValue: string,
        nodesAlreadyLabeled = []
    ): Promise<string> {
        const nodeWithGpuNotLabeled = await this.getNodeWithGpus(nodesAlreadyLabeled);
        if (!nodeWithGpuNotLabeled) {
            return '';
        }
        await kubeCli.labelNode(nodeWithGpuNotLabeled.metadata.name, nodePoolLabelKey, nodePoolLabelValue);
        return nodeWithGpuNotLabeled.metadata.name;
    }

    async annotateAGPUNode(
        nodePoolLabelKey: string,
        nodePoolLabelValue: string,
        nodesAlreadyLabeled = []
    ): Promise<string> {
        const nodeWithGpuNotLabeled = await this.getNodeWithGpus(nodesAlreadyLabeled);
        if (!nodeWithGpuNotLabeled) {
            return '';
        }
        await kubeCli.annotateNode(nodeWithGpuNotLabeled.metadata.name, nodePoolLabelKey, nodePoolLabelValue);
        return nodeWithGpuNotLabeled.metadata.name;
    }

    async getNodePoolCrd(nodePoolName: string): Promise<KubernetesObject> {
        return await kubeCli.getCrd(await resolveNodePoolCrd(), nodePoolName);
    }

    async deleteNodePoolByIdWithWait(
        nodepoolId: any,
        nodepoolName: string,
        shouldWaitForCrd = true,
        timeoutSeconds = 60,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        backendClusterApi.nodePoolsApi.accessToken = accessToken;
        await backendClusterApi.nodePoolsApi.deleteNodePool(flowRootState.cluster.uuid, nodepoolId as any);
        if (shouldWaitForCrd) await this.waitForNodePoolCrd(nodepoolName, timeoutSeconds, false);
    }

    async deleteNodePool(nodePool: NodePool, shouldWaitForCrd = true, timeoutSeconds = 60): Promise<void> {
        return this.deleteNodePoolByIdWithWait(nodePool.id, nodePool.name, shouldWaitForCrd, timeoutSeconds);
    }

    async getNodePoolsScheduler(nodePoolName: string): Promise<V1Deployment> {
        const schedulerName = `${runaiSchedulerName}-${nodePoolName}`;

        try {
            return await kubeCli.getDeployment(schedulerName, 'runai');
        } catch (error) {
            logger.warn(`Failed getting deployment ${schedulerName}, err: ${error}`);
            return undefined;
        }
    }

    generateKeyValueLabels(): {
        key: string;
        value: string;
    } {
        const nodePoolLabelKey = objectGenerator.generateNameByPrefix('nodepoolkey');
        const nodePoolLabelValue = objectGenerator.generateNameByPrefix('nodepoolval');
        return { key: nodePoolLabelKey, value: nodePoolLabelValue };
    }

    async getNodePoolsSchedulerArgs(
        nodePoolName: string
    ): Promise<{ nodePoolLabelKey: string; nodePoolLabelValue: string }> {
        const scheduler = await this.getNodePoolsScheduler(nodePoolName);

        const containers = scheduler.spec.template.spec.containers;
        const schedulerContainer = containers.find((containers) =>
            containers.name.includes(runaiSchedulerName)
        );
        expect(schedulerContainer).not.to.be.undefined;
        const args = schedulerContainer.args;

        // From 2.24+, flags are passed as --key=value. Before 2.24, they were --key value (two args).
        const isNewFormat = versionUtils.isVersionSupported(flowRootState.cluster.version, '2.24');

        let nodePoolLabelKey: string | undefined;
        let nodePoolLabelValue: string | undefined;

        if (isNewFormat) {
            // New format: --nodepool-label-key=..., --partition-label-value=...
            const labelKeyArg = args.find((a: string) => a.startsWith('--nodepool-label-key='));
            const labelValArg = args.find((a: string) => a.startsWith('--partition-label-value='));
            if (labelKeyArg && labelValArg) {
                nodePoolLabelKey = labelKeyArg.split('=')[1] ?? '';
                nodePoolLabelValue = labelValArg.split('=')[1] ?? '';
            }
        }

        // Fallback to legacy parsing or for clusters < 2.24
        if (nodePoolLabelKey === undefined || nodePoolLabelValue === undefined) {
            const labelKeyArgIndex = args.indexOf('--nodepool-label-key');
            const labelValueArgIndex = args.indexOf('--partition-label-value');
            expect(labelKeyArgIndex).greaterThanOrEqual(0);
            expect(labelValueArgIndex).greaterThanOrEqual(0);

            const nodePoolKeyArgIndex = labelKeyArgIndex + 1;
            const nodePoolValueArgIndex = labelValueArgIndex + 1;
            expect(args.length).greaterThan(nodePoolKeyArgIndex);
            expect(args.length).greaterThan(nodePoolValueArgIndex);

            nodePoolLabelKey = args[nodePoolKeyArgIndex];
            nodePoolLabelValue = args[nodePoolValueArgIndex];
        }

        return {
            nodePoolLabelKey: nodePoolLabelKey,
            nodePoolLabelValue: nodePoolLabelValue
        };
    }

    async validateNodePoolsSchedulerArgs(nodePoolName: string): Promise<void> {
        const { nodePoolLabelKey, nodePoolLabelValue } = await this.getNodePoolsSchedulerArgs(nodePoolName);
        expect(nodePoolLabelKey).to.be.equal(this.NodePoolNameLabel);

        if (nodePoolName === this.DefaultNodePoolName) {
            expect(nodePoolLabelValue).to.be.equal('');
        } else {
            expect(nodePoolLabelValue).to.be.equal(nodePoolName);
        }
    }

    async waitForNodePoolCrd(nodePoolName: string, timeout = 60, desiredResult = true): Promise<object> {
        return await waitUtils.waitForCrd(
            await resolveNodePoolCrd(),
            nodePoolName,
            '',
            timeout,
            desiredResult
        );
    }

    async waitForNodePoolCrdStatus(
        nodePoolName: string,
        exceptedStatuses = [NodePool.StatusEnum.Ready],
        timeout = 50,
        delay = 1
    ): Promise<object> {
        let res: object;
        let status;

        const doesNodePoolStatusMatch = async (): Promise<boolean> => {
            try {
                res = await this.getNodePoolCrd(nodePoolName);
                status = res['status'].phase;
                return exceptedStatuses.includes(status);
            } catch (e) {
                logger.warn(`Failed to get NodePool CRD status for ${nodePoolName}: ${e.message}`);
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                doesNodePoolStatusMatch,
                `waitForNodePoolCrdStatus nodePoolName: ${nodePoolName}; exceptedStatuses: ${exceptedStatuses}`,
                true,
                timeout,
                delay
            );
        } catch (e) {
            logger.debug(
                `Failed waiting for node pool with statuses ${exceptedStatuses}, status was: ${status}`
            );
            e.message = `${e.message} - node pool phase was: ${status}`;
            throw e;
        }
        return res;
    }

    async waitForClusterNodePoolProperty(
        nodePoolId: string,
        propertyName: string,
        predicate: (propValue: any) => boolean,
        timeout = 30
    ): Promise<Nodepool> {
        let nodePool: Nodepool;
        let nodePoolName: string;

        const isClusterNodePoolPropertyMatch = async (): Promise<boolean> => {
            try {
                const response = await clusterApi.nodepoolsApi.getNodepool(nodePoolId);
                nodePool = response.body;
                const value = get(nodePool, propertyName);
                nodePoolName = nodePool.name;
                logger.debug(
                    `waitForClusterNodePoolProperty: nodePool=${nodePoolName}, property=${propertyName}, value=${JSON.stringify(value)}`
                );
                return predicate(value);
            } catch (e) {
                logger.warn(
                    `Failed to get nodepool ${nodePoolId} for property ${propertyName}: ${e.message}`
                );
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                isClusterNodePoolPropertyMatch,
                `waitForClusterNodePoolProperty nodePoolId: ${nodePoolId}, property: ${propertyName}`,
                true,
                timeout
            );
        } catch (e) {
            logger.debug(
                `waitForClusterNodePoolProperty timed out for nodePool=${nodePoolName ?? nodePoolId}, expected ${propertyName}`
            );
            throw e;
        }

        return nodePool;
    }

    async waitForClusterNodesProperty(
        clusterUid: string,
        nodePoolName: string,
        predicate: (propValue: any) => boolean,
        timeout = 30
    ): Promise<Node[]> {
        let nodes: Node[];

        const isClusterNodePropertyMatch = async (): Promise<boolean> => {
            try {
                clusterApi.nodesApi.accessToken = flowRootState.users.system_administrator.accessToken;
                const response = await clusterApi.nodesApi.getTenantNodes([`clusterUuid==${clusterUid}`]);
                nodes = response.body.nodes;
                return predicate(nodes);
            } catch (e) {
                logger.warn(
                    `Failed to get nodes for cluster ${clusterUid}, nodepool: ${nodePoolName}: ${e.message}`
                );
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                isClusterNodePropertyMatch,
                `waitForClusterNodesProperty for nodepool: ${nodePoolName}`,
                true,
                timeout
            );
        } catch (e) {
            logger.debug(`waitForClusterNodesProperty for nodepool ${nodePoolName} - expected nodes`);
            throw e;
        }

        return nodes;
    }

    async waitForNodeCrdProperty(
        nodepoolLabel: string,
        predicate: (nodes: any) => V1Node,
        timeout = 40,
        delay = 1
    ): Promise<V1Node> {
        let node: V1Node;
        const verifyNodes = async (): Promise<boolean> => {
            try {
                const clusterNodesList = await kubeCli.getNodes(nodepoolLabel);
                node = predicate(clusterNodesList.items);
                return !isNil(node);
            } catch (e) {
                logger.warn(`Failed to get nodes with label ${nodepoolLabel}: ${e.message}`);
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                verifyNodes,
                `waitForNodeCrdProperty nodepoolLabel: ${nodepoolLabel}`,
                true,
                timeout,
                delay
            );
        } catch (e) {
            logger.debug(`Failed waiting for node items with label ${nodepoolLabel}`);
            e.message = `${e.message} - node with label ${nodepoolLabel} values are not met`;
            throw e;
        }
        return node;
    }

    /**
     * `propertyName` is addressed by its path on the legacy NodePool CRD; fields that moved in the
     * KAI CRD are resolved to their new location, see NodePoolCrdFields.getNodePoolCrdField.
     */
    async waitForNodePoolCrdProperty(
        nodePoolName: string,
        propertyName: string,
        predicate: (propValue: any) => boolean,
        timeout = 40,
        delay = 1
    ): Promise<object> {
        let res: object;

        const doesNodePoolValueMatch = async (): Promise<boolean> => {
            try {
                res = await this.getNodePoolCrd(nodePoolName);
                const propValue = getNodePoolCrdField(res, propertyName);
                return predicate(propValue);
            } catch (e) {
                logger.warn(
                    `Failed to get NodePool CRD property ${propertyName} for ${nodePoolName}: ${e.message}`
                );
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                doesNodePoolValueMatch,
                `waitForNodePoolCrdProperty nodePoolName: ${nodePoolName}; propertyName: ${propertyName}`,
                true,
                timeout,
                delay
            );
        } catch (e) {
            logger.debug(`Failed waiting for node pool with statuses ${propertyName}`);
            e.message = `${e.message} - node pool property ${propertyName} value not met`;
            throw e;
        }
        return res;
    }

    async waitForNodePoolCrdReadyOrEmpty(nodePoolName: string, timeout = 40): Promise<object> {
        return await this.waitForNodePoolCrdStatus(
            nodePoolName,
            [NodePool.StatusEnum.Ready, NodePool.StatusEnum.Empty],
            timeout
        );
    }

    async consistentlyNodePoolReady(nodePoolName: string, timeout = 60): Promise<void> {
        let phase: string;
        const isNodePoolReady = async (): Promise<boolean> => {
            try {
                const res = await this.getNodePoolCrd(nodePoolName);
                phase = res['status'].phase;
                return phase === String(NodePool.StatusEnum.Ready);
            } catch (e) {
                logger.warn(`Failed to check if NodePool ${nodePoolName} is ready: ${e.message}`);
                return false;
            }
        };

        try {
            await waitUtils.consistentlyCondition(
                isNodePoolReady,
                `consistentlyNodePoolReady for node pool ${nodePoolName}`,
                true,
                timeout
            );
        } catch (e) {
            logger.debug(
                `Node pool phase during consistentlyNodePoolReady for node pool "${nodePoolName}" was: ${phase}`
            );
            e.message = `${e.message} - node pool phase was: ${phase}`;
            throw e;
        }
        logger.debug(
            `Node pool's "${nodePoolName}" phase was: "${phase}" consistently for ${timeout} seconds`
        );
    }

    validateNodePoolCrdReady(
        nodePoolCrd: object,
        expectedNodePoolName: string,
        expectedNodePoolLabelKey: string,
        expectedNodePoolLabelValue: string
    ): void {
        expect(nodePoolCrd['metadata'].name).to.be.equal(expectedNodePoolName);
        expect(nodePoolCrd['spec'].labelKey).to.be.equal(expectedNodePoolLabelKey);
        expect(nodePoolCrd['spec'].labelValue).to.be.equal(expectedNodePoolLabelValue);
        expect(nodePoolCrd['status'].phase).to.be.oneOf([
            NodePool.StatusEnum.Ready,
            NodePool.StatusEnum.Empty
        ]);
        expect(nodePoolCrd['status'].message).to.be.equal('');
    }

    async waitForDeletedNodePool(nodePoolId: string): Promise<void> {
        const isNodePoolDeleted = async (): Promise<boolean> => {
            try {
                await clusterApi.nodepoolsApi.getNodepool(nodePoolId);
                return false;
            } catch (e) {
                if (
                    e.statusCode &&
                    (e.statusCode === StatusCodes.FORBIDDEN || e.statusCode === StatusCodes.NOT_FOUND)
                ) {
                    return true;
                }
                logger.error(`Failed getting deleted node pool ${nodePoolId}; reason: ${JSON.stringify(e)}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isNodePoolDeleted,
            `waitForDeletedNodePool node pool: ${nodePoolId}`,
            true,
            60
        );
    }

    async createClusterNodePoolWithRequest(
        nodePoolCreateRequest: NodepoolCreateFields,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body?: Nodepool }> {
        logger.debug(`Creating nodepool ${nodePoolCreateRequest.name}`);
        clusterApi.nodepoolsApi.accessToken = accessToken;
        return await clusterApi.nodepoolsApi.createNodepool(nodePoolCreateRequest);
    }

    async createClusterNodePool(
        nodePoolName: string,
        nodePoolLabelKey: string,
        nodePoolLabelValue: string,
        options: GpuNetworkAccelerationDetection = {},
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body?: Nodepool }> {
        const nodePoolCreateRequest = new NodepoolCreateFields();
        nodePoolCreateRequest.name = nodePoolName;
        nodePoolCreateRequest.labelKey = nodePoolLabelKey;
        nodePoolCreateRequest.labelValue = nodePoolLabelValue;
        nodePoolCreateRequest.clusterId = flowRootState.cluster.uuid;
        if (!isEmpty(options)) {
            nodePoolCreateRequest.gpuNetworkAccelerationDetection = options.gpuNetworkAccelerationDetection;
            nodePoolCreateRequest.gpuNetworkAccelerationLabelKey =
                options.gpuNetworkAccelerationLabelKey ?? '';
        }

        logger.debug(`Creating nodepool ${nodePoolName}`);
        clusterApi.nodepoolsApi.accessToken = accessToken;
        return await clusterApi.nodepoolsApi.createNodepool(nodePoolCreateRequest);
    }

    async updateClusterNodePool(
        nodePoolId: string,
        nodepoolUpdateFields: NodepoolUpdateFields,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body?: Nodepool }> {
        logger.debug(`Updating nodepool ${nodePoolId}`);
        clusterApi.nodepoolsApi.accessToken = accessToken;
        return await clusterApi.nodepoolsApi.updateNodepool(nodePoolId, nodepoolUpdateFields);
    }

    // Set the GPU device-level placement strategy on a node pool (node-level is set to binpack since
    // it is irrelevant on a single-node pool) and wait until the cluster CRD reflects it.
    async setGpuDevicePlacementStrategyAndWait(
        nodePoolId: string,
        nodePoolName: string,
        nodePoolLabelKey: string,
        nodePoolLabelValue: string,
        device: PlacementStrategy
    ): Promise<void> {
        const gpuPodsPlacementStrategy = new GpuPodsPlacementStrategy();
        gpuPodsPlacementStrategy.node = PlacementStrategy.Binpack;
        gpuPodsPlacementStrategy.device = device;

        const schedulingConfiguration = new SchedulingConfiguration();
        schedulingConfiguration.gpuPodsPlacementStrategy = gpuPodsPlacementStrategy;

        const updateFields = new NodepoolUpdateFields();
        updateFields.labelKey = nodePoolLabelKey;
        updateFields.labelValue = nodePoolLabelValue;
        updateFields.schedulingConfiguration = schedulingConfiguration;

        await this.updateClusterNodePool(nodePoolId, updateFields);

        // the control-plane maps the device-level strategy onto the cluster CRD's
        // spec.placementStrategy.gpuDevice
        await this.waitForNodePoolCrdProperty(
            nodePoolName,
            'spec.placementStrategy.gpuDevice',
            (value) => value === device.toString(),
            60
        );
    }

    async syncNodepoolsStatus(
        clusterUuid: string,
        syncNodepoolsStatusRequest: SyncNodepoolsStatusRequest,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body?: Nodepool }> {
        logger.debug(`Sync nodepool Status cluster ${clusterUuid}`);
        clusterApi.nodepoolsApi.accessToken = accessToken;
        return await clusterApi.nodepoolsApi.syncNodepoolsStatus(clusterUuid, syncNodepoolsStatusRequest);
    }

    async waitForClusterNodePoolPhase(
        nodePoolId: string,
        exceptedStatuses = [NodepoolPhase.Ready],
        timeout = 50
    ): Promise<Nodepool> {
        let nodePool: Nodepool;
        let phase: NodepoolPhase;
        let nodes: string[];
        let nodePoolName: string;

        const isClusterNodePoolPhase = async (): Promise<boolean> => {
            try {
                const response = await clusterApi.nodepoolsApi.getNodepool(nodePoolId);
                nodePool = response.body;
                logger.debug(`waitForClusterNodePoolPhase gotten nodepool: ${JSON.stringify(nodePool)}`);
                phase = nodePool.phase;
                nodes = nodePool.status.nodes;
                nodePoolName = nodePool.name;
                return exceptedStatuses.includes(phase);
            } catch (e) {
                logger.warn(`Failed to get cluster nodepool ${nodePoolId}: ${e.message}`);
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                isClusterNodePoolPhase,
                `waitForClusterNodePoolPhase nodePool: ${nodePoolName}`,
                true,
                timeout
            );
        } catch (e) {
            logger.debug(
                `waitForClusterNodePoolPhase for node pool ${nodePoolName}, expected phases: ${exceptedStatuses}, status actually was: ${phase}, nodes in node pool were: ${JSON.stringify(
                    nodes
                )}`
            );
            e.message = `${
                e.message
            } - status actually was: ${phase}, nodes in node pool were: ${JSON.stringify(nodes)}`;
            throw e;
        }

        return nodePool;
    }

    async waitForSchedulingShardCrdArgs(
        name: string,
        argPath: string,
        desiredArgValue = undefined
    ): Promise<object> {
        return this.waitForCrdCondition(crds.runaiSchedulingShardCrd, name, argPath, desiredArgValue);
    }

    /**
     * Path of the min runtime section on the SchedulingShard.
     * Until cluster 3.0 the node pool's minGuaranteedRuntime was rendered into the
     * shard's scheduler args; from 3.0 (kai.scheduler/v1) it has its own spec section.
     */
    schedulingShardMinRuntimePath(): string {
        return versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V3_0)
            ? 'spec.minRuntime'
            : 'spec.args';
    }

    async waitForCrdCondition(
        crdDef: CustomResourceDefinition,
        name: string,
        argPath: string,
        desiredArgValue = undefined
    ): Promise<object> {
        let crdInstance: object;
        let argValue: string;

        const isCrdHasCorrectArgValue = async (): Promise<boolean> => {
            try {
                crdInstance = await waitUtils.waitForCrd(crdDef, name, undefined, 10);
                argValue = get(crdInstance, argPath);
                return desiredArgValue === argValue || (isNil(desiredArgValue) && isNil(argValue));
            } catch (e) {
                logger.error(`Failed getting crdInstance crd ${name}; reason: ${JSON.stringify(e)}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isCrdHasCorrectArgValue,
            `waitForCrdCondition crdInstance name: ${name}; arg value was: ${argValue}`,
            true,
            90
        );
        return crdInstance;
    }

    async deleteAllClusterNodePools(
        accessToken = flowRootState.users.system_administrator.accessToken,
        clusterId = flowRootState.cluster.uuid
    ): Promise<void> {
        clusterApi.nodepoolsApi.accessToken = accessToken;
        const allNodePools = await clusterApi.nodepoolsApi.getNodepools(['clusterId==' + clusterId]);
        for (const nodePool of allNodePools.body.nodepools) {
            if (nodePool.name !== 'default') {
                try {
                    logger.debug(`Deleting nodepool ${nodePool.id}`);
                    await clusterApi.nodepoolsApi.deleteNodepool(nodePool.id);
                } catch (e) {
                    logger.debug(
                        `Failed to delete nodepool ${nodePool.id}, it may already be deleted: ${e.message}`
                    );
                }
            }
        }
    }

    async createBackendNodePool(
        nodePoolName: string,
        nodePoolLabelKey: string,
        nodePoolLabelValue: string
    ): Promise<{ response: http.IncomingMessage; body?: NodePool }> {
        const nodePoolCreateRequest = new NodePoolCreateRequest();
        nodePoolCreateRequest.name = nodePoolName;
        nodePoolCreateRequest.labelKey = nodePoolLabelKey;
        nodePoolCreateRequest.labelValue = nodePoolLabelValue;

        logger.debug(`Creating nodepool ${nodePoolName}`);
        backendClusterApi.nodePoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const res = await backendClusterApi.nodePoolsApi.createNodePool(
            flowRootState.cluster.uuid,
            nodePoolCreateRequest
        );
        return { response: res.response, body: res.body as NodePool };
    }

    async createClusterNodePoolAndValidate(
        nodePoolLabelKey: string,
        nodePoolLabelValue: string,
        nodePoolNamePrefix = 'np',
        options: {
            gpuNetworkAccelerationLabelKey?: string;
            gpuNetworkAccelerationDetection?: NodepoolGPUNetworkAccelerationDetection;
        } = {}
    ): Promise<{ nodePoolName: string; nodePoolId: string }> {
        const nodePoolName = objectGenerator.generateShortNameByPrefix(nodePoolNamePrefix);
        // create nodepool
        const createResponse = await this.createClusterNodePool(
            nodePoolName,
            nodePoolLabelKey,
            nodePoolLabelValue,
            options
        );
        expect(createResponse.response.statusCode.valueOf()).equal(StatusCodes.CREATED);
        expect(createResponse.body).property('id');

        return {
            nodePoolName: nodePoolName,
            nodePoolId: createResponse.body.id
        };
    }

    async updateClusterNodePoolAndValidate(
        nodePoolId: string,
        nodepoolUpdateFields: NodepoolUpdateFields
    ): Promise<{ nodePoolName: string; nodePoolId: string }> {
        // update nodepool
        const updateResponse = await this.updateClusterNodePool(nodePoolId, nodepoolUpdateFields);
        expect(updateResponse.response.statusCode.valueOf()).equal(StatusCodes.OK);
        expect(updateResponse.body).property('id');

        return {
            nodePoolName: updateResponse.body.name,
            nodePoolId: updateResponse.body.id
        };
    }

    async createBackendNodePoolAndValidate(
        nodePoolLabelKey: string,
        nodePoolLabelValue: string,
        nodePoolNamePrefix = 'np'
    ): Promise<{ nodePoolName: string; nodePoolId: number }> {
        const nodePoolName = objectGenerator.generateShortNameByPrefix(nodePoolNamePrefix);

        // create nodepool
        const createResponse = await this.createBackendNodePool(
            nodePoolName,
            nodePoolLabelKey,
            nodePoolLabelValue
        );
        expect(createResponse.response.statusCode.valueOf()).equal(StatusCodes.CREATED);
        expect(createResponse.body).property('id');

        return {
            nodePoolName: nodePoolName,
            nodePoolId: createResponse.body.id
        };
    }

    async updateNodePool(
        nodePoolId: number,
        newNodePoolLabelKey: string,
        newNodePoolLabelValue: string
    ): Promise<http.IncomingMessage> {
        const nodePoolLabelsRequest = new NodePoolLabelsRequest();
        nodePoolLabelsRequest.labelKey = newNodePoolLabelKey;
        nodePoolLabelsRequest.labelValue = newNodePoolLabelValue;

        const res = await backendClusterApi.nodePoolsApi.updateNodePoolLabels(
            flowRootState.cluster.uuid,
            nodePoolId,
            nodePoolLabelsRequest
        );
        return res.response;
    }

    async updateNodePoolAndValidate(
        nodePoolId: number,
        newNodePoolLabelKey: string,
        newNodePoolLabelValue: string
    ): Promise<void> {
        const res = await this.updateNodePool(nodePoolId, newNodePoolLabelKey, newNodePoolLabelValue);
        expect(res.statusCode.valueOf()).equal(StatusCodes.ACCEPTED);
    }

    async cleanupDeleteNodePool(nodePoolCreated: {
        nodePoolName: string;
        nodePoolId: string | number;
    }): Promise<void> {
        if (!nodePoolCreated || !nodePoolCreated.nodePoolId) {
            return;
        }
        try {
            await backendClusterApi.nodePoolsApi.deleteNodePool(
                flowRootState.cluster.uuid,
                +nodePoolCreated.nodePoolId
            );

            await this.waitForNodePoolCrd(nodePoolCreated.nodePoolName, 40, false);

            await waitUtils.waitForCondition(
                async () => {
                    const nodepools = await backendClusterApi.nodePoolsApi.getNodePools(
                        flowRootState.cluster.uuid
                    );

                    return nodepools.body.some((np) => np.id === nodePoolCreated.nodePoolId);
                },
                'Wait for node pool to be deleted from the backend',
                false
            );
        } catch (e) {
            logger.error(
                `Could not cleanup, unable to delete nodepool ${JSON.stringify(
                    nodePoolCreated
                )}; reason: ${JSON.stringify(e)}`
            );
        }
    }

    async deleteAllNodePools(): Promise<void> {
        backendClusterApi.nodePoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const allNodePools = await this.getAllBackendNodePoolsAndValidate();
        const promisesToAwait = [];
        for (const nodePool of allNodePools) {
            if (nodePool.name !== this.DefaultNodePoolName) {
                promisesToAwait.push(this.deleteNodePool(nodePool));
                promisesToAwait.push(this.waitForNodePoolCrd(nodePool.name, 30, false));
            }
        }
        try {
            await Promise.all(promisesToAwait);
            logger.debug('All nodepools deleted');
        } catch (e) {
            logger.debug(`Failed to delete all nodepools, error: ${JSON.stringify(e)}`);
        }
    }

    async createNodeType(labelName: string, accessToken?: string): Promise<NodeType> {
        logger.debug(`Creating node type ${labelName}`);
        orgUnitService.nodeTypeApi.accessToken = accessToken
            ? accessToken
            : flowRootState.users.editor.accessToken;
        const nodeType = (
            await orgUnitService.nodeTypeApi.createNodeType({
                name: labelName,
                clusterId: flowRootState.cluster.uuid
            })
        ).body;
        logger.info(`Node Type ${nodeType.id} with label name ${nodeType.name} created`);
        return nodeType;
    }

    async getAllBackendNodePoolsAndValidate(
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Array<NodePool>> {
        backendClusterApi.nodePoolsApi.accessToken = accessToken;
        const getResponse = await backendClusterApi.nodePoolsApi.getNodePools(flowRootState.cluster.uuid);
        expect(getResponse.response.statusCode.valueOf()).equal(StatusCodes.OK);
        return getResponse.body;
    }

    async getBackendNodePoolAndValidate(nodePoolId: number): Promise<NodePool> {
        const allNodePools = await this.getAllBackendNodePoolsAndValidate();

        const currentNodePool = allNodePools.find((nodePool) => nodePool.id === nodePoolId);
        expect(currentNodePool).not.to.be.undefined;
        return currentNodePool;
    }

    async getAndValidateBackendNodePool(
        nodePoolId: number,
        expectedNodePoolName: string,
        expectedNodePoolLabelKey: string,
        expectedNodePoolLabelValue: string,
        isDefault = false
    ): Promise<NodePool> {
        const currentNodePool = await this.getBackendNodePoolAndValidate(nodePoolId);
        expect(currentNodePool.labelKey).to.be.equal(expectedNodePoolLabelKey);
        expect(currentNodePool.labelValue).to.be.equal(expectedNodePoolLabelValue);
        expect(currentNodePool.name).to.be.equal(expectedNodePoolName);
        expect(currentNodePool.isDefault).to.be.equal(isDefault);

        return currentNodePool;
    }

    async getDefaultNodePoolFromBackend(): Promise<NodePool> {
        backendClusterApi.nodePoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const allNodePools = await this.getAllBackendNodePoolsAndValidate();
        const defaultNodePool = allNodePools.find((nodePool) => nodePool.name === this.DefaultNodePoolName);
        expect(defaultNodePool).not.to.be.undefined;
        return defaultNodePool;
    }

    async getAllNodePoolsByClusterAndValidate(
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<Array<Nodepool>> {
        clusterApi.nodepoolsApi.accessToken = accessToken;
        const getResponse = await clusterApi.nodepoolsApi.getNodepools([
            'clusterId==' + flowRootState.cluster.uuid
        ]);
        expect(getResponse.response.statusCode.valueOf()).equal(StatusCodes.OK);
        return getResponse.body.nodepools;
    }

    async getDefaultNodePoolFromCluster(): Promise<Nodepool> {
        const nodePools = await clusterApi.nodepoolsApi.getNodepools([
            'clusterId==' + flowRootState.cluster.uuid,
            'name==default'
        ]);
        expect(nodePools.body.nodepools).not.to.be.undefined;
        expect(nodePools.body.nodepools.length).to.equal(1);
        return nodePools.body.nodepools[0];
    }

    async waitForBackendNodePoolStatus(
        nodePoolId: number,
        exceptedStatuses = [NodePool.StatusEnum.Ready],
        timeout = 30
    ): Promise<NodePool> {
        let nodePool: NodePool;
        let status;
        let nodes;
        let nodePoolName;

        const isBackendNodePoolStatus = async (): Promise<boolean> => {
            try {
                nodePool = await this.getBackendNodePoolAndValidate(nodePoolId);
                logger.debug(`waitForBackendNodePoolStatus gotten nodepool: ${JSON.stringify(nodePool)}`);
                status = nodePool.status;
                nodes = nodePool.nodes;
                nodePoolName = nodePool.name;
                return exceptedStatuses.includes(status);
            } catch (e) {
                logger.warn(`Failed to get backend nodepool ${nodePoolId}: ${e.message}`);
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                isBackendNodePoolStatus,
                `waitForBackendNodePoolStatus nodePool: ${nodePoolName}`,
                true,
                timeout
            );
        } catch (e) {
            logger.debug(
                `waitForBackendNodePoolStatus for node pool ${nodePoolName}, expected statuses: ${exceptedStatuses}, status actually was: ${status}, nodes in node pool were: ${JSON.stringify(
                    nodes
                )}`
            );
            e.message = `${
                e.message
            } - status actually was: ${status}, nodes in node pool were: ${JSON.stringify(nodes)}`;
            throw e;
        }

        return nodePool;
    }

    async waitForBackendNodePoolReadyOrEmpty(
        nodePoolId: number,
        timeout = 30,
        checkNonEmptyNodePools = false
    ): Promise<NodePool> {
        return await this.waitForBackendNodePoolStatus(
            nodePoolId,
            checkNonEmptyNodePools
                ? [NodePool.StatusEnum.Ready]
                : [NodePool.StatusEnum.Ready, NodePool.StatusEnum.Empty],
            timeout
        );
    }

    async waitForBackendJobWithNodePool(
        jobName: string,
        expectedNodePool: string,
        timeout = 30,
        shouldBeRunning = true
    ): Promise<DisplayedJob> {
        let jobsNodePool: string;
        let jobStatus: string;
        let backendJob: DisplayedJob;

        const isBackendJobWithCorrectNodePool = async (): Promise<boolean> => {
            try {
                backendJob = await waitUtils.waitForBackendJob(jobName);
                jobsNodePool = backendJob.nodePool;
                jobStatus = backendJob.status;
                return jobsNodePool === expectedNodePool && (!shouldBeRunning || jobStatus === 'Running');
            } catch (e) {
                logger.warn(`Failed to get backend job ${jobName}: ${e.message}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackendJobWithCorrectNodePool,
            `waitForBackendJobWithNodePool jobName: ${jobName}, expectedNodePool: ${expectedNodePool}, jobs node pool was: ${jobsNodePool}; job's status was: ${jobStatus}`,
            true,
            timeout
        );

        return backendJob;
    }

    async waitForBackendWorkloadWithNodePool(
        workloadName: string,
        projectId: string,
        expectedNodePool: string,
        timeout = 30,
        shouldBeRunning = true
    ): Promise<Workload> {
        let workloadNodePool: string;
        let workloadPhase: Phase;
        let backendWorkload: Workload;

        const isWorkloadWithCorrectNodePool = async (): Promise<boolean> => {
            try {
                backendWorkload = await waitUtils.waitForBackendWorkload(workloadName, projectId);
                workloadNodePool = workloadNodePool = backendWorkload?.currentNodePools?.length
                    ? backendWorkload.currentNodePools[0]
                    : undefined;
                workloadPhase = backendWorkload.phase;
                return (
                    workloadNodePool === expectedNodePool &&
                    (!shouldBeRunning || workloadPhase === Phase.Running)
                );
            } catch (e) {
                logger.warn(
                    `Failed to get backend workload ${workloadName} in project ${projectId}: ${e.message}`
                );
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                isWorkloadWithCorrectNodePool,
                `waitForBackendWorkloadWithNodePool jobName: ${workloadName}, expectedNodePool: ${expectedNodePool}`,
                true,
                timeout
            );
        } catch (e) {
            e.message = `${e.message} - workload node pool was: ${workloadNodePool}; workload phase: ${workloadPhase}`;
            throw e;
        }

        return backendWorkload;
    }

    async waitForBackendPodsOfJobWithNodePool(
        backendJob: DisplayedJob,
        expectedNodePool: string,
        timeout = 30
    ): Promise<void> {
        let podsNodePool: string;

        const isBackendPodsOfJobWithCorrectNodePool = async (): Promise<boolean> => {
            try {
                const backendJobPods = (
                    await backendClusterApi.jobsApi.getPodsByJobId(
                        backendJob.podGroupId,
                        flowRootState.cluster.uuid
                    )
                ).body;
                for (const pod of backendJobPods) {
                    podsNodePool = pod.nodePool;
                    if (!expectedNodePool) {
                        if (podsNodePool) {
                            return false;
                        }
                    } else if (podsNodePool !== expectedNodePool) {
                        return false;
                    }
                }
                return true;
            } catch (e) {
                logger.warn(
                    `Failed to get pods for job ${backendJob.jobName} (podGroupId: ${backendJob.podGroupId}): ${e.message}`
                );
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                isBackendPodsOfJobWithCorrectNodePool,
                `waitForBackendPodsOfJobWithNodePool jobName: ${backendJob.jobName}, expectedNodePool: ${expectedNodePool}`,
                true,
                timeout
            );
        } catch (e) {
            e.message = `${e.message} - pods node pool was: ${podsNodePool}`;
            throw e;
        }
    }

    async waitForBackendPodsOfWorkloadWithNodePool(
        workload: Workload,
        expectedNodePool: string,
        timeout = 30
    ): Promise<void> {
        let podsNodePool: string;

        const isBackendPodsOfJobWithCorrectNodePool = async (): Promise<boolean> => {
            try {
                const backendJobPods = (await workloadsUtilsV2.getWorkloadPods(workload.id)).body;
                for (const pod of backendJobPods.pods) {
                    podsNodePool = pod.currentNodePool;
                    if (!expectedNodePool) {
                        if (podsNodePool) {
                            return false;
                        }
                    } else if (podsNodePool !== expectedNodePool) {
                        return false;
                    }
                }
                return true;
            } catch (e) {
                logger.warn(
                    `Failed to get pods for workload ${workload.name} (id: ${workload.id}): ${e.message}`
                );
                return false;
            }
        };

        try {
            await waitUtils.waitForCondition(
                isBackendPodsOfJobWithCorrectNodePool,
                `waitForBackendPodsOfWorkloadWithNodePool workloadName: ${workload.name}, expectedNodePool: ${expectedNodePool}`,
                true,
                timeout
            );
        } catch (e) {
            e.message = `${e.message} - pods node pool was: ${podsNodePool}`;
            throw e;
        }
    }

    async waitForBackendJobAndPodsWithCorrectNodePool(
        jobName: string,
        expectedNodePool: string,
        timeout = 30,
        shouldBeRunning = true
    ): Promise<void> {
        const backendJob = await this.waitForBackendJobWithNodePool(
            jobName,
            expectedNodePool,
            timeout,
            shouldBeRunning
        );

        await this.waitForBackendPodsOfJobWithNodePool(backendJob, expectedNodePool, timeout);
    }

    async waitForBackendWorkloadAndPodsWithCorrectNodePool(
        jobName: string,
        projectId: number,
        expectedNodePool: string,
        timeout = 30,
        shouldBeRunning = true
    ): Promise<void> {
        const workload = await this.waitForBackendWorkloadWithNodePool(
            jobName,
            projectId + '',
            expectedNodePool,
            timeout,
            shouldBeRunning
        );
        await this.waitForBackendPodsOfWorkloadWithNodePool(workload, expectedNodePool, timeout);
    }

    async createEmptyNodePool(
        nodePoolLabelKey: string,
        nodePoolLabelValue: string
    ): Promise<{ nodePoolName: string; nodePoolId: number; errorStr: string }> {
        let nodePoolName = '';
        let nodePoolId = 0;

        try {
            const res = await this.createBackendNodePoolAndValidate(
                nodePoolLabelKey,
                nodePoolLabelValue,
                'empty-np'
            );
            nodePoolName = res.nodePoolName;
            nodePoolId = res.nodePoolId;

            // wait for crd to exist in cluster - with status "Empty"
            await this.waitForNodePoolCrdStatus(nodePoolName, [NodePool.StatusEnum.Empty]);

            return {
                nodePoolName: nodePoolName,
                nodePoolId: nodePoolId,
                errorStr: ''
            };
        } catch (error) {
            return {
                nodePoolName: nodePoolName,
                nodePoolId: nodePoolId,
                errorStr: error
            };
        }
    }

    async createNodePoolAndLabelNodeToAssignToNodePool(
        nodePoolLabelKey: string,
        nodePoolLabelValue: string,
        nodesAlreadyLabeled = []
    ): Promise<{ nodePoolName: string; nodePoolId: number; nodeName: string }> {
        let nodeName = '',
            nodePoolName = '';
        let nodePoolId = 0;
        try {
            // label a node to be assigned to the nodepool
            nodeName = await this.labelAGPUNode(nodePoolLabelKey, nodePoolLabelValue, nodesAlreadyLabeled);
            expect(nodeName).not.to.equal('');

            // create nodepool
            const res = await this.createBackendNodePoolAndValidate(nodePoolLabelKey, nodePoolLabelValue);
            nodePoolName = res.nodePoolName;
            nodePoolId = res.nodePoolId;

            // wait for crd to be Ready in cluster - it has a node assigned to it
            const nodePoolCrd = await this.waitForNodePoolCrdStatus(nodePoolName, [
                NodePool.StatusEnum.Ready
            ]);
            logger.debug(`Nodepool ${nodePoolName} is in status Ready`);
            await this.validateNodeLabeledWithNodePool(nodePoolName, nodeName);
            logger.debug(
                `Node ${nodeName} is labeled as part of nodepool ${nodePoolName}; label: ${nodePoolLabelKey}/${nodePoolLabelValue}`
            );

            await this.validateNodesStatusOfNodePoolById(nodePoolCrd, nodePoolId, [nodeName]);

            return {
                nodePoolName: nodePoolName,
                nodePoolId: nodePoolId,
                nodeName: nodeName
            };
        } catch (e) {
            logger.error(`Failed to create nodepool with node ${nodeName}: ${e.message}`);
            return {
                nodePoolName: nodePoolName,
                nodePoolId: nodePoolId,
                nodeName: nodeName
            };
        }
    }

    async createEmptyClusterNodePool(
        nodePoolLabelKey: string,
        nodePoolLabelValue: string
    ): Promise<{ nodePoolName: string; nodePoolId: string; errorStr: string }> {
        let nodePoolName = '';
        let nodePoolId = '';

        try {
            const res = await this.createClusterNodePoolAndValidate(
                nodePoolLabelKey,
                nodePoolLabelValue,
                'empty-np'
            );
            nodePoolName = res.nodePoolName;
            nodePoolId = res.nodePoolId;

            // wait for crd to exist in cluster - with status "Empty"
            await this.waitForClusterNodePoolPhase(nodePoolId, [NodepoolPhase.Empty]);

            return {
                nodePoolName: nodePoolName,
                nodePoolId: nodePoolId,
                errorStr: ''
            };
        } catch (error) {
            return {
                nodePoolName: nodePoolName,
                nodePoolId: nodePoolId,
                errorStr: error
            };
        }
    }

    async createClusterNodePoolAndLabelNodeToAssignToNodePool(
        nodePoolLabelKey: string,
        nodePoolLabelValue: string,
        nodesAlreadyLabeled = []
    ): Promise<{ nodePoolName: string; nodePoolId: string; nodeName: string }> {
        let nodeName = '',
            nodePoolName = '';
        let nodePoolId = '';
        try {
            // label a node to be assigned to the nodepool
            nodeName = await this.labelAGPUNode(nodePoolLabelKey, nodePoolLabelValue, nodesAlreadyLabeled);
            expect(nodeName).not.to.equal('');

            // create nodepool
            const res = await this.createClusterNodePoolAndValidate(nodePoolLabelKey, nodePoolLabelValue);
            nodePoolName = res.nodePoolName;
            nodePoolId = res.nodePoolId;

            // wait for crd to be Ready in cluster - it has a node assigned to it
            const nodePoolCrd = await this.waitForNodePoolCrdStatus(nodePoolName, [
                NodePool.StatusEnum.Ready
            ]);
            logger.debug(`Nodepool ${nodePoolName} is in status Ready`);
            await this.validateNodeLabeledWithNodePool(nodePoolName, nodeName);
            logger.debug(
                `Node ${nodeName} is labeled as part of nodepool ${nodePoolName}; label: ${nodePoolLabelKey}/${nodePoolLabelValue}`
            );

            this.validateNodesStatusOfNodePoolCrd(nodePoolCrd, [nodeName]);

            return {
                nodePoolName: nodePoolName,
                nodePoolId: nodePoolId,
                nodeName: nodeName
            };
        } catch (e) {
            logger.error(`Failed to create nodepool with node ${nodeName}: ${e.message}`);
            return {
                nodePoolName: nodePoolName,
                nodePoolId: nodePoolId,
                nodeName: nodeName
            };
        }
    }

    async cleanupDeleteClusterNodePool(nodePoolCreated: {
        nodePoolName: string;
        nodePoolId: string;
    }): Promise<void> {
        if (!nodePoolCreated || !nodePoolCreated.nodePoolId) {
            return;
        }
        try {
            await clusterApi.nodepoolsApi.deleteNodepool(nodePoolCreated.nodePoolId);
            await this.waitForDeletedNodePool(nodePoolCreated.nodePoolId);
        } catch (e) {
            logger.error(
                `Could not cleanup, unable to delete nodepool ${JSON.stringify(
                    nodePoolCreated
                )}; reason: ${JSON.stringify(e)}`
            );
        }
    }

    async updateClusterNodePoolLabelsAndValidate(
        nodePoolId: string,
        labelKey: string,
        labelValue: string
    ): Promise<void> {
        const updateFields: NodepoolUpdateFields = {
            labelKey: labelKey,
            labelValue: labelValue
        };
        const res = await this.updateClusterNodePool(nodePoolId, updateFields);
        expect(res.response.statusCode.valueOf()).equal(StatusCodes.OK);
    }

    async waitForDesiredNodesStatusOfClusterNodePool(
        nodePoolCrd: object,
        expectedNodeNames?: Array<string>,
        timeout = 30
    ): Promise<void> {
        const isDesiredNodesStatusOfNodePool = (): boolean => {
            try {
                this.validateNodesStatusOfNodePoolCrd(nodePoolCrd, expectedNodeNames);
                return true;
            } catch (e) {
                logger.warn(`Failed to validate nodes status for nodepool CRD: ${e.message}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isDesiredNodesStatusOfNodePool,
            `waitForDesiredNodesStatusOfClusterNodePool expectedNodeNames: ${expectedNodeNames}`,
            true,
            timeout
        );
    }

    validateNodesStatusOfNodePoolInner(
        nodePoolNodes: Array<string>,
        expectedNodeNames?: Array<string>,
        isCrd?: boolean
    ): void {
        if (expectedNodeNames === null || expectedNodeNames === undefined || expectedNodeNames.length === 0) {
            if (isCrd) {
                expect(nodePoolNodes).to.equal(undefined);
            } else {
                expect(nodePoolNodes).to.equal(null);
            }
            return;
        } else if (nodePoolNodes === null || nodePoolNodes === undefined || nodePoolNodes.length === 0) {
            // validation cannot be performed on backend since there is no wait on cluster-sync results
            if (isCrd) {
                expect(nodePoolNodes).to.equal(expectedNodeNames);
            }
            return;
        }

        expectedNodeNames.sort();
        nodePoolNodes.sort();
        expect(nodePoolNodes).to.eql(expectedNodeNames);
    }

    validateNodesStatusOfBackendNodePool(backendNodePool: NodePool, expectedNodeNames?: Array<string>): void {
        const nodePoolNodesFromBackend = JSON.parse(backendNodePool.nodes);
        this.validateNodesStatusOfNodePoolInner(nodePoolNodesFromBackend, expectedNodeNames, false);
    }

    validateNodesStatusOfNodePoolCrd(nodePoolCrd: object, expectedNodeNames?: Array<string>): void {
        const nodePoolNodesFromCrd = nodePoolCrd['status'].nodes?.map((node) => node['name']);
        this.validateNodesStatusOfNodePoolInner(nodePoolNodesFromCrd, expectedNodeNames, true);
    }

    validateNodesStatusOfNodePool(
        nodePoolCrd: object,
        backendNodePool: NodePool,
        expectedNodeNames?: Array<string>
    ): void {
        this.validateNodesStatusOfBackendNodePool(backendNodePool, expectedNodeNames);
        this.validateNodesStatusOfNodePoolCrd(nodePoolCrd, expectedNodeNames);
    }

    async validateNodesStatusOfNodePoolById(
        nodePoolCrd: object,
        nodePoolId: number,
        expectedNodeNames?: Array<string>
    ): Promise<void> {
        const backendNodePool = await this.getBackendNodePoolAndValidate(nodePoolId);
        this.validateNodesStatusOfNodePool(nodePoolCrd, backendNodePool, expectedNodeNames);
    }

    async waitForDesiredNodesStatusOfNodePool(
        nodePoolCrd: object,
        nodePoolId: number,
        expectedNodeNames?: Array<string>,
        timeout = 30
    ): Promise<void> {
        const isDesiredNodesStatusOfNodePool = async (): Promise<boolean> => {
            try {
                await this.validateNodesStatusOfNodePoolById(nodePoolCrd, nodePoolId, expectedNodeNames);
                return true;
            } catch (e) {
                logger.warn(`Failed to validate nodes status for nodepool ${nodePoolId}: ${e.message}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isDesiredNodesStatusOfNodePool,
            `waitForDesiredNodesStatusOfNodePool nodePoolId: ${nodePoolId}, expectedNodeNames: ${expectedNodeNames}`,
            true,
            timeout
        );
    }

    async validateDefaultNodePoolNodes(expectedNodeNames: string[]): Promise<void> {
        const defaultNodePoolCrd = await this.getNodePoolCrd(this.DefaultNodePoolName);
        const defaultNodePoolFromBackend = await this.getDefaultNodePoolFromBackend();
        this.validateNodesStatusOfNodePool(defaultNodePoolCrd, defaultNodePoolFromBackend, expectedNodeNames);
    }

    async waitForDefaultNodePoolNodes(expectedNodeNames: string[], timeout: number = 30): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                try {
                    await this.validateDefaultNodePoolNodes(expectedNodeNames);
                    return true;
                } catch (e) {
                    logger.debug(`Node pool validation failed: ${e}`);
                    return false;
                }
            },
            `Waiting for default node pool to match expected nodes: ${JSON.stringify(expectedNodeNames)}`,
            true,
            timeout
        );
    }

    async createDepartmentWithAllNodePools(
        nodePoolsToCreateWithResources: Array<NodePoolParamsForCreateProjectWithResources>
    ): Promise<Department> {
        const departmentNodePoolResources = [];
        for (const nodePoolResourcesParams of nodePoolsToCreateWithResources) {
            const nodePoolResources = Object.assign({
                nodePool: {
                    id: nodePoolResourcesParams.nodePoolId + '',
                    name: nodePoolResourcesParams.nodePoolName
                },
                gpu: {
                    deserved: nodePoolResourcesParams.deservedGPUs
                }
            });
            departmentNodePoolResources.push(nodePoolResources);
        }

        const departmentName = objectGenerator.generateNameByPrefix('d');

        const createRequest = {
            name: departmentName,
            resources: departmentNodePoolResources,
            clusterId: flowRootState.cluster.uuid
        };

        return await orgUnitUtils.createDepartment(createRequest);
    }

    async createProjectWithSomeNodePoolResources(
        nodePoolsToCreateWithResources: Array<NodePoolParamsForCreateProjectWithResources>,
        options: {
            nodeAffinityForProject?: any;
            defaultNodePools?: string[];
            departmentId?: string;
        } = { nodeAffinityForProject: undefined, defaultNodePools: [], departmentId: undefined }
    ): Promise<ProjectFromProjectsDepartments> {
        const nodePoolResourcesForProject = [];
        for (const nodePoolResourcesParams of nodePoolsToCreateWithResources) {
            const nodePoolResources = Object.assign(
                projectResourcesUtils.createResourcesWithOverQuotaAllowed(
                    nodePoolResourcesParams.deservedGPUs
                ),
                {
                    nodePool: {
                        id: nodePoolResourcesParams.nodePoolId,
                        name: nodePoolResourcesParams.nodePoolName
                    }
                }
            );
            nodePoolResourcesForProject.push(nodePoolResources);
        }

        const projectName = objectGenerator.generateNameByPrefix('p');
        const namespace = Utils.namespaceFromProjectName(projectName);

        const projectCreateRequest = {
            name: projectName,
            nodePoolsResources: nodePoolResourcesForProject
        } as Partial<ProjectCreateRequest>;
        if (options.nodeAffinityForProject) {
            projectCreateRequest['nodeAffinity'] = options.nodeAffinityForProject;
        }
        if (options.defaultNodePools) {
            projectCreateRequest['defaultNodePools'] = options.defaultNodePools;
        }
        if (options.departmentId) {
            projectCreateRequest.departmentId = Number(options.departmentId);
        }

        return await projectResourcesUtils.createProject(namespace, projectCreateRequest);
    }

    async createProjectWithNodePoolResourcesAndTrainNodeAffinity(
        nodePoolsToCreateWithResources: Array<NodePoolParamsForCreateProjectWithResources>,
        nodeTypeLabels: string[]
    ): Promise<ProjectFromProjectsDepartments> {
        const selectedTypes = new Array(0);
        for (const nodeTypeLabel of nodeTypeLabels) {
            const nodeType = await this.createNodeType(nodeTypeLabel);
            selectedTypes.push({ id: nodeType.id });
        }

        const nodeAffinityForProject = {
            train: {
                affinityType: NodeAffinity.AffinityTypeEnum.OnlySelected,
                selectedTypes: selectedTypes
            }
        };

        return await this.createProjectWithSomeNodePoolResources(nodePoolsToCreateWithResources, {
            nodeAffinityForProject: nodeAffinityForProject
        });
    }

    async labelNodeAsRunaiNodeType(nodeName: string, nodeTypeLabel: string): Promise<string> {
        await kubeCli.labelNode(nodeName, this.RunaiNodeTypeLabelKey, nodeTypeLabel);
        return this.RunaiNodeTypeLabelKey;
    }

    async createNodePool(
        nodePoolName: string,
        labelKey: string,
        labelValue: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<NodePool> {
        backendClusterApi.nodePoolsApi.accessToken = accessToken;
        const nodePoolRequest = {
            name: nodePoolName,
            labelKey: labelKey,
            labelValue: labelValue
        };
        const createNodePoolResponse = await backendClusterApi.nodePoolsApi.createNodePool(
            flowRootState.cluster.uuid,
            nodePoolRequest
        );
        return createNodePoolResponse.body;
    }

    async getAllNodeTypes(): Promise<Array<NodeType>> {
        logger.debug('Getting all node types from cluster');
        orgUnitService.nodeTypeApi.accessToken = flowRootState.users.editor.accessToken;
        try {
            const res = await orgUnitService.nodeTypeApi.getNodeTypes(flowRootState.cluster.uuid);
            return res?.body?.nodeTypes ?? [];
        } catch (e) {
            logger.debug(`Failed to get node types: ${e}`);
            return [];
        }
    }

    async deleteNodeType(nodeTypeId: string): Promise<void> {
        logger.debug(`Deleting node type ${nodeTypeId}`);
        orgUnitService.nodeTypeApi.accessToken = flowRootState.users.editor.accessToken;
        await orgUnitService.nodeTypeApi.deleteNodeType(nodeTypeId);
    }

    getObjectsNodePool(obj: object): string {
        const metadata = obj['metadata'];
        if (metadata === undefined) {
            return this.DefaultNodePoolName;
        }

        const labels = metadata['labels'];
        if (labels === undefined) {
            return this.DefaultNodePoolName;
        }

        const nodePoolName = labels[this.NodePoolNameLabel];
        if (nodePoolName === undefined || nodePoolName === '') {
            return this.DefaultNodePoolName;
        }
        return nodePoolName;
    }

    async getCurrentNodePoolOfJob(jobName: string, namespace: string, instances = 1): Promise<string> {
        const podGroups = await waitUtils.waitForPodGroups(jobName, namespace, instances);
        const nodePoolName = this.getObjectsNodePool(podGroups[0]);

        if (!podGroups.every((podGroup) => this.getObjectsNodePool(podGroup) === nodePoolName)) {
            return '';
        }
        return nodePoolName;
    }

    async getNodePoolOfRunningPodsInJob(jobName: string, namespace: string, instances = 1): Promise<string> {
        const pods = await waitUtils.waitForPods(jobName, namespace, instances, 30, [], true);
        const runningPods = pods.filter((pod) => pod.status.phase === 'Running');
        if (runningPods.length === 0) {
            logger.warning(`Couldn't find running pods in job ${namespace}/${jobName}`);
            return '';
        }
        const nodePoolName = this.getObjectsNodePool(runningPods[0]);

        if (!runningPods.every((runningPod) => this.getObjectsNodePool(runningPod) === nodePoolName)) {
            logger.warning(
                `All pods in job ${namespace}/${jobName} did not have same nodepool: ${nodePoolName}`
            );
            return '';
        }
        return nodePoolName;
    }

    submitJobWithMultiNodePools(
        projectName: string,
        image: string,
        nodePools: string[] = [],
        wlType: cliV2WorkloadType = cliV2WorkloadType.training,
        gpus = '1'
    ): string {
        const jobName = objectGenerator.generateNameByPrefix('job');
        const submitFlag = {
            [SubmitFlagsMerged.image]: image,
            [SubmitFlagsMerged.project]: projectName,
            [SubmitFlagsMerged.gpuRequest]: gpus,
            [SubmitFlagsMerged.cpuCoreRequest]: '0.05',
            [SubmitFlagsMerged.cpuMemoryRequest]: '0.05'
        } as Partial<typeof SubmitFlagsMerged>;
        if (nodePools.length !== 0) {
            submitFlag[SubmitFlagsMerged.nodePools] = nodePools;
        }
        runaiCliV2.submit(submitFlag, jobName, wlType);
        return jobName;
    }

    async isJobLabeledCorrectly(
        jobName: string,
        namespace: string,
        expectedNodePool: string,
        instances = 1
    ): Promise<boolean> {
        const jobNodePool = await this.getCurrentNodePoolOfJob(jobName, namespace, instances);
        const result = jobNodePool === expectedNodePool;
        if (!result) {
            logger.warn(
                `Job ${namespace}/${jobName} had node pool ${jobNodePool}, while should have had node pool: ${expectedNodePool}`
            );
        }
        return result;
    }

    async areRunningPodsInJobLabeledCorrectly(
        jobName: string,
        namespace: string,
        expectedNodePool: string,
        instances = 1
    ): Promise<boolean> {
        const runningPodsNodePool = await this.getNodePoolOfRunningPodsInJob(jobName, namespace, instances);
        const result = runningPodsNodePool === expectedNodePool;

        if (!result) {
            logger.warn(
                `Job's ${namespace}/${jobName} pods had node pool ${runningPodsNodePool}, while should have had node pool: ${expectedNodePool}`
            );
        }
        return result;
    }

    async waitForJobAndRunningPodsLabeledCorrectly(
        jobName: string,
        namespace: string,
        expectedNodePool: string,
        instances = 1,
        desiredResult = true
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                return await this.areRunningPodsInJobLabeledCorrectly(
                    jobName,
                    namespace,
                    expectedNodePool,
                    instances
                );
            },
            `waitForJobAndRunningPodsLabeledCorrectly areRunningPodsInJobLabeledCorrectly jobName: ${jobName}, namespace: ${namespace}, expectedNodePool: ${expectedNodePool}`,
            desiredResult,
            60,
            2
        );
        await waitUtils.waitForCondition(
            async () => {
                return await this.isJobLabeledCorrectly(jobName, namespace, expectedNodePool, instances);
            },
            `waitForJobAndRunningPodsLabeledCorrectly isJobLabeledCorrectly jobName: ${jobName}, namespace: ${namespace}, expectedNodePool: ${expectedNodePool}`,
            desiredResult,
            40,
            2
        );
    }

    async waitForJobsToRunInNodePool(
        submittedJobsNames: string[],
        namespace: string,
        expectedNodePoolName: string
    ): Promise<void> {
        const promisesToAwait = [];
        for (const jobName of submittedJobsNames) {
            promisesToAwait.push(waitUtils.waitForPods(jobName, namespace, 1, 60, ['Running']));
        }
        await Promise.all(promisesToAwait);

        for (const jobName of submittedJobsNames) {
            await this.waitForJobAndRunningPodsLabeledCorrectly(jobName, namespace, expectedNodePoolName);
        }
    }

    async getNodepoolsCountByClusterId(
        clusterId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<number> {
        clusterApi.nodepoolsApi.accessToken = accessToken;
        const getResponse = await clusterApi.nodepoolsApi.countNodepools(['clusterId==' + clusterId]);
        expect(getResponse.response.statusCode.valueOf()).equal(StatusCodes.OK);
        return getResponse.body.count;
    }

    async getSchedulingConditions(jobName: string, namespace: string, instances = 1): Promise<object> {
        const podGroups = await waitUtils.waitForPodGroups(jobName, namespace, instances);
        const result = {};
        for (const podGroup of podGroups) {
            if (podGroup['status']) {
                result[podGroup.metadata.name] = podGroup['status']['schedulingConditions'];
            } else {
                result[podGroup.metadata.name] = new Array(0);
            }
        }
        return result;
    }

    async waitForUnschedulableOnNodePool(
        jobName: string,
        namespace: string,
        nodePool: string,
        numOfNodePoolsSubmittedForJob: number,
        instances = 1,
        timeout = 30
    ): Promise<void> {
        await this.waitForUnschedulableOnNodePoolsInCorrectOrder(
            jobName,
            namespace,
            [nodePool],
            numOfNodePoolsSubmittedForJob,
            instances,
            timeout
        );
    }

    isPodGroupUnschedulableOnNodePoolsInCorrectOrder(
        podGroupSchedulingConditions,
        expectedUnschedulableOnNodePools: string[],
        numOfNodePoolsSubmittedForJob: number
    ): { isCorrectOrder: boolean; foundUnschedulableNodePools: Array<string> } {
        const foundUnschedulableNodePools = new Array(0);
        if (podGroupSchedulingConditions && podGroupSchedulingConditions.length > 0) {
            for (let i = 0; i < podGroupSchedulingConditions.length; i++) {
                const schedulingCondition = podGroupSchedulingConditions[i];
                if (schedulingCondition['type'] === this.UnschedulableOnNodePool) {
                    foundUnschedulableNodePools.push(schedulingCondition['nodePool']);
                }

                if (foundUnschedulableNodePools.length === numOfNodePoolsSubmittedForJob) {
                    break;
                }
            }
        }

        if (!expectedUnschedulableOnNodePools || expectedUnschedulableOnNodePools.length === 0) {
            return {
                isCorrectOrder: foundUnschedulableNodePools.length === 0,
                foundUnschedulableNodePools: foundUnschedulableNodePools
            };
        }

        if (foundUnschedulableNodePools.length < expectedUnschedulableOnNodePools.length) {
            return { isCorrectOrder: false, foundUnschedulableNodePools: foundUnschedulableNodePools };
        }

        // validate the UnschedulableOnNodePool condition was found in the same order,
        // but not necessarily starting with the same one
        let firstExpectedNpIndex = -1;
        for (let i = 0; i < expectedUnschedulableOnNodePools.length; i++) {
            if (foundUnschedulableNodePools[i] === expectedUnschedulableOnNodePools[0]) {
                firstExpectedNpIndex = i;
                break;
            }
        }

        for (let i = 0; i < expectedUnschedulableOnNodePools.length; i++) {
            const nodePool = expectedUnschedulableOnNodePools[i];
            if (
                foundUnschedulableNodePools[(firstExpectedNpIndex + i) % numOfNodePoolsSubmittedForJob] !==
                nodePool
            ) {
                return { isCorrectOrder: false, foundUnschedulableNodePools: foundUnschedulableNodePools };
            }
        }

        return { isCorrectOrder: true, foundUnschedulableNodePools: foundUnschedulableNodePools };
    }

    async waitForUnschedulableOnNodePoolsInCorrectOrder(
        jobName: string,
        namespace: string,
        unschedulableOnNodePools: string[],
        numOfNodePoolsSubmittedForJob: number,
        instances = 1,
        timeout = 30
    ): Promise<void> {
        logger.debug(
            `Waiting for job ${namespace}/${jobName} to be UnschedulableOnNodePool on node pools: ${JSON.stringify(
                unschedulableOnNodePools
            )}`
        );
        let lastFoundUnschedulableNodePools = new Array(0);
        const isUnschedulableOnNodePool = async (): Promise<boolean> => {
            const allSchedulingConditions = await this.getSchedulingConditions(jobName, namespace, instances);
            for (const podGroupSchedulingConditions of Object.values(allSchedulingConditions)) {
                const { isCorrectOrder, foundUnschedulableNodePools } =
                    this.isPodGroupUnschedulableOnNodePoolsInCorrectOrder(
                        podGroupSchedulingConditions,
                        unschedulableOnNodePools,
                        numOfNodePoolsSubmittedForJob
                    );
                lastFoundUnschedulableNodePools = foundUnschedulableNodePools;
                if (!isCorrectOrder) {
                    return false;
                }
            }
            return true;
        };

        try {
            await waitUtils.waitForCondition(
                isUnschedulableOnNodePool,
                `waitForUnschedulableOnNodePoolsInCorrectOrder jobName: ${namespace}/${jobName}, nodePools: ${JSON.stringify(
                    unschedulableOnNodePools
                )}`,
                true,
                timeout
            );
        } catch (e) {
            logger.debug(
                `UnschedulableOnNodePool conditions before reached timeout in waitForUnschedulableOnNodePoolsInCorrectOrder for job "${namespace}/${jobName}": ${JSON.stringify(
                    lastFoundUnschedulableNodePools
                )}`
            );
            e.message = `${e.message} - unschedulable node pools order was: ${JSON.stringify(
                lastFoundUnschedulableNodePools
            )}`;
            throw e;
        }
    }

    getMarkUnschedulableOfPodGroup(podGroup: KubernetesObject): boolean {
        const result = podGroup['spec']['markUnschedulable'];
        if (!result) {
            return false;
        }
        return true;
    }

    getSchedulingBackoffOfPodGroup(podGroup: KubernetesObject): number {
        return Number(podGroup['spec']['schedulingBackoff']);
    }

    prerequisitesForTestsWithGpus(
        numOfGpusRequiredForThisTest: number,
        numOfGpus: number,
        numOfNodesWithGpusRequiredForThisTest: number,
        numOfNodesWithGpus: number,
        shouldLog = true
    ): boolean {
        if (numOfGpus < numOfGpusRequiredForThisTest) {
            if (shouldLog) {
                logger.info(
                    `not enough GPUs to perform test; numOfGpus: ${numOfGpus}, numOfGpusRequired: ${numOfGpusRequiredForThisTest}`
                );
            }
            return false;
        }

        if (numOfNodesWithGpus < numOfNodesWithGpusRequiredForThisTest) {
            if (shouldLog) {
                logger.info(
                    `not enough nodes with GPUs to perform test; numOfNodesWithGpus: ${numOfNodesWithGpus}, numOfNodesWithGpusRequired: ${numOfNodesWithGpusRequiredForThisTest}`
                );
            }
            return false;
        }

        return true;
    }

    buildNodeAffinityForNodePools(
        nodePools: { name: string; labelKey: string; labelValue: string }[]
    ): V1Affinity {
        const nodeSelectorTerms = new Array(0);
        for (const nodePool of nodePools) {
            const matchExpression = this.createMatchExpressionForNodePool(nodePool);
            nodeSelectorTerms.push({ matchExpressions: [matchExpression] });
        }

        return {
            nodeAffinity: {
                requiredDuringSchedulingIgnoredDuringExecution: { nodeSelectorTerms: nodeSelectorTerms }
            }
        };
    }

    createMatchExpressionForNodePool(nodePool: {
        name: string;
        labelKey: string;
        labelValue: string;
    }): V1NodeSelectorRequirement {
        if (nodePool.name === this.DefaultNodePoolName) {
            return {
                key: this.NodePoolNameLabel,
                operator: 'DoesNotExist'
            };
        }

        return {
            key: nodePool.labelKey,
            operator: 'In',
            values: [nodePool.labelValue]
        };
    }

    async setupNodeGroupInfrastructure(options: CreateNodeGroupOptions): Promise<ProjectFromOrgUnit> {
        const {
            nodeLabelKey,
            nodeLabelVal1,
            nodeLabelVal2,
            nodeGroupSize = 1,
            createTopology,
            attachTopology,
            topoLevels,
            createTopoRequest,
            topologyIds,
            topoRegionKey,
            topoCityKey,
            region,
            cityName,
            nodePoolCreateRequest,
            nodePoolCreateRequest2,
            nodePools,
            npName1,
            npName2,
            npProjectName,
            gpuDeserved = 3,
            topologyReadyTimeout = 120
        } = options;

        let { workerNodeNames, nodeGroup1, nodeGroup2, midIndex } = options;
        workerNodeNames = await nodeUtils.getWorkerNodeNames();
        const onlyWorkerNodes = workerNodeNames.filter((node: string) => !node.includes('system'));
        logger.info(`Worker nodes in cluster: ${onlyWorkerNodes.length}`);

        if (nodeGroupSize === 2) {
            midIndex = Math.floor(onlyWorkerNodes.length / 2) + 1;
            nodeGroup1 = onlyWorkerNodes.slice(0, midIndex);
            nodeGroup2 = onlyWorkerNodes.slice(midIndex);

            await Promise.all(
                nodeGroup2.map((node: string) => kubeCli.labelNode(node, nodeLabelKey, nodeLabelVal2!))
            );
            logger.info(`Labeled ${nodeGroup2.length} nodes with ${nodeLabelKey}=${nodeLabelVal2}`);
        } else {
            nodeGroup1 = onlyWorkerNodes.length > 0 ? onlyWorkerNodes : workerNodeNames;
            await Promise.all(
                nodeGroup1.map((node: string) => kubeCli.labelNode(node, nodeLabelKey, nodeLabelVal1))
            );
            logger.info(`Labeled ${nodeGroup1.length} nodes with ${nodeLabelKey}=${nodeLabelVal1}`);
        }

        if (createTopology) {
            createTopoRequest.levels = topoLevels;

            if (topoLevels.length > 0) {
                const nodesToLabel = nodeGroupSize === 2 ? nodeGroup1 : workerNodeNames;
                const labels: { key: string; value: string }[] = [];

                if (topoLevels.includes(topoRegionKey)) {
                    labels.push({ key: topoRegionKey, value: region });
                }
                if (topoLevels.includes(topoCityKey)) {
                    labels.push({ key: topoCityKey, value: cityName });
                }

                await Promise.all(
                    nodesToLabel.flatMap((node: string) =>
                        labels.map((label) => kubeCli.labelNode(node, label.key, label.value))
                    )
                );
            }

            const topoRsp = await networkTopologiesUtils.createNetworkTopologyViaApi(createTopoRequest);
            topologyIds.push(topoRsp.body.id);
            logger.info(`Created topology with level: ${topoLevels.join(', ')}`);
            logger.info(`Topology Body: ${JSON.stringify(topoRsp.body)}`);

            if (attachTopology) {
                nodePoolCreateRequest.networkTopologyId = topoRsp.body.id;
            }
        }

        const npRsp1 = await this.createClusterNodePoolWithRequest(nodePoolCreateRequest);

        if (nodeGroupSize === 2 && nodePoolCreateRequest2) {
            const npRsp2 = await this.createClusterNodePoolWithRequest(nodePoolCreateRequest2);
            await Promise.all(
                [npRsp2.body.id].map((id: string) =>
                    this.waitForClusterNodePoolPhase(id, [NodepoolPhase.Ready, NodepoolPhase.Empty], 30)
                )
            );
            nodePools.push({ name: npName2!, id: npRsp2.body.id });
        }

        await Promise.all(
            [npRsp1.body.id].map((id: string) =>
                this.waitForClusterNodePoolPhase(id, [NodepoolPhase.Ready, NodepoolPhase.Empty], 30)
            )
        );
        nodePools.push({ name: npName1, id: npRsp1.body.id });

        const res = [npRsp1.body].map((np: any) => ({
            nodePool: { name: np.name, id: np.id.toString() },
            gpu: { deserved: gpuDeserved, limit: -1 }
        }));

        const npProject = await orgUnitUtils.createProjectAndValidate({
            clusterId: flowRootState.cluster.uuid,
            resources: res,
            name: npProjectName,
            defaultNodePools: [npName1]
        });

        if (createTopology) {
            let lastTopologyState: any = null;

            try {
                await waitUtils.waitForCondition(
                    async () => {
                        try {
                            const res = await networkTopologiesUtils.getNetworkTopologyByIdViaApi(
                                topologyIds[0]
                            );
                            lastTopologyState = res.body;
                            logger.info(
                                `Topology ${topologyIds[0]} - Phase: ${res.body?.phase}, Status: ${JSON.stringify(res.body)}`
                            );
                            return res.body?.phase === 'Ready';
                        } catch (error) {
                            logger.warn(`Failed to get topology ${topologyIds[0]}: ${error.message}`);
                            lastTopologyState = { error: error.message };
                            return false;
                        }
                    },
                    `Wait for topology ${topologyIds[0]} to be Ready`,
                    true,
                    topologyReadyTimeout
                );
                logger.info(`Topology ${topologyIds[0]} is now Ready`);
            } catch (error) {
                const lastPhase = lastTopologyState?.phase || 'Unknown';
                const errorDetails = lastTopologyState?.error || JSON.stringify(lastTopologyState, null, 2);

                logger.error(
                    `Topology ${topologyIds[0]} failed to become Ready after ${topologyReadyTimeout}s. ` +
                        `Last known phase: ${lastPhase}. ` +
                        `Last state: ${errorDetails}`
                );

                error.message = `${error.message} - Topology ID: ${topologyIds[0]}, Last phase: ${lastPhase}, Last state: ${errorDetails}`;
                throw error;
            }
        }

        options.workerNodeNames = workerNodeNames;
        options.nodeGroup1 = nodeGroup1;
        if (nodeGroup2) options.nodeGroup2 = nodeGroup2;
        if (midIndex !== undefined) options.midIndex = midIndex;

        return npProject;
    }

    ///////////////// NodePool API Wrappers with Full Response /////////////////

    /**
     * Gets nodepools minimal data and returns the full response including status code
     */
    async getNodepoolsMinimalApi(
        filterBy?: Array<string>,
        sortBy?: any,
        sortOrder?: 'asc' | 'desc',
        offset?: number,
        limit?: number,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: GetNodepoolsMinimal200Response }> {
        clusterApi.nodepoolsApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await clusterApi.nodepoolsApi.getNodepoolsMinimal(filterBy, sortBy, sortOrder, offset, limit);
    }

    /**
     * Gets nodepool minimal data by ID and returns the full response including status code
     */
    async getNodepoolMinimalApi(
        nodepoolId: string,
        accessToken?: string
    ): Promise<{ response: http.IncomingMessage; body: NodepoolMinimal }> {
        clusterApi.nodepoolsApi.accessToken =
            accessToken ?? flowRootState.users.system_administrator.accessToken;
        return await clusterApi.nodepoolsApi.getNodepoolMinimal(nodepoolId);
    }

    /**
     * Label all nodes in a nodepool with topology level labels
     * @param nodePoolId - The ID of the nodepool
     * @param topologyLevels - Array of topology level keys (e.g., ['cloud.provider.com/topology-rack', 'kubernetes.io/hostname'])
     * @param accessToken - Optional access token
     * @returns Promise<Array<{nodeName: string; labels: Array<{key: string; value: string}>}>> - Array of nodes with their applied labels
     */
    async labelNodePoolNodesWithTopologyLevels(
        nodePoolId: string,
        topologyLevels: string[],
        accessToken: string = flowRootState.users.system_administrator.accessToken
    ): Promise<Array<{ nodeName: string; labels: Array<{ key: string; value: string }> }>> {
        if (!topologyLevels || topologyLevels.length === 0) {
            logger.info('No topology levels provided, skipping node labeling');
            return [];
        }

        // Get the nodepool to retrieve its labelKey and labelValue
        const nodePoolResponse = await this.getNodePoolById(nodePoolId, accessToken);
        const nodePool = nodePoolResponse.body;

        if (!nodePool) {
            throw new Error(`NodePool with ID ${nodePoolId} not found`);
        }

        logger.info(
            `Labeling nodes in nodepool ${nodePool.name} (ID: ${nodePoolId}) with topology levels: ${topologyLevels.join(', ')}`
        );

        // Get node names from nodepool status
        const nodeNames = nodePool.status?.nodes || [];

        if (nodeNames.length === 0) {
            logger.warn(`No nodes found in nodepool ${nodePool.name}`);
            return [];
        }

        logger.debug(`Found ${nodeNames.length} nodes in nodepool ${nodePool.name}: ${nodeNames.join(', ')}`);

        const labeledNodesInfo: Array<{ nodeName: string; labels: Array<{ key: string; value: string }> }> =
            [];

        // Label each node with each topology level
        for (const nodeName of nodeNames) {
            const nodeLabels: Array<{ key: string; value: string }> = [];

            for (const topologyLevel of topologyLevels) {
                // Generate a random value for this node and topology level
                // Use a combination of node name hash and random string to ensure uniqueness
                const randomValue = objectGenerator.generateNameByPrefix(`topo-${nodeName.substring(0, 8)}`);

                try {
                    await kubeCli.labelNode(nodeName, topologyLevel, randomValue);
                    logger.info(`Labeled node ${nodeName} with ${topologyLevel}=${randomValue}`);
                    nodeLabels.push({ key: topologyLevel, value: randomValue });
                } catch (error) {
                    logger.error(`Failed to label node ${nodeName} with ${topologyLevel}: ${error.message}`);
                    throw error;
                }
            }

            labeledNodesInfo.push({ nodeName, labels: nodeLabels });
        }

        logger.info(
            `Successfully labeled ${labeledNodesInfo.length} nodes with ${topologyLevels.length} topology levels each`
        );
        return labeledNodesInfo;
    }

    // Helper function to create nodepools and label GPU nodes
    async createNodepoolsAndValidate(addNodepools: NodepoolConfig[]): Promise<{
        nodepoolsNames: string[];
        nodeLabels: {
            nodeName: string;
            labelKeys: string[];
        }[];
    }> {
        const allNodePools = await this.getAllBackendNodePoolsAndValidate();
        const nodePoolMap = new Map(allNodePools.map((np) => [np.name, np]));
        const result = { nodepoolsNames: [], nodeLabels: [] };
        for (const nodePoolConfig of addNodepools) {
            const nodePoolName = nodePoolConfig.name;

            if (!nodePoolMap.has(nodePoolName)) {
                if (nodePoolName === DEFAULT_NODE_POOL_NAME) {
                    logger.info(`Node pool ${DEFAULT_NODE_POOL_NAME} always exists, skipping creation`);
                    continue;
                }

                logger.info(`Node pool ${nodePoolName} does not exist, creating it...`);

                const labelKey = nodePoolConfig.labelKey;
                const labelValue = nodePoolConfig.labelValue;

                const createResponse = await this.createBackendNodePool(nodePoolName, labelKey, labelValue);

                expect(createResponse.response.statusCode).to.equal(201);
                expect(createResponse.body).to.exist;

                const createdNodePool = createResponse.body;
                nodePoolMap.set(nodePoolName, createdNodePool);

                await this.waitForNodePoolCrdStatus(
                    nodePoolName,
                    [NodePool.StatusEnum.Ready, NodePool.StatusEnum.Empty],
                    60
                );

                result.nodepoolsNames.push(nodePoolName);

                logger.info(
                    `Created node pool ${nodePoolName} with labelKey=${labelKey}, labelValue=${labelValue}, ID: ${createdNodePool.id}`
                );
            }

            if (nodePoolConfig.setGpuNode && nodePoolName !== DEFAULT_NODE_POOL_NAME) {
                const labelKey = nodePoolConfig.labelKey;
                const labelValue = nodePoolConfig.labelValue;

                logger.info(
                    `Labeling GPU node for nodepool ${nodePoolName} with labelKey=${labelKey}, labelValue=${labelValue}`
                );

                const nodeName = await this.labelAGPUNode(labelKey, labelValue, []);

                expect(nodeName).not.to.equal('');
                result.nodeLabels.push({
                    nodeName: nodeName,
                    labelKeys: [labelKey]
                });

                logger.info(
                    `Labeled GPU node ${nodeName} with labelKey=${labelKey}, labelValue=${labelValue}`
                );
            }
        }
        return result;
    }

    async prepareNodepoolInformation(nodepoolNames: string[]): Promise<NodepoolConfig[]> {
        const allNodePools = await this.getAllBackendNodePoolsAndValidate();
        const nodePoolMap = new Map(allNodePools.map((np) => [np.name, np]));

        const nodepoolsWithLabels = nodepoolNames.map((npName) => {
            if (npName === DEFAULT_NODE_POOL_NAME) {
                return { name: DEFAULT_NODE_POOL_NAME };
            } else {
                const nodePool = nodePoolMap.get(npName);
                if (nodePool) {
                    return {
                        name: npName,
                        labelKey: nodePool.labelKey,
                        labelValue: nodePool.labelValue
                    };
                } else {
                    throw new Error(`Nodepool ${npName} not found`);
                }
            }
        });

        logger.info(`Prepared nodepool information: ${JSON.stringify(nodepoolsWithLabels)}`);
        return nodepoolsWithLabels;
    }
}
