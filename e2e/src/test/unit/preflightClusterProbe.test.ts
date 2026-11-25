/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import { ClusterProbe, CommandResult, CommandRunner } from '../../modules/preflight/clusterProbe';
import { getSupportMatrix } from '../../modules/preflight/supportMatrix';
import type { ComponentSpec } from '../../modules/preflight/supportMatrix';

const matrix = getSupportMatrix('2.26');

const component = (id: string): ComponentSpec => {
    const spec = matrix.components.find((candidate) => candidate.id === id);
    if (!spec) {
        throw new Error(`no component '${id}' in the support matrix`);
    }
    return spec;
};

const ok = (payload: unknown): CommandResult => ({ code: 0, stdout: JSON.stringify(payload), stderr: '' });
const notFound = (): CommandResult => ({ code: 1, stdout: '', stderr: 'Error from server (NotFound)' });
const emptyList = (): CommandResult => ok({ items: [] });

const gpuOperatorDeployment = (labels: Record<string, string>, image: string): unknown => ({
    metadata: { name: 'gpu-operator', namespace: 'gpu-operator', labels },
    spec: { template: { spec: { containers: [{ name: 'gpu-operator', image }] } } }
});

/** Records every invocation so tests can assert which lookups were needed. */
const recordingRunner = (handler: CommandRunner): { runner: CommandRunner; calls: string[][] } => {
    const calls: string[][] = [];
    const runner: CommandRunner = (commandName, args) => {
        calls.push([commandName, ...args]);
        return handler(commandName, args);
    };
    return { runner, calls };
};

