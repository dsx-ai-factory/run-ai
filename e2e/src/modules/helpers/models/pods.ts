/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { logger, waitUtils } from '../index';
import { expect } from 'chai';
import { kubeCli } from '../../cli';
import { V1Pod } from '@kubernetes/client-node';

export enum PodStatus {
    RUNNING = 'Running',
    PENDING = 'Pending',
    FAILED = 'Failed',
    SUCCEEDED = 'Succeeded',
    START_ERROR = 'StartError'
}

// Label keys
export const podLabelKeys = {
    INFERENCE_WORKLOAD: 'run.ai/inferenceworkload'
};

export async function extractJobPodsFromCluster(
    jobName: string,
    namespace: string,
    minExpectedNumOfPods: number,
    expectedStatus: string[],
    clusterTimeout: number
): Promise<V1Pod[]> {
    let clusterPods;
    try {
        await waitUtils.waitForCondition(
            async () => {
                clusterPods = await waitUtils.waitForPods(
                    jobName,
                    namespace,
                    minExpectedNumOfPods,
                    clusterTimeout
                );
                // logger.debug(`Get pods data for job ${jobName}`, clusterPods);
                return (
                    clusterPods.filter((pod) =>
                        expectedStatus.includes(pod.metadata.annotations['runai-calculated-status'])
                    ).length >= minExpectedNumOfPods
                );
            },
            `wait for ${minExpectedNumOfPods} number of pods with ${expectedStatus} status`,
            true,
            clusterTimeout
        );
        logger.debug(`Extracted pod cluster-data for ${jobName}`, { ClusterPods: clusterPods });
    } catch (e) {
        logger.error(e.message, { ClusterPods: clusterPods });

        expect(e).equal(
            undefined,
            `Failed to find at least ${minExpectedNumOfPods} pods with 
            status ${expectedStatus} for the job ${jobName}`
        );
    }
    return clusterPods;
}

export function validatePodsGpuMetric(clusterPods: V1Pod[], jobGpu: number): void {
    clusterPods.forEach((pod) => {
        expect(
            pod.metadata.annotations['received-resource-type'],
            `Incorrect received-resource-type annotation for fraction pod ${pod.metadata.name}.
                 Annotations: ${JSON.stringify(pod.metadata.annotations)}`
        ).equal('Fraction');
        expect(
            pod.metadata.annotations['gpu-fraction'],
            `Incorrect gpu-fraction annotation for pod ${pod.metadata.name} (should be given fraction).
                 Annotations: ${JSON.stringify(pod.metadata.annotations)}`
        ).equal(String(jobGpu));
        if (pod.metadata.annotations['runai-calculated-status'] === 'Running') {
            expect(
                Number(pod.metadata.annotations['runai-allocated-gpus']),
                `Incorrect runai-allocated-gpus annotation for pod ${
                    pod.metadata.name
                } (should be given fraction).
                     Annotations: ${JSON.stringify(pod.metadata.annotations)}`
            ).to.lessThanOrEqual(jobGpu);
            expect(
                pod.metadata.annotations['runai-allocated-mig-gpus'],
                `Incorrect runai-allocated-mig-gpus annotation for pod ${
                    pod.metadata.name
                } (no migs for fractions).
                     Annotations: ${JSON.stringify(pod.metadata.annotations)}`
            ).equal(String(0));
            expect(
                Number(pod.metadata.annotations['runai-allocated-gpu-memory']),
                `Incorrect runai-allocated-mig-gpus annotation for pod ${
                    pod.metadata.name
                } (our image should be using part of the memory).
                     Annotations: ${JSON.stringify(pod.metadata.annotations)}`
            ).to.greaterThan(0);
        }
    });
}

export function extractFractionConfigMap(pod: V1Pod): string {
    for (const container of pod.spec.containers) {
        for (const containerEnvsMap of container.env) {
            if (containerEnvsMap.name === 'NVIDIA_VISIBLE_DEVICES') {
                return containerEnvsMap.valueFrom.configMapKeyRef.name;
            }
        }
    }
    throw new Error(`Couldn't extract fraction config map name from pod ${pod.metadata.name}`);
}

export async function extractGpuGroupIndex(pod: V1Pod): Promise<string> {
    const configMapName = extractFractionConfigMap(pod);
    const fractionConfigMap = await kubeCli.getConfigMap(configMapName, pod.metadata.namespace);
    return fractionConfigMap.data['RUNAI-VISIBLE-DEVICES'];
}
