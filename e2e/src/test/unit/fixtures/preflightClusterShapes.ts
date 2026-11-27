/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { CommandResult, CommandRunner } from '../../../modules/preflight/clusterProbe';

/**
 * Install shapes recorded from real clusters, with the detection each one is
 * expected to produce.
 *
 * Every prerequisite can be installed in more than one way, and each way puts
 * the version somewhere else. Guessing the possibilities up front does not
 * work; what does work is recording each shape the kit is shown so it can
 * never regress. When a cluster reports something unexpected, capture its
 * objects here and add a row.
 */

export interface ClusterObject {
    resource: string;
    namespace?: string;
    body: Record<string, any>;
}

export interface HelmRelease {
    name: string;
    chart: string;
    app_version?: string;
}

export interface ClusterShape {
    /** Where the shape was seen, so the fixture can be traced back. */
    name: string;
    serverVersion?: string;
    helmReleases: HelmRelease[];
    objects: ClusterObject[];
    /** Component id to the version the probes should read, or undefined. */
    expected: Record<string, string | undefined>;
}

const labelsOf = (object: ClusterObject): Record<string, string> => object.body?.metadata?.labels ?? {};

/** `key=value` matches that label; a bare `key` matches its presence. */
const matchesSelector = (object: ClusterObject, selector: string): boolean =>
    selector.split(',').every((term) => {
        const [key, value] = term.split('=');
        const labels = labelsOf(object);
        return value === undefined ? key in labels : labels[key] === value;
    });

const valueOfFlag = (args: string[], flag: string): string | undefined => {
    const index = args.indexOf(flag);
    return index === -1 ? undefined : args[index + 1];
};

const ok = (payload: unknown): CommandResult => ({ code: 0, stdout: JSON.stringify(payload), stderr: '' });

/**
 * Enough of `kubectl get` and `helm list` to drive `ClusterProbe` over a shape:
 * resource, namespace, object name, label selector and name field selector.
 */
export const runnerFor = (shape: ClusterShape): CommandRunner => {
    return (command, args) => {
        if (command === 'helm') {
            return ok(shape.helmReleases);
        }
        if (args[0] === 'version') {
            return shape.serverVersion
                ? ok({ serverVersion: { gitVersion: shape.serverVersion } })
                : { code: 1, stdout: '', stderr: 'the server could not be reached' };
        }

        const resource = args[1];
        const positionalName = args[2] && !args[2].startsWith('-') ? args[2] : undefined;
        const namespace = valueOfFlag(args, '-n');
        const selector = valueOfFlag(args, '-l');
        const fieldSelector = args
            .find((arg) => arg.startsWith('--field-selector=metadata.name='))
            ?.split('=')
            .pop();
        const wantedName = positionalName ?? fieldSelector;

        const items = shape.objects.filter(
            (object) =>
                object.resource === resource &&
                (namespace === undefined || object.namespace === namespace) &&
                (wantedName === undefined || object.body?.metadata?.name === wantedName) &&
                (selector === undefined || matchesSelector(object, selector))
        );

        // kubectl fails rather than returning an empty list for a named lookup.
        if (positionalName && items.length === 0) {
            return { code: 1, stdout: '', stderr: `Error from server (NotFound): ${resource} not found` };
        }
        return ok(positionalName ? items[0].body : { items: items.map((object) => object.body) });
    };
};

const deployment = (
    name: string,
    namespace: string,
    containers: { name: string; image: string }[],
    labels: Record<string, string> = {}
): ClusterObject => ({
    resource: 'deployment',
    namespace,
    body: {
        metadata: { name, namespace, labels },
        spec: { template: { spec: { containers } } }
    }
});

/**
 * NVIDIA Run:ai self-hosted on bare metal, built by the env-in-a-click
 * pipeline. Recorded from `sh-op-real-a15`, September 2026.
 *
 * Two things about it are easy to get wrong: the Prometheus instance comes
 * straight from a custom resource that pins no version, and the Training
 * Operator image is tagged with a build id rather than a release.
 */
