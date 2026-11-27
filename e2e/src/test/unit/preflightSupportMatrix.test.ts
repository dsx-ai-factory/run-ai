/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import {
    checkRange,
    compareVersions,
    evaluatePreflight,
    formatRange,
    parseVersion,
    satisfiesRange
} from '../../modules/preflight/preflight';
import { getSupportMatrix } from '../../modules/preflight/supportMatrix';
import type { CheckResult, ProbeResult } from '../../modules/preflight/supportMatrix';

test.describe('preflight version handling', () => {
    test.describe('parseVersion', () => {
        test('should strip prefixes and build-metadata suffixes', () => {
            expect(parseVersion('v25.10.0')).toEqual([25, 10, 0]);
            expect(parseVersion('1.34.2-eks-1234567')).toEqual([1, 34, 2]);
            expect(parseVersion('  26.3  ')).toEqual([26, 3]);
        });

        test('should return undefined for values without a version', () => {
            expect(parseVersion(undefined)).toBe(undefined);
            expect(parseVersion('latest')).toBe(undefined);
        });
    });

    test.describe('compareVersions', () => {
        test('should compare only as deep as the shorter version', () => {
            expect(compareVersions([26, 3], [26, 3, 5])).toBe(0);
            expect(compareVersions([25, 9], [25, 10])).toBeLessThan(0);
        });
    });

    test.describe('satisfiesRange', () => {
        const gpuOperator = { min: '25.10', max: '26.3' };

        test('should accept versions inside the documented range', () => {
            expect(satisfiesRange('v25.10.0', gpuOperator)).toBe(true);
            expect(satisfiesRange('26.0.1', gpuOperator)).toBe(true);
        });

        test('should accept patch releases of the upper bound', () => {
            expect(satisfiesRange('26.3.5', gpuOperator)).toBe(true);
        });

        test('should reject versions outside the documented range', () => {
            expect(satisfiesRange('25.3.0', gpuOperator)).toBe(false);
            expect(satisfiesRange('26.4.0', gpuOperator)).toBe(false);
        });

        test('should treat a missing upper bound as open ended', () => {
            expect(satisfiesRange('0.7.0', { min: '0.7.0' })).toBe(true);
            expect(satisfiesRange('1.2.0', { min: '0.7.0' })).toBe(true);
            expect(satisfiesRange('0.6.9', { min: '0.7.0' })).toBe(false);
        });

        test('should reject an unreadable version', () => {
            expect(satisfiesRange(undefined, gpuOperator)).toBe(false);
            expect(satisfiesRange('latest', gpuOperator)).toBe(false);
        });

        test('should reject a version too coarse to place inside the range', () => {
            expect(satisfiesRange('v1-5170a36', { min: '1.9.2', max: '1.9.2' })).toBe(false);
        });
    });

    test.describe('checkRange', () => {
        test('should report a build tag as indeterminate rather than a match', () => {
            expect(checkRange('v1-5170a36', { min: '1.9.2', max: '1.9.2' })).toBe('indeterminate');
            expect(checkRange('latest', { min: '3.5' })).toBe('indeterminate');
        });

        test('should place a short version when the rest of the bound is zeroes', () => {
            expect(checkRange('76', { min: '76.0' })).toBe('inside');
        });

        test('should still decide a short version that differs on the shared prefix', () => {
            expect(checkRange('2', { min: '1.9.2', max: '1.9.2' })).toBe('outside');
        });
    });

    test.describe('formatRange', () => {
        test('should render bounded and open ended ranges', () => {
            expect(formatRange({ min: '25.10', max: '26.3' })).toBe('25.10 - 26.3');
            expect(formatRange({ min: '0.7.0' })).toBe('0.7.0 or above');
        });
    });
});

const matrix = getSupportMatrix('2.26');

const probe = (componentId: string, detected?: string, detectedVia = 'test'): ProbeResult => ({
    componentId,
    detected,
    detectedVia
});

/** Every required component in range, so a test only has to state its deviation. */
const healthyCluster = (): ProbeResult[] => [
    probe('kubernetes', 'v1.35.2'),
    probe('nvidia-gpu-operator', 'v25.10.0'),
    probe('kube-prometheus-stack', '76.4.1'),
    probe('prometheus', '3.5.0'),
    probe('knative-serving', '1.20.0'),
    probe('kubeflow-training-operator', 'v1.9.2'),
    probe('leader-worker-set', 'v0.7.0')
];

const replaceProbe = (probes: ProbeResult[], componentId: string, detected?: string): ProbeResult[] => [
    ...probes.filter((result) => result.componentId !== componentId),
    probe(componentId, detected)
];

const checkFor = (checks: CheckResult[], componentId: string): CheckResult => {
    const check = checks.find((candidate) => candidate.componentId === componentId);
    if (!check) {
        throw new Error(`no check produced for ${componentId}`);
    }
    return check;
};

