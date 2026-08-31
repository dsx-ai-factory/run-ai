/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it, vi, beforeEach } from "vitest";
import { allDepartmentColumnsMap, departmentIndexColumns, EDepartmentColumnLabel } from "./department.table-model";
import { ECustomCell } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import type { IDepartmentTable } from "@/models/department.model";
import { exportFile } from "quasar";

vi.mock("quasar", async () => {
  const actual = await vi.importActual("quasar");
  return {
    ...(actual as object),
    exportFile: vi.fn(() => true),
    Notify: { create: vi.fn() },
  };
});

describe("department.table-model - workloads column (viewLinkColumn)", () => {
  const column = allDepartmentColumnsMap.workloads;

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

describe("department.table-model - CSV export flow (departmentIndexColumns via tableUtil.exportTableAsCsv)", () => {
  const mockedExportFile = vi.mocked(exportFile);

  beforeEach(() => {
    mockedExportFile.mockClear();
  });

  const department = { name: "department-1" } as IDepartmentTable;

  const getCsvHeader = (): string[] => {
    const call = mockedExportFile.mock.calls[0];
    if (!call) throw new Error("exportFile was not called");
    const [headerLine] = (call[1] as string).split("\r\n");
    return headerLine.split(",");
  };

  it("drops the workloads drill-in column from the exported CSV", () => {
    tableUtil.exportTableAsCsv("departments", [department], departmentIndexColumns);

    const header = getCsvHeader();
    expect(header).not.toContain(EDepartmentColumnLabel.Workloads);
  });

  it("keeps the other columns in the exported CSV", () => {
    tableUtil.exportTableAsCsv("departments", [department], departmentIndexColumns);

    const header = getCsvHeader();
    expect(header).toContain(EDepartmentColumnLabel.Name);
  });
});
