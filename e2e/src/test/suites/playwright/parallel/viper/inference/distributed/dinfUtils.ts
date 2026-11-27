/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { DistributedInferenceSpecSpec } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceSpecSpec';
import {
    config,
    logger,
    waitUtils,
    workloadsUtilsV2,
    workloadsCrudUtils
} from '../../../../../../../modules/helpers';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { Workload } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/workload';
import { bashCli } from '../../../../../../../modules/cli';

export function getBaseWorkloadSpec(): DistributedInferenceSpecSpec {
    return {
        workers: 1,
        replicas: 1,
        servingPort: {
            port: 8080
        },
        leader: {
            image: config.quickStartImage.inferenceHelloWorld
        },
        worker: {
            image: config.quickStartImage.inferenceHelloWorld
        }
    };
}

export function getNumWorkers(workloadSpec: DistributedInferenceSpecSpec): number {
    return workloadSpec?.workers ?? 0;
}

export function getNumReplicas(workloadSpec: DistributedInferenceSpecSpec): number {
    return workloadSpec?.replicas ?? 1;
}

export function getReadyReplicas(workload: Workload): number {
    return workload?.additionalFields?.readyReplicas ?? 0;
}

export function getNumReplicasForWorkload(workload: Workload): number {
    return workload?.additionalFields?.specReplicas ?? 0;
}

export async function waitForNumReplicas(
    workloadId: string,
    expectedNumReplicas: number,
    timeout: number = 60
): Promise<Workload> {
    let workload: Workload;

    const isWorkloadWithCorrectNumReplicas = async (): Promise<boolean> => {
        try {
            workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;
            return getNumReplicasForWorkload(workload) === expectedNumReplicas;
        } catch (e) {
            logger.error(`Got an error: ${e.message}`);
            return false;
        }
    };

    logger.debug(`waiting for workload ${workloadId} to have ${expectedNumReplicas} replicas`);
    await waitUtils.waitForCondition(
        isWorkloadWithCorrectNumReplicas,
        `workload: ${workloadId}, expected replicas: ${expectedNumReplicas}`,
        true,
        timeout
    );

    return workload;
}

export async function waitForReadyReplicas(
    workloadId: string,
    expectedReadyReplicas: number,
    timeout: number = 60
): Promise<Workload> {
    let workload: Workload;

    const isWorkloadWithCorrectReadyReplicas = async (): Promise<boolean> => {
        try {
            workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;
            return getReadyReplicas(workload) === expectedReadyReplicas;
        } catch (e) {
            logger.error(`Got an error: ${e.message}`);
            return false;
        }
    };

    logger.debug(`waiting for workload ${workloadId} to have ${expectedReadyReplicas} ready replicas`);
    await waitUtils.waitForCondition(
        isWorkloadWithCorrectReadyReplicas,
        `workload: ${workloadId}, expected ready replicas: ${expectedReadyReplicas}`,
        true,
        timeout
    );

    return workload;
}

export function validateResourceAllocation(
    workload: Workload,
    workloadSpec: DistributedInferenceSpecSpec
): void {
    // Calculate total expected resources
    const leaderResources =
        workloadSpec.leader?.compute?.gpuDevicesRequest ??
        workloadSpec.leader?.compute?.gpuPortionRequest ??
        0;

    const workerResources =
        workloadSpec.worker?.compute?.gpuDevicesRequest ??
        workloadSpec.worker?.compute?.gpuPortionRequest ??
        0;

    const workers = getNumWorkers(workloadSpec);
    const replicas = getNumReplicas(workloadSpec);
    const readyReplicas = getReadyReplicas(workload);

    const expectedRequestedGpuResources = (leaderResources + workers * workerResources) * replicas;
    const expectedAllocatedGpuResources = (leaderResources + workers * workerResources) * readyReplicas;

    // Validate with floating-point tolerance
    const podsRequestedGpu = workload.podsRequestedResources?.gpu?.request ?? 0;
    expect(podsRequestedGpu).to.be.closeTo(expectedRequestedGpuResources, 0.0001);

    const workloadRequestedGpu = workload.workloadRequestedResources?.gpu?.request ?? 0;
    expect(workloadRequestedGpu).to.be.closeTo(expectedRequestedGpuResources, 0.0001);

    const allocatedGpu = workload.allocatedResources?.gpu ?? 0;
    expect(allocatedGpu).to.be.closeTo(expectedAllocatedGpuResources, 0.0001);
}

