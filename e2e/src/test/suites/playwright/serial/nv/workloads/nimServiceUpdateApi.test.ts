/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
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
    workloadsUtilsV2,
    workloadUtils
} from '../../../../../../modules/helpers';
import {
    createAndWaitForNimService,
    createDefaultNimServiceSpec,
    createNgcAuthSecret,
    createNimService,
    executeWithTokenSwitch,
    getNimServiceById,
    NIM_TEST_TIMEOUT_MS,
    NIM_UPDATE_SYNC_TEST_TIMEOUT_MS,
    NIM_UPDATE_TEST_TIMEOUT_MS,
    updateNimServiceSpec,
    validateNimServiceUpdateResponse,
    verifyErrorResponse,
    verifyNimServiceFieldUnchanged
} from '../../../../../../modules/helpers/utils/nimServiceUtils';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';
import { isNimOperatorInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';

import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { NimServiceSpec } from '../../../../../../modules/api/apis/workloadsManagerApi/generated/model/nimServiceSpec';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, TestInfo } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';
import { randomBytes } from 'crypto';
import { workloadsManagerService } from '../../../../../../modules/api';
import { NimServiceRemovableFields } from '../../../../../../modules/api/apis/workloadsManagerApi/generated/model/nimServiceRemovableFields';

