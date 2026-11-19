/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ENodePoolSchedulingModeId as ESchedulingMode } from "@/models/node-pool.model";
import type { SchedulingCapabilitiesNullable } from "@/swagger-models/cluster-service-client";

import { PROPORTIONAL_CAPABILITIES, type ENodePoolSchedulingModeId } from "@/models/node-pool.model";
export type SchedulingMode = ENodePoolSchedulingModeId;
export interface IDepartmentSchedulingEdit {
  id: string;
  weight?: number;
  quota?: number;
  limit?: number | null;
  rank?: string;
}
export interface ISchedulingSnapshot {
  mode?: SchedulingMode;
  automaticWeights: boolean;
  cpuQuota: boolean;
  maxAllocation?: boolean;
  timeFairshare?: boolean;
  historicalWeight?: number | null;
  historicalWindow?: Record<string, number | null>;
  minimumRuntime?: Record<string, number | null>;
  departments: Record<string, IDepartmentSchedulingEdit>;
  capabilities?: SchedulingCapabilitiesNullable;
}
export type SchedulingDraft = ISchedulingSnapshot;
export function createSchedulingDraft(snapshot: ISchedulingSnapshot): SchedulingDraft {
  return JSON.parse(JSON.stringify(snapshot));
}
export function toSchedulingCapabilities(draft: SchedulingDraft): SchedulingCapabilitiesNullable {
  if (!draft.mode) return { ...draft.capabilities };
  if (draft.mode === ESchedulingMode.Proportional) return { ...PROPORTIONAL_CAPABILITIES };
  return {
    adjustableWeights: draft.mode !== ESchedulingMode.Guaranteed || !draft.automaticWeights,
    quota: draft.mode === ESchedulingMode.Guaranteed,
    ranks: true,
    limits: draft.mode === ESchedulingMode.Guaranteed && !!draft.maxAllocation,
    cpuQuota: draft.mode === ESchedulingMode.Guaranteed && draft.cpuQuota,
  };
}

