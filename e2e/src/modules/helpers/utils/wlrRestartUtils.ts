/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from '@playwright/test';
import { KubernetesObject } from '@kubernetes/client-node';
import { kubeCli, bashCli } from '../../cli';
import { crds, CustomResourceDefinition } from '../models/customResourceDefinition';
import { waitUtils, Utils, distributedUtils, workloadsUtilsV2 } from '..';
import { DistributedFramework } from '../../api/apis/policyApi';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled,
    resolveWorkloadsContainerIndex,
    setAnyWorkloadSubmissionEnabledCache,
    WORKLOADS_DEPLOYMENT,
    WORKLOADS_NAMESPACE
} from './AnyWorkloadSubmissionUtils';
import { clusterApi } from '../../api';
import { flowRootState } from '../config/config';

// WLR restart protocol constants (designs/restart-controller-workloadrunner/design.md)
export const REQUEST_REF_LABEL = 'run.ai/request-ref';
export const WORKFLOW_ID_LABEL = 'heimdall.nvidia.com/workflow-id';
export const ADVICE_NAME_LABEL = 'run.ai/advice-name';
export const CREATED_BY_LABEL = 'run.ai/created-by';
export const WORKLOAD_NAME_LABEL = 'run.ai/workload-name';
export const WORKLOAD_ID_LABEL = 'run.ai/workload-id';
export const WORKLOAD_ID_ANNOTATION = 'workloadId';
export const NODE_AFFINITY_OVERRIDE_PATH = '/spec/instructions/nodeAffinity';
export const ADVICE_RESTART_OPERAND = 'advice-restart-controller';

export const DISTRIBUTED_SUBMISSION_ENV = 'ANY_WORKLOAD_DISTRIBUTED_SUBMISSION';

// waitForDeploymentToBeAvailable returns mid-rollout on a stale Progressing
// condition; this waits until the old pod is actually gone.
export async function waitForDeploymentRollout(
    name: string,
    namespace: string,
    timeout = 180
): Promise<void> {
    await waitUtils.waitForCondition(
        async () => {
            try {
                const d = await kubeCli.getDeployment(name, namespace);
                const status: any = d.status ?? {};
                const generation = d.metadata?.generation ?? 0;
                const want = d.spec?.replicas ?? 1;
                // counters are omitted when zero; replicas must equal want, or
                // maxSurge lets this pass while the old pod still serves
                return (
                    (status.observedGeneration ?? 0) >= generation &&
                    (status.updatedReplicas ?? 0) >= want &&
                    (status.availableReplicas ?? 0) >= want &&
                    (status.replicas ?? 0) === want
                );
            } catch {
                return false;
            }
        },
        `${namespace}/${name} rollout complete`,
        true,
        timeout,
        3
    );
}

// The distributed WLR path ships dark (chart default off). Turn it on for the
// suite instead of skipping, so the tests actually run in CI. Returns restore
// bookkeeping for afterAll.
export async function ensureDistributedWlrSubmission(): Promise<{
    changed: boolean;
    priorValue?: string;
}> {
    if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Distributed, true)) {
        return { changed: false };
    }
    const deployment = await kubeCli.getDeployment(WORKLOADS_DEPLOYMENT, WORKLOADS_NAMESPACE);
    const containers = deployment.spec!.template!.spec!.containers;
    const idx = resolveWorkloadsContainerIndex(containers);
    const env = (containers[idx].env = containers[idx].env ?? []);
    const entry = env.find((e) => e.name === DISTRIBUTED_SUBMISSION_ENV);
    const priorValue = entry?.value;
    if (entry) {
        entry.value = 'true';
    } else {
        env.push({ name: DISTRIBUTED_SUBMISSION_ENV, value: 'true' });
    }
    await kubeCli.AppsV1Api.replaceNamespacedDeployment({
        name: WORKLOADS_DEPLOYMENT,
        namespace: WORKLOADS_NAMESPACE,
        body: deployment
    });
    await waitForDeploymentRollout(WORKLOADS_DEPLOYMENT, WORKLOADS_NAMESPACE);
    setAnyWorkloadSubmissionEnabledCache(undefined);
    return { changed: true, priorValue };
}

