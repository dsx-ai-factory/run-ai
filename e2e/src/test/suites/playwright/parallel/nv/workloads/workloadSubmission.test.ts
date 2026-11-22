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
    flowRootState,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { EWorkloadStatus, WorkloadPage } from '../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { TrainingPage } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { ProjectPage } from '../../../../../../modules/ui/runaiV2/pages/ProjectPage';
import { WorkloadType } from '../../../../../../modules/api/apis/assetsApi';
import { DISTRIBUTED_TOOLS_LABELS } from '../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { ServingPortProtocol } from '../../../../../../modules/api/apis/workloadsApi/generated/model/servingPortProtocol';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { kubeCli } from '../../../../../../modules/cli';
import { WorkloadType as WorkloadTypeCrud } from '../../../../../../modules/helpers/utils/WorkloadsCrudUtils';

test.describe(`Workload Submission Tests using Hybrid Submission  ${tags.tool.ui}  ${tags.resource.workloads} ${tags.runOn.dgxc} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let browserEngine: BrowserEngine;
        let videoName: string;
        let auth: AuthPage;
        let training: TrainingPage;
        let workloads: WorkloadPage;
        let projectPage: ProjectPage;
        let trainingName: string;

        test.beforeAll(async function () {
            browserEngine = new BrowserEngine();
            projectName = objectGenerator.generateShortNameByPrefix('project-e2e-for-training');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 6,
                maxAllowedGpus: 6
            });
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

        test.beforeEach(async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            training = new TrainingPage(browserEngine.currentPage);
            workloads = new WorkloadPage(browserEngine.currentPage);
            projectPage = new ProjectPage(browserEngine.currentPage);
            await auth.signIn(
                flowRootState.users.l1_researcher.email,
                flowRootState.users.l1_researcher.password
            );
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        // Multi-resource workload submission coverage remains disabled pending revalidation.
        test.skip('T3793928: Verify the cluster have reasonable set of compute resources and tenant user can submit jobs with different resources', async () => {
            videoName = 'T3793928-compute-resources-different-jobs';
            test.setTimeout(600 * 1000);
            const computeResources = {
                'cpu-only': '0',
                'one-gpu': '1',
                'two-gpus': '2'
            };

            for (const [computeLabel, expectedGpu] of Object.entries(computeResources)) {
                // Test single node workload
                trainingName = objectGenerator.generateNameByPrefix(`sn-${computeLabel}`);
                await test.step(`Create workload with ${computeLabel} compute resource`, async () => {
                    await training.createFlexibleTraining(
                        projectName,
                        trainingName,
                        undefined,
                        undefined,
                        'tensorboard',
                        computeLabel
                    );
                    await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Running);
                });

                await test.step(`Verify workload details for ${computeLabel}`, async () => {
                    await workloads.checkDisplayedWorkloadDetails({
                        jobName: trainingName,
                        requestedGpu: expectedGpu === '0' ? '-' : expectedGpu,
                        type: WorkloadType.Training,
                        projectName,
                        status: EWorkloadStatus.Running,
                        createdBy: flowRootState.users.l1_researcher.email,
                        pods: '1/1'
                    });
                });

                await test.step(`Stop workload ${trainingName}`, async () => {
                    await workloads.stopWorkloadByName(trainingName);
                    await workloads.waitForWorkloadUIStatus(trainingName, EWorkloadStatus.Stopped);
                });

                await test.step(`Delete workload ${trainingName}`, async () => {
                    await workloads.deleteWorkload(trainingName);
                });

                // Test distributed workload
                const distributedName = objectGenerator.generateNameByPrefix(`dist-${computeLabel}`);
                await test.step(`Create distributed training job with ${computeLabel} compute resource`, async () => {
                    await training.fillFirstNewTrainingPage(
                        projectName,
                        distributedName,
                        true,
                        DISTRIBUTED_TOOLS_LABELS.PY_TORCH,
                        true,
                        true
                    );
                    await training.fillHybridDistributedTrainingSecondPage(config.quickStartImage.ubuntu, {
                        command: 'sleep infinity',
                        computeLabel: computeLabel,
                        workersCount: '2'
                    });
                    await training.waitForWorkloadUIStatus(distributedName, EWorkloadStatus.Running);
                    const requestedGpu = expectedGpu === '0' ? '-' : (Number(expectedGpu) * 3).toString(); // 2 workers + 1 master
                    await training.checkDisplayedWorkloadDetails({
                        jobName: distributedName,
                        requestedGpu,
                        type: WorkloadType.Training,
                        projectName,
                        status: EWorkloadStatus.Running,
                        createdBy: flowRootState.users.l1_researcher.email,
                        pods: '3/3'
                    });
                });

                await test.step(`Stop distributed training job ${distributedName}`, async () => {
                    await workloads.stopWorkloadByName(distributedName);
                    await workloads.waitForWorkloadUIStatus(distributedName, EWorkloadStatus.Stopped);
                });

                await test.step(`Delete distributed training job ${distributedName}`, async () => {
                    await workloads.deleteWorkload(distributedName);
                });
            }
        });
    });
});

test.describe(`ML Engineer can submit inference job, tags: ${tags.resource.workloads} ${tags.teams.loong} ${tags.runOn.dgxc}`, () => {
    supportedFrom(SupportedVersions.V2_20)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let inferenceName: string;
        test.beforeAll(async () => {
            projectName = objectGenerator.generateNameByPrefix('e2e-project');
            inferenceName = objectGenerator.generateNameByPrefix('e2e-inference');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            workloadsCrudUtils.accessToken = flowRootState.users.ml_engineer.accessToken;
        });

        test.afterAll(async () => {
            if (project && project.id) {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`deleted project ${projectName} and namespace ${namespace}`);
                } catch {
                    logger.error('failed to delete project ' + projectName + ' and namespace ' + namespace);
                }
            }
        });

        // skip this test because the image is large and it will get no space error in automation pipeline
        test.skip('T3893341: Verify ML engineer can create inference job', async () => {
            test.setTimeout(480 * 1000);
            let inference: any;
            let inferenceDetails: any;
            let url: string;
            let clientWorkloadName: string;

            await test.step('Create inference workload with ML Engineer', async () => {
                const { body: inferenceResponse } = await workloadsCrudUtils.createInference(
                    inferenceName,
                    false,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: config.quickStartImage.nvidiaTritonServer,
                        compute: { gpuDevicesRequest: 1 },
                        servingPort: {
                            container: 8000,
                            protocol: ServingPortProtocol.Http
                        }
                    }
                );
                inference = inferenceResponse;

                await workloadsCrudUtils.waitForWorkloadPhase(
                    inference.workloadId,
                    Phase.Running,
                    WorkloadTypeCrud.INFERENCE,
                    undefined,
                    300
                );

                inferenceDetails = await workloadsUtilsV2.getWorkloadByWorkloadId(inference.workloadId);
                url = inferenceDetails.body.urls[0];
                logger.info(`url: ${url}`);
            });

            await test.step('Create client workload to connect to inference server', async () => {
                const healthCheckCommand = `curl -i ${url}/v2/health/ready`;
                const clientCommand = `/workspace/install/bin/image_client -m densenet_onnx -c 3 -s INCEPTION -u ${url} /workspace/images/mug.jpg`;
                const bashCommand = '/bin/bash -c';
                const bashArgs = `"${healthCheckCommand} && ${clientCommand} && sleep infinity"`;
                const tritonClientImage = 'nvcr.io/nvidia/tritonserver:24.07-py3-sdk';
                clientWorkloadName = objectGenerator.generateNameByPrefix('inference-client');

                workloadsCrudUtils.accessToken = flowRootState.users.l1_researcher.accessToken;
                const { body: healthCheckWorkload } = await workloadsCrudUtils.createTraining(
                    clientWorkloadName,
                    false,
                    flowRootState.cluster.uuid,
                    project.id.toString(),
                    {
                        image: tritonClientImage,
                        command: bashCommand,
                        args: bashArgs
                    }
                );
                await workloadsCrudUtils.waitForWorkloadPhase(
                    healthCheckWorkload.workloadId,
                    Phase.Running,
                    WorkloadTypeCrud.TRAINING,
                    undefined,
                    300
                );

                const clientWorkloadLogs = await kubeCli.getLogs(clientWorkloadName, namespace);
                logger.info('clientWorkloadLogs', clientWorkloadLogs);
                expect(clientWorkloadLogs).toContain('200 OK');
                expect(clientWorkloadLogs).toContain('/workspace/images/mug.jpg');
                expect(clientWorkloadLogs).toContain('COFFEE MUG');
                expect(clientWorkloadLogs).toContain('CUP');
                expect(clientWorkloadLogs).toContain('COFFEEPOT');
            });

            await test.step('Delete workloads', async () => {
                await workloadsUtilsV2.deleteAllWorkloads();
            });
        });
    });
});
