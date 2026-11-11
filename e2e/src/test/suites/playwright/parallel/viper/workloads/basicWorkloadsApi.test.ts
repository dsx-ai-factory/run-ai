/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    crds,
    flowRootState,
    inferenceClientCreator,
    logger,
    objectGenerator,
    operations,
    projectUtils,
    tags,
    Utils,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2,
    workloadUtils
} from '../../../../../../modules/helpers';
import { createProject } from '../../../../../../modules/helpers/models/projects';
import { GpuRequestType } from '../../../../../../modules/api/apis/workloadsApi/generated/model/gpuRequestType';
import { assert, expect } from 'chai';
import { kubeCli, runaiCliV2 } from '../../../../../../modules/cli';
import { validateDeploymentCreatedSuccessfullyInCluster } from '../../../../../../modules/helpers/Validations';
import {
    isKnativeInstalled,
    isKubeflowInstalled
} from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { WorkloadType } from '../../../../../../modules/helpers/config/types';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { DistributedFramework } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedFramework';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';
import { AnyWorkloadSubmissionType } from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

const sleepInfinity = 'sleep infinity';

test.describe(
    `Basic workloads API tests, tags: ${tags.teams.viper} ${tags.resource.workloads} ${tags.runOn.parallelCI}`,
    { tag: [tags.services.workloads, tags.services.agent, tags.services.clusterSync, tags.runOn.parallelCI] },
    function () {
        supportedFrom(SupportedVersions.V2_16)(function () {
            let workloadId;
            let project;
            let namespace;
            let workload;

            test.beforeAll(async function () {
                const options = { maxAllowedGpus: -1, deservedGpus: 1 };
                project = await createProject('project1', options);
                namespace = Utils.namespaceFromProjectName(project.name);
            });

            test.afterAll(async function () {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch (error) {
                    logger.error(`failed to delete project: ${project.name}`, error);
                }
                await projectUtils.waitForProjectCrdToBeDeleted(project.name);
            });

            test.afterEach(async function () {
                if (project?.id) {
                    try {
                        await workloadsUtilsV2.deleteWorkloadsByProjectId(String(project.id));
                    } catch (error) {
                        logger.debug(`failed to delete workloads of project ${project.id}: ${error}`);
                    }
                }
                workload = null;
            });

            test('Should validate resources of training workload', async function () {
                ({ body: workload } = await workloadsCrudUtils.createTraining(
                    'training-workload',
                    false,
                    flowRootState.cluster.uuid,
                    String(project.id),
                    {
                        image: config.quickStartImage.training,
                        compute: {
                            gpuRequestType: GpuRequestType.Portion,
                            gpuPortionRequest: 1
                        }
                    }
                ));
                await workloadUtils.waitForWorkloadClusterResource(
                    workload.name,
                    namespace,
                    AnyWorkloadSubmissionType.Training,
                    crds.runaiTrainingWorkloadCrd
                );
                expect(workload.spec.compute.gpuRequestType).to.equal(GpuRequestType.Portion);
                expect(workload.spec.compute.gpuPortionRequest).to.equal(1);
            });

            test('Should validate type of interactive workload', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'WorkloadRunner phase semantics need investigation'
                );
                ({ body: workload } = await workloadsCrudUtils.createWorkspace(
                    'interactive-workload',
                    false,
                    flowRootState.cluster.uuid,
                    String(project.id),
                    {
                        image: config.quickStartImage.interactive,
                        command: sleepInfinity,
                        compute: {
                            gpuRequestType: GpuRequestType.Portion,
                            gpuDevicesRequest: 4000,
                            gpuPortionRequest: 1
                        }
                    }
                ));
                await workloadUtils.waitForWorkloadRunnerPhase(workload.name, namespace, Phase.Pending);
                expect(workload.spec.command).to.equal(sleepInfinity);
            });

            test(`Creation of training workload with non-existing node pool should fail, ${tags.teams.viper}`, function () {
                // NOTE: IMPORTANT: locally, if you have the env var "ALLOW_DEVELOP_CLI_VERSION=true"
                // then this test will fail!! you need to set: "export ALLOW_DEVELOP_CLI_VERSION=false"
                // in order to run this test successfully. in CI envs this should be the case anyway.
                const workloadName = objectGenerator.generateNameByPrefix('no-nodepool-workload');
                const nodePoolName = 'non-existing-nodepool';
                assert.throw(() => {
                    const submitFlags = {
                        [SubmitFlagsMerged.image]: config.quickStartImage.interactive,
                        [SubmitFlagsMerged.project]: project.name,
                        [SubmitFlagsMerged.nodePools]: nodePoolName
                    } as Partial<typeof SubmitFlagsMerged>;
                    runaiCliV2.submit(submitFlags, workloadName, cliV2WorkloadType.training);
                }, `Error: failed to submit standard training. the following node pools do not exist: ${nodePoolName}`);
            });

            test('Should submit deployment and validate workload', async function () {
                const deploymentName = objectGenerator.generateNameByPrefix('d');

                const deployment = await kubeCli.createDeployment(
                    inferenceClientCreator.getDeployment(
                        namespace,
                        project.name,
                        deploymentName,
                        1,
                        config.quickStartImage.inference
                    )
                );

                validateDeploymentCreatedSuccessfullyInCluster(
                    deployment,
                    deploymentName,
                    namespace,
                    project.name,
                    1,
                    config.quickStartImage.inference
                );

                const workloadId = deployment.metadata.uid;
                const expectedSource = 'Other';
                const expectedType = WorkloadType.deployment;

                await workloadsUtilsV2.waitForWorkloadInControlPlane(workloadId);
                await workloadsUtilsV2.getWorkloadAndValidate(workloadId, {
                    expectedSource: expectedSource,
                    expectedType: expectedType
                });

                workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;
            });

            test.describe('Inference workload', function () {
                test.beforeAll(async function () {
                    if (!(await isKnativeInstalled())) {
                        logger.warn('knative is not installed on the cluster, skipping test');
                        test.skip();
                    }
                });

                test('Should submit inference workload and validate min/max replicas', async function () {
                    const workloadName = objectGenerator.generateNameByPrefix('inference-workload');
                    ({ body: workload } = await workloadsCrudUtils.createInference(
                        workloadName,
                        false,
                        flowRootState.cluster.uuid,
                        String(project.id),
                        {
                            image: config.quickStartImage.inferenceHelloWorld,
                            servingPort: { container: 8080 },
                            autoscaling: {
                                minReplicas: 2,
                                maxReplicas: 4,
                                metric: 'throughput',
                                metricThreshold: 50
                            }
                        }
                    ));
                    await workloadUtils.waitForWorkloadClusterResource(
                        workload.name,
                        namespace,
                        AnyWorkloadSubmissionType.Inference,
                        crds.runaiInferenceWorkloadCrd
                    );
                    const minReplicas = 2;
                    const maxReplicas = 4;

                    expect(workload.spec.autoscaling.minReplicas).to.equal(minReplicas);
                    expect(workload.spec.autoscaling.maxReplicas).to.equal(maxReplicas);
                });
            });

            test.describe('Pytorch workload', function () {
                test.beforeAll(async function () {
                    if (!(await isKubeflowInstalled())) {
                        logger.warn('kubeflow is not installed on the cluster, skipping test');
                        test.skip();
                    }
                });

                test('Should submit distributed training workload and validate distributed type', async function () {
                    ({ body: workload } = await workloadsCrudUtils.createDistributed(
                        'pytorch-workload',
                        false,
                        flowRootState.cluster.uuid,
                        String(project.id),
                        {
                            distributedFramework: DistributedFramework.PyTorch,
                            numWorkers: 1,
                            command: sleepInfinity,
                            image: config.quickStartImage.interactive
                        },
                        true
                    ));
                    await workloadUtils.waitForWorkloadClusterResource(
                        workload.name,
                        namespace,
                        AnyWorkloadSubmissionType.Distributed,
                        crds.runaiDistributedWorkloadCrd
                    );
                    expect(workload.spec.distributedFramework).to.equal(DistributedFramework.PyTorch);
                });
            });

            test.describe('Delete workload', function () {
                const workloadPrefix = 'workload-to-delete';
                let deletedWorkloadName: string;

                test.beforeAll(async function () {
                    // submit workload
                    ({ body: workload } = await workloadsCrudUtils.createTraining(
                        workloadPrefix,
                        false,
                        flowRootState.cluster.uuid,
                        String(project.id),
                        { image: config.quickStartImage.training }
                    ));
                    deletedWorkloadName = workload.name;
                    await workloadsUtilsV2.waitForWorkloadInControlPlane(workload.workloadId);

                    await workloadsCrudUtils.deleteTraining(workload.workloadId);

                    await waitUtils.waitForWorkloadToBeDeleted(deletedWorkloadName);

                    // afterEach hook will skip its own delete since the workload is already gone
                    workload = null;
                });

                test('Should validate that workload is deleted and validate it is in deleted workloads', async function () {
                    // validate workload was deleted
                    const allWorkloadsAfterDelete = (await workloadsUtilsV2.getAllWorkloads()).body.workloads;

                    // expect all the workloads in the workloads array to not contain the deleted workload
                    for (const workloadAfterDelete of allWorkloadsAfterDelete) {
                        expect(workloadAfterDelete.name).to.not.equal(deletedWorkloadName);
                    }
                });
            });
        });
    }
);
