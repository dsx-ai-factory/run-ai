/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi, beforeEach } from "vitest";
import { allProjectColumnsMap, projectIndexColumns, EProjectColumnLabel } from "./project.table-model";
import { ECustomCell } from "@/models/table.model";
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

describe("project.table-model - CPU memory allocation ratio column", () => {
  const column = allProjectColumnsMap.memoryUtilization;
  const getCellValue = (project: IProjectTable): string =>
    column.format?.((column.field as (row: IProjectTable) => unknown)(project), project);

  it("compares the allocated memory and the memory quota in the same unit (MB)", () => {
    const project = {
      status: { quotaStatus: { allocated: { memory: 512 } } },
      totalResources: { memoryQuota: 1024 },
    } as IProjectTable;

    expect(getCellValue(project)).toBe("50.00%");
  });

  it("shows '-' when the project has no memory quota", () => {
    const project = {
      status: { quotaStatus: { allocated: { memory: 512 } } },
      totalResources: { memoryQuota: 0 },
    } as IProjectTable;

    expect(getCellValue(project)).toBe("-");
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
