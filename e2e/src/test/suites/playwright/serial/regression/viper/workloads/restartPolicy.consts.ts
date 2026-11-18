/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { config, crds } from '../../../../../../../modules/helpers';
import {
    submitInteractiveWorkload,
    submitTrainWorkload
} from '../../../../../../../modules/helpers/models/workloads';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadUtils';
import { cliV2WorkloadType } from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';

interface jobCrdParams {
    specName: string;
    workerName?: string;
    leaderName?: string;
}

export const mpiCrdParams: jobCrdParams = {
    specName: 'mpiReplicaSpecs',
    leaderName: 'Launcher',
    workerName: 'Worker'
};
export const jaxCrdParams: jobCrdParams = {
    specName: 'jaxReplicaSpecs',
    workerName: 'Worker'
};
export const xgbCrdParams: jobCrdParams = {
    specName: 'xgbReplicaSpecs',
    leaderName: 'Master',
    workerName: 'Worker'
};
export const tfCrdParams: jobCrdParams = {
    specName: 'tfReplicaSpecs',
    leaderName: 'Chief',
    workerName: 'Worker'
};
export const pytCrdParams: jobCrdParams = {
    specName: 'pytorchReplicaSpecs',
    leaderName: 'Master',
    workerName: 'Worker'
};
export const trainingCrdParams: jobCrdParams = {
    specName: 'spec'
};
export const workspaceCrdParams: jobCrdParams = {
    specName: 'spec'
};

export const workloadCrdParamsMapper = {
    [cliV2WorkloadType.workspace]: workspaceCrdParams,
    [cliV2WorkloadType.training]: trainingCrdParams,
    [cliV2WorkloadType.jax]: jaxCrdParams,
    [cliV2WorkloadType.xgBoost]: xgbCrdParams,
    [cliV2WorkloadType.pyTorch]: pytCrdParams,
    [cliV2WorkloadType.tensorFlow]: tfCrdParams,
    [cliV2WorkloadType.mpi]: mpiCrdParams
};
