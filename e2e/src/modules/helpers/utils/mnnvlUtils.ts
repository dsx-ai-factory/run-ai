/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { bashCli } from '../../cli';
import { logger, nodePoolUtils, operations } from '../index';
import { isNvidiaDraDriverInstalled } from './IsInstalledUtils';
import {
    NodepoolGPUNetworkAccelerationDetection,
    NodepoolUpdateFields
} from '../../api/apis/clusterApi';

export class MnnvlUtils {
    static mnnvlLabel = 'nvidia.com/gpu.clique';

    isDraSupportedInCluster(): boolean {
        try {
            bashCli.exec('kubectl get resourceslices');
        } catch {
            // if the command fails, it means that DRA is not enabled in this cluster.
            return false;
        }
        return true;
    }

    async installMnnvlOperator(): Promise<void> {
        if (await isNvidiaDraDriverInstalled()) {
            logger.info('Nvidia dra driver is already installed.');
            return;
        }
        try {
            console.log(
                bashCli.execFile(`helm repo add nvidia https://helm.ngc.nvidia.com/nvidia
    helm repo update
    helm repo update && helm install nvidia-dra-driver-gpu nvidia/nvidia-dra-driver-gpu --version="25.8.1" --create-namespace --namespace nvidia-dra-driver-gpu --set resources.gpus.enabled=false --set nvidiaDriverRoot=/run/nvidia/driver
    `)
            );
        } catch (error) {
            throw new Error(`Failed to Nvidia dra driver: ${error.message}`);
        }
    }

    // Creates a plain GPU node pool (no GPU-network-acceleration detection). Detection is
    // turned on/off atomically by enableGpuNetworkAcceleration/disableGpuNetworkAcceleration
    // so callers can pair them in beforeAll/afterAll and never leave the cluster half-configured.
    async createMnnvlNodePool(
        nodePoolLabelKey: string,
        nodePoolLabelValue: string,
        nodePoolNamePrefix = 'np'
    ): Promise<{ nodePoolName: string; nodePoolId: string }> {
        await nodePoolUtils.labelAGPUNode(nodePoolLabelKey, nodePoolLabelValue);

        return await nodePoolUtils.createClusterNodePoolAndValidate(
            nodePoolLabelKey,
            nodePoolLabelValue,
            nodePoolNamePrefix
        );
    }

    // Atomically turns GPU network acceleration ON: flips the cluster-wide runaiconfig flag,
    // marks the given node pool as acceleration-capable (detection: Use) and waits until the
    // node pool reports it is detected. Pair with disableGpuNetworkAcceleration in afterAll.
    async enableGpuNetworkAcceleration(
        nodePoolId: string,
        accelerationLabelKey: string = MnnvlUtils.mnnvlLabel
    ): Promise<void> {
        await operations.updateRunaiConfig({
            spec: { 'anyworkload-controller': { GPUNetworkAccelerationEnabled: true } }
        });

        const nodePool = (await nodePoolUtils.getNodePoolById(nodePoolId)).body;

        const updateFields = new NodepoolUpdateFields();
        updateFields.labelKey = nodePool.labelKey;
        updateFields.labelValue = nodePool.labelValue;
        updateFields.gpuNetworkAccelerationDetection = NodepoolGPUNetworkAccelerationDetection.Use;
        updateFields.gpuNetworkAccelerationLabelKey = accelerationLabelKey;

        await nodePoolUtils.updateClusterNodePool(nodePoolId, updateFields);

        await nodePoolUtils.waitForClusterNodePoolProperty(
            nodePoolId,
            'status.gpuNetworkAccelerationDetected',
            (value) => value === true,
            120
        );

        logger.info(
            `Enabled GPU network acceleration (runaiconfig + node pool ${nodePoolId} detection=Use)`
        );
    }

    // Reverts enableGpuNetworkAcceleration: resets the node pool detection to Auto and turns
    // the cluster-wide runaiconfig flag back off, leaving the cluster in its default state.
    async disableGpuNetworkAcceleration(nodePoolId: string): Promise<void> {
        const nodePool = (await nodePoolUtils.getNodePoolById(nodePoolId)).body;

        const updateFields = new NodepoolUpdateFields();
        updateFields.labelKey = nodePool.labelKey;
        updateFields.labelValue = nodePool.labelValue;
        updateFields.gpuNetworkAccelerationDetection = NodepoolGPUNetworkAccelerationDetection.Auto;

        await nodePoolUtils.updateClusterNodePool(nodePoolId, updateFields);

        await operations.updateRunaiConfig({
            spec: { 'anyworkload-controller': { GPUNetworkAccelerationEnabled: false } }
        });

        logger.info(
            `Disabled GPU network acceleration (node pool ${nodePoolId} detection=Auto + runaiconfig off)`
        );
    }
}
