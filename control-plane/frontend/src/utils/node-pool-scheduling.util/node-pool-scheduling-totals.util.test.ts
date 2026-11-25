import { describe, expect, it } from "vitest";
import {
  completePeerSchedulingTotals,
  effectiveDepartmentEdit,
} from "@/utils/node-pool-scheduling.util/node-pool-scheduling-totals.util";

describe("complete department peer totals", () => {
  const rows = [
    { id: "a", weight: 2, quota: 3, rank: "Low" },
    { id: "hidden", weight: 4, quota: -1, rank: "Low" },
    { id: "b", weight: 0, quota: 0, rank: "High" },
  ];
  it("keeps server snapshot quota and independent automatic/custom weights", () => {
    expect(completePeerSchedulingTotals(rows).all).toEqual({
      rawWeightsSum: 6,
      automaticWeightsSum: 3.01,
      finiteQuotaSum: 3,
      unlimitedQuotaCount: 1,
      deservedSum: -1,
    });
    expect(completePeerSchedulingTotals(rows).High.rawWeightsSum).toBe(0);
  });
  it("preserves unknown custom peers and excludes only an unapplied edited row", () => {
    const unknown = rows.map((row) => (row.id === "hidden" ? { ...row, weight: undefined } : row));
    expect(completePeerSchedulingTotals(unknown).all.rawWeightsSum).toBeUndefined();
    expect(completePeerSchedulingTotals(unknown, { excludeId: "hidden" }).all.rawWeightsSum).toBe(2);
    expect(completePeerSchedulingTotals(rows, { rawWeightsAvailable: false }).all.rawWeightsSum).toBeUndefined();
  });
  it("counts stored fractional weights on reopened custom pools", () => {
    const peers = [{ id: "a", weight: 1.5 }];
    expect(completePeerSchedulingTotals(peers).all.rawWeightsSum).toBe(1.5);
  });
  it("merges sparse edits before rank and weight recovery", () => {
    const baseline = Object.fromEntries(rows.map((row) => [row.id, row]));
    const edits = { a: { id: "a", rank: "High" } };
    const effective = rows.map((row) => effectiveDepartmentEdit(baseline, edits, row.id));
    expect(completePeerSchedulingTotals(effective).High.rawWeightsSum).toBe(2);
    expect(completePeerSchedulingTotals(effective).Low.rawWeightsSum).toBe(4);
  });
});
