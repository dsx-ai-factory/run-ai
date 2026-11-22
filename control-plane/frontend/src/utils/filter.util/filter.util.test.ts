/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, test, expect } from "vitest";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";

describe("filter.util", () => {
  test("parseEnumToSelectOptions", () => {
    enum ReportStatus {
      Pending = "pending",
      Processing = "processing",
    }
    const result = filterUtil.parseEnumToSelectOptions(ReportStatus);

    expect(result).toEqual([
      { label: "Pending", value: "Pending" },
      { label: "Processing", value: "Processing" },
    ]);
  });

  test("getFreeTextFilterQuery", () => {
    const name = "workloadName";
    const term = "test";
    const label = "Workload Name";
    const operator = EFilterOperator.Contains;

    const result = filterUtil.getFreeTextFilterQuery(name, term, label, operator);

    expect(JSON.parse(result)).toEqual({ name, term, label, value: operator, type: EColumnFilterType.FreeText });
  });

  test("getFreeTextFilterQuery escapes quotes in the term", () => {
    const term = 'gpu"batch';

    const result = filterUtil.getFreeTextFilterQuery("nodepool", term, "Node pool", EFilterOperator.Equals);

    expect(JSON.parse(result).term).toBe(term);
  });
});
