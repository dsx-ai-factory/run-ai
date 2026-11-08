/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it } from "vitest";
import { matchesSchedulingFilters } from "./node-pool-scheduling-filters";
import { getMockSchedulingResult, MOCK_SCHEDULING_IDS } from "./node-pool-scheduling-simulation.mock";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";
describe("scheduling preview boundaries", () => {
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
