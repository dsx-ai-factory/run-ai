/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { filterClustersByAdvancedFilters } from "./cluster-filter.util";
import type { IAdvancedFilterModel, IEnumBasedFilterModel, IFreeTextFilterModel } from "@/models/filter.model";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";

interface IRow {
  name: string;
  status: string;
  version: string;
}

const rows: IRow[] = [
  { name: "alpha-cluster", status: "Connected", version: "2.20.0" },
  { name: "beta-cluster", status: "Disconnected", version: "2.21.0" },
  { name: "gamma-cluster", status: "Connected", version: "2.21.1" },
];

const columns: ITableColumn[] = [
  { name: "name", label: "Name", field: "name" },
  {
    name: "status",
    label: "Status",
    field: (row: IRow) => row.status,
    format: (status: string) => ({ status, color: "ok" }),
    filterKey: "status",
  },
  { name: "version", label: "Version", field: "version" },
];

describe("filterClustersByAdvancedFilters", () => {
  it("returns the original list when no filters are provided", () => {
    expect(filterClustersByAdvancedFilters(rows, undefined, columns)).toEqual(rows);
    expect(filterClustersByAdvancedFilters(rows, [], columns)).toEqual(rows);
  });

  it("filters by free-text contains (case-insensitive)", () => {
    const filters: IFreeTextFilterModel[] = [
      {
        name: "name",
        label: "Name",
        type: EColumnFilterType.FreeText,
        term: "ALPHA",
        value: EFilterOperator.Contains,
      },
    ];

    const result = filterClustersByAdvancedFilters(rows, filters, columns);
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe("alpha-cluster");
  });

  it("filters by free-text does-not-contain", () => {
    const filters: IFreeTextFilterModel[] = [
      {
        name: "name",
        label: "Name",
        type: EColumnFilterType.FreeText,
        term: "beta",
        value: EFilterOperator.DoesNotContain,
      },
    ];

    const result = filterClustersByAdvancedFilters(rows, filters, columns);
    expect(result.map((r) => r.name)).toEqual(["alpha-cluster", "gamma-cluster"]);
  });

  it("filters by enum selectedValues using a column with format + filterKey", () => {
    const filters: IEnumBasedFilterModel[] = [
      {
        name: "status",
        label: "Status",
        type: EColumnFilterType.EnumBased,
        selectOptions: [
          { label: "Connected", value: "Connected" },
          { label: "Disconnected", value: "Disconnected" },
        ],
        selectedValues: ["Connected"],
      },
    ];

    const result = filterClustersByAdvancedFilters(rows, filters, columns);
    expect(result.map((r) => r.name)).toEqual(["alpha-cluster", "gamma-cluster"]);
  });

  it("does not let 'Connected' match 'Disconnected' (the original regression case)", () => {
    const filters: IEnumBasedFilterModel[] = [
      {
        name: "status",
        label: "Status",
        type: EColumnFilterType.EnumBased,
        selectOptions: [{ label: "Connected", value: "Connected" }],
        selectedValues: ["Connected"],
      },
    ];

    const result = filterClustersByAdvancedFilters(rows, filters, columns);
    expect(result.every((r) => r.status === "Connected")).toBe(true);
  });

  it("AND-combines multiple filters", () => {
    const filters: IAdvancedFilterModel[] = [
      {
        name: "status",
        label: "Status",
        type: EColumnFilterType.EnumBased,
        selectOptions: [{ label: "Connected", value: "Connected" }],
        selectedValues: ["Connected"],
      } as IEnumBasedFilterModel,
      {
        name: "version",
        label: "Version",
        type: EColumnFilterType.FreeText,
        term: "2.21",
        value: EFilterOperator.StartsWith,
      } as IFreeTextFilterModel,
    ];

    const result = filterClustersByAdvancedFilters(rows, filters, columns);
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe("gamma-cluster");
  });

  it("ignores filters for columns not in the model", () => {
    const filters: IFreeTextFilterModel[] = [
      {
        name: "unknown",
        label: "Unknown",
        type: EColumnFilterType.FreeText,
        term: "anything",
        value: EFilterOperator.Contains,
      },
    ];

    expect(filterClustersByAdvancedFilters(rows, filters, columns)).toEqual(rows);
  });
});
