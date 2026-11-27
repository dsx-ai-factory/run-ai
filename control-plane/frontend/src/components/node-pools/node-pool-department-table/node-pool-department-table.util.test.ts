/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it } from "vitest";
import { previewDepartmentWeightPercentage } from "@/components/node-pools/node-pool-department-table/node-pool-department-table.util";

describe("unapplied weight percentages", () => {
  it("includes hidden peers and leaves applied weights unchanged", () => {
    const row = { id: "a", weight: 2, rank: "Low" };
    const peers = [row, { id: "hidden", weight: 8, rank: "Low" }];
    expect(previewDepartmentWeightPercentage(row, 8, undefined, peers, true)).toBe(50);
    expect(row.weight).toBe(2);
  });
  it("can complete the only unknown peer in its rank", () => {
    const row = { id: "a", weight: undefined, rank: "Low" };
    const peers = [row, { id: "other-rank", weight: undefined, rank: "High" }];
    expect(previewDepartmentWeightPercentage(row, 3, undefined, peers, true)).toBe(100);
    expect(previewDepartmentWeightPercentage(row, 3, undefined, peers, false)).toBeUndefined();
  });
});
