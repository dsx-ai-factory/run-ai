/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ENodePoolSchedulingMode as ESchedulingMode } from "@/models/node-pool.model";
import { effectiveDepartmentEdit } from "@/utils/node-pool-scheduling.util/node-pool-scheduling-totals.util";
import type { SchedulingCapabilitiesNullable } from "@/swagger-models/cluster-service-client/index";

import { PROPORTIONAL_CAPABILITIES, isSupportedSchedulingMode } from "@/models/node-pool.model";
export interface IDepartmentSchedulingEdit {
  id: string;
  weight?: number;
  quota?: number;
  limit?: number | null;
  rank?: string;
}
export interface ISchedulingSnapshot {
  mode?: ESchedulingMode;
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
  if (!isSupportedSchedulingMode(draft.mode)) return { ...draft.capabilities };
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
export function isValidSchedulingWeight(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= 0 && value <= 20;
}
export function isFiniteSchedulingAllocation(value: unknown): value is number {
  return typeof value === "number" && Number.isFinite(value) && value >= 0;
}
export function isDepartmentLimitBelowQuota(
  limit: number | null | undefined,
  quota: number | null | undefined,
): boolean {
  return limit != null && limit !== -1 && quota != null && limit < quota;
}
// Primitive field checks are shared by editors, saves and simulation. Project
// context is requested separately, only when the quota itself is being edited.
export function validateDepartmentValues(edit: IDepartmentSchedulingEdit) {
  const errors: string[] = [];
  if (edit.weight !== undefined && !isValidSchedulingWeight(edit.weight))
    errors.push("Weight must be a whole number from 0 to 20.");
  if (edit.quota !== undefined && !isFiniteSchedulingAllocation(edit.quota))
    errors.push("Guaranteed GPUs must be 0 or greater.");
  const finiteLimit = edit.limit === -1 ? undefined : edit.limit;
  if (finiteLimit != null && !isFiniteSchedulingAllocation(finiteLimit))
    errors.push("Maximum allocation must be unlimited or 0 or greater.");
  if (isDepartmentLimitBelowQuota(edit.limit, edit.quota))
    errors.push("Maximum allocation must be at least guaranteed GPUs.");
  return errors;
}
export function validateDepartmentEdit(edit: IDepartmentSchedulingEdit, context: IDepartmentValidationContext) {
  const errors = validateDepartmentValues(edit);
  const warnings: string[] = [];
  if (edit.quota !== undefined) {
    if (context.childQuotaSum === undefined) errors.push("Project quota totals are unavailable.");
    else if (edit.quota < context.childQuotaSum)
      (context.enforceChildQuota ? errors : warnings).push("Guaranteed GPUs are below the child project quota sum.");
  }
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
  for (const id of Object.keys(edits)) {
    const original = baseline[id];
    if (!original) continue;
    const merged = effectiveDepartmentEdit(baseline, edits, id);
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

export interface ISchedulingScope {
  poolId: string;
  clusterId: string;
  rank?: string;
}
