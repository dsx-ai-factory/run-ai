/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export type WorkloadRunnerResource = Record<string, any>;

export function findWorkloadRunnerByWorkloadName(
    runners: WorkloadRunnerResource[],
    workloadName: string
): WorkloadRunnerResource | undefined {
    return runners.find(
        (runner) =>
            runner.metadata?.name?.includes(workloadName) ||
            runner.metadata?.labels?.workloadName === workloadName ||
            runner.spec?.workload?.name === workloadName
    );
}
