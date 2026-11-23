/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { KubernetesObject, V1OwnerReference, V1PersistentVolumeClaim, V1Pod } from '@kubernetes/client-node';
import { crds, waitUtils, workloadsManagerUtils } from '../index';
import { kubeCli } from '../../cli';

// Wire-friendly shape for a metadata.configuration.pvcs[] entry, mirroring the generated
// WorkloadV2UpdateConfigurationPvc but typing accessModes as string[] (its Set<enum> does not serialize).
export interface WorkloadV2PvcEntry {
    metadata: { name: string };
    spec: {
        accessModes: string[];
        resources: { requests: { storage: string } };
        storageClassName?: string;
    };
}

// The cases assume a default dynamic StorageClass named `standard` exists on the cluster.
export const STORAGE_CLASS = 'standard';

// WorkloadRunner CR shape — only the storage.newPvcs path the storage-plugin cases assert on.
export interface WorkloadRunnerResource extends KubernetesObject {
    spec?: {
        instructions?: {
            storage?: {
                newPvcs?: Array<{ metadata?: { name?: string }; spec?: Record<string, unknown> }>;
            };
        };
    };
}

// A Deployment (not a bare Pod): the V2 API requires a non-empty GVK group, which core/v1 Pod lacks.
// The managed PVC is driven by configuration.pvcs; the volume reference here is the realistic mount.
export const deploymentManifest = (
    name: string,
    volumes: { volName: string; claimName: string }[]
): Record<string, unknown> => ({
    apiVersion: 'apps/v1',
    kind: 'Deployment',
    metadata: { name },
    spec: {
        replicas: 1,
        selector: { matchLabels: { app: name } },
        template: {
            metadata: { labels: { app: name } },
            spec: {
                containers: [
                    {
                        name: 'app',
                        image: 'busybox',
                        command: ['sleep', '300'],
                        volumeMounts: volumes.map((v) => ({
                            name: v.volName,
                            mountPath: `/mnt/${v.volName}`
                        }))
                    }
                ],
                volumes: volumes.map((v) => ({
                    name: v.volName,
                    persistentVolumeClaim: { claimName: v.claimName }
                }))
            }
        }
    }
});

export const pvcEntry = (
    name: string,
    storage: string,
    opts: { storageClassName?: string; accessModes?: string[] } = {}
): WorkloadV2PvcEntry => ({
    metadata: { name },
    spec: {
        accessModes: opts.accessModes ?? ['ReadWriteOnce'],
        resources: { requests: { storage } },
        storageClassName: opts.storageClassName ?? STORAGE_CLASS
    }
});

export const findWorkloadRunner = async (
    forWorkload: string,
    namespace: string
): Promise<WorkloadRunnerResource> => {
    let runner: WorkloadRunnerResource | undefined;
    await waitUtils.waitForCondition(
        async () => {
            const items = (await kubeCli.getCrds(
                crds.workloadRunnerCrd,
                namespace
            )) as WorkloadRunnerResource[];
            runner = items.find((item) => item.metadata?.name?.includes(forWorkload));
            return runner !== undefined;
        },
        `WorkloadRunner for ${forWorkload} to be created`,
        true,
        120
    );
    return runner!;
};

export const waitForManagedPvc = async (
    pvcName: string,
    namespace: string
): Promise<V1PersistentVolumeClaim> => {
    let pvc: V1PersistentVolumeClaim | undefined;
    await waitUtils.waitForCondition(
        async () => {
            try {
                pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                    name: pvcName,
                    namespace
                });
                return true;
            } catch {
                return false;
            }
        },
        `managed PVC ${pvcName} to be created by the storage plugin`,
        true,
        120
    );
    return pvc!;
};

// A pod reaching Running already proves its PVC bound and mounted — an unmountable volume
// keeps the pod in ContainerCreating and it never becomes Running.
export const waitForRunningPod = async (appLabel: string, namespace: string): Promise<V1Pod> => {
    let running: V1Pod | undefined;
    await waitUtils.waitForCondition(
        async () => {
            const pods = (
                await kubeCli.CoreV1Api.listNamespacedPod({
                    namespace,
                    labelSelector: `app=${appLabel}`
                })
            ).items;
            running = pods.find((p) => p.status?.phase === 'Running');
            return running !== undefined;
        },
        `a pod labelled app=${appLabel} to reach Running (implies its PVC bound + mounted)`,
        true,
        180
    );
    return running!;
};

// Returns the controller ownerRef linking the PVC to its managing WorkloadRunner, or undefined if
// none — the caller asserts on it, keeping this util assertion-free.
export const workloadRunnerController = (pvc: V1PersistentVolumeClaim): V1OwnerReference | undefined =>
    (pvc.metadata?.ownerReferences ?? []).find(
        (ref) => ref.kind === 'WorkloadRunner' && ref.controller === true
    );

export const isPvcGone = async (pvcName: string, namespace: string): Promise<boolean> => {
    try {
        await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
            name: pvcName,
            namespace
        });
        return false;
    } catch {
        return true;
    }
};

// Deletes the workload, then waits until every named managed PVC is garbage-collected. The caller
// owns workloadId bookkeeping (reset it after) — this helper holds no test state.
export const deleteWorkloadAndWaitForPvcGc = async (
    workloadId: string,
    pvcNames: string[],
    namespace: string
): Promise<void> => {
    await workloadsManagerUtils.deleteWorkloadV2(workloadId);
    await waitUtils.waitForCondition(
        async () => (await Promise.all(pvcNames.map((n) => isPvcGone(n, namespace)))).every(Boolean),
        `managed PVCs [${pvcNames.join(', ')}] to be garbage-collected after workload delete`,
        true,
        120
    );
};
