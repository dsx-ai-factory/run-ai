/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { tableUtil } from "@/utils/table.util";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { columnA, columnB, rowA, rowB, rowC, rowD } from "@/utils/table.util/table.util.mock";
import type { ITableColumn } from "@/models/table.model";
import { exportFile } from "quasar";
import { nodeIndexColumns, nodeColumnName } from "@/table-models/node.table-model";

vi.mock("quasar", async () => {
  const actual = await vi.importActual("quasar");
  return {
    ...(actual as object),
    exportFile: vi.fn(() => true),
    Notify: { create: vi.fn() },
  };
});

describe("table.util", () => {
  describe("exportTableAsCsv", () => {
    const mockedExportFile = vi.mocked(exportFile);

    beforeEach(() => {
      mockedExportFile.mockClear();
    });

    const getCsvContent = (): string => {
      const call = mockedExportFile.mock.calls[0];
      if (!call) throw new Error("exportFile was not called");
      return call[1] as string;
    };

    it("includes a regular column's header and per-row values", () => {
      const nameColumn: ITableColumn = { name: "name", label: "Name", field: "name" };
      tableUtil.exportTableAsCsv("nodes", [{ name: "node-1" }, { name: "node-2" }], [nameColumn]);

      const [header, row1, row2] = getCsvContent().split("\r\n");
      expect(header).toBe("Name");
      expect(row1).toBe("node-1");
      expect(row2).toBe("node-2");
    });

    it("omits a hideExport column's header and values entirely", () => {
      const nameColumn: ITableColumn = { name: "name", label: "Name", field: "name" };
      const hiddenColumn: ITableColumn = { name: "hidden", label: "Hidden", field: () => "secret", hideExport: true };
      tableUtil.exportTableAsCsv("nodes", [{ name: "node-1" }], [nameColumn, hiddenColumn]);

      const [header, row1] = getCsvContent().split("\r\n");
      expect(header).toBe("Name");
      expect(row1).toBe("node-1");
      expect(header).not.toContain("Hidden");
      expect(row1).not.toContain("secret");
    });

    it("uses exportFormat over field when both are present", () => {
      const column: ITableColumn = {
        name: "status",
        label: "Status",
        field: () => "raw-field-value",
        exportFormat: () => "formatted-export-value",
      };
      tableUtil.exportTableAsCsv("nodes", [{}], [column]);

      const [, row1] = getCsvContent().split("\r\n");
      expect(row1).toBe("formatted-export-value");
    });

    it("never invokes exportFormat when hideExport: true is also set — hideExport wins", () => {
      const exportFormat = vi.fn(() => "formatted-export-value");
      const column: ITableColumn = { name: "status", label: "Status", field: "status", hideExport: true, exportFormat };
      tableUtil.exportTableAsCsv("nodes", [{ status: "ready" }], [column]);

      expect(exportFormat).not.toHaveBeenCalled();
      expect(getCsvContent().split("\r\n")).toEqual(["", ""]);
    });

    it("omits the numaNodes column from CSV export, matching its hideExport: true default from viewLinkColumn", () => {
      const numaNodesColumn = nodeIndexColumns.find((col) => col.name === nodeColumnName.numaNodes);
      if (!numaNodesColumn) throw new Error("numaNodes column not found in nodeIndexColumns");
      const nameColumn: ITableColumn = { name: "name", label: "Name", field: "name" };

      tableUtil.exportTableAsCsv("nodes", [{ name: "node-1" }], [nameColumn, numaNodesColumn]);

      const [header, row1] = getCsvContent().split("\r\n");
      expect(header).toBe("Name");
      expect(row1).toBe("node-1");
    });
  });

  describe("customColumnSort", () => {
    it("should sort strings", () => {
      const expected = "a-department".localeCompare("b-department");
      const result = tableUtil.customColumnSort(columnA, rowA, rowB);

      expect(result).toBe(expected);
    });

    it("should sort numbers ", () => {
      const expected = 1 - 2;
      const result = tableUtil.customColumnSort(columnB, rowC, rowD);

      expect(result).toBe(expected);
    });
  });

  describe("_wrapCsvValue", () => {
    it("should handle undefined value", () => {
      expect(tableUtil._wrapCsvValue(undefined)).toBe("-");
    });

    it("should handle null value", () => {
      expect(tableUtil._wrapCsvValue(null)).toBe("-");
    });

    it("should handle empty string", () => {
      expect(tableUtil._wrapCsvValue("")).toBe("-");
    });

    it("should handle dash string", () => {
      expect(tableUtil._wrapCsvValue("-")).toBe("-");
    });

    it("should convert number to string", () => {
      expect(tableUtil._wrapCsvValue(123)).toBe("123");
    });

    it("should handle regular string", () => {
      expect(tableUtil._wrapCsvValue("hello")).toBe("hello");
    });

    it("should apply format function", () => {
      const formatFn = (val: string) => val.toUpperCase();
      expect(tableUtil._wrapCsvValue("hello", formatFn)).toBe("HELLO");
    });

    it("should handle format function with row context", () => {
      const row = { prefix: "Mr." };
      const formatFn = (val: string, r: typeof row) => `${r.prefix} ${val}`;
      expect(tableUtil._wrapCsvValue("Smith", formatFn, row)).toBe("Mr. Smith");
    });

    it("should sanitize cell when column.sanitizeCell is true", () => {
      const column: ITableColumn = {
        name: "test",
        label: "Test",
        field: "test",
        sanitizeCell: true,
      };
      expect(tableUtil._wrapCsvValue("=count(E2:E3)", undefined, undefined, column)).toBe("'=count(E2:E3)");
    });

    it("should not sanitize cell when column.sanitizeCell is false", () => {
      const column: ITableColumn = {
        name: "test",
        label: "Test",
        field: "test",
        sanitizeCell: false,
      };
      expect(tableUtil._wrapCsvValue("=count(E2:E3)", undefined, undefined, column)).toBe("=count(E2:E3)");
    });

    it("should not sanitize cell when column.sanitizeCell is undefined", () => {
      const column: ITableColumn = {
        name: "test",
        label: "Test",
        field: "test",
      };
      expect(tableUtil._wrapCsvValue("=count(E2:E3)", undefined, undefined, column)).toBe("=count(E2:E3)");
    });

    it("should handle complex format function with sanitization", () => {
      const row = { value: 100 };
      const formatFn = (val: number, r: typeof row) => `=${r.value}+${val}`;
      const column: ITableColumn = {
        name: "test",
        label: "Test",
        field: "test",
        sanitizeCell: true,
      };
      expect(tableUtil._wrapCsvValue(50, formatFn, row, column)).toBe("'=100+50");
    });

    describe("date format comma removal", () => {
      it("should remove comma from date format M/dd/yyyy, HH:mm", () => {
        expect(tableUtil._wrapCsvValue("12/25/2023, 14:30")).toBe("12/25/2023 14:30");
      });

      it("should remove comma from date format with single digit month", () => {
        expect(tableUtil._wrapCsvValue("1/5/2023, 09:15")).toBe("1/5/2023 09:15");
      });

      it("should remove comma from date format with seconds", () => {
        expect(tableUtil._wrapCsvValue("12/25/2023, 14:30:45")).toBe("12/25/2023 14:30:45");
      });

      it("should remove comma from date format with milliseconds", () => {
        expect(tableUtil._wrapCsvValue("12/25/2023, 14:30:45:123")).toBe("12/25/2023 14:30:45:123");
      });

      it("should handle date format with single digit hour", () => {
        expect(tableUtil._wrapCsvValue("6/15/2023, 9:05")).toBe("6/15/2023 9:05");
      });

      it("should not modify date format without comma", () => {
        expect(tableUtil._wrapCsvValue("12/25/2023 14:30")).toBe("12/25/2023 14:30");
      });

      it("should wrap non-date strings with commas in quotes", () => {
        expect(tableUtil._wrapCsvValue("hello, world")).toBe('"hello, world"');
      });

      it("should wrap strings that look like dates but don't match exact format", () => {
        expect(tableUtil._wrapCsvValue("2023/12/25, 14:30")).toBe('"2023/12/25, 14:30"');
      });

      it("should handle date format comma removal with sanitization", () => {
        const column: ITableColumn = {
          name: "test",
          label: "Test",
          field: "test",
          sanitizeCell: true,
        };
        expect(tableUtil._wrapCsvValue("12/25/2023, 14:30", undefined, undefined, column)).toBe("'12/25/2023 14:30");
      });
    });
  });
});
