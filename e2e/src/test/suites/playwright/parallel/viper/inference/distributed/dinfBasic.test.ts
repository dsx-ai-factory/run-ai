/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    config,
    distributedInferenceUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    versionUtils,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { isLwsInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { test, TestInfo } from '@playwright/test';
import { DistributedInferenceSpecSpec } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceSpecSpec';
import { DistributedInferenceStartupPolicy } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceStartupPolicy';
import { DistributedInferenceRestartPolicy } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceRestartPolicy';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { DistributedInferenceLeaderWorkerSpecCompute } from '../../../../../../../modules/api/apis/workloadsApi/generated/api';

supportedFrom(SupportedVersions.V2_23)(function () {
    test.describe(
        `distributed inference workload basic tests: ${tags.resource.distInferenceWorkload} ${tags.tool.api} ${tags.teams.viper}`,
        { tag: [tags.resource.distInferenceWorkload] },
        function () {
            let project: Project;
            let workloadId: string;
            let workloadName: string;
            let workloadSpec: DistributedInferenceSpecSpec;
            test.beforeAll(async function () {
                if (!(await isLwsInstalled())) {
                    logger.warn('LeaderWorkerSet (LWS) is not installed on the cluster, skipping test');
                    test.skip();
                }

                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('project'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 2, limit: 2 } }]
                });
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateNameByPrefix('workload');
                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;

                workloadSpec = distributedInferenceUtils.getBaseWorkloadSpec();
            });

            test.afterAll(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and project ${project.status.namespace}; error: ${error}`
                    );
                }
            });

            test.afterEach(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads of project ${project.status.namespace}; error: ${error}`
                    );
                }
            });

            const basicReadinessTestParameters = [
                { workers: 0, replicas: 1 },
                { workers: 1, replicas: 1 },
                { workers: 3, replicas: 1 },
                { workers: 1, replicas: 0 },
                { workers: 1, replicas: 2 }
            ];

            basicReadinessTestParameters.forEach(({ workers, replicas }) => {
                test(`creates a workload with ${workers} workers and ${replicas} replicas`, async function ({}, testInfo: TestInfo) {
                    workloadSpec.workers = workers;
                    workloadSpec.replicas = replicas;

                    if (workers === 0) {
                        workloadSpec.worker = null;
                    }

                    // create the workload
                    [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                        workloadName,
                        workloadSpec,
                        String(project.id)
                    );

                    const expectedPods = (1 + workers) * replicas;
                    // wait for the workload to be in phase Running
                    const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                        workloadId,
                        Phase.Running,
                        120,
                        expectedPods
                    );

                    // validate get workload result
                    distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec);
                });
            });

            supportedFrom(SupportedVersions.V2_23)(function () {
                test('creates a workload and manual scale it', async function () {
                    workloadSpec.workers = 1;
                    workloadSpec.replicas = 1;

                    // create the workload
                    [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                        workloadName,
                        workloadSpec,
                        String(project.id)
                    );

                    // wait for the workload to be in phase Running
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                    logger.info(`Testing scaling workflow for distributed inference workload ${workloadId}`);

                    // Capture original pod ids before scaling
                    const originalPodIds = await distributedInferenceUtils.getWorkloadPodIds(workloadId);

                    // Test API scaling
                    const apiScaleTarget = 2;
                    await distributedInferenceUtils.scaleDistributedInferenceViaApi(
                        workloadId,
                        apiScaleTarget
                    );
                    const newPodIds = await distributedInferenceUtils.validateScalingResults(
                        workloadId,
                        apiScaleTarget,
                        workloadSpec,
                        originalPodIds
                    );

                    // Test direct cluster scaling
                    const clusterScaleTarget = 1;
                    const expectedPodsAfterClusterScale =
                        (1 + distributedInferenceUtils.getNumWorkers(workloadSpec)) *
                        clusterScaleTarget;
                    if (
                        versionUtils.isVersionSupported(
                            flowRootState.cluster.version,
                            SupportedVersions.V3_0
                        )
                    ) {
                        await distributedInferenceUtils.scaleWorkloadRunnerViaKubectl(
                            workloadId,
                            workloadName,
                            project.status.namespace,
                            clusterScaleTarget,
                            expectedPodsAfterClusterScale
                        );
                    } else {
                        await distributedInferenceUtils.scaleDistributedInferenceViaKubectl(
                            workloadId,
                            workloadName,
                            project.status.namespace,
                            clusterScaleTarget
                        );
                    }
                    await distributedInferenceUtils.validateScalingResults(
                        workloadId,
                        clusterScaleTarget,
                        workloadSpec,
                        newPodIds
                    );

                    logger.info(
                        `Successfully completed scaling workflow test for distributed inference workload ${workloadId}`
                    );
                });
            });

            test('creates a workload with more advanced configuration', async function () {
                // set non-default startupPolicy and restartPolicy
                workloadSpec.startupPolicy = DistributedInferenceStartupPolicy.LeaderReady;
                workloadSpec.restartPolicy = DistributedInferenceRestartPolicy.None;

                // set different images for leader and worker
                workloadSpec.leader.image = config.quickStartImage.inferenceHelloWorld;
                workloadSpec.worker.image = config.quickStartImage.inferenceAutoScale;

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                const expectedPods = (1 + workloadSpec.workers) * workloadSpec.replicas;
                // wait for the workload to be in phase Running
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Running,
                    120,
                    expectedPods
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec);
            });

            test('creates a workload with compute requirements', async function () {
                // set compute requirements
                workloadSpec.leader.compute = new DistributedInferenceLeaderWorkerSpecCompute();
                workloadSpec.leader.compute.gpuDevicesRequest = 1;
                workloadSpec.worker.compute = new DistributedInferenceLeaderWorkerSpecCompute();
                workloadSpec.worker.compute.gpuPortionRequest = 0.5;

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                const expectedPods = (1 + workloadSpec.workers) * workloadSpec.replicas;
                // wait for the workload to be in phase Running
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Running,
                    120,
                    expectedPods
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec);
            });

            test('creates a workload with compute requirements, multiple workers and replicas', async function ({}, testInfo: TestInfo) {
                // set compute requirements
                workloadSpec.workers = 3;
                workloadSpec.replicas = 2;
                workloadSpec.leader.compute = new DistributedInferenceLeaderWorkerSpecCompute();
                workloadSpec.leader.compute.gpuPortionRequest = 0.2;
                workloadSpec.worker.compute = new DistributedInferenceLeaderWorkerSpecCompute();
                workloadSpec.worker.compute.gpuPortionRequest = 0.1;

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                const expectedPods = (1 + workloadSpec.workers) * workloadSpec.replicas;
                // wait for the workload to be in phase Running
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Running,
                    120,
                    expectedPods
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec);
            });

            test('creates distributed inference with mixed GPU allocation and validates allocation', async function () {
                // Test case: Leader with fraction, worker with whole GPU
                workloadSpec.workers = 1;
                workloadSpec.replicas = 1;
                workloadSpec.leader.compute = new DistributedInferenceLeaderWorkerSpecCompute();
                workloadSpec.leader.compute.gpuPortionRequest = 0.5;
                workloadSpec.worker.compute = new DistributedInferenceLeaderWorkerSpecCompute();
                workloadSpec.worker.compute.gpuDevicesRequest = 1;

                // create the workload
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                const expectedPods = (1 + workloadSpec.workers) * workloadSpec.replicas;
                // wait for the workload to be in phase Running
                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Running,
                    120,
                    expectedPods
                );

                // validate get workload result
                distributedInferenceUtils.validateGetDistributedInference(workload, workloadSpec);

                // Additional explicit validation that allocation equals request
                const leaderGpu = workloadSpec.leader.compute.gpuPortionRequest;
                const workerGpu = workloadSpec.worker.compute.gpuDevicesRequest;
                const expectedTotal = (leaderGpu + workloadSpec.workers * workerGpu) * workloadSpec.replicas;

                expect(workload.allocatedResources?.gpu ?? 0).to.be.closeTo(expectedTotal, 0.001);

                logger.info(
                    `Validated distributed inference GPU allocation: leader=${leaderGpu}, worker=${workerGpu}, total=${expectedTotal}`
                );
            });
        }
    );
});
