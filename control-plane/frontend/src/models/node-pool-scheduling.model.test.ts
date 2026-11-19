/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it } from "vitest";
import {
  recommendedDepartmentQuotaMinimum,
  departmentQuotaBudget,
  weightPercentage,
  validateDepartmentEdit,
  validateDepartmentValues,
  isDepartmentLimitBelowQuota,
  schedulingTotalsWithEdits,
} from "@/models/node-pool-scheduling.model";
describe("department scheduling validation", () => {
  it("recommends the maximum known project quota and non-preemptible usage floor", () => {
    expect(recommendedDepartmentQuotaMinimum({ childQuotaSum: 3, nonPreemptibleUsage: 10 })).toBe(10);
    expect(recommendedDepartmentQuotaMinimum({ childQuotaSum: 3 })).toBe(3);
    expect(recommendedDepartmentQuotaMinimum({})).toBeUndefined();
  });
  it("keeps quota budgets unknown for partial totals and clamps known exhausted capacity to zero", () => {
    expect(
      departmentQuotaBudget(6, { capacity: 144, finiteQuotaSum: 30, unlimitedQuotaCount: 0, totalsComplete: false }),
    ).toBeUndefined();
    expect(departmentQuotaBudget(6, { capacity: 144, finiteQuotaSum: 206, unlimitedQuotaCount: 0 })).toEqual({
      available: 0,
      otherFiniteQuota: 200,
    });
    expect(departmentQuotaBudget(6, { capacity: 0, finiteQuotaSum: 6, unlimitedQuotaCount: 0 })).toEqual({
      available: 0,
      otherFiniteQuota: 0,
    });
  });
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
  it("recomputes finite quotas when the final unlimited department becomes finite", () => {
    const server = { all: { weightsSum: 4, deservedSum: -1, finiteQuotaSum: 10, unlimitedQuotaCount: 1 } };
    const baseline = { a: { id: "a", quota: -1, weight: 2 }, b: { id: "b", quota: 10, weight: 2 } };
    expect(schedulingTotalsWithEdits(server, baseline, { a: { id: "a", quota: 5 } }, false).all.deservedSum).toBe(15);
    expect(schedulingTotalsWithEdits(server, baseline, { b: { id: "b", quota: 12 } }, false).all.deservedSum).toBe(-1);
    expect(
      schedulingTotalsWithEdits(
        { all: { weightsSum: 4, deservedSum: -1 } },
        baseline,
        { b: { id: "b", quota: 12 } },
        false,
      ).all.deservedSum,
    ).toBe(-1);
  });
});

describe("primitive scheduling values without project context", () => {
  it("validates a merged limit without requiring child project totals", () => {
    expect(validateDepartmentValues({ id: "a", quota: 3, limit: 2 })).toEqual([
      "Maximum allocation must be at least guaranteed GPUs.",
    ]);
    expect(validateDepartmentValues({ id: "a", quota: 3, limit: null })).toEqual([]);
    expect(validateDepartmentEdit({ id: "a", limit: 4 }, {}).errors).toEqual([]);
  });
  it.each([NaN, Infinity, -1, 21, 1.5])("rejects an invalid weight %s consistently", (weight) => {
    expect(validateDepartmentValues({ id: "a", weight })).toEqual(["Weight must be a whole number from 0 to 20."]);
  });
});

describe("effective quota and maximum allocation", () => {
  it("compares finite limits while preserving unlimited and unknown values", () => {
    expect(isDepartmentLimitBelowQuota(2, 3)).toBe(true);
    expect(isDepartmentLimitBelowQuota(3, 3)).toBe(false);
    expect(isDepartmentLimitBelowQuota(-1, 3)).toBe(false);
    expect(isDepartmentLimitBelowQuota(null, 3)).toBe(false);
    expect(isDepartmentLimitBelowQuota(undefined, 3)).toBe(false);
    expect(isDepartmentLimitBelowQuota(2, undefined)).toBe(false);
  });
});