export async function restoreDistributedWlrSubmission(priorValue?: string): Promise<void> {
    const deployment = await kubeCli.getDeployment(WORKLOADS_DEPLOYMENT, WORKLOADS_NAMESPACE);
    const containers = deployment.spec!.template!.spec!.containers;
    const idx = resolveWorkloadsContainerIndex(containers);
    const env = containers[idx].env ?? [];
    const entryIdx = env.findIndex((e) => e.name === DISTRIBUTED_SUBMISSION_ENV);
    if (priorValue === undefined) {
        if (entryIdx >= 0) env.splice(entryIdx, 1);
    } else if (entryIdx >= 0) {
        env[entryIdx].value = priorValue;
    } else {
        env.push({ name: DISTRIBUTED_SUBMISSION_ENV, value: priorValue });
    }
    containers[idx].env = env;
    await kubeCli.AppsV1Api.replaceNamespacedDeployment({
        name: WORKLOADS_DEPLOYMENT,
        namespace: WORKLOADS_NAMESPACE,
        body: deployment
    });
    await waitForDeploymentRollout(WORKLOADS_DEPLOYMENT, WORKLOADS_NAMESPACE);
    setAnyWorkloadSubmissionEnabledCache(undefined);
}

// The WLR restart flow has no released version label yet, so version gating cannot
// detect it. Detect the capability directly instead: a restart-capable cluster ships
// a WorkloadRunner CRD whose schema knows spec.managedBy and
// spec.instructions.nodeAffinity - the two fields the restart machinery writes.
export async function clusterSupportsWlrRestart(): Promise<boolean> {
    try {
        const crd: any = await kubeCli.ApiextensionsV1Api.readCustomResourceDefinition({
            name: 'workloadrunners.run.ai'
        });
        const specProps = crd?.spec?.versions?.[0]?.schema?.openAPIV3Schema?.properties?.spec?.properties;
        return (
            specProps?.managedBy !== undefined &&
            specProps?.instructions?.properties?.nodeAffinity !== undefined
        );
    } catch {
        return false;
    }
}

export interface NodeSelectorTerm {
    matchExpressions?: { key: string; operator: string; values: string[] }[];
}

export function getExcludedNodes(terms: NodeSelectorTerm[]): string[] {
    for (const term of terms) {
        for (const expr of term.matchExpressions ?? []) {
            if (expr.key === 'kubernetes.io/hostname' && expr.operator === 'NotIn') {
                return [...expr.values].sort();
            }
        }
    }
    return [];
}

// union of every kubernetes.io/hostname NotIn value across ALL terms and ALL
// matchExpressions - user affinity and plugin-merged exclusions may live in
// separate expressions, so first-match readers under-report
export function allHostnameExclusions(terms: NodeSelectorTerm[]): string[] {
    const values = new Set<string>();
    for (const term of terms) {
        for (const expr of term.matchExpressions ?? []) {
            if (expr.key === 'kubernetes.io/hostname' && expr.operator === 'NotIn') {
                expr.values.forEach((v) => values.add(v));
            }
        }
    }
    return [...values].sort();
}

export function childReplicaTerms(pytorchJob: any, replica: 'Master' | 'Worker'): NodeSelectorTerm[] {
    return (
        pytorchJob.spec?.pytorchReplicaSpecs?.[replica]?.template?.spec?.affinity?.nodeAffinity
            ?.requiredDuringSchedulingIgnoredDuringExecution?.nodeSelectorTerms ?? []
    );
}

// all Ready, schedulable, non-control-plane nodes - workload pods on kind/dev
// clusters schedule on system AND worker nodes, so worker-role filters under-count
export async function getSchedulableNodeNames(): Promise<string[]> {
    const nodes = await kubeCli.CoreV1Api.listNode();
    return (nodes.items || [])
        .filter((n: any) => {
            const labels = n.metadata?.labels ?? {};
            if (
                'node-role.kubernetes.io/control-plane' in labels ||
                'node-role.kubernetes.io/master' in labels
            ) {
                return false;
            }
            if (n.spec?.unschedulable) return false;
            const ready = (n.status?.conditions || []).find((c: any) => c.type === 'Ready');
            return ready?.status === 'True';
        })
        .map((n: any) => n.metadata.name)
        .sort();
}

