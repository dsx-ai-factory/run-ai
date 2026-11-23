/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { isSupportedSchedulingMode } from "@/models/node-pool.model";
import { ENodePoolSchedulingMode as ESchedulingMode } from "@/models/node-pool.model";
import type { Nodepool, NodepoolUpdateFields } from "@/swagger-models/cluster-service-client/index";
import type { NodePoolResourcesPatchRequest } from "@/swagger-models/org-unit-service-client/index";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";
import {
  toSchedulingCapabilities,
  validateDepartmentValues,
  type SchedulingDraft,
} from "@/models/node-pool-scheduling.model";
import {
  schedulingDurationSeconds,
  serializeSchedulingDuration,
} from "@/utils/node-pool-scheduling.util/scheduling-duration.util";

export function buildSchedulingSave(pool: Nodepool, draft: SchedulingDraft) {
  if (!isSupportedSchedulingMode(draft.mode)) throw new Error("Select a supported scheduling mode.");
  if (pool.schedulingCapabilities?.quota && draft.mode !== ESchedulingMode.Guaranteed)
    throw new Error("A saved Guaranteed quota pool cannot downgrade.");
  const capabilities = toSchedulingCapabilities(draft);
  const minimum = serializeSchedulingDuration(draft.minimumRuntime);
  const window = draft.timeFairshare ? serializeSchedulingDuration(draft.historicalWindow, false) : undefined;
  const savedWindow = pool.schedulingConfiguration?.timeBasedFairShare?.windowDuration;
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
      schedulingDurationSeconds(minimum) !==
        schedulingDurationSeconds(pool.schedulingConfiguration?.minGuaranteedRuntime)
        ? { minGuaranteedRuntime: minimum }
        : {}),
      ...(draft.timeFairshare !== undefined
        ? {
            timeBasedFairShare: {
              ...(draft.timeFairshare ? pool.schedulingConfiguration?.timeBasedFairShare : {}),
              enabled: draft.timeFairshare,
              ...(window !== undefined &&
              (savedWindow == null || schedulingDurationSeconds(window) !== schedulingDurationSeconds(savedWindow))
                ? { windowDuration: window }
                : {}),
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
    const active = {
      id: edit.id,
      ...(capabilities.adjustableWeights ? { weight: edit.weight } : {}),
      ...(capabilities.quota ? { quota: edit.quota } : {}),
      ...(capabilities.limits ? { limit: edit.limit } : {}),
    };
    const errors = validateDepartmentValues(active);
    if (errors.length) throw new Error(errors[0]);
    if (active.weight !== undefined) gpu.overQuotaWeight = active.weight;
    if (active.quota !== undefined) gpu.deserved = active.quota;
    // Use the explicit unlimited sentinel. Null is omitted by the Go patch model.
    if (active.limit !== undefined) gpu.limit = active.limit ?? -1;
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
