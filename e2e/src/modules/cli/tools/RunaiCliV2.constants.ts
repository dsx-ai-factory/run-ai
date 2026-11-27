/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    accessKeyUtils,
    crds,
    flowRootState,
    objectGenerator,
    User
} from '../../helpers';
import { runaiCliV2 } from '../index';

function excludeProperties<T extends object, K extends keyof T>(obj: T, ...excludeKeys: K[]): Omit<T, K> {
    const result = { ...obj };
    excludeKeys.forEach((key) => delete result[key]);
    return result;
}

export interface childProcessOutput {
    stdout: string[];
    stderr: string;
}

// for type safety of the unify logs command
export enum workloadLog {
    noun = 'workload'
}

export enum cliV2WorkloadType {
    workspace = 'workspace',
    training = 'training',
    inference = 'inference',
    inferenceNim = 'inference nim',
    mpi = 'mpi',
    tensorFlow = 'tf',
    xgBoost = 'xgboost',
    pyTorch = 'pytorch',
    jax = 'jax'
}

export const cliV2workloadNoun = 'workload';

export function getWorkloadTypeToCrd(): object {
    return {
        [cliV2WorkloadType.inference]: crds.runaiInferenceWorkloadCrd,
        [cliV2WorkloadType.workspace]: crds.runaiInteractiveWorkloadCrd,
        [cliV2WorkloadType.training]: crds.runaiTrainingWorkloadCrd,
        [cliV2WorkloadType.jax]: crds.jaxJobCrd,
        [cliV2WorkloadType.xgBoost]: crds.xgboostJobCrd,
        [cliV2WorkloadType.pyTorch]: crds.pyTorchJobCrd,
        [cliV2WorkloadType.tensorFlow]: crds.tfJobCrd,
        [cliV2WorkloadType.mpi]: crds.mpiJobCrd
    };
}

export const workloadTypesNoInference = [
    cliV2WorkloadType.jax,
    cliV2WorkloadType.mpi,
    cliV2WorkloadType.pyTorch,
    cliV2WorkloadType.tensorFlow,
    cliV2WorkloadType.xgBoost,
    cliV2WorkloadType.workspace,
    cliV2WorkloadType.training
];

export enum cliV2TrainingType {
    standard = 'standard',
    mpi = 'mpi',
    tensorFlow = 'tf',
    xgBoost = 'xgboost',
    pyTorch = 'pytorch',
    jax = 'jax'
}

export enum MyCredentialType {
    genericSecret = 'genericSecret',
    dockerRegistry = 'dockerRegistry',
    ngcApiKey = 'ngcApiKey'
}


export enum preemptibilityOptions {
    preemptible = 'preemptible',
    nonPreemptible = 'non-preemptible'
}

export enum ConfigFlags {
    setControlPlane = '--cp-url',
    setAuthServer = '--auth-url'
}

export enum LogsFlags {
    follow = '-f ',
    project = '--project '
}

export enum listFlags {
    json = '--json',
    yaml = '--yaml',
    framework = '--framework',
    limit = '--limit',
    offset = '--offset',
    allScopes = '-A',
    project = '-p',
    columns = '--columns',
    table = '--table',
    noHeaders = '--no-headers',
    pageSize = '--page-size',
    maxItems = '--max-items',
    noPagination = '--no-pagination',
    nextToken = '--next-token',
    type = '--type'
}

export enum collectDiagnosticLogsFlags {
    outputDir = '--output-dir',
    noPrevious = '--no-previous',
    namespaces = '--namespaces'
}

export enum CommonWorkloadFlags {
    project = '--project ',
    uuid = '--uuid '
}

export enum cliV2Actions {
    submit = 'submit',
    delete = 'delete',
    list = 'list',
    update = 'update',
    describe = 'describe'
}

enum SubmitFlags {
    image = '--image',
    externalUrl = '--external-url',
    command = '-c ',
    servingPort = '--serving-port ', //Inference only
    gpuPortionLimit = '--gpu-portion-limit',
    gpuPortionRequest = '--gpu-portion-request',
    gpuRequest = '-g',
    cpuCoreRequest = '--cpu-core-request',
    cpuCoreLimit = '--cpu-core-limit',
    cpuMemoryRequest = '--cpu-memory-request',
    cpuMemoryLimit = '--cpu-memory-limit',
    backoffLimit = '--backoff-limit',
    parallelism = '--parallelism',
    slotsPerWorker = '--slots-per-worker', //MPI only
    workers = '--workers',
    envVar = '-e',
    noMaster = '--no-master', //TensorFlow and PyTorch only
    maxReplicas = '--max-replicas', //TensorFlow and PyTorch only
    minReplicas = '--min-replicas', //TensorFlow and PyTorch only
    cleanPodPolicy = '--clean-pod-policy',
    attach = '--attach',
    priority = '--priority',
    privileged = '--privileged',
    masterRestartPolicy = '--master-restart-policy',
    restartPolicy = '--restart-policy',
    excludeNode = '--exclude-node',
    largeShm = '--large-shm',
    template = '--template',
    preemptibility = '--preemptibility',
    preemptible = '--preemptible',
    workingDir = '--working-dir',
    dryRun = '--dry-run',
    annotation = '--annotation',
    nodePools = '--node-pools',

    // User identity / security
    createHomeDir = '--create-home-dir',
    runAsUser = '--run-as-user',
    runAsUid = '--run-as-uid',
    runAsGid = '--run-as-gid',
    supplementalGroups = '--supplemental-groups',

    // Container options
    imagePullPolicy = '--image-pull-policy',
    stdin = '--stdin',
    tty = '--tty',

    // Networking
    hostIpc = '--host-ipc',
    hostNetwork = '--host-network',
    port = '--port',

