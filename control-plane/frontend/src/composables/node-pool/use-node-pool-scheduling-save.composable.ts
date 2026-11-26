/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ref } from "vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { ResourcesResetAction } from "@/swagger-models/org-unit-service-client";
import type { Nodepool, SchedulingCapabilitiesNullable } from "@/swagger-models/cluster-service-client/index";
import { weightFromGuaranteedQuota, type SchedulingDraft } from "@/models/node-pool-scheduling.model";
import { buildSchedulingSave } from "@/utils/node-pool-scheduling.util/node-pool-scheduling-save.util";
export function switchingToCustomWeights(saved?: SchedulingCapabilitiesNullable, next?: SchedulingCapabilitiesNullable) {
  return (
    saved?.quota === true && saved.adjustableWeights === false && next?.quota === true && next.adjustableWeights === true
  );
}

type OrgUnitKind = "departments" | "projects";

async function normalizeResetGpuWeights(pool: Nodepool, kind: OrgUnitKind) {
  const list =
    kind === "departments" ? orgUnitService.getDepartmentNodePoolResources : orgUnitService.getProjectNodePoolResources;
  const patch =
    kind === "departments"
      ? orgUnitService.batchPatchDepartmentNodePoolResources
      : orgUnitService.batchPatchProjectNodePoolResources;
  const filters = [`clusterId==${pool.clusterId}`, `nodePoolId==${pool.id}`];
  let offset = 0;
  for (;;) {
    // Sort by a stable ID; updating weights must not move rows between pages.
    const page = await list(filters, "orgUnitId", "asc", offset, 100);
    const patchRequests = page.nodePoolResources.flatMap(({ orgUnitId, nodePoolId, resources }) => {
      const { deserved, overQuotaWeight } = resources.gpu ?? {};
      // The existing reset copies quota literally. Only repair the two sentinels it copied.
      if ((deserved !== -1 && deserved !== 0) || overQuotaWeight !== deserved) return [];
      return [{ orgUnitId, nodePoolId, resources: { gpu: { overQuotaWeight: weightFromGuaranteedQuota(deserved) } } }];
    });
    if (patchRequests.length) {
      const result = await patch({ patchRequests });
      if (result.totalFailed || result.totalSuccessful !== patchRequests.length) {
        throw new Error(result.failed?.[0]?.message || `Could not initialize weights for all ${kind}.`);
      }
    }
    if (page.next == null) return;
    if (page.next <= offset) throw new Error(`Could not load the next page of ${kind}.`);
    offset = page.next;
  }
}

const resetPhases = ["departments", "projects", "departmentWeights", "projectWeights"] as const;
type ResetPhase = (typeof resetPhases)[number];

export function useNodePoolSchedulingSave(getPool: () => Nodepool) {
  // The scheduling step is mounted per pool. Keep its recovery record across reloads in this tab.
  const { clusterId, id } = getPool();
  const resetKey = `node-pool-weight-reset:${clusterId}:${id}`;
  const storedPhase = sessionStorage.getItem(resetKey);
  const pendingReset = ref(resetPhases.find((phase) => phase === storedPhase));
  function setPendingReset(phase?: ResetPhase) {
    if (phase) sessionStorage.setItem(resetKey, phase);
    else sessionStorage.removeItem(resetKey);
    pendingReset.value = phase;
  }

  let busy = false;
  let poolKey = "";
  let savedPool: Nodepool | undefined;
  let poolSettingsSaved = false;
  const completed = new Map<string, string>();
  return {
    get resetPending() {
      return pendingReset.value !== undefined;
    },
    get savedPool() {
      return savedPool;
    },
    get poolSettingsSaved() {
      return poolSettingsSaved;
    },
    clearSaveProgress() {
      poolKey = "";
      savedPool = undefined;
      poolSettingsSaved = false;
      completed.clear();
      // An unfinished server-side weight reset must still resume on the next save.
    },
    async save(draft: SchedulingDraft): Promise<Nodepool> {
      if (busy) throw new Error("A save is already in progress.");
      busy = true;
      poolSettingsSaved = false;
      try {
        const pool = getPool();
        const latest = savedPool ?? pool;
        const plan = buildSchedulingSave(latest, draft);
        const settings = { ...draft, departments: undefined };
        const key = JSON.stringify([pool.id, settings]);
        if (poolKey !== key) {
          poolKey = "";
          // A reload can interrupt the response after the server has already enabled custom weights.
          if (switchingToCustomWeights(latest.schedulingCapabilities, plan.pool.schedulingCapabilities))
            setPendingReset("departments");
          const updated = await clusterService.patchNodepool(pool.id, plan.pool);
          if (
            Object.entries(plan.pool.schedulingCapabilities ?? {}).some(
              ([name, value]) =>
                updated.schedulingCapabilities?.[name as keyof NonNullable<typeof updated.schedulingCapabilities>] !==
                value,
            )
          )
            throw new Error(
              "The backend did not apply the requested scheduling capabilities. Department changes were not sent.",
            );
          if (
            pendingReset.value &&
            !(updated.schedulingCapabilities?.quota && updated.schedulingCapabilities.adjustableWeights)
          )
            setPendingReset();
          savedPool = updated;
          poolKey = key;
          completed.clear();
        }
        poolSettingsSaved = true;
        if (pendingReset.value) {
          const request = {
            clusterId: pool.clusterId,
            nodePoolIds: [pool.id],
            action: ResourcesResetAction.WeightsToDeserved,
          };
          if (pendingReset.value === "departments") {
            await orgUnitService.resetDepartmentResources(request);
            setPendingReset("projects");
          }
          if (pendingReset.value === "projects") {
            await orgUnitService.resetProjectResources(request);
            setPendingReset("departmentWeights");
          }
          if (pendingReset.value === "departmentWeights") {
            await normalizeResetGpuWeights(pool, "departments");
            setPendingReset("projectWeights");
          }
          await normalizeResetGpuWeights(pool, "projects");
          setPendingReset();
        }
        const pending = plan.departments.filter((row) => completed.get(row.orgUnitId) !== JSON.stringify(row));
        for (let start = 0; start < pending.length; start += 100) {
          const batch = pending.slice(start, start + 100);
          const result = await orgUnitService.batchPatchDepartmentNodePoolResources({ patchRequests: batch });
          const succeeded = new Set(
            (result.succeeded ?? []).filter((row) => row.nodePoolId === pool.id).map((row) => row.orgUnitId),
          );
          for (const row of batch) if (succeeded.has(row.orgUnitId)) completed.set(row.orgUnitId, JSON.stringify(row));
          if (result.totalFailed || batch.some((row) => !succeeded.has(row.orgUnitId)))
            throw new Error(
              result.failed
                ?.map((row) => row.message)
                .filter(Boolean)
                .join("; ") || "Some department changes were not confirmed. Retry to save the remaining changes.",
            );
        }
        if (!savedPool) throw new Error("The pool update was not confirmed.");
        return savedPool;
      } finally {
        busy = false;
      }
    },
  };
}
