/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// models
import type { ITableColumn } from "@/models/table.model";
import { getField, getFieldContent } from "@/utils/table-format.util";
import { exportFile } from "quasar";
import { alertUtil } from "@/utils/alert.util";
import { Notify } from "quasar";
import type { IStatusColOptions } from "@/models/table.model";

export const tableUtil = {
  getDisplayColumns,
  customColumnSort,
  exportTableAsCsv,
  getStatusColOptions,
  ratioSort,

  // for unit tests
  _wrapCsvValue,
};

function getDisplayColumns(columns: Array<ITableColumn>, displayedColumns: Array<string> = []): Array<ITableColumn> {
  const displayedColumnsMap: Set<string> = new Set(displayedColumns);
  return columns.filter((col: ITableColumn) => displayedColumnsMap.has(col.name));
}

function customColumnSort<T>(
  column: ITableColumn,
  rowA: T,
  rowB: T,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  convertFunc: ((val: any) => any) | null = null,
): number {
  let aValue = getFieldContent(column, rowA);
  let bValue = getFieldContent(column, rowB);

  if (convertFunc) {
    aValue = convertFunc(aValue);
    bValue = convertFunc(bValue);
  }

  const isPercentage = (value: string): boolean => /^(\d+(\.\d+)?|\.\d+)%$/.test(value);

  const sortPercentages = (valueA: string, valueB: string): number => parseFloat(valueA) - parseFloat(valueB);

  switch (typeof aValue) {
    case "number":
      return aValue - bValue;

    case "string":
      if (isPercentage(aValue) && isPercentage(bValue)) {
        return sortPercentages(aValue, bValue);
      }
      return aValue.localeCompare(bValue);
    case "object":
      if (Array.isArray(aValue) && Array.isArray(bValue)) {
        return aValue.length - bValue.length;
      }
      return 0;
    default:
      return 0;
  }
}
//use when need to sort 'x/y' for e.g workload table runningPodsVsRequestedPods
function ratioSort(x1: number, y1: number, x2: number, y2: number): number {
  const number1 = x1 * 10 + y1;
  const number2 = x2 * 10 + y2;
  return number1 - number2;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function _wrapCsvValue(val: any, formatFn?: (val: any, row: any) => any, row?: any, column?: ITableColumn) {
  let formatted = formatFn !== void 0 ? formatFn(val, row) : val;

  formatted = formatted === void 0 || formatted === null ? "" : String(formatted);

  if (!formatted || formatted === "-") return "-";

  // Remove commas from date formats to prevent CSV parsing issues
  // Detect date formats like "M/dd/yyyy, HH:mm" and remove the comma
  const dateFormatRegex = /^\d{1,2}\/\d{1,2}\/\d{4},\s\d{1,2}:\d{2}(?::\d{2})?(?::\d{3})?$/;
  if (dateFormatRegex.test(formatted)) {
    formatted = formatted.replace(",", "");
  }

  // Only escape quotes if we're going to wrap the value
  if (column?.sanitizeCell) {
    formatted = `'${formatted}`;
  }

  // Wrap in quotes if value contains comma, newline, or quotes
  if (formatted.includes(",") || formatted.includes("\n") || formatted.includes('"')) {
    // Escape any existing quotes by doubling them
    formatted = formatted.replace(/"/g, '""');
    formatted = `"${formatted}"`;
  }

  return formatted;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function exportTableAsCsv(tableName: string, rows: any[], columns: ITableColumn[]): void {
  try {
    const content = [columns.filter((col) => col.hideExport !== true).map((col) => _wrapCsvValue(col.label))]
      .concat(
        rows.map((row) =>
          columns
            .filter((col) => col.hideExport !== true)
            .map((col) => {
              if (col.exportFormat) {
                return _wrapCsvValue(col.exportFormat(row), undefined, undefined, col);
              } else {
                return _wrapCsvValue(getField(col, row), col.format, row, col);
              }
            })
            .join(","),
        ),
      )
      .join("\r\n");

    const status = exportFile(`${tableName}.csv`, content, "text/csv");
    if (status !== true) {
      Notify.create(alertUtil.getError(status.message));
    } else {
      Notify.create(alertUtil.getSuccess("Report downloaded"));
    }
  } catch (e) {
    console.error(e);
    Notify.create(alertUtil.getError("Failed to export file"));
  }
}

function getStatusColOptions(statusOptions: IStatusColOptions, tooltipText?: string): IStatusColOptions {
  if (statusOptions) {
    return {
      status: statusOptions.status,
      tooltipText: tooltipText || statusOptions.tooltipText || "",
      color: statusOptions.color,
      displayAnimation: statusOptions.displayAnimation,
      filterKey: statusOptions.filterKey,
      statusUpdatedTimeInMs: statusOptions.statusUpdatedTimeInMs,
    };
  }
  return {
    status: "-",
    tooltipText: "",
    displayAnimation: false,
    filterKey: "status",
  };
}
