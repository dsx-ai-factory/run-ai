/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BaseWorkload } from '../models/baseWorkload';
import {
    cliUtil,
    flowRootState,
    InteractiveWorkload,
    logger,
    objectGenerator,
    operations,
    PodStatus,
    TrainingWorkload,
    versionUtils,
    waitUtils
} from '../index';
import { DistributedJobType, DistributedWorkload } from '../models/distributedWorkload';
import { expect } from 'chai';
import {bashCli, runaiCliV2} from '../../cli';
import { DisplayedJob } from '../../api/apis/backendClusterApi';
import { backendClusterApi } from '../../api';
export class JobUtils {
    static async createRunaiTypeJobs(
        projectOptions: { maxAllowedGpus: number; deservedGpus: number },
        namespace: string,
        imagePath: string,
        interactiveWorkloadName: string,
        interactivePreemptibleWorkloadName: string,
        trainWorkloadName: string
    ): Promise<{ [p: string]: BaseWorkload }> {
        const workloads: { [p: string]: BaseWorkload } = {};
        const jobNameSuffix = objectGenerator.generateNameByPrefix('-job');
        if (interactiveWorkloadName !== '') {
            this.createInteractiveWorkload(
                workloads,
                interactiveWorkloadName,
                namespace,
                jobNameSuffix,
                imagePath,
                false
            );
        }
        if (interactivePreemptibleWorkloadName !== '') {
            this.createInteractiveWorkload(
                workloads,
                interactivePreemptibleWorkloadName,
                namespace,
                jobNameSuffix,
                imagePath,
                true
            );
        }
        if (trainWorkloadName !== '') {
            this.createTrainingWorkload(workloads, trainWorkloadName, namespace, jobNameSuffix, imagePath);
        }

        if (flowRootState.numOfGpus < projectOptions.deservedGpus) {
            expect.fail('not enough GPUs in cluster to run this test');
        }

        await operations.submitWorkloads(Object.values(workloads), PodStatus.RUNNING);
        return workloads;
    }

    static createTrainingWorkload(
        workloads: {
            [p: string]: BaseWorkload;
        },
        trainWorkloadName: string,
        namespace: string,
        jobNameSuffix: string,
        image: string,
        command = 'sleep 1000000',
        gpu = '0.3'
    ): void {
        const tw = (workloads[trainWorkloadName] = new TrainingWorkload(
            image,
            namespace,
            trainWorkloadName + jobNameSuffix
        ));
        tw.gpu = gpu;
        tw.username = 'user-training';
        tw.name = trainWorkloadName + jobNameSuffix;
        tw.command = command;
    }

    static createInteractiveWorkload(
        workloads: {
            [p: string]: BaseWorkload;
        },
        interactiveWorkloadName: string,
        namespace: string,
        jobNameSuffix: string,
        image: string,
        preemptible: boolean,
        command = 'sleep 1000000',
        gpu = '0.2'
    ): void {
        const iw = (workloads[interactiveWorkloadName] = new InteractiveWorkload(
            image,
            namespace,
            interactiveWorkloadName + jobNameSuffix
        ));
        iw.gpu = gpu;
        iw.username = 'user-interactive';
        iw.name = interactiveWorkloadName + jobNameSuffix;
        if (preemptible) iw.name = 'interprejob';
        iw.command = command;
        iw.preemptible = preemptible;
    }

