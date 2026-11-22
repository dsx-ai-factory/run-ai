/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState } from '../config/config';
import { clusterApi } from '../../api';
import { logger, nodeUtils, waitUtils, operations } from '../index';
import { expect } from 'chai';
import {
    MetricsResponse,
    Node,
    NodeMetricType,
    Nodes,
    NodeStatusConditionDetails,
    NodeTaint
} from '../../api/apis/clusterApi';
import http from 'http';
import { kubeCli } from '../../cli';
import { includes } from 'lodash-es';
import { V1Node } from '@kubernetes/client-node';

export const GPU_PRODUCT_LABEL = 'nvidia.com/gpu.product';

export const GPU_FRACTIONING_READY_CONDITION = 'gpu-fractioning.nvidia.com/Ready';

// cluster-sync only evaluates nodes carrying this label; NFD sets it from a real
// NVIDIA PCI device, so fake-GPU nodes need it added before they are considered.
export const NVIDIA_GPU_LABEL = 'feature.node.kubernetes.io/pci-10de.present';

// The reason fake-gpu-operator stamps on conditions it owns. A test must never
// reuse it: conditions carrying it are reconciled back to True, anything else is
// left alone, which is what makes staging a failing condition possible at all.
export const SIMULATED_GPU_FRACTIONING_REASON = 'FakeGpuSimulation';

export interface StagedNodeCondition {
    status: 'True' | 'False' | 'Unknown';
    reason?: string;
    message?: string;
}

export class NodeUtils {
    async getAllNodes(
        accessToken = flowRootState.users.system_administrator.accessToken,
        nodeName?: string
    ): Promise<{ response: http.IncomingMessage; body: Nodes }> {
        clusterApi.nodesApi.accessToken = accessToken;
        const filters = [`clusterUuid==${flowRootState.cluster.uuid}`];
        if (nodeName) {
            filters.push(`name==${nodeName}`);
        }
        return await clusterApi.nodesApi.getTenantNodes(filters);
    }

    async unlabelNodes(nodeName: string, removeLabels: string[]): Promise<void> {
        const labels = await kubeCli.getNodeLabels(nodeName);
        const labelKeys = Object.keys(labels);

        for (const removeLabel of removeLabels) {
            if (includes(labelKeys, removeLabel)) {
                await kubeCli.unLabelNode(nodeName, removeLabel);
            }
        }
    }

    async expectNoNodeCondition(nodeName: string, conditionType: string): Promise<void> {
        const condition = await nodeUtils.findNodeCondition(nodeName, conditionType);
        expect(condition, `node ${nodeName} should not report condition ${conditionType}`).to.be
            .undefined;
    }

    async findNodeCondition(
        nodeName: string,
        conditionType: string
    ): Promise<NodeStatusConditionDetails | undefined> {
        const allNodes = await nodeUtils.getAllNodes(undefined, nodeName);
        const node = allNodes.body.nodes.find((candidate) => candidate.name === nodeName);
        return node?.conditions?.find((condition) => condition.type === conditionType);
    }

    // cluster-sync reports asynchronously, so every staged change needs a wait before the
    // assertion. Pass expectedReason whenever the previous case also produced this condition:
    // waiting on existence alone returns immediately and asserts against the stale reason.
    async waitForNodeCondition(
        nodeName: string,
        conditionType: string,
        shouldExist = true,
        expectedReason?: string
    ): Promise<void> {
        const target = shouldExist
            ? `appear${expectedReason ? ` with reason ${expectedReason}` : ''}`
            : 'disappear';
        await waitUtils.waitForCondition(
            async () => {
                try {
                    const condition = await nodeUtils.findNodeCondition(nodeName, conditionType);
                    if ((condition !== undefined) !== shouldExist) {
                        return false;
                    }
                    return expectedReason === undefined || condition?.reason === expectedReason;
                } catch (error) {
                    logger.debug(`Reading condition ${conditionType} on ${nodeName} failed: ${error}`);
                    return false;
                }
            },
            `wait for node ${nodeName} condition ${conditionType} to ${target}`,
            true,
            60
        );
    }

