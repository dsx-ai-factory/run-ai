/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    flowRootState,
    logger,
    tags,
    operations,
    projectResourcesUtils,
    workloadsCrudUtils,
    waitUtils,
    computeResourceUtils,
    environmentUtils,
    assetsApiV2Utils,
    dataSourcesUtils,
    workloadsManagerUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { StatusCodes } from 'http-status-codes';
import { AssetKind } from '../../../../../../../modules/api/apis/assetsApi';
import { DISTRIBUTED_TOOLS_LABELS } from '../../../../../../../modules/ui/runaiV2/pages/TrainingPage';
import { assetsApi } from '../../../../../../../modules/api';
import { test } from '@playwright/test';
import {
    CleanPodPolicy,
    DistributedFramework,
    DistributedSpecSpec
} from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

let project: Project;

test.describe(
    `Submit distributed training V2, tags: ${tags.resource.workloads} ${tags.operations.submit} ${tags.teams.viper}`,
    { tag: [tags.services.workloads, tags.services.agent, tags.runOn.serialCI] },
    function () {
        test.beforeAll('creating project', async function () {
            project = await projectResourcesUtils.createProjectAndNamespace('project');
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;
            environmentUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            dataSourcesUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            computeResourceUtils.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterAll('Cleanup', async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch {
                logger.error(`Failed to delete project ${project.name} and namespace ${project.namespace}`);
            }
        });

        async function getDistributedTrainingCreationPayload(
            environmentId: string,
            computeId?: string,
            dataSources?: { id: string; kind: AssetKind }[],
            master?: { sameAsWorker: boolean }
        ): Promise<any> {
            return {
                meta: {
                    name: 'distributed-training',
                    useGivenNameAsPrefix: true,
                    projectId: project.id.toString(),
                    clusterId: flowRootState.cluster.uuid
                },
                assets: {
                    environment: environmentId,
                    command: 'sleep infinity',
                    compute: computeId || (await computeResourceUtils.getId('cpu-only')),
                    dataSources: dataSources || [] // Add dataSources if provided
                },
                distributed: {
                    numWorkers: 1,
                    framework: DISTRIBUTED_TOOLS_LABELS.PY_TORCH,
                    master: master || null
                }
            };
        }

        test('create assets and then create distributed training', async function () {
            const environmentId = (await environmentUtils.createEnvironmentWithArbitraryFields(project.id))
                .meta.id;
            const computeId = (await computeResourceUtils.createComputeResourceWithArbitraryFields()).meta.id;
            const hostPathId = (
                await dataSourcesUtils.createHostPathAsset(dataSourcesUtils.createHostPathPayload(project.id))
            ).meta.id;

            const result = await assetsApiV2Utils.createTraining(
                await getDistributedTrainingCreationPayload(environmentId, computeId, [
                    {
                        id: hostPathId,
                        kind: AssetKind.HostPath
                    }
                ])
            );
            expect(result.response.statusCode).to.equal(StatusCodes.OK);
            const workloadName = result.body.meta.name;
            await waitUtils.waitForPods(workloadName, project.namespace, 1, 90, ['Running']);
            expect(
                (await workloadsCrudUtils.deleteTraining(result.body.meta.workloadId)).response.statusCode
            ).to.equal(StatusCodes.ACCEPTED);

            await waitUtils.waitForWorkloadToBeDeleted(workloadName);
            await assetsApi.hostPathApi.deleteHostPathById(hostPathId);
            await environmentUtils.delete(environmentId);
            await computeResourceUtils.delete(computeId);
        });

        test('create distributed training with master', async function () {
            const environmentId = (await environmentUtils.createEnvironmentWithArbitraryFields(project.id))
                .meta.id;
            const result = await assetsApiV2Utils.createTraining(
                await getDistributedTrainingCreationPayload(environmentId, undefined, undefined, {
                    sameAsWorker: true
                })
            );
            expect(result.response.statusCode).to.equal(StatusCodes.OK);
            const workloadName = result.body.meta.name;
            await waitUtils.waitForPods(workloadName, project.namespace, 2, 90, ['Running']);
            expect(
                (await workloadsCrudUtils.deleteTraining(result.body.meta.workloadId)).response.statusCode
            ).to.equal(StatusCodes.ACCEPTED);
            await waitUtils.waitForWorkloadToBeDeleted(result.body.meta.name);
            await environmentUtils.delete(environmentId);
        });

        test('distributed training with clean pod policy', async function () {
            const workloadNamePrefix = 'dist-clean-pod-policy';
            const distSpec: DistributedSpecSpec = {
                numWorkers: 1,
                cleanPodPolicy: CleanPodPolicy.All,
                distributedFramework: DistributedFramework.PyTorch,
                image: 'ubuntu',
                command: 'sleep infinity'
            };
            const createResult = await workloadsCrudUtils.createDistributed(
                workloadNamePrefix,
                true,
                flowRootState.cluster.uuid,
                project.id.toString(),
                distSpec,
                true
            );
            expect(createResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);
            const workloadId = createResult.body.workloadId;

            if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Distributed)) {
                const getResult = await workloadsManagerUtils.getWorkloadV2ById(workloadId);
                expect(
                    getResult.body.workload.workloadManifest['spec']['runPolicy']['cleanPodPolicy']
                ).to.equal(CleanPodPolicy.All);
            } else {
                const getResult = await workloadsCrudUtils.getDistributed(workloadId);
                expect(getResult.body.spec.cleanPodPolicy).to.equal(CleanPodPolicy.All);
            }

            const workloadName = createResult.body.name;
            await waitUtils.waitForPods(workloadName, project.namespace, 2, 90, ['Running']);

            expect((await workloadsCrudUtils.deleteDistributed(workloadId)).response.statusCode).to.equal(
                StatusCodes.ACCEPTED
            );
            await waitUtils.waitForWorkloadToBeDeleted(workloadName);
        });
    }
);
