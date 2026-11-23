/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    inferenceUtils,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsManagerUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import {
    createNgcAuthSecret,
    NGC_AUTH_SECRET_NAME
} from '../../../../../../modules/helpers/utils/nimServiceUtils';
import { workloadsManagerService } from '../../../../../../modules/api';
import { test, expect } from '@playwright/test';
import { WorkloadV2UpdateRequest } from '../../../../../../modules/api/apis/workloadsManagerApi/generated/model/workloadV2UpdateRequest';
import { WorkloadV2MetadataUpdateParams } from '../../../../../../modules/api/apis/workloadsManagerApi/generated/model/workloadV2MetadataUpdateParams';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

supportedFrom(SupportedVersions.V2_24)(function () {
    test.describe(
        `Workloads V2 Update API tests: ${tags.resource.workloadsV2} ${tags.tool.api} ${tags.teams.viper} ${tags.teams.loong}`,
        { tag: [tags.resource.workloadsV2] },
        function () {
            let project: Project;
            let namespace: string;
            let workloadName: string;
            let workloadId: string;

            test.beforeAll(async function () {
                workloadsManagerService.workloadsV2Api.accessToken =
                    flowRootState.users.system_administrator.idToken;

                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('project'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 2, limit: 2 } }]
                });
                namespace = project.status.namespace;
                await createNgcAuthSecret(namespace);
                logger.info(
                    `Created project: ${project.name} (ID: ${project.id}) in namespace: ${namespace} with NGC secret`
                );
                workloadsManagerUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            });

            test.afterAll(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id, 120);
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and project ${project.name} namespace ${project.status.namespace}; error ${error}`
                    );
                }
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateShortNameByPrefix('wl-v2-update');
            });

            test.afterEach(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
                } catch (error) {
                    logger.error(`Failed to delete workloads of project ${project.name}; error: ${error}`);
                }
            });

            test('Update Workload v2 Manifest Directly', async function () {
                const initialReplicas = 1;
                const updatedReplicas = 2;
                const updatedImage = config.quickStartImage.inferenceAutoScale;

                await test.step('Create a workload v2 directly via Workloads v2 API', async () => {
                    workloadId = await workloadsManagerUtils.createDeploymentWorkloadV2({
                        name: workloadName,
                        projectId: String(project.id),
                        replicas: initialReplicas,
                        useStandardManifest: true
                    });
                });

                await test.step('Get workload ID and record initial manifest', async () => {
                    const getResponse =
                        await workloadsManagerService.workloadsV2Api.getWorkloadV2ById(workloadId);
                    expect(getResponse.response.statusCode).toBe(200);

                    const workload = getResponse.body.workload;
                    workloadsManagerUtils.validateDeploymentManifestStructure(workload);

                    // Validate actual values
                    expect(workload.workloadManifest.spec.replicas).toBe(initialReplicas);
                    expect(workload.workloadManifest.spec.template.spec.containers[0].image).toBe('busybox');
                    logger.info(`Initial manifest: replicas=${initialReplicas}, image=busybox`);
                });

                await test.step('Wait for initial workload to reach Running state', async () => {
                    logger.info(
                        `Waiting for workload ${workloadId} to reach Running state before updating...`
                    );
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                    logger.info(`Workload ${workloadId} is now Running`);
                });

                await test.step('Update workload manifest via PUT API', async () => {
                    const updatedManifest = workloadsManagerUtils.createDeploymentManifest(
                        workloadName,
                        updatedReplicas,
                        updatedImage
                    );

                    const updateRequest: WorkloadV2UpdateRequest = {
                        manifest: updatedManifest
                    };

                    const updateResponse = await workloadsManagerService.workloadsV2Api.updateWorkloadV2(
                        workloadId,
                        updateRequest
                    );

                    expect(updateResponse.response.statusCode).toBe(202);

                    workloadsManagerUtils.validateWorkloadV2UpdateResponse(
                        updateResponse.body,
                        workloadId,
                        workloadName,
                        String(project.id)
                    );

                    logger.info(
                        `Updated workload ${workloadId} with image=${updatedImage} (replicas unchanged: ${updatedReplicas})`
                    );
                });

                await test.step('Verify workload manifest updated in control-plane', async () => {
                    const getResponse =
                        await workloadsManagerService.workloadsV2Api.getWorkloadV2ById(workloadId);
                    expect(getResponse.response.statusCode).toBe(200);

                    const workload = getResponse.body.workload;
                    workloadsManagerUtils.validateDeploymentManifestStructure(workload);

                    // Validate actual values
                    expect(workload.workloadManifest.spec.replicas).toBe(updatedReplicas);
                    expect(workload.workloadManifest.spec.template.spec.containers[0].image).toBe(
                        updatedImage
                    );
                    logger.info(
                        `Verified updated manifest in control-plane: image=${updatedImage} (replicas unchanged: ${updatedReplicas})`
                    );
                });

                await test.step('Verify workload manifest and data synced in cluster', async () => {
                    logger.info(
                        `Verifying workload ${workloadId} update synced to cluster (image=${updatedImage}, replicas=${updatedReplicas})...`
                    );

                    // Wait for workload phase and image to be updated
                    const workload = await waitUtils.waitForWorkloadPhaseAndImage(
                        workloadId,
                        String(project.id),
                        Phase.Running,
                        updatedImage,
                        120,
                        5
                    );

                    expect(workload, 'Workload should be found').toBeDefined();
                    expect(workload.phase).toBe(Phase.Running);

                    const actualImage = workload.images[0];
                    expect(
                        actualImage === updatedImage || actualImage.startsWith(updatedImage + '@'),
                        `Expected image to be ${updatedImage} (or with digest), but got ${actualImage}`
                    ).toBeTruthy();

                    expect(workload.requestedPods?.number).toBe(updatedReplicas);
                    expect(workload.runningPods).toBe(updatedReplicas);

                    logger.info(
                        `Workload updated in cluster: phase=${workload.phase}, image=${actualImage}, replicas=${workload.requestedPods?.number}, runningPods=${workload.runningPods}`
                    );
                });
            });

            test('Update with Different GVK (Negative Test)', async function () {
                const initialReplicas = 2;
                const initialImage = 'nginx:1.21';
                let updateFailed = false;
                let capturedError: any;

                await test.step('Create a workload v2 with Deployment manifest', async () => {
                    workloadId = await workloadsManagerUtils.createDeploymentWorkloadV2({
                        name: workloadName,
                        projectId: String(project.id),
                        replicas: initialReplicas,
                        image: initialImage,
                        containerName: 'nginx',
                        containerPort: 80
                    });
                });

                await test.step('Verify workload created with GVK: apps/v1, Deployment', async () => {
                    const getResponse =
                        await workloadsManagerService.workloadsV2Api.getWorkloadV2ById(workloadId);
                    expect(getResponse.response.statusCode).toBe(200);

                    const workload = getResponse.body.workload;
                    workloadsManagerUtils.validateDeploymentManifestStructure(workload);

                    // Validate actual values
                    expect(workload.workloadManifest.apiVersion).toBe('apps/v1');
                    expect(workload.workloadManifest.kind).toBe('Deployment');
                    expect(workload.workloadManifest.spec.replicas).toBe(initialReplicas);
                    expect(workload.workloadManifest.spec.template.spec.containers[0].image).toBe(
                        initialImage
                    );
                    logger.info(
                        `Verified workload created with GVK: ${workload.workloadManifest.apiVersion}, ${workload.workloadManifest.kind}`
                    );
                });

                await test.step('Attempt to update with different GVK (NIMService)', async () => {
                    const nimServiceManifest = workloadsManagerUtils.createNIMServiceManifest(
                        'test-nim-service',
                        config.quickStartImage.nim,
                        1,
                        NGC_AUTH_SECRET_NAME
                    );

                    const updateRequest: WorkloadV2UpdateRequest = {
                        manifest: nimServiceManifest
                    };

                    try {
                        await workloadsManagerService.workloadsV2Api.updateWorkloadV2(
                            workloadId,
                            updateRequest
                        );
                    } catch (error) {
                        updateFailed = true;
                        capturedError = error;
                    }
                    expect(updateFailed).toBe(true);
                    expect(capturedError).toBeDefined();
                    expect(capturedError.statusCode).toBe(400);
                    logger.info(`Update correctly rejected with status code: ${capturedError.statusCode}`);

                    const errorMessage =
                        capturedError.message ||
                        capturedError.body?.message ||
                        JSON.stringify(capturedError.body);
                    logger.info(`Error message: ${errorMessage}`);

                    const errorMessageLower = errorMessage.toLowerCase();
                    const containsGvkInfo =
                        errorMessageLower.includes('gvk') && errorMessageLower.includes('does not match');
                    expect(containsGvkInfo).toBe(true);
                    logger.info('Error message correctly indicates GVK mismatch');
                });

                await test.step('Verify original Deployment workload remains unchanged', async () => {
                    const getResponse =
                        await workloadsManagerService.workloadsV2Api.getWorkloadV2ById(workloadId);
                    expect(getResponse.response.statusCode).toBe(200);

                    const workload = getResponse.body.workload;
                    workloadsManagerUtils.validateDeploymentManifestStructure(workload);

                    // Validate actual values
                    expect(workload.workloadManifest.apiVersion).toBe('apps/v1');
                    expect(workload.workloadManifest.kind).toBe('Deployment');
                    expect(workload.workloadManifest.spec.replicas).toBe(initialReplicas);
                    expect(workload.workloadManifest.spec.template.spec.containers[0].image).toBe(
                        initialImage
                    );
                    logger.info('Verified original workload remains unchanged after rejected update');
                });
            });

            test('Update Priority via Workloads v2 API', async function () {
                const initialPriority = 'medium';
                const updatedPriority = 'high';
                const initialReplicas = 1;

                await test.step('Create workload v2 with Deployment manifest and initial priority', async () => {
                    workloadId = await workloadsManagerUtils.createDeploymentWorkloadV2({
                        name: workloadName,
                        projectId: String(project.id),
                        replicas: initialReplicas,
                        priority: initialPriority,
                        useStandardManifest: true
                    });
                });

                await test.step('Wait for workload to reach Running state', async () => {
                    logger.info(
                        `Waiting for workload ${workloadId} to reach Running state before updating priority...`
                    );
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);
                    logger.info(`Workload ${workloadId} is now Running`);
                });

                await test.step('Update priority via workloads v2 API', async () => {
                    const manifest = workloadsManagerUtils.createStandardManifest(
                        workloadName,
                        initialReplicas
                    );
                    const updateResponse = await workloadsManagerUtils.updateWorkloadV2Metadata(
                        workloadId,
                        manifest,
                        {
                            priority: updatedPriority
                        }
                    );
                    expect(updateResponse.response.statusCode).toBe(202);
                    logger.info(`Priority update API call accepted (202): ${updatedPriority}`);
                });

                await test.step('Verify priority updated in control-plane', async () => {
                    const getResponse =
                        await workloadsManagerService.workloadsV2Api.getWorkloadV2ById(workloadId);
                    expect(getResponse.response.statusCode).toBe(200);
                    expect(getResponse.body.workload.metadata.priority).toBe(updatedPriority);
                    logger.info(`Verified priority in control-plane: ${updatedPriority}`);
                });

                await test.step('Verify priority updated in cluster', async () => {
                    logger.info(
                        `Waiting for workload to have priority ${updatedPriority} synced from cluster`
                    );
                    await waitUtils.waitForBackendWorkloadByIdAndPriority(workloadId, updatedPriority, 120);
                });
                logger.info('Priority update test completed successfully!');
            });

            test('Update Category via Workloads v2 API', async function () {
                const initialCategory = 'Train';
                const updatedCategory = 'Deploy';
                const initialReplicas = 1;

                await test.step('Create workload v2 with Deployment manifest and initial category', async () => {
                    workloadId = await workloadsManagerUtils.createDeploymentWorkloadV2({
                        name: workloadName,
                        projectId: String(project.id),
                        replicas: initialReplicas,
                        category: initialCategory,
                        useStandardManifest: true
                    });
                });

                await test.step('Update category via workloads v2 API', async () => {
                    const manifest = workloadsManagerUtils.createStandardManifest(
                        workloadName,
                        initialReplicas
                    );
                    const updateResponse = await workloadsManagerUtils.updateWorkloadV2Metadata(
                        workloadId,
                        manifest,
                        {
                            category: updatedCategory
                        }
                    );
                    expect(updateResponse.response.statusCode).toBe(202);
                    logger.info(`Category update API call accepted (202): ${updatedCategory}`);
                });

                await test.step('Verify category updated in control-plane', async () => {
                    const getResponse =
                        await workloadsManagerService.workloadsV2Api.getWorkloadV2ById(workloadId);
                    expect(getResponse.response.statusCode).toBe(200);
                    expect(getResponse.body.workload.metadata.category).toBe(updatedCategory);
                    logger.info(`Verified category in control-plane: ${updatedCategory}`);
                });

                await test.step('Verify category updated in cluster', async () => {
                    logger.info(
                        `Waiting for workload to have category ${updatedCategory} synced from cluster`
                    );
                    await waitUtils.waitForBackendWorkloadByIdAndCategory(workloadId, updatedCategory, 120);
                });
                logger.info('Category update test completed successfully!');
            });

            test('Update Non-V2 Workload (Negative Test)', async function () {
                let capturedError: any;
                const v1InferenceName = objectGenerator.generateShortNameByPrefix('v1-inference');
                const attemptedPriority = 'high';

                await test.step('Create a runai inference workload (v1)', async () => {
                    const inferenceSpec = {
                        image: config.quickStartImage.inferenceHelloWorld,
                        compute: {
                            gpuDevicesRequest: 1
                        },
                        servingPort: {
                            container: 8080
                        }
                    };

                    const [id, name] = await inferenceUtils.createInference(
                        v1InferenceName,
                        inferenceSpec,
                        String(project.id),
                        false,
                        flowRootState.users.system_administrator.accessToken
                    );

                    workloadId = id;
                    workloadName = name;
                    logger.info(`Created v1 inference workload: ${workloadName} with ID: ${workloadId}`);
                });

                await test.step('Get workload ID from the created inference workload', () => {
                    expect(workloadId).toBeDefined();
                    expect(workloadId).not.toBe('');
                    logger.info(`V1 Workload ID: ${workloadId}`);
                });

                await test.step('Attempt to update non-v2 workload via Workloads v2 API', async () => {
                    const v1Workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(v1Workload).toBeDefined();
                    expect(v1Workload.body).toBeDefined();
                    logger.info('Retrieved v1 workload via v1 API successfully');

                    const minimalManifest = workloadsManagerUtils.createDeploymentManifest(
                        workloadName,
                        1,
                        config.quickStartImage.inferenceHelloWorld,
                        'inference'
                    );

                    const metadataUpdate: WorkloadV2MetadataUpdateParams = {
                        priority: attemptedPriority
                    };

                    const updateRequest: WorkloadV2UpdateRequest = {
                        metadata: metadataUpdate,
                        manifest: minimalManifest
                    };

                    try {
                        await workloadsManagerService.workloadsV2Api.updateWorkloadV2(
                            workloadId,
                            updateRequest
                        );
                        throw new Error('v1 workload should not be updatable via Workloads v2 API');
                    } catch (error) {
                        if (error.message === 'v1 workload should not be updatable via Workloads v2 API') {
                            throw error;
                        }
                        capturedError = error;
                    }

                    expect(capturedError).toBeDefined();
                    // Authorization may deliberately conceal a workload that is outside the V2 store
                    // with 403. All accepted statuses prove the V2 update was rejected.
                    expect([400, 403, 404]).toContain(capturedError.statusCode);
                    logger.info(
                        `v1 workload correctly rejected by v2 UPDATE API, status code: ${capturedError.statusCode}`
                    );

                    const errorMessage =
                        capturedError.message ||
                        capturedError.body?.message ||
                        JSON.stringify(capturedError.body);
                    logger.info(`Error message: ${errorMessage}`);
                });

                await test.step('Verify original inference workload remains unchanged', async () => {
                    const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workload).toBeDefined();
                    expect(workload.body).toBeDefined();
                    expect(workload.body.id).toBe(workloadId);
                    expect(workload.body.name).toBe(workloadName);
                    logger.info('Verified original v1 workload remains unchanged');
                });
                logger.info('Update non-v2 workload negative test completed successfully!');
            });
        }
    );
});
