/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { VersionUtils } from './VersionUtils';

// Mirrors checker.MinVersionOfAnyWorkloadForV1Workloads in the main repository.
export const MIN_ANY_WORKLOAD_SUBMISSION_CLUSTER_VERSION = '2.26.0';
// Mirrors checker.MinVersionNoNativeSubmission in the main repository.
const MIN_WORKLOAD_RUNNER_REQUIRED_CLUSTER_VERSION = '3.0.0';

export enum AnyWorkloadSubmissionType {
    Workspace = 'workspace',
    Training = 'training',
    Distributed = 'distributed',
    Inference = 'inference',
    DistributedInference = 'distributed-inference'
}

export enum WorkloadSubmissionRoute {
    Legacy = 'legacy',
    WorkloadRunner = 'workload-runner'
}

const versionUtils = new VersionUtils();

// Undefined means this version/type combination still depends on deployment flags.
export function getVersionBasedSubmissionRoute(
    clusterVersion: string | undefined,
    workloadType: AnyWorkloadSubmissionType
): WorkloadSubmissionRoute | undefined {
    // Build hashes must not make a released cluster compare below its version boundary.
    const releaseVersion = clusterVersion?.split('-')[0] ?? '';
    // Master builds and 3.0+ have no native submission path, regardless of deployment flags.
    if (
        versionUtils.isMasterBuild(clusterVersion) ||
        versionUtils.isVersionSupported(releaseVersion, MIN_WORKLOAD_RUNNER_REQUIRED_CLUSTER_VERSION)
    ) {
        return WorkloadSubmissionRoute.WorkloadRunner;
    }

    if (!versionUtils.isVersionSupported(releaseVersion, MIN_ANY_WORKLOAD_SUBMISSION_CLUSTER_VERSION)) {
        return WorkloadSubmissionRoute.Legacy;
    }

    // The product only redirects training/workspace/distributed workloads on 2.26.
    if (
        workloadType === AnyWorkloadSubmissionType.Inference ||
        workloadType === AnyWorkloadSubmissionType.DistributedInference
    ) {
        return WorkloadSubmissionRoute.Legacy;
    }
    return undefined;
}
