/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { crds, logger, nodePoolUtils, nodeUtils, orgUnitUtils } from '../index';
import { kubeCli } from '../../cli';
import { Node, NodepoolPhase } from '../../api/apis/clusterApi';

export interface NumaTestNodePool {
    nodePoolName: string;
    nodePoolId: string;
    labelKey: string;
    labelValue: string;
    nodeName: string;
}

export class NumaUtils {
    // Condition type the nodepool-controller sets on a node, and on its node pool, when NUMA-aware
    // scheduling is enabled but the node's NodeResourceTopology (NRT) data is missing or unusable.
    static missingNrtConditionType = 'MissingNrtHealthyPrerequisite';

    static missingNrtObjectMessage = 'NodeResourceTopology custom resource missing';

    // The KAI numa plugin rejects nodes that cannot fit the pod's GPUs in a single NUMA zone under
    // the single-numa-node policy - its fit error mentions NUMA alignment.
    static numaRejectionPattern = /NUMA-align/i;

    async getNodeNamesWithNrt(): Promise<string[]> {
        const nrts = await kubeCli.getCrds(crds.nodeResourceTopologyCrd);
        return nrts.map((nrt) => nrt.metadata.name);
    }

    // GPUs per NUMA zone, read from the node's NRT object - the same data the scheduler uses -
    // so scheduling tests don't depend on cluster-sync publishing it to the control plane.
    async getNrtZoneGpus(nodeName: string): Promise<number[]> {
        const nrt = await kubeCli.getCrd<any>(crds.nodeResourceTopologyCrd, nodeName);
        return (nrt.zones ?? [])
            .filter((zone) => zone.type === 'Node')
            .map((zone) => {
                const gpu = (zone.resources ?? []).find((resource) => resource.name === 'nvidia.com/gpu');
                return Number(gpu?.allocatable ?? 0);
            });
    }

    // A Ready node without GPUs and without an NRT object - adding it to a NUMA-aware node pool
    // should surface the missing-prerequisite status on both the node and the node pool.
    async findNodeWithoutNrt(): Promise<string | undefined> {
        const nodeNamesWithNrt = new Set(await this.getNodeNamesWithNrt());
        const nodes = (await nodeUtils.getAllNodes()).body.nodes;
        const candidate = nodes.find(
            (node) =>
                node.status === Node.StatusEnum.Ready &&
                !(node.gpuInfo?.gpuCount > 0) &&
                !nodeNamesWithNrt.has(node.name)
        );
        return candidate?.name;
    }

    async createNodePoolForGpuNode(): Promise<NumaTestNodePool> {
        const { key, value } = nodePoolUtils.generateKeyValueLabels();
        const created = await nodePoolUtils.createClusterNodePoolAndLabelNodeToAssignToNodePool(key, value);
        expect(created.nodeName, 'a GPU node should be labeled into the node pool').to.not.equal('');
        expect(created.nodePoolId, 'the node pool should be created').to.be.ok;
        await nodePoolUtils.waitForClusterNodePoolPhase(created.nodePoolId, [NodepoolPhase.Ready], 120);
        return { ...created, labelKey: key, labelValue: value };
    }

    async createNodePoolForNode(nodeName: string): Promise<NumaTestNodePool> {
        const { key, value } = nodePoolUtils.generateKeyValueLabels();
        await kubeCli.labelNode(nodeName, key, value);
        const created = await nodePoolUtils.createClusterNodePoolAndValidate(key, value);
        await nodePoolUtils.waitForClusterNodePoolPhase(created.nodePoolId, [NodepoolPhase.Ready], 120);
        return {
            nodePoolName: created.nodePoolName,
            nodePoolId: created.nodePoolId,
            labelKey: key,
            labelValue: value,
            nodeName
        };
    }

    async deleteNumaTestNodePool(nodePool: NumaTestNodePool | undefined): Promise<void> {
        if (!nodePool) {
            return;
        }
        await nodePoolUtils.cleanupDeleteClusterNodePool(nodePool);
        try {
            if (nodePool.nodeName) {
                await kubeCli.unLabelNode(nodePool.nodeName, nodePool.labelKey);
            }
        } catch (error) {
            logger.error(`Failed to unlabel node ${nodePool.nodeName}`, error);
        }
    }

    // Departments are enabled by default, so the new node pool needs department quota as well.
    async setDefaultDepartmentGpuQuota(nodePoolName: string, gpus: number): Promise<void> {
        const defaultDepartment = await orgUnitUtils.getDefaultDepartment();
        const nodePoolResources = defaultDepartment.resources.find(
            (resource) => resource.nodePool.name === nodePoolName
        );
        if (nodePoolResources) {
            nodePoolResources.gpu.deserved = gpus;
            // with over-quota off, the weight must be empty or equal to deserved
            nodePoolResources.gpu.overQuotaWeight = gpus;
            await orgUnitUtils.patchDepartmentResources(defaultDepartment.id, [nodePoolResources]);
        }
    }
}
