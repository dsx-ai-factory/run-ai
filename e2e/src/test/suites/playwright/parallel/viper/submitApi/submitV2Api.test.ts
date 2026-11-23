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
    crds,
    computeResourceUtils,
    environmentUtils,
    assetsApiV2Utils,
    objectGenerator,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { StatusCodes } from 'http-status-codes';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { WorkloadStatus } from '../../../../../../modules/helpers/models/baseWorkload';
import { EnvironmentCreationRequest, Scope } from '../../../../../../modules/api/apis/assetsApi';
import { description, tag } from 'allure-js-commons';
import { test } from '@playwright/test';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { kubeCli } from '../../../../../../modules/cli';

let project: Project;
let workloadId: string;
let workloadName: string;

const WorkloadTypeParams = [
    {
        workloadType: 'training',
        workloadName: 'test-training',
        createFunction: assetsApiV2Utils.createTraining,
        getFunction: workloadsCrudUtils.getTraining,
        deleteFunction: workloadsCrudUtils.deleteTraining,
        jobKind: crds.runaiTrainingWorkloadCrd.kind
    },
    {
        workloadType: 'interactive',
        workloadName: 'test-interactive',
        createFunction: assetsApiV2Utils.createWorkspace,
        getFunction: workloadsCrudUtils.getWorkspace,
        deleteFunction: workloadsCrudUtils.deleteWorkspace,
        jobKind: crds.runaiInteractiveWorkloadCrd.kind
    }
];

WorkloadTypeParams.forEach(function (params) {
    let anyWorkloadSubmissionEnabled: boolean;

    test.describe(
        `Submit ${params.workloadType} V2 | ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp}`,
        { tag: [tags.services.workloads, tags.services.agent, tags.runOn.parallelCI] },
        function () {
            test.beforeAll('creating project', async function () {
                project = await projectResourcesUtils.createProjectAndNamespace('project');
                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
                assetsApiV2Utils.accessToken = flowRootState.users.system_administrator.accessToken;
                anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
                    params.workloadType === 'interactive'
                        ? AnyWorkloadSubmissionType.Workspace
                        : AnyWorkloadSubmissionType.Training
                );
            });

            test.afterAll('Cleanup', async function () {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch {
                    logger.error(
                        `Failed to delete project ${project.name} and namespace ${project.namespace}`
                    );
                }
            });

            async function createEnvAsset(): Promise<any> {
                const environmentCreationRequest: EnvironmentCreationRequest = {
                    meta: {
                        name: objectGenerator.generateNameByPrefix('environment'),
                        scope: Scope.Tenant,
                        workloadSupportedTypes: {
                            workspace: true,
                            training: true
                        }
                    },
                    spec: {
                        image: (await environmentUtils.get('tensorboard')).spec.image,
                        tty: true,
                        stdin: true
                    }
                };

                return await environmentUtils.create(environmentCreationRequest);
            }

            async function getWorkloadCreationPayload(envAssetId): Promise<any> {
                return {
                    meta: {
                        name: params.workloadName,
                        useGivenNameAsPrefix: true,
                        projectId: project.id.toString(),
                        clusterId: flowRootState.cluster.uuid
                    },
                    assets: {
                        environment: envAssetId,
                        compute: await computeResourceUtils.getId('cpu-only')
                    }
                };
            }
            test('Create 2 workloads using the same PVC', async function (): Promise<void> {
                function addInfo(): void {
                    tag(`${tags.teams.viper}`);
                    description(
                        'Flow:\n' +
                            `1. Submit ${params.workloadType} workload using the V2 Api.\n` +
                            '2. Check:\n' +
                            '* request status Code\n' +
                            '* workload phase is "running"\n' +
                            '* pod status is "running"\n' +
                            '* workload receives TTY and Stdin from env \n' +
                            `3. Delete the ${params.workloadType} workload using the V1 Api`
                    );
                }
                await test.step(`Create ${params.workloadType} workload`, async () => {
                    addInfo();
                    const environmentAsset = await createEnvAsset();
                    const workloadPayload = await getWorkloadCreationPayload(environmentAsset.meta.id);
                    const result = await params.createFunction(workloadPayload);
                    workloadId = result.body.meta.workloadId;
                    workloadName = result.body.meta.name;

                    expect(result.response.statusCode).to.equal(StatusCodes.OK);
                    expect(result.body.meta.desiredPhase).to.equal(WorkloadStatus.RUNNING);
                    await waitUtils.waitForPods(
                        params.workloadName,
                        project.namespace,
                        1,
                        120,
                        ['Running'],
                        true
                    );

                    if (anyWorkloadSubmissionEnabled) {
                        const job = await kubeCli.getBatchJob(workloadName, project.namespace);
                        const container = job.spec?.template.spec?.containers[0];
                        expect(container?.tty).to.be.true;
                        expect(container?.stdin).to.be.true;

                        const workload = await workloadsUtilsV2.waitForBackendWorkloadByNameAndProjectName(
                            workloadName,
                            project.name
                        );
                        expect(workload.sourceApi).to.equal('WorkloadsV2');
                    } else {
                        const workload = await params.getFunction(workloadId);
                        expect(workload.body.spec.tty).to.be.true;
                        expect(workload.body.spec.stdin).to.be.true;
                    }
                });

                await test.step(`Delete ${params.workloadType} workload`, async () => {
                    addInfo();
                    const deleteResult = await params.deleteFunction(workloadId);
                    expect(deleteResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                    await waitUtils.waitForWorkloadToBeDeleted(workloadName);
                });
            });
        }
    );
});
