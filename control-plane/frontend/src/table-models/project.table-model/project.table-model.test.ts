/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi, beforeEach } from "vitest";
import { allProjectColumnsMap, projectIndexColumns, EProjectColumnLabel } from "./project.table-model";
import { ECustomCell, type ITableColumn } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import type { IProjectTable } from "@/models/project.model";
import { exportFile } from "quasar";

vi.mock("quasar", async () => {
  const actual = await vi.importActual("quasar");
  return {
    ...(actual as object),
    exportFile: vi.fn(() => true),
    Notify: { create: vi.fn() },
  };
});

describe("project.table-model - workloads column (viewLinkColumn)", () => {
  const column = allProjectColumnsMap.workloads;

  it("renders the 'View' drill-in link (unsortable, unfiltered, no CSV export)", () => {
    expect(column).toMatchObject({
      sortable: false,
      align: "left",
      hideFilter: true,
      hideExport: true,
      customCell: ECustomCell.LINK_COL,
    });
    expect((column.field as () => string)()).toBe("View");
  });

  it('emits "workloads-clicked" when clicked', () => {
    expect(column.customCellEvent).toEqual({ emitName: "workloads-clicked" });
  });
});

describe("project.table-model - CSV export flow (projectIndexColumns via tableUtil.exportTableAsCsv)", () => {
  const mockedExportFile = vi.mocked(exportFile);

  beforeEach(() => {
    mockedExportFile.mockClear();
  });

  const project = { name: "project-1", resources: [] } as unknown as IProjectTable;

  const getCsvHeader = (): string[] => {
    const call = mockedExportFile.mock.calls[0];
    if (!call) throw new Error("exportFile was not called");
    const [headerLine] = (call[1] as string).split("\r\n");
    return headerLine.split(",");
  };

  it("drops the workloads drill-in column from the exported CSV", () => {
    tableUtil.exportTableAsCsv("projects", [project], projectIndexColumns);

    const header = getCsvHeader();
    expect(header).not.toContain(EProjectColumnLabel.Workloads);
  });

  it("keeps the other columns in the exported CSV", () => {
    tableUtil.exportTableAsCsv("projects", [project], projectIndexColumns);

    const header = getCsvHeader();
    expect(header).toContain(EProjectColumnLabel.ProjectName);
  });
});

describe("project.table-model - GPU allocation breakdown columns", () => {
  const getProject = (project: Partial<IProjectTable>): IProjectTable =>
    ({ name: "project-1", resources: [], ...project }) as IProjectTable;
  const getCellValue = (column: ITableColumn, project: IProjectTable): string => {
    const value = (column.field as (row: IProjectTable) => unknown)(project);
    return column.format ? column.format(value, project) : value;
  };

  it("shows the non-preemptible part of the GPU allocation", () => {
    const project = getProject({
      status: { quotaStatus: { allocated: { gpu: 3 }, allocatedNonPreemptible: { gpu: 1.5 } } },
    });

    expect(getCellValue(allProjectColumnsMap.allocatedNonPreemptibleGpus, project)).toBe("1.50");
  });

  it("shows '-' when no non-preemptible GPUs are allocated", () => {
    const project = getProject({ status: { quotaStatus: { allocated: { gpu: 3 } } } });

    expect(getCellValue(allProjectColumnsMap.allocatedNonPreemptibleGpus, project)).toBe("-");
  });

  it("shows the GPUs allocated above the node pool quota", () => {
    const project = getProject({
      resources: [{ nodePool: { id: "1", name: "default" }, gpu: { deserved: 0 } }],
      status: { nodePoolQuotaStatuses: [{ nodePoolId: "1", nodePoolName: "default", allocated: { gpu: 0.5 } }] },
    });

    expect(getCellValue(allProjectColumnsMap.overQuotaGpus, project)).toBe("0.50");
  });

  it("shows '-' when the allocation is within the quota", () => {
    const project = getProject({
      resources: [{ nodePool: { id: "1", name: "default" }, gpu: { deserved: 2 } }],
      status: { nodePoolQuotaStatuses: [{ nodePoolId: "1", nodePoolName: "default", allocated: { gpu: 2 } }] },
    });

    expect(getCellValue(allProjectColumnsMap.overQuotaGpus, project)).toBe("-");
  });

  it.each([allProjectColumnsMap.allocatedNonPreemptibleGpus, allProjectColumnsMap.overQuotaGpus])(
    "offers the $label column in the projects table, hidden by default",
    (column: ITableColumn) => {
      const indexColumn = projectIndexColumns.find((col) => col.name === column.name);

      expect(indexColumn).toMatchObject({ display: false, sortable: false, hideFilter: true });
    },
  );
});
