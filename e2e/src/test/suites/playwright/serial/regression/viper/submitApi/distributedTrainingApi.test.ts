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
    workloadsManagerUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { StatusCodes } from 'http-status-codes';
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
    `Distributed training submission, tags: ${tags.resource.workloads} ${tags.operations.submit} ${tags.teams.viper}`,
    { tag: [tags.services.workloads, tags.services.agent, tags.runOn.serialCI] },
    function () {
        test.beforeAll('creating project', async function () {
            project = await projectResourcesUtils.createProjectAndNamespace('project');
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterAll('Cleanup', async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch {
                logger.error(`Failed to delete project ${project.name} and namespace ${project.namespace}`);
            }
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