export function wlrInstructionExclusions(wlr: any): string[] {
    return getExcludedNodes(
        wlr.spec?.instructions?.nodeAffinity?.nodeAffinity?.requiredDuringSchedulingIgnoredDuringExecution
            ?.nodeSelectorTerms ?? []
    );
}

export function childReplicaExclusions(pytorchJob: any, replica: 'Master' | 'Worker'): string[] {
    return getExcludedNodes(
        pytorchJob.spec?.pytorchReplicaSpecs?.[replica]?.template?.spec?.affinity?.nodeAffinity
            ?.requiredDuringSchedulingIgnoredDuringExecution?.nodeSelectorTerms ?? []
    );
}

export function buildAdviceBody(
    name: string,
    namespace: string,
    jobName: string,
    jobUid: string,
    state: string,
    exclusions?: string[],
    jobGvk: { apiVersion: string; kind: string } = { apiVersion: 'kubeflow.org/v1', kind: 'PyTorchJob' }
): object {
    const body: any = {
        apiVersion: 'heimdall.nvidia.com/v1alpha1',
        kind: 'Advice',
        metadata: { name, namespace },
        spec: {
            job: {
                heimdallId: `heimdall-${name}`,
                jobId: jobUid,
                apiVersion: jobGvk.apiVersion,
                kind: jobGvk.kind,
                name: jobName,
                attemptIndex: 0
            },
            advice: { state }
        }
    };
    if (exclusions && exclusions.length > 0) {
        body.spec.advice.nodeConstraints = { exclusions };
    }
    return body;
}

// canonical valid RestartRequest body; negative tests then break one field per case
export function buildRestartRequestBody(opts: {
    name: string;
    namespace: string;
    wlName: string;
    workloadId: string;
    generation: number;
    specOverrides?: { op: string; path: string; value: unknown }[];
}): any {
    const body: any = {
        apiVersion: 'run.ai/v1alpha1',
        kind: 'RestartRequest',
        metadata: {
            name: opts.name,
            namespace: opts.namespace,
            labels: {
                [WORKLOAD_NAME_LABEL]: opts.wlName,
                [WORKLOAD_ID_LABEL]: opts.workloadId
            }
        },
        spec: {
            workloadRef: {
                gvk: { group: 'run.ai', version: 'v1alpha1', kind: 'WorkloadRunner' },
                name: opts.wlName,
                generation: opts.generation
            }
        }
    };
    if (opts.specOverrides) {
        body.spec.specOverrides = opts.specOverrides;
    }
    return body;
}

// thin domain alias over kubeCli.getCrd (throws when the WLR is missing)
export async function getWlr(name: string, namespace: string): Promise<any> {
    return (await kubeCli.getCrd(crds.workloadRunnerCrd, name, namespace)) as any;
}

// unlike kubeCli.getCrd this resolves to undefined when the child is missing -
// restart-window polls rely on that to observe the delete/recreate transition
export async function getPytorchJob(name: string, namespace: string): Promise<any> {
    try {
        return (await kubeCli.getCrd(crds.pyTorchJobCrd, name, namespace)) as any;
    } catch {
        return undefined;
    }
}

// after resume the WLR and its child PyTorchJob are recreated from scratch
export async function waitForWlrAndChildBack(name: string, namespace: string, timeoutSec = 180): Promise<void> {
    await waitUtils.waitForCondition(
        async () => {
            const wlrs = (await kubeCli.getCrds(crds.workloadRunnerCrd, namespace)) || [];
            return (
                wlrs.some((w: any) => w.metadata?.name === name) &&
                (await getPytorchJob(name, namespace)) !== undefined
            );
        },
        `WLR and child ${name} recreated after resume`,
        true,
        timeoutSec
    );
}

export async function getRestartRequestsForAdvice(
    advName: string,
    namespace: string
): Promise<KubernetesObject[]> {
    const allRRs = await kubeCli.getCrds(crds.restartRequestCrd, namespace);
    return (allRRs || []).filter((rr: any) => rr.metadata?.labels?.[ADVICE_NAME_LABEL] === advName);
}

