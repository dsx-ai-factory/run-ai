/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { kubeCli } from '../../../../../../modules/cli';
import { flowRootState, versionUtils, waitUtils, workloadsUtilsV2 } from '../../../../../../modules/helpers';
import { crds } from '../../../../../../modules/helpers/models/customResourceDefinition';
import { CustomResourceDefinition } from '../../../../../../modules/helpers/models/customResourceDefinition';
import { SoftAssert } from '../../../../../../modules/helpers/SoftAssert';
import {
    isKubeflowInstalled,
    isLwsInstalled
} from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { SupportedVersions } from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

export const runningWlTestName = 'running workloads';
export const completedWlTestName = 'completed workloads';
export const failedWlTestName = 'failed workloads';
export const suspendedWlTestName = 'suspended workloads';
export const initializingWlTestName = 'initializing workloads';
export const pendingWlTestName = 'pending workloads';

interface testNamePhase {
    name: string;
    phase: Phase;
}
const running: testNamePhase = {
    name: runningWlTestName,
    phase: Phase.Running
};
const failed: testNamePhase = {
    name: failedWlTestName,
    phase: Phase.Failed
};
const suspended: testNamePhase = {
    name: suspendedWlTestName,
    phase: Phase.Stopped
};
const completed: testNamePhase = {
    name: completedWlTestName,
    phase: Phase.Completed
};
const pending: testNamePhase = {
    name: pendingWlTestName,
    phase: Phase.Pending
};
const initializing: testNamePhase = {
    name: initializingWlTestName,
    phase: Phase.Initializing
};

export const wlTestPhases = [running, failed, suspended, completed, initializing, pending];
export const workloadCategory = 'workload';
export const podsId = 'podsId';
export const workloadGen = 'workloadGen';
export const podsYaml = 'podsYaml';

// ---------------- v2 workload upgrade coverage ----------------
// Two submission paths share the same snapshot/verify shape:
//   Path A: workloads-v2 API -> WorkloadRunner CR
//   Path B: kubectl apply of a framework CRD -> ExternalWorkload wrapper picked up by EWI
export const v2WorkloadsKey = 'v2-workloads';

export type V2WorkloadKind = 'LeaderWorkerSet' | 'PyTorchJob';
export type V2WorkloadSubmissionPath = 'workloads-v2-api' | 'kubectl-apply-ewi';

export interface V2WorkloadSeed {
    kind: V2WorkloadKind;
    submissionPath: V2WorkloadSubmissionPath;
    id: string;
    name: string;
    namespace: string;
    projectId: string;
    podsId: string[];
    podsYaml: object[];
    crdKey: 'workloadRunnerCrd' | 'externalWorkloadCrd';
    externalWorkloadWrapperName?: string;
    yaml: object;
    workloadGen: number;
    innerManifest: object;
    phase: Phase;
}

/**
 * Returns the v2 workload kinds the pre-upgrade suite is expected to seed on
 * this cluster, applying the same gates the pre-upgrade tests use:
 *   - LeaderWorkerSet (Path A): cluster >= 2.24 (WorkloadRunner CRD) AND LWS installed
 *   - PyTorchJob (Path B): kubeflow training-operator installed
 * Single source of truth so the post-upgrade tests skip on exactly the same
 * condition as the pre-upgrade tests, rather than on an empty seed list (which
 * would hide a real bug if the pre-upgrade ran but a seed went missing).
 */
export async function expectedV2SeedKinds(): Promise<V2WorkloadKind[]> {
    const kinds: V2WorkloadKind[] = [];
    if (
        versionUtils.isVersionSupported(flowRootState.cluster.version, SupportedVersions.V2_24) &&
        (await isLwsInstalled())
    ) {
        kinds.push('LeaderWorkerSet');
    }
    if (await isKubeflowInstalled()) {
        kinds.push('PyTorchJob');
    }
    return kinds;
}

/**
 * Captures the post-submission snapshot for a v2 workload — mirrors the v1
 * `setWorkloadPodsData` / `setWorkloadsData` flow in workloadPreUpgrade.test.ts.
 * Caller passes the identifying fields and we fill in pod IDs, pod YAMLs,
 * wrapper-CR YAML, and metadata.generation.
 */
