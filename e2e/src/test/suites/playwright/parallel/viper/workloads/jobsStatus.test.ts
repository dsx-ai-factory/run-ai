/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    settings,
    tags,
    Utils,
    versionUtils,
    waitUtils,
    workloadsUtilsV2,
    workloadsCrudUtils,
    workloadUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { kubeCli } from '../../../../../../modules/cli';
import { createProject } from '../../../../../../modules/helpers/models/projects';
import { assetsApi, backendClusterApi } from '../../../../../../modules/api';
import { isMpiInstalled } from '../../../../../../modules/helpers/utils/Mpi';
import { isKubeflowInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { JobUtils } from '../../../../../../modules/helpers/utils/jobUtils';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadCreationRequest } from '../../../../../../modules/api/apis/assetsApi';
import {
    shouldRunTests,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { WorkloadType as WorkloadTypeV2 } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { test } from '@playwright/test';
import { DistributedFramework } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedFramework';
import { GpuRequestType } from '../../../../../../modules/api/apis/workloadsApi/generated/model/gpuRequestType';

test.describe(
    `Job status tests, tags: ${tags.resource.jobs} ${tags.teams.viper}`,
    { tag: [tags.services.agent, tags.services.clusterSync, tags.runOn.parallelCI] },
    function () {
        const defaultImagePath = config.quickStartImage.ubuntu;
        const sleepInfinity = 'sleep infinity';
        const sleepOneSec = 'sleep 1s';

        let project: Project;
        let projectWithTimeout: Project;
        let hasTrainTimeoutInCluster: boolean;

        test.beforeAll(async function () {
            backendClusterApi.jobsApi.accessToken = flowRootState.users.system_administrator.accessToken;
            const options = { maxAllowedGpus: -1, deservedGpus: 1 };
            project = await createProject('jobs-status', options);
            hasTrainTimeoutInCluster = versionUtils.isVersionSupported(
                flowRootState.cluster.version,
                SupportedVersions.V2_13
            );
            const optionsTimeout = {
                maxAllowedGpus: -1,
                deservedGpus: 1,
                trainingJobTimeLimitSecs: hasTrainTimeoutInCluster ? 10 : undefined,
                interactiveJobTimeLimitSecs: 10
            };
            projectWithTimeout = await createProject('jobs-status-timeout', optionsTimeout);
        });
        test.afterAll(async function () {
            try {
                if (project && project.id) {
                    logger.info(`deleting temporary project with id ${project.id}`);
                    await operations.deleteProjectAndNamespace(project.id);
                }
                if (projectWithTimeout && projectWithTimeout.id) {
                    logger.info(`deleting temporary project with id ${projectWithTimeout.id}`);
                    await operations.deleteProjectAndNamespace(projectWithTimeout.id);
                }
            } catch (error) {
                logger.error(`failed to delete project ${projectWithTimeout.name}`, error);
                return;
            }
            project = null;
        });
        test.afterEach(async function () {
            await workloadsUtilsV2.deleteWorkloadsByProjectId(String(project.id));
            await workloadsUtilsV2.deleteWorkloadsByProjectId(String(projectWithTimeout.id));
        });

        test.describe('training jobs', function () {
            test('training jobs - Workload should succeed', async function () {
                const { body: trainWorkload } = await workloadsCrudUtils.createTraining(
                    'succeeded-training',
                    false,
                    flowRootState.cluster.uuid,
                    String(project.id),
                    { image: defaultImagePath, command: sleepOneSec }
                );
                await waitUtils.waitForBackendWorkloadPhase(trainWorkload.name, project.id, Phase.Completed);
            });

            test('Workload should become running', async function () {
                const { body: trainWorkload } = await workloadsCrudUtils.createTraining(
                    'initialized-training',
                    false,
                    flowRootState.cluster.uuid,
                    String(project.id),
                    { image: defaultImagePath, command: sleepInfinity }
                );
                await workloadUtils.waitForWorkloadRunnerPhase(
                    trainWorkload.name,
                    Utils.namespaceFromProjectName(project.name),
                    Phase.Running
                );
            });

            test('Workload should be creating', async function () {
                assetsApi.trainingApi.accessToken = flowRootState.users.system_administrator.accessToken;
                const trainName = objectGenerator.generateNameByPrefix('created-training');
                const anyEnv = await environmentUtils.get('tensorboard');
                const trainWorkloadRequest: WorkloadCreationRequest = {
                    assets: {
                        environment: anyEnv.meta.id
                    },
                    clusterId: flowRootState.cluster.uuid,
                    name: trainName,
                    namespace: Utils.namespaceFromProjectName(project.name),
                    projectId: project.id
                };
                const trainingWorkload = await assetsApi.trainingApi.createTraining(trainWorkloadRequest);

                await waitUtils.waitForBackendWorkloadPhaseInHistory(
                    trainingWorkload.body.meta.id,
                    trainingWorkload.body.meta.name,
                    Phase.Creating
                );

                await workloadsUtilsV2.deleteWorkload(trainingWorkload.body.meta.id, WorkloadTypeV2.TRAINING);
            });
        });

        test.describe('hpo jobs', function () {
            test(`Workload should be running ${tags.runOn.parallelCI}`, async function () {
                const { body: trainWorkload } = await workloadsCrudUtils.createTraining(
                    'running-hpo',
                    false,
                    flowRootState.cluster.uuid,
                    String(project.id),
                    {
                        image: defaultImagePath,
                        command: sleepInfinity,
                        parallelism: 2,
                        completions: 3,
                        compute: {
                            gpuRequestType: GpuRequestType.Portion,
                            gpuDevicesRequest: 1,
                            gpuPortionRequest: 1
                        }
                    }
                );
                await waitUtils.waitForBackendWorkloadPhase(trainWorkload.name, project.id, Phase.Running);
            });
        });

        test.describe('interactive jobs', function () {
            test('Workload should be timeout (SUSPENDED)', async function () {
                const { body: interactiveWorkload } = await workloadsCrudUtils.createWorkspace(
                    'timeout-inter',
                    false,
                    flowRootState.cluster.uuid,
                    String(projectWithTimeout.id),
                    { image: defaultImagePath, command: sleepInfinity }
                );
                await Utils.delay(6000);
                await waitUtils.waitForBackendWorkloadPhase(
                    interactiveWorkload.name,
                    projectWithTimeout.id,
                    Phase.Stopped
                );
            });
        });

        test.describe('pytorch jobs', function () {
            test.beforeAll(async function () {
                if (!(await isKubeflowInstalled())) {
                    logger.warn('kubeflow is not installed on the cluster, skipping test');
                    this.skip();
                }
            });

            test('Workload should succeed', async function () {
                const { body: distributedWorkload } = await workloadsCrudUtils.createDistributed(
                    'succeeded-pytorch',
                    false,
                    flowRootState.cluster.uuid,
                    String(project.id),
                    {
                        distributedFramework: DistributedFramework.PyTorch,
                        numWorkers: 1,
                        command: sleepOneSec,
                        image: defaultImagePath
                    },
                    true
                );
                await waitUtils.waitForBackendWorkloadPhase(
                    distributedWorkload.name,
                    project.id,
                    Phase.Completed
                );
            });
        });

        test.describe('MPI jobs', function () {
            test.beforeAll(async function () {
                if (!(await isMpiInstalled())) {
                    logger.warn('Mpi operator doesnt exists on the cluster - skip');
                    this.skip();
                }
            });

            test('Workload should be pending', async function () {
                const { body: distributedWorkload } = await workloadsCrudUtils.createDistributed(
                    'pending-mpi',
                    false,
                    flowRootState.cluster.uuid,
                    String(project.id),
                    {
                        distributedFramework: DistributedFramework.Mpi,
                        numWorkers: 1,
                        command: sleepInfinity,
                        image: defaultImagePath,
                        compute: {
                            gpuRequestType: GpuRequestType.Portion,
                            gpuDevicesRequest: 100,
                            gpuPortionRequest: 1
                        }
                    },
                    true
                );
                await waitUtils.waitForBackendWorkloadPhase(
                    distributedWorkload.name,
                    project.id,
                    Phase.Pending
                );
            });
        });
    }
);

test.describe(`preemptive training job status, tags: ${tags.resource.jobs} ${tags.teams.viper}`, function () {
    let projectWithoutGpus: Project;
    let projectWithGpus: Project;
    let workloadInProjectWithGpus: string;
    let isEnvUnfitToTest: boolean;
    const gpus = 16;

    async function createTrainingWLThatUsesOverQuota(): Promise<string> {
        const workloadName = objectGenerator.generateNameByPrefix('training-workload');
        const { body: workload } = await workloadsCrudUtils.createTraining(
            workloadName,
            false,
            flowRootState.cluster.uuid,
            String(projectWithoutGpus.id),
            {
                image: config.quickStartImage.emptyImage,
                terminateAfterPreemption: true,
                compute: {
                    gpuRequestType: GpuRequestType.Portion,
                    gpuDevicesRequest: gpus,
                    gpuPortionRequest: 1
                }
            }
        );
        await waitUtils.waitForBackendWorkloadByIdAndPhase(workload.workloadId, Phase.Running);
        return workload.name;
    }

    const isOnlyOneOfTheJobsFailed = (): boolean => {
        const jobsInfo = JobUtils.getJobsInfo(projectWithoutGpus.name);
        logger.info(`Total number of jobs: ${jobsInfo.length}`);
        jobsInfo.forEach((job) => {
            logger.info(`Job Name: ${job.NAME}, Job Status: ${job.STATUS}, Project: ${job.PROJECT}`);
        });
        const isOneOfTheJobsFailed = jobsInfo.some((job) => job['STATUS'] === 'Failed');
        const isThreeJobsRunning = jobsInfo.filter((job) => job['STATUS'] === 'Running').length === 3;
        return isOneOfTheJobsFailed && isThreeJobsRunning;
    };

    test.beforeAll(async function () {
        const numOfWorkers = (await kubeCli.getNodesResources()).filter((node) =>
            node.name.includes('worker')
        ).length;
        isEnvUnfitToTest = settings.gpuType === 'real' || !settings.isSelfHosted || numOfWorkers !== 4;

        projectWithGpus = await createProject('16-gpus-quota', { maxAllowedGpus: -1, deservedGpus: gpus });

        projectWithoutGpus = await createProject('1-gpus-quota', { maxAllowedGpus: -1, deservedGpus: 1 });
    });
    test.afterAll(async function () {
        await operations.deleteProjectAndNamespace(projectWithoutGpus.id);
        await operations.deleteProjectAndNamespace(projectWithGpus.id);
    });
    test('terminate after preemption', function () {
        shouldRunTests(!isEnvUnfitToTest)(async function () {
            await createTrainingWLThatUsesOverQuota();
            await createTrainingWLThatUsesOverQuota();
            await createTrainingWLThatUsesOverQuota();
            await createTrainingWLThatUsesOverQuota();

            const workloadName = objectGenerator.generateNameByPrefix('training-workload');
            const { body: workload } = await workloadsCrudUtils.createTraining(
                workloadName,
                false,
                flowRootState.cluster.uuid,
                String(projectWithGpus.id),
                {
                    image: config.quickStartImage.emptyImage,
                    compute: {
                        gpuRequestType: GpuRequestType.Portion,
                        gpuDevicesRequest: gpus,
                        gpuPortionRequest: 1
                    }
                }
            );
            workloadInProjectWithGpus = workload.name;
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workload.workloadId, Phase.Running);
            await JobUtils.waitForJobStatus(projectWithGpus.name, workloadInProjectWithGpus, 'Running');

            await waitUtils.waitForCondition(
                isOnlyOneOfTheJobsFailed,
                `validate one of the trainings in project ${projectWithoutGpus.name} failed and one still running`,
                true,
                180
            );
        });
    });
});