test.describe('preflight evaluation against the NVIDIA Run:ai 2.26 support matrix', () => {
    test('should pass a cluster whose prerequisites are all inside the documented ranges', () => {
        const report = evaluatePreflight({ matrix, platform: 'k8s', probeResults: healthyCluster() });

        expect(report.passed).toBe(true);
        expect(report.summary.fail).toBe(0);
        expect(checkFor(report.checks, 'nvidia-gpu-operator').status).toBe('pass');
    });

    test('should fail and report expected versus actual when the GPU Operator is too old', () => {
        const probeResults = replaceProbe(healthyCluster(), 'nvidia-gpu-operator', 'v25.3.0');

        const report = evaluatePreflight({ matrix, platform: 'k8s', probeResults });
        const check = checkFor(report.checks, 'nvidia-gpu-operator');

        expect(report.passed).toBe(false);
        expect(check.status).toBe('fail');
        expect(check.expected).toBe('supported 25.10 - 26.3');
        expect(check.actual).toContain('v25.3.0');
        expect(check.detail).toContain('outside the supported range 25.10 - 26.3');
    });

    test('should fail when a required component is not deployed at all', () => {
        const probeResults = healthyCluster().filter(
            (result) => result.componentId !== 'nvidia-gpu-operator'
        );

        const report = evaluatePreflight({ matrix, platform: 'k8s', probeResults });
        const check = checkFor(report.checks, 'nvidia-gpu-operator');

        expect(check.status).toBe('fail');
        expect(check.actual).toBe('not detected');
        expect(report.passed).toBe(false);
    });

    test('should not call a component missing when the probe matched it but read no version', () => {
        const probeResults = [
            ...healthyCluster().filter((result) => result.componentId !== 'nvidia-gpu-operator'),
            {
                componentId: 'nvidia-gpu-operator',
                attempts: [
                    { via: 'helm release of chart gpu-operator', matched: 0 },
                    { via: 'deployment/gpu-operator -n gpu-operator', matched: 1 }
                ]
            }
        ];

        const report = evaluatePreflight({ matrix, platform: 'k8s', probeResults });
        const check = checkFor(report.checks, 'nvidia-gpu-operator');

        expect(check.status).toBe('fail');
        expect(check.actual).toBe('deployed, no version readable');
        expect(check.detail).toContain('deployment/gpu-operator -n gpu-operator matched 1 object(s)');
        expect(check.detail).not.toContain('was not found');
    });

    test('should only warn when a conditional component is not deployed', () => {
        const probeResults = healthyCluster().filter((result) => result.componentId !== 'leader-worker-set');

        const report = evaluatePreflight({ matrix, platform: 'k8s', probeResults });
        const check = checkFor(report.checks, 'leader-worker-set');

        expect(check.status).toBe('warn');
        expect(check.detail).toContain('distributed (multi-node) inference');
        expect(report.passed).toBe(true);
    });

    test('should only warn when a conditional component is deployed at an unsupported version', () => {
        const probeResults = replaceProbe(healthyCluster(), 'knative-serving', '1.18.0');

        const report = evaluatePreflight({ matrix, platform: 'k8s', probeResults });
        const check = checkFor(report.checks, 'knative-serving');

        expect(check.status).toBe('warn');
        expect(check.detail).toContain('outside the supported range 1.19 - 1.21');
        expect(check.detail).toContain('serving inference workloads');
        expect(report.passed).toBe(true);
    });

    test('should never fail the gate on a conditional component, whatever its state', () => {
        const states = ['1.18.0', 'v1-5170a36', undefined];

        for (const detected of states) {
            const probeResults = replaceProbe(healthyCluster(), 'knative-serving', detected);
            const report = evaluatePreflight({ matrix, platform: 'k8s', probeResults });

            expect(checkFor(report.checks, 'knative-serving').status).toBe('warn');
            expect(report.summary.fail).toBe(0);
            expect(report.passed).toBe(true);
        }
    });

    test('should warn rather than pass a conditional component whose version cannot be read', () => {
        const probeResults = replaceProbe(healthyCluster(), 'kubeflow-training-operator', 'v1-5170a36');

        const report = evaluatePreflight({ matrix, platform: 'k8s', probeResults });
        const check = checkFor(report.checks, 'kubeflow-training-operator');

        expect(check.status).toBe('warn');
        expect(check.actual).toContain('v1-5170a36');
        expect(check.detail).toContain('is deployed');
        expect(report.passed).toBe(true);
    });

    test('should fail a required component whose version cannot be read', () => {
        const probeResults = replaceProbe(healthyCluster(), 'nvidia-gpu-operator', 'v25-abc1234');

        const report = evaluatePreflight({ matrix, platform: 'k8s', probeResults });
        const check = checkFor(report.checks, 'nvidia-gpu-operator');

        expect(check.status).toBe('fail');
        expect(check.detail).toContain('does not give a version');
        expect(report.passed).toBe(false);
    });

    test('should skip Kubernetes and Kube-Prometheus Stack checks on OpenShift', () => {
        const probeResults = [
            probe('openshift', '4.19.3'),
            probe('nvidia-gpu-operator', 'v25.10.0'),
            probe('prometheus', '3.5.0')
        ];

        const report = evaluatePreflight({ matrix, platform: 'ocp', probeResults });

        expect(checkFor(report.checks, 'kubernetes').status).toBe('skip');
        expect(checkFor(report.checks, 'kube-prometheus-stack').status).toBe('skip');
        expect(checkFor(report.checks, 'openshift').status).toBe('pass');
    });

    test('should skip the OpenShift check on upstream Kubernetes', () => {
        const report = evaluatePreflight({ matrix, platform: 'k8s', probeResults: healthyCluster() });

        expect(checkFor(report.checks, 'openshift').status).toBe('skip');
    });
});
