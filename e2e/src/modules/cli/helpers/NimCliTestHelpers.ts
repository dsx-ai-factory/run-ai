/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger, workloadsUtilsV2, waitUtils, config } from '../../helpers';
import { runaiCliV2 } from '../index';
import { cliV2WorkloadType, describeFlags, NimSubmitFlagsMerged, NimScaleFlags } from '../tools/RunaiCliV2.constants';
import type { Phase } from '../../api/apis/workloadsApi/generated/model/phase';
import { expect } from 'chai';

const WORKLOAD_APPEAR_POLL_INTERVAL_MS = 2000;
const WORKLOAD_APPEAR_DEFAULT_TIMEOUT_SEC = 45;
export const NGC_AUTH_SECRET_NAME = 'ngc-secret';

export function createBasicNimSubmitFlagsWithPvc(
    pvcName: string,
    overrides?: Record<string, any>
): Partial<typeof NimSubmitFlagsMerged> {
    return {
        [NimSubmitFlagsMerged.image]: config.quickStartImage.nim,
        [NimSubmitFlagsMerged.servingPort]: '8000',
        [NimSubmitFlagsMerged.modelExistingPvc]: `claimname=${pvcName}`,
        [NimSubmitFlagsMerged.ngcAuthSecret]: NGC_AUTH_SECRET_NAME,
        [NimSubmitFlagsMerged.cpuCoreRequest]: '0.1',
        [NimSubmitFlagsMerged.cpuMemoryRequest]: '100M',
        ...overrides
    } as Partial<typeof NimSubmitFlagsMerged>;
}

export function createBasicNimSubmitFlags(
    overrides?: Record<string, any>
): Partial<typeof NimSubmitFlagsMerged> {
    return createBasicNimSubmitFlagsWithPvc('nim-model-pvc', overrides);
}

export function createNimWithExistingPvcFlags(
    pvcName: string,
    overrides?: Record<string, any>
): Partial<typeof NimSubmitFlagsMerged> {
    return createBasicNimSubmitFlagsWithPvc(pvcName, overrides);
}

export function createNimWithAutoscalingFlags(
    pvcName: string,
    minReplicas: number,
    maxReplicas: number,
    metric: string,
    metricThreshold: number,
    overrides?: Record<string, any>
): Partial<typeof NimSubmitFlagsMerged> {
    return {
        ...createBasicNimSubmitFlagsWithPvc(pvcName, overrides),
        [NimSubmitFlagsMerged.minReplicas]: minReplicas.toString(),
        [NimSubmitFlagsMerged.maxReplicas]: maxReplicas.toString(),
        [NimSubmitFlagsMerged.metric]: metric,
        [NimSubmitFlagsMerged.metricThreshold]: metricThreshold.toString()
    } as Partial<typeof NimSubmitFlagsMerged>;
}

export function createNimWithWorkersFlags(
    pvcName: string,
    workerCount: number,
    overrides?: Record<string, any>
): Partial<typeof NimSubmitFlagsMerged> {
    return {
        ...createBasicNimSubmitFlagsWithPvc(pvcName, overrides),
        [NimSubmitFlagsMerged.workers]: workerCount.toString()
    } as Partial<typeof NimSubmitFlagsMerged>;
}

export function createNimWithFixedReplicasFlags(
    pvcName: string,
    replicas: number,
    overrides?: Record<string, any>
): Partial<typeof NimSubmitFlagsMerged> {
    return {
        ...createBasicNimSubmitFlagsWithPvc(pvcName, overrides),
        [NimSubmitFlagsMerged.replicas]: replicas.toString()
    } as Partial<typeof NimSubmitFlagsMerged>;
}

export function submitNimWorkload(
    submitFlags: Partial<typeof NimSubmitFlagsMerged>,
    workloadName: string,
    createdWorkloadNames: string[],
    logMessage?: string
): void {
    runaiCliV2.submit(submitFlags as any, workloadName, cliV2WorkloadType.inferenceNim);
    createdWorkloadNames.push(workloadName);
    logger.info(logMessage || `Submitted NIM workload ${workloadName}`);
}

export function describeNimWorkload(workloadName: string, projectName?: string): string {
    const flags = projectName ? { [describeFlags.project]: projectName } : {};
    const result = runaiCliV2.describe(flags, workloadName, cliV2WorkloadType.inferenceNim);
    logger.info(`CLI describe result:\n${result}`);
    return result;
}

export function describeNimWorkloadJson(workloadName: string, projectName?: string): any {
    const flags: Record<string, string> = { [describeFlags.output]: 'json' };
    if (projectName) {
        flags[describeFlags.project] = projectName;
    }
    const result = runaiCliV2.describe(flags, workloadName, cliV2WorkloadType.inferenceNim);
    return JSON.parse(result);
}

export function scaleNimWorkload(
    workloadName: string,
    replicas: number,
    force: boolean = false,
    projectFlag?: string
): string {
    const flags: Record<string, string> = {
        [NimScaleFlags.replicas]: replicas.toString()
    };
    if (force) {
        flags[NimScaleFlags.force] = '';
    }
    if (projectFlag) {
        flags[describeFlags.project] = projectFlag;
    }
    return runaiCliV2.scaleNim(workloadName, flags as Partial<typeof NimScaleFlags>);
}

