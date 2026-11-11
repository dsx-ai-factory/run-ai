/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    crds,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    workloadsManagerUtils
} from '../../../../../../modules/helpers';
import { WorkloadV2SyncStatus } from '../../../../../../modules/api/apis/workloadsManagerApi/generated/model/workloadV2SyncStatus';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { kubeCli } from '../../../../../../modules/cli';
import {
    STORAGE_CLASS,
    WorkloadRunnerResource,
    deploymentManifest,
    pvcEntry,
    findWorkloadRunner,
    waitForManagedPvc,
    waitForRunningPod,
    workloadRunnerController,
    isPvcGone,
    deleteWorkloadAndWaitForPvcGc
} from '../../../../../../modules/helpers/utils/storagePluginNewPvcsHelpers';

// Storage plugin (AnyWorkload), V2 API: metadata.configuration.pvcs[] -> WorkloadRunner
// .spec.instructions.storage.newPvcs -> dromos plugin creates/owns/GCs the managed PVC. Gated >=2.26.0.

test.describe(
    `Storage plugin (anyworkload) - V2 configuration.pvcs (API), tags: ${tags.tool.api} ${tags.resource.workloadsV2} ${tags.teams.loong}`,
    function () {
        supportedFrom(SupportedVersions.V2_26)(function () {
            let project: Project;
            let workloadId: string | undefined;
            let pvcsToCleanup: string[];

            test.beforeAll('create a shared project + namespace', async function () {
                project = await projectResourcesUtils.createProjectAndNamespace('storage-plugin');
            });

            test.beforeEach(function () {
                workloadsManagerUtils.accessToken = flowRootState.users.system_administrator.accessToken;
                workloadId = undefined;
                pvcsToCleanup = [];
            });

            test.afterEach('clean up workload + any tracked PVCs', async function () {
                if (workloadId) {
                    try {
                        await workloadsManagerUtils.deleteWorkloadV2(workloadId);
                    } catch (e) {
                        logger.info(`Workload cleanup skipped for ${workloadId}: ${e.message}`);
                    }
                }
                for (const pvcName of pvcsToCleanup) {
                    try {
                        await kubeCli.deletePvc(pvcName, project.namespace);
                    } catch (e) {
                        logger.info(`PVC cleanup skipped for ${pvcName}: ${e.message}`);
                    }
                }
            });

            test.afterAll('delete the shared project + namespace', async function () {
                if (project) {
                    try {
                        await operations.deleteProjectAndNamespace(project.id);
                    } catch (e) {
                        logger.error(`Failed to delete project ${project.name}`, e);
                    }
                }
            });

            test('T6140989: submits a V2 workload with configuration.pvcs and the plugin creates an owned, GC-tracked PVC', async function () {
                const workloadName = objectGenerator.generateShortNameByPrefix('storage-wl', 5);
                const pvcName = objectGenerator.generateShortNameByPrefix('managed-pvc', 5);

                await test.step('Submit the V2 workload carrying metadata.configuration.pvcs', async () => {
                    workloadId = await workloadsManagerUtils.createWorkloadV2WithPvcs(
                        workloadName,
                        project.id.toString(),
                        [pvcEntry(pvcName, '1Gi')],
                        deploymentManifest(workloadName, [{ volName: 'new-vol', claimName: pvcName }])
                    );
                });

                await test.step('GET returns the workload (desiredPhase=Running) and it syncs to the cluster', async () => {
                    const res = await workloadsManagerUtils.getWorkloadV2ById(workloadId!);
                    expect(res.response.statusCode, 'GET workload should return 200').toBe(200);
                    expect(res.body.workload.desiredPhase, 'workload should be desired Running').toBe(
                        'Running'
                    );
                    // Wait until the control-plane reports the workload applied (syncInfo.status=Synced) — the
                    // API-level "running" signal (not pod readiness, which the raw V2 API omits).
                    const synced = await workloadsManagerUtils.waitForWorkloadV2Synced(workloadId!);
                    expect(synced.syncInfo?.status, 'workload should reach Synced').toBe('Synced');
                });

                await test.step('WorkloadRunner declares the PVC under .spec.instructions.storage.newPvcs, separate from mounts', async () => {
                    const runner = await findWorkloadRunner(workloadName, project.namespace);
                    const declared = runner.spec?.instructions?.storage?.newPvcs ?? [];
                    expect(declared.map((p) => p.metadata?.name)).toContain(pvcName);
                    // The declaration only lists PVCs to create; mount semantics live in the manifest.
                    const entry = declared.find((p) => p.metadata?.name === pvcName);
                    expect(JSON.stringify(entry ?? {})).not.toContain('volumeMounts');
                    expect(JSON.stringify(entry ?? {})).not.toContain('mountPath');
                });

                await test.step('The plugin creates the managed PVC in the runner namespace, owned by the WorkloadRunner', async () => {
                    const pvc = await waitForManagedPvc(pvcName, project.namespace);
                    expect(
                        workloadRunnerController(pvc),
                        `${pvcName} should carry a controller ownerRef to its WorkloadRunner`
                    ).toBeDefined();
                    expect(pvc.spec?.storageClassName).toBe(STORAGE_CLASS);
                    expect(pvc.spec?.accessModes).toEqual(['ReadWriteOnce']);
                    expect(pvc.spec?.resources?.requests?.storage).toBe('1Gi');
                });

                await test.step('The pod runs and the managed PVC is bound + mounted (writable) at its mountPath', async () => {
                    // Pod Running already implies the PVC bound and mounted; assert both explicitly.
                    const pod = await waitForRunningPod(workloadName, project.namespace);

                    const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                        name: pvcName,
                        namespace: project.namespace
                    });
                    expect(pvc.status?.phase, 'managed PVC should be Bound').toBe('Bound');

                    // Prove the volume is mounted AND writable at the manifest's mountPath (/mnt/new-vol).
                    const probe = `runai-storage-probe-${pvcName}`;
                    const { stdout } = await kubeCli.execInPod(pod.metadata!.name!, project.namespace, [
                        'sh',
                        '-c',
                        `echo ${probe} > /mnt/new-vol/probe && cat /mnt/new-vol/probe`
                    ]);
                    expect(stdout, 'PVC should be mounted and writable at /mnt/new-vol').toContain(probe);
                });

                await test.step('Deleting the workload garbage-collects the managed PVC', async () => {
                    await deleteWorkloadAndWaitForPvcGc(workloadId!, [pvcName], project.namespace);
                    workloadId = undefined;
                });
            });

            test('T6140999: creates each PVC independently when a workload declares multiple pvcs', async function () {
                const workloadName = objectGenerator.generateShortNameByPrefix('storage-multi-wl', 5);
                const pvcA = objectGenerator.generateShortNameByPrefix('multi-pvc-a', 5);
                const pvcB = objectGenerator.generateShortNameByPrefix('multi-pvc-b', 5);

                await test.step('Submit one workload declaring two distinct pvcs', async () => {
                    workloadId = await workloadsManagerUtils.createWorkloadV2WithPvcs(
                        workloadName,
                        project.id.toString(),
                        [pvcEntry(pvcA, '1Gi'), pvcEntry(pvcB, '2Gi')],
                        deploymentManifest(workloadName, [
                            { volName: 'vol-a', claimName: pvcA },
                            { volName: 'vol-b', claimName: pvcB }
                        ])
                    );
                });

                await test.step('The workload syncs to the cluster (V2 API syncInfo.status=Synced)', async () => {
                    const synced = await workloadsManagerUtils.waitForWorkloadV2Synced(workloadId!);
                    expect(synced.syncInfo?.status, 'multi-pvc workload should reach Synced').toBe('Synced');
                });

                await test.step('WorkloadRunner preserves both PVC declarations', async () => {
                    const runner = await findWorkloadRunner(workloadName, project.namespace);
                    const names = (runner.spec?.instructions?.storage?.newPvcs ?? []).map(
                        (p) => p.metadata?.name
                    );
                    expect(names).toContain(pvcA);
                    expect(names).toContain(pvcB);
                    expect(names.length).toBe(2);
                });

                await test.step('Both PVCs are created independently, owned by the same WorkloadRunner, each with its own size', async () => {
                    const a = await waitForManagedPvc(pvcA, project.namespace);
                    const b = await waitForManagedPvc(pvcB, project.namespace);
                    expect(
                        workloadRunnerController(a),
                        `${pvcA} should carry a controller ownerRef to its WorkloadRunner`
                    ).toBeDefined();
                    expect(
                        workloadRunnerController(b),
                        `${pvcB} should carry a controller ownerRef to its WorkloadRunner`
                    ).toBeDefined();
                    expect(a.spec?.resources?.requests?.storage).toBe('1Gi');
                    expect(b.spec?.resources?.requests?.storage).toBe('2Gi');
                    expect(a.metadata?.uid).not.toBe(b.metadata?.uid);
                });

                await test.step('The pod runs and both managed PVCs are bound + mounted (writable) at their mountPaths', async () => {
                    // One pod mounts both volumes; Running implies both PVCs bound + mounted.
                    const pod = await waitForRunningPod(workloadName, project.namespace);

                    for (const { name, mount } of [
                        { name: pvcA, mount: '/mnt/vol-a' },
                        { name: pvcB, mount: '/mnt/vol-b' }
                    ]) {
                        const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                            name,
                            namespace: project.namespace
                        });
                        expect(pvc.status?.phase, `${name} should be Bound`).toBe('Bound');

                        const probe = `runai-storage-probe-${name}`;
                        const { stdout } = await kubeCli.execInPod(pod.metadata!.name!, project.namespace, [
                            'sh',
                            '-c',
                            `echo ${probe} > ${mount}/probe && cat ${mount}/probe`
                        ]);
                        expect(stdout, `${name} should be mounted and writable at ${mount}`).toContain(probe);
                    }
                });

                await test.step('Deleting the workload garbage-collects both PVCs', async () => {
                    await deleteWorkloadAndWaitForPvcGc(workloadId!, [pvcA, pvcB], project.namespace);
                    workloadId = undefined;
                });
            });

            test('T6140995: leaves a pre-existing PVC unmanaged when the workload declares no pvcs', async function () {
                const workloadName = objectGenerator.generateShortNameByPrefix('storage-existing-wl', 5);
                const existingPvc = objectGenerator.generateShortNameByPrefix('pre-existing-pvc', 5);

                await test.step('Pre-create a PVC outside the workload path', async () => {
                    await kubeCli.CoreV1Api.createNamespacedPersistentVolumeClaim({
                        namespace: project.namespace,
                        body: {
                            apiVersion: 'v1',
                            kind: 'PersistentVolumeClaim',
                            metadata: { name: existingPvc },
                            spec: {
                                accessModes: ['ReadWriteOnce'],
                                resources: { requests: { storage: '1Gi' } },
                                storageClassName: STORAGE_CLASS
                            }
                        }
                    });
                    pvcsToCleanup.push(existingPvc);
                });

                await test.step('Submit a workload that mounts the pre-existing PVC but declares no pvcs', async () => {
                    workloadId = await workloadsManagerUtils.createWorkloadV2WithPvcs(
                        workloadName,
                        project.id.toString(),
                        [],
                        deploymentManifest(workloadName, [
                            { volName: 'existing-vol', claimName: existingPvc }
                        ])
                    );
                });

                await test.step('The workload syncs to the cluster (V2 API syncInfo.status=Synced)', async () => {
                    const synced = await workloadsManagerUtils.waitForWorkloadV2Synced(workloadId!);
                    expect(synced.syncInfo?.status, 'no-pvcs workload should reach Synced').toBe('Synced');
                });

                await test.step('WorkloadRunner declares no managed PVCs', async () => {
                    const runner = await findWorkloadRunner(workloadName, project.namespace);
                    const declared = runner.spec?.instructions?.storage?.newPvcs ?? [];
                    expect(declared.length).toBe(0);
                });

                await test.step('The plugin does not take ownership of the pre-existing PVC', async () => {
                    const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                        name: existingPvc,
                        namespace: project.namespace
                    });
                    const wrOwner = (pvc.metadata?.ownerReferences ?? []).find(
                        (ref) => ref.kind === 'WorkloadRunner'
                    );
                    expect(
                        wrOwner,
                        'pre-existing PVC must NOT gain a WorkloadRunner ownerRef'
                    ).toBeUndefined();
                });

                await test.step('The pod runs and the pre-existing PVC is mounted (writable) at its mountPath', async () => {
                    // The workload mounts the pre-existing PVC; confirm it actually runs and mounts,
                    // even though the storage plugin does not manage/own that PVC.
                    const pod = await waitForRunningPod(workloadName, project.namespace);

                    const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                        name: existingPvc,
                        namespace: project.namespace
                    });
                    expect(pvc.status?.phase, 'pre-existing PVC should be Bound').toBe('Bound');

                    const probe = `runai-storage-probe-${existingPvc}`;
                    const { stdout } = await kubeCli.execInPod(pod.metadata!.name!, project.namespace, [
                        'sh',
                        '-c',
                        `echo ${probe} > /mnt/existing-vol/probe && cat /mnt/existing-vol/probe`
                    ]);
                    expect(
                        stdout,
                        'pre-existing PVC should be mounted and writable at /mnt/existing-vol'
                    ).toContain(probe);
                });

                await test.step('Deleting the workload does NOT garbage-collect the pre-existing PVC', async () => {
                    await workloadsManagerUtils.deleteWorkloadV2(workloadId!);
                    workloadId = undefined;
                    await waitUtils.waitForCondition(
                        async () => {
                            const runners = (await kubeCli.getCrds(
                                crds.workloadRunnerCrd,
                                project.namespace
                            )) as WorkloadRunnerResource[];
                            return (
                                runners.find((r) => r.metadata?.name?.includes(workloadName)) === undefined
                            );
                        },
                        `WorkloadRunner for ${workloadName} to be removed`,
                        true,
                        120
                    );
                    expect(
                        await isPvcGone(existingPvc, project.namespace),
                        'pre-existing PVC must still exist after workload delete'
                    ).toBe(false);
                });
            });

            test('T6140996: rejects malformed PVC entries (unknown fields and missing required fields) at the V2 API', async function () {
                // V2 create rejects malformed PVC entries at the API (4xx): unknown fields (apiVersion/kind/
                // namespace/bogus spec) via additionalProperties:false, and missing required accessModes/storage.
                const validSpec = {
                    accessModes: ['ReadWriteOnce'],
                    resources: { requests: { storage: '1Gi' } },
                    storageClassName: STORAGE_CLASS
                };
                const variants: { label: string; pvc: Record<string, unknown> }[] = [
                    {
                        label: 'extra top-level apiVersion/kind',
                        pvc: {
                            apiVersion: 'v999/wrong',
                            kind: 'WrongKind',
                            metadata: { name: 'bad-pvc' },
                            spec: validSpec
                        }
                    },
                    {
                        label: 'namespace in metadata (namespace is set by the WorkloadRunner)',
                        pvc: {
                            metadata: { name: 'bad-pvc', namespace: 'totally-different-namespace' },
                            spec: validSpec
                        }
                    },
                    {
                        label: 'unknown field inside spec',
                        pvc: { metadata: { name: 'bad-pvc' }, spec: { ...validSpec, bogusField: 'nope' } }
                    },
                    {
                        label: 'missing required spec.accessModes',
                        pvc: {
                            metadata: { name: 'bad-pvc' },
                            spec: {
                                resources: { requests: { storage: '1Gi' } },
                                storageClassName: STORAGE_CLASS
                            }
                        }
                    },
                    {
                        label: 'missing required spec.resources.requests.storage',
                        pvc: {
                            metadata: { name: 'bad-pvc' },
                            spec: {
                                accessModes: ['ReadWriteOnce'],
                                resources: { requests: {} },
                                storageClassName: STORAGE_CLASS
                            }
                        }
                    }
                ];

                for (const { label, pvc } of variants) {
                    await test.step(`V2 API rejects a newPvcs entry with ${label}`, async () => {
                        const name = objectGenerator.generateShortNameByPrefix('storage-neg', 5);
                        const { status, data } = await workloadsManagerUtils.tryCreateWorkloadV2(
                            { name, projectId: project.id.toString(), configuration: { pvcs: [pvc] } },
                            deploymentManifest(name, [])
                        );
                        expect(
                            status,
                            `should be a 4xx rejection for: ${label} (got ${status}, body: ${JSON.stringify(data)})`
                        ).toBeGreaterThanOrEqual(400);
                        expect(status, `should be a client error (<500) for: ${label}`).toBeLessThan(500);
                    });
                }
            });

            test('T6140998: a second workload declaring the same PVC name does not hijack the PVC owned by the first', async function () {
                const sharedPvc = objectGenerator.generateShortNameByPrefix('shared-conflict-pvc', 5);
                const nameA = objectGenerator.generateShortNameByPrefix('wl-owner-a', 5);
                const nameB = objectGenerator.generateShortNameByPrefix('wl-owner-b', 5);
                let idA: string | undefined;
                let idB: string | undefined;

                try {
                    await test.step('Workload A creates and owns the shared PVC', async () => {
                        idA = await workloadsManagerUtils.createWorkloadV2WithPvcs(
                            nameA,
                            project.id.toString(),
                            [pvcEntry(sharedPvc, '1Gi')],
                            deploymentManifest(nameA, [{ volName: 'shared-vol', claimName: sharedPvc }])
                        );
                        const pvc = await waitForManagedPvc(sharedPvc, project.namespace);
                        expect(
                            workloadRunnerController(pvc),
                            `${sharedPvc} should carry a controller ownerRef to its WorkloadRunner`
                        ).toBeDefined();
                        await findWorkloadRunner(nameA, project.namespace);
                    });

                    await test.step('Workload B in the same project declares the same PVC name (API accepts it)', async () => {
                        idB = await workloadsManagerUtils.createWorkloadV2WithPvcs(
                            nameB,
                            project.id.toString(),
                            [pvcEntry(sharedPvc, '1Gi')],
                            deploymentManifest(nameB, [{ volName: 'shared-vol', claimName: sharedPvc }])
                        );
                        // Wait until B's WorkloadRunner exists so its storage plugin has had a reconcile pass
                        // (the AdditionalResourcePlugin's Get + IsControlledBy ownership check has run).
                        await findWorkloadRunner(nameB, project.namespace);
                    });

                    await test.step('The shared PVC is NOT hijacked — controller ownerRef still points to workload A', async () => {
                        const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                            name: sharedPvc,
                            namespace: project.namespace
                        });
                        const wrOwners = (pvc.metadata?.ownerReferences ?? []).filter(
                            (ref) => ref.kind === 'WorkloadRunner'
                        );
                        expect(wrOwners.length, 'PVC should have exactly one WorkloadRunner owner').toBe(1);
                        expect(
                            wrOwners[0].name?.includes(nameA),
                            `PVC owner must still be workload A (${nameA}), not B (${nameB}) — the plugin refuses to take over a PVC owned by another WorkloadRunner`
                        ).toBe(true);
                    });
                } finally {
                    // Delete B first (owns nothing), then A (GCs the shared PVC). afterAll project-delete is the backstop.
                    for (const id of [idB, idA]) {
                        if (id) {
                            try {
                                await workloadsManagerUtils.deleteWorkloadV2(id);
                            } catch (e) {
                                logger.info(`owner-conflict cleanup skipped for ${id}: ${e.message}`);
                            }
                        }
                    }
                }
            });

            test('T6141000: updating a workload preserves its managed PVC (patched, not recreated)', async function () {
                const wlName = objectGenerator.generateShortNameByPrefix('wl-update-pvc', 5);
                const pvcName = objectGenerator.generateShortNameByPrefix('update-pvc', 5);

                await test.step('Submit the workload with one managed PVC', async () => {
                    workloadId = await workloadsManagerUtils.createWorkloadV2WithPvcs(
                        wlName,
                        project.id.toString(),
                        [pvcEntry(pvcName, '1Gi')],
                        deploymentManifest(wlName, [{ volName: 'vol', claimName: pvcName }])
                    );
                });

                let originalUid: string | undefined;
                await test.step('The managed PVC is created; capture its uid', async () => {
                    const pvc = await waitForManagedPvc(pvcName, project.namespace);
                    expect(
                        workloadRunnerController(pvc),
                        `${pvcName} should carry a controller ownerRef to its WorkloadRunner`
                    ).toBeDefined();
                    originalUid = pvc.metadata?.uid;
                    expect(originalUid, 'PVC uid should be captured').toBeTruthy();
                });

                await test.step('PUT-update the manifest (replicas 1→2); the update body carries no storage config', async () => {
                    // UpdateRequest.metadata has no configuration field, so an update cannot change pvcs; a
                    // manifest-only update re-drives the plugin via SSA re-apply and must NOT recreate the PVC.
                    const updatedManifest = deploymentManifest(wlName, [
                        { volName: 'vol', claimName: pvcName }
                    ]);
                    (updatedManifest.spec as Record<string, unknown>).replicas = 2;
                    await workloadsManagerUtils.updateWorkloadV2Metadata(workloadId!, updatedManifest);
                });

                await test.step('The PVC is the SAME object (uid unchanged) — re-applied via SSA, not recreated', async () => {
                    await workloadsManagerUtils.waitForWorkloadV2Synced(workloadId!);
                    const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                        name: pvcName,
                        namespace: project.namespace
                    });
                    expect(pvc.metadata?.uid, 'PVC must not be recreated on update (same uid)').toBe(
                        originalUid
                    );
                    expect(
                        workloadRunnerController(pvc),
                        `${pvcName} should carry a controller ownerRef to its WorkloadRunner`
                    ).toBeDefined();
                });

                await test.step('Deleting the workload garbage-collects the managed PVC', async () => {
                    await deleteWorkloadAndWaitForPvcGc(workloadId!, [pvcName], project.namespace);
                    workloadId = undefined;
                });
            });

            test('T6140997: rejects duplicate pvcs names cluster-side (API accepts, sync reports Failed)', async function () {
                // The API accepts duplicate pvcs names (no cross-entry uniqueness check); the cluster-side
                // ValidatorPlugin then rejects the WorkloadRunner, surfaced back as syncInfo.status=Failed.
                const wlName = objectGenerator.generateShortNameByPrefix('wl-dup-pvc', 5);
                const dupName = objectGenerator.generateShortNameByPrefix('dup-pvc', 5);

                await test.step('Submit a workload declaring two pvcs with the same name (API accepts it)', async () => {
                    workloadId = await workloadsManagerUtils.createWorkloadV2WithPvcs(
                        wlName,
                        project.id.toString(),
                        [pvcEntry(dupName, '1Gi'), pvcEntry(dupName, '1Gi')],
                        deploymentManifest(wlName, [{ volName: 'vol', claimName: dupName }])
                    );
                });

                await test.step('The cluster-side validator rejects it — sync reaches Failed citing the duplicate', async () => {
                    const terminal = await workloadsManagerUtils.waitForWorkloadV2SyncTerminal(workloadId!);
                    expect(
                        terminal.syncInfo?.status,
                        `duplicate-name workload should fail to sync (message: ${terminal.syncInfo?.message ?? 'n/a'})`
                    ).toBe(WorkloadV2SyncStatus.Failed);
                    expect(
                        terminal.syncInfo?.message ?? '',
                        'the sync failure should reference the duplicate PVC name / rule'
                    ).toMatch(new RegExp(`${dupName}|duplicate`, 'i'));
                });

                await test.step('No WorkloadRunner is persisted (admission rejected the create)', async () => {
                    const runners = (await kubeCli.getCrds(
                        crds.workloadRunnerCrd,
                        project.namespace
                    )) as WorkloadRunnerResource[];
                    expect(
                        runners.find((r) => r.metadata?.name?.includes(wlName)),
                        'WorkloadRunner must NOT be persisted when admission rejects the duplicate'
                    ).toBeUndefined();
                });
            });

            test('creates but does not bind a PVC whose StorageClass does not exist (left to k8s)', async function () {
                // The ValidatorPlugin does NOT check StorageClass existence (left to k8s), so a nonexistent
                // storageClassName is accepted: the WorkloadRunner persists and the PVC is created but Pending.
                const wlName = objectGenerator.generateShortNameByPrefix('wl-unbindable', 5);
                const pvcName = objectGenerator.generateShortNameByPrefix('unbindable-pvc', 5);

                await test.step('Submit a workload whose pvcs entry references a nonexistent StorageClass', async () => {
                    workloadId = await workloadsManagerUtils.createWorkloadV2WithPvcs(
                        wlName,
                        project.id.toString(),
                        [pvcEntry(pvcName, '1Gi', { storageClassName: 'nonexistent-sc' })],
                        deploymentManifest(wlName, [{ volName: 'vol', claimName: pvcName }])
                    );
                });

                await test.step('The workload still syncs (a nonexistent StorageClass is not a validation failure)', async () => {
                    // A nonexistent SC is applied, not rejected, so the apply succeeds even though it won't bind.
                    await workloadsManagerUtils.waitForWorkloadV2Synced(workloadId!);
                });

                await test.step('The WorkloadRunner is created (admission did NOT reject the nonexistent SC)', async () => {
                    await findWorkloadRunner(wlName, project.namespace);
                });

                await test.step('The plugin creates the PVC, owned by the WorkloadRunner, but it stays Pending', async () => {
                    const pvc = await waitForManagedPvc(pvcName, project.namespace);
                    expect(
                        workloadRunnerController(pvc),
                        `${pvcName} should carry a controller ownerRef to its WorkloadRunner`
                    ).toBeDefined();
                    await waitUtils.waitForCondition(
                        async () => {
                            const p = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                                name: pvcName,
                                namespace: project.namespace
                            });
                            return p.status?.phase === 'Pending';
                        },
                        `PVC ${pvcName} to be observed Pending (no provisioner for the missing StorageClass)`,
                        true,
                        60
                    );
                });

                await test.step('Deleting the workload garbage-collects even the Pending PVC', async () => {
                    await deleteWorkloadAndWaitForPvcGc(workloadId!, [pvcName], project.namespace);
                    workloadId = undefined;
                });
            });

            test('T6141001: creates each pvcs entry independently when one references a nonexistent StorageClass (worst-of-N)', async function () {
                // One healthy + one unbindable (nonexistent SC) entry: both PVCs are created and owned (one bad
                // entry doesn't block the others); we assert the bad one is independently created and Pending.
                const wlName = objectGenerator.generateShortNameByPrefix('wl-worst-of-n', 5);
                const healthyPvc = objectGenerator.generateShortNameByPrefix('healthy-pvc', 5);
                const badPvc = objectGenerator.generateShortNameByPrefix('bad-pvc', 5);

                await test.step('Submit a workload with one healthy + one unbindable pvcs entry', async () => {
                    workloadId = await workloadsManagerUtils.createWorkloadV2WithPvcs(
                        wlName,
                        project.id.toString(),
                        [
                            pvcEntry(healthyPvc, '1Gi'),
                            pvcEntry(badPvc, '1Gi', { storageClassName: 'nonexistent-sc' })
                        ],
                        deploymentManifest(wlName, [
                            { volName: 'vol-good', claimName: healthyPvc },
                            { volName: 'vol-bad', claimName: badPvc }
                        ])
                    );
                });

                await test.step('The workload syncs (a nonexistent StorageClass is not a validation failure)', async () => {
                    await workloadsManagerUtils.waitForWorkloadV2Synced(workloadId!);
                });

                await test.step('Both PVCs are created and owned by the WorkloadRunner; the bad one stays Pending', async () => {
                    const healthy = await waitForManagedPvc(healthyPvc, project.namespace);
                    const bad = await waitForManagedPvc(badPvc, project.namespace);
                    expect(
                        workloadRunnerController(healthy),
                        `${healthyPvc} should carry a controller ownerRef to its WorkloadRunner`
                    ).toBeDefined();
                    expect(
                        workloadRunnerController(bad),
                        `${badPvc} should carry a controller ownerRef to its WorkloadRunner`
                    ).toBeDefined();
                    await waitUtils.waitForCondition(
                        async () => {
                            const p = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                                name: badPvc,
                                namespace: project.namespace
                            });
                            return p.status?.phase === 'Pending';
                        },
                        `bad PVC ${badPvc} to be observed Pending`,
                        true,
                        60
                    );
                });

                await test.step('Deleting the workload garbage-collects both PVCs (including the Pending one)', async () => {
                    await deleteWorkloadAndWaitForPvcGc(workloadId!, [healthyPvc, badPvc], project.namespace);
                    workloadId = undefined;
                });
            });

            test('T6140994: creates a single ReadWriteMany PVC for a workload whose replicas share one volume', async function () {
                // Flagship distributed shape: ONE ReadWriteMany entry shared by every replica (one managed PVC,
                // one ownerRef, GC-once). Concurrent RWX mount is infra-gated (local-path RWO-only), so not asserted.
                const wlName = objectGenerator.generateShortNameByPrefix('wl-rwx', 5);
                const rwxPvc = objectGenerator.generateShortNameByPrefix('shared-rwx-pvc', 5);

                await test.step('Submit a 2-replica workload sharing one ReadWriteMany pvcs entry', async () => {
                    const manifest = deploymentManifest(wlName, [
                        { volName: 'shared-vol', claimName: rwxPvc }
                    ]);
                    (manifest.spec as Record<string, unknown>).replicas = 2;
                    workloadId = await workloadsManagerUtils.createWorkloadV2WithPvcs(
                        wlName,
                        project.id.toString(),
                        [pvcEntry(rwxPvc, '1Gi', { accessModes: ['ReadWriteMany'] })],
                        manifest
                    );
                });

                await test.step('The workload syncs and the WorkloadRunner declares exactly one PVC', async () => {
                    await workloadsManagerUtils.waitForWorkloadV2Synced(workloadId!);
                    const runner = await findWorkloadRunner(wlName, project.namespace);
                    const declared = runner.spec?.instructions?.storage?.newPvcs ?? [];
                    expect(declared.length, 'exactly one PVC declaration (not one per replica)').toBe(1);
                    expect(declared[0]?.metadata?.name).toBe(rwxPvc);
                });

                await test.step('Exactly one managed PVC is created, ReadWriteMany, owned once by the WorkloadRunner', async () => {
                    const pvc = await waitForManagedPvc(rwxPvc, project.namespace);
                    expect(
                        workloadRunnerController(pvc),
                        `${rwxPvc} should carry a controller ownerRef to its WorkloadRunner`
                    ).toBeDefined();
                    expect(pvc.spec?.accessModes, 'the managed PVC should be ReadWriteMany').toEqual([
                        'ReadWriteMany'
                    ]);
                    const wrOwners = (pvc.metadata?.ownerReferences ?? []).filter(
                        (ref) => ref.kind === 'WorkloadRunner'
                    );
                    expect(wrOwners.length, 'a single WorkloadRunner owner (not one per replica)').toBe(1);
                });

                await test.step('Deleting the workload garbage-collects the shared PVC exactly once', async () => {
                    await deleteWorkloadAndWaitForPvcGc(workloadId!, [rwxPvc], project.namespace);
                    workloadId = undefined;
                });
            });

            test('T6141002: applies a workload without waiting for a WaitForFirstConsumer PVC to bind', async function () {
                // Branch #23: apply must not block on PVC bind. With a WaitForFirstConsumer SC the PVC binds only
                // after its consumer is scheduled; a synced workload while the PVC is deferred is the signal.
                const sc = await kubeCli.StorageV1Api.readStorageClass({ name: STORAGE_CLASS });
                test.skip(
                    sc.volumeBindingMode !== 'WaitForFirstConsumer',
                    `requires a WaitForFirstConsumer StorageClass; '${STORAGE_CLASS}' is '${sc.volumeBindingMode}'`
                );

                const wlName = objectGenerator.generateShortNameByPrefix('wl-wfc', 5);
                const pvcName = objectGenerator.generateShortNameByPrefix('wfc-pvc', 5);

                await test.step('Submit a workload whose managed PVC uses the WFC StorageClass', async () => {
                    workloadId = await workloadsManagerUtils.createWorkloadV2WithPvcs(
                        wlName,
                        project.id.toString(),
                        [pvcEntry(pvcName, '1Gi')],
                        deploymentManifest(wlName, [{ volName: 'vol', claimName: pvcName }])
                    );
                });

                await test.step('The workload applies (syncs) and the WorkloadRunner + PVC are created without blocking on bind', async () => {
                    await workloadsManagerUtils.waitForWorkloadV2Synced(workloadId!);
                    await findWorkloadRunner(wlName, project.namespace);
                    // Under WFC the PVC exists but stays deferred until a consumer is scheduled — the apply
                    // did not wait for it to bind.
                    await waitForManagedPvc(pvcName, project.namespace);
                });

                await test.step('Once the consumer pod is scheduled, WFC binds the PVC and the pod runs', async () => {
                    const pod = await waitForRunningPod(wlName, project.namespace);
                    expect(pod.metadata?.name, 'a pod should be Running').toBeTruthy();
                    const pvc = await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({
                        name: pvcName,
                        namespace: project.namespace
                    });
                    expect(pvc.status?.phase, 'WFC PVC should bind once its consumer is scheduled').toBe(
                        'Bound'
                    );
                    const selectedNode = pvc.metadata?.annotations?.['volume.kubernetes.io/selected-node'];
                    logger.info(
                        `WFC PVC ${pvcName} bound; selected-node annotation: ${selectedNode ?? 'n/a'}`
                    );
                });

                await test.step('Deleting the workload garbage-collects the PVC', async () => {
                    await deleteWorkloadAndWaitForPvcGc(workloadId!, [pvcName], project.namespace);
                    workloadId = undefined;
                });
            });
        });
    }
);
