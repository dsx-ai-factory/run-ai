/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
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
    createNimService,
    createNgcAuthSecret,
    getNimServiceById,
    createDefaultNimServiceSpec,
    NIM_TEST_TIMEOUT_MS
} from '../../../../../../modules/helpers/utils/nimServiceUtils';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { isNimOperatorInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';

import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { NimServiceSpec } from '../../../../../../modules/api/apis/workloadsManagerApi/generated/model/nimServiceSpec';
import { ServingPortExposedProtocol } from '../../../../../../modules/api/apis/workloadsManagerApi/generated/model/servingPortExposedProtocol';
import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, TestInfo } from '@playwright/test';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';
import { WorkloadDetailed } from '../../../../../../modules/api/apis/workloadsApi/generated/model/workloadDetailed';
import { randomBytes } from 'crypto';

let project: Project;
let workloadName: string;
let nimServiceSpec: NimServiceSpec;

function validateSpecInNimServiceResponse(spec: NimServiceSpec, expectedSpec: NimServiceSpec): void {
    // expectedSpec may be partial but spec (the returned item) has all the fields
    expect(JSON.parse(JSON.stringify(spec))).to.deep.equal(expectedSpec);
}

function validateNimServiceUrl(workloadSpec: NimServiceSpec, urls: string[]): void {
    expect(urls).to.not.be.undefined;
    expect(urls).to.have.lengthOf(1);
    const url = urls[0];

    // either external or internal url, depends on the provided spec
    if (workloadSpec.servingPort.exposeExternally !== false) {
        if (workloadSpec.servingPort.exposedUrl !== undefined) {
            expect(url).to.equal(workloadSpec.servingPort.exposedUrl);
        } else {
            expect(url).to.contain('https://');
        }
    } else {
        expect(url).to.contain('http://');
        expect(url).to.contain('cluster.local');
    }
}

async function validateGetWorkloadResponseForNimService(
    workloadId: string,
    workloadSpec: NimServiceSpec
): Promise<void> {
    const workloadResponse = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
    const workload: WorkloadDetailed = workloadResponse.body;

    expect(workload.type).to.equal('NIMService');
    expect(workload.category).to.equal('Deploy');
    expect(workload.priorityClassName).to.equal('very-high');
    expect(workload.requestedPods.number).to.equal(workloadSpec.replicas);
    expect(workload.runningPods).to.equal(workloadSpec.replicas);
    expect(workload.conditions).to.not.be.empty;

    const expectedCpuRequest = workloadSpec.compute.cpuCoreRequest * workloadSpec.replicas;
    expect(workload.workloadRequestedResources.cpu.request).to.be.closeTo(expectedCpuRequest, 0.0001);
    expect(workload.workloadRequestedResources.cpuMemory.request).to.not.be.empty;
    expect(workload.podsRequestedResources.cpu.request).to.be.closeTo(expectedCpuRequest, 0.0001);
    expect(workload.podsRequestedResources.cpuMemory.request).to.not.be.empty;
    expect(workload.allocatedResources.cpu).to.be.closeTo(expectedCpuRequest, 0.0001);
    expect(workload.allocatedResources.cpuMemory).to.not.be.empty;

    // Validate GPU resources if specified
    if (
        workloadSpec.compute.gpuPortionRequest !== undefined ||
        workloadSpec.compute.gpuDevicesRequest !== undefined
    ) {
        const gpuPerPod =
            workloadSpec.compute.gpuPortionRequest ?? workloadSpec.compute.gpuDevicesRequest ?? 0;
        const expectedGpuRequest = gpuPerPod * workloadSpec.replicas;

        // Validate allocation equals request
        const allocatedGpu = workload.allocatedResources?.gpu ?? 0;
        expect(allocatedGpu).to.be.closeTo(
            expectedGpuRequest,
            0.001,
            `allocatedResources.gpu should equal requested ${expectedGpuRequest}, but got ${allocatedGpu}`
        );

        logger.info(
            `GPU allocation validated for NIM service: gpuPerPod=${gpuPerPod}, replicas=${workloadSpec.replicas}, total=${expectedGpuRequest}, allocated=${allocatedGpu}`
        );
    }

    let expectedImage = workloadSpec.image;
    if (!expectedImage.includes(':')) {
        expectedImage += ':latest';
    }
    expect(workload.images).to.have.lengthOf(1);
    expect(workload.images[0]).to.equal(expectedImage);

    validateNimServiceUrl(workloadSpec, workload.urls);
}

