/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    InteractiveWorkload,
    logger,
    objectGenerator,
    operations,
    Utils,
    waitUtils,
    workloadUtils,
    NodeResources,
    BaseWorkload,
    projectResourcesUtils
} from '..';
import { Project } from '../../api/apis/projectsDepartmentsApi';
import { V1Pod } from '@kubernetes/client-node';

export class SchedulingUtils {
    private static workloads: Array<BaseWorkload>;

    private static getValidClusterResources(
        desiredNodesCount: number,
        desiredAvailableGPUsPerNode: number
    ): NodeResources[] {
        const compatibleNodes = flowRootState.nodesResources.filter(
            (node) => node.gpu >= desiredAvailableGPUsPerNode
        );
        if (compatibleNodes.length < desiredNodesCount) {
            throw new Error(
                `There are not enough nodes with ${desiredAvailableGPUsPerNode} GPUs in the cluster. There should be at least ${desiredNodesCount} nodes with ${desiredAvailableGPUsPerNode} GPUs`
            );
        }
        return compatibleNodes.slice(0, desiredNodesCount);
    }

    private static async fillNodeGpuResources(
        node: NodeResources,
        desiredAvailableGPUsPerNode: number
    ): Promise<Project> {
        const gpusToFill = node.gpu - desiredAvailableGPUsPerNode;
        if (gpusToFill > 0) {
            const jobName = objectGenerator.generateNameByPrefix('filler-job');
            logger.debug(`About to fill ${gpusToFill} GPUs on node ${node.name}`);
            const projectName = objectGenerator.generateNameByPrefix('filler-p');
            const namespace = Utils.namespaceFromProjectName(projectName);
            const project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                maxAllowedGpus: gpusToFill,
                deservedGpus: gpusToFill
            });

            const workload = new InteractiveWorkload(config.quickStartImage.interactive, namespace, jobName);
            workload.gpu = gpusToFill.toString();
            await workloadUtils.submit(workload);
            await waitUtils.waitForPods(jobName, namespace, 1);

            if (!this.workloads) {
                this.workloads = [];
            }
            this.workloads.push(workload);
            return project;
        }
    }

    private static async setNodesResourcesState(
        chosenNodes: NodeResources[],
        desiredAvailableGPUsPerNode: number
    ): Promise<Array<Project>> {
        const fillingProjects = new Array<Project>();
        // this function fills the nodes with
        let desiredAvailableGPUs;
        for (const node of flowRootState.nodesResources) {
            if (node.gpu) {
                logger.debug(`About to fill node ${node.name} with ${desiredAvailableGPUsPerNode} GPUs`);

                if (chosenNodes.includes(node)) {
                    desiredAvailableGPUs = desiredAvailableGPUsPerNode;
                } else {
                    desiredAvailableGPUs = 0;
                }
                logger.debug(`About to fill node ${node.name} to ${desiredAvailableGPUs} available GPUs`);
                const fillingProject = await this.fillNodeGpuResources(node, desiredAvailableGPUs);
                fillingProjects.push(fillingProject);
            }
        }
        return fillingProjects;
    }

    static isPodUnschedulable(pod: V1Pod): boolean {
        if (!pod.status.conditions) {
            return false;
        }

        const scheduledCondition = pod.status.conditions.filter((cond) => cond.type === 'PodScheduled');
        if (scheduledCondition.length !== 1) {
            return false;
        }
        return scheduledCondition[0].reason === 'Unschedulable';
    }

    // checks if the pod has PodScheduled condition, and it is either true, or false with a reason,
    // which means that the scheduler handled the pod and made at least an initial decision about it,
    // (either scheduled or not yet, but handled). if so, it returns the last transition time of this condition.
    // otherwise, returns undefined
    static getPodSchedulingDecisionTime(pod: V1Pod): Date {
        if (!pod.status.conditions) {
            return undefined;
        }

        const scheduledCondition = pod.status.conditions.filter((cond) => cond.type === 'PodScheduled');
        if (scheduledCondition.length === 0) {
            return undefined;
        }

        if (
            scheduledCondition[0].status === 'True' ||
            (scheduledCondition[0].reason !== undefined && scheduledCondition[0].reason !== '')
        ) {
            return scheduledCondition[0].lastTransitionTime;
        }

        return undefined;
    }

    static async setClusterResourcesState(
        desiredNodesCount: number,
        desiredAvailableGPUsPerNode: number,
        forceCleanup = true
    ): Promise<Array<Project>> {
        logger.info(
            `Setting cluster resources state for ${desiredNodesCount} nodes with ${desiredAvailableGPUsPerNode} GPUs`
        );
        const chosenNodes = this.getValidClusterResources(desiredNodesCount, desiredAvailableGPUsPerNode);
        if (forceCleanup) {
            logger.debug('About to cleanup cluster resources');
            await operations.forceCleanup();
        }

        return await this.setNodesResourcesState(chosenNodes, desiredAvailableGPUsPerNode);
    }

    static async clearResources(): Promise<void> {
        await operations.deleteWorkloads(this.workloads);
    }
}
