/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { type Report, ReportStatus } from "@/swagger-models/org-unit-service-client";
import { ECustomCell, type ITableColumn } from "@/models/table.model";
import { EColumnFilterType } from "@/models/filter.model";
import { ReportFilterAndSortFields } from "@/swagger-models/metrics-service-client";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";
import { dateUtil } from "@/utils/date.util";
import { reportUtil } from "@/utils/report.util/report.util";

export type IReportTable = Report;

export const reportColumnName = {
  name: ReportFilterAndSortFields.Name,
  description: "description",
  status: ReportFilterAndSortFields.Status,
  type: "type",
  createdBy: ReportFilterAndSortFields.CreatedBy,
  createdAt: ReportFilterAndSortFields.CreatedAt,
  collectionPeriod: "collectionPeriod",
};

enum EReportColumnLabel {
  Name = "Report",
  Description = "Description",
  Status = "Status",
  Type = "Type",
  CreatedBy = "Created by",
  CreatedAt = "Creation time",
  CollectionPeriod = "Collection period",
}

export const allReportsColumnsMap: Record<keyof typeof reportColumnName, ITableColumn> = {
  name: {
    name: reportColumnName.name,
    label: EReportColumnLabel.Name,
    field: (report: Report): string => report.name,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.reports(ReportFilterAndSortFields.Name, (r) => r.name),
    },
  },
  description: {
    name: reportColumnName.description,
    label: EReportColumnLabel.Description,
    field: (report: Report): string => report.description || "-",
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  status: {
    name: reportColumnName.status,
    label: EReportColumnLabel.Status,
    field: (report: Report) => report,
    format: (report: Report) => reportUtil.getStatusColOptions(report),
    sortable: true,
    align: "left",
    customCell: ECustomCell.STATUS_COL,
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: filterUtil.parseEnumToSelectOptions(ReportStatus),
    },
  },
  type: {
    name: reportColumnName.type,
    label: EReportColumnLabel.Type,
    field: () => "Consumption",
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  createdAt: {
    name: reportColumnName.createdAt,
    label: EReportColumnLabel.CreatedAt,
    field: (report: Report): string => dateUtil.dateAndTimeFormat(new Date(report.createdAt)),
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.Date,
    },
  },
  createdBy: {
    name: reportColumnName.createdBy,
    label: EReportColumnLabel.CreatedBy,
    field: (report: Report): string => report.createdBy,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.reports(ReportFilterAndSortFields.CreatedBy, (r) => r.createdBy),
    },
  },
  collectionPeriod: {
    name: reportColumnName.collectionPeriod,
    label: EReportColumnLabel.CollectionPeriod,
    field: (report: Report): string => dateUtil.formatDateRange(report.start, report.end),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
};

export const reportIndexColumns: Array<ITableColumn> = [
  { ...allReportsColumnsMap.name, display: true, mandatory: true },
  { ...allReportsColumnsMap.description, display: true },
  { ...allReportsColumnsMap.status, display: true },
  { ...allReportsColumnsMap.type, display: true },
  { ...allReportsColumnsMap.createdBy, display: true },
  { ...allReportsColumnsMap.createdAt, display: true },
  { ...allReportsColumnsMap.collectionPeriod, display: true },
];
