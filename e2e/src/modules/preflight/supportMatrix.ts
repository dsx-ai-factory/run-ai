/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * What NVIDIA Run:ai publishes as supported for a cluster release, and how to
 * find each component's deployed version on a cluster.
 *
 * This is the only file that has to change when a new cluster release ships.
 */

// ---------------------------------------------------------------------------
// Shapes
// ---------------------------------------------------------------------------

/** Inclusive version range. An omitted `max` means "this version and above". */
export interface VersionRange {
    min: string;
    max?: string;
}

/**
 * `required` - NVIDIA Run:ai documents the component as mandatory for every cluster.
 * `conditional` - only mandatory for a subset of hardware or workload types, so a
 * missing component is reported but does not fail the gate. A component that is
 * present must still be within the supported range.
 */
export type Requirement = 'required' | 'conditional';

/** Reads a version string out of a single Kubernetes object. */
export type VersionExtractor = (object: Record<string, any>) => string | undefined;

export interface K8sServerVersionProbe {
    kind: 'k8sServerVersion';
}

export interface KubectlProbe {
    kind: 'kubectl';
    /** Resource to query, for example `deployment` or `clusterpolicies.nvidia.com`. */
    resource: string;
    /** Object name. When omitted the first object matching `labelSelector` is used. */
    name?: string;
    /** Namespaces to try in order. Omit for cluster-scoped or all-namespace lookups. */
    namespaces?: string[];
    /** Used for the all-namespace fallback when the named lookup finds nothing. */
    labelSelector?: string;
    extract: VersionExtractor;
}

export interface HelmProbe {
    kind: 'helm';
    /** Matched against the chart name, for example `gpu-operator-v25.10.0`. */
    chart: string;
    /** Prefer the chart's `app_version` over its chart version. Defaults to true. */
    preferAppVersion?: boolean;
}

export type Probe = K8sServerVersionProbe | KubectlProbe | HelmProbe;

/** Distribution family, which decides whether some components apply at all. */
export type Platform = 'k8s' | 'ocp';

export interface ComponentSpec {
    /** Stable identifier, used as the key in reports. */
    id: string;
    displayName: string;
    requirement: Requirement;
    supported: VersionRange;
    /** Explains when a `conditional` component becomes mandatory. */
    appliesWhen?: string;
    /** Platform on which the component does not apply and the check is skipped. */
    skipOnPlatform?: Platform;
    /** Tried in order; the first probe that yields a version wins. */
    probes: Probe[];
}

export interface SupportMatrix {
    /** NVIDIA Run:ai cluster version the matrix describes, for example `2.26`. */
    runaiVersion: string;
    /** Documentation page the values were taken from. */
    source: string;
    components: ComponentSpec[];
}

/**
 * One lookup a probe performed. Recorded for every lookup, so a component that
 * was not found can be told apart from one that was found without a readable
 * version, without re-running anything against the cluster.
 */
export interface ProbeAttempt {
    /** The lookup, worded as in `ProbeResult.detectedVia`. */
    via: string;
    /** Objects the lookup returned. */
    matched: number;
    /** Version read from one of those objects, when one could be read. */
    extracted?: string;
    /** Whatever the command wrote to stderr when it returned nothing. */
    note?: string;
}

export interface ProbeResult {
    componentId: string;
    /** Raw version string as found on the cluster, before normalisation. */
    detected?: string;
    /** How it was found, for operator-facing output. */
    detectedVia?: string;
    /** Set when every probe failed to run rather than simply finding nothing. */
    error?: string;
    /** Every lookup performed, in the order they were tried. */
    attempts?: ProbeAttempt[];
}

export type CheckStatus = 'pass' | 'fail' | 'warn' | 'skip';

export interface CheckResult {
    componentId: string;
    displayName: string;
    status: CheckStatus;
    /** What the check demanded, rendered for humans. */
    expected: string;
    /** What the cluster actually has, rendered for humans. */
    actual: string;
    detail: string;
    /** The lookups behind `actual`, kept in the report for diagnosis. */
    attempts?: ProbeAttempt[];
}

export interface PreflightReport {
    runaiVersion: string;
    source: string;
    platform: Platform;
    generatedAt: string;
    checks: CheckResult[];
    summary: Record<CheckStatus, number>;
    passed: boolean;
}

// ---------------------------------------------------------------------------
// Version extractors, used to declare the probes below
// ---------------------------------------------------------------------------

/** Reads the conventional `app.kubernetes.io/version` label. */
export const versionLabel: VersionExtractor = (object) =>
    object?.metadata?.labels?.['app.kubernetes.io/version'];

/** Follows a dotted path, for example `status.version`. */
export const fieldPath =
    (path: string): VersionExtractor =>
    (object) =>
        path.split('.').reduce<any>((current, key) => current?.[key], object);

