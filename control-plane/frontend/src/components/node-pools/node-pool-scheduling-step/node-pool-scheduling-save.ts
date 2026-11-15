/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { Nodepool, NodepoolUpdateFields } from "@/swagger-models/cluster-service-client";
import type {
  BatchPatchNodePoolResourcesRequest,
  BatchPatchDepartmentResourcesResponse,
  NodePoolResourcesPatchRequest,
} from "@/swagger-models/org-unit-service-client";
import { dateUtil } from "@/utils/date.util/date.util";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";
import { toSchedulingCapabilities, type SchedulingDraft } from "./node-pool-scheduling.model";

function duration(fields: Record<string, number | null> | undefined): string | undefined {
  if (!fields) return undefined;
  const seconds = Object.entries({ Days: 86400, Hours: 3600, Minutes: 60, Seconds: 1 }).reduce(
    (total, [unit, multiplier]) => {
      const value = fields[unit] ?? 0;
      if (!Number.isFinite(value) || value < 0) throw new Error("Enter a valid nonnegative duration.");
      return total + value * multiplier;
    },
    0,
  );
  if (!Number.isFinite(seconds) || seconds > Number.MAX_SAFE_INTEGER) throw new Error("The duration is too large.");
  return `${seconds}s`;
}

export function buildSchedulingSave(pool: Nodepool, draft: SchedulingDraft) {
  if (!draft.mode) throw new Error("Select a supported scheduling mode.");
  if (pool.schedulingCapabilities?.quota && draft.mode !== "guaranteed")
    throw new Error("A saved Guaranteed quota pool cannot downgrade.");
  if (draft.maxAllocation && draft.mode !== "guaranteed")
    throw new Error("Max GPU allocation in this mode is preview-only. Turn it off before saving.");
  const capabilities = toSchedulingCapabilities(draft);
  const minimum = duration(draft.minimumRuntime);
  const window = draft.timeFairshare ? duration(draft.historicalWindow) : undefined;
  if (
    draft.timeFairshare &&
    draft.historicalWeight != null &&
    (!Number.isFinite(draft.historicalWeight) || draft.historicalWeight < 0)
  )
    throw new Error("Historical usage weight must be 0 or greater.");
  const poolUpdate: NodepoolUpdateFields = {
    schedulingCapabilities: capabilities,
    schedulingConfiguration: {
      ...(minimum !== undefined &&
      dateUtil.parseDurationStringToSeconds(minimum) !==
        dateUtil.parseDurationStringToSeconds(pool.schedulingConfiguration?.minGuaranteedRuntime ?? "")
        ? { minGuaranteedRuntime: minimum }
        : {}),
      ...(draft.timeFairshare !== undefined
        ? {
            timeBasedFairShare: {
              ...(draft.timeFairshare ? pool.schedulingConfiguration?.timeBasedFairShare : {}),
              enabled: draft.timeFairshare,
              ...(window !== undefined ? { windowDuration: window } : {}),
              ...(draft.timeFairshare && draft.historicalWeight != null
                ? { historicalUsageWeight: draft.historicalWeight }
                : {}),
            },
          }
        : {}),
    },
  };
  const departments: NodePoolResourcesPatchRequest[] = [];
  for (const edit of Object.values(draft.departments)) {
    const gpu: NonNullable<NonNullable<NodePoolResourcesPatchRequest["resources"]>["gpu"]> = {};
    if (capabilities.adjustableWeights && edit.weight !== undefined) {
      if (!Number.isInteger(edit.weight) || edit.weight < 0 || edit.weight > 20)
        throw new Error("Weight must be a whole number from 0 to 20.");
      gpu.overQuotaWeight = edit.weight;
    }
    if (capabilities.quota && edit.quota !== undefined) {
      if (!Number.isFinite(edit.quota) || edit.quota < 0) throw new Error("Guaranteed GPUs must be 0 or greater.");
      gpu.deserved = edit.quota;
    }
    if (capabilities.limits && edit.limit !== undefined) {
      if (edit.limit != null && (!Number.isFinite(edit.limit) || (edit.limit < 0 && edit.limit !== -1)))
        throw new Error("Maximum allocation must be unlimited or 0 or greater.");
      if (edit.limit != null && edit.limit !== -1 && gpu.deserved != null && edit.limit < gpu.deserved)
        throw new Error("Maximum allocation must be at least guaranteed GPUs.");
      // Use the explicit unlimited sentinel. Null is omitted by the Go patch model.
      gpu.limit = edit.limit ?? -1;
    }
    const rank = capabilities.ranks ? edit.rank : undefined;
    if (rank !== undefined && !ENTITY_PRIORITY_SELECT_OPTIONS.some((option) => option.value === rank))
      throw new Error("Select a valid rank.");
    if (Object.keys(gpu).length || rank !== undefined)
      departments.push({
        orgUnitId: edit.id,
        nodePoolId: pool.id,
        ...(rank !== undefined ? { rank } : {}),
        resources: { gpu },
      });
  }
  return { pool: poolUpdate, departments };
}
export interface SchedulingSaveTransport {
  patchPool: (id: string, payload: NodepoolUpdateFields) => Promise<Nodepool>;
  patchDepartments: (request: BatchPatchNodePoolResourcesRequest) => Promise<BatchPatchDepartmentResourcesResponse>;
}
export function createSchedulingSaver(transport: SchedulingSaveTransport) {
  let busy = false;
  let poolKey = "";
  let savedPool: Nodepool | undefined;
  const completed = new Map<string, string>();
  return {
    get savedPool() {
      return savedPool;
    },
    async save(pool: Nodepool, draft: SchedulingDraft): Promise<Nodepool> {
      if (busy) throw new Error("A save is already in progress.");
      busy = true;
      try {
        const latest = savedPool?.id === pool.id ? savedPool : pool;
        const plan = buildSchedulingSave(latest, draft);
        const settings = { ...draft, departments: undefined };
        const key = JSON.stringify([pool.id, settings]);
        if (poolKey !== key) {
          const updated = await transport.patchPool(pool.id, plan.pool);
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
        const pending = plan.departments.filter((row) => completed.get(row.orgUnitId) !== JSON.stringify(row));
        for (let start = 0; start < pending.length; start += 100) {
          const batch = pending.slice(start, start + 100);
          const result = await transport.patchDepartments({ patchRequests: batch });
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