test.describe('preflight cluster probe', () => {
    test('should read the Kubernetes server version', () => {
        const runner: CommandRunner = (commandName, args) =>
            commandName === 'kubectl' && args[0] === 'version'
                ? ok({ serverVersion: { gitVersion: 'v1.35.2' } })
                : notFound();

        const result = new ClusterProbe(runner).resolve(component('kubernetes'));

        expect(result.detected).toBe('v1.35.2');
        expect(result.detectedVia).toBe('kubectl version');
    });

    test('should prefer the Helm release app version for the GPU Operator', () => {
        const { runner, calls } = recordingRunner((commandName) =>
            commandName === 'helm'
                ? ok([{ name: 'gpu-operator', chart: 'gpu-operator-v25.10.0', app_version: 'v25.10.0' }])
                : notFound()
        );

        const result = new ClusterProbe(runner).resolve(component('nvidia-gpu-operator'));

        expect(result.detected).toBe('v25.10.0');
        expect(result.detectedVia).toContain('helm release');
        expect(calls.filter(([commandName]) => commandName === 'kubectl')).toHaveLength(0);
    });

    test('should fall back to the deployment version label when Helm metadata is unavailable', () => {
        const runner: CommandRunner = (commandName, args) => {
            if (commandName === 'helm') {
                return notFound();
            }
            if (args.includes('gpu-operator') && args.includes('-n')) {
                return ok(
                    gpuOperatorDeployment(
                        { 'app.kubernetes.io/version': 'v26.0.1' },
                        'nvcr.io/nvidia/gpu-operator:v26.0.1'
                    )
                );
            }
            return emptyList();
        };

        const result = new ClusterProbe(runner).resolve(component('nvidia-gpu-operator'));

        expect(result.detected).toBe('v26.0.1');
        expect(result.detectedVia).toContain('-n gpu-operator');
    });

    test('should fall back to the container image tag when no version label is set', () => {
        const runner: CommandRunner = (commandName, args) => {
            if (commandName === 'helm') {
                return notFound();
            }
            if (args.includes('nvidia-gpu-operator')) {
                return ok(gpuOperatorDeployment({}, 'nvcr.io:5000/nvidia/gpu-operator:v25.10.2'));
            }
            return notFound();
        };

        const result = new ClusterProbe(runner).resolve(component('nvidia-gpu-operator'));

        expect(result.detected).toBe('v25.10.2');
    });

    test('should report no version when the component is absent', () => {
        const runner: CommandRunner = () => notFound();

        const result = new ClusterProbe(runner).resolve(component('leader-worker-set'));

        expect(result.detected).toBe(undefined);
        expect(result.error).toBe(undefined);
    });

    test('should not report an unrelated deployment as a component that is absent', () => {
        const { runner, calls } = recordingRunner((commandName, args) => {
            if (commandName === 'helm') {
                return notFound();
            }
            if (args.includes('-A')) {
                const selector = args.find((arg) => arg.startsWith('--field-selector='));
                const wanted = selector?.split('=').pop();
                const items = [
                    gpuOperatorDeployment(
                        { 'app.kubernetes.io/version': 'v25.10.0' },
                        'nvcr.io/nvidia/gpu-operator:v25.10.0'
                    )
                ].filter(() => wanted === 'gpu-operator');
                return ok({ items });
            }
            return notFound();
        });

        const result = new ClusterProbe(runner).resolve(component('leader-worker-set'));

        expect(result.detected).toBe(undefined);
        expect(
            calls.some((call) => call.includes('--field-selector=metadata.name=lws-controller-manager'))
        ).toBe(true);
    });

    test('should read the Prometheus version from spec.image when spec.version is unset', () => {
        const runner: CommandRunner = (commandName, args) =>
            args.includes('prometheuses.monitoring.coreos.com')
                ? ok({
                      items: [
                          {
                              metadata: { name: 'kube-prometheus-stack-prometheus' },
                              spec: { image: 'quay.io/prometheus/prometheus:v3.5.0' }
                          }
                      ]
                  })
                : notFound();

        const result = new ClusterProbe(runner).resolve(component('prometheus'));

        expect(result.detected).toBe('v3.5.0');
    });

    test('should fall back to the Prometheus StatefulSet when the custom resource carries no version', () => {
        const runner: CommandRunner = (commandName, args) => {
            if (args.includes('prometheuses.monitoring.coreos.com')) {
                return ok({ items: [{ metadata: { name: 'prometheus' }, spec: {} }] });
            }
            if (args.includes('statefulset')) {
                return ok({
                    items: [
                        {
                            metadata: { name: 'prometheus-kube-prometheus-stack-prometheus' },
                            spec: {
                                template: {
                                    spec: {
                                        containers: [
                                            { name: 'config-reloader', image: 'quay.io/reloader:v0.87.0' },
                                            { name: 'prometheus', image: 'quay.io/prometheus:v3.5.1' }
                                        ]
                                    }
                                }
                            }
                        }
                    ]
                });
            }
            return notFound();
        };

        const result = new ClusterProbe(runner).resolve(component('prometheus'));

        expect(result.detected).toBe('v3.5.1');
    });

    test.describe('Prometheus created from a custom resource', () => {
        /** Shape prometheus-operator gives the StatefulSet it governs. */
        const operatorManaged = (labels: Record<string, string>): unknown => ({
            metadata: { name: 'prometheus-runai', namespace: 'runai', labels },
            spec: {
                template: {
                    spec: {
                        containers: [
                            { name: 'prometheus', image: 'quay.io/prometheus/prometheus:v3.0.1' },
                            {
                                name: 'config-reloader',
                                image: 'quay.io/prometheus-operator/prometheus-config-reloader:v0.79.2'
                            }
                        ]
                    }
                }
            }
        });

        /** The custom resource pins neither `spec.version` nor `spec.image`. */
        const runnerServing = (selector: string, item: unknown): CommandRunner => {
            return (commandName, args) => {
                if (args.includes('prometheuses.monitoring.coreos.com')) {
                    return ok({ items: [{ metadata: { name: 'runai' }, spec: {} }] });
                }
                if (args.includes('statefulset') && args[args.indexOf('-l') + 1] === selector) {
                    return ok({ items: [item] });
                }
                return emptyList();
            };
        };

        test('should read the version from the server-mode StatefulSet', () => {
            const runner = runnerServing(
                'operator.prometheus.io/mode=server',
                operatorManaged({
                    'managed-by': 'prometheus-operator',
                    'operator.prometheus.io/mode': 'server',
                    'operator.prometheus.io/name': 'runai'
                })
            );

            expect(new ClusterProbe(runner).resolve(component('prometheus')).detected).toBe('v3.0.1');
        });

        test('should still find it when the operator sets no mode label', () => {
            const runner = runnerServing(
                'operator.prometheus.io/name',
                operatorManaged({
                    'managed-by': 'prometheus-operator',
                    'operator.prometheus.io/name': 'runai'
                })
            );

            expect(new ClusterProbe(runner).resolve(component('prometheus')).detected).toBe('v3.0.1');
        });
    });

    test('should use the chart version rather than the app version for the Kube-Prometheus Stack', () => {
        const runner: CommandRunner = (commandName) =>
            commandName === 'helm'
                ? ok([
                      {
                          name: 'kube-prometheus-stack',
                          chart: 'kube-prometheus-stack-76.4.1',
                          app_version: 'v0.87.0'
                      }
                  ])
                : notFound();

        const result = new ClusterProbe(runner).resolve(component('kube-prometheus-stack'));

        expect(result.detected).toBe('76.4.1');
    });

    test.describe('platform detection', () => {
        test('should report ocp when a ClusterVersion object exists', () => {
            const runner: CommandRunner = () =>
                ok({ items: [{ status: { desired: { version: '4.19.3' } } }] });

            expect(new ClusterProbe(runner).detectPlatform()).toBe('ocp');
        });

        test('should report k8s when the OpenShift CRD is not served', () => {
            const runner: CommandRunner = () => ({
                code: 1,
                stdout: '',
                stderr: 'error: the server doesn\'t have a resource type "clusterversions"'
            });

            expect(new ClusterProbe(runner).detectPlatform()).toBe('k8s');
        });
    });
});