/**
 * Tests gRPC external access with URL validation.
 * gRPC external access only works with host-based routing (e.g., workload.domain.com).
 * Path-based routing (e.g., domain.com/project/workload) is not supported by gRPC protocol.
 * If the URL uses path-based routing, the test will be skipped.
 */
async function testGrpcExternalAccess(
    externalUrl: string,
    workloadName: string,
    namespace: string
): Promise<void> {
    // Check if URL is host-based (no path) or path-based
    const urlWithoutProtocol = externalUrl.replace(/^https:\/\//, '');
    const hasPath = urlWithoutProtocol.includes('/');

    if (hasPath) {
        logger.warn(
            `Skipping gRPC external access test: URL "${externalUrl}" uses path-based routing. ` +
                'gRPC protocol only supports host-based routing (e.g., workload.domain.com).'
        );
        return;
    } else {
        logger.info(`Testing gRPC external access with host-based URL: ${externalUrl}`);
        await inferenceUtils.sendGrpcRequestExternalUrl(externalUrl, workloadName, namespace);
    }
}

test.describe(
    `nimservice api tests, tags: ${tags.resource.mimService} ${tags.teams.viper}`,
    { tag: [tags.resource.mimService] },
    function () {
        supportedFrom(SupportedVersions.V2_24)(function () {
            test.beforeAll(async function () {
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
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and project ${project.name} namespace ${project.status.namespace}; error ${error}`
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
                    logger.error(`Failed to delete workloads of project ${project.name}; error: ${error}`);
                }
            });

            test('creates a nimservice with external url', async function ({}, testInfo: TestInfo) {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                // create the nimservice with external access (HTTP only, no gRPC)
                delete nimServiceSpec.servingPort.grpcPort;
                nimServiceSpec.servingPort.exposeExternally = true;

                const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                const workloadId = createResponse.metadata.id;
                validateSpecInNimServiceResponse(createResponse.spec, nimServiceSpec);

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);

                const getResponse = await getNimServiceById(workloadId);
                validateSpecInNimServiceResponse(getResponse.spec, nimServiceSpec);
                expect(getResponse.spec.servingPort.grpcPort).to.be.undefined;

                await validateGetWorkloadResponseForNimService(workloadId, nimServiceSpec);

                const externalUrl = await inferenceUtils.getServingUrl(workloadId);
                expect(externalUrl).to.contain('https://');
                expect(externalUrl).to.not.contain('cluster.local');

                await inferenceUtils.sendInferenceRequestExternalUrl(
                    externalUrl,
                    '',
                    true,
                    false,
                    workloadName,
                    project.status.namespace
                );
            });

            shouldRunNativeWorkloadCrdTests(() =>
            test('creates a nimservice with internal url', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                nimServiceSpec.servingPort.exposeExternally = false;
                const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                const workloadId = createResponse.metadata.id;
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);

                await validateGetWorkloadResponseForNimService(workloadId, nimServiceSpec);

                const internalUrl = await inferenceUtils.getServingUrl(workloadId);
                expect(internalUrl).to.contain('cluster.local');

                await inferenceUtils.sendInferenceRequest(project.status.namespace, internalUrl);
                })
            );

            test('creates a nimservice with custom external url', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                const customUrl =
                    process.env.CONTROL_PLANE_URL + '/test-nim-' + randomBytes(2).toString('hex');
                nimServiceSpec.servingPort.port = 8000;
                nimServiceSpec.servingPort.exposeExternally = true;
                nimServiceSpec.servingPort.exposedUrl = customUrl;
                delete nimServiceSpec.servingPort.grpcPort;

                const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                const workloadId = createResponse.metadata.id;
                validateSpecInNimServiceResponse(createResponse.spec, nimServiceSpec);

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);

                const getResponse = await getNimServiceById(workloadId);
                validateSpecInNimServiceResponse(getResponse.spec, nimServiceSpec);

                const serviceUrl = await inferenceUtils.getServingUrl(workloadId);
                expect(serviceUrl).to.equal(customUrl);

                await inferenceUtils.sendInferenceRequestExternalUrl(
                    customUrl,
                    '',
                    true,
                    false,
                    workloadName,
                    project.status.namespace
                );
            });

            test('creates a nimservice with gRPC port external access', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                nimServiceSpec.servingPort.port = 8000;
                nimServiceSpec.servingPort.grpcPort = 8001;
                nimServiceSpec.servingPort.exposeExternally = true;
                nimServiceSpec.servingPort.exposedProtocol = ServingPortExposedProtocol.Grpc;

                const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                const workloadId = createResponse.metadata.id;
                validateSpecInNimServiceResponse(createResponse.spec, nimServiceSpec);

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);

                const getResponse = await getNimServiceById(workloadId);
                expect(getResponse.spec.servingPort.grpcPort).to.equal(8001);
                expect(getResponse.spec.servingPort.exposedProtocol).to.equal(
                    ServingPortExposedProtocol.Grpc
                );

                const externalUrl = await inferenceUtils.getServingUrl(workloadId);
                expect(externalUrl).to.contain('https://');
                expect(externalUrl).to.not.contain('cluster.local');

                await testGrpcExternalAccess(externalUrl, workloadName, project.status.namespace);
            });

            shouldRunNativeWorkloadCrdTests(() =>
            test('creates a nimservice with both HTTP and gRPC ports internal access', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                nimServiceSpec.servingPort.port = 8000;
                nimServiceSpec.servingPort.grpcPort = 8001;
                nimServiceSpec.servingPort.exposeExternally = false;

                const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                const workloadId = createResponse.metadata.id;
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);

                await validateGetWorkloadResponseForNimService(workloadId, nimServiceSpec);

                const getResponse = await getNimServiceById(workloadId);
                expect(getResponse.spec.servingPort.port).to.equal(8000);
                expect(getResponse.spec.servingPort.grpcPort).to.equal(8001);

                    const httpUrl = inferenceUtils.buildServiceUrl(
                        workloadName,
                        project.status.namespace,
                        8000
                    );
                await inferenceUtils.sendInferenceRequest(project.status.namespace, `http://${httpUrl}`);

                    const grpcUrl = inferenceUtils.buildServiceUrl(
                        workloadName,
                        project.status.namespace,
                        8001
                    );
                // Wait for gRPC service to be fully ready
                await waitUtils.sleep(2000);
                await inferenceUtils.sendGrpcRequest(project.status.namespace, grpcUrl);
                })
            );

            test('creates a nimservice with exposedProtocol set to HTTP', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                nimServiceSpec.servingPort.port = 8000;
                nimServiceSpec.servingPort.grpcPort = 8001;
                nimServiceSpec.servingPort.exposeExternally = true;
                nimServiceSpec.servingPort.exposedProtocol = ServingPortExposedProtocol.Http;

                const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                const workloadId = createResponse.metadata.id;
                validateSpecInNimServiceResponse(createResponse.spec, nimServiceSpec);

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);

                const getResponse = await getNimServiceById(workloadId);
                expect(getResponse.spec.servingPort.exposedProtocol).to.equal(
                    ServingPortExposedProtocol.Http
                );
                expect(getResponse.spec.servingPort.grpcPort).to.equal(8001);

                const externalUrl = await inferenceUtils.getServingUrl(workloadId);
                expect(externalUrl).to.contain('https://');
                expect(externalUrl).to.not.contain('cluster.local');

                await inferenceUtils.sendInferenceRequestExternalUrl(
                    externalUrl,
                    '',
                    true,
                    false,
                    workloadName,
                    project.status.namespace
                );
            });

            test('creates a nimservice with GPU fraction and validates allocation', async function () {
                test.setTimeout(NIM_TEST_TIMEOUT_MS);
                // Configure NIM service with GPU fraction
                nimServiceSpec.compute = {
                    cpuCoreRequest: 0.1,
                    cpuMemoryRequest: '10M',
                    gpuPortionRequest: 0.5 // GPU fraction
                };
                nimServiceSpec.replicas = 2;
                nimServiceSpec.servingPort.exposeExternally = false;
                delete nimServiceSpec.servingPort.grpcPort;

                const createResponse = await createNimService(workloadName, project, nimServiceSpec);
                const workloadId = createResponse.metadata.id;
                validateSpecInNimServiceResponse(createResponse.spec, nimServiceSpec);

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 240);

                // Validate workload response including GPU allocation
                await validateGetWorkloadResponseForNimService(workloadId, nimServiceSpec);

                const getResponse = await getNimServiceById(workloadId);
                validateSpecInNimServiceResponse(getResponse.spec, nimServiceSpec);

                // Additional explicit check that allocation equals request
                const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                const expectedGpu = nimServiceSpec.compute.gpuPortionRequest * nimServiceSpec.replicas;
                expect(workload.body.allocatedResources.gpu).to.be.closeTo(expectedGpu, 0.0001);

                logger.info(
                    `NIM service GPU fraction test passed: ${expectedGpu} GPU allocated across ${nimServiceSpec.replicas} replicas`
                );
            });
        });
    }
);