    // Storage
    existingPvc = '--existing-pvc',
    newPvc = '--new-pvc',
    nfs = '--nfs',
    gitSync = '--git-sync',
    s3 = '--s3',
    label = '--label',

    // Asset flags
    environment = '--environment',
    compute = '--compute',
    datasource = '--datasource',
    datavolume = '--datavolume',
    hostpath = '--host-path',
    envMyCredentials = '--env-my-credentials',
    imagePullMyCredentials = '--image-pull-my-credentials'}

enum DeleteFlags {
    byPassDialog = '-y'
}

// NIM-specific submit flags
enum NimSubmitFlags {
    image = '--image',
    ngcAuthSecret = '--ngc-auth-secret',
    modelExistingPvc = '--model-existing-pvc',
    modelNewPvc = '--model-new-pvc',
    modelNimCache = '--model-nim-cache',
    modelProfile = '--model-profile',
    servingPort = '--serving-port',
    replicas = '--replicas',
    minReplicas = '--min-replicas',
    maxReplicas = '--max-replicas',
    metric = '--metric',
    metricThreshold = '--metric-threshold',
    scaleWindowSeconds = '--scale-window-seconds',
    workers = '--workers',
    envVar = '-e',
    imagePullPolicy = '--image-pull-policy',
    imagePullSecret = '--image-pull-secret',
    cpuCoreRequest = '--cpu-core-request',
    cpuCoreLimit = '--cpu-core-limit',
    cpuMemoryRequest = '--cpu-memory-request',
    cpuMemoryLimit = '--cpu-memory-limit',
    gpuRequest = '-g',
    gpuMemoryRequest = '--gpu-memory-request',
    gpuMemoryLimit = '--gpu-memory-limit',
    gpuPortionRequest = '--gpu-portion-request',
    gpuPortionLimit = '--gpu-portion-limit',
    gpuRequestType = '--gpu-request-type',
    runAsUid = '--run-as-uid',
    runAsGid = '--run-as-gid',
    nodePools = '--node-pools',
    toleration = '--toleration',
    category = '--category',
    priority = '--priority',
    preemptibility = '--preemptibility',
    annotation = '--annotation',
    label = '--label',
    readinessProbeFailure = '--readiness-probe-failure',
    readinessProbeInitialDelay = '--readiness-probe-initial-delay',
    readinessProbePath = '--readiness-probe-path',
    readinessProbePeriod = '--readiness-probe-period',
    readinessProbePort = '--readiness-probe-port',
    readinessProbeSuccess = '--readiness-probe-success',
    readinessProbeTimeout = '--readiness-probe-timeout'
}

export enum NimScaleFlags {
    replicas = '--replicas',
    force = '--force'
}

export const ArrayFlags = [
    '--exclude-node',
    '-e',
    '--datasource',
    '--environment',
    '--datavolume',
    '--env-my-credentials',
    '--image-pull-my-credentials',
    '--node-pools',
    '--existing-pvc',
    '--new-pvc',
    '--nfs',
    '--git-sync',
    '--s3',
    '--label',
    '--port',
    '--external-url'
];
export const SubmitFlagsMerged = {
    ...SubmitFlags,
    ...excludeProperties(CommonWorkloadFlags, 'uuid')
};

export const NimSubmitFlagsMerged = {
    ...NimSubmitFlags,
    ...excludeProperties(CommonWorkloadFlags, 'uuid')
};

export enum SuspendFlags {}

export const SuspendFlagsMerged = {
    ...SuspendFlags,
    ...CommonWorkloadFlags
};

export enum ResumeFlags {}

export const ResumeFlagsMerged = {
    ...ResumeFlags,
    ...CommonWorkloadFlags
};

export const DeleteFlagsMerged = {
    ...DeleteFlags,
    ...CommonWorkloadFlags
};
export const portForwardFlags = {
    port: '--port',
    address: '--address',
    pod: '--pod'
};
export const describeFlags = {
    output: '--output',
    project: '-p'
};

export const applicationSubmitFlags = {
    project: '-p',
    chartUrl: '--chart-url',
    values: '--values',
    set: '--set',
    credential: '--credential'
};
export const execFlags = {
    command: '--',
    container: '--container',
    project: '--project',
    pod: '--pod'
};

export const bashFlags = {
    pod: '--pod',
    container: '--container',
    project: '--project',
    podRunningTimeout: '--pod-running-timeout',
    waitTimeout: '--wait-timeout',
    stdin: '--stdin',
    tty: '--tty'
};

export async function loginWithAccessKey(): Promise<void> {
    const accessKey = await accessKeyUtils.createAccessKey(
        flowRootState.users.system_administrator.accessToken,
        { name: objectGenerator.generateShortNameByPrefix('cli-v2') }
    );
    runaiCliV2.loginWithAccessKey(accessKey.body.clientId, accessKey.body.clientSecret);
}

export function loginWithUser(user: User): void {
    runaiCliV2.loginAsUser(user.email, user.password);
}

export function loginUser(user: User): void {
    runaiCliV2.loginAsUser(user.email, user.password);
}

export function cliLoginAsSystemAdmin(): void {
    loginUser(flowRootState.users.system_administrator);
}

export function ensureCommandLast(
    flags: Partial<typeof SubmitFlagsMerged>
): Partial<typeof SubmitFlagsMerged> {
    const entries = Object.entries(flags);
    const commandIndex = entries.findIndex(([key]) => key === SubmitFlags.command);
    if (commandIndex === -1 || commandIndex === entries.length - 1) {
        return flags;
    }
    const commandEntry = entries.splice(commandIndex, 1)[0];
    entries.push(commandEntry);
    return Object.fromEntries(entries) as Partial<typeof SubmitFlagsMerged>;
}
