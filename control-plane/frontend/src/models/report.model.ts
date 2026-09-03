/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ReportGroupByOptions, ReportStatus } from "@/swagger-models/org-unit-service-client";
import type { IStatusColOptions } from "@/models/table.model";

export const ReportStatusMap: Record<ReportStatus, IStatusColOptions> = {
  [ReportStatus.Pending]: {
    status: "Pending",
    displayAnimation: false,
    color: "warning",
  },
  [ReportStatus.Ready]: {
    status: "Ready",
    color: "success",
    displayAnimation: false,
  },
  [ReportStatus.Failed]: {
    status: "Failed",
    color: "negative",
    displayAnimation: false,
  },
  [ReportStatus.Processing]: {
    status: "Processing...",
    displayAnimation: true,
  },
  [ReportStatus.PendingDelete]: {
    status: "Deleting...",
    displayAnimation: true,
  },
  [ReportStatus.Deleting]: {
    status: "Deleting...",
    displayAnimation: true,
  },
  [ReportStatus.DeletionFailed]: {
    status: "Deletion failed",
    color: "negative",
    displayAnimation: false,
  },
};

export enum EReportExportedDocumentColumn {
  Project = "Project",
  Department = "Department",
  Cluster = "Cluster",
  GpuAllocationHours = "GPU allocation hours",
  InQuotaGpuHoursTitle = "GPU deserved quota hours",
  OverQuotaGpuHoursTitle = "GPU over quota hours",
  GpuIdleHours = "GPU idle hours",
  CpuAllocationHours = "CPU allocation hours",
  CpuMemoryAllocationHours = "CPU memory allocation hours",
  CpuUsageHours = "CPU usage hours",
  CpuMemoryUsageHours = "CPU memory usage hours",
}

export const ReportExportedDocumentColumns = Object.values(EReportExportedDocumentColumn);

export const hiddenExportedColumnsByReportGroupBy: Record<ReportGroupByOptions, string[]> = {
  [ReportGroupByOptions.Cluster]: [
    EReportExportedDocumentColumn.Cluster,
    EReportExportedDocumentColumn.Department,
    EReportExportedDocumentColumn.Project,
  ],
  [ReportGroupByOptions.Department]: [EReportExportedDocumentColumn.Department, EReportExportedDocumentColumn.Project],
  [ReportGroupByOptions.Project]: [EReportExportedDocumentColumn.Project],
  [ReportGroupByOptions.Nodepool]: [],
};

export enum EReportAction {
  Delete = "Delete",
  Download = "Download",
  Create = "Create",
}

export enum EReportActionTooltip {
  NoDeletePermission = "You don't have permission to delete this report.",
}

export enum EReportType {
  Consumption = "Consumption",
}