test.describe(
    `nimservice update api tests, tags: ${tags.teams.viper} ${tags.teams.loong}`,
    function () {
        supportedFrom(SupportedVersions.V2_24)(function () {
            let project: Project;
            let workloadName: string;
            let initialImage: string;
            let nimServiceSpec: NimServiceSpec;
            let workloadId: string;

            test.beforeAll(async function () {
                initialImage = config.quickStartImage.nim;
                if (!(await isNimOperatorInstalled())) {
                    logger.warn('k8s-nim-operator is not installed on the cluster, skipping test');
                    test.skip();
                }

                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('project'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 2, limit: 2 } }]
                });
                await createNgcAuthSecret(project.status.namespace);
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateNameByPrefix('workload');

                nimServiceSpec = createDefaultNimServiceSpec();

                workloadsManagerUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            });

            test.afterAll(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch {
                    logger.error(
                        `Failed to delete workloads and project ${project.name} namespace ${project.status.namespace}`
                    );
                }
            });

            test.afterEach(async function () {
                if (project === undefined) {
                    return;
                }

                // Fall back to the heavy CRD+pod cleanup only if the API delete failed,
                // to avoid firing the admission-controller webhook twice on the happy path.
                let apiDeleteOk = true;
                try {
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(String(project.id));
                    logger.info(`Deleted all workloads in project ${project.id} via API`);
                } catch (error) {
                    apiDeleteOk = false;
                    logger.warn(`Failed to delete workloads via API for project ${project.id}: ${error}`);
                }

                if (!apiDeleteOk) {
                    try {
                        await workloadUtils.clearProjectWorkloadsAndPods(project.name, 120);
                    } catch (error) {
                        logger.error(
                            `Failed to delete workloads and pods of project ${project.name} namespace ${project.status.namespace}: ${error}`
                        );
                    }
                }
            });

            test('Update a NIM service with a single field', async function () {
                test.setTimeout(NIM_UPDATE_TEST_TIMEOUT_MS);
                const initialReplicas = 1;
                const updatedReplicas = 2;

                await test.step('Create NIM service with initial replicas and wait for Running', async () => {
                    nimServiceSpec.image = initialImage;
                    nimServiceSpec.replicas = initialReplicas;
                    workloadId = await createAndWaitForNimService(workloadName, project, nimServiceSpec);
                    logger.info(`Created NIM service with ID: ${workloadId}, replicas: ${initialReplicas}`);
                });

                await test.step('Record initial workload state', async () => {
                    const getResponse = await getNimServiceById(workloadId);
                    expect(getResponse.spec.replicas).to.equal(initialReplicas);

                    const workloadResponse = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workloadResponse.body.requestedPods.number).to.equal(initialReplicas);
                    expect(workloadResponse.body.runningPods).to.equal(initialReplicas);

                    logger.info('Verified initial replicas in both NIM service and workload APIs');
                });

                await test.step('Update replicas via PATCH API', async () => {
                    const updateResponse = await updateNimServiceSpec(workloadId, {
                        replicas: updatedReplicas
                    });
                    expect(updateResponse.body.spec.replicas).to.equal(updatedReplicas);
                    logger.info(`Successfully updated NIM service replicas to: ${updatedReplicas}`);
                });

                await test.step('Verify updated spec via GET NIM service API', async () => {
                    const getResponse = await getNimServiceById(workloadId);
                    expect(getResponse).to.not.be.undefined;
                    expect(getResponse.spec).to.not.be.undefined;
                    expect(getResponse.spec.replicas).to.equal(updatedReplicas);
                    logger.info(`Verified spec.replicas = ${updatedReplicas} in NIM service GET response`);
                });

                await test.step('Verify workload manifest and data synced in cluster', async () => {
                    logger.info(
                        `Verifying workload ${workloadId} update synced to cluster (replicas=${updatedReplicas})...`
                    );

                    await waitUtils.waitForCondition(
                        async () => {
                            const workloadResponse =
                                await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                            const requestedReplicas = workloadResponse.body.requestedPods?.number || 0;
                            const runningReplicas = workloadResponse.body.runningPods || 0;
                            const phase = workloadResponse.body.phase;
                            return (
                                requestedReplicas === updatedReplicas &&
                                runningReplicas === updatedReplicas &&
                                phase === Phase.Running
                            );
                        },
                        `workload ${workloadId} to run with ${updatedReplicas} replicas`,
                        true,
                        120,
                        5
                    );

                    const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workload.body.type).to.equal('NIMService');
                    expect(workload.body.phase).to.equal(Phase.Running);
                    expect(workload.body.requestedPods.number).to.equal(updatedReplicas);
                    expect(workload.body.runningPods).to.equal(updatedReplicas);

                    logger.info(
                        `Workload updated in cluster: type=${workload.body.type}, phase=${workload.body.phase}, requestedPods=${workload.body.requestedPods.number}, runningPods=${workload.body.runningPods}`
                    );
                    logger.info('NIM service replicas update test completed successfully!');
                });
            });

            test('Update a NIM service with multiple compute fields simultaneously', async function () {
                test.setTimeout(NIM_UPDATE_SYNC_TEST_TIMEOUT_MS);
                const initialCpuRequest = 0.2;
                const initialMemoryRequest = '50M';
                const updatedCpuRequest = 0.4;
                const updatedCpuLimit = 0.8;
                const updatedMemoryRequest = '100M';

                await test.step('Create NIM service with initial configuration and wait for Running', async () => {
                    nimServiceSpec.image = initialImage;
                    nimServiceSpec.compute = {
                        cpuCoreRequest: initialCpuRequest,
                        cpuMemoryRequest: initialMemoryRequest
                    };
                    nimServiceSpec.replicas = 1;

                    workloadId = await createAndWaitForNimService(workloadName, project, nimServiceSpec);
                    logger.info(
                        `Created NIM service with ID: ${workloadId}, cpu: ${initialCpuRequest}, memory: ${initialMemoryRequest}`
                    );
                });

                await test.step('Record initial workload state', async () => {
                    const getResponse = await getNimServiceById(workloadId);
                    expect(getResponse.spec.compute.cpuCoreRequest).to.equal(initialCpuRequest);
                    expect(getResponse.spec.compute.cpuMemoryRequest).to.equal(initialMemoryRequest);
                    const workloadResponse = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workloadResponse.body.workloadRequestedResources.cpu.request).to.be.closeTo(
                        initialCpuRequest,
                        0.0001
                    );
                    logger.info('Verified initial configuration in both NIM service and workload APIs');
                });

                await test.step('Wait for workload to reach Running phase', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                    logger.info(`Workload ${workloadId} reached Running phase`);
                });

                await test.step('Update multiple compute fields simultaneously via PATCH API', async () => {
                    const updateResponse = await updateNimServiceSpec(workloadId, {
                        compute: {
                            cpuCoreRequest: updatedCpuRequest,
                            cpuCoreLimit: updatedCpuLimit,
                            cpuMemoryRequest: updatedMemoryRequest
                        }
                    });

                    validateNimServiceUpdateResponse(
                        updateResponse.body,
                        workloadId,
                        workloadName,
                        String(project.id)
                    );

                    expect(updateResponse.body.spec.compute.cpuCoreRequest).to.equal(updatedCpuRequest);
                    expect(updateResponse.body.spec.compute.cpuCoreLimit).to.equal(updatedCpuLimit);
                    logger.info(
                        `Successfully updated NIM service - cpu: ${updatedCpuRequest}, cpuLimit: ${updatedCpuLimit}, memory: ${updatedMemoryRequest}`
                    );
                });

                await test.step('Verify updated spec stored in control-plane via GET NIM service API', async () => {
                    const getResponse = await getNimServiceById(workloadId);
                    expect(getResponse).to.not.be.undefined;
                    expect(getResponse.spec).to.not.be.undefined;
                    expect(getResponse.spec.compute.cpuCoreRequest).to.equal(updatedCpuRequest);
                    expect(getResponse.spec.compute.cpuCoreLimit).to.equal(updatedCpuLimit);
                    logger.info(
                        `Verified updated spec in control-plane: cpu=${updatedCpuRequest}, cpuLimit=${updatedCpuLimit}, memory=${updatedMemoryRequest}`
                    );
                });

                await test.step('Verify workload manifest and data synced in cluster', async () => {
                    logger.info(
                        `Verifying workload ${workloadId} update synced to cluster (cpu=${updatedCpuRequest}, memory=${updatedMemoryRequest})...`
                    );

                    await waitUtils.waitForCondition(
                        async () => {
                            const workloadResponse =
                                await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                            const workload = workloadResponse.body;
                            const actualCpu = workload.workloadRequestedResources?.cpu?.request || 0;
                            const phase = workload.phase;
                            const runningPods = workload.runningPods || 0;
                            return (
                                phase === Phase.Running &&
                                runningPods === nimServiceSpec.replicas &&
                                Math.abs(actualCpu - updatedCpuRequest) < 0.0001
                            );
                        },
                        `workload ${workloadId} to sync updated CPU resources (${updatedCpuRequest}) and reach Running state`,
                        true,
                        180,
                        5
                    );

                    const workloadResponse = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    const workload = workloadResponse.body;

                    expect(workload, 'Workload should be found').to.not.be.undefined;
                    expect(workload.type).to.equal('NIMService');
                    expect(workload.phase).to.equal(Phase.Running);
                    expect(workload.runningPods).to.equal(nimServiceSpec.replicas);

                    const actualCpu = workload.workloadRequestedResources?.cpu?.request || 0;
                    expect(
                        Math.abs(actualCpu - updatedCpuRequest) < 0.0001,
                        `Expected CPU request to be ${updatedCpuRequest}, but got ${actualCpu}`
                    ).to.be.true;
                    logger.info(
                        `Workload updated in cluster: phase=${workload.phase}, cpu=${actualCpu}, runningPods=${workload.runningPods}`
                    );
                });
                logger.info('NIM service multiple fields update test completed successfully!');
            });

            test('Updates existing environment variables', async function () {
                test.setTimeout(NIM_UPDATE_TEST_TIMEOUT_MS);
                const initialEnvVars = [{ name: 'LOG_LEVEL', value: 'info' }];
                const updatedEnvVars = [{ name: 'LOG_LEVEL', value: 'debug' }];

                await test.step('Create NIM service with initial environment variable', async () => {
                    nimServiceSpec.environmentVariables = initialEnvVars;

                    const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                    workloadId = createResponse.metadata.id;
                    expect(createResponse.spec.environmentVariables).to.not.be.undefined;
                    expect(createResponse.spec.environmentVariables).to.have.lengthOf(1);
                    expect(createResponse.spec.environmentVariables[0].name).to.equal('LOG_LEVEL');
                    expect(createResponse.spec.environmentVariables[0].value).to.equal('info');
                    logger.info(`Created NIM service with ID: ${workloadId}, LOG_LEVEL=info`);
                });

                await test.step('Wait for workload to reach Running phase', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                    logger.info(`Workload ${workloadId} reached Running phase`);
                });

                await test.step('Update LOG_LEVEL environment variable to debug', async () => {
                    const updateResponse = await updateNimServiceSpec(workloadId, {
                        environmentVariables: updatedEnvVars
                    });
                    expect(updateResponse.body.spec.environmentVariables).to.have.lengthOf(1);
                    expect(updateResponse.body.spec.environmentVariables[0].value).to.equal('debug');
                    logger.info('Successfully updated LOG_LEVEL to debug');
                });

                await test.step('Wait for workload to restart with new environment variable', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                    logger.info(`Workload ${workloadId} restarted and is Running`);
                });

                await test.step('Verify updated spec via GET NIM service API', async () => {
                    const getResponse = await getNimServiceById(workloadId);

                    expect(getResponse.spec.environmentVariables).to.not.be.undefined;
                    expect(getResponse.spec.environmentVariables).to.have.lengthOf(1);
                    expect(getResponse.spec.environmentVariables[0].name).to.equal('LOG_LEVEL');
                    expect(getResponse.spec.environmentVariables[0].value).to.equal('debug');
                    logger.info('Verified LOG_LEVEL value is debug in control-plane');
                });
                logger.info('Environment variable update test completed successfully!');
            });

            test('Removes environment variables', async function () {
                test.setTimeout(NIM_UPDATE_TEST_TIMEOUT_MS);
                const initialEnvVars = [
                    { name: 'LOG_LEVEL', value: 'info' },
                    { name: 'DEBUG_MODE', value: 'true' }
                ];

                await test.step('Create NIM service with 2 environment variables', async () => {
                    nimServiceSpec.environmentVariables = initialEnvVars;

                    const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                    workloadId = createResponse.metadata.id;
                    expect(createResponse.spec.environmentVariables).to.not.be.undefined;
                    expect(createResponse.spec.environmentVariables).to.have.lengthOf(2);
                    logger.info(`Created NIM service with ID: ${workloadId}, with 2 environment variables`);
                });

                await test.step('Wait for workload to reach Running phase', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                    logger.info(`Workload ${workloadId} reached Running phase`);
                });

                await test.step('Update to remove environment variables', async () => {
                    await updateNimServiceSpec(workloadId, { environmentVariables: [] });
                    logger.info('Successfully called PATCH API to remove environment variables');
                });

                await test.step('Wait for workload to restart after removing environment variables', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                    logger.info(`Workload ${workloadId} restarted and is Running`);
                });

                await test.step('Verify updated spec via GET NIM service API', async () => {
                    const getResponse = await getNimServiceById(workloadId);
                    const envVars = getResponse.spec.environmentVariables;

                    if (envVars !== undefined) {
                        expect(envVars).to.be.empty;
                        logger.info('Verified environmentVariables is empty array in control-plane');
                    } else {
                        expect(envVars).to.be.undefined;
                        logger.info('Verified environmentVariables is undefined (removed) in control-plane');
                    }
                });
                logger.info('Environment variables removal test completed successfully!');
            });

            test('Disable autoscaling on NIM service', async ({}, testInfo: TestInfo) => {
                test.setTimeout(NIM_UPDATE_TEST_TIMEOUT_MS);
                const minReplicas = 2;
                const maxReplicas = 3;
                const fixedReplicas = 5;

                await test.step('Create NIM service with autoscaling enabled', async () => {
                    nimServiceSpec.autoscaling = {
                        minReplicas: minReplicas,
                        maxReplicas: maxReplicas,
                        metric: 'cpu',
                        metricThreshold: 50
                    };
                    delete nimServiceSpec.replicas;

                    const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                    workloadId = createResponse.metadata.id;
                    expect(createResponse.spec.autoscaling).to.not.be.undefined;
                    expect(createResponse.spec.autoscaling.minReplicas).to.equal(minReplicas);
                    expect(createResponse.spec.autoscaling.maxReplicas).to.equal(maxReplicas);
                    logger.info(
                        `Created NIM service with ID: ${workloadId}, autoscaling enabled (min=${minReplicas}, max=${maxReplicas})`
                    );
                });

                await test.step('Wait for workload to reach Running phase', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                    logger.info(`Workload ${workloadId} reached Running phase`);
                });

                await test.step('Disable autoscaling by updating to fixed replicas', async () => {
                    await updateNimServiceSpec(
                        workloadId,
                        { replicas: fixedReplicas },
                        StatusCodes.ACCEPTED,
                        [NimServiceRemovableFields.Autoscaling]
                    );
                    logger.info(
                        `Successfully called PATCH API to disable autoscaling and set replicas to ${fixedReplicas}`
                    );
                });

                await test.step('Verify autoscaling is removed and fixed replicas are set', async () => {
                    const getResponse = await getNimServiceById(workloadId);

                    expect(getResponse.spec.autoscaling).to.satisfy(
                        (value: any) => value === null || value === undefined,
                        'autoscaling should be null or undefined (disabled)'
                    );
                    logger.info(`Verified autoscaling is disabled (value: ${getResponse.spec.autoscaling})`);
                    expect(getResponse.spec.replicas).to.equal(fixedReplicas);
                    logger.info(`Verified fixed replicas set to ${fixedReplicas}`);
                });

                await test.step('Verify workload runs with fixed replicas', async () => {
                    await waitUtils.waitForCondition(
                        async () => {
                            const workloadResponse =
                                await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                            const requestedReplicas = workloadResponse.body.requestedPods?.number || 0;
                            const runningReplicas = workloadResponse.body.runningPods || 0;
                            const phase = workloadResponse.body.phase;
                            return (
                                requestedReplicas === fixedReplicas &&
                                runningReplicas === fixedReplicas &&
                                phase === Phase.Running
                            );
                        },
                        `workload ${workloadId} to run with ${fixedReplicas} fixed replicas`,
                        true,
                        120,
                        5
                    );

                    const workloadResponse = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workloadResponse.body.requestedPods.number).to.equal(fixedReplicas);
                    expect(workloadResponse.body.runningPods).to.equal(fixedReplicas);
                    logger.info(`Verified workload runs with ${fixedReplicas} fixed replicas`);
                });
                logger.info('Autoscaling disable test completed successfully!');
            });

            test('UPDATE with invalid spec - API validation error (conflicting replicas and autoscaling)', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                const fixedReplicas = 2;

                await test.step('Create NIM service with fixed replicas and wait for Running', async () => {
                    nimServiceSpec.replicas = fixedReplicas;
                    delete nimServiceSpec.autoscaling;
                    workloadId = await createAndWaitForNimService(workloadName, project, nimServiceSpec);
                    logger.info(
                        `Created NIM service with ID: ${workloadId}, fixed replicas: ${fixedReplicas}`
                    );
                });

                await test.step('Attempt to update with conflicting configuration', async () => {
                    const invalidUpdateRequest = {
                        spec: {
                            replicas: 3,
                            autoscaling: {
                                minReplicas: 1,
                                maxReplicas: 5,
                                metric: 'cpu',
                                metricThreshold: 70
                            }
                        }
                    };

                    try {
                        await workloadsManagerService.nimServicesApi.updateNimServiceSpec(
                            workloadId,
                            invalidUpdateRequest
                        );
                        throw new Error('Expected 400 Bad Request, but update succeeded');
                    } catch (error) {
                        verifyErrorResponse(
                            error,
                            StatusCodes.BAD_REQUEST,
                            ['replica', 'autoscal', 'conflict', 'cannot specify both'],
                            'validation failure'
                        );
                    }
                });

                await test.step('Verify workload unchanged', async () => {
                    await verifyNimServiceFieldUnchanged(workloadId, 'spec.replicas', fixedReplicas);

                    const getResponse = await getNimServiceById(workloadId);
                    if (getResponse.spec.autoscaling !== undefined) {
                        expect(Object.keys(getResponse.spec.autoscaling)).to.be.empty;
                    }
                    logger.info('Verified autoscaling not enabled');
                });
                logger.info('Invalid spec API validation test completed successfully!');
            });

            test('Update with invalid spec - Operator validation error (ingress URL conflict)', async ({}, testInfo: TestInfo) => {
                test.setTimeout(NIM_UPDATE_SYNC_TEST_TIMEOUT_MS);
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'NIM service workload not transitioning to Degraded after invalid spec update on BC 2.24 (PodGrouper StorageError on podgroups.scheduling.run.ai)'
                );
                const sharedIngressUrl =
                    process.env.CONTROL_PLANE_URL + '/nim-shared-' + randomBytes(2).toString('hex');
                const service2IngressUrl =
                    process.env.CONTROL_PLANE_URL + '/nim-service-2-' + randomBytes(2).toString('hex');
                let workloadId1: string;
                let workloadId2: string;
                const workloadName1 = objectGenerator.generateNameByPrefix('nim-service-1');
                const workloadName2 = objectGenerator.generateNameByPrefix('nim-service-2');

                await test.step('Setup two NIM services with distinct ingress URLs', async () => {
                    const nimServiceSpec1 = { ...nimServiceSpec };
                    nimServiceSpec1.servingPort = {
                        port: 8000,
                        exposeExternally: true,
                        exposedUrl: sharedIngressUrl
                    };

                    const nimServiceSpec2 = { ...nimServiceSpec };
                    nimServiceSpec2.servingPort = {
                        port: 8000,
                        exposeExternally: true,
                        exposedUrl: service2IngressUrl
                    };

                    const [id1, id2] = await Promise.all([
                        createAndWaitForNimService(workloadName1, project, nimServiceSpec1),
                        createAndWaitForNimService(workloadName2, project, nimServiceSpec2)
                    ]);

                    workloadId1 = id1;
                    workloadId2 = id2;
                    workloadId = workloadId2;

                    const [getResponse1, getResponse2] = await Promise.all([
                        getNimServiceById(workloadId1),
                        getNimServiceById(workloadId2)
                    ]);

                    expect(getResponse1.spec.servingPort.exposedUrl).to.equal(sharedIngressUrl);
                    expect(getResponse2.spec.servingPort.exposedUrl).to.equal(service2IngressUrl);
                    logger.info(
                        `Created nim-service-1 (${workloadId1}) and nim-service-2 (${workloadId2}) in parallel`
                    );
                });

                await test.step('Attempt to UPDATE nim-service-2 to use same URL as nim-service-1', async () => {
                    const updateResponse = 
                    await updateNimServiceSpec(workloadId2, {
                        servingPort: {
                            port: 8000,
                            exposeExternally: true,
                            exposedUrl: sharedIngressUrl
                        }
                    });
                    logger.info(
                        `Update API returned ${updateResponse.response.statusCode}, monitoring for operator validation error...`
                    );
                });

                await test.step('Wait for workload to enter degraded state', async () => {
                    logger.info(`Waiting for workload ${workloadId2} to enter degraded state...`);

                    const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                        workloadId2,
                        Phase.Degraded,
                        120
                    );
                    expect(
                        workload.phaseMessage.trim(),
                        'BUG: Error condition detected but message is empty. Expected detailed error message about ingress URL conflict.'
                    ).to.not.be.empty;

                    logger.info(
                        `Verified workload entered degraded state - phase: ${workload.phase}, reason: ${workload.phaseReason}, message: ${workload.phaseMessage}`
                    );
                });

                await test.step('Verify nim-service-1 remains unaffected', async () => {
                    const workload1Response = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId1);
                    expect(workload1Response.body.phase).to.equal(Phase.Running);
                });
                logger.info('Operator validation error test completed!');
            });

            test('Update non-existent NIM service', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                await test.step('Attempt to update non-existent workload ID', async () => {
                    const nonExistentId = randomBytes(16).toString('hex');
                    logger.info(`Generated non-existent workload ID: ${nonExistentId}`);

                    try {
                        await workloadsManagerService.nimServicesApi.updateNimServiceSpec(nonExistentId, {
                            spec: { replicas: 3 }
                        });
                        throw new Error('Expected 404 Not Found, but update succeeded');
                    } catch (error) {
                        verifyErrorResponse(
                            error,
                            StatusCodes.NOT_FOUND,
                            ['requested resource does not exist', 'does not exist', 'not found'],
                            'workload not found'
                        );

                        const errorBody = error.response?.body || error.body || {};
                        expect(errorBody).to.not.be.empty;
                        logger.info('Verified error response structure includes error details');
                    }
                });
                logger.info('Non-existent workload test completed successfully!');
            });

            test('Update without authentication', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                await test.step('Create NIM service with authenticated user and wait for Running', async () => {
                    workloadId = await createAndWaitForNimService(workloadName, project, nimServiceSpec);
                    logger.info(`Created NIM service with ID: ${workloadId}`);
                });

                await test.step('Attempt UPDATE without authentication', async () => {
                    try {
                        await executeWithTokenSwitch(workloadsManagerService.nimServicesApi, '', async () => {
                            await workloadsManagerService.nimServicesApi.updateNimServiceSpec(workloadId, {
                                spec: { replicas: 3 }
                            });
                            throw new Error('Expected 401 Unauthorized');
                        });
                    } catch (error) {
                        const statusCode = error.response?.statusCode || error.statusCode;
                        expect(statusCode).to.equal(StatusCodes.UNAUTHORIZED);
                        logger.info(`Verified HTTP status code: ${statusCode}`);

                        const errorBody = error.response?.body || error.body || {};
                        expect(
                            errorBody.message,
                            `Unexpected 401 response body: ${JSON.stringify(errorBody)}`
                        ).to.equal('authorization header is missing or invalid');

                        logger.info(`401 error response body: ${JSON.stringify(errorBody)}`);
                    }
                });

                await test.step('Verify workload unchanged', async () => {
                    workloadsManagerUtils.accessToken = flowRootState.users.system_administrator.accessToken;
                    await verifyNimServiceFieldUnchanged(
                        workloadId,
                        'spec.replicas',
                        nimServiceSpec.replicas
                    );
                });
                logger.info('Unauthorized update test completed successfully!');
            });

            test('Update without sufficient permissions', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                await test.step('Create NIM service with admin user and wait for Running', async () => {
                    workloadId = await createAndWaitForNimService(workloadName, project, nimServiceSpec);
                    logger.info(`Created NIM service with ID: ${workloadId}`);
                });

                await test.step('Authenticate as non-admin user and attempt UPDATE', async () => {
                    const researcherToken = flowRootState.users.l1_researcher.accessToken;

                    try {
                        await executeWithTokenSwitch(
                            workloadsManagerService.nimServicesApi,
                            researcherToken,
                            async () => {
                                await workloadsManagerService.nimServicesApi.updateNimServiceSpec(
                                    workloadId,
                                    {
                                        spec: { replicas: 3 }
                                    }
                                );
                                throw new Error('Expected 403 Forbidden');
                            }
                        );
                    } catch (error) {
                        verifyErrorResponse(
                            error,
                            StatusCodes.FORBIDDEN,
                            ['forbidden', 'permission', 'not allowed', 'access denied'],
                            'insufficient permissions'
                        );
                    }
                });

                await test.step('Verify workload unchanged', async () => {
                    workloadsManagerUtils.accessToken = flowRootState.users.system_administrator.accessToken;
                    await verifyNimServiceFieldUnchanged(
                        workloadId,
                        'spec.replicas',
                        nimServiceSpec.replicas
                    );
                });
                logger.info('Insufficient permissions test completed successfully!');
            });

            test('Update after workload failure', async function () {
                test.setTimeout(NIM_UPDATE_TEST_TIMEOUT_MS);
                const invalidImage = 'invalid-image-that-does-not-exist:' + randomBytes(2).toString('hex');
                const validImage = config.quickStartImage.nim.includes(':')
                    ? config.quickStartImage.nim
                    : `${config.quickStartImage.nim}:latest`;

                await test.step('Create NIM service with invalid image', async () => {
                    nimServiceSpec.image = invalidImage;

                    const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                    workloadId = createResponse.metadata.id;
                    expect(createResponse.spec.image).to.equal(invalidImage);
                    logger.info(`Created NIM service with ID: ${workloadId}, invalid image: ${invalidImage}`);
                });

                await test.step('Wait for workload to fail due to invalid image', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Creating, 60);
                    await waitUtils.waitForImagePullFailure(project.status.namespace, workloadName, 60, 5);
                    logger.info('Workload failed due to invalid image, proceeding with recovery update');
                });

                await test.step('Update with correct configuration', async () => {
                    const updateResponse = await updateNimServiceSpec(workloadId, { image: validImage }, [
                        StatusCodes.OK,
                        StatusCodes.ACCEPTED
                    ]);
                    expect(updateResponse.body.spec.image).to.equal(validImage);
                    logger.info(`Successfully updated NIM service with valid image: ${validImage}`);
                });

                await test.step('Verify workload recovers and reaches Running state', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                    logger.info(`Workload ${workloadId} recovered and reached Running state`);

                    const workloadResponse = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workloadResponse.body.images[0]).to.equal(validImage);
                    logger.info(`Verified workload is running with corrected image: ${validImage}`);
                });
                logger.info('Workload recovery after failure test completed successfully!');
            });

            test('Update priority of NIM service', async function () {
                test.setTimeout(NIM_UPDATE_SYNC_TEST_TIMEOUT_MS);
                const initialPriority = 'medium';
                const updatedPriority = 'high';

                await test.step('Create NIM service with initial priority and wait for Running', async () => {
                    workloadId = await createAndWaitForNimService(workloadName, project, {
                        ...nimServiceSpec,
                        priorityClass: initialPriority
                    });
                    logger.info(`Created NIM service with ID: ${workloadId}, priority: ${initialPriority}`);
                });

                await test.step('Wait for workload to reach Running phase', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                    logger.info(`Workload ${workloadId} reached Running phase`);
                });

                await test.step('Update priority to high', async () => {
                    await updateNimServiceSpec(workloadId, { priorityClass: updatedPriority }, [
                        StatusCodes.OK,
                        StatusCodes.ACCEPTED
                    ]);
                    logger.info('Successfully called PATCH API to update priority');
                });

                await test.step('Wait for workload to restart with new priority', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                    logger.info(`Workload ${workloadId} restarted and is Running`);
                });

                await test.step('Verify priority updated via GET NIM service API', async () => {
                    const getResponse = await getNimServiceById(workloadId);
                    expect(getResponse.metadata.priority).to.equal(updatedPriority);
                    logger.info(`Verified priority field is "${updatedPriority}" in NIM service API`);
                });

                await test.step('Verify priority reflected in cluster via GET workload API', async () => {
                    await waitUtils.waitForCondition(
                        async () => {
                            const workloadResponse =
                                await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                            const priorityClassName = workloadResponse.body.priorityClassName || '';
                            return priorityClassName.includes(updatedPriority);
                        },
                        `workload ${workloadId} priorityClassName to be updated to include '${updatedPriority}'`,
                        true,
                        120,
                        5
                    );

                    const workloadResponse = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workloadResponse.body.priorityClassName).to.contain(updatedPriority);
                    logger.info(
                        `Verified workload priority reflected in cluster: ${workloadResponse.body.priorityClassName}`
                    );
                });
                logger.info('Priority update test completed successfully!');
            });

            test('Update category of NIM service', async ({}, testInfo: TestInfo) => {
                test.setTimeout(NIM_UPDATE_SYNC_TEST_TIMEOUT_MS);
                // skipPlaywrightTest(
                //     test,
                //     testInfo,
                //     'Skip this test case since there is a bug - category update not reflected in Workload API',

                // );

                const initialCategory = 'Deploy';
                const updatedCategory = 'Train';

                await test.step('Create NIM service with initial category and wait for Running', async () => {
                    workloadId = await createAndWaitForNimService(workloadName, project, {
                        ...nimServiceSpec,
                        category: initialCategory
                    });
                    logger.info(`Created NIM service with ID: ${workloadId}, category: ${initialCategory}`);
                });

                await test.step('Update category to Train', async () => {
                    await updateNimServiceSpec(workloadId, { category: updatedCategory }, [
                        StatusCodes.OK,
                        StatusCodes.ACCEPTED
                    ]);
                    logger.info('Successfully called PATCH API to update category');
                });

                await test.step('Wait for workload to apply new category', async () => {
                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);
                    logger.info(`Workload ${workloadId} is Running with new category`);
                });

                await test.step('Verify category updated via GET NIM service API', async () => {
                    const getResponse = await getNimServiceById(workloadId);
                    expect(getResponse.metadata.category).to.equal(updatedCategory);
                    logger.info(`Verified category field is "${updatedCategory}" in NIM service API`);
                });

                await test.step('Verify category reflected in cluster via GET workload API', async () => {
                    await waitUtils.waitForCondition(
                        async () => {
                            const workloadResponse =
                                await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                            const category = workloadResponse.body.category || '';
                            return category === updatedCategory;
                        },
                        `workload ${workloadId} category to be updated to '${updatedCategory}'`,
                        true,
                        120,
                        5
                    );

                    const workloadResponse = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workloadResponse.body.category).to.equal(updatedCategory);
                    logger.info(
                        `Verified workload category reflected in cluster metadata: ${workloadResponse.body.category}`
                    );
                });
                logger.info('Category update test completed successfully!');
            });

            test('Attempt to update v1 inference workload via NIM service API (negative test)', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                let inferenceWorkloadId: string;

                await test.step('Create a v1 inference workload', async () => {
                    const workloadSpec = {
                        image: config.quickStartImage.inferenceHelloWorld,
                        servingPort: {
                            container: 8080,
                            protocol: 'http'
                        }
                    };

                    const [workloadIdFromCreate] = await inferenceUtils.createInference(
                        workloadName,
                        workloadSpec,
                        String(project.id)
                    );

                    inferenceWorkloadId = workloadIdFromCreate;
                    workloadId = inferenceWorkloadId;

                    logger.info(`Created v1 inference workload with ID: ${inferenceWorkloadId}`);
                });

                await test.step('Attempt to update inference workload via NIM Service UPDATE API', async () => {
                    try {
                        await workloadsManagerService.nimServicesApi.updateNimServiceSpec(
                            inferenceWorkloadId,
                            {
                                spec: { replicas: 2 }
                            }
                        );
                        throw new Error('Expected error (400/404/422), but update succeeded');
                    } catch (error) {
                        verifyErrorResponse(
                            error,
                            StatusCodes.BAD_REQUEST,
                            ['not created via nim-services route'],
                            'v1 inference workload was not created via the NIM Service API'
                        );
                    }
                });

                await test.step('Verify the original inference workload remains unchanged', async () => {
                    const workloadResponse =
                        await workloadsUtilsV2.getWorkloadByWorkloadId(inferenceWorkloadId);
                    expect(workloadResponse.body.type).to.equal('Inference');
                    expect(workloadResponse.body.type).to.not.equal('NIMService');
                    logger.info(`Verified inference workload unchanged, type: ${workloadResponse.body.type}`);
                });

                logger.info('V1 inference workload negative test completed successfully!');
            });

            test('Attempt to update v2 Deployment workload via NIM service API (negative test)', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                let deploymentWorkloadId: string;

                await test.step('Create a Workloads v2 Deployment', async () => {
                    deploymentWorkloadId = await workloadsManagerUtils.createDeploymentWorkloadV2({
                        name: workloadName,
                        projectId: String(project.id),
                        replicas: 2,
                        useStandardManifest: true
                    });
                    workloadId = deploymentWorkloadId;
                    logger.info(`Created v2 Deployment workload with ID: ${deploymentWorkloadId}`);
                });

                await test.step('Attempt to update Deployment workload via NIM Service UPDATE API', async () => {
                    try {
                        await workloadsManagerService.nimServicesApi.updateNimServiceSpec(
                            deploymentWorkloadId,
                            {
                                spec: { replicas: 3 }
                            }
                        );
                        throw new Error('Expected error (400/404/422), but update succeeded');
                    } catch (error) {
                        verifyErrorResponse(
                            error,
                            [
                                StatusCodes.BAD_REQUEST,
                                StatusCodes.NOT_FOUND,
                                StatusCodes.UNPROCESSABLE_ENTITY
                            ],
                            [
                                'not created via nim-service',
                                'not a nim service',
                                'not nimservice',
                                'deployment',
                                'workload type',
                                'cannot update'
                            ],
                            'workload is not a NIM service'
                        );
                    }
                });

                await test.step('Verify the original Deployment workload remains unchanged', async () => {
                    const workloadResponse =
                        await workloadsUtilsV2.getWorkloadByWorkloadId(deploymentWorkloadId);
                    expect(workloadResponse.body.type).to.not.equal('NIMService');
                    expect(workloadResponse.body.type).to.equal('Deployment');
                    logger.info(
                        `Verified Deployment workload unchanged, type: ${workloadResponse.body.type}`
                    );
                });
                logger.info('V2 Deployment workload negative test completed successfully!');
            });
        });
    }
);
