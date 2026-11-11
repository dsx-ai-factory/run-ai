/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ECustomCell, USER_DISPLAY_NAME_COL_DEFAULTS, type ITableColumn } from "@/models/table.model";
import { type WorkloadActionRecord, WorkloadActionType } from "@/swagger-models/workloads-client";
import { EColumnFilterType, type IAdvancedFilterOption, type IFilterBy } from "@/models/filter.model";
import { EWorkloadActionSortField } from "@/models/workload-action.model";
import { dateUtil } from "@/utils/date.util";

export const actionLabels = new Map<WorkloadActionType, string>([
  [WorkloadActionType.Create, "Created"],
  [WorkloadActionType.Update, "Updated"],
  [WorkloadActionType.Delete, "Deleted"],
  [WorkloadActionType.Suspend, "Stopped"],
  [WorkloadActionType.Resume, "Resumed"],
]);

export const workloadActivityColumns: ITableColumn[] = [
  {
    name: EWorkloadActionSortField.AcceptedAt,
    label: "Date & time",
    field: EWorkloadActionSortField.AcceptedAt,
    align: "left",
    sortable: true,
    style: "width: 38%",
    format: (value: string) => {
      const date = new Date(value);
      return Number.isNaN(date.getTime()) ? "-" : dateUtil.dateAndTimeFormat(date);
    },
  },
  {
    name: EWorkloadActionSortField.Action,
    label: "Event",
    field: EWorkloadActionSortField.Action,
    align: "left",
    sortable: true,
    style: "width: 22%",
    format: (value: WorkloadActionType) => actionLabels.get(value) || value || "-",
  },
  {
    name: EWorkloadActionSortField.Subject,
    label: "Subject",
    field: EWorkloadActionSortField.Subject,
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
  {
    name: EWorkloadActionSortField.AcceptedAt,
    field: EWorkloadActionSortField.AcceptedAt,
    label: "Date & time",
    type: EColumnFilterType.Date,
  },
  {
    name: EWorkloadActionSortField.Action,
    field: EWorkloadActionSortField.Action,
    label: "Event",
    type: EColumnFilterType.EnumBased,
    selectOptions: [...actionLabels].map(([value, label]) => ({ value, label })),
  },
  {
    name: EWorkloadActionSortField.Subject,
    field: EWorkloadActionSortField.Subject,
    label: "Subject",
    type: EColumnFilterType.FreeText,
  },
];

export const WORKLOAD_ACTIONS_PAGE_SIZE = 50;

export function getDefaultWorkloadActionsFilters(): IFilterBy {
  return {
    page: 1,
    rowsPerPage: 0,
    rowsNumber: 0,
    sortBy: EWorkloadActionSortField.AcceptedAt,
    descending: true,
    advancedFilters: [],
  };
}
