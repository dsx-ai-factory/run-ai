/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { SchedulingCapabilitiesNullable } from "@/swagger-models/cluster-service-client";

export type SchedulingMode = "proportional" | "ranked" | "guaranteed";
export interface DepartmentSchedulingEdit {
  id: string;
  weight?: number;
  quota?: number;
  limit?: number | null;
  rank?: string;
}
export interface SchedulingSnapshot {
  mode?: SchedulingMode;
  automaticWeights: boolean;
  cpuQuota: boolean;
  departments: Record<string, DepartmentSchedulingEdit>;
  capabilities?: SchedulingCapabilitiesNullable;
}
export type SchedulingDraft = SchedulingSnapshot;
export function createSchedulingDraft(snapshot: SchedulingSnapshot): SchedulingDraft {
  return JSON.parse(JSON.stringify(snapshot));
}
export function toSchedulingCapabilities(draft: SchedulingDraft): SchedulingCapabilitiesNullable {
  if (!draft.mode) return { ...draft.capabilities };
  return {
    adjustableWeights: draft.mode !== "guaranteed" || !draft.automaticWeights,
    quota: draft.mode === "guaranteed",
    ranks: draft.mode !== "proportional",
    cpuQuota: draft.mode === "guaranteed" && draft.cpuQuota,
  };
}

export interface DepartmentValidationContext {
  childQuotaSum?: number;
  enforceChildQuota?: boolean;
  capacity?: number;
  totalQuota?: number;
}
export function weightPercentage(weight: number, total: number | undefined): number | undefined {
  if (total == null || !Number.isFinite(total) || total < 0 || !Number.isFinite(weight)) return undefined;
  return total === 0 ? 0 : (weight / total) * 100;
}
export function validateDepartmentEdit(edit: DepartmentSchedulingEdit, context: DepartmentValidationContext) {
  const errors: string[] = [];
  const warnings: string[] = [];
  if (
    edit.weight !== undefined &&
    (typeof edit.weight !== "number" ||
      !Number.isFinite(edit.weight) ||
      edit.weight < 0 ||
      !Number.isInteger(edit.weight))
  )
    errors.push("Weight must be a whole number of 0 or greater.");
  if (edit.quota !== undefined) {
    if (typeof edit.quota !== "number" || !Number.isFinite(edit.quota) || edit.quota < 0)
      errors.push("Guaranteed GPUs must be 0 or greater.");
    if (context.childQuotaSum === undefined) errors.push("Project quota totals are unavailable.");
    else if (edit.quota < context.childQuotaSum)
      (context.enforceChildQuota ? errors : warnings).push("Guaranteed GPUs are below the child project quota sum.");
  }
  if (edit.limit != null && (!Number.isFinite(edit.limit) || edit.limit < 0))
    errors.push("Maximum allocation must be 0 or greater.");
  if (edit.limit != null && edit.quota != null && edit.limit < edit.quota)
    errors.push("Maximum allocation must be at least guaranteed GPUs.");
  if (context.capacity != null && context.totalQuota != null && context.totalQuota > context.capacity)
    warnings.push("Assigned quota exceeds node pool capacity.");
  return { errors, warnings };
}
export interface SchedulingTotals {
  rawWeightsSum?: number;
  automaticWeightsSum?: number;
  weightsSum: number;
  deservedSum: number;
}
export function schedulingTotalsWithEdits(
  server: Record<string, SchedulingTotals>,
  baseline: Record<string, DepartmentSchedulingEdit>,
  edits: Record<string, DepartmentSchedulingEdit>,
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
      if (original.quota != null && merged.quota != null)
        total.deservedSum += (after ? merged.quota : 0) - (before ? original.quota : 0);
    }
  }
  return totals;
}
