/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authV2Utils,
    config,
    flowRootState,
    versionUtils,
    logger,
    objectGenerator,
    trainingUtils,
    computeResourceUtils,
    environmentUtils,
    workloadsUtilsV2,
    serviceAccountUtils,
    authUtils,
    accessKeyUtils,
    orgUnitUtils
} from '../../../../../../modules/helpers';
import { tags } from '../../../../../../modules/helpers';
import { test, expect } from '@playwright/test';
import { kubeCli } from '../../../../../../modules/cli';
import { invalidUuid } from '../../../../../../modules/helpers/config/const';
import { AxiosResponse } from 'axios';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { TrainingV2 } from '../../../../../../modules/api/apis/assetsApi/generated/api';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { isOCP } from '../../../../../../modules/helpers/utils/EnvironmentUtils';

test.describe(
    `Workload auth validate redirect_uri middleware tests, tags: ${tags.teams.phoenix} ${tags.suites.legacySphinx} ${tags.runOn.k8sFlavCertKit}`,
    { tag: [tags.services.identityManager, tags.services.keycloak] },
    function () {
        let project: Project;
        let training: TrainingV2;
        
            let clientId: string;
            const maliciousRedirectUri = 'https://www.malicious-domain.com';
            const validRedirectUri = isOCP()
                ? `${config.appUrl}/test/test/oauth2/callback`
                : `${config.clusterUrl}/test/test/oauth2/callback`;
            const codeResponseType = 'code';
            let response: AxiosResponse;

            const logSchedulingDiagnostics = async (
                workloadId: string,
                workloadName: string
            ): Promise<void> => {
                try {
                    const nodes = await kubeCli.CoreV1Api.listNode();
                    for (const node of nodes.items) {
                        const conditions = (node.status?.conditions ?? [])
                            .filter((c) =>
                                ['Ready', 'DiskPressure', 'MemoryPressure', 'PIDPressure'].includes(
                                    c.type ?? ''
                                )
                            )
                            .map((c) => `${c.type}=${c.status}`)
                            .join(', ');
                        const alloc = node.status?.allocatable ?? {};
                        logger.warn(
                            `[scheduling] node=${node.metadata?.name} conditions=[${conditions}] ` +
                                `cpu alloc=${alloc.cpu} memory alloc=${alloc.memory} pods alloc=${alloc.pods}`
                        );
                    }
                    const workload = (await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId)).body;
                    logger.warn(
                        `[scheduling] workload=${workloadName} phase=${workload.phase} namespace=${workload.namespace}`
                    );
                    const pods = await kubeCli.listPodsByNamepsace(workload.namespace);
                    for (const pod of pods) {
                        const scheduled = (pod.status?.conditions ?? []).find(
                            (c) => c.type === 'PodScheduled'
                        );
                        logger.warn(
                            `[scheduling] pod=${pod.metadata?.name} phase=${pod.status?.phase} ` +
                                `PodScheduled=${scheduled?.status} reason=${scheduled?.reason} message=${scheduled?.message}`
                        );
                    }
                    const events = await kubeCli.CoreV1Api.listNamespacedEvent({
                        namespace: workload.namespace
                    });
                    for (const event of events.items) {
                        if (event.type === 'Warning') {
                            logger.warn(
                                `[scheduling] event reason=${event.reason} object=${event.involvedObject?.name} message=${event.message}`
                            );
                        }
                    }
                } catch (err) {
                    logger.warn(`[scheduling] failed to collect diagnostics: ${err.message}`);
                }
            };

            test('Should return login form for valid auth request', async () => {
                clientId = `runai-internal.${flowRootState.cluster.uuid}.workloads_auth`;
                response = await authV2Utils.sendWorkloadAuthAuthRequest(
                    clientId,
                    validRedirectUri,
                    codeResponseType
                );
                expect(response.status, 'valid auth request should return 200 OK').toBe(200);
                expect(response.data, 'login form should contain Sign in heading').toContain('Sign in to your account');
                expect(response.data, 'login form should contain Email field').toContain('Email');
                expect(response.data, 'login form should contain Password field').toContain('Password');
                expect(response.data, 'login form should contain Remember me option').toContain('Remember me');
                expect(response.data, 'login form should contain Forgot Password link').toContain('Forgot Password?');
            });

            test('Should fail in case of invalid redirect_uri', async () => {
                clientId = `runai-internal.${flowRootState.cluster.uuid}.workloads_auth`;
                response = await authV2Utils.sendWorkloadAuthAuthRequest(clientId, maliciousRedirectUri);
                expect(response.status, 'malicious redirect_uri should result in an error response (>=400)').toBeGreaterThanOrEqual(400);
                expect(response.data, 'response should indicate invalid redirect_uri parameter').toContain('Invalid parameter: redirect_uri');
            });

            test('Should fail in case of wrong cluster', async () => {
                clientId = `runai-internal.${invalidUuid}.workloads_auth`;
                response = await authV2Utils.sendWorkloadAuthAuthRequest(clientId, maliciousRedirectUri);
                expect(response.status, 'wrong cluster in client_id should result in an error response (>=400)').toBeGreaterThanOrEqual(400);
                expect(response.data, 'response should indicate failure to get cluster URL for unknown cluster uuid').toContain('Failed to get cluster URL');
            });

            test('Should fail in case of invalid client_id', async () => {
                clientId = 'runai-internal.invalid.workloads_auth';
                response = await authV2Utils.sendWorkloadAuthAuthRequest(clientId, maliciousRedirectUri);
                expect(response.status, 'invalid client_id should result in an error response (>=400)').toBeGreaterThanOrEqual(400);
                expect(response.data, 'response should indicate invalid client_id parameter').toContain('Invalid parameter: client_id');
            });

            test('Should fail in keycloak in case of missing redirect_uri', async () => {
                test.skip(
                    !versionUtils.isVersionSupported(flowRootState.cluster.version, '2.23'),
                    'Requires version 2.23+'
                );
                clientId = `runai-internal.${flowRootState.cluster.uuid}.workloads_auth`;
                response = await authV2Utils.sendWorkloadAuthAuthRequest(clientId);
                expect(response.status, 'missing redirect_uri should result in an error response (>=400)').toBeGreaterThanOrEqual(400);
                expect(response.data.error, 'error response should indicate missing redirect_uri parameter').toContain('Invalid parameter: redirect_uri');
            });

            test('Should allow missing redirect_uri for * urls', async () => {
                clientId = 'runai-internal.invalid.workloads_auth';
                response = await authV2Utils.sendWorkloadAuthAuthRequest(clientId);
                expect(response.status, 'unknown client_id with no redirect_uri should result in an error response (>=400)').toBeGreaterThanOrEqual(400);
                expect(response.data.error, 'error response should indicate client not found for unknown client_id').toContain('Client not found.');
            });

            test('Should create jupyter workload and connect to it with service account token', async () => {
                test.setTimeout(5 * 60 * 1000);
                const projectName = objectGenerator.generateShortNameByPrefix('jupyter');
                const workloadName = objectGenerator.generateShortNameByPrefix('jupyterw');

                project = await orgUnitUtils.createProject({
                    name: projectName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 1, limit: 1 } }]
                });
                logger.info(`Created project: ${projectName} with ID: ${project.id}`);
    
                await orgUnitUtils.waitForProjectFromBackendToBeReady(projectName, project.id.toString(), 120);
                logger.info(`Project ${projectName} is now ready and propagated to all services`);

                const isHostBased = await environmentUtils.isHostBasedRouting();
                const jupyterArgs = isHostBased
                    ? '--NotebookApp.token=""'
                    : '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""';

                training = await trainingUtils.createTraining({
                    meta: {
                        name: workloadName,
                        projectId: project.id.toString(),
                        clusterId: project.clusterId
                    },
                    assets: {
                        environment: await environmentUtils.getId('jupyter-lab'),
                        compute: await computeResourceUtils.getId('cpu-only')
                    },
                    specificRunParams: {
                        command: 'start-notebook.sh',
                        args: jupyterArgs,
                        environmentVariables: [],
                        nodePools: ['default'],
                        autoDeletionTimeAfterCompletionSeconds: 2592000,
                        backoffLimit: 6,
                        annotations: [],
                        labels: [],
                        tolerations: [],
                        connections: [
                            {
                                name: 'Jupyter'
                            }
                        ],
                        completions: 1,
                        allowOverQuota: false
                    }
                });
                logger.info(`Created Jupyter workload: ${workloadName}`);
                try {
                    await workloadsUtilsV2.waitForApiWorkloadWithPhase(
                        training.meta.workloadId,
                        Phase.Running,
                        workloadName,
                        true,
                        180
                    );
                } catch (err) {
                    await logSchedulingDiagnostics(training.meta.workloadId, workloadName);
                    throw err;
                }
                logger.info('Jupyter workload is running');
                expect(training.meta.name, 'created training workload name should match the requested name').toBe(workloadName);
                const workloadDetails = await workloadsUtilsV2.getWorkloadByWorkloadId(
                    training.meta.workloadId
                );
                const workload = workloadDetails.body;
                logger.info('Workload Connection Details:');
                logger.info(`URLs: ${JSON.stringify(workload.urls)}`);
                const serviceAccount = await serviceAccountUtils.createServiceAccount(
                    flowRootState.users.system_administrator.accessToken,
                    { name: objectGenerator.generateNameByPrefix('jupyter-sa') }
                );
                const serviceAccountToken = await authUtils.authenticateAppAndGetToken(
                    serviceAccount.body.clientId,
                    serviceAccount.body.clientSecret
                );

                const jupyterClient = await authV2Utils.connectToJupyterWithAppCredentials(
                    workload.urls[0],
                    serviceAccountToken.accessToken
                );
                logger.info(`Jupyter client: ${JSON.stringify(jupyterClient)}`);
                const response = await jupyterClient.client.get('/api/kernelspecs');
                logger.info(`Jupyter response: ${JSON.stringify(response.data)}`);
                expect(response.status, 'Jupyter kernelspecs API should return 200 OK with service account token').toBe(200);
                expect(response.data, 'Jupyter kernelspecs response should have kernelspecs property').toHaveProperty('kernelspecs');
                expect(response.data.default, 'default kernel should be python3').toBe('python3');
                await workloadsUtilsV2.deleteWorkload(training.meta.workloadId, WorkloadType.TRAINING);
                await serviceAccountUtils.deleteServiceAccount(
                    flowRootState.users.system_administrator.accessToken,
                    serviceAccount.body.id
                );
                await orgUnitUtils.deleteProjectAndNamespace(
                    project.id,
                    project.name,
                    flowRootState.users.system_administrator.accessToken
                );
            });

            test('Should create jupyter workload and connect to it with access key token', async () => {
                test.setTimeout(5 * 60 * 1000);
                const projectName = objectGenerator.generateShortNameByPrefix('jupyter');
                const workloadName = objectGenerator.generateShortNameByPrefix('jupyter');

                project = await orgUnitUtils.createProject({
                    name: projectName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 1, limit: 1 } }]
                });
                logger.info(`Created project: ${projectName} with ID: ${project.id}`);

                logger.info(`Waiting for project ${projectName} to be ready...`);
                await orgUnitUtils.waitForProjectFromBackendToBeReady(projectName, project.id.toString(), 120);
                logger.info(`Project ${projectName} is now ready and propagated to all services`);

                const isHostBased = await environmentUtils.isHostBasedRouting();
                const jupyterArgs = isHostBased
                    ? '--NotebookApp.token=""'
                    : '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""';

                training = await trainingUtils.createTraining({
                    meta: {
                        name: workloadName,
                        projectId: project.id.toString(),
                        clusterId: project.clusterId
                    },
                    assets: {
                        environment: await environmentUtils.getId('jupyter-lab'),
                        compute: await computeResourceUtils.getId('cpu-only')
                    },
                    specificRunParams: {
                        command: 'start-notebook.sh',
                        args: jupyterArgs,
                        environmentVariables: [],
                        nodePools: ['default'],
                        autoDeletionTimeAfterCompletionSeconds: 2592000,
                        backoffLimit: 6,
                        annotations: [],
                        labels: [],
                        tolerations: [],
                        connections: [
                            {
                                name: 'Jupyter'
                            }
                        ],
                        completions: 1,
                        allowOverQuota: false
                    }
                });
                logger.info(`Created Jupyter workload: ${workloadName}`);
                try {
                    await workloadsUtilsV2.waitForApiWorkloadWithPhase(
                        training.meta.workloadId,
                        Phase.Running,
                        workloadName,
                        true,
                        180
                    );
                } catch (err) {
                    await logSchedulingDiagnostics(training.meta.workloadId, workloadName);
                    throw err;
                }
                logger.info('Jupyter workload is running');
                expect(training.meta.name, 'created training workload name should match the requested name').toBe(workloadName);
                const workloadDetails = await workloadsUtilsV2.getWorkloadByWorkloadId(
                    training.meta.workloadId
                );
                const workload = workloadDetails.body;
                logger.info('Workload Connection Details:');
                logger.info(`URLs: ${JSON.stringify(workload.urls)}`);
                const accessKey = await accessKeyUtils.createAccessKey(
                    flowRootState.users.system_administrator.accessToken,
                    { name: objectGenerator.generateNameByPrefix('jupyter-personal-app') }
                );
                const appToken = await authUtils.authenticateAppAndGetToken(
                    accessKey.body.clientId,
                    accessKey.body.clientSecret
                );

                const jupyterClient = await authV2Utils.connectToJupyterWithAppCredentials(
                    workload.urls[0],
                    appToken.accessToken
                );
                logger.info(`Jupyter client: ${jupyterClient}`);
                const response = await jupyterClient.client.get('/api/kernelspecs');
                logger.info(`Jupyter response: ${JSON.stringify(response.data)}`);
                expect(response.status, 'Jupyter kernelspecs API should return 200 OK with access key token').toBe(200);
                expect(response.data, 'Jupyter kernelspecs response should have kernelspecs property').toHaveProperty('kernelspecs');
                expect(response.data.default, 'default kernel should be python3').toBe('python3');
                await workloadsUtilsV2.deleteWorkload(training.meta.workloadId, WorkloadType.TRAINING);
                await accessKeyUtils.deleteAccessKey(
                    flowRootState.users.system_administrator.accessToken,
                    accessKey.body.id
                );
                await orgUnitUtils.deleteProjectAndNamespace(
                    project.id,
                    project.name,
                    flowRootState.users.system_administrator.accessToken
                );
            });
        }
    );
