/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { KubernetesObject } from '@kubernetes/client-node';
import { expect } from 'chai';
import { flowRootState, versionUtils, WorkloadConditionType, workloadUtils } from '../../index';
import { Phase } from '../../../api/apis/workloadsApi/generated/model/phase';
import { kubeCli } from '../../../cli';

export function expectRunningWorkload(workload: KubernetesObject): void {
    expect(workload['status']['runningPods']).to.equal(1);
    expect(workload['status']['pendingPods']).to.oneOf([0, undefined]);
    expect(workload['status']['failedPods']).to.oneOf([0, undefined]);
    expect(workload['status']['succeededPods']).to.oneOf([0, undefined]);
    expect(workload['status']['startTime']).not.to.be.undefined;
    expect(workload['status']['completionTime']).to.be.undefined;
    expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Ready)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Failed)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Succeeded)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Pending)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Suspended)).to.be.true;
    if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.20')) {
        expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Initialized)).to.be.true;
    } else {
        expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Running)).to.be.true;
    }
    expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Ready)).to.be.true;
}

export function expectCompleteSuccessfully(workload: KubernetesObject, podsCount = 1): void {
    expect(workload['status']['pendingPods']).to.oneOf([0, undefined]);
    expect(workload['status']['failedPods']).to.oneOf([0, undefined]);
    expect(workload['status']['runningPods']).to.oneOf([0, undefined]);
    expect(workload['status']['succeededPods']).to.equal(podsCount);
    expect(workload['status']['completionTime']).not.to.be.undefined;
    expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Succeeded)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Failed)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Initialized)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Pending)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Suspended)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Ready)).to.be.true;
}

export function expectPendingWorkload(workload: KubernetesObject): void {
    expect(workload['status']['runningPods']).to.oneOf([0, undefined]);
    expect(workload['status']['pendingPods']).to.equal(1);
    expect(workload['status']['failedPods']).to.oneOf([0, undefined]);
    expect(workload['status']['succeededPods']).to.oneOf([0, undefined]);
    expect(workload['status']['startTime']).to.be.undefined;
    expect(workload['status']['completionTime']).to.be.undefined;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Succeeded)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Failed)).to.be.true;
    if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.20')) {
        expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Initialized)).to.be.true;
    } else {
        expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Running)).to.be.true;
    }
    expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Pending)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Suspended)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Ready)).to.be.true;
}

export function expectSuspendedWorkload(workload: KubernetesObject): void {
    expect(workload['status']['runningPods'], 'runningPods').to.oneOf([0, undefined]);
    expect(workload['status']['pendingPods'], 'pendingPods').to.oneOf([0, undefined]);
    expect(workload['status']['failedPods'], 'failedPods').to.oneOf([0, undefined]);
    expect(workload['status']['succeededPods'], 'succeededPods').to.oneOf([0, undefined]);
    expect(workload['status']['completionTime'], 'completionTime').to.be.undefined;
    expect(workload['status']['startTime'], 'startTime').to.be.undefined;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Failed)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Succeeded)).to.be.true;
    if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.20')) {
        expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Initialized)).to.be.true;
    } else {
        expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Running)).to.be.true;
    }
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Pending)).to.be.true;
    expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Suspended)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Ready)).to.be.true;
}

export function expectFailedWorkload(workload: KubernetesObject): void {
    if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.16')) {
        expect(workload['status']['phase'], 'phase').to.equal(Phase.Failed);
    }
    expect(workload['status']['runningPods'], 'runningPods').to.oneOf([0, undefined]);
    expect(workload['status']['pendingPods'], 'pendingPods').to.oneOf([0, undefined]);
    // check failed pods as at least 1 pod should fail
    expect(workload['status']['failedPods'], 'failedPods').to.greaterThan(0);
    expect(workload['status']['succeededPods'], 'succeededPods').to.oneOf([0, undefined]);
    expect(workload['status']['completionTime'], 'completionTime').not.to.be.undefined;
    expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Failed)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Succeeded)).to.be.true;
    if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.20')) {
        expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Initialized)).to.be.true;
    } else {
        expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Running)).to.be.true;
    }
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Pending)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Suspended)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Ready)).to.be.true;
}

export function expectRunningPodsForWorkload(workload: KubernetesObject, runningPodsCount: number): void {
    expect(workload['status']['runningPods'], 'runningPods').to.equal(runningPodsCount);
    expect(workload['status']['failedPods'], 'failedPods').to.oneOf([0, undefined]);
    expect(workload['status']['succeededPods'], 'succeededPods').to.oneOf([0, undefined]);
    expect(workload['status']['startTime'], 'startTime').not.to.be.undefined;
    expect(workload['status']['completionTime'], 'completionTime').to.be.undefined;
    expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Ready)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Failed)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Succeeded)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Pending)).to.be.true;
    expect(workloadUtils.isConditionFalseOrUnset(workload, WorkloadConditionType.Suspended)).to.be.true;
    if (versionUtils.isVersionSupported(flowRootState.cluster.version, '2.20')) {
        expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Initialized)).to.be.true;
    } else {
        expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Running)).to.be.true;
    }
    expect(workloadUtils.isConditionTrue(workload, WorkloadConditionType.Ready)).to.be.true;
}

export async function isPendingPodsZeroOrUndefined(namespace: string, jobName: string): Promise<boolean> {
    const workload = await kubeCli.getTrainWorkloadOfRunaiJob(jobName, namespace);
    if ('pendingPods' in workload['status']) {
        return workload['status']['pendingPods'] === 0;
    }
    return true;
}
