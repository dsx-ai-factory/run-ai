/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, test } from "vitest";
import {
  allWorkloadColumnsMap,
  allWorkloadPodsModalColumnsMap,
  deletedWorkloadIndexColumns,
  workloadIndexColumns,
} from "./workload.table-model";
import { Preemptibility, WorkloadSortFilterFields, type Pod, type Workload } from "@/swagger-models/workloads-client";
import type { ITableColumn } from "@/models/table.model";
import { EColumnFilterType } from "@/models/filter.model";

describe("workload table model - pods modal image field", () => {
  const imageField = allWorkloadPodsModalColumnsMap.image.field as (row: Pod) => string;

  test("should return user-container image when present", () => {
    const pod: Pod = {
      containers: [
        { name: "sidecar", image: "sidecar:latest" },
        { name: "user-container-1", image: "user-image:latest" },
        { name: "init", image: "init:latest" },
      ],
    } as Pod;

    expect(imageField(pod)).toBe("user-image:latest");
  });

  test("should return first container image when no user-container present", () => {
    const pod: Pod = {
      containers: [
        { name: "sidecar", image: "sidecar:latest" },
        { name: "init", image: "init:latest" },
      ],
    } as Pod;

    expect(imageField(pod)).toBe("sidecar:latest");
  });
});

describe("deletedWorkloadIndexColumns", () => {
  const getColumnNames = (columns: ITableColumn[]) => columns.map((col) => col.label);
  const getDefaultVisibleColumns = (columns: ITableColumn[]) =>
    columns.filter((col) => col.display).map((col) => col.label);

  test("should have correct default visible columns", () => {
    const defaultVisible = getDefaultVisibleColumns(deletedWorkloadIndexColumns);

    expect(defaultVisible).toEqual([
      "Workload",
      "Type",
      "Project",
      "Department",
      "GPU compute request",
      "GPU memory request",
      "Idle GPU devices",
      "Cluster",
    ]);
  });

  test("should have correct column order", () => {
    const columnOrder = getColumnNames(deletedWorkloadIndexColumns);

    expect(columnOrder).toEqual([
      "Workload",
      "Type",
      "Project",
      "Department",
      "Category",
      "Priority",
      "Preemptibility",
      "AI application",
      "Requested runs",
      "Created by",
      "Creation time",
      "Completion time",
      "Total pending time",
      "Total runtime",
      "Guaranteed runtime left",
      "Connection(s)",
      "Image",
      "Workload architecture",
      "GPU compute request",
      "GPU memory request",
      "Idle GPU devices",
      "CPU compute request",
      "CPU memory request",
      "Cluster",
    ]);
  });

  test("should not include active-only columns", () => {
    const columnLabels = getColumnNames(deletedWorkloadIndexColumns);

    const activeOnlyColumns = [
      "Status",
      "Node pool",
      "Running / requested pods",
      "GPU compute allocation",
      "GPU memory allocation",
      "CPU compute allocation",
      "CPU memory allocation",
    ];

    activeOnlyColumns.forEach((label) => {
      expect(columnLabels).not.toContain(label);
    });
  });

  test("should not have priority, preemptibility, or connections displayed by default", () => {
    const priority = deletedWorkloadIndexColumns.find((col) => col.label === "Priority");
    const preemptible = deletedWorkloadIndexColumns.find((col) => col.label === "Preemptibility");
    const connections = deletedWorkloadIndexColumns.find((col) => col.label === "Connection(s)");

    expect(priority?.display).toBe(false);
    expect(preemptible?.display).toBe(false);
    expect(connections?.display).toBe(false);
  });

  test("should have project and department right after type", () => {
    const columnLabels = getColumnNames(deletedWorkloadIndexColumns);
    const typeIndex = columnLabels.indexOf("Type");
    const projectIndex = columnLabels.indexOf("Project");
    const departmentIndex = columnLabels.indexOf("Department");

    expect(projectIndex).toBe(typeIndex + 1);
    expect(departmentIndex).toBe(typeIndex + 2);
  });

  test("should render project column as plain text (not clickable link)", () => {
    const projectCol = deletedWorkloadIndexColumns.find((col) => col.label === "Project");
    const mockWorkload = { projectName: "my-project" };

    // In deleted view, field returns a string (plain text) instead of an array (clickable link)
    const fieldValue = (projectCol?.field as (row: typeof mockWorkload) => unknown)(mockWorkload);
    expect(fieldValue).toBe("my-project");
    expect(typeof fieldValue).toBe("string");
  });

  test("should differ from workloadIndexColumns in default visibility", () => {
    const activeDefaults = getDefaultVisibleColumns(workloadIndexColumns);
    const deletedDefaults = getDefaultVisibleColumns(deletedWorkloadIndexColumns);

    expect(activeDefaults).not.toEqual(deletedDefaults);
    expect(activeDefaults).toContain("Status");
    expect(deletedDefaults).not.toContain("Status");
  });
});

