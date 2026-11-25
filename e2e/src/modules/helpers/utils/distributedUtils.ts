/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState } from '../config/config';
import { policyUtils, workloadsCrudUtils } from '..';
import http from 'http';
import { DistributedFramework, MatchExpressionOperator } from '../../api/apis/policyApi';
import { DistributedWorkload as DistributedModel } from '../../api/apis/workloadsApi/generated/model/distributedWorkload';
import { Compute } from '../config/types';
import { PodAffinity } from '../../api/apis/workloadsApi/generated/api';

interface WorkerSpec {
    distributedFramework: DistributedFramework;
    image: string;
    command: string;
    numWorkers: number;
    compute: Compute;
    nodePools: Array<string>;
    podAffinity: PodAffinity;
}

export class DistributedUtils {
    async createDistributedTraining(
        projectId: string,
        workloadName: string,
        workerSpec: Partial<WorkerSpec> = {
            distributedFramework: DistributedFramework.Mpi,
            image: 'alpine:latest',
            command: 'sleep infinity',
            numWorkers: 1,
            compute: {
                gpuDevicesRequest: 0,
                cpuCoreRequest: 0.1,
                cpuMemoryRequest: '100M'
            }
        },
        masterSpec?: {
            distributedFramework: DistributedFramework;
            image: string;
            command: string;
            compute: Compute;
        },
        workerExcludeNodeNames?: string[],
        masterExcludeNodeNames?: string[]
    ): Promise<{ response: http.IncomingMessage; body: DistributedModel }> {
        if (workerExcludeNodeNames && workerExcludeNodeNames.length !== 0) {
            workerSpec['nodeAffinityRequired'] = policyUtils.getNodeSelectorTerm(
                MatchExpressionOperator.NotIn,
                workerExcludeNodeNames
            ).nodeAffinityRequired;
        }

        if (masterSpec && masterExcludeNodeNames && masterExcludeNodeNames.length !== 0) {
            masterSpec['nodeAffinityRequired'] = policyUtils.getNodeSelectorTerm(
                MatchExpressionOperator.NotIn,
                masterExcludeNodeNames
            ).nodeAffinityRequired;
        }

        workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;

        return workloadsCrudUtils.createDistributedV2(
            workloadName,
            false,
            flowRootState.cluster.uuid,
            projectId,
            workerSpec,
            masterSpec
        );
    }
}
