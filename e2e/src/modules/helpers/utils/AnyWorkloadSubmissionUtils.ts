/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { V1Container } from '@kubernetes/client-node';

import { kubeCli } from '../../cli';
import { flowRootState } from '../config/config';
import {
    AnyWorkloadSubmissionType,
    getVersionBasedSubmissionRoute,
    WorkloadSubmissionRoute
} from './AnyWorkloadSubmissionRules';

export {
    AnyWorkloadSubmissionType,
    MIN_ANY_WORKLOAD_SUBMISSION_CLUSTER_VERSION,
    WorkloadSubmissionRoute
} from './AnyWorkloadSubmissionRules';

export const WORKLOADS_DEPLOYMENT = 'runai-backend-workloads';
export const WORKLOADS_NAMESPACE = 'runai-backend';
export const ANY_WORKLOAD_SUBMISSION_ENV = 'ANY_WORKLOAD_SUBMISSION';

const submissionTypeEnv: Record<AnyWorkloadSubmissionType, string> = {
    [AnyWorkloadSubmissionType.Workspace]: 'ANY_WORKLOAD_WORKSPACE_SUBMISSION',
    [AnyWorkloadSubmissionType.Training]: 'ANY_WORKLOAD_TRAINING_SUBMISSION',
    [AnyWorkloadSubmissionType.Distributed]: 'ANY_WORKLOAD_DISTRIBUTED_SUBMISSION',
    [AnyWorkloadSubmissionType.Inference]: 'ANY_WORKLOAD_INFERENCE_SUBMISSION',
    [AnyWorkloadSubmissionType.DistributedInference]: 'ANY_WORKLOAD_DINF_SUBMISSION'
};

interface SubmissionFlags {
    global: boolean;
    byType: Record<AnyWorkloadSubmissionType, boolean>;
}

let cachedFlags: SubmissionFlags | undefined;

export function resolveWorkloadsContainerIndex(containers: V1Container[]): number {
    if (containers.length === 0) {
        throw new Error(`deployment ${WORKLOADS_NAMESPACE}/${WORKLOADS_DEPLOYMENT} has no containers`);
    }
    const index = containers.findIndex(
        (container) => container.name === WORKLOADS_DEPLOYMENT || container.name.includes('workload')
    );
    return index >= 0 ? index : 0;
}

export async function isAnyWorkloadSubmissionEnabled(
    workloadType: AnyWorkloadSubmissionType,
    refresh = false
): Promise<boolean> {
    const requiredRoute = getVersionBasedSubmissionRoute(flowRootState.cluster?.version, workloadType);
    if (requiredRoute !== undefined) {
        return requiredRoute === WorkloadSubmissionRoute.WorkloadRunner;
    }

    if (cachedFlags !== undefined && !refresh) {
        return cachedFlags.global || cachedFlags.byType[workloadType];
    }

    const deployment = await kubeCli.getDeployment(WORKLOADS_DEPLOYMENT, WORKLOADS_NAMESPACE);
    const containers = deployment.spec?.template?.spec?.containers ?? [];
    const containerIndex = resolveWorkloadsContainerIndex(containers);
    const env = containers[containerIndex].env ?? [];
    const isEnabled = (name: string): boolean =>
        env.find((entry) => entry.name === name)?.value?.toLowerCase() === 'true';

    cachedFlags = {
        global: isEnabled(ANY_WORKLOAD_SUBMISSION_ENV),
        byType: Object.fromEntries(
            Object.entries(submissionTypeEnv).map(([type, name]) => [type, isEnabled(name)])
        ) as Record<AnyWorkloadSubmissionType, boolean>
    };
    return cachedFlags.global || cachedFlags.byType[workloadType];
}

export function resolveWorkloadSubmissionRoute(enabled: boolean): WorkloadSubmissionRoute {
    return enabled ? WorkloadSubmissionRoute.WorkloadRunner : WorkloadSubmissionRoute.Legacy;
}

export async function getWorkloadSubmissionRoute(
    workloadType: AnyWorkloadSubmissionType,
    refresh = false
): Promise<WorkloadSubmissionRoute> {
    return resolveWorkloadSubmissionRoute(await isAnyWorkloadSubmissionEnabled(workloadType, refresh));
}

export function setAnyWorkloadSubmissionEnabledCache(enabled: boolean | undefined): void {
    cachedFlags =
        enabled === undefined
            ? undefined
            : {
                  global: enabled,
                  byType: Object.fromEntries(
                      Object.values(AnyWorkloadSubmissionType).map((type) => [type, false])
                  ) as Record<AnyWorkloadSubmissionType, boolean>
              };
}