describe("workloadIndexColumns", () => {
  const getColumnNames = (columns: ITableColumn[]) => columns.map((col) => col.label);
  const getDefaultVisibleColumns = (columns: ITableColumn[]) =>
    columns.filter((col) => col.display).map((col) => col.label);

  test("should have correct default visible columns per Figma", () => {
    const defaultVisible = getDefaultVisibleColumns(workloadIndexColumns);

    expect(defaultVisible).toEqual([
      "Workload",
      "Type",
      "Status",
      "Project",
      "Department",
      "Node pool",
      "Running / requested pods",
      "GPU compute request",
      "GPU compute allocation",
      "GPU memory request",
      "GPU memory allocation",
      "Idle GPU devices",
    ]);
  });

  test("should have correct column order per Figma", () => {
    const columnOrder = getColumnNames(workloadIndexColumns);

    expect(columnOrder).toEqual([
      "Workload",
      "Type",
      "Status",
      "Project",
      "Department",
      "Category",
      "Priority",
      "Preemptibility",
      "Cluster",
      "AI application",
      "Node pool",
      "Running / requested pods",
      "Requested runs",
      "Created by",
      "Creation time",
      "Completion time",
      "Total pending time",
      "Total runtime",
      "Guaranteed runtime left",
      "Connection(s)",
      "Image",
      "Workload architecture",
      "GPU compute request",
      "GPU compute allocation",
      "GPU memory request",
      "GPU memory allocation",
      "Idle GPU devices",
      "CPU compute request",
      "CPU compute allocation",
      "CPU memory request",
      "CPU memory allocation",
    ]);
  });

  test("priority, preemptibility, connections, and cluster should be hidden by default", () => {
    const priority = workloadIndexColumns.find((col) => col.label === "Priority");
    const preemptible = workloadIndexColumns.find((col) => col.label === "Preemptibility");
    const connections = workloadIndexColumns.find((col) => col.label === "Connection(s)");
    const cluster = workloadIndexColumns.find((col) => col.label === "Cluster");

    expect(priority?.display).toBe(false);
    expect(preemptible?.display).toBe(false);
    expect(connections?.display).toBe(false);
    expect(cluster?.display).toBe(false);
  });

  test("supports filtering and sorting by all preemptibility values", () => {
    const column = allWorkloadColumnsMap.preemptible;
    const field = column.field as (workload: Workload) => string | null | undefined;

    expect(column.name).toBe(WorkloadSortFilterFields.Preemptibility);
    expect(column.sortable).toBe(true);
    expect(column.filter?.type).toBe(EColumnFilterType.EnumBased);
    expect(column.filter?.selectOptions).toEqual([
      { label: "Preemptible", value: Preemptibility.Preemptible },
      { label: "Semi-preemptible", value: Preemptibility.SemiPreemptible },
      { label: "Non-preemptible", value: Preemptibility.NonPreemptible },
    ]);
    expect(column.format?.(Preemptibility.Preemptible, {})).toBe("Preemptible");
    expect(column.format?.(Preemptibility.SemiPreemptible, {})).toBe("Semi-preemptible");
    expect(column.format?.(Preemptibility.NonPreemptible, {})).toBe("Non-preemptible");
    expect(column.format?.(null, {})).toBe("-");
    expect(field({ preemptibility: Preemptibility.SemiPreemptible, preemptible: false } as Workload)).toBe(
      Preemptibility.SemiPreemptible,
    );
    expect(field({ preemptible: true } as Workload)).toBe(Preemptibility.Preemptible);
    expect(field({ preemptible: false } as Workload)).toBe(Preemptibility.NonPreemptible);
    expect(field({} as Workload)).toBeUndefined();
  });
});
