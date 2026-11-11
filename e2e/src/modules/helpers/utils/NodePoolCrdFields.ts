/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { get } from 'lodash-es';
import { kubeCli } from '../../cli';
import { flowRootState } from '../config/config';
import { logger } from '../logger/logger';
import {
    crds,
    CustomResourceDefinition,
    KAI_NODE_POOL_GROUP,
    LEGACY_NODE_POOL_GROUP
} from '../models/customResourceDefinition';

/** The node pool CRD descriptor as served by `group`, built lazily to stay clear of import cycles. */
const nodePoolCrdOfGroup = (group: string): CustomResourceDefinition => ({
    ...crds.nodePoolCrd,
    group
});

let isNodePoolCrdDetected = false;

let nodePoolCrdDetection: Promise<CustomResourceDefinition | undefined> | undefined;

async function hasNodePools(crd: CustomResourceDefinition): Promise<boolean> {
    try {
        return (await kubeCli.getCrds(crd)).length > 0;
    } catch (error) {
        logger.debug(`Failed listing node pools of ${crd.plural}.${crd.group}: ${error}`);
        return false;
    }
}

async function detectNodePoolCrd(): Promise<CustomResourceDefinition | undefined> {
    const kaiNodePoolCrd = nodePoolCrdOfGroup(KAI_NODE_POOL_GROUP);
    const legacyNodePoolCrd = nodePoolCrdOfGroup(LEGACY_NODE_POOL_GROUP);
    const [isKaiInstalled, isLegacyInstalled] = await Promise.all([
        kubeCli.isCrdInstalled(kaiNodePoolCrd),
        kubeCli.isCrdInstalled(legacyNodePoolCrd)
    ]);

    if (isKaiInstalled && isLegacyInstalled) {
        // Both CRDs are installed during the migration window - the KAI one ships with the scheduler
        // before the cutover, and the legacy one is only uninstalled after it. The CRD that actually
        // holds node pools is the one the cluster reconciles into, so prefer that one; a cluster that
        // has node pools in neither (or in both) is treated as migrated.
        if ((await hasNodePools(legacyNodePoolCrd)) && !(await hasNodePools(kaiNodePoolCrd))) {
            return legacyNodePoolCrd;
        }
        return kaiNodePoolCrd;
    }
    if (isKaiInstalled) {
        return kaiNodePoolCrd;
    }
    if (isLegacyInstalled) {
        return legacyNodePoolCrd;
    }
    return undefined;
}

/** Applies a resolved descriptor to the shared `crds.nodePoolCrd` holder. */
function applyNodePoolCrd(group: string, version: string): CustomResourceDefinition {
    crds.nodePoolCrd.group = group;
    crds.nodePoolCrd.version = version;
    isNodePoolCrdDetected = true;
    return crds.nodePoolCrd;
}

/**
 * The NodePool CRD this cluster serves.
 *
 * Resolved once, in global setup, and recorded on `flowRootState` before it is written to
 * globalState.json - so every worker reads the answer instead of probing the cluster again. Which
 * CRD is installed cannot change during a run; only an upgrade changes it, and the upgrade suites
 * call `resetNodePoolCrdDetection` at that point.
 *
 * The result is written back into `crds.nodePoolCrd` rather than returned as a new object, so that
 * any other holder of that descriptor sees the resolved group too.
 */
export async function resolveNodePoolCrd(): Promise<CustomResourceDefinition> {
    if (isNodePoolCrdDetected) {
        return crds.nodePoolCrd;
    }

    // Recorded by global setup: workers take it from shared state and never probe.
    const shared = flowRootState.nodePoolCrd;
    if (shared?.group && shared?.version) {
        return applyNodePoolCrd(shared.group, shared.version);
    }

    nodePoolCrdDetection ??= detectNodePoolCrd();
    let detected: CustomResourceDefinition | undefined;
    try {
        detected = await nodePoolCrdDetection;
    } finally {
        nodePoolCrdDetection = undefined;
    }

    if (!detected) {
        logger.warn(
            `Could not detect an installed NodePool CRD, falling back to ${crds.nodePoolCrd.plural}.${crds.nodePoolCrd.group}`
        );
        return crds.nodePoolCrd;
    }

    // Global setup runs this before globalState.json is written, so recording it here is what
    // makes it available to the workers.
    flowRootState.nodePoolCrd = { group: detected.group, version: detected.version };
    logger.debug(`Detected NodePool CRD ${detected.plural}.${detected.group}`);
    return applyNodePoolCrd(detected.group, detected.version);
}

