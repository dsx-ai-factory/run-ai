/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsManagerUtils
} from '../../../../../../modules/helpers';
import { expect } from 'chai';
import { test } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';
import { invalidUuid } from '../../../../../../modules/helpers/config/const';
import { Workload } from '../../../../../../modules/api/apis/workloadsApi/generated/model/workload';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadV2 } from '../../../../../../modules/api/apis/workloadsManagerApi/generated/model/workloadV2';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

supportedFrom(SupportedVersions.V2_23)(function () {
    test.describe(
        `Workloads V2 API tests: ${tags.resource.workloadsV2} ${tags.tool.api} ${tags.teams.viper}`,
        { tag: [tags.resource.workloadsV2] },
        function () {
            let project: Project;
            let workloadName: string;
            let createdWorkloadIds: string[] = [];
            let defaultNodePoolId: string;

            // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
            function createStandardManifest(name: string) {
                return {
                    apiVersion: 'apps/v1',
                    kind: 'Deployment',
                    metadata: {
                        name: name
                    },
                    spec: {
                        replicas: 1,
                        selector: {
                            matchLabels: {
                                app: name
                            }
                        },
                        template: {
                            metadata: {
                                labels: {
                                    app: name
                                }
                            },
                            spec: {
                                containers: [
                                    {
                                        name: 'app',
                                        image: 'busybox',
                                        command: ['sh', '-c', 'while true; do echo hello; sleep 5; done']
                                    }
                                ]
                            }
                        }
                    }
                };
            }

            test.beforeAll(async function () {
                defaultNodePoolId = String((await nodePoolUtils.getDefaultNodePoolFromBackend()).id);

                project = await orgUnitUtils.createProjectAndValidate({
                    resources: [
                        {
                            nodePool: { id: defaultNodePoolId, name: 'default' },
                            gpu: { deserved: 1 }
                        }
                    ],
                    name: objectGenerator.generateNameByPrefix('workloads-v2'),
                    clusterId: flowRootState.cluster.uuid
                });
                logger.info(`Created test project: ${project.name} with ID: ${project.id}`);
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateNameByPrefix('w-v2');
            });

            test.afterEach(async function () {
                for (const workloadId of createdWorkloadIds) {
                    try {
                        await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                        logger.debug(`Cleaned up workload: ${workloadId}`);
                    } catch (error) {
                        logger.warn(`Failed to clean up workload ${workloadId}: ${error.message}`);
                    }
                }
                createdWorkloadIds = [];
            });

            test.afterAll(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await orgUnitUtils.deleteProject(project.id);
                } catch (error) {
                    logger.error(`Failed to delete workloads and project ${project.name}:`, error);
                }
            });

            test.describe('POST /api/v2/workloads - Create workload', () => {
                test('should create a simple workload with basic manifest', async function () {
                    const manifest = createStandardManifest(workloadName);

                    const result = await workloadsManagerUtils.createWorkloadV2(
                        workloadName,
                        String(project.id),
                        manifest,
                        { priority: 'medium', category: 'Train' }
                    );

                    expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                    expect(result.body).to.not.be.undefined;
                    expect(result.body.metadata).to.not.be.undefined;
                    expect(result.body.metadata.name).to.equal(workloadName);
                    expect(result.body.metadata.projectId).to.equal(String(project.id));
                    expect(result.body.metadata.priority).to.equal('medium');
                    expect(result.body.metadata.category).to.equal('Train');
                    expect(result.body.metadata.id).to.not.be.undefined;
                    expect(result.body.desiredPhase).to.equal('Running');
                    expect(result.body.rawManifest).to.deep.equal(manifest);
                    expect(result.body.workloadManifest).to.not.be.undefined;

                    createdWorkloadIds.push(result.body.metadata.id);

                    await waitUtils.waitForBackendWorkloadByIdAndPhase(
                        result.body.metadata.id,
                        Phase.Running,
                        120
                    );

                    const workloadId = result.body.metadata.id;
                    const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(getResult.response.statusCode).to.equal(200);
                    expect(getResult.body.workload.metadata.name).to.equal(result.body.metadata.name);
                });

                test('should create a workload with base64 encoded manifest', async function () {
                    const yamlManifest = JSON.stringify(createStandardManifest(workloadName), null, 2);
                    const manifestBase64Encoded = Buffer.from(yamlManifest).toString('base64');

                    const result = await workloadsManagerUtils.createWorkloadV2(
                        workloadName,
                        String(project.id),
                        undefined,
                        undefined,
                        manifestBase64Encoded
                    );

                    expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                    expect(result.body).to.not.be.undefined;
                    expect(result.body.metadata.name).to.equal(workloadName);
                    expect(result.body.metadata.projectId).to.equal(String(project.id));
                    expect(result.body.metadata.id).to.not.be.undefined;

                    createdWorkloadIds.push(result.body.metadata.id);

                    await waitUtils.waitForBackendWorkloadByIdAndPhase(
                        result.body.metadata.id,
                        Phase.Running,
                        120
                    );

                    const workloadId = result.body.metadata.id;
                    const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(getResult.response.statusCode).to.equal(200);
                    expect(getResult.body.workload.metadata.name).to.equal(result.body.metadata.name);
                });

                test('should fail to create workload with invalid project ID', async function () {
                    const manifest = createStandardManifest(workloadName);

                    try {
                        await workloadsManagerUtils.createWorkloadV2(workloadName, invalidUuid, manifest);
                        expect.fail('Expected request to fail with invalid project ID');
                    } catch (error) {
                        expect(error.statusCode).to.be.oneOf([StatusCodes.FORBIDDEN, StatusCodes.NOT_FOUND]);
                    }
                });

                test('should fail to create workload with missing required metadata', async function () {
                    try {
                        await workloadsManagerUtils.createWorkloadV2(
                            '',
                            '',
                            createStandardManifest(workloadName)
                        );
                        expect.fail('Expected request to fail with missing metadata');
                    } catch (error) {
                        expect(error.statusCode).to.equal(StatusCodes.BAD_REQUEST);
                    }
                });

                test('should fail to create workload with both manifest and manifestBase64Encoded', async function () {
                    try {
                        const manifestBase64Encoded = 'dGVzdA==';
                        await workloadsManagerUtils.createWorkloadV2(
                            workloadName,
                            String(project.id),
                            createStandardManifest(workloadName),
                            undefined,
                            manifestBase64Encoded
                        );
                        expect.fail('Expected request to fail with both manifest formats');
                    } catch (error) {
                        expect(error.statusCode).to.equal(StatusCodes.BAD_REQUEST);
                    }
                });
            });

            test.describe('GET /api/v2/workloads/{id} - Get workload by ID', () => {
                let workloadId: string;
                let created: WorkloadV2;

                test.beforeEach(async function () {
                    if (!created) {
                        const manifest = createStandardManifest(workloadName);

                        const createResult = await workloadsManagerUtils.createWorkloadV2(
                            workloadName,
                            String(project.id),
                            manifest,
                            { priority: 'high' }
                        );
                        expect(createResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                        workloadId = createResult.body.metadata.id;
                        createdWorkloadIds.push(workloadId);
                        created = createResult.body;

                        await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                        const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                        expect(getResult.response.statusCode).to.equal(200);
                        expect(getResult.body.workload.metadata.name).to.equal(workloadName);
                    }
                });

                test.afterAll(async function () {
                    if (created) {
                        const result = await workloadsManagerUtils.deleteWorkloadV2(created.metadata.id);
                        expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                    }
                });

                test('should retrieve and validate workload by valid ID', async function () {
                    const result = await workloadsManagerUtils.getWorkloadV2ById(workloadId);

                    expect(result.response.statusCode).to.equal(StatusCodes.OK);
                    expect(result.body).to.not.be.undefined;
                    expect(result.body.workload).to.not.be.undefined;

                    expect(result.body.workload.metadata.id).to.equal(workloadId);
                    expect(result.body.workload.metadata.name).to.equal(workloadName);
                    expect(result.body.workload.metadata.projectId).to.equal(String(project.id));
                    expect(result.body.workload.metadata.priority).to.equal('high');
                    expect(result.body.workload.metadata.id).to.not.be.undefined;
                    expect(result.body.workload.metadata.createdAt).to.not.be.undefined;

                    expect(result.body.workload.desiredPhase).to.equal('Running');
                    expect(result.body.workload.workloadManifest).to.not.be.undefined;
                    expect(result.body.workload.rawManifest).to.not.be.undefined;
                });

                test('should fail to retrieve workload with invalid ID', async function () {
                    try {
                        await workloadsManagerUtils.getWorkloadV2ById(invalidUuid);
                        expect.fail('Expected request to fail with invalid ID');
                    } catch (error) {
                        expect(error.statusCode).to.equal(StatusCodes.NOT_FOUND);
                    }
                });

                test('should fail to retrieve workload with malformed ID', async function () {
                    const malformedId = 'not-a-uuid';

                    try {
                        await workloadsManagerUtils.getWorkloadV2ById(malformedId);
                        expect.fail('Expected request to fail with malformed ID');
                    } catch (error) {
                        expect(error.statusCode).to.equal(StatusCodes.BAD_REQUEST);
                    }
                });
            });

            test.describe('DELETE /api/v2/workloads/{id} - Delete workload', () => {
                let workloadId: string;

                test.beforeEach(async function () {
                    workloadName = objectGenerator.generateNameByPrefix('w-v2');
                    const manifest = createStandardManifest(workloadName);

                    const createResult = await workloadsManagerUtils.createWorkloadV2(
                        workloadName,
                        String(project.id),
                        manifest
                    );
                    workloadId = createResult.body.metadata.id;

                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                    const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(getResult.response.statusCode).to.equal(200);
                    expect(getResult.body.workload.metadata.name).to.equal(createResult.body.metadata.name);
                });

                test('should delete workload by valid ID', async function () {
                    const result = await workloadsManagerUtils.deleteWorkloadV2(workloadId);

                    expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                    expect(result.body).to.not.be.undefined;
                    expect(result.body.metadata.id).to.equal(workloadId);
                    expect(result.body.desiredPhase).to.equal('Deleted');

                    const getDeletedResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(getDeletedResult.response.statusCode).to.equal(200);
                    expect(getDeletedResult.body.workload.desiredPhase).to.equal('Deleted');
                    expect(getDeletedResult.body.workload.metadata.deletedAt).to.not.be.null;

                    await waitUtils.waitForWorkloadToBeDeleted(workloadName, 120);
                });

                test('should fail to delete workload with invalid ID', async function () {
                    try {
                        await workloadsManagerUtils.deleteWorkloadV2(invalidUuid);
                        expect.fail('Expected request to fail with invalid ID');
                    } catch (error) {
                        expect(error.statusCode).to.equal(StatusCodes.NOT_FOUND);
                    }
                });

                test('should fail to delete already deleted workload', async function () {
                    const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    expect(deleteResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                    expect(deleteResult.body.desiredPhase).to.equal('Deleted');

                    const getDeletedResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(getDeletedResult.response.statusCode).to.equal(200);
                    expect(getDeletedResult.body.workload.desiredPhase).to.equal('Deleted');

                    await waitUtils.waitForWorkloadToBeDeleted(workloadName, 120);
                });
            });

            test.describe('Workload lifecycle integration tests', () => {
                test('should handle complete workload lifecycle', async function () {
                    const manifest = createStandardManifest(workloadName);

                    const createResult = await workloadsManagerUtils.createWorkloadV2(
                        workloadName,
                        String(project.id),
                        manifest,
                        { priority: 'medium', category: 'Build' }
                    );
                    expect(createResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                    const workloadId = createResult.body.metadata.id;

                    await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                    const verifyResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(verifyResult.response.statusCode).to.equal(200);
                    expect(verifyResult.body.workload.metadata.name).to.equal(
                        createResult.body.metadata.name
                    );

                    const v1Result: Workload = await waitUtils.waitForBackendWorkload(
                        createResult.body.metadata.name,
                        String(project.id),
                        120
                    );
                    expect(v1Result).to.not.be.undefined;
                    expect(v1Result.name).to.equal(workloadName);

                    const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(getResult.response.statusCode).to.equal(StatusCodes.OK);
                    expect(getResult.body.workload.metadata.id).to.equal(workloadId);
                    expect(getResult.body.workload.desiredPhase).to.equal('Running');

                    const deleteResult = await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    expect(deleteResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                    expect(deleteResult.body.desiredPhase).to.equal('Deleted');

                    const getDeletedResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(getDeletedResult.body.workload.desiredPhase).to.equal('Deleted');
                    expect(getDeletedResult.body.workload.metadata.deletedAt).to.not.be.null;

                    await waitUtils.waitForWorkloadToBeDeleted(workloadName);
                });

                test('should handle workload with complex manifest', async function () {
                    const manifest = {
                        apiVersion: 'apps/v1',
                        kind: 'Deployment',
                        metadata: {
                            name: workloadName,
                            labels: {
                                app: workloadName,
                                version: 'v1',
                                environment: 'test'
                            },
                            annotations: {
                                'runai.ai/test': 'true',
                                'runai.ai/category': 'e2e-test'
                            }
                        },
                        spec: {
                            replicas: 2,
                            selector: {
                                matchLabels: {
                                    app: workloadName
                                }
                            },
                            template: {
                                metadata: {
                                    labels: {
                                        app: workloadName,
                                        version: 'v1'
                                    }
                                },
                                spec: {
                                    containers: [
                                        {
                                            name: 'main-container',
                                            image: config.quickStartImage.inference,
                                            ports: [
                                                {
                                                    containerPort: 8080,
                                                    name: 'http'
                                                }
                                            ],
                                            env: [
                                                {
                                                    name: 'ENV_VAR',
                                                    value: 'test-value'
                                                }
                                            ],
                                            resources: {
                                                requests: {
                                                    cpu: '100m',
                                                    memory: '128Mi'
                                                },
                                                limits: {
                                                    cpu: '500m',
                                                    memory: '512Mi'
                                                }
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    };

                    const result = await workloadsManagerUtils.createWorkloadV2(
                        workloadName,
                        String(project.id),
                        manifest,
                        { priority: 'high', category: 'Deploy' }
                    );

                    expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                    expect(result.body.metadata.name).to.equal(workloadName);
                    expect(result.body.metadata.priority).to.equal('high');
                    expect(result.body.metadata.category).to.equal('Deploy');
                    expect(result.body.rawManifest).to.deep.equal(manifest);

                    createdWorkloadIds.push(result.body.metadata.id);

                    await waitUtils.waitForBackendWorkloadByIdAndPhase(
                        result.body.metadata.id,
                        Phase.Running,
                        120
                    );

                    const workloadId = result.body.metadata.id;
                    const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(getResult.response.statusCode).to.equal(200);
                    expect(getResult.body.workload.metadata.name).to.equal(result.body.metadata.name);
                });

                test('should verify workload data consistency in V2 API', async function () {
                    const manifest = createStandardManifest(workloadName);

                    const createResult = await workloadsManagerUtils.createWorkloadV2(
                        workloadName,
                        String(project.id),
                        manifest,
                        { priority: 'low', category: 'Train' }
                    );
                    expect(createResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);

                    createdWorkloadIds.push(createResult.body.metadata.id);

                    await waitUtils.waitForBackendWorkloadByIdAndPhase(
                        createResult.body.metadata.id,
                        Phase.Running,
                        120
                    );

                    const workloadId = createResult.body.metadata.id;
                    const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                    expect(getResult.response.statusCode).to.equal(200);
                    expect(getResult.body.workload.metadata.name).to.equal(workloadName);
                    expect(getResult.body.workload.metadata.projectId).to.equal(String(project.id));
                    expect(getResult.body.workload.metadata.priority).to.equal('low');
                    expect(getResult.body.workload.metadata.category).to.equal('Train');
                });
            });
        }
    );
});
