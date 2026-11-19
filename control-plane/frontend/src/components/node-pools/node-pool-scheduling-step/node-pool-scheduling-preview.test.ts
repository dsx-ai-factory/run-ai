/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it } from "vitest";
import { matchesSchedulingFilters } from "./node-pool-scheduling-filters";
import {
  getMockSchedulingResult,
  MOCK_SCHEDULING_IDS,
  MOCK_SCHEDULING_TEN_IDS,
} from "../../../../unit-test/fixtures/node-pool-scheduling-simulation";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";
describe("scheduling preview boundaries", () => {
  it("uses authored guaranteed and over-quota values for the five-row Automatic profile", () => {
    const result = getMockSchedulingResult(MOCK_SCHEDULING_IDS, 144, {}, "guaranteed-automatic");
    expect(result).toBeDefined();
    const outputs = Object.values(result ?? {});
    expect(outputs.reduce((sum, row) => sum + row.allocated, 0)).toBe(144);
    expect(outputs.map((row) => row.guaranteed)).toEqual([20, 20, 20, 20, 20]);
    expect(outputs.map((row) => row.overQuota)).toEqual([9, 9, 9, 9, 8]);
    expect(
      outputs.every((row) => row.guaranteed + row.overQuota === row.allocated && row.allocated + row.unmet === 144),
    ).toBe(true);
  });
  it("keeps the authored overcommit output internally consistent and separates guaranteed allocation from over-quota", () => {
    const result = getMockSchedulingResult(MOCK_SCHEDULING_TEN_IDS, 144, {}, "guaranteed-overcommit");
    expect(result).toBeDefined();
    const outputs = Object.values(result ?? {});
    expect(outputs.reduce((sum, row) => sum + row.allocated, 0)).toBe(144);
    expect(
      outputs.every(
        (row) => row.guaranteed === row.allocated && row.overQuota === 0 && row.allocated + row.unmet === 144,
      ),
    ).toBe(true);
  });
  it("applies multiple ranks and a negative name condition together", () => {
    const filters = [
      {
        name: "rank",
        label: "Rank",
        type: EColumnFilterType.EnumBased,
        selectOptions: [],
        selectedValues: ["High", "Low"],
      },
      {
        name: "department",
        label: "Department",
        type: EColumnFilterType.FreeText,
        value: EFilterOperator.DoesNotContain,
        term: "test",
      },
    ];
    expect(matchesSchedulingFilters({ name: "Production", rank: "High" }, filters)).toBe(true);
    expect(matchesSchedulingFilters({ name: "Test production", rank: "High" }, filters)).toBe(false);
    expect(matchesSchedulingFilters({ name: "Production", rank: "MediumLow" }, filters)).toBe(false);
  });
  it("never assigns fixture outputs to real departments or modified demand", () => {
    expect(getMockSchedulingResult(["real-department"], 144, {})).toBeUndefined();
    expect(getMockSchedulingResult(MOCK_SCHEDULING_IDS, 144, {})).toBeDefined();
    expect(getMockSchedulingResult(MOCK_SCHEDULING_IDS, 144, { [MOCK_SCHEDULING_IDS[0]]: 20 })).toBeUndefined();
    expect(getMockSchedulingResult(MOCK_SCHEDULING_IDS, 0, {})).toBeUndefined();
  });
});