export function updateNimWorkload(
    workloadName: string,
    flags: Record<string, string>
): string {
    return runaiCliV2.updateNim(workloadName, flags as Partial<typeof NimSubmitFlagsMerged>);
}

export function deleteNimWorkload(workloadName: string, createdWorkloadNames: string[]): void {
    runaiCliV2.delete({}, workloadName, cliV2WorkloadType.inferenceNim, false);
    const index = createdWorkloadNames.indexOf(workloadName);
    if (index > -1) {
        createdWorkloadNames.splice(index, 1);
    }
    logger.info(`Deleted NIM workload ${workloadName}`);
}

export async function waitForNimWorkloadByNameAndProjectId(
    workloadName: string,
    projectId: string | number,
    timeoutSeconds: number = WORKLOAD_APPEAR_DEFAULT_TIMEOUT_SEC
): Promise<void> {
    await waitUtils.waitForCondition(
        async () => {
            try {
                const { body } = await workloadsUtilsV2.getWorkloadByName(workloadName, projectId);
                return body.workloads?.some((w) => w.name === workloadName) === true;
            } catch {
                return false;
            }
        },
        `NIM workload ${workloadName} to appear in project ${projectId}`,
        true,
        timeoutSeconds,
        WORKLOAD_APPEAR_POLL_INTERVAL_MS / 1000
    );
    logger.debug(`NIM workload ${workloadName} appeared in project ${projectId}`);
}

export async function describeNimWorkloadAsync(
    workloadName: string,
    projectId: string | number,
    projectName?: string,
    timeoutSeconds: number = WORKLOAD_APPEAR_DEFAULT_TIMEOUT_SEC
): Promise<string> {
    await waitForNimWorkloadByNameAndProjectId(workloadName, projectId, timeoutSeconds);
    return describeNimWorkload(workloadName, projectName);
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
    logger.info(`Retrieved workload ID for NIM ${workloadName}: ${workloadId}`);
    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, phase, timeoutSeconds);
    logger.info(`NIM workload ${workloadName} is in ${phase} phase`);
    return workloadId;
}

export function verifyNimBasicInfo(describeResult: string, workloadName: string): void {
    expect(describeResult).to.include(workloadName);
    expect(describeResult).to.match(/NIM Service|NIMService|InferenceNim/i);
}

export function verifyNimImage(describeResult: string, imageName: string): void {
    expect(describeResult).to.include(imageName);
}

export function verifyNimServingPort(describeResult: string, port: number): void {
    expect(describeResult).to.match(new RegExp(`Port\\s*:\\s*${port}`, 'i'));
}

export function verifyNimPhase(describeResult: string, phase: string): void {
    expect(describeResult).to.match(new RegExp(`Phase\\s*:\\s*${phase}`, 'i'));
}

export function verifyNimReplicas(describeResult: string, replicas: number): void {
    expect(describeResult).to.match(new RegExp(`Replicas\\s*.*${replicas}`, 'i'));
}

export function verifyNimAutoscaling(
    describeResult: string,
    minReplicas: number,
    maxReplicas: number,
    metric: string
): void {
    expect(describeResult).to.match(/Autoscaling/i);
    expect(describeResult).to.match(new RegExp(`Min Replicas\\s*:\\s*${minReplicas}`, 'i'));
    expect(describeResult).to.match(new RegExp(`Max Replicas\\s*:\\s*${maxReplicas}`, 'i'));
    expect(describeResult).to.include(metric);
}

export function verifyNimMultiNode(describeResult: string, workers: number): void {
    expect(describeResult).to.match(new RegExp(`Workers\\s*.*${workers}`, 'i'));
}

export function verifyNimPvcType(describeResult: string, pvcType: 'new' | 'existing'): void {
    if (pvcType === 'existing') {
        expect(describeResult).to.match(/Existing PVC|existingPvc/i);
    } else {
        expect(describeResult).to.match(/New PVC|newPvc|Model Store/i);
    }
}

export function verifyNimModelStore(describeResult: string): void {
    expect(describeResult).to.match(/Model Store/i);
}

export function verifyNimCpuResources(
    describeResult: string,
    cpuCores: string | number,
    cpuMemory: string
): void {
    expect(describeResult).to.match(new RegExp(`CPU.*${cpuCores}`, 'i'));
    expect(describeResult).to.match(new RegExp(`Memory.*${cpuMemory}`, 'i'));
}

export function verifyNimEnvironmentVariable(describeResult: string, key: string, value: string): void {
    expect(describeResult).to.match(new RegExp(`${key}\\s*:\\s*${value}`, 'i'));
}

export function verifyNimRunningPods(describeResult: string, running: number, total: number): void {
    expect(describeResult).to.match(new RegExp(`Running Pods\\s*:\\s*${running}/${total}`, 'i'));
}