    static async createAllTypeJobs(
        projectOptions: { maxAllowedGpus: number; deservedGpus: number },
        namespace: string,
        image: string,
        interactiveWorkloadName: string,
        trainWorkloadName: string,
        mpiWorkloadName: string,
        pytorchWorkloadName: string,
        tensorWorkloadName: string,
        xgboostWorkloadName?: string,
        command?: string,
        expectedStatus = PodStatus.RUNNING,
        gpu?: string,
        NameSuffix = ''
    ): Promise<{ [p: string]: BaseWorkload }> {
        const workloads: { [p: string]: BaseWorkload } = {};
        const jobNameSuffix = NameSuffix || objectGenerator.generateNameByPrefix('-job');
        if (interactiveWorkloadName !== '') {
            this.createInteractiveWorkload(
                workloads,
                interactiveWorkloadName,
                namespace,
                jobNameSuffix,
                image,
                false,
                command,
                gpu
            );
        }
        if (trainWorkloadName !== '') {
            this.createTrainingWorkload(
                workloads,
                trainWorkloadName,
                namespace,
                jobNameSuffix,
                image,
                command,
                gpu
            );
        }
        if (mpiWorkloadName !== '') {
            this.createMpiWorkload(workloads, mpiWorkloadName, image, namespace, jobNameSuffix, command, gpu);
        }
        if (pytorchWorkloadName !== '') {
            this.createPytorchWorkload(
                workloads,
                pytorchWorkloadName,
                image,
                namespace,
                jobNameSuffix,
                command,
                gpu
            );
        }
        if (tensorWorkloadName !== '') {
            this.createTensorflowWorkload(
                workloads,
                tensorWorkloadName,
                image,
                namespace,
                jobNameSuffix,
                command,
                gpu
            );
        }
        if (xgboostWorkloadName !== '') {
            this.createXgboostWorkload(
                workloads,
                xgboostWorkloadName,
                image,
                namespace,
                jobNameSuffix,
                command,
                gpu
            );
        }

        if (flowRootState.numOfGpus < projectOptions.deservedGpus) {
            expect.fail('not enough GPUs in cluster to run this test');
        }

        await operations.submitWorkloads(Object.values(workloads), expectedStatus);
        return workloads;
    }

    static createMpiWorkload(
        workloads: {
            [p: string]: BaseWorkload;
        },
        mpiWorkloadName: string,
        image: string,
        namespace: string,
        jobNameSuffix: string,
        command?: string,
        gpu = '1'
    ): void {
        const mpiw = (workloads[mpiWorkloadName] = new DistributedWorkload(
            image,
            namespace,
            mpiWorkloadName + jobNameSuffix
        ));
        mpiw.gpu = gpu;
        mpiw.workers = 2;
        mpiw.jobType = DistributedJobType.MpiJobType;
        mpiw.username = 'user-mpi';
        mpiw.command = command;
        mpiw.name = mpiWorkloadName + jobNameSuffix;
    }

    private static createPytorchWorkload(
        workloads: {
            [p: string]: BaseWorkload;
        },
        pytorchWorkloadName: string,
        image: string,
        namespace: string,
        jobNameSuffix: string,
        command?: string,
        gpu = '0.2'
    ): void {
        const pytorchw = (workloads[pytorchWorkloadName] = new DistributedWorkload(
            image,
            namespace,
            pytorchWorkloadName + jobNameSuffix
        ));
        pytorchw.gpu = gpu;
        pytorchw.workers = 2;
        pytorchw.jobType = DistributedJobType.PyTorchJobType;
        pytorchw.username = 'user-pytorch';
        pytorchw.command = command;
        pytorchw.name = pytorchWorkloadName + jobNameSuffix;
    }

    private static createXgboostWorkload(
        workloads: {
            [p: string]: BaseWorkload;
        },
        xgboostWorkloadName: string,
        image: string,
        namespace: string,
        jobNameSuffix: string,
        command?: string,
        gpu = '0.2'
    ): void {
        const pytorchw = (workloads[xgboostWorkloadName] = new DistributedWorkload(
            image,
            namespace,
            xgboostWorkloadName + jobNameSuffix
        ));
        pytorchw.gpu = gpu;
        pytorchw.workers = 2;
        pytorchw.jobType = DistributedJobType.PyTorchJobType;
        pytorchw.username = 'user-xgboost';
        pytorchw.command = command;
        pytorchw.name = xgboostWorkloadName + jobNameSuffix;
    }

    static createTensorflowWorkload(
        workloads: {
            [p: string]: BaseWorkload;
        },
        tensorWorkloadName: string,
        image: string,
        namespace: string,
        jobNameSuffix: string,
        command?: string,
        gpu = '0.2'
    ): void {
        const tfw = (workloads[tensorWorkloadName] = new DistributedWorkload(
            image,
            namespace,
            tensorWorkloadName + jobNameSuffix
        ));
        tfw.gpu = gpu;
        tfw.workers = 2;
        tfw.jobType = DistributedJobType.TensorJobType;
        tfw.username = 'user-tf';
        tfw.command = command;
        tfw.name = tensorWorkloadName + jobNameSuffix;
    }

    static getJobsInfo(projectName: string): { [p: string]: string }[] {
        const listCmd1 = `${runaiCliV2.cliPath} list jobs -p ${projectName} `;
        const result = bashCli.exec(listCmd1);
        return cliUtil.parseListJobsTable(result, 1 /* column headers in 2nd line following info line */);
    }

