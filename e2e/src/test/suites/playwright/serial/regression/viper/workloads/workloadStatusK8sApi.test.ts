/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    crds,
    InteractiveWorkload,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    TrainingWorkload,
    Utils,
    waitUtils,
    allureUtils,
    settings
} from '../../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../../modules/cli';
import {
    expectCompleteSuccessfully,
    expectFailedWorkload,
    expectPendingWorkload,
    expectRunningPodsForWorkload,
    expectRunningWorkload,
    expectSuspendedWorkload,
    isPendingPodsZeroOrUndefined
} from '../../../../../../../modules/helpers/utils/testing/WorkloadStateTestUtils';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import { shouldRunNativeWorkloadCrdTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

import { test, expect } from '@playwright/test';

shouldRunNativeWorkloadCrdTests(() => {
test.describe(`workload status api tests:  ${tags.tool.api} ${tags.resource.jobs} ${tags.teams.viper}`, function () {
    let projectName: string, namespace: string, jobName: string;
    test.beforeAll(async function () {
        try {
            await operations.forceCleanup();
        } catch (error) {
            logger.error(error);
        }
        projectName = objectGenerator.generateNameByPrefix('project');
        namespace = Utils.namespaceFromProjectName(projectName);
        await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 2,
            maxAllowedGpus: 2
        });
    });
    test.afterAll('force cleanup', async function () {
        try {
            await operations.forceCleanup();
        } catch (error) {
            logger.error('Force Cleanup failed', error);
        }
    });

    test.describe('interactive workload', function () {
        test.describe('sanity  test', function () {
            let workload: InteractiveWorkload;

            test.beforeEach(function () {
                jobName = objectGenerator.generateNameByPrefix('job');

                workload = new InteractiveWorkload(config.quickStartImage.ubuntu, namespace, jobName);
                workload.cpu = '1';
                workload.createHomeDir = true;
                workload.gpu = '0.5';
                workload.imagePullPolicy = 'Always';
                workload.memory = '100M';
                if (!isOCP()) {
                    workload.runAsGid = 20;
                    workload.runAsUid = 501;
                }
                workload.runAsUser = true;
                workload.stdin = true;
                workload.supplementalGroups = '20;12;61;79;80;81;98;33;100;204;250;395;101;399;400;701';
                workload.tty = true;
                workload.username = 'ofer.lahav';
                workload.workingDir = '/home/test';
            });

            test.afterEach(async function () {
                try {
                    await operations.deleteWorkloads([workload]);
                } catch (error) {
                    logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
                }
            });

            test('Should create a valid job and validate its running', async function () {
                workload.image = settings.isArm
                    ? config.quickStartImage.ubuntu
                    : config.quickStartImage.sleepInfinity;
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForJobToBeRunning(
                    workload.name,
                    namespace,
                    crds.runaiInteractiveWorkloadCrd.kind
                );
                const interactiveWorkload = await kubeCli.getInteractiveWorkload(jobName, namespace);
                expectRunningWorkload(interactiveWorkload);
            });
            test('should create a valid job and validate it completes successfully', async function () {
                workload.command = 'sleep 1';
                workload.completions = 1;
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForJobToBeFinished(
                    jobName,
                    namespace,
                    crds.runaiInteractiveWorkloadCrd.kind
                );

                const interactiveWorkload = await kubeCli.getInteractiveWorkload(jobName, namespace);
                expectCompleteSuccessfully(interactiveWorkload);
            });
            test('job should be in pending, job has to many GPUs', async function () {
                workload.gpu = '100';
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForJobToBePending(
                    jobName,
                    namespace,
                    crds.runaiInteractiveWorkloadCrd.kind
                );
                const interactiveWorkload = await kubeCli.getInteractiveWorkload(jobName, namespace);
                expectPendingWorkload(interactiveWorkload);
            });
            test('should create a job with invalid exit code and validate it failed', async function () {
                workload.command = '/bin/sh -c `exit 1`';
                workload.backoffLimit = 1;
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForJobToBeFinished(
                    jobName,
                    namespace,
                    crds.runaiInteractiveWorkloadCrd.kind
                );
                const interactiveWorkload = await kubeCli.getInteractiveWorkload(jobName, namespace);
                expectFailedWorkload(interactiveWorkload);
            });
            test.describe('interactive job is already running', function () {
                test.beforeEach(async function () {
                    workload.image = settings.isArm
                        ? config.quickStartImage.ubuntu
                        : config.quickStartImage.sleepInfinity;
                    await kubeCli.createWorkload(workload);
                    await waitUtils.waitForJobToBeRunning(
                        workload.name,
                        namespace,
                        crds.runaiInteractiveWorkloadCrd.kind
                    );
                });
                test('should suspend the job', async function () {
                    workload.active = false;
                    await kubeCli.updateWorkload(workload);
                    await waitUtils.waitForJobToBeSuspend(
                        jobName,
                        namespace,
                        crds.runaiInteractiveWorkloadCrd.kind
                    );
                    const interactiveWorkload = await kubeCli.getInteractiveWorkload(jobName, namespace);

                    expectSuspendedWorkload(interactiveWorkload);
                });
                test('should resume the job', async function () {
                    workload.active = false;
                    await kubeCli.updateWorkload(workload);
                    await waitUtils.waitForJobToBeSuspend(
                        jobName,
                        namespace,
                        crds.runaiInteractiveWorkloadCrd.kind
                    );

                    workload.active = true;
                    await kubeCli.updateWorkload(workload);
                    await waitUtils.waitForJobToBeRunning(
                        jobName,
                        namespace,
                        crds.runaiInteractiveWorkloadCrd.kind
                    );

                    const interactiveWorkload = await kubeCli.getInteractiveWorkload(jobName, namespace);
                    expectRunningWorkload(interactiveWorkload);
                });
            });
        });
    });
    test.describe('training workload', function () {
        test.describe('sanity test  ', function () {
            let workload: TrainingWorkload;

            test.beforeEach(function () {
                jobName = objectGenerator.generateNameByPrefix('job');
                workload = new TrainingWorkload(config.quickStartImage.ubuntu, namespace, jobName);
                workload.cpu = '1';
                workload.createHomeDir = true;
                workload.gpu = '0.5';
                workload.imagePullPolicy = 'Always';
                workload.memory = '100M';
                if (!isOCP()) {
                    workload.runAsGid = 20;
                    workload.runAsUid = 501;
                }
                workload.runAsUser = true;
                workload.stdin = true;
                workload.supplementalGroups = '20;12;61;79;80;81;98;33;100;204;250;395;101;399;400;701';
                workload.tty = true;
                workload.username = 'ofer.lahav';
                workload.workingDir = '/home/test';
            });
            test.afterEach(async function () {
                try {
                    await operations.deleteWorkloads([workload]);
                } catch (error) {
                    logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
                }
            });

            test('Should create a valid job and validate its running', async function () {
                workload.image = settings.isArm
                    ? config.quickStartImage.ubuntu
                    : config.quickStartImage.sleepInfinity;
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForJobToBeRunning(workload.name, namespace, workload.getCrd().kind);
                const trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
                expectRunningWorkload(trainWorkload);
            });
            test('should create a valid job and validate it completes successfully', async function () {
                workload.command = 'sleep 1';
                workload.completions = 1;
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForJobToBeFinished(jobName, namespace, workload.getCrd().kind);

                const trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
                expectCompleteSuccessfully(trainWorkload);
            });
            test('job should be in pending, job is has to many GPUs', async function () {
                workload.gpu = '100';
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForJobToBePending(jobName, namespace, workload.getCrd().kind);
                await waitUtils.sleep(1000);
                const trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
                expectPendingWorkload(trainWorkload);
            });
            test('should create a job with failed command and validate it failed', async function () {
                workload.command = '/bin/sh -c `exit 1`';
                workload.backoffLimit = 1;
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForJobToBeFinished(jobName, namespace, workload.getCrd().kind);
                const trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
                expectFailedWorkload(trainWorkload);
            });
            test.describe('training job is already running', function () {
                test.beforeEach(async function () {
                    workload.image = settings.isArm
                        ? config.quickStartImage.ubuntu
                        : config.quickStartImage.sleepInfinity;
                    await kubeCli.createWorkload(workload);
                        await waitUtils.waitForJobToBeRunning(
                            workload.name,
                            namespace,
                            workload.getCrd().kind
                        );
                });
                test('should suspend the job', async function () {
                    workload.active = false;
                    await kubeCli.updateWorkload(workload);
                    await waitUtils.waitForJobToBeSuspend(jobName, namespace, workload.getCrd().kind);

                    const trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);

                    expectSuspendedWorkload(trainWorkload);
                });
                test('should resume the job', async function () {
                    workload.active = false;
                    await kubeCli.updateWorkload(workload);
                    await waitUtils.waitForJobToBeSuspend(jobName, namespace, workload.getCrd().kind);

                    workload.active = true;
                    await kubeCli.updateWorkload(workload);
                    await waitUtils.waitForJobToBeRunning(jobName, namespace, workload.getCrd().kind);

                    const trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
                    expectRunningWorkload(trainWorkload);
                });
            });
        });
        test.describe('hpo workload - cluster level metrics behavior', function () {
            let workload: TrainingWorkload;

            test.beforeEach(function () {
                jobName = objectGenerator.generateNameByPrefix('job');
                workload = new TrainingWorkload(config.quickStartImage.ubuntu, namespace, jobName);
                workload.cpu = '0';
                workload.createHomeDir = true;
                workload.gpu = '0.5';
                workload.imagePullPolicy = 'Always';
                workload.memory = '100M';
                if (!isOCP()) {
                    workload.runAsGid = 20;
                    workload.runAsUid = 501;
                }
                workload.runAsUser = true;
                workload.stdin = true;
                workload.supplementalGroups = '20;12;61;79;80;81;98;33;100;204;250;395;101;399;400;701';
                workload.tty = true;
                workload.username = 'user1';
                workload.workingDir = '/home/test';
                workload.parallelism = 2;
                workload.completions = 3;
            });
            test.afterEach(async function () {
                try {
                    await operations.deleteWorkloads([workload]);
                } catch (error) {
                    logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
                }
            });

            test('Should create a valid job and validate its running', async function () {
                workload.image = settings.isArm
                    ? config.quickStartImage.ubuntu
                    : config.quickStartImage.sleepInfinity;
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForJobToBeRunning(workload.name, namespace, workload.getCrd().kind);
                let trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
                if (trainWorkload['status']['runningPods'] < workload.parallelism) {
                    await waitUtils.waitForPods(
                        jobName,
                        namespace,
                        workload.parallelism,
                        waitUtils.waitForPodTimeout,
                        ['Running']
                    );
                    trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
                }
                expectRunningPodsForWorkload(trainWorkload, workload.parallelism);
            });
            test('should create a valid job and validate it completes successfully', async function () {
                workload.command = 'sleep 1';
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForJobToBeFinished(jobName, namespace, workload.getCrd().kind);
                const trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
                expectCompleteSuccessfully(trainWorkload, workload.completions);
            });
            test('should create a job with failed command and validate it failed', async function () {
                workload.command = '/bin/sh -c `exit 1`';
                workload.backoffLimit = 1;
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForJobToBeFinished(jobName, namespace, workload.getCrd().kind);
                const trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
                await waitUtils.waitForCondition(
                    () => isPendingPodsZeroOrUndefined(namespace, jobName),
                    'wait for pending pods to be updated',
                    true,
                    30
                );
                expectFailedWorkload(trainWorkload);
            });

            test.describe('training job is already running', function () {
                test.beforeEach(async function () {
                    workload.image = settings.isArm
                        ? config.quickStartImage.ubuntu
                        : config.quickStartImage.sleepInfinity;
                    await kubeCli.createWorkload(workload);
                        await waitUtils.waitForJobToBeRunning(
                            workload.name,
                            namespace,
                            workload.getCrd().kind
                        );
                });
                test('should suspend the job', async function () {
                    workload.active = false;
                    const updatedWorkload = await kubeCli.updateWorkload(workload);
                    expect(updatedWorkload['spec'].active.value).toBeFalsy();
                        await waitUtils.waitForJobToBeSuspend(
                            jobName,
                            namespace,
                            workload.getCrd().kind,
                            150
                        );
                    const trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
                    expectSuspendedWorkload(trainWorkload);
                });
                test('should resume the job', async function () {
                    workload.active = false;
                    await kubeCli.updateWorkload(workload);
                    await waitUtils.waitForJobToBeSuspend(jobName, namespace, workload.getCrd().kind);

                    workload.active = true;
                    await kubeCli.updateWorkload(workload);
                    await waitUtils.waitForJobToBeRunning(jobName, namespace, workload.getCrd().kind);
                    await waitUtils.waitForPods(jobName, namespace, 2, 60, ['Running']);

                    const trainWorkload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
                    expectRunningPodsForWorkload(trainWorkload, workload.parallelism);
                });
            });
        });
    });
});
});
