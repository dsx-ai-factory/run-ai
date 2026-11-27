/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { dataSourceIndexColumns } from "./data-source.table-model";
import { filterService } from "@/services/filter.service/filter.service";
import { useAuthStore } from "@/stores/auth.store";
import { CLUSTER_COLUMN_FILTER_NAME, type IFilterModel } from "@/models/filter.model";
import { Scope, type DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";
import { ScopeType, type OrgUnit } from "@/swagger-models/org-unit-service-client";

// Department Research owns A, its project Vision owns B, department Other owns C.
// D belongs to a project missing from the org tree; E is cluster-scoped.
const row = (name: string, meta: object): DatasourceListResponseEntry =>
  ({ meta: { name, kind: "nfs", createdAt: "2026-01-01T00:00:00Z", clusterId: "c1", ...meta } }) as never;

const rows: DatasourceListResponseEntry[] = [
  row("A", { scope: Scope.Department, departmentId: "2" }),
  row("B", { scope: Scope.Project, projectId: 10, departmentId: "2" }),
  row("C", { scope: Scope.Department, departmentId: "3" }),
  row("D", { scope: Scope.Project, projectId: 99 }),
  row("E", { scope: Scope.Cluster, clusterId: "c2" }),
];

const research: OrgUnit = { id: "2", type: ScopeType.Department, name: "Research" } as OrgUnit;
const other: OrgUnit = { id: "3", type: ScopeType.Department, name: "Other" } as OrgUnit;
const vision: OrgUnit = {
  id: "10",
  type: ScopeType.Project,
  name: "Vision",
  parentId: "2",
  parentType: ScopeType.Department,
} as OrgUnit;
const fullTree: OrgUnit[] = [research, other, vision];

const SCOPE_FILTERS = ["project", "department"];

describe("data-source table: project and department filters", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it.each([
    {
      title: "department matches its own and descendant-project assets",
      filters: { department: "research" },
      expected: ["A", "B"],
    },
    { title: "department match ignores case", filters: { department: "RESEARCH" }, expected: ["A", "B"] },
    { title: "project matches only that project's assets", filters: { project: "Vision" }, expected: ["B"] },
    { title: "an unrelated department is excluded", filters: { department: "Other" }, expected: ["C"] },
    {
      title: "a project without a parent in the tree is never matched by department",
      filters: { department: "Research" },
      tree: [research, other, { ...vision, parentId: undefined, parentType: undefined }],
      expected: ["A"],
    },
    { title: "an empty org tree guesses no department", filters: { department: "Research" }, tree: [], expected: [] },
    {
      title: "the cluster filter keeps working on its own",
      filters: { [CLUSTER_COLUMN_FILTER_NAME]: "c1" },
      expected: ["A", "B", "C", "D"],
    },
    {
      title: "department combines with the cluster filter",
      filters: { department: "Research", [CLUSTER_COLUMN_FILTER_NAME]: "c1" },
      expected: ["A", "B"],
    },
  ])("$title", ({ filters, tree = fullTree, expected }) => {
    useAuthStore().orgUnits = tree;
    const scopeNames = Object.keys(filters).filter((n) => SCOPE_FILTERS.includes(n));

    // The "+ Filter" menu offers the scope filters even though their columns are hidden by default.
    // The cluster filter has its own dropdown, so it is not part of this menu.
    const offered = filterService
      .getAdvancedFilterOptions(dataSourceIndexColumns.filter((c) => !c.hideFilter))
      .map((o) => o.name);
    expect(offered).toEqual(expect.arrayContaining(scopeNames));
    for (const name of scopeNames) {
      expect(dataSourceIndexColumns.find((c) => c.name === name)?.display).toBe(false);
    }

    const columnFilters = Object.entries(filters).map(([name, term]) => ({ name, term }) as IFilterModel);
    const matched = filterService.filterByColumns(rows, columnFilters, dataSourceIndexColumns);
    expect(matched.map((r) => r.meta.name)).toEqual(expected);
  });
});
