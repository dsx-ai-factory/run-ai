/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import {
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadActionsUtils,
    workloadsCrudUtils,
    workloadUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { WorkloadActionType } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadActionType';
import { BatchResponse } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/batchResponse';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Batch workload action history ${tags.resource.workloads} ${tags.tool.api} ${tags.teams.viper}`, function () {
    supportedFrom(SupportedVersions.V3_1)(function () {
        let project: Project;

        test.beforeAll('create project', async function () {
            const systemAdministrator = flowRootState.users.system_administrator;
            project = await orgUnitUtils.createProjectAndValidate({
                name: objectGenerator.generateShortNameByPrefix('wl-actions-batch'),
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 0, limit: 0 } }]
            });
            workloadsCrudUtils.accessToken = systemAdministrator.accessToken;
        });

        test.afterAll('delete project and namespace', async function () {
            if (project === undefined) {
                return;
            }

            try {
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            } catch (error) {
                logger.error(`Failed to delete project ${project.name}; error ${error}`);
            }
        });

        test('records actions only for accepted batch items', async function () {
            test.setTimeout(300 * 1000);
            const acceptedAfter = new Date();
            const expectedActions: WorkloadActionType[] = [];
            const projectId = String(project.id);

            const createTrainingWorkload = async (name: string): Promise<{ id: string; name: string }> => {
                const result = await workloadsCrudUtils.createTraining(
                    objectGenerator.generateShortNameByPrefix(name),
                    false,
                    flowRootState.cluster.uuid,
                    projectId,
                    { image: 'ubuntu', command: 'sleep', args: '500' }
                );
                expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);
                return { id: result.body.workloadId, name: result.body.name };
            };

            const firstWorkload = await test.step('Create first training workload', function () {
                return createTrainingWorkload('action-training-first');
            });

            const secondWorkload = await test.step('Create second training workload', function () {
                return createTrainingWorkload('action-training-second');
            });

            const workloadRejectedByResume =
                await test.step('Create running workload for rejected resume', function () {
                    return createTrainingWorkload('action-training-rejected');
                });

            const workloads = [firstWorkload, secondWorkload];

            await test.step('Verify create action history', async function () {
                expectedActions.push(WorkloadActionType.Create);
                await workloadActionsUtils.waitForBatchActions(workloads, expectedActions, acceptedAfter);
            });

            const historyBeforeRejectedResume =
                await test.step('Save running workload history before rejected resume', function () {
                    return workloadActionsUtils.waitForActions(
                        workloadRejectedByResume.id,
                        workloadRejectedByResume.name,
                        [WorkloadActionType.Create],
                        acceptedAfter
                    );
                });

            const workloadIds = workloads.map((workload) => workload.id);

            const waitForWorkloadsPhase = async (ids: string[], phase: Phase): Promise<void> => {
                await Promise.all(
                    ids.map((workloadId) =>
                        waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, phase, 120)
                    )
                );
            };

            await test.step('Wait for workloads to reach Running', function () {
                return waitForWorkloadsPhase([...workloadIds, workloadRejectedByResume.id], Phase.Running);
            });

            const runBatchAction = async (
                name: string,
                action: WorkloadActionType,
                request: () => Promise<BatchResponse>,
                expectedFailure?: { id: string; code: number }
            ): Promise<void> => {
                await test.step(`${name} workloads`, async function () {
                    workloadActionsUtils.validateBatchResult(await request(), workloadIds, expectedFailure);
                });
                await test.step(`Verify ${action} action history`, async function () {
                    expectedActions.push(action);
                    await workloadActionsUtils.waitForBatchActions(workloads, expectedActions, acceptedAfter);
                });
            };

            await runBatchAction('Suspend', WorkloadActionType.Suspend, () =>
                workloadUtils.suspendMultipleWorkloads(workloadIds)
            );
            await test.step('Wait for accepted workloads to reach Stopped', function () {
                return waitForWorkloadsPhase(workloadIds, Phase.Stopped);
            });
            await runBatchAction(
                'Resume',
                WorkloadActionType.Resume,
                () => workloadUtils.resumeMultipleWorkloads([...workloadIds, workloadRejectedByResume.id]),
                { id: workloadRejectedByResume.id, code: StatusCodes.BAD_REQUEST }
            );

            await test.step('Verify rejected resume was not recorded', async function () {
                const historyAfterRejectedResume = (
                    await workloadActionsUtils.getWorkloadActions(workloadRejectedByResume.id)
                ).actions;
                expect(
                    historyAfterRejectedResume,
                    'the rejected resume must not be added to action history'
                ).to.deep.equal(historyBeforeRejectedResume);
            });

            await runBatchAction('Delete', WorkloadActionType.Delete, () =>
                workloadUtils.deleteMultipleWorkloads(workloadIds)
            );
        });
    });
});
