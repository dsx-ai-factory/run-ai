/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { workloadsService } from '../../../../../../modules/api';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { isLwsInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    createInference,
    createNimMetricsFixtureSpec,
    createDistributedInference
} from '../../../parallel/viper/inference/inferenceUtils.js';
import { InferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceSpecSpec';
import { DistributedInferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceSpecSpec';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { WorkloadDetailsDrawer } from '../../../../../../modules/ui/runaiV2/pages/WorkloadDetailsDrawer';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { test } from '@playwright/test';

test.describe(
    `NIM metrics tests, tags: ${tags.resource.inferenceWorkload} ${tags.teams.loong} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        supportedFrom(SupportedVersions.V2_24)(function () {
            let browserEngine: BrowserEngine;
            let authPage: AuthPage;
            let workloadPage: WorkloadPage;
            let project: Project;
            let singleReplicaWorkloadId: string;
            let singleReplicaWorkloadName: string;
            let singleReplicaWorkloadSpec: InferenceSpecSpec;
            let twoReplicaWorkloadId: string;
            let twoReplicaWorkloadName: string;
            let twoReplicaWorkloadSpec: InferenceSpecSpec;
            let distributedworkloadId: string;
            let distributedworkloadName: string;
            let distributedworkloadSpec: DistributedInferenceSpecSpec;
            let trainingId: string;
            let trainingJobName: string;
            let detailsDrawer: WorkloadDetailsDrawer;
            let videoName: string;

            const NAME_PREFIX = {
                SINGLE_REPLICA: 'singlereplica',
                TWO_REPLICA: 'tworeplica'
            } as const;

            let nimFixtureSpec: Pick<InferenceSpecSpec, 'image' | 'command' | 'args' | 'compute'>;

            const LABEL = {
                NIM_WORKLOAD: 'run.ai/nim-workload'
            } as const;

            test.beforeAll(async function () {
                if (!(await isKnativeInstalled())) {
                    logger.warn('knative is not installed on the cluster, skipping test');
                    test.skip();
                }
                nimFixtureSpec = createNimMetricsFixtureSpec();
                project = await projectResourcesUtils.createProjectAndNamespace('project', null, 8, 8);
                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
                workloadsService.workloadsApi.accessToken =
                    flowRootState.users.system_administrator.accessToken;

                singleReplicaWorkloadName = objectGenerator.generateNameByPrefix(NAME_PREFIX.SINGLE_REPLICA);
                twoReplicaWorkloadName = objectGenerator.generateNameByPrefix(NAME_PREFIX.TWO_REPLICA);
                distributedworkloadName = objectGenerator.generateNameByPrefix('distinf');
                trainingJobName = objectGenerator.generateShortNameByPrefix('training-job');

                singleReplicaWorkloadSpec = {
                    ...nimFixtureSpec,
                    servingPort: {
                        container: 8000
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1,
                        metric: 'throughput',
                        metricThreshold: 100
                    },
                    labels: [
                        {
                            name: LABEL.NIM_WORKLOAD,
                            value: 'true'
                        }
                    ]
                };
                twoReplicaWorkloadSpec = {
                    ...nimFixtureSpec,
                    servingPort: {
                        container: 8000
                    },
                    autoscaling: {
                        minReplicas: 2,
                        maxReplicas: 2,
                        metric: 'throughput',
                        metricThreshold: 100
                    },
                    labels: [
                        {
                            name: LABEL.NIM_WORKLOAD,
                            value: 'true'
                        }
                    ]
                };
                distributedworkloadSpec = {
                    workers: 2,
                    replicas: 2,
                    servingPort: {
                        port: 8000
                    },
                    leader: {
                        ...nimFixtureSpec,
                        labels: [
                            {
                                name: LABEL.NIM_WORKLOAD,
                                value: 'true'
                            }
                        ]
                    },
                    worker: {
                        ...nimFixtureSpec
                    }
                };

                logger.info('Creating single-replica inference workload...');
                [singleReplicaWorkloadId, singleReplicaWorkloadName] = await createInference(
                    singleReplicaWorkloadName,
                    singleReplicaWorkloadSpec,
                    project
                );
                logger.info(`Created: ${singleReplicaWorkloadName} (ID: ${singleReplicaWorkloadId})`);

                logger.info('Creating two-replica inference workload...');
                [twoReplicaWorkloadId, twoReplicaWorkloadName] = await createInference(
                    twoReplicaWorkloadName,
                    twoReplicaWorkloadSpec,
                    project
                );
                logger.info(`Created: ${twoReplicaWorkloadName} (ID: ${twoReplicaWorkloadId})`);

                logger.info('Creating distributed inference workload...');
                [distributedworkloadId, distributedworkloadName] = await createDistributedInference(
                    distributedworkloadName,
                    distributedworkloadSpec,
                    project
                );
                logger.info(`Created: ${distributedworkloadName} (ID: ${distributedworkloadId})`);

                logger.info('Creating training workload...');
                const { body: training } = await workloadsCrudUtils.createTraining(
                    trainingJobName,
                    true,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.ubuntu,
                        command: 'sleep infinity'
                    }
                );
                trainingId = training.workloadId;
                logger.info(`Created: ${trainingJobName} (ID: ${trainingId})`);

                logger.info(`Waiting for ${singleReplicaWorkloadName} to be Running...`);
                await waitUtils.waitForBackendWorkloadPhase(
                    singleReplicaWorkloadName,
                    project.id,
                    Phase.Running,
                    150000
                );
                logger.info(`${singleReplicaWorkloadName} is now in Running status`);

                logger.info(`Waiting for ${twoReplicaWorkloadName} to be Running...`);
                await waitUtils.waitForBackendWorkloadPhase(
                    twoReplicaWorkloadName,
                    project.id,
                    Phase.Running,
                    150000
                );
                logger.info(`${twoReplicaWorkloadName} is now in Running status`);

                logger.info(`Waiting for ${trainingJobName} to be Running...`);
                await workloadsCrudUtils.waitForWorkloadPhase(
                    trainingId,
                    Phase.Running,
                    WorkloadType.TRAINING
                );
                logger.info(`${trainingJobName} is now in Running status`);

                logger.info(`Waiting for ${distributedworkloadName} to be Running...`);
                await waitUtils.waitForBackendWorkloadPhase(
                    distributedworkloadName,
                    project.id,
                    Phase.Running,
                    150000
                );
                logger.info(`${distributedworkloadName} is now in Running status`);
            });

            test.beforeEach(async function () {
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                authPage = new AuthPage(browserEngine.currentPage);
                workloadPage = new WorkloadPage(browserEngine.currentPage);
                detailsDrawer = new WorkloadDetailsDrawer(browserEngine.currentPage);
                await authPage.signIn(
                    flowRootState.users.system_administrator.email,
                    flowRootState.users.system_administrator.password
                );
            });

            test.afterEach(async ({}, testInfo) => {
                await browserEngine?.stopAndArchive(testInfo, 'training-' + videoName);
            });

            test.afterAll(async function () {
                if (!project) return;
                try {
                    await operations.deleteProjectNamespaceAndWorkloads(project, 120);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and project ${project.name} namespace ${project.namespace} with error ${error}`
                    );
                }
            });

            test('T5583200: Submit a nim workload(1-replica) and check the metrics on UI', async function () {
                test.setTimeout(360 * 1000);
                videoName = 'nim-metrics-1-replica-ui';
                logger.info('Metrics wait completed');
                await test.step('Check the workload metrics on UI', async () => {
                    await workloadPage.navigateToWorkloads();
                    await detailsDrawer.clickShowDetailsBtn(singleReplicaWorkloadName);
                    await workloadPage.validateNimMetricsForWorkload();
                });

                await test.step('Check the pod metrics on UI', async () => {
                    await workloadPage.validateNimMetricsForEachPod();
                });
            });

            test('T5583201: Submit a nim workload(2-replica) and check the metrics on UI', async function () {
                test.setTimeout(360 * 1000);
                videoName = 'nim-metrics-2-replica-ui';
                await test.step('Check the workload metrics on UI', async () => {
                    await workloadPage.navigateToWorkloads();
                    await detailsDrawer.clickShowDetailsBtn(twoReplicaWorkloadName);
                    await workloadPage.validateNimMetricsForWorkload();
                });

                await test.step('Check the pod metrics on UI', async () => {
                    await workloadPage.validateNimMetricsForEachPod();
                });
            });

            // Coverage remains disabled pending revalidation of this case
            test.skip('T5583202: Submit a distributed inference workload and check the metrics on UI', async function () {
                if (!isLwsInstalled()) {
                    logger.info('LWS is not installed, skipping test');
                    test.skip();
                }
                test.setTimeout(360 * 1000);
                videoName = 'nim-metrics-distributed-ui';
                await test.step('Check the workload metrics on UI', async () => {
                    await workloadPage.navigateToWorkloads();
                    await detailsDrawer.clickShowDetailsBtn(distributedworkloadName);
                    await workloadPage.validateNimMetricsForWorkload();
                });

                await test.step('Check the pod metrics on UI', async () => {
                    await workloadPage.validateNimMetricsForEachPod();
                });
            });

            test('T5583203: Submit a training workload and check no NIM metrics option exists on UI', async function () {
                videoName = 'nim-metrics-training-no-option-ui';
                await test.step('Check the workload metrics on UI', async () => {
                    await workloadPage.navigateToWorkloads();
                    await detailsDrawer.clickShowDetailsBtn(trainingJobName);
                    await workloadPage.page.locator(workloadPage.metricTab).click();
                    await waitUtils.waitForElementToExist(workloadPage, workloadPage.metricTypeSelect);
                    await workloadPage.page.locator(workloadPage.metricTypeSelect).click();
                    const nimOptionCount = await workloadPage.page
                        .locator(workloadPage.nimMetricOption)
                        .count();
                    expect(nimOptionCount).to.equal(0);
                    logger.info('Verified: NVIDIA NIM option does not exist for training workload');
                });
            });
        });
    }
);
