/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import type { Nodepool } from "@/swagger-models/cluster-service-client/index";
import type { SchedulingDraft } from "@/models/node-pool-scheduling.model";
import { buildSchedulingSave } from "@/utils/node-pool-scheduling.util/node-pool-scheduling-save.util";
export function useNodePoolSchedulingSave() {
  let busy = false;
  let poolKey = "";
  let savedPool: Nodepool | undefined;
  let poolSettingsSaved = false;
  const completed = new Map<string, string>();
  return {
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
          savedPool = updated;
          poolKey = key;
          completed.clear();
        }
        poolSettingsSaved = true;
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