export function validateUrls(workload: Workload, workloadSpec: DistributedInferenceSpecSpec): void {
    const readyReplicas = getReadyReplicas(workload);
    const urls = workload.urls || [];

    if (readyReplicas === 0) {
        expect(urls).to.have.lengthOf(0);
        return;
    }

    const servingPort = workloadSpec.servingPort;

    expect(urls).to.have.length.at.least(1);

    // Validate each URL format
    for (const url of urls) {
        const startsWithHttps = url.startsWith('https://');
        const startsWithHttpAndHasClusterLocal = url.startsWith('http://') && url.includes('cluster.local');

        expect(startsWithHttps || startsWithHttpAndHasClusterLocal).to.be.true;
    }
}

export function validateGetDistributedInference(
    workload: Workload,
    workloadSpec: DistributedInferenceSpecSpec,
    validateAllocation = true
): void {
    logger.debug('validating the get workload response');
    const workers = getNumWorkers(workloadSpec);
    const replicas = getNumReplicas(workloadSpec);
    const expectedPods = (1 + workers) * replicas;

    const expectedImages = new Set([
        workloadSpec.leader.image,
        ...(workloadSpec.worker && workloadSpec.worker.image ? [workloadSpec.worker.image] : [])
    ]);
    expect(workload.images).to.have.lengthOf(expectedImages.size);
    expect(workload.images).to.have.members(Array.from(expectedImages));
    expect(workload.requestedPods.number).to.equal(expectedPods);

    expect(workload.preemptible).to.equal(false);
    expect(workload.category).to.equal('Deploy');
    expect(workload.priorityClassName).to.equal('very-high');

    expect(workload).to.have.property('additionalFields');
    const additionalFields = workload.additionalFields;

    if (workloadSpec.startupPolicy) {
        expect(additionalFields).to.have.property('startupPolicy');
        expect(additionalFields.startupPolicy).to.equal(workloadSpec.startupPolicy);
    }

    if (workloadSpec.restartPolicy) {
        expect(additionalFields).to.have.property('restartPolicy');
        expect(additionalFields.restartPolicy).to.equal(workloadSpec.restartPolicy);
    }

    if (workload.phase === Phase.Running) {
        expect(workload.runningPods).to.equal(expectedPods);
    }

    if (workload.phase !== Phase.Creating && workload.phase !== Phase.Deleting) {
        expect(workload.phaseMessage).not.to.be.undefined;
        expect(workload.phaseMessage).not.to.be.empty;
        expect(workload.k8sPhase).not.to.be.undefined;
        expect(workload.k8sPhase).not.to.be.empty;
    }

    if (workload.phase !== Phase.Degraded) {
        validateUrls(workload, workloadSpec);
    }

    if (validateAllocation) {
        validateResourceAllocation(workload, workloadSpec);
    }
}

/**
 * Get pod ids for a workload
 * @param workloadId - The workload ID
 * @returns Array of pod ids
 */
export async function getWorkloadPodIds(workloadId: string): Promise<string[]> {
    logger.debug(`Getting pod ids for workload ${workloadId}`);

    try {
        const podsResponse = await workloadsUtilsV2.getWorkloadPods(workloadId);
        const podIds = podsResponse.body.pods.map((pod) => pod.id);

        logger.debug(`Found ${podIds.length} pods for workload ${workloadId}: ${podIds.join(', ')}`);
        return podIds;
    } catch (error) {
        logger.error(`Failed to get pod ids for workload ${workloadId}: ${error.message}`);
        throw error;
    }
}