export async function captureV2WorkloadSeed(
    seed: Omit<V2WorkloadSeed, 'podsId' | 'podsYaml' | 'yaml' | 'workloadGen' | 'phase'>
): Promise<V2WorkloadSeed> {
    const pods = (await workloadsUtilsV2.getWorkloadPods(seed.id)).body.pods;
    const podsIdArr = pods.map((p) => p.id);
    const podsYamlArr: object[] = [];
    for (const pod of pods) {
        podsYamlArr.push(await kubeCli.getPodYaml(pod.name, seed.namespace));
    }

    const crd: CustomResourceDefinition = crds[seed.crdKey];
    const wrapperName = seed.externalWorkloadWrapperName ?? seed.name;
    const yaml = (await kubeCli.getCrd(crd, wrapperName, seed.namespace)) as {
        metadata: { generation: number };
    };

    return {
        ...seed,
        podsId: podsIdArr,
        podsYaml: podsYamlArr,
        yaml,
        workloadGen: yaml.metadata.generation,
        phase: Phase.Running
    };
}

/**
 * Runs the same 4 checks both before and after the upgrade — single source of
 * truth for what "healthy" means. Pre-upgrade: catches a bad seed before it
 * propagates into the post-upgrade test. Post-upgrade: validates the upgrade
 * preserved the workload.
 */
export async function verifyV2WorkloadSeedHealthy(
    seed: V2WorkloadSeed,
    softAssert: SoftAssert
): Promise<void> {
    const crd: CustomResourceDefinition = crds[seed.crdKey];
    const name = seed.externalWorkloadWrapperName ?? seed.name;

    try {
        const live = (await kubeCli.getCrd(crd, name, seed.namespace)) as {
            metadata: { generation: number };
            status?: {
                conditions?: { type: string; status: string }[];
                observedStatus?: { phase?: string };
            };
        };

        // Reconcile-health signal differs by wrapper kind:
        //   WorkloadRunner (Path A) sets status.conditions[LastReconcileSucceeded=True]
        //     once the workload-runner controller has fully observed + applied the
        //     embedded manifest.
        //   ExternalWorkload (Path B) has no LastReconcileSucceeded condition - EWI
        //     surfaces the underlying framework controller's phase under
        //     status.observedStatus.phase, so we check that === 'Running'.
        if (seed.crdKey === 'workloadRunnerCrd') {
            const reconcileOk = (live.status?.conditions ?? []).some(
                (c) => c.type === 'LastReconcileSucceeded' && c.status === 'True'
            );
            softAssert.assert(
                reconcileOk,
                `${seed.kind} (${seed.name}) LastReconcileSucceeded=True missing on ${crd.kind} ${name}`
            );
        } else {
            const observedPhase = live.status?.observedStatus?.phase;
            softAssert.assert(
                observedPhase === 'Running',
                `${seed.kind} (${seed.name}) ${crd.kind} ${name} status.observedStatus.phase = ${observedPhase} (expected Running)`
            );
        }

        softAssert.assertNonStrictEqual(
            live.metadata.generation,
            seed.workloadGen,
            `${seed.kind} (${seed.name}) generation drift on ${crd.kind} ${name}`
        );
    } catch (err) {
        softAssert.assert(
            false,
            `${seed.kind} (${seed.name}) unable to fetch ${crd.kind} ${name}: ${err}`
        );
    }

    try {
        const livePods = (await workloadsUtilsV2.getWorkloadPods(seed.id)).body.pods;
        const livePodIds = livePods.map((p) => p.id).sort();
        const snapshotPodIds = [...seed.podsId].sort();
        softAssert.assert(
            JSON.stringify(livePodIds) === JSON.stringify(snapshotPodIds),
            `${seed.kind} (${seed.name}) pod IDs changed: pre=${JSON.stringify(snapshotPodIds)} post=${JSON.stringify(livePodIds)}`
        );
    } catch (err) {
        softAssert.assert(false, `${seed.kind} (${seed.name}) unable to fetch live pods: ${err}`);
    }

    try {
        await waitUtils.waitForBackendWorkloadPhase(seed.name, seed.projectId, Phase.Running, 5);
    } catch (err) {
        softAssert.assert(
            false,
            `${seed.kind} (${seed.name}) backend phase not Running within 5s: ${err}`
        );
    }
}
