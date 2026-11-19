/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { QTableColumn } from "quasar";
import { ENodePoolSchedulingMode as ESchedulingMode } from "@/models/node-pool.model";
import { entityPriorityMap } from "@/models/org-unit.model";
import { weightPercentage, type IDepartmentSchedulingEdit } from "@/models/node-pool-scheduling.model";
import { completePeerSchedulingTotals } from "@/utils/node-pool-scheduling.util/node-pool-scheduling-totals.util";

export function departmentTableColumns(mode: ESchedulingMode | undefined, maxAllocation: boolean): QTableColumn[] {
  const ranked = mode === ESchedulingMode.Ranked || mode === ESchedulingMode.Guaranteed;
  return [
    { name: "selection", label: "", field: "id", align: "left" },
    ...(mode === ESchedulingMode.Guaranteed
      ? [{ name: "quota", label: "Guaranteed GPUs", field: "quota", align: "left" as const }]
      : []),
    { name: "department", label: ranked ? "Rank" : "Department", field: "name", align: "left" },
    { name: "weight", label: ranked ? "Weight within rank" : "Weight", field: "weight", align: "left" },
    ...(maxAllocation ? [{ name: "limit", label: "Max GPU allocation", field: "limit", align: "left" as const }] : []),
    { name: "result", label: "Result", field: "id", align: "left" },
  ];
}
export function schedulingColumnOrder(name: string, mode: ESchedulingMode | undefined) {
  const ranked = mode === ESchedulingMode.Ranked || mode === ESchedulingMode.Guaranteed;
  if (name === "quota") return "#1";
  if (name === "department") return ranked ? (mode === ESchedulingMode.Guaranteed ? "#2" : "#1") : "";
  if (name === "weight") return mode === ESchedulingMode.Guaranteed ? "#3" : ranked ? "#2" : "#1";
  if (name === "limit") return mode === ESchedulingMode.Guaranteed ? "#4" : ranked ? "#3" : "#2";
  return "";
}
export function rankedDepartmentRows<T extends IDepartmentSchedulingEdit>(
  effectiveRows: T[],
  options: {
    ranks: string[];
    expanded: Set<string>;
    rankCounts?: Record<string, number>;
    filtered: boolean;
    rankFilter?: string;
    baseline: Record<string, IDepartmentSchedulingEdit>;
    edits: Record<string, IDepartmentSchedulingEdit>;
  },
) {
  const { ranks, expanded, rankCounts, filtered, rankFilter, baseline, edits } = options;
  const knownRanks = [
    ...new Set([...ranks, ...effectiveRows.map((row) => row.rank).filter((rank): rank is string => !!rank)]),
  ];
  type Group = { key: string; group: true; rank: string; name: string; count: number };
  return knownRanks.flatMap<T | Group>((rank) => {
    if (rankFilter && rankFilter !== rank) return [];
    const rows = effectiveRows.filter((row) => row.rank === rank);
    let count = filtered ? rows.length : (rankCounts?.[rank] ?? rows.length);
    if (!filtered && rankCounts?.[rank] !== undefined)
      for (const [id, edit] of Object.entries(edits)) {
        const original = baseline[id];
        if (edit.rank && original?.rank && edit.rank !== original.rank)
          count += (edit.rank === rank ? 1 : 0) - (original.rank === rank ? 1 : 0);
      }
    if (!count && !rows.length) return [];
    const label = entityPriorityMap[rank] ?? rank;
    return [
      {
        key: `rank-${rank}`,
        group: true,
        rank,
        name: label.charAt(0) + label.slice(1).toLowerCase(),
        count: Math.max(count, rows.length),
      },
      ...(expanded.has(rank) ? rows : []),
    ];
  });
}
export function previewDepartmentWeightPercentage(
  row: IDepartmentSchedulingEdit,
  weight: number,
  total: number | undefined,
  completePeers: IDepartmentSchedulingEdit[] | undefined,
  ranked: boolean,
) {
  if (total != null && Number.isFinite(total) && row.weight != null && row.weight >= 0 && total >= row.weight)
    return weightPercentage(weight, total - row.weight + weight);
  if (!completePeers) return undefined;
  const key = ranked ? (row.rank ?? "") : "all";
  const otherWeights = completePeerSchedulingTotals(completePeers, { excludeId: row.id, requireIntegerWeights: true })[
    key
  ]?.rawWeightsSum;
  return otherWeights == null ? undefined : weightPercentage(weight, otherWeights + weight);
}
