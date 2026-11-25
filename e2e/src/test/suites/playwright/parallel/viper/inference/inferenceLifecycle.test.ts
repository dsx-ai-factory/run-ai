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
    orgUnitUtils,
    tags,
    versionUtils,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { inferenceUtils } from '../../../../../../modules/helpers';
import { InferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceSpecSpec';
import { SupportedVersions } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { InferenceUpdateSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceUpdateSpecSpec';
import { WorkloadDetailed } from '../../../../../../modules/api/apis/workloadsApi/generated/model/workloadDetailed';
import { Event } from '../../../../../../modules/api/apis/workloadsApi/generated/model/event';
import { WorkspaceSpecSpecAllOfCompute } from '../../../../../../modules/api/apis/workloadsApi/generated/api';

test.describe(
    `inference workload lifecycle tests, tags: ${tags.resource.inferenceWorkload} ${tags.teams.viper} ${tags.runOn.parallelCI}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        test.describe.configure({ retries: 1 });
        let project: Project;
        let workloadId: string;
        let workloadName: string;
        let workloadSpec: InferenceSpecSpec;
        let updatedWorkloadSpec: InferenceUpdateSpecSpec;

        async function validateBackendGetInferenceWorkloadResponse(
            workloadId,
            workloadSpec,
            workloadName,
            expectedLatestCreatedRevisionId,
            expectedLatestReadyRevisionId
        ): Promise<WorkloadDetailed> {
            const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);

            expect(workload.body.images).to.have.lengthOf(1);
            expect(workload.body.images[0]).to.equal(workloadSpec.image);
            expect(workload.body.requestedPods.min).to.equal(workloadSpec.autoscaling.minReplicas);
            expect(workload.body.requestedPods.max).to.equal(workloadSpec.autoscaling.maxReplicas);

            if (
                expectedLatestCreatedRevisionId !== undefined ||
                expectedLatestReadyRevisionId !== undefined
            ) {
                expect(workload.body.additionalFields).to.not.be.undefined;

                if (expectedLatestCreatedRevisionId !== undefined) {
                    expect(workload.body.additionalFields).to.have.property('latestCreatedRevisionName');
                    expect(workload.body.additionalFields['latestCreatedRevisionName']).to.equal(
                        workloadName + '-' + expectedLatestCreatedRevisionId
                    );
                } else {
                    expect(workload.body.additionalFields).to.not.have.property('latestCreatedRevisionName');
                }

                if (expectedLatestReadyRevisionId !== undefined) {
                    expect(workload.body.additionalFields).to.have.property('latestReadyRevisionName');
                    expect(workload.body.additionalFields['latestReadyRevisionName']).to.equal(
                        workloadName + '-' + expectedLatestReadyRevisionId
                    );
                } else {
                    expect(workload.body.additionalFields).to.not.have.property('latestReadyRevisionName');
                }
            }

            return workload.body;
        }
        test.beforeAll(async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            project = await orgUnitUtils.createProjectAndValidate({
                name: objectGenerator.generateShortNameByPrefix('proj'),
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 2, limit: 2 } }]
            });
        });

        test.beforeEach(function () {
            workloadName = objectGenerator.generateNameByPrefix('wl');

            workloadSpec = {
                image: config.quickStartImage.inferenceHelloWorld,
                servingPort: {
                    container: 8080
                },
                autoscaling: {
                    minReplicas: 1,
                    maxReplicas: 2,
                    metric: 'throughput',
                    metricThreshold: 100
                }
            };

            inferenceUtils.setKnativeProgressDeadlineSeconds(600);

            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterAll(async function () {
            if (project === undefined) {
                return;
            }

            try {
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            } catch (error) {
                logger.error(
                    `Failed to delete workloads and project ${project.name} namespace ${project.status.namespace}`
                );
            }
        });

        test.afterEach(async function () {
            if (project === undefined) {
                return;
            }

            try {
                const projectWorkloadNames = (
                    await workloadsUtilsV2.getWorkloadsByProjectId(String(project.id))
                ).body.workloads.map((workload) => workload.name);

                await workloadsUtilsV2.deleteWorkloadsByProjectId(
                    String(project.id),
                    120,
                    flowRootState.users.system_administrator.accessToken,
                    false
                );
                await waitUtils.waitForWorkloadPodsNotExist(
                    project.status.namespace,
                    projectWorkloadNames,
                    120
                );
            } catch (error) {
                logger.error(
                    `Failed to isolate project ${project.name} after deleting its workloads: ${error}`
                );
                throw error;
            }
        });

        test('creates a workload that becomes ready', async function () {
            // create the workload
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // wait for the workload to be in phase Running
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

            // validate get workload result
            const workload = await validateBackendGetInferenceWorkloadResponse(
                workloadId,
                workloadSpec,
                workloadName,
                '00001',
                '00001'
            );

            logger.info(`sending inference request to workload ${workloadName}`);
            const clientWorkloadName = await inferenceUtils.sendInferenceRequest(
                project.status.namespace,
                inferenceUtils.getInternalUrlForInferenceWorkload(workload)
            );

            // Verify that cluster-sync propagates Knative events to workloads-service.
            // propagation is asynchronous, so poll until both controller events arrive.
            if (versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_20)) {
                logger.info(`waiting for knative controller events for workload ${workloadName}`);

                const requiredSources = ['configuration-controller', 'revision-controller'] as const;
                let workloadEvents: Event[] = [];

                await waitUtils.waitForCondition(
                    async () => {
                        workloadEvents =
                            (await workloadsUtilsV2.getWorkloadEvents(workloadId)).body?.events ?? [];
                        return requiredSources.every((source) =>
                            workloadEvents.some((event) => event.source === source)
                        );
                    },
                    `workload ${workloadName} has events from knative ${requiredSources.join(' and ')}`,
                    true,
                    60,
                    2
                );

                const observedSources = Array.from(
                    new Set(workloadEvents.map((event) => event.source).filter(Boolean))
                );
                for (const source of requiredSources) {
                    expect(
                        workloadEvents.some((event) => event.source === source),
                        `expected a ${source} event for workload ${workloadName}; observed sources: ${observedSources.join(', ')}`
                    ).to.equal(true);
                }
            }
            await inferenceUtils.deleteInferenceClientJob(clientWorkloadName, project.status.namespace);
        });

        test('creates a workload that fails due to non-existing image', async function () {
            // create the workload with a non-existing image
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            workloadSpec.image = config.quickStartImage.nonExistingImage;
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // wait for the workload to be in phase Running
            logger.info(`waiting for inference workload ${workloadName} to be in phase Failed`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Failed);

            // validate get workload result
            logger.info(`validating get workload result for workload ${workloadName}`);
            await validateBackendGetInferenceWorkloadResponse(
                workloadId,
                workloadSpec,
                workloadName,
                '00001',
                undefined
            );
        });

        test('creates a workload that fails due to a runtime error', async function () {
            // set a short knative progress deadline to speed up the test, making the revision fail faster
            inferenceUtils.setKnativeProgressDeadlineSeconds(30);

            // create the workload with a command that fails
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            workloadSpec.image = config.quickStartImage.ubuntu; // no entrypoint, so it exists immediately
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // wait for the workload to be in phase Initializing, and finally in phase Failed
            // due to backoff limit, failure takes time, which allows us test the Initializing phase as well
            logger.info(`waiting for inference workload ${workloadName} to be in phase Initializing`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Initializing);
            logger.info(`waiting for inference workload ${workloadName} to be in phase Failed`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Failed);
        });

        test('creates a workload that is pending', async function () {
            // create the workload with a request for too many gpus, so it will get stuck on pending phase
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            workloadSpec.compute = new WorkspaceSpecSpecAllOfCompute();
            workloadSpec.compute.gpuDevicesRequest = 256;
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // wait for the workload to be in phase Pending
            // verify it's not just a temporary phase as with any workload by sleeping for a while and checking again
            logger.info(`waiting for inference workload ${workloadName} to be in phase Pending`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Pending);
            await waitUtils.sleep(10);
            logger.info(
                `waiting for inference workload ${workloadName} to be in phase Pending (second time)`
            );
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Pending);
        });

        test('updates a workload that is ready, new revision becomes ready', async function () {
            // create the workload
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // wait for the workload to be in phase Running
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 180);

            // update the workload spec with a new image and new autoscaling settings
            workloadSpec.image = config.quickStartImage.inferenceAutoScale;
            workloadSpec.autoscaling.maxReplicas = 3;
            logger.info(`updating inference workload ${workloadName} with new spec`);
            await inferenceUtils.updateInferenceSpec(workloadId, workloadSpec);

            // wait for the workload to have an additional field with the new revision name,
            // this is the indication that update is in progress
            logger.info('waiting for workload additional field with the new revision name');
            await waitUtils.waitForWorkloadAdditionalField(
                workloadId,
                'latestReadyRevisionName',
                workloadName + '-' + '00002'
            );

            // validate get workload result - new spec and new revision names
            logger.info(`validating get workload result for workload ${workloadName}`);
            const workload = await validateBackendGetInferenceWorkloadResponse(
                workloadId,
                workloadSpec,
                workloadName,
                '00002',
                '00002'
            );

            // verify the workload is indeed serving requests
            logger.info(`sending inference request to workload ${workloadName}`);
            const clientWorkloadName = await inferenceUtils.sendInferenceRequest(
                project.status.namespace,
                inferenceUtils.getInternalUrlForInferenceWorkload(workload)
            );
            await inferenceUtils.deleteInferenceClientJob(clientWorkloadName, project.status.namespace);
        });

        test('updates a workload that is ready, new revision is pending', async function () {
            // create the workload, using 1 gpu so when we update we use many gpu replicas to get it stuck on pending
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            workloadSpec.compute = new WorkspaceSpecSpecAllOfCompute();
            workloadSpec.compute.gpuDevicesRequest = 1;
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // wait for the workload to be in phase Running
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

            // update the workload spec with too many replicas
            workloadSpec.autoscaling.minReplicas = 256;
            workloadSpec.autoscaling.maxReplicas = 256;
            logger.info(`updating inference workload ${workloadName} with new spec`);
            await inferenceUtils.updateInferenceSpec(workloadId, workloadSpec);

            // wait for the workload to be in phase Running (just to make sure, as it's supposed to be Running already from the start)
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running);

            // wait for the workload to have an additional field with the new revision name,
            // this is the indication that update is in progress
            logger.info('waiting for workload additional field with the new revision name');
            await waitUtils.waitForWorkloadAdditionalField(
                workloadId,
                'latestCreatedRevisionName',
                workloadName + '-' + '00002'
            );

            // validate get workload result - new spec and new revision names
            // the previous revision should still be the latest ready revision
            logger.info(`validating get workload result for workload ${workloadName}`);
            const workload = await validateBackendGetInferenceWorkloadResponse(
                workloadId,
                workloadSpec,
                workloadName,
                '00002',
                '00001'
            );

            // verify the workload is still serving requests, using the previous revision
            logger.info(`sending inference request to workload ${workloadName}`);
            const clientWorkloadName = await inferenceUtils.sendInferenceRequest(
                project.status.namespace,
                inferenceUtils.getInternalUrlForInferenceWorkload(workload)
            );
            await inferenceUtils.deleteInferenceClientJob(clientWorkloadName, project.status.namespace);
        });

        test('creates inference workload with GPU fraction and validates allocation', async function () {
            // Create workload with GPU fraction
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            workloadSpec.compute = new WorkspaceSpecSpecAllOfCompute();
            workloadSpec.compute.gpuPortionRequest = 0.5;
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // Wait for the workload to be in phase Running
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

            // Get workload and validate allocation
            const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);

            const expectedGpuRequest =
                workloadSpec.compute.gpuPortionRequest * workloadSpec.autoscaling.minReplicas;

            // Validate allocation equals request (for running workload)
            logger.info('validating allocation equals request');
            const allocatedGpu = workload.body.allocatedResources?.gpu ?? 0;
            expect(allocatedGpu).to.be.closeTo(expectedGpuRequest, 0.001);

            logger.info(
                `Successfully validated GPU fraction allocation: requested=${expectedGpuRequest}, allocated=${allocatedGpu}`
            );
        });

        test('updates a workload that is ready, new revision fails due to non-existing image', async function () {
            // create the workload
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // wait for the workload to be in phase Running
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

            // update the workload spec
            workloadSpec.image = config.quickStartImage.nonExistingImage;
            logger.info(`updating inference workload ${workloadName} with new spec`);
            await inferenceUtils.updateInferenceSpec(workloadId, workloadSpec);

            // wait for the new knative revision to be created in the cluster first
            logger.info(`waiting for knative revision ${workloadName}-00002 to be created`);
            await waitUtils.waitForKnativeRevision(workloadName + '-00002', project.status.namespace, 120);

            // wait for the workload to be in phase Running (just to make sure, as it's supposed to be Running already from the start)
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running);

            // wait for the workload to have an additional field with the new revision name,
            // this is the indication that update is in progress
            logger.info('waiting for workload additional field with the new revision name');
            await waitUtils.waitForWorkloadAdditionalField(
                workloadId,
                'latestCreatedRevisionName',
                workloadName + '-' + '00002'
            );

            // validate get workload result - new spec and new revision names
            // the previous revision should still be the latest ready revision
            logger.info(`validating get workload result for workload ${workloadName}`);
            const workload = await validateBackendGetInferenceWorkloadResponse(
                workloadId,
                workloadSpec,
                workloadName,
                '00002',
                '00001'
            );

            // verify the workload is still serving requests, using the previous revision
            const clientWorkloadName = await inferenceUtils.sendInferenceRequest(
                project.status.namespace,
                inferenceUtils.getInternalUrlForInferenceWorkload(workload)
            );
            await inferenceUtils.deleteInferenceClientJob(clientWorkloadName, project.status.namespace);
        });

        test('updates a workload that is currently being updated', async function () {
            test.setTimeout(180 * 1000);

            // create the workload
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // wait for the workload to be in phase Running
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

            // update the workload spec and wait for it to have a new revision
            // waiting for a revision in the cluster and not in the control-plane as sync times may be long
            // and it's not necessary
            workloadSpec.image = config.quickStartImage.nonExistingImage;
            logger.info(`updating inference workload ${workloadName} - first update`);
            await inferenceUtils.updateInferenceSpec(workloadId, workloadSpec);
            logger.info('waiting for new knative revision to be created');
            await waitUtils.waitForKnativeRevision(workloadName + '-00002', project.status.namespace, 120);

            // update the workload spec again and wait for it to be ready
            workloadSpec.image = config.quickStartImage.inferenceHelloWorld;
            logger.info(`updating inference workload ${workloadName} - second update`);
            await inferenceUtils.updateInferenceSpec(workloadId, workloadSpec);
            // wait for the new revision to be synced as part of the workload additional fields
            logger.info('waiting for workload additional field with the new revision name');
            await waitUtils.waitForWorkloadAdditionalField(
                workloadId,
                'latestReadyRevisionName',
                workloadName + '-' + '00003'
            );

            // wait for the workload to be in phase Running
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

            // validate get workload result - new spec and new revision names
            // the previous revision should still be the latest ready revision
            logger.info(`validating get workload result for workload ${workloadName}`);
            await validateBackendGetInferenceWorkloadResponse(
                workloadId,
                workloadSpec,
                workloadName,
                '00003',
                '00003'
            );
        });

        test('updates a workload that is failed', async function () {
            test.setTimeout(240 * 1000); // this test might take longer than the default timeout

            // create the workload
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            workloadSpec.image = config.quickStartImage.nonExistingImage;
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // wait for the workload to be in phase Failed
            logger.info(`waiting for inference workload ${workloadName} to be in phase Failed`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Failed, 180);

            // update the workload spec and wait for it to be ready
            workloadSpec.image = config.quickStartImage.inferenceHelloWorld;
            logger.info(`updating inference workload ${workloadName} with new spec`);
            await inferenceUtils.updateInferenceSpec(workloadId, workloadSpec);

            // wait for the workload to be in phase Running
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

            logger.info('waiting for workload additional field with the new revision name');
            await waitUtils.waitForWorkloadAdditionalField(
                workloadId,
                'latestReadyRevisionName',
                workloadName + '-' + '00002'
            );

            // validate get workload result - new spec and new revision names
            // the previous revision should still be the latest ready revision
            logger.info(`validating get workload result for workload ${workloadName}`);
            const workload = await validateBackendGetInferenceWorkloadResponse(
                workloadId,
                workloadSpec,
                workloadName,
                '00002',
                '00002'
            );

            // verify the workload is indeed serving requests
            logger.info(`sending inference request to workload ${workloadName}`);
            const clientWorkloadName = await inferenceUtils.sendInferenceRequest(
                project.status.namespace,
                inferenceUtils.getInternalUrlForInferenceWorkload(workload)
            );
            await inferenceUtils.deleteInferenceClientJob(clientWorkloadName, project.status.namespace);
        });

        test('updates a workload that is initializing', async function () {
            test.setTimeout(180 * 1000);

            // create the workload with a command that fails
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            workloadSpec.image = config.quickStartImage.ubuntu; // no entrypoint, so it exists immediately
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // wait for the workload to be in phase Initializing (the failures repeat in a crash loop backoff)
            logger.info(`waiting for inference workload ${workloadName} to be in phase Initializing`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Initializing, 120);

            // update the workload spec and wait for it to be ready
            workloadSpec.image = config.quickStartImage.inferenceHelloWorld;
            logger.info(`updating inference workload ${workloadName} with new spec`);
            await inferenceUtils.updateInferenceSpec(workloadId, workloadSpec);

            // wait for the workload to be in phase Running
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running);

            logger.info('waiting for workload additional field with the new revision name');
            await waitUtils.waitForWorkloadAdditionalField(
                workloadId,
                'latestReadyRevisionName',
                workloadName + '-' + '00002'
            );

            // validate get workload result - new spec and new revision names
            // the previous revision should still be the latest ready revision
            logger.info(`validating get workload result for workload ${workloadName}`);
            const workload = await validateBackendGetInferenceWorkloadResponse(
                workloadId,
                workloadSpec,
                workloadName,
                '00002',
                '00002'
            );

            // verify the workload is indeed serving requests
            logger.info(`sending inference request to workload ${workloadName}`);
            const clientWorkloadName = await inferenceUtils.sendInferenceRequest(
                project.status.namespace,
                inferenceUtils.getInternalUrlForInferenceWorkload(workload)
            );
            await inferenceUtils.deleteInferenceClientJob(clientWorkloadName, project.status.namespace);
        });

        test('updates a workload using compute fields', async function () {
            if (!versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_23)) {
                return;
            }

            // create the workload
            logger.info(
                `creating inference workload ${workloadName} in namespace ${project.status.namespace}`
            );
            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            // wait for the workload to be in phase Running
            logger.info(`waiting for inference workload ${workloadName} to be in phase Running`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 180);

            // update the workload spec with new compute fields
            updatedWorkloadSpec = {};
            updatedWorkloadSpec.compute = new WorkspaceSpecSpecAllOfCompute();
            updatedWorkloadSpec.compute.cpuCoreRequest = 0.2;
            updatedWorkloadSpec.compute.cpuMemoryRequest = '123M';
            logger.info(`updating inference workload ${workloadName} with new spec`);
            await inferenceUtils.updateInferenceSpec(workloadId, updatedWorkloadSpec);

            // wait for the workload to have an additional field with the new revision name,
            // this is the indication that update is in progress
            logger.info('waiting for workload additional field with the new revision name');
            await waitUtils.waitForWorkloadAdditionalField(
                workloadId,
                'latestReadyRevisionName',
                workloadName + '-' + '00002'
            );

            // validate get workload result - new spec and new revision names
            logger.info(`validating get workload result for workload ${workloadName}`);
            const workload = await validateBackendGetInferenceWorkloadResponse(
                workloadId,
                workloadSpec,
                workloadName,
                '00002',
                '00002'
            );

            // verify the new spec is applied using the supported workload API
            logger.info('verifying new spec is applied to the workload');
            expect(workload.workloadRequestedResources).to.not.be.undefined;
            expect(workload.workloadRequestedResources.cpu.request).to.equal(
                updatedWorkloadSpec.compute.cpuCoreRequest
            );
            expect(workload.workloadRequestedResources.cpuMemory.request).to.equal(
                updatedWorkloadSpec.compute.cpuMemoryRequest
            );

            // Verify the update reached the actual Knative revision, not only the backend representation.
            const revision = await waitUtils.waitForKnativeRevision(
                workloadName + '-00002',
                project.status.namespace,
                120
            );
            const revisionSpec = revision['spec'] as {
                containers: Array<{
                    resources?: { requests?: { cpu?: string; memory?: string } };
                }>;
            };
            const resourceRequests = revisionSpec.containers[0]?.resources?.requests;
            expect(resourceRequests, 'expected resource requests on the new Knative revision').to.not.be
                .undefined;
            expect(['0.2', '200m']).to.include(resourceRequests.cpu);
            expect(resourceRequests.memory).to.equal(updatedWorkloadSpec.compute.cpuMemoryRequest);
        });
    }
);
