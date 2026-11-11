/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState, settings } from '../config/config';
import { clusterApi, workloadsService } from '../../api';
import { WorkloadTelemetryType } from '../../api/apis/workloadsApi/generated/model/workloadTelemetryType';
import { TelemetryResponse } from '../../api/apis/workloadsApi/generated/model/telemetryResponse';
import { NodeTelemetryType } from '../../api/apis/clusterApi';
import { logger, waitUtils } from '../index';
import { Phase } from '../../api/apis/workloadsApi/generated/api';

export class TelemetryUtils {
    async getWorkloadTelemetryFilteredByClusterId(
        clusterId: string,
        telemetryType: WorkloadTelemetryType,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<TelemetryResponse> {
        workloadsService.workloadsApi.accessToken = accessToken;
        const res = await workloadsService.workloadsApi.getWorkloadsTelemetry(telemetryType, clusterId);
        return res.body;
    }

    async getWorkloadTelemetryUsingGroupBy(
        telemetryType: WorkloadTelemetryType,
        groupBy?: Array<
            'ClusterId' | 'DepartmentId' | 'ProjectId' | 'Type' | 'CurrentNodepools' | 'Phase' | 'Category'
        >,
        clusterId?: string,
        nodePoolName?: string,
        filterByPhase?: Array<Phase>,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<TelemetryResponse> {
        workloadsService.workloadsApi.accessToken = accessToken;
        if (!clusterId) {
            clusterId = settings.isSelfHosted ? undefined : flowRootState.cluster.uuid;
        }

        const res = await workloadsService.workloadsApi.getWorkloadsTelemetry(
            telemetryType,
            clusterId,
            nodePoolName,
            undefined, // departmentId
            undefined, // projectId
            groupBy,
            filterByPhase
        );
        return res.body;
    }

    async getNodeTelemetryFilteredByClusterId(
        clusterId: string,
        telemetryType: NodeTelemetryType,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<TelemetryResponse> {
        clusterApi.nodesApi.accessToken = accessToken;
        const res = await clusterApi.nodesApi.getNodeTelemetry(telemetryType, clusterId);
        return res.body;
    }

    async getNodeTelemetryUsingGroupBy(
        telemetryType: NodeTelemetryType,
        groupBy: Array<'ClusterId' | 'Nodepool' | 'Node'>,
        clusterId?: string,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<TelemetryResponse> {
        clusterApi.nodesApi.accessToken = accessToken;
        const res = await clusterApi.nodesApi.getNodeTelemetry(telemetryType, clusterId, undefined, groupBy);
        return res.body;
    }

    async waitForMinWorkloadTelemetryValue(
        telemetryType: WorkloadTelemetryType,
        minValue: number,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<void> {
        const doesTelemetryValueMeetCriteria = async (): Promise<boolean> => {
            const telemetry = await this.getWorkloadTelemetryFilteredByClusterId(
                flowRootState.cluster.uuid,
                telemetryType,
                accessToken
            );
            return Number(telemetry.values[0].value) >= minValue;
        };

        await waitUtils.waitForCondition(
            doesTelemetryValueMeetCriteria,
            'wait for telemetry value to meet criteria',
            true,
            60
        );
    }

    /**
     * Waits for the workload telemetry value to stabilize (same value on consecutive reads).
     * Useful after cleanup to ensure telemetry has caught up with deletions before capturing a baseline.
     */
    async waitForWorkloadTelemetryStabilization(
        clusterId: string,
        telemetryType: WorkloadTelemetryType,
        stableReadsRequired = 3,
        delayBetweenReadsSec = 5,
        timeoutSec = 60,
        accessToken = flowRootState.users.system_administrator.accessToken
    ): Promise<TelemetryResponse> {
        let lastValue: string | undefined;
        let consecutiveStableReads = 0;
        let latestResponse: TelemetryResponse;

        const isStable = async (): Promise<boolean> => {
            latestResponse = await this.getWorkloadTelemetryFilteredByClusterId(
                clusterId,
                telemetryType,
                accessToken
            );
            const currentValue = latestResponse.values[0]?.value;
            logger.debug(
                `Telemetry stabilization check: type=${telemetryType}, value=${currentValue}, ` +
                    `lastValue=${lastValue}, consecutiveStableReads=${consecutiveStableReads}/${stableReadsRequired}`
            );

            if (currentValue === lastValue) {
                consecutiveStableReads++;
            } else {
                consecutiveStableReads = 1;
                lastValue = currentValue;
            }

            return consecutiveStableReads >= stableReadsRequired;
        };

        await waitUtils.waitForCondition(
            isStable,
            `wait for telemetry ${telemetryType} to stabilize`,
            true,
            timeoutSec,
            delayBetweenReadsSec
        );

        return latestResponse;
    }
}