export interface IDepartmentValidationContext {
  totalsComplete?: boolean;
  childQuotaSum?: number;
  nonPreemptibleUsage?: number;
  enforceChildQuota?: boolean;
  capacity?: number;
  totalQuota?: number;
  finiteQuotaSum?: number;
  unlimitedQuotaCount?: number;
}
export function recommendedDepartmentQuotaMinimum(context: IDepartmentValidationContext) {
  const known = [context.childQuotaSum, context.nonPreemptibleUsage].filter(
    (value): value is number => value != null && Number.isFinite(value) && value >= 0,
  );
  return known.length ? Math.max(...known) : undefined;
}
export function departmentQuotaBudget(originalQuota: number | undefined, context: IDepartmentValidationContext) {
  const { capacity, finiteQuotaSum, unlimitedQuotaCount } = context;
  if (
    context.totalsComplete === false ||
    capacity == null ||
    !Number.isFinite(capacity) ||
    capacity < 0 ||
    finiteQuotaSum == null ||
    !Number.isFinite(finiteQuotaSum) ||
    finiteQuotaSum < 0 ||
    unlimitedQuotaCount == null ||
    !Number.isInteger(unlimitedQuotaCount) ||
    unlimitedQuotaCount < 0 ||
    originalQuota == null ||
    !Number.isFinite(originalQuota) ||
    (originalQuota < 0 && originalQuota !== -1)
  )
    return undefined;
  const otherUnlimited = unlimitedQuotaCount - (originalQuota === -1 ? 1 : 0);
  const otherFiniteQuota = finiteQuotaSum - (originalQuota >= 0 ? originalQuota : 0);
  if (otherUnlimited !== 0 || otherFiniteQuota < 0) return undefined;
  return { available: Math.max(0, capacity - otherFiniteQuota), otherFiniteQuota };
}
export function weightPercentage(weight: number, total: number | undefined): number | undefined {
  if (total == null || !Number.isFinite(total) || total < 0 || !Number.isFinite(weight)) return undefined;
  return total === 0 ? 0 : (weight / total) * 100;
}
export function validateDepartmentEdit(edit: IDepartmentSchedulingEdit, context: IDepartmentValidationContext) {
  const errors: string[] = [];
  const warnings: string[] = [];
  if (
    edit.weight !== undefined &&
    (typeof edit.weight !== "number" ||
      !Number.isFinite(edit.weight) ||
      edit.weight < 0 ||
      edit.weight > 20 ||
      !Number.isInteger(edit.weight))
  )
    errors.push("Weight must be a whole number from 0 to 20.");
  if (edit.quota !== undefined) {
    if (typeof edit.quota !== "number" || !Number.isFinite(edit.quota) || edit.quota < 0)
      errors.push("Guaranteed GPUs must be 0 or greater.");
    if (context.childQuotaSum === undefined) errors.push("Project quota totals are unavailable.");
    else if (edit.quota < context.childQuotaSum)
      (context.enforceChildQuota ? errors : warnings).push("Guaranteed GPUs are below the child project quota sum.");
  }
  const finiteLimit = edit.limit === -1 ? undefined : edit.limit;
  if (finiteLimit != null && (!Number.isFinite(finiteLimit) || finiteLimit < 0))
    errors.push("Maximum allocation must be 0 or greater.");
  if (finiteLimit != null && edit.quota != null && finiteLimit < edit.quota)
    errors.push("Maximum allocation must be at least guaranteed GPUs.");
  if (
    context.capacity != null &&
    ((context.totalsComplete !== false && context.totalQuota != null && context.totalQuota > context.capacity) ||
      (edit.quota != null && edit.quota > context.capacity))
  )
    warnings.push("Assigned quota exceeds node pool capacity.");
  return { errors, warnings };
}
export interface ISchedulingTotals {
  rawWeightsSum?: number;
  automaticWeightsSum?: number;
  weightsSum: number;
  deservedSum: number;
  finiteQuotaSum?: number;
  unlimitedQuotaCount?: number;
}
export function schedulingTotalsWithEdits(
  server: Record<string, ISchedulingTotals>,
  baseline: Record<string, IDepartmentSchedulingEdit>,
  edits: Record<string, IDepartmentSchedulingEdit>,
  automatic: boolean,
) {
  const totals = Object.fromEntries(
    Object.entries(server).map(([key, value]) => [
      key,
      {
        ...value,
        weightsSum: automatic
          ? (value.automaticWeightsSum ?? value.deservedSum)
          : "rawWeightsSum" in value
            ? value.rawWeightsSum
            : value.weightsSum,
      },
    ]),
  );
  for (const [id, edit] of Object.entries(edits)) {
    const original = baseline[id];
    if (!original) continue;
    const merged = { ...original, ...edit };
    const beforeWeight = automatic ? Math.max(0, original.quota ?? 0) : original.weight;
    const afterWeight = automatic ? Math.max(0, merged.quota ?? 0) : merged.weight;
    for (const [key, total] of Object.entries(totals)) {
      const before = key === "all" || key === original.rank;
      const after = key === "all" || key === merged.rank;
      if (beforeWeight != null && afterWeight != null && total.weightsSum != null)
        total.weightsSum += (after ? afterWeight : 0) - (before ? beforeWeight : 0);
      if (total.finiteQuotaSum != null && total.unlimitedQuotaCount != null) {
        const beforeFinite = original.quota != null && original.quota >= 0;
        const afterFinite = merged.quota != null && merged.quota >= 0;
        total.finiteQuotaSum +=
          (after && afterFinite ? (merged.quota ?? 0) : 0) - (before && beforeFinite ? (original.quota ?? 0) : 0);
        total.unlimitedQuotaCount += (after && !afterFinite ? 1 : 0) - (before && !beforeFinite ? 1 : 0);
        total.deservedSum = total.unlimitedQuotaCount > 0 ? -1 : total.finiteQuotaSum;
      } else if (
        total.deservedSum >= 0 &&
        original.quota != null &&
        original.quota >= 0 &&
        merged.quota != null &&
        merged.quota >= 0
      ) {
        total.deservedSum += (after ? merged.quota : 0) - (before ? original.quota : 0);
      }
    }
  }
  return totals;
}

export enum EResourceEditorKind {
  Weight = "weight",
  Quota = "quota",
  Limit = "limit",
}
