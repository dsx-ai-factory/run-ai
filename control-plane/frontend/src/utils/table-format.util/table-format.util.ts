/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// models
import type { TEntity } from "@/models/entity.model";
import type { ICustomCellConfig, ITableColumn } from "@/models/table.model";
import { MAX_STRING_LENGTH_BEFORE_TRUNCATION } from "@/models/global.model";

// services
import { numberFormat, numberWithPercentFormat, onlyPercentFormat } from "@/utils/format.util";
import { dateUtil } from "@/utils/date.util";
import { stringUtil } from "@/utils/string.util";
import type { ECustomCell } from "@/models/table.model";

export type TCalcPercentFormatFunc = (val: string | number, row: TEntity) => string;
export type TNumberFormatFunc = (val: string | number) => string;
export const tableFormatUtil = {
  getFormattedNumberOrDefault,
  tableNumberFormat,
  tableNumberWithPercentFormat,
  percentFormat,
  getMemoryUnitsForUI,
  getCpuCoresForUI,
  getFieldContent,
  getCustomCellConfig,
  getField,
  formatDate,
  formatArrayStrings,
  getFormattedDateOrDefault,
};
/**
 * Returns the input value if it is a number, otherwise returns "-".
 * @param val - The input value to check.
 * @returns The input value if it is a number, otherwise "-".
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function getFormattedNumberOrDefault(val: any): string {
  return typeof val === "number" ? numberFormat(val) : "-";
}

function getFormattedDateOrDefault(date?: string): string {
  return date ? dateUtil.dateAndTimeFormat(new Date(date)) : "-";
}

export function tableNumberFormat(suffix?: string): TNumberFormatFunc {
  return (val: string | number): string => {
    return numberFormat(val, undefined, suffix);
  };
}

export function tableNumberWithPercentFormat(compareField: string, isMemory = false): TCalcPercentFormatFunc {
  return (val: string | number, row: TEntity) => {
    const compareValue = +(row[compareField as keyof TEntity] as string);
    return numberWithPercentFormat(+val, compareValue, isMemory);
  };
}

export function percentFormat(compareField: string): TCalcPercentFormatFunc {
  return (val: string | number, row: TEntity) => {
    const compareValue = row[compareField as keyof TEntity];
    if (!compareValue) {
      return "-";
    }
    return onlyPercentFormat(val == null ? 0 : +val, +compareValue);
  };
}

export function getMemoryUnitsForUI(val: string): string {
  if (!val) return "";
  return `${val.slice(0, val.length - 1)} ${(val.slice(val.length - 1) + "b").toUpperCase()}`;
}

export function getCpuCoresForUI(cpuCores: number): string {
  if (typeof cpuCores !== "number") return "";
  let postfix = "cores";
  let amount = cpuCores;
  if (cpuCores < 1 && cpuCores !== 0) {
    amount *= 1000;
    postfix = "millicores";
  }
  return `${amount} ${postfix}`;
}

/* eslint-disable-next-line @typescript-eslint/no-explicit-any */
export function getFieldContent(column: ITableColumn, row: unknown): any {
  const val = getField(column, row as Record<string, unknown>);
  return column.format ? column.format(val, row) : val;
}

export function getCustomCellConfig(column: ITableColumn, customCellName: ECustomCell) {
  return column?.customCellConfig?.[customCellName as keyof ICustomCellConfig];
}
/* eslint-disable-next-line @typescript-eslint/no-explicit-any */
export function getField(column: ITableColumn, row: Record<string, unknown>): any {
  return typeof column.field === "function" ? column.field(row) : row[column.field];
}

export function formatDate(date: string | number) {
  if (!date) return "-";
  return dateUtil.dateAndTimeFormat(new Date(date));
}

export function formatArrayStrings(
  array: Array<string>,
  separator = ", ",
  maxLength = MAX_STRING_LENGTH_BEFORE_TRUNCATION,
): string {
  const string = array.join(separator);
  return stringUtil.truncateString(string, maxLength);
}