/** Tag of an image reference. Digest-only references yield nothing. */
const tagOf = (image: unknown): string | undefined => {
    if (typeof image !== 'string') {
        return undefined;
    }
    const withoutDigest = image.split('@')[0];
    const lastColon = withoutDigest.lastIndexOf(':');
    // A colon before the last slash is a registry port, not a tag.
    if (lastColon === -1 || lastColon < withoutDigest.lastIndexOf('/')) {
        return undefined;
    }
    return withoutDigest.slice(lastColon + 1);
};

/**
 * Reads the image tag of a pod template container. Without `containerName` the
 * first container is used.
 */
export const imageTag =
    (containerName?: string): VersionExtractor =>
    (object) => {
        const containers: any[] = object?.spec?.template?.spec?.containers ?? [];
        const container = containerName
            ? containers.find((candidate) => candidate?.name === containerName)
            : containers[0];
        return tagOf(container?.image);
    };

/** Reads the image tag out of a field holding a reference, such as `spec.image`. */
export const imageFieldTag =
    (path: string): VersionExtractor =>
    (object) =>
        tagOf(fieldPath(path)(object));

/** Returns the first extractor result that is a non-empty string. */
export const firstOf =
    (...extractors: VersionExtractor[]): VersionExtractor =>
    (object) => {
        for (const extract of extractors) {
            const value = extract(object);
            if (typeof value === 'string' && value.trim() !== '') {
                return value.trim();
            }
        }
        return undefined;
    };

// ---------------------------------------------------------------------------
// The published matrix
// ---------------------------------------------------------------------------

/**
 * The versions NVIDIA Run:ai publicly commits to for a cluster release.
 *
 * Values are transcribed from the NVIDIA Run:ai documentation, which is the
 * contract the kit is asked to enforce. Keep this file the single place the
 * numbers live, and update it together with the docs on every cluster release.
 *
 * Sources:
 *   https://run-ai-docs.nvidia.com/self-hosted/getting-started/installation/support-matrix
 *   https://run-ai-docs.nvidia.com/self-hosted/getting-started/installation/install-using-helm/system-requirements
 */
