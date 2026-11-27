/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../api/apis/projectsDepartmentsApi';
import { TrainingWorkload } from './trainingWorkload';
import {
    config,
    InteractiveWorkload,
    logger,
    objectGenerator,
    PodStatus,
    Utils,
    waitUtils,
    workloadUtils
} from '../index';
import { DistributedJobType, DistributedWorkload } from './distributedWorkload';
import { extractGpuGroupIndex } from './pods';

export async function submitTrainWorkload(
    workloadPrefix: string,
    projectName: string,
    imagePath = config.quickStartImage.ubuntu,
    command = 'sleep infinity',
    numGpus = 1,
    podStatus = PodStatus.RUNNING,
    nodePoolName?: string,
    cpu?: string,
    memory?: string,
    cpuLimit?: string,
    memoryLimit?: string,
    username?: string
): Promise<TrainingWorkload> {
    const jobName = objectGenerator.generateNameByPrefix(workloadPrefix);
    const testWorkload = new TrainingWorkload(
        imagePath,
        Utils.namespaceFromProjectName(projectName),
        jobName
    );
    testWorkload.gpu = String(numGpus);
    testWorkload.command = command;

    if (nodePoolName) {
        testWorkload.nodePools = nodePoolName;
    }
    if (cpu) {
        testWorkload.cpu = cpu;
    }
    if (memory) {
        testWorkload.memory = memory;
    }
    if (cpuLimit) {
        testWorkload.cpuLimit = cpuLimit;
    }
    if (memoryLimit) {
        testWorkload.memoryLimit = memoryLimit;
    }
    if (username) {
        testWorkload.username = username;
        testWorkload.userId = username;
    }

    logger.info(`submitting workload ${jobName} in project ${projectName}`);
    await workloadUtils.submit(testWorkload);
    await waitUtils.waitForWorkloadStatus(testWorkload, podStatus);
    return testWorkload;
}

// Submit `count` GPU-fraction training workloads to a single node pool and return, for each, the
// physical GPU index (RUNAI-VISIBLE-DEVICES) its pod landed on. The workloads are submitted together
// so the scheduler places the fractions concurrently, and the indices are then resolved together.
export async function submitFractionWorkloadsAndGetGpuIndices(
    projectName: string,
    nodePoolName: string,
    count: number,
    gpuFraction: number,
    imagePath = config.quickStartImage.ubuntu
): Promise<{ workloads: TrainingWorkload[]; gpuIndices: string[] }> {
    const namespace = Utils.namespaceFromProjectName(projectName);
    const workloads = await Promise.all(
        Array.from({ length: count }, () =>
            submitTrainWorkload(
                'gpu-frac',
                projectName,
                imagePath,
                'sleep infinity',
                gpuFraction,
                PodStatus.RUNNING,
                nodePoolName
            )
        )
    );
    const gpuIndices = await Promise.all(
        workloads.map(async (workload) => {
            const pods = await waitUtils.waitForPods(
                workload.name,
                namespace,
                1,
                waitUtils.waitForPodTimeout,
                ['Running']
            );
            return extractGpuGroupIndex(pods[0]);
        })
    );
    return { workloads, gpuIndices };
}

export async function submitInteractiveWorkload(
    workloadPrefix: string,
    projectName: string,
    imagePath: string,
    command: string,
    numGpus = 1,
    podStatus = PodStatus.RUNNING
): Promise<InteractiveWorkload> {
    const jobName = objectGenerator.generateNameByPrefix(workloadPrefix);
    const testWorkload = new InteractiveWorkload(
        imagePath,
        Utils.namespaceFromProjectName(projectName),
        jobName
    );
    testWorkload.gpu = String(numGpus);
    testWorkload.command = command;

    logger.info(`submitting interactive workload ${jobName} in project ${projectName}`);
    await workloadUtils.submit(testWorkload);
    await waitUtils.waitForWorkloadStatus(testWorkload, podStatus);
    return testWorkload;
}

export async function submitDistributedWorkload(
    jobType: string,
    workloadPrefix: string,
    project: Project,
    command: string,
    imagePath: string,
    numGpus: string,
    workers: number
): Promise<DistributedWorkload> {
    const jobName = objectGenerator.generateNameByPrefix(workloadPrefix);
    const namespace = Utils.namespaceFromProjectName(project.name);
    const testWorkload = new DistributedWorkload(imagePath, namespace, jobName);
    testWorkload.imagePullPolicy = 'IfNotPresent';
    testWorkload.gpu = numGpus;
    testWorkload.preemptible = true;
    testWorkload.workers = workers;
    testWorkload.jobType = jobType;
    testWorkload.command = command;

    logger.info(`submitting ${jobType} workload ${jobName} in project ${project.name}`);
    await workloadUtils.submit(testWorkload);
    await waitUtils.waitForPods(jobName, namespace, workers);
    return testWorkload;
}

export async function submitPytorchWorkload(
    workloadPrefix: string,
    project: Project,
    command: string,
    imagePath = config.quickStartImage.distributedPytorch,
    numGpus = '1',
    workers = 2
): Promise<DistributedWorkload> {
    return submitDistributedWorkload(
        DistributedJobType.PyTorchJobType,
        workloadPrefix,
        project,
        command,
        imagePath,
        numGpus,
        workers
    );
}

export async function submitMpiWorkload(
    workloadPrefix: string,
    project: Project,
    command: string,
    imagePath = config.quickStartImage.distributedPytorch,
    numGpus = '1',
    workers = 2
): Promise<DistributedWorkload> {
    return submitDistributedWorkload(
        DistributedJobType.MpiJobType,
        workloadPrefix,
        project,
        command,
        imagePath,
        numGpus,
        workers
    );
}

export async function submitHpoWorkload(
    workloadPrefix: string,
    project: Project,
    imagePath: string,
    command: string,
    numGpus = 1,
    podStatus = PodStatus.RUNNING,
    parallelism = 2,
    completions = 3
): Promise<TrainingWorkload> {
    const jobName = objectGenerator.generateNameByPrefix(workloadPrefix);
    const testWorkload = new TrainingWorkload(
        imagePath,
        Utils.namespaceFromProjectName(project.name),
        jobName
    );
    testWorkload.gpu = String(numGpus);
    testWorkload.command = command;
    testWorkload.parallelism = parallelism;
    testWorkload.completions = completions;

    logger.info(`submitting workload ${jobName} in project ${project.name}`);
    await workloadUtils.submit(testWorkload);
    await waitUtils.waitForWorkloadStatus(testWorkload, podStatus);
    return testWorkload;
}

export async function submitWorkload(
    workloadPrefix: string,
    project: Project,
    imagePath: string,
    command: string,
    numGpus = 1,
    podStatus = PodStatus.RUNNING
): Promise<TrainingWorkload> {
    return await submitTrainWorkload(workloadPrefix, project.name, imagePath, command, numGpus, podStatus);
}