export const runaiSelfHostedOnPrem: ClusterShape = {
    name: 'runai self-hosted, env-in-a-click on bare metal (sh-op-real-a15)',
    serverVersion: 'v1.35.5',
    helmReleases: [
        { name: 'gpu-operator', chart: 'gpu-operator-v25.10.1', app_version: 'v25.10.1' },
        { name: 'kube-prometheus-stack', chart: 'kube-prometheus-stack-68.3.0', app_version: 'v0.79.2' }
    ],
    objects: [
        {
            resource: 'prometheuses.monitoring.coreos.com',
            namespace: 'runai',
            body: { metadata: { name: 'runai', namespace: 'runai' }, spec: {} }
        },
        {
            resource: 'statefulset',
            namespace: 'runai',
            body: {
                metadata: {
                    name: 'prometheus-runai',
                    namespace: 'runai',
                    labels: {
                        'managed-by': 'prometheus-operator',
                        'operator.prometheus.io/mode': 'server',
                        'operator.prometheus.io/name': 'runai',
                        'operator.prometheus.io/shard': '0'
                    }
                },
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
            }
        },
        deployment(
            'training-operator',
            'kubeflow',
            [{ name: 'training-operator', image: 'docker.io/kubeflow/training-operator:v1-5170a36' }],
            { 'control-plane': 'kubeflow-training-operator' }
        ),
        deployment('mpi-operator', 'mpi-operator', [{ name: 'mpi-operator', image: 'mpioperator/mpi-operator:0.6.0' }])
    ],
    expected: {
        kubernetes: 'v1.35.5',
        'nvidia-gpu-operator': 'v25.10.1',
        'kube-prometheus-stack': '68.3.0',
        prometheus: 'v3.0.1',
        'kubeflow-training-operator': 'v1-5170a36',
        'mpi-operator': '0.6.0',
        'nvidia-network-operator': undefined,
        'nvidia-dra-driver': undefined,
        'knative-serving': undefined,
        'leader-worker-set': undefined
    }
};

/**
 * A cluster whose Prometheus was created by the Kube-Prometheus Stack chart
 * itself. The custom resource leaves `spec.version` unset and pins the release
 * through `spec.image`, and the StatefulSet carries the conventional labels.
 */
export const kubePrometheusStackManaged: ClusterShape = {
    name: 'prometheus installed by the kube-prometheus-stack chart',
    serverVersion: 'v1.34.9',
    helmReleases: [
        { name: 'kube-prometheus-stack', chart: 'kube-prometheus-stack-76.4.1', app_version: 'v0.87.0' }
    ],
    objects: [
        {
            resource: 'prometheuses.monitoring.coreos.com',
            namespace: 'monitoring',
            body: {
                metadata: { name: 'kube-prometheus-stack-prometheus', namespace: 'monitoring' },
                spec: { image: 'quay.io/prometheus/prometheus:v3.6.0' }
            }
        },
        {
            resource: 'statefulset',
            namespace: 'monitoring',
            body: {
                metadata: {
                    name: 'prometheus-kube-prometheus-stack-prometheus',
                    namespace: 'monitoring',
                    labels: { 'app.kubernetes.io/name': 'prometheus' }
                },
                spec: {
                    template: {
                        spec: {
                            containers: [{ name: 'prometheus', image: 'quay.io/prometheus/prometheus:v3.6.0' }]
                        }
                    }
                }
            }
        }
    ],
    expected: {
        kubernetes: 'v1.34.9',
        'kube-prometheus-stack': '76.4.1',
        prometheus: 'v3.6.0',
        'nvidia-gpu-operator': undefined
    }
};

export const clusterShapes: ClusterShape[] = [runaiSelfHostedOnPrem, kubePrometheusStackManaged];
