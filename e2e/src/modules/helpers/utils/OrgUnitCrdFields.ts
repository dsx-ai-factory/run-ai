/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { flowRootState } from '../config/config';
import { logger } from '../logger/logger';
import { kubeCli } from '../../cli';
import { crds, KAI_ORG_UNIT_GROUP, KAI_ORG_UNIT_VERSION } from '../models/customResourceDefinition';

let orgUnitCrdsResolved = false;

/**
 * Whether the cluster serves the KAI org-unit CRDs (projects.kai.resources,
 * departments.kai.resources) vs the legacy ones.
 *
 * Meaningful only after resolveOrgUnitCrds() has been awaited. Returns false
 * before resolution so that any accidental early access falls back safely.
 */
export function usesKaiOrgUnitCrds(): boolean {
    return flowRootState.orgUnitCrdsKai === true;
}

/** @internal Alias kept for call sites that used the old private name. */
const useNewOrgUnitCrds = usesKaiOrgUnitCrds;

/**
 * Applies the resolved org-unit CRD group to the shared descriptors in `crds`.
 * Downgrades to legacy groups when KAI CRDs are absent so that every test that
 * calls getCrd/getCrds/waitForCrd with crds.departmentCrd / crds.runaiProjectCrd
 * automatically hits the right API group.
 */
function applyOrgUnitCrds(isKai: boolean): void {
    if (!isKai) {
        crds.departmentCrd.group = 'scheduling.run.ai';
        crds.departmentCrd.version = 'v2alpha2';
        crds.runaiProjectCrd.group = 'run.ai';
        crds.runaiProjectCrd.version = 'v2';
    }
    // If isKai: keep the defaults (kai.resources/v1alpha1) already on the descriptors.
    flowRootState.orgUnitCrdsKai = isKai;
    orgUnitCrdsResolved = true;
}

/**
 * Resolves which org-unit CRD group this cluster serves and records it on
 * flowRootState so that workers read the answer from shared state instead of
 * probing the cluster again.
 *
 * Must be called once in global setup (before globalState.json is written).
 * Workers that receive a populated flowRootState.orgUnitCrdsKai skip the probe.
 */
export async function resolveOrgUnitCrds(): Promise<void> {
    if (orgUnitCrdsResolved) {
        return;
    }

    // Workers: take the answer from shared state written by global setup.
    if (flowRootState.orgUnitCrdsKai !== undefined) {
        applyOrgUnitCrds(flowRootState.orgUnitCrdsKai);
        return;
    }

    // Global setup: probe the cluster.
    const kaiDeptCrd = { ...crds.departmentCrd, group: KAI_ORG_UNIT_GROUP, version: KAI_ORG_UNIT_VERSION };
    let isKai: boolean;
    try {
        isKai = await kubeCli.isCrdInstalled(kaiDeptCrd);
    } catch (error) {
        logger.warn(`Could not probe org-unit CRDs (${error}); assuming KAI`);
        isKai = true;
    }
    logger.debug(`Org-unit CRDs: ${isKai ? 'KAI (kai.resources)' : 'legacy (run.ai / scheduling.run.ai)'}`);
    applyOrgUnitCrds(isKai);
}

/**
 * Drops the resolved CRD variant so the next access re-probes the cluster.
 * Needed by upgrade suites where the cluster agent is replaced underneath a running worker.
 */
export function resetOrgUnitCrdDetection(): void {
    orgUnitCrdsResolved = false;
    flowRootState.orgUnitCrdsKai = undefined;
    // Reset the descriptors back to KAI defaults so the next probe starts clean.
    crds.departmentCrd.group = KAI_ORG_UNIT_GROUP;
    crds.departmentCrd.version = KAI_ORG_UNIT_VERSION;
    crds.runaiProjectCrd.group = KAI_ORG_UNIT_GROUP;
    crds.runaiProjectCrd.version = KAI_ORG_UNIT_VERSION;
}

/**
 * Field accessors for the Project / Department CRDs across the KAI migration.
 */

/** Annotation keys on the KAI Project/Department CRDs (constants.go). */
export const orgUnitAnnotations = {
    updatedAt: 'run.ai/updated-at',
    lastSyncTimestamp: 'run.ai/last-sync-timestamp',
    forceDelete: 'kai/force-delete',
    nodeAffinityInteractive: 'run.ai/node-affinity-interactive',
    nodeAffinityTrain: 'run.ai/node-affinity-train',
    interactiveJobTimeLimitSecs: 'run.ai/interactive-job-time-limit-seconds',
    interactiveJobMaxIdleDurationSecs: 'run.ai/interactive-job-max-idle-duration-seconds',
    interactivePreemptibleJobMaxIdleDurationSecs:
        'run.ai/interactive-preemptible-job-max-idle-duration-seconds',
    trainingJobTimeLimitSecs: 'run.ai/training-job-time-limit-seconds',
    trainingJobMaxIdleDurationSecs: 'run.ai/training-job-max-idle-duration-seconds'
} as const;

/** Legacy `spec` field name -> KAI annotation key, for the numeric project job limits. */
export const projectJobLimitFields = {
    interactiveJobTimeLimitSecs: orgUnitAnnotations.interactiveJobTimeLimitSecs,
    interactiveJobMaxIdleDurationSecs: orgUnitAnnotations.interactiveJobMaxIdleDurationSecs,
    interactivePreemptibleJobMaxIdleDurationSecs:
        orgUnitAnnotations.interactivePreemptibleJobMaxIdleDurationSecs,
    trainingJobTimeLimitSecs: orgUnitAnnotations.trainingJobTimeLimitSecs,
    trainingJobMaxIdleDurationSecs: orgUnitAnnotations.trainingJobMaxIdleDurationSecs
} as const;