function isSubset(superset: string[], subset: string[]): void {
    logger.info(
        `Validating subset: checking if ${JSON.stringify(subset)} is contained in ${JSON.stringify(superset)}`
    );

    for (const item of subset) {
        expect(superset).to.include(item, `Expected ${JSON.stringify(superset)} to include ${item}`);
    }
}

/**
 * Scale distributed inference workload via API
 * @param workloadId - The workload ID
 * @param newReplicas - The new number of replicas
 */
export async function scaleDistributedInferenceViaApi(
    workloadId: string,
    newReplicas: number
): Promise<void> {
    logger.info(`Scaling distributed inference workload ${workloadId} to ${newReplicas} replicas via API`);

    // Create update request with new replicas
    const updateRequest = {
        spec: {
            replicas: newReplicas
        }
    };

    // Update the workload via API
    await workloadsCrudUtils.updateDistributedInferenceSpec(workloadId, updateRequest);

    // Wait for the scaling to take effect
    await waitForReadyReplicas(workloadId, newReplicas, 120);

    logger.info(
        `Successfully scaled distributed inference workload ${workloadId} to ${newReplicas} replicas via API`
    );
}

/**
 * Scale distributed inference workload via kubectl
 * @param workloadId - The workload id
 * @param workloadName - The workload name
 * @param namespace - The namespace
 * @param newReplicas - The new number of replicas
 */
export async function scaleDistributedInferenceViaKubectl(
    workloadId: string,
    workloadName: string,
    namespace: string,
    newReplicas: number
): Promise<void> {
    logger.info(
        `Scaling distributed inference workload ${workloadName} to ${newReplicas} replicas via kubectl`
    );

    const scaleCommand = `kubectl scale distributedinferenceworkload ${workloadName} --replicas=${newReplicas} -n ${namespace}`;

    try {
        bashCli.exec(scaleCommand);
        logger.info(`Successfully executed kubectl scale command: ${scaleCommand}`);

        await waitForNumReplicas(workloadId, newReplicas);

        logger.info(
            `Successfully scaled distributed inference workload ${workloadName} to ${newReplicas} replicas via kubectl`
        );
    } catch (error) {
        logger.error(`Failed to scale distributed inference workload ${workloadName} via kubectl: ${error}`);
        throw error;
    }
}

/**
 * Validate scaling results for distributed inference workloads
 * @param workloadId - The workload ID
 * @param expectedReplicas - The expected number of replicas after scaling
 * @param workloadSpec - The original workload specification
 * @param originalPodIds - Optional array of pod ids that existed before scaling (to validate no recreation)
 */
export async function validateScalingResults(
    workloadId: string,
    expectedReplicas: number,
    workloadSpec: DistributedInferenceSpecSpec,
    originalPodIds?: string[]
): Promise<string[]> {
    logger.info(`Validating scaling results for workload ${workloadId}`);

    // Get the updated workload
    const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);

    // Validate the replicas were updated correctly
    const workers = getNumWorkers(workloadSpec);
    const expectedPods = (1 + workers) * expectedReplicas;

    expect(workload.body.requestedPods.number).to.equal(expectedPods);

    let newPodIds = originalPodIds;
    // Validate that old pods were not recreated if original pod ids are provided
    if (originalPodIds && originalPodIds.length > 0) {
        logger.info(`Validating that ${originalPodIds.length} original pods were not recreated`);

        // Get current pods after scaling
        const currentPodsResponse = await workloadsUtilsV2.getWorkloadPods(workloadId);
        const currentPods = currentPodsResponse.body.pods;
        newPodIds = currentPods.map((pod) => pod.id);

        if (originalPodIds.length < expectedPods) {
            // we scaled up, all original pods should remain
            isSubset(newPodIds, originalPodIds);
        } else {
            // we scaled down, current pods should be part of the old ones
            isSubset(originalPodIds, newPodIds);
        }
    }

    workloadSpec.replicas = expectedReplicas;

    // Validate the scaling didn't break the workload configuration
    validateGetDistributedInference(workload.body, workloadSpec);

    logger.info(`Successfully validated scaling results for workload ${workloadId}`);

    return newPodIds;
}
