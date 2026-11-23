/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import { findWorkloadRunnerByWorkloadName } from '../../modules/helpers/utils/WorkloadRunnerUtils';

test.describe('findWorkloadRunnerByWorkloadName', () => {
    test('finds a WorkloadRunner by each supported workload-name location', () => {
        const runners = [
            { metadata: { name: 'runner-for-name-match' } },
            { metadata: { name: 'runner-by-label', labels: { workloadName: 'label-match' } } },
            { metadata: { name: 'runner-by-spec' }, spec: { workload: { name: 'spec-match' } } }
        ];

        expect(findWorkloadRunnerByWorkloadName(runners, 'name-match')).toBe(runners[0]);
        expect(findWorkloadRunnerByWorkloadName(runners, 'label-match')).toBe(runners[1]);
        expect(findWorkloadRunnerByWorkloadName(runners, 'spec-match')).toBe(runners[2]);
        expect(findWorkloadRunnerByWorkloadName(runners, 'missing')).toBeUndefined();
    });
});
