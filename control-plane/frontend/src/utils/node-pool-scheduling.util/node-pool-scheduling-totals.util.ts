/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { IDepartmentSchedulingEdit } from "@/models/node-pool-scheduling.model";

export function effectiveDepartmentEdit(
  baseline: Record<string, IDepartmentSchedulingEdit>,
  edits: Record<string, IDepartmentSchedulingEdit>,
  id: string,
): IDepartmentSchedulingEdit {
  return { ...baseline[id], ...edits[id], id };
}
interface ICompletePeerTotal {
  rawWeightsSum?: number;
  automaticWeightsSum: number;
  finiteQuotaSum: number;
  unlimitedQuotaCount: number;
  deservedSum: number;
}
// Call only with the complete peer snapshot, never the filtered/visible page.
// Excluding the popup row lets its unapplied value complete an unknown total.
export function completePeerSchedulingTotals(
  rows: IDepartmentSchedulingEdit[],
  {
    rawWeightsAvailable = true,
    excludeId,
    requireIntegerWeights = false,
  }: {
    rawWeightsAvailable?: boolean;
    excludeId?: string;
    requireIntegerWeights?: boolean;
  } = {},
) {
  const emptyTotal = (): ICompletePeerTotal => ({
    rawWeightsSum: rawWeightsAvailable ? 0 : undefined,
    automaticWeightsSum: 0,
    finiteQuotaSum: 0,
    unlimitedQuotaCount: 0,
    deservedSum: 0,
  });
  const totals: Record<string, ICompletePeerTotal> = { all: emptyTotal() };
  for (const row of rows) {
    for (const key of new Set(["all", row.rank ?? ""])) {
      const total = totals[key] ?? (totals[key] = emptyTotal());
      if (row.id === excludeId) continue;
      const weight = row.weight;
      if (
        weight == null ||
        !Number.isFinite(weight) ||
        weight < 0 ||
        (requireIntegerWeights && !Number.isInteger(weight))
      )
        total.rawWeightsSum = undefined;
      else if (total.rawWeightsSum !== undefined) total.rawWeightsSum += weight;
      const quota = row.quota;
      if (quota != null && Number.isFinite(quota) && quota >= 0) {
        total.automaticWeightsSum += quota;
        total.finiteQuotaSum += quota;
      } else total.unlimitedQuotaCount++;
      total.deservedSum = total.unlimitedQuotaCount ? -1 : total.finiteQuotaSum;
    }
  }
  return totals;
}