/**
 * Drops the resolved CRD, so the next access probes the cluster again. Needed by the upgrade
 * suites, where the cluster is replaced underneath a running worker.
 */
export function resetNodePoolCrdDetection(): void {
    isNodePoolCrdDetected = false;
    nodePoolCrdDetection = undefined;
    flowRootState.nodePoolCrd = undefined;
}

/** Whether the resolved NodePool CRD is the KAI one. Meaningful only after `resolveNodePoolCrd`. */
export function usesKaiNodePoolCrd(): boolean {
    return crds.nodePoolCrd.group === KAI_NODE_POOL_GROUP;
}

/**
 * Field accessors for the NodePool CRD across the KAI migration.
 */

/** Annotation keys on the KAI NodePool CRD, replacing spec / status fields of the legacy one. */
export const nodePoolAnnotations = {
    gpuNetworkAccelerationDetected: 'kai/gpu-network-acceleration-detected',
    gpuNetworkAccelerationLabelKey: 'kai/gpu-network-acceleration-label-key'
} as const;

/** The label key the KAI node pool controller detects on when the annotation is not set. */
export const DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY = 'nvidia.com/gpu.clique';

/**
 * A fetched node pool CRD. Deliberately loose: call sites hold these as plain objects, as
 * `KubernetesObject`, or as the generated REST models, none of which share an index signature.
 */
type Crd = object;

const annotation = (crd: Crd, key: string): string | undefined => {
    const metadata = (crd as Record<string, unknown>)?.['metadata'] as Record<string, unknown> | undefined;
    const annotations = metadata?.['annotations'] as Record<string, string> | undefined;
    return annotations?.[key];
};

/**
 * Whether a fetched object is a KAI node pool. The object's own `apiVersion` is authoritative and
 * survives a mid-run migration; the detected CRD is only used for objects that were fetched without
 * one (list items on some api-server versions).
 */
const isKaiNodePool = (crd: Crd): boolean => {
    const apiVersion = (crd as Record<string, unknown>)?.['apiVersion'];
    if (typeof apiVersion === 'string' && apiVersion.includes('/')) {
        return apiVersion.split('/')[0] === KAI_NODE_POOL_GROUP;
    }
    return usesKaiNodePoolCrd();
};

/**
 * Legacy field path -> how to read the same value off a KAI node pool. Paths that are absent here
 * (`status.phase`, `status.nodes`, `status.conditions`, `spec.labelKey`, `spec.labelValue`) are
 * identical on both CRDs and are read as-is.
 *
 * `spec.gpuResourceOptimization` and `spec.overProvisioningRatio` were dropped without an
 * equivalent, so they are deliberately not mapped - reading them off a KAI node pool yields
 * undefined.
 *
 * NOTE: `spec.placementStrategy.*` is NOT identical - on the KAI CRD it is nested under
 * `spec.schedulingShardConfig.placementStrategy.*` and is therefore mapped below.
 */
const kaiFieldAccessors: Record<string, (crd: Crd) => unknown> = {
    'status.gpuNetworkAccelerationDetected': (crd) => {
        const raw = annotation(crd, nodePoolAnnotations.gpuNetworkAccelerationDetected);
        return raw === undefined || raw === '' ? undefined : raw === 'true';
    },
    'spec.gpuNetworkAccelerationLabelKey': (crd) =>
        annotation(crd, nodePoolAnnotations.gpuNetworkAccelerationLabelKey) ??
        DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY,
    'spec.networkTopologyName': (crd) => get(crd, 'spec.preferredNetworkTopologyName'),
    'spec.minGuaranteedRuntime': (crd) => get(crd, 'spec.schedulingShardConfig.minRuntime.preemptMinRuntime'),
    'spec.placementStrategy.cpu': (crd) => get(crd, 'spec.schedulingShardConfig.placementStrategy.cpu'),
    'spec.placementStrategy.gpu': (crd) => get(crd, 'spec.schedulingShardConfig.placementStrategy.gpu')
};

/**
 * Reads a node pool CRD field addressed by its legacy path, from wherever the installed CRD keeps
 * it. Legacy paths stay the vocabulary of the call sites so that a single mapping table describes
 * the whole migration.
 */
export function getNodePoolCrdField<T = unknown>(crd: Crd, legacyPath: string): T {
    const accessor = isKaiNodePool(crd) ? kaiFieldAccessors[legacyPath] : undefined;
    return (accessor ? accessor(crd) : get(crd, legacyPath)) as T;
}