// settle window then assert the advice produced exactly `expected` RestartRequests.
// The sleep is load-bearing: it asserts ABSENCE, so it cannot be replaced by a poll.
export async function expectRestartRequestCountAfterSettle(
    adviceName: string,
    namespace: string,
    expected = 0,
    settleSec = 15
): Promise<void> {
    await Utils.delay(settleSec * 1000);
    const rrs = await getRestartRequestsForAdvice(adviceName, namespace);
    expect(rrs.length, `RestartRequests for advice ${adviceName} after ${settleSec}s settle`).toBe(expected);
}

export async function getWlrEvents(wlrName: string, namespace: string): Promise<any[]> {
    const events = await kubeCli.CoreV1Api.listNamespacedEvent({ namespace });
    return (events.items || []).filter((e: any) => e.involvedObject?.name === wlrName);
}

// node names the workload's live pods are scheduled on (unique, sorted),
// read through the cp pods api (completed leftovers are filtered out)
export async function getWorkloadPodNodes(workloadId: string): Promise<string[]> {
    const pods = (await workloadsUtilsV2.getWorkloadPods(workloadId)).body.pods ?? [];
    const nodes = pods
        .filter((p: any) => !p.completedAt)
        .map((p: any) => p.nodeName)
        .filter((n: string | null | undefined): n is string => !!n);
    return [...new Set(nodes)].sort();
}

// wait until every pod of the workload is scheduled, return their nodes
export async function waitForScheduledNodes(workloadId: string, timeoutSec = 120): Promise<string[]> {
    let nodes: string[] = [];
    await waitUtils.waitForCondition(
        async () => {
            nodes = await getWorkloadPodNodes(workloadId);
            return nodes.length > 0;
        },
        `pods of workload ${workloadId} scheduled on nodes`,
        true,
        timeoutSec
    );
    return nodes;
}

// old terminating pods linger after a restart and pollute a read-once
// node check - poll until every live pod sits OFF the excluded set
export async function waitForNodesOffExcluded(
    workloadId: string,
    excluded: string[],
    timeoutSec = 240
): Promise<void> {
    await waitUtils.waitForCondition(
        async () => {
            const nodes = await getWorkloadPodNodes(workloadId);
            return nodes.length > 0 && nodes.every((n) => !excluded.includes(n));
        },
        `pods of workload ${workloadId} rescheduled off [${excluded.join(',')}]`,
        true,
        timeoutSec
    );
}

export async function createPytorchWorkload(
    projectId: string,
    name: string,
    workerExcludeNodeNames?: string[]
): Promise<string> {
    return Utils.retry(
        async () => {
            const { body } = await distributedUtils.createDistributedTraining(
                projectId,
                name,
                {
                    distributedFramework: DistributedFramework.PyTorch,
                    image: 'alpine:latest',
                    command: 'sleep infinity',
                    numWorkers: 1,
                    compute: {
                        gpuDevicesRequest: 0,
                        cpuCoreRequest: 0.1,
                        cpuMemoryRequest: '100M'
                    }
                },
                undefined,
                workerExcludeNodeNames
            );
            return body.workloadId;
        },
        { retries: 3, delay: 10 * 1000 }
    );
}

// Playwright restarts the worker process after any test failure, wiping
// describe-scope state. Re-create / re-derive the shared workload so later
// tests never cascade-fail.
export async function ensureWorkload(
    projectId: string,
    namespace: string,
    name: string,
    cachedId?: string
): Promise<string> {
    const wlrs = (await kubeCli.getCrds(crds.workloadRunnerCrd, namespace)) || [];
    if (!wlrs.some((w: any) => w.metadata?.name === name)) {
        const workloadId = await createPytorchWorkload(projectId, name);
        await waitUtils.waitForCrd(crds.workloadRunnerCrd, name, namespace, 90);
        await waitUtils.waitForCrd(crds.pyTorchJobCrd, name, namespace, 90);
        return workloadId;
    }
    if (cachedId) {
        return cachedId;
    }
    const wlr = await getWlr(name, namespace);
    return wlr.metadata?.annotations?.[WORKLOAD_ID_ANNOTATION];
}

