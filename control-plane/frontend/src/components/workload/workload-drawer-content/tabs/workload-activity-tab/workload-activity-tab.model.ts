/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ECustomCell, USER_DISPLAY_NAME_COL_DEFAULTS, type ITableColumn } from "@/models/table.model";
import type { WorkloadActionRecord, WorkloadActionType } from "@/swagger-models/workloads-client";
import { EColumnFilterType, type IAdvancedFilterOption } from "@/models/filter.model";
import { dateUtil } from "@/utils/date.util";

export const actionLabels = new Map<WorkloadActionType, string>([
  ["create", "Created"],
  ["update", "Updated"],
  ["delete", "Deleted"],
  ["suspend", "Suspended"],
  ["resume", "Resumed"],
]);

export const workloadActivityColumns: ITableColumn[] = [
  {
    name: "acceptedAt",
    label: "Date & time",
    field: "acceptedAt",
    align: "left",
    sortable: true,
    style: "width: 38%",
    format: (value: string) => {
      const date = new Date(value);
      return Number.isNaN(date.getTime()) ? "-" : dateUtil.dateAndTimeFormat(date);
    },
  },
  {
    name: "action",
    label: "Event",
    field: "action",
    align: "left",
    sortable: true,
    style: "width: 22%",
    format: (value: WorkloadActionType) => actionLabels.get(value) || value || "-",
  },
  {
    name: "subject",
    label: "Subject",
    field: "subject",
    align: "left",
    sortable: true,
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    style: "width: 40%; max-width: 0",
    customCellConfig: {
      [ECustomCell.USER_DISPLAY_NAME_COL]: {
        skipWhen: (row) => (row as WorkloadActionRecord).subjectType.toLowerCase() !== "user",
      },
    },
  },
];

export const workloadActivityFilterOptions: IAdvancedFilterOption[] = [
  { name: "acceptedAt", field: "acceptedAt", label: "Date & time", type: EColumnFilterType.Date },
  {
    name: "action",
    field: "action",
    label: "Event",
    type: EColumnFilterType.EnumBased,
    selectOptions: [...actionLabels].map(([value, label]) => ({ value, label })),
  },
  { name: "subject", field: "subject", label: "Subject", type: EColumnFilterType.FreeText },
];