    // Writes a node status condition directly. A fake-GPU environment never reaches a
    // missing-prerequisite state on its own, so the failing cases have to be staged.
    // Indices are guarded with a JSON Patch test op: the client negotiates json-patch,
    // so a stale index would otherwise rewrite an unrelated condition.
    async stageNodeCondition(
        nodeName: string,
        conditionType: string,
        staged: StagedNodeCondition
    ): Promise<void> {
        const now = new Date();
        const value = {
            type: conditionType,
            status: staged.status,
            reason: staged.reason ?? '',
            message: staged.message ?? '',
            lastHeartbeatTime: now,
            lastTransitionTime: now
        };

        const index = await nodeUtils.nodeConditionIndex(nodeName, conditionType);
        const patch =
            index < 0
                ? [{ op: 'add', path: '/status/conditions/-', value }]
                : [
                      { op: 'test', path: `/status/conditions/${index}/type`, value: conditionType },
                      { op: 'replace', path: `/status/conditions/${index}`, value }
                  ];
        await kubeCli.patchNodeStatus(nodeName, patch);
    }

    async clearNodeCondition(nodeName: string, conditionType: string): Promise<void> {
        const index = await nodeUtils.nodeConditionIndex(nodeName, conditionType);
        if (index < 0) {
            return;
        }
        await kubeCli.patchNodeStatus(nodeName, [
            { op: 'test', path: `/status/conditions/${index}/type`, value: conditionType },
            { op: 'remove', path: `/status/conditions/${index}` }
        ]);
    }

    async stageGpuFractioningCondition(
        nodeName: string,
        staged: StagedNodeCondition
    ): Promise<void> {
        if (staged.reason === SIMULATED_GPU_FRACTIONING_REASON) {
            throw new Error(
                `refusing to stage ${GPU_FRACTIONING_READY_CONDITION} with reason ` +
                    `${SIMULATED_GPU_FRACTIONING_REASON}: fake-gpu-operator owns that reason and will revert it`
            );
        }
        await nodeUtils.stageNodeCondition(nodeName, GPU_FRACTIONING_READY_CONDITION, staged);
    }

    async readGpuFractioningCondition(nodeName: string): Promise<StagedNodeCondition | undefined> {
        const node = await kubeCli.getNode(nodeName);
        const condition = (node.status?.conditions ?? []).find(
            (c) => c.type === GPU_FRACTIONING_READY_CONDITION
        );
        return condition
            ? {
                  status: condition.status as StagedNodeCondition['status'],
                  reason: condition.reason,
                  message: condition.message
              }
            : undefined;
    }

    async clearGpuFractioningCondition(nodeName: string): Promise<void> {
        await nodeUtils.clearNodeCondition(nodeName, GPU_FRACTIONING_READY_CONDITION);
    }

    private async nodeConditionIndex(nodeName: string, conditionType: string): Promise<number> {
        const node = await kubeCli.getNode(nodeName);
        return (node.status?.conditions ?? []).findIndex(
            (condition) => condition.type === conditionType
        );
    }

    async getNodeStatusAndValidate(
        nodeName: string,
        desiredStatus: Node.StatusEnum,
        desiredCondition?: string,
        desiredConditionsMessage?: string,
        desiredConditionReason?: string
    ): Promise<void> {
        const allNodes = await nodeUtils.getAllNodes(undefined, nodeName);
        const workerNode = allNodes.body.nodes.find((node) => node.name === nodeName);
        expect(workerNode.status).to.equal(desiredStatus);
        if (desiredCondition) {
            // Matched by type rather than by position: a node can report several
            // conditions at once, and the one under test is not necessarily first.
            const condition = workerNode.conditions?.find((c) => c.type === desiredCondition);
            expect(condition, `node ${nodeName} should report condition ${desiredCondition}`).to.not
                .be.undefined;
            if (desiredConditionsMessage) expect(condition.message).to.equal(desiredConditionsMessage);
            if (desiredConditionReason) expect(condition.reason).to.equal(desiredConditionReason);
        }
    }

