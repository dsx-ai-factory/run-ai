/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import { ClusterProbe } from '../../modules/preflight/clusterProbe';
import { evaluatePreflight, renderProbeTrace } from '../../modules/preflight/preflight';
import { getSupportMatrix } from '../../modules/preflight/supportMatrix';
import { clusterShapes, runaiSelfHostedOnPrem, runnerFor } from './fixtures/preflightClusterShapes';

const matrix = getSupportMatrix('2.26');

test.describe('preflight against recorded cluster shapes', () => {
    for (const shape of clusterShapes) {
        for (const [componentId, version] of Object.entries(shape.expected)) {
            test(`should read ${componentId} as ${version ?? 'not detected'} on ${shape.name}`, () => {
                const component = matrix.components.find((candidate) => candidate.id === componentId);
                if (!component) {
                    throw new Error(`no component '${componentId}' in the support matrix`);
                }

                const result = new ClusterProbe(runnerFor(shape)).resolve(component);

                expect(result.detected).toBe(version);
            });
        }
    }

    test('should record every lookup it made, including the ones that found nothing', () => {
        const component = matrix.components.find((candidate) => candidate.id === 'leader-worker-set')!;

        const result = new ClusterProbe(runnerFor(runaiSelfHostedOnPrem)).resolve(component);

        expect(result.detected).toBe(undefined);
        expect(result.attempts?.length).toBeGreaterThan(1);
        expect(result.attempts?.every((attempt) => attempt.matched === 0)).toBe(true);
        expect(result.attempts?.map((attempt) => attempt.via)).toContain('helm release of chart lws');
    });

    test('should show the lookups behind a component in the rendered trace', () => {
        const probe = new ClusterProbe(runnerFor(runaiSelfHostedOnPrem));
        const probeResults = matrix.components
            .filter((component) => component.skipOnPlatform !== 'k8s')
            .map((component) => probe.resolve(component));

        const trace = renderProbeTrace(evaluatePreflight({ matrix, platform: 'k8s', probeResults }));

        expect(trace).toContain('read v3.0.1');
        expect(trace).toContain('no release of that chart in helm list -A');
    });
});
