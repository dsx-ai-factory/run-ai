/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    assetsApiV2Utils,
    computeResourceUtils,
    config,
    clusterUtils,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadAuthUtils,
    workloadsCrudUtils,
    workloadsUtilsV2,
    workloadUtils
} from '../../../../../../modules/helpers';
import {
    InferenceCreationRequestV2,
    ServingPortAccessAuthorizationTypeEnum,
    WorkspaceCreationRequestV2
} from '../../../../../../modules/api/apis/assetsApi';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { expect } from 'chai';
import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { inferenceUtils } from '../../../../../../modules/helpers';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { InferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceSpecSpec';
import { AuthorizationType } from '../../../../../../modules/api/apis/workloadsApi/generated/model/authorizationType';
import { assetsApi } from '../../../../../../modules/api';
import { test } from '@playwright/test';

supportedFrom(SupportedVersions.V2_19)(function () {
    test.describe(
        `inference workload authorization tests: ${tags.runOn.parallelCI} ${tags.resource.inferenceWorkload} ${tags.tool.kubeCli} ${tags.teams.viper}`,
        { tag: [tags.resource.inferenceWorkload] },
        function () {
            test.describe.configure({ retries: 1 });
            let project: Project;
            let workloadName: string;

            test.beforeAll(async function () {
                if (!(await isKnativeInstalled())) {
                    logger.warn('knative is not installed on the cluster, skipping test');
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
            });

            test.afterAll(async function () {
                if (!project) {
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
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(String(project.id));
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and pods of project ${project.name} namespace ${project.status.namespace}`
                    );
                }
            });

            async function createInference(
                authorizationType: ServingPortAccessAuthorizationTypeEnum,
                authorizedUsers?: string[]
            ): Promise<void> {
                const spec: InferenceSpecSpec = {
                    image: config.quickStartImage.inferenceHelloWorld,
                    servingPort: {
                        container: 8080,
                        authorizationType,
                        authorizedUsers
                    }
                };
                await inferenceUtils.createInference(workloadName, spec, String(project.id));
                await workloadUtils.waitForWorkloadRunnerPhase(
                    workloadName,
                    project.status.namespace,
                    Phase.Running,
                    120
                );
            }

            shouldRunNativeWorkloadCrdTests(() =>
            test('creates a workload with authorization type = public', async function () {
                await createInference(ServingPortAccessAuthorizationTypeEnum.Public);
                logger.info('creating a workload with authorization type = public');
                const pods = await waitUtils.waitForPods(
                    workloadName,
                    project.status.namespace,
                    1,
                    60,
                    ['Running'],
                    true
                );
                expect(pods.length).to.equal(1);

                // expected to succeed without a token
                logger.info('sending inference request without a token, expected to succeed');
                await inferenceUtils.sendInferenceRequest(
                    project.status.namespace,
                    `http://${workloadName}.${project.status.namespace}.svc.cluster.local`
                );
                })
            );

            shouldRunNativeWorkloadCrdTests(() =>
            test('creates a workload with authorization type = authenticatedUsers', async function () {
                    await createInference(ServingPortAccessAuthorizationTypeEnum.AuthenticatedUsers);
                logger.info('creating a workload with authorization type = authenticatedUsers');
                const pods = await waitUtils.waitForPods(
                    workloadName,
                    project.status.namespace,
                    1,
                    60,
                    ['Running'],
                    true
                );
                expect(pods.length).to.equal(1);

                const internalUrl = `http://${workloadName}.${project.status.namespace}.svc.cluster.local`;

                // expected to fail without a token and succeed with a token
                logger.info('sending inference request without a token, expected to fail');
                await inferenceUtils.sendInferenceRequest(
                    project.status.namespace,
                    internalUrl,
                    '',
                    false
                );
                logger.info('sending inference request with a token, expected to succeed');
                await inferenceUtils.sendInferenceRequest(
                    project.status.namespace,
                    internalUrl,
                    flowRootState.users.ml_engineer.accessToken
                );
                })
            );

            shouldRunNativeWorkloadCrdTests(() =>
            test('creates a workload with authorization type = authorizedUsers', async function () {
                    await createInference(ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups, [
                        flowRootState.users.ml_engineer.email
                    ]);
                logger.info('creating a workload with authorization type = authorizedUsers');
                const pods = await waitUtils.waitForPods(
                    workloadName,
                    project.status.namespace,
                    1,
                    60,
                    ['Running'],
                    true
                );
                expect(pods.length).to.equal(1);

                const internalUrl = `http://${workloadName}.${project.status.namespace}.svc.cluster.local`;

                // expected to fail without a token and succeed with a token
                logger.info('sending inference request without a token, expected to fail');
                await inferenceUtils.sendInferenceRequest(
                    project.status.namespace,
                    internalUrl,
                    '',
                    false
                );
                logger.info('sending inference request with a token, expected to succeed');
                await inferenceUtils.sendInferenceRequest(
                    project.status.namespace,
                    internalUrl,
                    flowRootState.users.ml_engineer.accessToken
                );
                })
            );
        }
    );
});

supportedFrom(SupportedVersions.V2_20)(function () {
    test.describe(
        `inference workload authorization api tests: ${tags.resource.inferenceWorkload} ${tags.tool.api} ${tags.teams.viper}`,
        { tag: [tags.resource.inferenceWorkload] },
        function () {
            test.describe.configure({ retries: 1 });
            let project: Project;
            let workloadId: string;
            let workloadName: string;
            let workloadSpec: InferenceSpecSpec;

            async function validateConnectionsInGetWorkloadResponse(
                workloadId,
                numConnections,
                authorizationType
            ): Promise<void> {
                const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                const servingPortUrl = inferenceUtils.getKnativeUrlForInferenceWorkload(workload.body);

                expect(workload.body.environments).to.have.lengthOf(1);
                expect(workload.body.environments[0].connections).to.have.lengthOf(numConnections);
                expect(
                    workload.body.environments[0].connections[numConnections - 1].authorizationType
                ).to.equal(authorizationType);
                expect(workload.body.environments[0].connections[numConnections - 1].url).to.not.be.undefined;
                expect(workload.body.environments[0].connections[numConnections - 1].url).to.equal(
                    servingPortUrl
                );
                logger.info('connections in get workload response are as expected');
            }

            test.beforeAll(async function () {
                if (!(await isKnativeInstalled())) {
                    logger.warn('knative is not installed on the cluster, skipping test');
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

                workloadSpec = {
                    image: config.quickStartImage.inferenceHelloWorld,
                    servingPort: {
                        container: 8080
                    }
                };

                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
                assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;
                assetsApi.inferenceApi.accessToken = flowRootState.users.system_administrator.accessToken;
            });

            test.afterAll(async function () {
                if (!project) {
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
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(String(project.id));
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and pods of project ${project.name} namespace ${project.status.namespace}`
                    );
                }
            });

            test('creates a workload using param-based api with authorization type = public', async function () {
                workloadSpec.servingPort.authorizationType = ServingPortAccessAuthorizationTypeEnum.Public;
                [workloadId, workloadName] = await inferenceUtils.createInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                await validateConnectionsInGetWorkloadResponse(
                    workloadId,
                    1,
                    ServingPortAccessAuthorizationTypeEnum.Public
                );
            });

            test('creates a workload using param-based api with authorization type = authorizedUsers', async function () {
                workloadSpec.servingPort.authorizationType =
                    ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups;
                workloadSpec.servingPort.authorizedUsers = ['user1', 'user2'];
                [workloadId, workloadName] = await inferenceUtils.createInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                await validateConnectionsInGetWorkloadResponse(
                    workloadId,
                    1,
                    ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups
                );
            });

            supportedFrom(SupportedVersions.V2_23)(function () {
                shouldRunNativeWorkloadCrdTests(() =>
                test('protects inference serving port and external URL independently', async function () {
                    test.setTimeout(300 * 1000);

                    const servingPortUser = flowRootState.users.system_administrator;
                    const externalUrlUser = flowRootState.users.ml_engineer;
                    workloadSpec.servingPort.authorizationType =
                        ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups;
                    workloadSpec.servingPort.authorizedUsers = [servingPortUser.email];
                    workloadSpec.exposedUrls = [
                        {
                            name: 'authenticated-external-url',
                            container: 8080,
                            authorizationType: AuthorizationType.AuthorizedUsers,
                            authorizedUsers: [externalUrlUser.email]
                        }
                    ];

                    [workloadId, workloadName] = await inferenceUtils.createInference(
                        workloadName,
                        workloadSpec,
                        String(project.id)
                    );
                    await workloadUtils.waitForWorkloadRunnerPhase(
                        workloadName,
                        project.status.namespace,
                        Phase.Running,
                        120
                    );

                    let servingPortUrl: string | undefined;
                    let externalUrl: string | undefined;
                    await waitUtils.waitForCondition(
                        async () => {
                            const workloads = (
                                await workloadsUtilsV2.getWorkloadByName(workloadName, project.id)
                            ).body.workloads;
                            const workload = workloads.find((candidate) => {
                                const connections = candidate.environments?.flatMap(
                                    (environment) => environment.connections || []
                                );
                                return (
                                    connections?.some(
                                        (connection) => connection.connectionType === 'ServingPort'
                                    ) &&
                                    connections?.some(
                                        (connection) => connection.connectionType === 'ExternalUrl'
                                    )
                                );
                            });
                            const connections = workload?.environments?.flatMap(
                                (environment) => environment.connections || []
                            );
                            workloadId = workload?.id;
                            servingPortUrl = connections?.find(
                                (connection) => connection.connectionType === 'ServingPort'
                            )?.url;
                            externalUrl = connections?.find(
                                (connection) => connection.connectionType === 'ExternalUrl'
                            )?.url;
                            return Boolean(servingPortUrl && externalUrl);
                        },
                        'serving port and external URL connections to be reported',
                        true,
                        120,
                        5
                    );

                    expect(servingPortUrl).to.not.equal(externalUrl);

                    const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    const servingPortInternalUrl = inferenceUtils.getInternalUrlForInferenceWorkload(
                        workload.body
                    );

                    await inferenceUtils.sendInferenceRequest(
                        project.status.namespace,
                        servingPortInternalUrl,
                        '',
                        false
                    );
                    await inferenceUtils.sendInferenceRequest(
                        project.status.namespace,
                        servingPortInternalUrl,
                        servingPortUser.accessToken
                    );
                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl!, '', [401, 302]);
                    await workloadAuthUtils.verifyExternalUrlHttpStatus(
                        externalUrl!,
                        servingPortUser.accessToken,
                        403
                    );
                    await workloadAuthUtils.verifyExternalUrlHttpStatus(
                        externalUrl!,
                        externalUrlUser.accessToken,
                        200
                    );
                    })
                );
            });

            test('creates a workspace with huggingface-chat-ui and authorized users on the connection', async function () {
                // Step 1: Create a mock LLM inference workload
                const mockLlmName = objectGenerator.generateNameByPrefix('mock-llm');
                const mockLlmSpec: InferenceSpecSpec = {
                    image: 'zerob13/mock-openai-api:latest',
                    servingPort: {
                        container: 3000
                    }
                };
                const [mockLlmWorkloadId, mockLlmWorkloadName] = await inferenceUtils.createInference(
                    mockLlmName,
                    mockLlmSpec,
                    String(project.id)
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(mockLlmWorkloadId, Phase.Running, 120);

                const mockLlmInternalUrl = `http://${mockLlmWorkloadName}.${project.status.namespace}.svc.cluster.local/v1`;

                // Step 2: Detect cluster routing type and pick the correct chat-ui environment
                const clusterConfig = await clusterUtils.getClusterReportedConfig();
                const isHostBased = clusterConfig?.subdomainEnabled === true;
                const chatUiEnvName = isHostBased ? 'huggingface-chat-ui' : 'huggingface-chat-ui-path';
                logger.info(
                    `Cluster routing: ${isHostBased ? 'host-based' : 'path-based'}, using env: ${chatUiEnvName}`
                );
                const chatUiEnv = await environmentUtils.get(chatUiEnvName);
                const connectionName = chatUiEnv.spec.connections[0]?.name;
                const environmentId = chatUiEnv.meta.id;
                const computeId = await computeResourceUtils.getId('cpu-only');

                // Step 3: Build env vars — host-based needs PUBLIC_ORIGIN
                const environmentVariables = [
                    {
                        name: 'OPENAI_BASE_URL',
                        value: mockLlmInternalUrl
                    }
                ];
                if (isHostBased) {
                    const clusterDomain = flowRootState.cluster.domain.replace(/^https?:\/\//, '');
                    environmentVariables.push({
                        name: 'PUBLIC_ORIGIN',
                        value: `https://${project.name}-${workloadName}.${clusterDomain}`
                    });
                }

                // Step 4: Create a workspace with the chat-ui environment and authorized users on the connection
                const workspacePayload: WorkspaceCreationRequestV2 = {
                    meta: {
                        name: workloadName,
                        useGivenNameAsPrefix: false,
                        projectId: project.id.toString(),
                        clusterId: project.clusterId
                    },
                    assets: {
                        environment: environmentId,
                        compute: computeId
                    },
                    specificRunParams: {
                        connections: [
                            {
                                name: connectionName,
                                authorizedUsers: [config.controlPlaneAdminUsername, 'user1', 'user2']
                            }
                        ],
                        environmentVariables
                    }
                };

                const resultWorkload = await assetsApiV2Utils.createWorkspace(workspacePayload);
                workloadName = resultWorkload.body.meta.name;
                workloadId = resultWorkload.body.meta.workloadId;

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                // Validate the workspace has the connection with correct authorization
                const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                expect(workload.body.environments).to.have.lengthOf(1);
                expect(workload.body.environments[0].connections).to.have.length.greaterThan(0);

                const chatUiConnection = workload.body.environments[0].connections.find((conn) => conn.url);
                expect(chatUiConnection).to.not.be.undefined;
                logger.info(`Chat UI URL: ${chatUiConnection.url}`);

                const authorizedConnection = workload.body.environments[0].connections.find(
                    (conn) => conn.authorizedUsers?.length > 0
                );
                expect(authorizedConnection).to.not.be.undefined;
                expect(authorizedConnection.authorizedUsers).to.include(config.controlPlaneAdminUsername);
                expect(authorizedConnection.authorizedUsers).to.include('user1');
                expect(authorizedConnection.authorizedUsers).to.include('user2');
                logger.info('workspace connections with authorized users are as expected');
            });

            test('creates a workspace with openwebui on host-based cluster', async function () {
                // OpenWebUI only works on host-based clusters
                const clusterConfig = await clusterUtils.getClusterReportedConfig();
                const isHostBased = clusterConfig?.subdomainEnabled === true;
                if (!isHostBased) {
                    logger.info('Skipping openwebui test — cluster is not host-based');
                    return;
                }

                // Step 1: Create a mock LLM inference workload
                const mockLlmName = objectGenerator.generateNameByPrefix('mock-llm');
                const mockLlmSpec: InferenceSpecSpec = {
                    image: 'zerob13/mock-openai-api:latest',
                    servingPort: {
                        container: 3000
                    }
                };
                const [mockLlmWorkloadId, mockLlmWorkloadName] = await inferenceUtils.createInference(
                    mockLlmName,
                    mockLlmSpec,
                    String(project.id)
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(mockLlmWorkloadId, Phase.Running, 120);

                const mockLlmInternalUrl = `http://${mockLlmWorkloadName}.${project.status.namespace}.svc.cluster.local/v1`;

                // Step 2: Get the openwebui environment
                const openwebuiEnv = await environmentUtils.get('openwebui');
                const connectionName = openwebuiEnv.spec.connections[0]?.name;
                const environmentId = openwebuiEnv.meta.id;
                const computeId = await computeResourceUtils.getId('cpu-only');

                // Step 3: Create workspace with openwebui env vars
                const workspacePayload: WorkspaceCreationRequestV2 = {
                    meta: {
                        name: workloadName,
                        useGivenNameAsPrefix: false,
                        projectId: project.id.toString(),
                        clusterId: project.clusterId
                    },
                    assets: {
                        environment: environmentId,
                        compute: computeId
                    },
                    specificRunParams: {
                        connections: [
                            {
                                name: connectionName,
                                authorizedUsers: [config.controlPlaneAdminUsername, 'user1', 'user2']
                            }
                        ],
                        environmentVariables: [
                            {
                                name: 'OPENAI_API_BASE_URL',
                                value: mockLlmInternalUrl
                            }
                        ]
                    }
                };

                const resultWorkload = await assetsApiV2Utils.createWorkspace(workspacePayload);
                workloadName = resultWorkload.body.meta.name;
                workloadId = resultWorkload.body.meta.workloadId;

                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                expect(workload.body.environments).to.have.lengthOf(1);
                expect(workload.body.environments[0].connections).to.have.length.greaterThan(0);

                const openwebuiConnection = workload.body.environments[0].connections.find(
                    (conn) => conn.url
                );
                expect(openwebuiConnection).to.not.be.undefined;
                logger.info(`OpenWebUI URL: ${openwebuiConnection.url}`);

                const authorizedConnection = workload.body.environments[0].connections.find(
                    (conn) => conn.authorizedUsers?.length > 0
                );
                expect(authorizedConnection).to.not.be.undefined;
                expect(authorizedConnection.authorizedUsers).to.include(config.controlPlaneAdminUsername);
                expect(authorizedConnection.authorizedUsers).to.include('user1');
                expect(authorizedConnection.authorizedUsers).to.include('user2');
                logger.info('openwebui workspace connections with authorized users are as expected');
            });
        }
    );
});
