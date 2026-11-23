/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    flowRootState,
    logger,
    tags,
    operations,
    projectResourcesUtils,
    workloadsCrudUtils,
    waitUtils,
    crds,
    computeResourceUtils,
    environmentUtils,
    objectGenerator,
    credentialUtils,
    dataSourcesUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { StatusCodes } from 'http-status-codes';
import {
    TrainingCreationRequest,
    TrainingSpecSpec
} from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import { kubeCli } from '../../../../../../../modules/cli';

let project: Project;

const WorkloadTypeParams = [
    {
        workloadType: 'training',
        workloadName: 'test-training',
        createFunction: workloadsCrudUtils.submitTraining,
        deleteFunction: workloadsCrudUtils.deleteTraining,
        jobKind: crds.runaiTrainingWorkloadCrd
    },
    {
        workloadType: 'interactive',
        workloadName: 'test-interactive',
        createFunction: workloadsCrudUtils.submitWorkspace,
        deleteFunction: workloadsCrudUtils.deleteWorkspace,
        jobKind: crds.runaiInteractiveWorkloadCrd
    }
];

test.describe(
    `Submit workloads with assets, tags: ${tags.operations.submit} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`,
    { tag: [tags.services.workloads, tags.services.agent, tags.runOn.serialCI] },
    () => {
        test.beforeAll(async () => {
            project = await projectResourcesUtils.createProjectAndNamespace('project');
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterAll(async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch {
                logger.error(`Failed to delete project ${project.name} and namespace ${project.namespace}`);
            }
        });

        async function getWorkloadCreationPayload(
            workloadName: string,
            environmentId?: string,
            computeId?: string,
            spec?: TrainingSpecSpec,
            dataSources?: string[]
        ): Promise<TrainingCreationRequest> {
            return {
                name: workloadName,
                useGivenNameAsPrefix: true,
                projectId: project.id.toString(),
                clusterId: flowRootState.cluster.uuid,
                assets: {
                    environment: environmentId || (await environmentUtils.getId('tensorboard')),
                    compute: computeId || (await computeResourceUtils.getId('cpu-only')),
                    datasources: dataSources || undefined
                },
                spec
            };
        }

        for (const params of WorkloadTypeParams) {
            test(`create assets and then create ${params.workloadType} workload with them`, async () => {
                const environmentId = (
                    await environmentUtils.createEnvironmentWithArbitraryFields(project.id)
                ).meta.id;
                const computeId = (await computeResourceUtils.createComputeResourceWithArbitraryFields()).meta
                    .id;

                const result = await params.createFunction(
                    await getWorkloadCreationPayload(params.workloadName, environmentId, computeId)
                );

                expect(result.response.statusCode).toBe(StatusCodes.ACCEPTED);

                const workloadName = result.body.name;
                const workloadId = result.body.workloadId;

                await waitUtils.waitForPods(workloadName, project.namespace, 1, 90, ['Running']);
                expect((await params.deleteFunction(workloadId)).response.statusCode).toBe(
                    StatusCodes.ACCEPTED
                );

                await waitUtils.waitForWorkloadToBeDeleted(workloadName);
                await environmentUtils.delete(environmentId);
                await computeResourceUtils.delete(computeId);
            });
        }

        test('should create training with environment variables based on credential', async () => {
            test.setTimeout(240 * 1000);

            const ak = await credentialUtils.createAccessKeyWithArbitraryFields();
            const workloadName = objectGenerator.generateNameByPrefix('training');
            const spec: TrainingSpecSpec = {
                environmentVariables: [
                    {
                        name: objectGenerator.generateNameByPrefix('ev'),
                        secret: { name: ak.clusterInfo.resources[0].name, key: 'AccessKeyId' }
                    }
                ]
            };
            const result = await workloadsCrudUtils.submitTraining(
                await getWorkloadCreationPayload(workloadName, undefined, undefined, spec)
            );

            expect(result.response.statusCode).toBe(StatusCodes.ACCEPTED);

            await waitUtils.waitForPods(workloadName, project.namespace, 1, 90, ['Running'], true);
            expect(
                (await workloadsCrudUtils.deleteTraining(result.body.workloadId)).response.statusCode
            ).toBe(StatusCodes.ACCEPTED);

            await waitUtils.waitForWorkloadToBeDeleted(workloadName);
            await credentialUtils.deleteAccessKeyAsset(ak.meta.id);
        });

        test('create config-map assets and then create training with it', async () => {
            test.setTimeout(240 * 1000);
            const workloadName = objectGenerator.generateNameByPrefix('training');
            const meta = await dataSourcesUtils.createConfigMapResource('cm-resource', project.namespace);
            await dataSourcesUtils.waitForConfigMapToBeVisible(meta.name, 'project', project.id.toString());
            const configMap = await dataSourcesUtils.createConfigMapWithArbitraryFields(meta, project.id);

            const result = await workloadsCrudUtils.submitTraining(
                await getWorkloadCreationPayload(workloadName, undefined, undefined, undefined, [
                    configMap.meta.id
                ])
            );

            await waitUtils.waitForPods(workloadName, project.namespace, 1, 90, ['Running'], true);
            expect(
                (await workloadsCrudUtils.deleteTraining(result.body.workloadId)).response.statusCode
            ).toBe(StatusCodes.ACCEPTED);

            await waitUtils.waitForWorkloadToBeDeleted(workloadName);
            await dataSourcesUtils.deleteConfigMapAsset(configMap.meta.id);
            await kubeCli.CoreV1Api.deleteNamespacedConfigMap({ name: meta.name, namespace: meta.namespace });
        });
    }
);