export type ProjectJobLimitField = keyof typeof projectJobLimitFields;

/** Legacy `spec` field name -> KAI annotation key, for the node-affinity lists. */
export const projectNodeAffinityFields = {
    nodeAffinityInteractive: orgUnitAnnotations.nodeAffinityInteractive,
    nodeAffinityTrain: orgUnitAnnotations.nodeAffinityTrain
} as const;

export type ProjectNodeAffinityField = keyof typeof projectNodeAffinityFields;

/**
 * A fetched org-unit CRD. Deliberately loose: call sites hold these as plain objects, as
 * `KubernetesObject`, or as the generated REST models, none of which share an index signature.
 */
type Crd = object;

const spec = (crd: Crd): Record<string, unknown> | undefined =>
    (crd as Record<string, unknown>)?.['spec'] as Record<string, unknown> | undefined;

const status = (crd: Crd): Record<string, unknown> | undefined =>
    (crd as Record<string, unknown>)?.['status'] as Record<string, unknown> | undefined;

const annotation = (crd: Crd, key: string): string | undefined => {
    const metadata = (crd as Record<string, unknown>)?.['metadata'] as
        | Record<string, unknown>
        | undefined;
    const annotations = metadata?.['annotations'] as Record<string, string> | undefined;
    return annotations?.[key];
};

/**
 * Control-plane last-modified timestamp (RFC3339).
 * legacy: spec.updatedAt -> KAI: run.ai/updated-at annotation.
 */
export function getOrgUnitUpdatedAt(crd: Crd): string | undefined {
    if (useNewOrgUnitCrds()) {
        return annotation(crd, orgUnitAnnotations.updatedAt);
    }
    return spec(crd)?.['updatedAt'] as string | undefined;
}

/**
 * Writes the control-plane last-modified timestamp into whichever location the CRD uses.
 * Used by tests that mutate an org unit to prove a webhook rejects the write, so the mutation
 * has to land on a field the CRD actually has (an unknown field would simply be pruned).
 */
export function setOrgUnitUpdatedAt(crd: Crd, value: string): void {
    const target = crd as Record<string, unknown>;
    if (useNewOrgUnitCrds()) {
        const metadata = (target['metadata'] ??= {}) as Record<string, unknown>;
        const annotations = (metadata['annotations'] ??= {}) as Record<string, string>;
        annotations[orgUnitAnnotations.updatedAt] = value;
        return;
    }
    const specObj = (target['spec'] ??= {}) as Record<string, unknown>;
    specObj['updatedAt'] = value;
}

/**
 * Timestamp of the last cluster-sync reconciliation (RFC3339).
 * legacy: status.syncStatus.lastSyncTimestamp -> KAI: run.ai/last-sync-timestamp annotation,
 * written by cluster-sync (project_parser.go:18).
 */
export function getOrgUnitLastSyncTimestamp(crd: Crd): string | undefined {
    if (useNewOrgUnitCrds()) {
        return annotation(crd, orgUnitAnnotations.lastSyncTimestamp);
    }
    const syncStatus = status(crd)?.['syncStatus'] as Record<string, unknown> | undefined;
    return syncStatus?.['lastSyncTimestamp'] as string | undefined;
}

/** True when the org unit carries a sync timestamp, wherever it lives. */
export function hasOrgUnitSyncTimestamp(crd: Crd): boolean {
    return Boolean(getOrgUnitLastSyncTimestamp(crd));
}

/**
 * The department a project belongs to.
 * legacy: spec.department -> KAI: spec.parent (project_types.go:38).
 */
export function getProjectParent(crd: Crd): string | undefined {
    const specObj = spec(crd);
    return (useNewOrgUnitCrds() ? specObj?.['parent'] : specObj?.['department']) as
        | string
        | undefined;
}

/**
 * A project job limit in seconds, addressed by its legacy spec field name.
 * legacy: spec.<field> (number) -> KAI: run.ai/*-seconds annotation (string).
 * Returns undefined when unset, so `toBeUndefined()` assertions keep working on both.
 */
export function getProjectJobLimitSecs(crd: Crd, field: ProjectJobLimitField): number | undefined {
    if (!useNewOrgUnitCrds()) {
        return spec(crd)?.[field] as number | undefined;
    }
    const raw = annotation(crd, projectJobLimitFields[field]);
    if (raw === undefined || raw === '') {
        return undefined;
    }
    const parsed = Number(raw);
    return Number.isNaN(parsed) ? undefined : parsed;
}

/**
 * A project node-affinity list, addressed by its legacy spec field name.
 * legacy: spec.<field> (string[]) -> KAI: run.ai/node-affinity-* annotation, JSON-encoded
 * []string (constants.go:70-71).
 */
export function getProjectNodeAffinity(
    crd: Crd,
    field: ProjectNodeAffinityField
): string[] | undefined {
    if (!useNewOrgUnitCrds()) {
        return spec(crd)?.[field] as string[] | undefined;
    }
    const raw = annotation(crd, projectNodeAffinityFields[field]);
    if (raw === undefined || raw === '') {
        return undefined;
    }
    try {
        const parsed = JSON.parse(raw);
        return Array.isArray(parsed) ? (parsed as string[]) : undefined;
    } catch {
        return undefined;
    }
}