// pods on kind/dev clusters schedule on system AND worker nodes - capacity
// must count every schedulable node, not just worker-role ones
export async function skipIfFewerThanTwoSchedulableNodes(test: any, what: string): Promise<string[]> {
    const schedulable = await getSchedulableNodeNames();
    // env-capability skip (not a bug), test.skip per repo convention
    test.skip(
        schedulable.length < 2,
        `${what} needs >=2 schedulable nodes, cluster has ${schedulable.length}`
    );
    return schedulable;
}

// Applies an Advice and drives one full restart cycle to Succeeded.
// Returns the terminal RestartRequest.
export async function runRestartCycle(
    wlName: string,
    namespace: string,
    adviceName: string,
    exclusions?: string[],
    hooks?: { onRequestCreated?(rr: any): void | Promise<void> }
): Promise<any> {
    const child = await getPytorchJob(wlName, namespace);
    expect(child, `child PyTorchJob ${wlName} exists before restart`).toBeDefined();
    const oldUid = child.metadata.uid;

    await kubeCli.createCrd(
        crds.adviceCrd,
        buildAdviceBody(adviceName, namespace, wlName, oldUid, 'ADVICE_STATE_RESTARTED', exclusions),
        namespace
    );

    await waitUtils.waitForCondition(
        async () => (await getRestartRequestsForAdvice(adviceName, namespace)).length >= 1,
        `RestartRequest created for advice ${adviceName}`,
        true,
        60
    );

    if (hooks?.onRequestCreated) {
        await hooks.onRequestCreated((await getRestartRequestsForAdvice(adviceName, namespace))[0]);
    }

    await waitUtils.waitForCondition(
        async () => {
            const rrs = await getRestartRequestsForAdvice(adviceName, namespace);
            return rrs.some((rr: any) => rr.status?.phase === 'Succeeded');
        },
        `RestartRequest for ${adviceName} reached Succeeded`,
        true,
        180
    );

    // handback: lock cleared, child recreated with a new UID
    await waitUtils.waitForCondition(
        async () => {
            const wlr = (await kubeCli.getCrd(crds.workloadRunnerCrd, wlName, namespace)) as any;
            const lockCleared = !wlr?.spec?.managedBy && !wlr?.metadata?.labels?.[REQUEST_REF_LABEL];
            const newChild = await getPytorchJob(wlName, namespace);
            return lockCleared && newChild !== undefined && newChild.metadata.uid !== oldUid;
        },
        `handback done and child recreated for ${wlName}`,
        true,
        120
    );

    const rrs = await getRestartRequestsForAdvice(adviceName, namespace);
    return rrs.find((rr: any) => (rr as any).status?.phase === 'Succeeded');
}

export async function deleteAllCrInstancesInNamespace(
    crd: CustomResourceDefinition,
    namespace: string
): Promise<void> {
    const items = await kubeCli.getCrds(crd, namespace);
    for (const item of items || []) {
        await kubeCli.deleteCrd(crd, (item as any).metadata.name, namespace);
    }
}

// kubectl-based mutation that is EXPECTED to be denied by a webhook/VAP.
export function expectKubectlDenied(cmd: string, expectedSubstring: string): void {
    const result = bashCli.execWithOutput(cmd, false);
    expect(result.status, `expected denial for: ${cmd}`).not.toBe(0);
    const errText = `${result.stderr ?? ''}${result.stdout ?? ''}`;
    expect(errText.toLowerCase()).toContain(expectedSubstring.toLowerCase());
}

// The CP stamps the workflow-id at SUBMISSION time, gated on the operand state it
// learns via cluster-sync (~10s cadence) behind a 600s client cache. Any submit
// that races that propagation is composed unstamped forever, so wait for the
// CP-visible operand before submitting anything.
export async function waitForCpVisibleAdviceOperand(): Promise<void> {
    await waitUtils.waitForCondition(
        async () => {
            clusterApi.clustersApi.accessToken = flowRootState.users.system_administrator.accessToken;
            const clusters = (await clusterApi.clustersApi.getClusters()).body;
            const cluster = clusters.find((c: any) => c.uuid === flowRootState.cluster.uuid) as any;
            return (
                cluster?.status?.state === 'Connected' &&
                cluster?.status?.operands?.[ADVICE_RESTART_OPERAND]?.ready === true
            );
        },
        'control plane reports advice-restart-controller operand Ready',
        true,
        120,
        5
    );
}
