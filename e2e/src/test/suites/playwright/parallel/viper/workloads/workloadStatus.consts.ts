/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    cliV2TrainingType,
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { config, crds, CustomResourceDefinition } from '../../../../../../modules/helpers';

const invalidImage = {
    [SubmitFlagsMerged.image]: 'invalid-image'
} as Partial<typeof SubmitFlagsMerged>;
const runningImage = {
    [SubmitFlagsMerged.image]: config.quickStartImage.sleepInfinity
} as Partial<typeof SubmitFlagsMerged>;
const parallelism = {
    [SubmitFlagsMerged.parallelism]: 2
} as Partial<typeof SubmitFlagsMerged>;
const singleInitPhaseMsg = 'Workload is initializing, the pod is still not ready';
const parallelInitPhaseMsg =
    'Workload is initializing, minimum ready pods not reached, waiting for 1 pods to be ready';

export interface WorkloadCondition {
    type: string;
    status: string;
    message: string;
}

export interface WorkloadParams {
    submitFlags: Partial<typeof SubmitFlagsMerged>;
    wlType: cliV2WorkloadType;
    name: string;
    expectedPhaseMsg: string;
    runningPods: number;
    crd?: CustomResourceDefinition;
    minRequiredReadyPods?: number;
    trainingType?: cliV2TrainingType;
    condition?: WorkloadCondition;
}

const runningCondition: WorkloadCondition = {
    type: 'Running',
    status: 'True',
    message: undefined
};

const initSingleCondition: WorkloadCondition = {
    type: 'Initialized',
    status: 'False',
    message: singleInitPhaseMsg
};
const initParallelCondition: WorkloadCondition = {
    type: 'Initialized',
    status: 'False',
    message: parallelInitPhaseMsg
};

const trainingParallelRunning: WorkloadParams = {
    wlType: cliV2WorkloadType.training,
    submitFlags: { ...runningImage, ...parallelism },
    name: 'train-parallel-run',
    expectedPhaseMsg: singleInitPhaseMsg,
    condition: runningCondition,
    minRequiredReadyPods: 1,
    runningPods: 1,
    crd: crds.runaiTrainingWorkloadCrd
};
const trainingSingleRunning: WorkloadParams = {
    wlType: cliV2WorkloadType.training,
    submitFlags: runningImage,
    name: 'train-single-run',
    expectedPhaseMsg: singleInitPhaseMsg,
    condition: runningCondition,
    minRequiredReadyPods: 1,
    runningPods: 1,
    crd: crds.runaiTrainingWorkloadCrd
};
const workspaceRunning: WorkloadParams = {
    wlType: cliV2WorkloadType.workspace,
    submitFlags: runningImage,
    name: 'workspace-run',
    expectedPhaseMsg: singleInitPhaseMsg,
    condition: runningCondition,
    minRequiredReadyPods: 1,
    runningPods: 1,
    crd: crds.runaiInteractiveWorkloadCrd
};
const workspaceInit: WorkloadParams = {
    wlType: cliV2WorkloadType.workspace,
    submitFlags: invalidImage,
    name: 'workspace-init',
    expectedPhaseMsg: singleInitPhaseMsg,
    condition: initSingleCondition,
    runningPods: 0,
    crd: crds.runaiInteractiveWorkloadCrd
};
const trainingSingleInit: WorkloadParams = {
    wlType: cliV2WorkloadType.training,
    submitFlags: invalidImage,
    name: 'training-single-pod-init',
    expectedPhaseMsg: singleInitPhaseMsg,
    condition: initSingleCondition,
    runningPods: 0,
    crd: crds.runaiTrainingWorkloadCrd
};
const trainingParallelInit: WorkloadParams = {
    wlType: cliV2WorkloadType.training,
    submitFlags: { ...invalidImage, ...parallelism },
    name: 'training-parallel-pods-init',
    expectedPhaseMsg: parallelInitPhaseMsg,
    condition: initParallelCondition,
    runningPods: 0,
    crd: crds.runaiTrainingWorkloadCrd
};
const mpiInit: WorkloadParams = {
    wlType: cliV2WorkloadType.training,
    submitFlags: invalidImage,
    name: 'mpi',
    expectedPhaseMsg: parallelInitPhaseMsg,
    trainingType: cliV2TrainingType.mpi,
    condition: initParallelCondition,
    runningPods: 0
};
const xgboostInit: WorkloadParams = {
    wlType: cliV2WorkloadType.training,
    submitFlags: invalidImage,
    name: 'xg-boost',
    expectedPhaseMsg: parallelInitPhaseMsg,
    trainingType: cliV2TrainingType.xgBoost,
    condition: initParallelCondition,
    runningPods: 0,
    crd: crds.xgboostJobCrd
};
const tfInit: WorkloadParams = {
    wlType: cliV2WorkloadType.training,
    submitFlags: invalidImage,
    name: 'tensor-flow',
    expectedPhaseMsg: parallelInitPhaseMsg,
    trainingType: cliV2TrainingType.xgBoost,
    condition: initParallelCondition,
    runningPods: 0,
    crd: crds.tfJobCrd
};
const pyTorchInit: WorkloadParams = {
    wlType: cliV2WorkloadType.training,
    submitFlags: invalidImage,
    name: 'pytorch',
    expectedPhaseMsg: parallelInitPhaseMsg,
    trainingType: cliV2TrainingType.pyTorch,
    condition: initParallelCondition,
    runningPods: 0,
    crd: crds.pyTorchJobCrd
};
export const runningWorkloadParams: WorkloadParams[] = [
    workspaceRunning,
    trainingSingleRunning,
    trainingParallelRunning
];
export const initWorkloadParams: WorkloadParams[] = [
    workspaceInit,
    trainingSingleInit,
    trainingParallelInit,
    mpiInit,
    xgboostInit,
    tfInit,
    pyTorchInit
];
