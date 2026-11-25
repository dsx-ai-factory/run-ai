/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
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
  while (true) {
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

export function useNodePoolSchedulingSave() {
  let busy = false;
  let poolKey = "";
  let savedPool: Nodepool | undefined;
  let poolSettingsSaved = false;
  let pendingReset: OrgUnitKind | "departmentWeights" | "projectWeights" | undefined;
  const completed = new Map<string, string>();
  return {
    get resetPending() {
      return pendingReset !== undefined;
    },
    get savedPool() {
      return savedPool;
    },
    get poolSettingsSaved() {
      return poolSettingsSaved;
    },
    async save(pool: Nodepool, draft: SchedulingDraft): Promise<Nodepool> {
      if (busy) throw new Error("A save is already in progress.");
      busy = true;
      poolSettingsSaved = false;
      try {
        const latest = savedPool?.id === pool.id ? savedPool : pool;
        const plan = buildSchedulingSave(latest, draft);
        const settings = { ...draft, departments: undefined };
        const key = JSON.stringify([pool.id, settings]);
        if (poolKey !== key) {
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
          if (switchingToCustomWeights(latest.schedulingCapabilities, updated.schedulingCapabilities))
            pendingReset = "departments";
          else if (
            savedPool?.id !== pool.id ||
            !updated.schedulingCapabilities?.adjustableWeights ||
            !updated.schedulingCapabilities.quota
          )
            pendingReset = undefined;
          savedPool = updated;
          poolKey = key;
          completed.clear();
        }
        poolSettingsSaved = true;
        if (pendingReset) {
          const request = {
            clusterId: pool.clusterId,
            nodePoolIds: [pool.id],
            action: ResourcesResetAction.WeightsToDeserved,
          };
          if (pendingReset === "departments") {
            await orgUnitService.resetDepartmentResources(request);
            pendingReset = "projects";
          }
          if (pendingReset === "projects") {
            await orgUnitService.resetProjectResources(request);
            pendingReset = "departmentWeights";
          }
          if (pendingReset === "departmentWeights") {
            await normalizeResetGpuWeights(pool, "departments");
            pendingReset = "projectWeights";
          }
          await normalizeResetGpuWeights(pool, "projects");
          pendingReset = undefined;
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
