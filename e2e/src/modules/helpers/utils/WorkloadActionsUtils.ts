/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { StatusCodes } from 'http-status-codes';
import jwt_decode from 'jwt-decode';
import { validate as isUuid } from 'uuid';
import { workloadsService } from '../../api';
import { BatchResponse } from '../../api/apis/workloadsApi/generated/model/batchResponse';
import { GetWorkloadActionsResponse } from '../../api/apis/workloadsApi/generated/model/getWorkloadActionsResponse';
import { WorkloadActionRecord } from '../../api/apis/workloadsApi/generated/model/workloadActionRecord';
import { WorkloadActionType } from '../../api/apis/workloadsApi/generated/model/workloadActionType';
import { flowRootState, waitUtils } from '..';

export interface WorkloadReference {
    id: string;
    name: string;
}

export class WorkloadActionsUtils {
    isWorkloadTypeIn<TParams extends { workloadType: string }, TType extends TParams['workloadType']>(
        params: TParams,
        workloadTypes: readonly TType[]
    ): params is Extract<TParams, { workloadType: TType }> {
        return (workloadTypes as readonly string[]).includes(params.workloadType);
    }

    validateBatchResult(
        result: BatchResponse,
        expectedSucceeded: string[],
        expectedFailure?: { id: string; code: number }
    ): void {
        const succeeded = result.succeeded ?? [];
        const failed = result.failed ?? [];

        expect(succeeded).to.have.members(expectedSucceeded);
        if (expectedFailure === undefined) {
            expect(failed).to.be.empty;
            return;
        }

        expect(failed).to.have.lengthOf(1);
        expect(failed[0]).to.include(expectedFailure);
    }

    async waitForAcceptedOperation(
        operation: () => Promise<{ response: { statusCode?: number } }>,
        description: string,
        timeout = 120
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                try {
                    const result = await operation();
                    return result.response.statusCode === StatusCodes.ACCEPTED;
                } catch (error) {
                    if ((error as { statusCode?: number }).statusCode === StatusCodes.BAD_REQUEST) {
                        return false;
                    }
                    throw error;
                }
            },
            description,
            true,
            timeout
        );
    }

    async waitForBatchActions(
        workloads: WorkloadReference[],
        expectedActions: WorkloadActionType[],
        acceptedAfter: Date
    ): Promise<void> {
        await Promise.all(
            workloads.map((workload) =>
                this.waitForActions(workload.id, workload.name, expectedActions, acceptedAfter)
            )
        );
    }

    async getWorkloadActions(
        workloadId: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<GetWorkloadActionsResponse> {
        workloadsService.workloadActionsApi.accessToken = accessToken;
        return (await workloadsService.workloadActionsApi.getWorkloadActionsHistory(workloadId)).body;
    }

    async waitForActions(
        workloadId: string,
        workloadName: string,
        expectedActions: WorkloadActionType[],
        acceptedAfter: Date,
        accessToken = flowRootState.users.system_administrator.accessToken,
        timeout = 60
    ): Promise<WorkloadActionRecord[]> {
        const expected = [...expectedActions].reverse();
        const actions = (await waitUtils.waitForObject(
            async (): Promise<WorkloadActionRecord[] | undefined> => {
                try {
                    const actions = (await this.getWorkloadActions(workloadId, accessToken)).actions;
                    const matches =
                        actions.length === expected.length &&
                        actions.every((record, index) => record.action === expected[index]);
                    return matches ? actions : undefined;
                } catch {
                    return undefined;
                }
            },
            `workload ${workloadId} action history to equal ${expectedActions.join(', ')}`,
            timeout
        )) as WorkloadActionRecord[];

        this.validateActions(actions, workloadId, workloadName, acceptedAfter, accessToken);
        return actions;
    }

    private validateActions(
        actions: WorkloadActionRecord[],
        workloadId: string,
        workloadName: string,
        acceptedAfter: Date,
        accessToken: string
    ): void {
        const subject = jwt_decode<{ sub: string }>(accessToken).sub;
        const recordIds = new Set(actions.map((action) => action.id));

        expect(recordIds.size).to.equal(actions.length);
        for (const action of actions) {
            expect(isUuid(action.id)).to.be.true;
            expect(action).to.include({ workloadId, workloadName, subject, subjectType: 'User' });
            expect(action.acceptedAt.getTime()).to.be.within(acceptedAfter.getTime(), Date.now());
        }
    }
}
