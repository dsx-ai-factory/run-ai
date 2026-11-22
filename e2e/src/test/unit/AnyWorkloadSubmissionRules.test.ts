/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import {
    AnyWorkloadSubmissionType,
    getVersionBasedSubmissionRoute,
    WorkloadSubmissionRoute
} from '../../modules/helpers/utils/AnyWorkloadSubmissionRules';

test.describe('workload submission version rules', () => {
    for (const workloadType of Object.values(AnyWorkloadSubmissionType)) {
        for (const version of ['2.23.9', '2.25.9', 'v2.25.9-staging.1', undefined]) {
            test(`${workloadType} uses native submission on ${version ?? 'an unknown version'}`, () => {
                expect(getVersionBasedSubmissionRoute(version, workloadType)).toBe(
                    WorkloadSubmissionRoute.Legacy
                );
            });
        }

        for (const version of [
            '3.0.0',
            'v3.0.0-rc.1',
            '3.0.40001788-7cad1d8',
            '3.0.0-build',
            '3.1.0',
            '0.0.0-build',
            'master',
            'v0.0.0-master',
            '2.26.0-master'
        ]) {
            test(`${workloadType} requires a runner on ${version} regardless of deployment flags`, () => {
                expect(getVersionBasedSubmissionRoute(version, workloadType)).toBe(
                    WorkloadSubmissionRoute.WorkloadRunner
                );
            });
        }
    }

    for (const workloadType of [
        AnyWorkloadSubmissionType.Workspace,
        AnyWorkloadSubmissionType.Training,
        AnyWorkloadSubmissionType.Distributed
    ]) {
        for (const version of ['2.26.0', 'v2.26.0-staging.1']) {
            test(`${workloadType} still consults deployment flags on ${version}`, () => {
                expect(getVersionBasedSubmissionRoute(version, workloadType)).toBeUndefined();
            });
        }
    }

    for (const workloadType of [
        AnyWorkloadSubmissionType.Inference,
        AnyWorkloadSubmissionType.DistributedInference
    ]) {
        test(`${workloadType} stays native on 2.26 even when deployment flags are enabled`, () => {
            expect(getVersionBasedSubmissionRoute('2.26.0', workloadType)).toBe(
                WorkloadSubmissionRoute.Legacy
            );
        });
    }
});
