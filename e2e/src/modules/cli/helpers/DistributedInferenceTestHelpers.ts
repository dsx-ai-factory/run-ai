/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger, workloadsUtilsV2, waitUtils, config } from '../../helpers';
import { runaiCliV2 } from '../index';
import { cliV2WorkloadType, SubmitFlagsMerged } from '../tools/RunaiCliV2.constants';
import type { Phase } from '../../api/apis/workloadsApi/generated/model/phase';
import type { V1Pod } from '@kubernetes/client-node';
import { expect } from 'chai';

const WORKLOAD_APPEAR_POLL_INTERVAL_MS = 2000;
const WORKLOAD_APPEAR_DEFAULT_TIMEOUT_SEC = 45;

export function createBasicSubmitFlags(overrides?: Record<string, any>): Partial<typeof SubmitFlagsMerged> {
    return {
        [SubmitFlagsMerged.image]: config.quickStartImage.inferenceHelloWorld,
        [SubmitFlagsMerged.workers]: '2',
        [SubmitFlagsMerged.gpuRequest]: '1',
        [SubmitFlagsMerged.servingPort]: '8080',
        ...overrides
    } as Partial<typeof SubmitFlagsMerged>;
}

export function submitDistributedInferenceWorkload(
    submitFlags: Partial<typeof SubmitFlagsMerged>,
    workloadName: string,
    createdWorkloadNames: string[],
    logMessage?: string
): void {
    runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.inference, undefined, true);
    createdWorkloadNames.push(workloadName);
    logger.info(logMessage || `Submitted distributed inference workload ${workloadName}`);
}

export async function waitForWorkloadByNameAndProjectId(
    workloadName: string,
    projectId: string | number,
    timeoutSeconds: number = WORKLOAD_APPEAR_DEFAULT_TIMEOUT_SEC
): Promise<void> {
    const deadline = Date.now() + timeoutSeconds * 1000;
    while (Date.now() < deadline) {
        try {
            const { body } = await workloadsUtilsV2.getWorkloadByName(workloadName, projectId);
            const found = body.workloads?.some((w) => w.name === workloadName);
            if (found) {
                logger.debug(`Workload ${workloadName} appeared in project ${projectId}`);
                return;
            }
        } catch {
            // ignore and retry
        }
        await waitUtils.sleep(WORKLOAD_APPEAR_POLL_INTERVAL_MS);
    }
    throw new Error(
        `Workload ${workloadName} did not appear in project ${projectId} within ${timeoutSeconds}s`
    );
}

export function identifyLeaderAndWorkerPods(pods: V1Pod[]): {
    leaderPod: V1Pod | undefined;
    workerPod: V1Pod | undefined;
    leaderPodName: string | undefined;
    workerPodName: string | undefined;
} {
    const leaderPod = pods.find((pod) =>
        pod.spec?.containers?.some((container) => container.name === 'leader')
    );
    const workerPod = pods.find((pod) =>
        pod.spec?.containers?.some((container) => container.name === 'worker')
    );

    const leaderPodName = leaderPod?.metadata?.name;
    const workerPodName = workerPod?.metadata?.name;

    if (leaderPodName && workerPodName) {
        logger.info(`Identified leader pod: ${leaderPodName}`);
        logger.info(`Identified worker pod: ${workerPodName}`);
    }

    return { leaderPod, workerPod, leaderPodName, workerPodName };
}

export function deleteDistributedInferenceWorkload(
    workloadName: string,
    createdWorkloadNames: string[]
): void {
    runaiCliV2.delete(undefined, workloadName, cliV2WorkloadType.inference, true);
    const index = createdWorkloadNames.indexOf(workloadName);
    if (index > -1) {
        createdWorkloadNames.splice(index, 1);
    }
    logger.info(`Deleted distributed inference workload ${workloadName}`);
}

export async function getWorkloadIdAndWaitForPhase(
    workloadName: string,
    projectId: string | number,
    phase: Phase,
    timeoutSeconds: number = 120
): Promise<string> {
    const workloadId = await workloadsUtilsV2.waitForWorkloadIdByWorkloadNameAndProjectId(
        workloadName,
        projectId.toString(),
        WORKLOAD_APPEAR_DEFAULT_TIMEOUT_SEC
    );
    logger.info(`Retrieved workload ID for ${workloadName}: ${workloadId}`);
    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, phase, timeoutSeconds);
    logger.info(`Workload ${workloadName} is in ${phase} phase`);
    return workloadId;
}

export function describeWorkload(workloadName: string): string {
    const result = runaiCliV2.describe({}, workloadName, cliV2WorkloadType.inference, true);
    logger.info(`CLI describe result:\n${result}`);
    return result;
}

export async function describeWorkloadAsync(
    workloadName: string,
    projectId: string | number,
    timeoutSeconds: number = WORKLOAD_APPEAR_DEFAULT_TIMEOUT_SEC
): Promise<string> {
    await waitForWorkloadByNameAndProjectId(workloadName, projectId, timeoutSeconds);
    return describeWorkload(workloadName);
}

export function verifyBasicWorkloadInfo(describeResult: string, workloadName: string): void {
    expect(describeResult).to.include(workloadName);
    expect(describeResult).to.include('DistributedInference');
}

export function verifyWorkerReplicas(describeResult: string, count: number): void {
    expect(describeResult).to.match(new RegExp(`Workers per Replica\\s*:\\s*${count}`, 'i'));
}

export function verifyReplicasCount(describeResult: string, requested: number): void {
    expect(describeResult).to.match(new RegExp(`Replicas \\(Requested\\)\\s*:\\s*${requested}`, 'i'));
}

export function verifyGpuCompute(describeResult: string, expectedValue: string | number): void {
    expect(describeResult).to.match(new RegExp(`GPU Compute\\s+${expectedValue}`, 'i'));
}

export function verifyImage(describeResult: string, imageName: string): void {
    expect(describeResult).to.include(imageName);
}

export function verifyCpuCores(describeResult: string, expectedValue: string | number): void {
    expect(describeResult).to.match(new RegExp(`CPU Cores\\s+${expectedValue}`, 'i'));
}

export function verifyCpuMemory(describeResult: string, expectedValue: string): void {
    expect(describeResult).to.match(new RegExp(`CPU Memory\\s+${expectedValue}`, 'i'));
}

export function verifyPhase(describeResult: string, phase: string): void {
    expect(describeResult).to.match(new RegExp(`Phase\\s*:\\s*${phase}`, 'i'));
}

export function verifyEnvironmentVariable(describeResult: string, key: string, value: string): void {
    expect(describeResult).to.match(new RegExp(`${key}\\s*:\\s*${value}`, 'i'));
}

export function verifyServingPort(describeResult: string, port: number): void {
    expect(describeResult).to.include('Serving Configuration');
    expect(describeResult).to.match(new RegExp(`Port\\s*:\\s*${port}`, 'i'));
    expect(describeResult).to.match(/Serving URL\s*:/i);
}

export function verifyRunningPods(
    describeResult: string,
    total: number,
    leaders: number,
    workers: number
): void {
    expect(describeResult).to.match(
        new RegExp(`Running Pods\\s*:\\s*${total}/${total}.*${leaders} leaders.*${workers} workers`, 'i')
    );
}