    async waitForNodeStatus(nodeName: string, desiredStatus: Node.StatusEnum): Promise<void> {
        const doesNodeStatusCorrect = async (): Promise<boolean> => {
            try {
                const allNodes = await nodeUtils.getAllNodes();
                const workerNode = allNodes.body.nodes.find((node) => node.name === nodeName);
                return workerNode.status === desiredStatus;
            } catch (e) {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            doesNodeStatusCorrect,
            `wait for node ${nodeName} to be in status ${desiredStatus}`,
            true,
            20
        );
    }

    async getIdleGpuNode(): Promise<V1Node> {
        const nodes = await kubeCli.getNodes('nvidia.com/gpu.present=true');
        const allPods = await kubeCli.listAllPods();

        const getPodGpuLimit = (pod: (typeof allPods)[number]): number => {
            return (pod.spec?.containers ?? []).reduce(
                (sum, c) => sum + parseInt((c.resources?.limits?.['nvidia.com/gpu'] as string) ?? '0'),
                0
            );
        };

        const activeGpuPods = allPods.filter((pod) => {
            const phase = pod.status?.phase;
            return (
                pod.spec?.nodeName && phase !== 'Succeeded' && phase !== 'Failed' && getPodGpuLimit(pod) > 0
            );
        });
        const busyGpuNodeNames = new Set(activeGpuPods.map((pod) => pod.spec?.nodeName));

        const idleNode = nodes.items.find((node) => !busyGpuNodeNames.has(node.metadata?.name));
        if (!idleNode) {
            const usage = nodes.items
                .map((node) => {
                    const podsOnNode = activeGpuPods.filter(
                        (pod) => pod.spec?.nodeName === node.metadata?.name
                    );
                    const reservedGpu = podsOnNode.reduce((sum, pod) => sum + getPodGpuLimit(pod), 0);
                    const occupants = podsOnNode.map(
                        (pod) =>
                            `${pod.metadata?.namespace}/${pod.metadata?.name}(${getPodGpuLimit(pod)} GPU)`
                    );
                    return `  ${node.metadata?.name}: ${reservedGpu} GPU reserved by [${occupants.join(', ')}]`;
                })
                .join('\n');
            throw new Error(
                'No idle GPU node available - a node with all GPUs free is required, but every GPU ' +
                    'node has reserved GPUs (likely leftover workloads from another test on this shared ' +
                    `cluster). Current GPU usage:\n${usage}`
            );
        }
        return idleNode;
    }

    async getNodesWithGpusCount(): Promise<number> {
        const gpuCountLabel = 'nvidia.com/gpu.count';
        const nodesList = await kubeCli.getNodes(gpuCountLabel);
        expect(nodesList.items.length).greaterThan(0);
        return nodesList.items.length;
    }

    async getAllGpusCount(): Promise<number> {
        // supported only for 2.15 and above
        clusterApi.nodesApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const nodes = await clusterApi.nodesApi.getTenantNodes([
            `clusterUuid==${flowRootState.cluster.uuid}`
        ]);
        let gpusCount = 0;
        for (const node of nodes.body.nodes) {
            if (node.gpuInfo) {
                gpusCount += node.gpuInfo.gpuCount;
            }
        }
        return gpusCount;
    }

    async getMaxGpuCountPerNode(): Promise<number> {
        // supported only for 2.15 and above
        clusterApi.nodesApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const nodes = await clusterApi.nodesApi.getTenantNodes([
            `clusterUuid==${flowRootState.cluster.uuid}`
        ]);

        const maxGpuCount = nodes.body.nodes
            .filter((node) => Node.StatusEnum.Ready && node.gpuInfo)
            .reduce((max, node) => {
                if (node.gpuInfo!.gpuCount > max) {
                    return node.gpuInfo!.gpuCount;
                }
                return max;
            }, 0);

        return maxGpuCount;
    }

    async getNodesCount(): Promise<number> {
        const nodesList = await kubeCli.getNodes();
        return nodesList.items.length;
    }

    async getAllReadyGpusCount(): Promise<number> {
        // supported only for 2.15 and above
        clusterApi.nodesApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const nodes = await clusterApi.nodesApi.getTenantNodes([
            `clusterUuid==${flowRootState.cluster.uuid}`
        ]);
        let readyGpusCount = 0;
        for (const node of nodes.body.nodes) {
            if (node.status === Node.StatusEnum.Ready && node.gpuInfo) {
                readyGpusCount += node.gpuInfo.gpuCount;
            }
        }
        return readyGpusCount;
    }

    async addManagedNodesConfig(labels: { [key: string]: string }): Promise<void> {
        const nodeSelectorTerms = Object.keys(labels).map((key) => ({
            matchExpressions: [
                {
                    key: key,
                    operator: 'In',
                    values: [labels[key]]
                }
            ]
        }));

        const patch = {
            spec: {
                global: {
                    managedNodes: {
                        inclusionCriteria: {
                            nodeSelectorTerms: nodeSelectorTerms
                        }
                    }
                }
            }
        };

        try {
            await operations.updateRunaiConfig(patch);
            logger.info(`RunaiConfig patched with managed nodes configuration: ${JSON.stringify(labels)}`);
        } catch (err) {
            logger.error(`Error patching RunaiConfig: ${err}`);
        }
    }

    async removeManagedNodesConfig(): Promise<void> {
        const patch = {
            spec: {
                global: {
                    managedNodes: null
                }
            }
        };

        await operations
            .updateRunaiConfig(patch)
            .then(() => {
                logger.info('RunaiConfig patched to remove managed nodes configuration.');
            })
            .catch((err) => {
                logger.error(`Error patching RunaiConfig: ${err}`);
            });
    }

    static async getGpuType(): Promise<string> {
        const nodesList = await kubeCli.getNodes();
        for (const node of nodesList.items) {
            const labels = await kubeCli.getNodeLabels(node.metadata.name);
            if (labels && labels[GPU_PRODUCT_LABEL]) {
                return labels[GPU_PRODUCT_LABEL];
            }
        }
        throw new Error(`No GPU node with ${GPU_PRODUCT_LABEL} label found`);
    }

    async getWorkerNodeNames(): Promise<string[]> {
        const nodes = await this.getAllNodes();
        const unschedulableEffects = [NodeTaint.EffectEnum.NoSchedule, NodeTaint.EffectEnum.NoExecute];
        return nodes.body.nodes
            .filter((node) => {
                if (node.status !== Node.StatusEnum.Ready) return false;
                return !(node.taints && node.taints.some((t) => unschedulableEffects.includes(t.effect)));
            })
            .map((node) => node.name);
    }

    async getSchedulableNodes(): Promise<Node[]> {
        const nodes = await this.getAllNodes();
        const unschedulableEffects = [NodeTaint.EffectEnum.NoSchedule, NodeTaint.EffectEnum.NoExecute];

        return nodes.body.nodes.filter((node) => {
            const isReady = node.status === Node.StatusEnum.Ready;
            const isSchedulable = !(
                node.taints && node.taints.some((t) => unschedulableEffects.includes(t.effect))
            );
            return isReady && isSchedulable;
        });
    }

    async getCpuWorkerNodeNames(): Promise<string[]> {
        const nodes = await this.getSchedulableNodes();
        return nodes
            .filter((node) => {
                return node.gpuInfo === undefined;
            })
            .map((node) => node.name);
    }

    async getGpuWorkerNodeNames(): Promise<string[]> {
        const nodes = await this.getSchedulableNodes();
        return nodes
            .filter((node) => {
                return node.gpuInfo !== undefined && node.gpuInfo.gpuCount > 0;
            })
            .map((node) => node.name);
    }

    async getNodeGpuDeviceCnt(nodeName: string): Promise<number> {
        const nodes = await this.getSchedulableNodes();
        return nodes.filter((node) => {
            return node.gpuInfo !== undefined && node.gpuInfo.gpuCount > 0 && node.name === nodeName;
        })[0].gpuInfo.gpuCount;
    }

    async getTotalGpuDeviceCnt(): Promise<number> {
        const nodes = await this.getSchedulableNodes();
        return nodes
            .filter((node) => {
                return node.gpuInfo !== undefined && node.gpuInfo.gpuCount > 0;
            })
            .reduce((sum, node) => sum + node.gpuInfo.gpuCount, 0);
    }

    async getGpuNodes(gpu_type: string): Promise<Node[]> {
        const nodes = await this.getAllNodes();
        return nodes.body.nodes.filter((node) => {
            // filter out not ready nodes
            if (node.status !== Node.StatusEnum.Ready) return false;
            return node.gpuInfo && node.gpuInfo.gpuType.toLowerCase() === gpu_type.toLowerCase();
        });
    }

    async getNodeMetrics(
        nodeId: string,
        metricType: NodeMetricType[],
        start: Date,
        end: Date,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<{ response: http.IncomingMessage; body: MetricsResponse }> {
        clusterApi.nodesApi.accessToken = accessToken;
        return await clusterApi.nodesApi.getNodeMetrics(
            nodeId,
            [metricType.join(',') as unknown as NodeMetricType],
            start,
            end
        );
    }
}