    static validateJobStatus(
        projectName: string,
        jobName: string,
        desiredStatus: string
    ): () => Promise<boolean> {
        // bashCli.exec is implemented using spawn.sync therefore it is not possible to await for it,
        // but it is premised that the function will return only after executing the CLI command.
        // eslint-disable-next-line @typescript-eslint/require-await
        return async (): Promise<boolean> => {
            try {
                logger.debug(`Validating status of ${jobName} is ${desiredStatus}`);
                const table = this.getJobsInfo(projectName);

                // filter the job we're looking for
                const jobInfo = table.filter((job) => {
                    return job['NAME'] === jobName;
                });
                expect(jobInfo.length).to.equal(1);

                if (jobInfo[0]['STATUS'] === desiredStatus) {
                    logger.debug(`Job ${jobName} status is ${desiredStatus}`);
                    return true;
                }
            } catch (e) {
                logger.debug(
                    `Failed validating that status of job ${jobName} in project ${projectName}, error: ${e}`
                );
                return false;
            }
            return false;
        };
    }

    static async waitForJobStatus(projectName: string, jobName: string, status: string): Promise<void> {
        await waitUtils.waitForCondition(
            JobUtils.validateJobStatus(projectName, jobName, status),
            `wait for job ${jobName} in project ${projectName} to become ${status}`,
            true,
            90
        );
    }

    async waitForBackendJobInProject(
        jobName: string,
        expectedProjectName: string,
        timeout = 240
    ): Promise<DisplayedJob> {
        let jobsProject: string;
        let backendJob: DisplayedJob;

        const isBackendJobWithCorrectProject = async (): Promise<boolean> => {
            try {
                backendJob = await waitUtils.waitForBackendJob(jobName, `text:${expectedProjectName}`, 20, 5);
                jobsProject = backendJob.project;
                return jobsProject === expectedProjectName;
            } catch (e) {
                logger.warn(`Error getting backend job, jobName: ${jobName}, error: ${e}`);
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackendJobWithCorrectProject,
            `waitForBackendJobInProject jobName: ${jobName}, expectedProjectName: ${expectedProjectName}, jobs project was: ${jobsProject}`,
            true,
            timeout,
            5
        );

        logger.debug(`waitForBackendJobInProject succeeded for job ${jobName}, project: ${jobsProject}`);
        return backendJob;
    }

    async waitForBackendPodsOfJobInNamespace(
        backendJob: DisplayedJob,
        expectedNamespace: string,
        timeout = 60
    ): Promise<void> {
        let podsNamespace: string;

        const isBackendPodsOfJobWithCorrectNamespace = async (): Promise<boolean> => {
            try {
                const backendJobPods = (
                    await backendClusterApi.jobsApi.getPodsByJobId(
                        backendJob.podGroupId,
                        flowRootState.cluster.uuid
                    )
                ).body;
                for (const pod of backendJobPods) {
                    podsNamespace = pod.namespace;
                    if (podsNamespace !== expectedNamespace) {
                        return false;
                    }
                }
                return true;
            } catch (e) {
                logger.warn(
                    `Error getting backend pods by podGroupId: ${backendJob.podGroupId}, error: ${e}`
                );
                return false;
            }
        };

        await waitUtils.waitForCondition(
            isBackendPodsOfJobWithCorrectNamespace,
            `waitForBackendPodsOfJobInNamespace jobName: ${backendJob.jobName}, expectedNamespace: ${expectedNamespace}, pods namespace was: ${podsNamespace}`,
            true,
            timeout,
            5
        );

        logger.debug(
            `waitForBackendPodsOfJobInNamespace succeeded for pods in job ${backendJob.jobName}, namespace: ${podsNamespace}`
        );
    }

    async waitForBackendJobAndPodsWithCorrectProjectAndNamespace(
        jobName: string,
        expectedProjectName: string,
        expectedNamespace: string,
        timeout = 60
    ): Promise<void> {
        const backendJob = await this.waitForBackendJobInProject(jobName, expectedProjectName, timeout);

        await waitUtils.waitForPods(jobName, expectedNamespace, 1, waitUtils.waitForPodTimeout, ['Running']);

        if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.11')) {
            await this.waitForBackendPodsOfJobInNamespace(backendJob, expectedNamespace, timeout);
        }
    }
}
