/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it } from "vitest";
import { weightPercentage, validateDepartmentEdit, schedulingTotalsWithEdits } from "./node-pool-scheduling.model";
describe("department scheduling validation", () => {
  it("uses known server denominators including zero", () => {
    expect(weightPercentage(2, 10)).toBe(20);
    expect(weightPercentage(2, 0)).toBe(0);
    expect(weightPercentage(2, undefined)).toBeUndefined();
  });
  it.each(["", "abc", -1, 1.5, Infinity])("rejects invalid raw weight %s", (weight) => {
    expect(validateDepartmentEdit({ id: "a", weight: weight as number }, {}).errors.length).toBeGreaterThan(0);
  });
  it("accepts zero and fractional quota", () => {
    expect(validateDepartmentEdit({ id: "a", weight: 0, quota: 0.5 }, { childQuotaSum: 0 }).errors).toEqual([]);
  });
  it("checks limits, child quota policy, capacity and missing required totals", () => {
    expect(validateDepartmentEdit({ id: "a", quota: 3, limit: 2 }, { childQuotaSum: 0 }).errors.join()).toContain(
      "Maximum",
    );
    expect(
      validateDepartmentEdit({ id: "a", quota: 3 }, { childQuotaSum: 4, enforceChildQuota: false }).warnings.length,
    ).toBe(1);
    expect(
      validateDepartmentEdit({ id: "a", quota: 3 }, { childQuotaSum: 4, enforceChildQuota: true }).errors.length,
    ).toBe(1);
    expect(validateDepartmentEdit({ id: "a", quota: 3 }, {}).errors.join()).toContain("unavailable");
    expect(
      validateDepartmentEdit({ id: "a", quota: 3 }, { childQuotaSum: 0, totalQuota: 5, capacity: 4 }).warnings.length,
    ).toBe(1);
  });
  it("applies hidden edits and rank moves to whole-pool denominators", () => {
    const baseline = { a: { id: "a", weight: 2, quota: 1, rank: "Low" } };
    const edits = { a: { id: "a", weight: 5, quota: 2, rank: "High" } };
    const totals = schedulingTotalsWithEdits(
      {
        all: { weightsSum: 10, deservedSum: 3 },
        Low: { weightsSum: 4, deservedSum: 1 },
        High: { weightsSum: 6, deservedSum: 2 },
      },
      baseline,
      edits,
      false,
    );
    expect(totals.all.weightsSum).toBe(13);
    expect(totals.Low.weightsSum).toBe(2);
    expect(totals.High.weightsSum).toBe(11);
  });
  it("uses independent raw totals when saved automatic sums have quota semantics", () => {
    const totals = { all: { weightsSum: 12, deservedSum: 12, rawWeightsSum: 6, automaticWeightsSum: 12 } };
    expect(schedulingTotalsWithEdits(totals, {}, {}, false).all.weightsSum).toBe(6);
    expect(schedulingTotalsWithEdits(totals, {}, {}, true).all.weightsSum).toBe(12);
  });
  it("keeps raw totals unknown when saved automatic mode hides raw weights", () => {
    const totals = { all: { weightsSum: 12, deservedSum: 12, rawWeightsSum: undefined, automaticWeightsSum: 12 } };
    expect(schedulingTotalsWithEdits(totals, {}, {}, false).all.weightsSum).toBeUndefined();
    expect(schedulingTotalsWithEdits(totals, {}, {}, true).all.weightsSum).toBe(12);
  });
});
