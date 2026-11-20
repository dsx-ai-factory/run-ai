/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    config,
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    flowRootState,
    workloadsCrudUtils,
    metricUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { WorkloadType as WorkloadTypeFromWorkloadsCrudUtils } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { WorkloadMetricType } from '../../../../../../modules/api/apis/workloadsApi/generated/model/workloadMetricType';
import { WorkloadDetailsDrawer } from '../../../../../../modules/ui/runaiV2/pages/WorkloadDetailsDrawer';
import { workloadsService } from '../../../../../../modules/api';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { DistributedFramework } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedFramework';
import { PodMetricType } from '../../../../../../modules/api/apis/workloadsApi/generated/model/podMetricType';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';

test.describe(`Workload info check ${tags.tool.ui} ${tags.resource.workloads} ${tags.runOn.dgxc} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_18)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let browserEngine: BrowserEngine;
        let authPage: AuthPage;
        let workloadPage: WorkloadPage;
        let videoName: string;
        let testWorkloads: any[];

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('project-workload-info');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 8,
                maxAllowedGpus: 8
            });

            workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;
            workloadsService.workloadsApi.accessToken =
                flowRootState.users.l1_researcher.accessToken;

            browserEngine = new BrowserEngine();

            // Constants for workload commands and args
            const tensorboardCommand = 'tensorboard';
            const tensorboardArgs =
                '--logdir ./ --host 0.0.0.0 --path_prefix /${RUNAI_PROJECT}/${RUNAI_JOB_NAME}';
            const bashCommand = '/bin/bash -c';
            const bashArgs = '\'echo "Hello, World!" && sleep 300\'';

            testWorkloads = [
                {
                    name: 'sn-training',
                    options: {
                        image: config.quickStartImage.distributedTf,
                        command: tensorboardCommand,
                        args: tensorboardArgs,
                        compute: { gpuDevicesRequest: 1 }
                    },
                    logMatch: 'TensorBoard'
                },
                {
                    name: 'dist-training',
                    options: {
                        image: config.quickStartImage.ubuntu,
                        command: bashCommand,
                        args: bashArgs,
                        compute: { gpuDevicesRequest: 1 },
                        distributedFramework: DistributedFramework.Mpi
                    },
                    logMatch: 'Hello, World!'
                }
            ];
        });

        test.afterAll('deleting project and namespace', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`);
                throw error;
            }
        });

        // Helper function to create workload based on configuration
        async function createWorkload(workloadConfig: any): Promise<any> {
            const workloadName = objectGenerator.generateNameByPrefix(workloadConfig.name);
            let createdTraining: any;

            if (workloadConfig.name.startsWith('sn')) {
                // Single node training
                const { body: training } = await workloadsCrudUtils.createTraining(
                    workloadName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    workloadConfig.options
                );
                createdTraining = training;

                await workloadsCrudUtils.waitForWorkloadPhase(
                    training.workloadId,
                    Phase.Running,
                    WorkloadTypeFromWorkloadsCrudUtils.TRAINING,
                    undefined,
                    120
                );
            } else if (workloadConfig.name.startsWith('dist')) {
                // Distributed training
                const { body: training } = await workloadsCrudUtils.createDistributedV2(
                    workloadName,
                    false,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    workloadConfig.options
                );
                createdTraining = training;

                await workloadsCrudUtils.waitForWorkloadPhase(
                    training.workloadId,
                    Phase.Running,
                    WorkloadTypeFromWorkloadsCrudUtils.DISTRIBUTED,
                    undefined,
                    240
                );
            }
            return createdTraining;
        }

        test.beforeEach(async function () {
            await browserEngine.start();
            authPage = new AuthPage(browserEngine.currentPage);
            workloadPage = new WorkloadPage(browserEngine.currentPage);

            await authPage.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        // Coverage remains disabled pending revalidation of this case
        test.skip('T3738315: Verify a L1 researcher can check job info: Verify a L1 researcher can check job info', async function () {
            test.setTimeout(300 * 1000);
            videoName = 'tenant-user-job-info-check';
            let workloadDetailsDrawer: WorkloadDetailsDrawer;

            for (const workloadConfig of testWorkloads) {
                let createdTraining: any;

                await test.step(`Create ${workloadConfig.name}`, async () => {
                    createdTraining = await createWorkload(workloadConfig);
                });

                await test.step(`Verify ${workloadConfig.name} event history`, async () => {
                    await waitUtils.waitForCondition(
                        async () => {
                            const eventHistory = await workloadsUtilsV2.getWorkloadHistory(
                                createdTraining.workloadId
                            );
                            const responseContent = eventHistory.body;
                            const responseString = JSON.stringify(responseContent);
                            return (
                                responseString.includes(createdTraining.name) &&
                                responseString.includes('Running')
                            );
                        },
                        `Wait for event history to contain workload name and Running status for ${createdTraining.workloadId}`,
                        true,
                        30,
                        5
                    );
                });

                await test.step(`Verify ${workloadConfig.name} logs`, async () => {
                    workloadDetailsDrawer = new WorkloadDetailsDrawer(browserEngine.currentPage);
                    await workloadDetailsDrawer.clickShowDetailsBtn(createdTraining.name);
                    await workloadDetailsDrawer.clickOnLogsTab();
                    await workloadDetailsDrawer.waitForLogContentToAppear();
                    const logs = await workloadDetailsDrawer.getLogContent();
                    expect(logs).toContain(workloadConfig.logMatch);
                });

                await workloadsCrudUtils.deleteTraining(createdTraining.workloadId);
            }
        });

        test('T3738312: Verify a L1 researcher can view the job telemetry', async function () {
            test.setTimeout(300 * 1000);
            for (const workloadConfig of testWorkloads) {
                let createdTraining: any;
                await test.step(`Create ${workloadConfig.name} and wait for it to be running`, async () => {
                    createdTraining = await createWorkload(workloadConfig);
                });
                const now = new Date();
                const start = new Date(now.getTime() - 2 * 60 * 1000);
                const end = new Date(now.getTime() + 2 * 60 * 1000);

                await test.step(`Verify ${workloadConfig.name} metrics`, async () => {
                    const workloadMetrics: WorkloadMetricType[] = [
                        WorkloadMetricType.GpuUtilization,
                        WorkloadMetricType.CpuUsageCores,
                        WorkloadMetricType.GpuMemoryUsageBytes,
                        WorkloadMetricType.CpuMemoryUsageBytes
                    ];

                    await waitUtils.waitForCondition(
                        async () => {
                            const {
                                body: { measurements }
                            } = await metricUtils.getWorkloadMetrics(
                                createdTraining.workloadId,
                                workloadMetrics,
                                start,
                                end
                            );
                            return metricUtils.validateMetrics(measurements);
                        },
                        `Wait for metrics data for workload ${createdTraining.workloadId}`,
                        true,
                        180,
                        10
                    );
                });

                await test.step(`Verify ${workloadConfig.name} pod metrics`, async () => {
                    const podMetrics: PodMetricType[] = [
                        PodMetricType.GpuUtilizationPerGpu,
                        PodMetricType.GpuUtilization,
                        PodMetricType.GpuMemoryUsageBytesPerGpu,
                        PodMetricType.CpuUsageCores
                    ];

                    const {
                        body: { pods }
                    } = await workloadsUtilsV2.getWorkloadPods(createdTraining.workloadId);
                    // Filter out launcher pod for distributed workloads
                    const workerPods = workloadConfig.name.startsWith('dist')
                        ? pods.filter((pod) => !pod.name.includes('launcher'))
                        : pods;
                    expect(workerPods.length).toBeGreaterThan(0);

                    await Promise.all(
                        workerPods.map(async (pod) => {
                            await waitUtils.waitForCondition(
                                async () => {
                                    const {
                                        body: { measurements }
                                    } = await metricUtils.getPodMetrics(
                                        createdTraining.workloadId,
                                        pod.id,
                                        podMetrics,
                                        start,
                                        end
                                    );
                                    return metricUtils.validateMetrics(measurements);
                                },
                                `Wait for metrics data for pod ${pod.id}`,
                                true,
                                180,
                                10
                            );
                        })
                    );
                });

                await workloadsCrudUtils.deleteTraining(createdTraining.workloadId);
            }
        });
    });
});
