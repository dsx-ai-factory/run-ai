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
    workloadsManagerUtils,
    waitUtils,
    crds
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { WorkloadStatus } from '../../../../../../modules/helpers/models/baseWorkload';
import { StatusCodes } from 'http-status-codes';
import { Training } from '../../../../../../modules/api/apis/workloadsApi/generated/model/training';
import { DistributedWorkload } from '../../../../../../modules/api/apis/workloadsApi/generated/model/distributedWorkload';
import { Workspace } from '../../../../../../modules/api/apis/workloadsApi/generated/model/workspace';
import { kubeCli } from '../../../../../../modules/cli';
import { test } from '@playwright/test';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

const ARBITRARY_BACKOFF_LIMIT = 5;
// const ARBITRARY_DISTRIBUTED_FRAMEWORK = DistributedFramework.PyTorch;

const workloadSpec = {
    image: 'ubuntu',
    command: 'sleep',
    args: '500',
    backoffLimit: ARBITRARY_BACKOFF_LIMIT
};

// const distributedWorkloadSpec = {
//     image: 'ubuntu',
//     command: 'sleep',
//     args: '500',
//     numWorkers: 1,
//     distributedFramework: ARBITRARY_DISTRIBUTED_FRAMEWORK,
//     backoffLimit: ARBITRARY_BACKOFF_LIMIT
// };

const WorkloadTypeParams = [
    {
        workloadType: 'training',
        getFunction: workloadsCrudUtils.getTraining,
        createWorkloadFunc: workloadsCrudUtils.createTraining,
        deleteWorkloadFunc: workloadsCrudUtils.deleteTraining,
        spec: workloadSpec,
        crd: crds.runaiTrainingWorkloadCrd
    },
    {
        workloadType: 'interactive',
        getFunction: workloadsCrudUtils.getWorkspace,
        createWorkloadFunc: workloadsCrudUtils.createWorkspace,
        deleteWorkloadFunc: workloadsCrudUtils.deleteWorkspace,
        spec: workloadSpec,
        crd: crds.runaiInteractiveWorkloadCrd
    }

    // {
    //     workloadType: 'distributed',
    //     getFunction: workloadsCrudUtils.getDistributed,
    //     createWorkloadFunc: workloadsCrudUtils.createDistributed,
    //     deleteWorkloadFunc: workloadsCrudUtils.deleteDistributed,
    //     spec: distributedWorkloadSpec
    // }
];

WorkloadTypeParams.forEach(function (params) {
    let project: Project;
    let workload: Training | DistributedWorkload | Workspace;

    test.describe(`${params.workloadType} workload BackOff limit - API, tags: ${tags.resource.workloads} ${tags.operations.submit} ${tags.teams.viper}`, function () {
        let anyWorkloadSubmissionEnabled: boolean;

        test.beforeAll('create project', async function () {
            project = await projectResourcesUtils.createProjectAndNamespace('project');
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
                params.workloadType === 'interactive'
                    ? AnyWorkloadSubmissionType.Workspace
                    : AnyWorkloadSubmissionType.Training
            );
        });

        test.afterAll('cleanup', async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch {
                logger.error(`Failed to delete project ${project.name} and namespace ${project.namespace}`);
            }
        });

        test(`Submit ${params.workloadType} with BackOff limit`, async () => {
            await test.step(`create ${params.workloadType} workload`, async function () {
                workload = (
                    await params.createWorkloadFunc(
                        params.workloadType,
                        true,
                        flowRootState.cluster.uuid,
                        String(project.id),
                        params.spec
                    )
                ).body;

                await waitUtils.waitForPods(workload.name, project.namespace, 1, 60, ['Running']);
            });

            await test.step(`wait for workload resource and check backoff limit | ${params.workloadType}`, async function () {
                if (anyWorkloadSubmissionEnabled) {
                    await waitUtils.waitForBatchJobToExist(workload.name, project.namespace);
                    const job = await kubeCli.getBatchJob(workload.name, project.namespace);
                    expect(job.spec?.backoffLimit).to.equal(ARBITRARY_BACKOFF_LIMIT);
                } else {
                    await waitUtils.waitForCrd(params.crd, workload.name, project.namespace);
                    const job = await kubeCli.getCrd(params.crd, workload.name, project.namespace);
                    expect(job['spec']['backoffLimit'].value).to.equal(ARBITRARY_BACKOFF_LIMIT);
                }
            });

            await test.step(`check backoff limit in GET request| ${params.workloadType}`, async function () {
                if (anyWorkloadSubmissionEnabled) {
                    const getResult = await workloadsManagerUtils.getWorkloadV2ById(workload.workloadId);
                    expect(getResult.response.statusCode).to.equal(StatusCodes.OK);
                    expect(getResult.body.workload.desiredPhase).to.equal(WorkloadStatus.RUNNING);
                    expect(getResult.body.workload.workloadManifest['spec']['backoffLimit']).to.equal(
                        ARBITRARY_BACKOFF_LIMIT
                    );
                } else {
                    const getResult = await params.getFunction(workload.workloadId);
                    expect(getResult.response.statusCode).to.equal(StatusCodes.OK);
                    expect(getResult.body.desiredPhase).to.equal(WorkloadStatus.RUNNING);
                    expect(getResult.body['spec']['backoffLimit']).to.equal(ARBITRARY_BACKOFF_LIMIT);
                }
            });

            await test.step(`delete ${params.workloadType} workload and verify deletion`, async function () {
                const deleteResult = await params.deleteWorkloadFunc(workload.workloadId);
                expect(deleteResult.response.statusCode).to.equal(StatusCodes.ACCEPTED);

                await waitUtils.waitForWorkloadToBeDeleted(workload.name);
            });
        });
    });
});