export const supportMatrix2_26: SupportMatrix = {
    runaiVersion: '2.26',
    source: 'https://run-ai-docs.nvidia.com/self-hosted/getting-started/installation/support-matrix',
    components: [
        {
            id: 'kubernetes',
            displayName: 'Kubernetes',
            requirement: 'required',
            supported: { min: '1.34', max: '1.36' },
            skipOnPlatform: 'ocp',
            probes: [{ kind: 'k8sServerVersion' }]
        },
        {
            id: 'openshift',
            displayName: 'Red Hat OpenShift',
            requirement: 'required',
            supported: { min: '4.18', max: '4.22' },
            skipOnPlatform: 'k8s',
            probes: [
                {
                    kind: 'kubectl',
                    resource: 'clusterversions.config.openshift.io',
                    name: 'version',
                    extract: fieldPath('status.desired.version')
                }
            ]
        },
        {
            id: 'nvidia-gpu-operator',
            displayName: 'NVIDIA GPU Operator',
            requirement: 'required',
            supported: { min: '25.10', max: '26.3' },
            probes: [
                { kind: 'helm', chart: 'gpu-operator' },
                {
                    kind: 'kubectl',
                    resource: 'deployment',
                    name: 'gpu-operator',
                    namespaces: ['gpu-operator', 'nvidia-gpu-operator'],
                    labelSelector: 'app.kubernetes.io/component=gpu-operator',
                    extract: firstOf(versionLabel, imageTag('gpu-operator'), imageTag())
                },
                {
                    kind: 'kubectl',
                    resource: 'clusterpolicies.nvidia.com',
                    extract: versionLabel
                }
            ]
        },
        {
            id: 'nvidia-network-operator',
            displayName: 'NVIDIA Network Operator',
            requirement: 'conditional',
            appliesWhen: 'the cluster has RDMA or Multi-Node NVLink capable nodes (for example B200, GB200)',
            supported: { min: '25.10', max: '26.7' },
            probes: [
                { kind: 'helm', chart: 'network-operator' },
                {
                    kind: 'kubectl',
                    resource: 'deployment',
                    name: 'network-operator',
                    namespaces: ['network-operator', 'nvidia-network-operator'],
                    extract: firstOf(versionLabel, imageTag())
                }
            ]
        },
        {
            id: 'nvidia-dra-driver',
            displayName: 'NVIDIA DRA driver',
            requirement: 'conditional',
            appliesWhen: 'the cluster has Multi-Node NVLink nodes (for example GB200)',
            supported: { min: '25.8', max: '25.12' },
            probes: [
                { kind: 'helm', chart: 'nvidia-dra-driver-gpu' },
                {
                    kind: 'kubectl',
                    resource: 'deployment',
                    namespaces: ['nvidia-dra-driver-gpu', 'nvidia-dra-driver'],
                    labelSelector: 'app.kubernetes.io/name=nvidia-dra-driver-gpu',
                    extract: firstOf(versionLabel, imageTag())
                }
            ]
        },
        {
            id: 'kube-prometheus-stack',
            displayName: 'Kube-Prometheus Stack',
            requirement: 'required',
            supported: { min: '76.0' },
            skipOnPlatform: 'ocp',
            probes: [{ kind: 'helm', chart: 'kube-prometheus-stack', preferAppVersion: false }]
        },
        {
            id: 'prometheus',
            displayName: 'Prometheus',
            requirement: 'required',
            supported: { min: '3.5' },
            probes: [
                {
                    kind: 'kubectl',
                    resource: 'prometheuses.monitoring.coreos.com',
                    // Kube-Prometheus Stack leaves spec.version unset by default
                    // and pins the release through spec.image instead.
                    extract: firstOf(fieldPath('spec.version'), imageFieldTag('spec.image'), versionLabel)
                },
                {
                    kind: 'kubectl',
                    resource: 'statefulset',
                    labelSelector: 'app.kubernetes.io/name=prometheus',
                    extract: firstOf(versionLabel, imageTag('prometheus'))
                },
                {
                    kind: 'kubectl',
                    resource: 'statefulset',
                    // An instance created straight from a Prometheus custom
                    // resource carries only prometheus-operator's own labels,
                    // so the container image is the single place the version
                    // appears. `mode` separates a server from a PrometheusAgent,
                    // whose StatefulSet is labelled and named identically.
                    labelSelector: 'operator.prometheus.io/mode=server',
                    extract: imageTag('prometheus')
                },
                {
                    kind: 'kubectl',
                    resource: 'statefulset',
                    // Operators older than the `mode` label still identify the
                    // owning custom resource by name.
                    labelSelector: 'operator.prometheus.io/name',
                    extract: imageTag('prometheus')
                }
            ]
        },
        {
            id: 'kubeflow-training-operator',
            displayName: 'Kubeflow Training Operator',
            requirement: 'conditional',
            appliesWhen: 'running TensorFlow, PyTorch, XGBoost or JAX distributed training workloads',
            supported: { min: '1.9.2', max: '1.9.2' },
            probes: [
                { kind: 'helm', chart: 'training-operator' },
                {
                    kind: 'kubectl',
                    resource: 'deployment',
                    name: 'training-operator',
                    namespaces: ['kubeflow'],
                    labelSelector: 'control-plane=kubeflow-training-operator',
                    extract: firstOf(versionLabel, imageTag())
                }
            ]
        },
        {
            id: 'mpi-operator',
            displayName: 'MPI Operator',
            requirement: 'conditional',
            appliesWhen: 'running MPI v2 distributed workloads',
            supported: { min: '0.6.0' },
            probes: [
                { kind: 'helm', chart: 'mpi-operator' },
                {
                    kind: 'kubectl',
                    resource: 'deployment',
                    name: 'mpi-operator',
                    namespaces: ['mpi-operator'],
                    extract: firstOf(versionLabel, imageTag())
                }
            ]
        },
        {
            id: 'knative-serving',
            displayName: 'Knative Serving',
            requirement: 'conditional',
            appliesWhen: 'serving inference workloads',
            supported: { min: '1.19', max: '1.21' },
            probes: [
                {
                    kind: 'kubectl',
                    resource: 'knativeservings.operator.knative.dev',
                    extract: firstOf(fieldPath('status.version'), fieldPath('spec.version'), versionLabel)
                },
                {
                    kind: 'kubectl',
                    resource: 'deployment',
                    name: 'controller',
                    namespaces: ['knative-serving'],
                    extract: firstOf(versionLabel, imageTag())
                }
            ]
        },
        {
            id: 'leader-worker-set',
            displayName: 'Leader Worker Set (LWS)',
            requirement: 'conditional',
            appliesWhen: 'running distributed (multi-node) inference workloads',
            supported: { min: '0.7.0' },
            probes: [
                { kind: 'helm', chart: 'lws' },
                {
                    kind: 'kubectl',
                    resource: 'deployment',
                    name: 'lws-controller-manager',
                    namespaces: ['lws-system', 'openshift-lws-operator'],
                    extract: firstOf(versionLabel, imageTag('manager'), imageTag())
                }
            ]
        }
    ]
};

/** Keyed by NVIDIA Run:ai cluster minor version. */
export const supportMatrices: Record<string, SupportMatrix> = {
    '2.26': supportMatrix2_26
};

export const defaultRunaiVersion = '2.26';

export const getSupportMatrix = (runaiVersion: string = defaultRunaiVersion): SupportMatrix => {
    const matrix = supportMatrices[runaiVersion];
    if (!matrix) {
        const known = Object.keys(supportMatrices).join(', ');
        throw new Error(
            `No support matrix defined for NVIDIA Run:ai version '${runaiVersion}'. Known versions: ${known}`
        );
    }
    return matrix;
};
