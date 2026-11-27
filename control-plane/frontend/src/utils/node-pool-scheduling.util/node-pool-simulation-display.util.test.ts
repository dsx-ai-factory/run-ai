/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it } from "vitest";
import { integerSimulationValues } from "@/utils/node-pool-scheduling.util/node-pool-simulation-display.util";

describe("integer simulation display", () => {
  it.each([
    {
      source: { demand: 2.8, allocated: 1.4, guaranteed: 0.6 },
      expected: { demand: 3, allocated: 1, guaranteed: 1, overQuota: 0, unmet: 2 },
    },
    {
      source: { demand: 1, allocated: 0.5, guaranteed: 0.25 },
      expected: { demand: 1, allocated: 1, guaranteed: 0, overQuota: 1, unmet: 0 },
    },
    {
      source: { demand: 0.4, allocated: 0.4, guaranteed: 0.4 },
      expected: { demand: 0, allocated: 0, guaranteed: 0, overQuota: 0, unmet: 0 },
    },
    {
      source: { demand: 4.5, allocated: 4.5, guaranteed: 1.5 },
      expected: { demand: 5, allocated: 5, guaranteed: 2, overQuota: 3, unmet: 0 },
    },
    {
      source: { demand: 3, allocated: 3, guaranteed: 1 },
      expected: { demand: 3, allocated: 3, guaranteed: 1, overQuota: 2, unmet: 0 },
    },
  ])("conserves the displayed totals for $source", ({ source, expected }) => {
    const original = { ...source };
    const display = integerSimulationValues(source);
    expect(display).toEqual(expected);
    expect(display.guaranteed + display.overQuota).toBe(display.allocated);
    expect(display.allocated + display.unmet).toBe(display.demand);
    expect(Object.values(display).every((value) => Number.isInteger(value) && value >= 0)).toBe(true);
    expect(source).toEqual(original);
  });
});
