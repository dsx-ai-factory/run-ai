/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ECustomCell, type ITableColumn, type IStatusColOptions } from "@/models/table.model";
import {
  EApplicationColumnLabel,
  EApplicationColumnName,
  ApplicationSortFilterFields,
  type IApplicationTable,
} from "@/models/ai-application.model";
import { EColumnFilterType, CLUSTER_COLUMN_FILTER_NAME } from "@/models/filter.model";
import { formatK8sResourceMemory, numberFormat } from "@/utils/format.util/format.util";
import { stringUtil } from "@/utils/string.util";
import { autocompleteUtil } from "@/utils/autocomplete.util";
import { aiApplicationUtil } from "@/utils/ai-application.util/ai-application.util";
import type { AiApplicationEndpoint } from "@/swagger-models/workloads-client";

export const allAIApplicationColumnsMap: Record<string, ITableColumn> = {
  name: {
    name: ApplicationSortFilterFields.Name,
    label: EApplicationColumnLabel.Name,
    field: (row: IApplicationTable) => row.name || "-",
    sortable: true,
    align: "left",
    display: true,
    mandatory: true,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.aiApplications(ApplicationSortFilterFields.Name, (app) => app.name),
    },
  },
  type: {
    name: ApplicationSortFilterFields.Type,
    label: EApplicationColumnLabel.Type,
    field: (row: IApplicationTable) => row.type || "-",
    sortable: true,
    align: "left",
    display: true,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.aiApplications(ApplicationSortFilterFields.Type, (app) => app.type),
    },
  },
  status: {
    name: "phase",
    label: EApplicationColumnLabel.Status,
    field: (row: IApplicationTable) => row,
    sortable: false,
    align: "left",
    display: true,
    format: (app: IApplicationTable): IStatusColOptions => aiApplicationUtil.getStatusColOptions(app),
    exportFormat: (app: IApplicationTable) => stringUtil.capitalizeString(app.phase || "Unknown"),
    customCell: ECustomCell.STATUS_COL,
    hideFilter: true,
  },
  department: {
    name: EApplicationColumnName.Department,
    label: EApplicationColumnLabel.Department,
    field: (row: IApplicationTable) => row.projectId || "-",
    sortable: false,
    align: "left",
    display: false,
    hideFilter: false,
    customCell: ECustomCell.APPLICATION_DEPARTMENT_COL,
  },
  project: {
    name: ApplicationSortFilterFields.ProjectName,
    label: EApplicationColumnLabel.Project,
    field: (row: IApplicationTable) => row.projectName || "-",
    sortable: true,
    align: "left",
    display: true,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.projects(),
    },
  },
  gpuComputeRequest: {
    name: EApplicationColumnName.GpuComputeRequest,
    label: EApplicationColumnLabel.GpuComputeRequest,
    field: (row: IApplicationTable) => row.requestedResources?.gpu?.request,
    sortable: false,
    align: "left",
    display: true,
    format: (val: number | null | undefined) => (val !== null && val !== undefined ? numberFormat(val) : "-"),
    hideFilter: true,
  },
  gpuComputeAllocation: {
    name: EApplicationColumnName.GpuComputeAllocation,
    label: EApplicationColumnLabel.GpuComputeAllocation,
    field: (row: IApplicationTable) => row.allocatedResources?.gpu,
    sortable: false,
    align: "left",
    display: true,
    format: (val: number | null | undefined) => (val !== null && val !== undefined ? numberFormat(val) : "-"),
    hideFilter: true,
  },
  gpuMemoryRequest: {
    name: EApplicationColumnName.GpuMemoryRequest,
    label: EApplicationColumnLabel.GpuMemoryRequest,
    field: (row: IApplicationTable) => row.requestedResources?.gpuMemory?.request,
    sortable: false,
    align: "left",
    display: true,
    format: (val: string | null | undefined) => (val ? formatK8sResourceMemory(val) : "-"),
    hideFilter: true,
  },
  gpuMemoryAllocation: {
    name: EApplicationColumnName.GpuMemoryAllocation,
    label: EApplicationColumnLabel.GpuMemoryAllocation,
    field: (row: IApplicationTable) => row.allocatedResources?.gpuMemory,
    sortable: false,
    align: "left",
    display: true,
    format: (val: string | null | undefined) => (val ? formatK8sResourceMemory(val) : "-"),
    hideFilter: true,
  },
  cpuComputeRequest: {
    name: EApplicationColumnName.CpuComputeRequest,
    label: EApplicationColumnLabel.CpuComputeRequest,
    field: (row: IApplicationTable) => row.requestedResources?.cpu?.request,
    sortable: false,
    align: "left",
    display: false,
    format: (val: number | null | undefined) => (val !== null && val !== undefined ? numberFormat(val) : "-"),
    hideFilter: true,
  },
  cpuComputeAllocation: {
    name: EApplicationColumnName.CpuComputeAllocation,
    label: EApplicationColumnLabel.CpuComputeAllocation,
    field: (row: IApplicationTable) => row.allocatedResources?.cpu,
    sortable: false,
    align: "left",
    display: false,
    format: (val: number | null | undefined) => (val !== null && val !== undefined ? numberFormat(val) : "-"),
    hideFilter: true,
  },
  cpuMemoryRequest: {
    name: EApplicationColumnName.CpuMemoryRequest,
    label: EApplicationColumnLabel.CpuMemoryRequest,
    field: (row: IApplicationTable) => row.requestedResources?.cpuMemory?.request,
    sortable: false,
    align: "left",
    display: false,
    format: (val: string | null | undefined) => (val ? formatK8sResourceMemory(val) : "-"),
    hideFilter: true,
  },
  cpuMemoryAllocation: {
    name: EApplicationColumnName.CpuMemoryAllocation,
    label: EApplicationColumnLabel.CpuMemoryAllocation,
    field: (row: IApplicationTable) => row.allocatedResources?.cpuMemory,
    sortable: false,
    align: "left",
    display: false,
    format: (val: string | null | undefined) => (val ? formatK8sResourceMemory(val) : "-"),
    hideFilter: true,
  },
  connections: {
    name: EApplicationColumnName.Connections,
    label: EApplicationColumnLabel.Connections,
    field: (row: IApplicationTable) => row.endpoints || [],
    sortable: false,
    align: "left",
    display: false,
    format: (endpoints: Array<AiApplicationEndpoint>): number[] | string[] | string => {
      const amount = endpoints?.length;
      if (amount === 0) return "-";
      if (amount === 1 && endpoints[0].url) {
        return [stringUtil.truncateString(endpoints[0].url, 30)];
      }
      return [amount];
    },
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "connections-clicked" },
    hideFilter: true,
  },
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: EApplicationColumnLabel.Cluster,
    field: (row: IApplicationTable) => row.clusterName || "-",
    sortable: true,
    align: "left",
    display: false,
    hideFilter: true,
  },
};

/**
 * Default columns for AI Applications table
 */
export const aiApplicationIndexColumns: Array<ITableColumn> = [
  { ...allAIApplicationColumnsMap.name },
  { ...allAIApplicationColumnsMap.type },
  { ...allAIApplicationColumnsMap.status },
  { ...allAIApplicationColumnsMap.department },
  { ...allAIApplicationColumnsMap.project },
  { ...allAIApplicationColumnsMap.gpuComputeRequest },
  { ...allAIApplicationColumnsMap.gpuComputeAllocation },
  { ...allAIApplicationColumnsMap.gpuMemoryRequest },
  { ...allAIApplicationColumnsMap.gpuMemoryAllocation },
  { ...allAIApplicationColumnsMap.cpuComputeRequest },
  { ...allAIApplicationColumnsMap.cpuComputeAllocation },
  { ...allAIApplicationColumnsMap.cpuMemoryRequest },
  { ...allAIApplicationColumnsMap.cpuMemoryAllocation },
  { ...allAIApplicationColumnsMap.connections },
  { ...allAIApplicationColumnsMap.cluster },
];

/**
 * Columns that depend on specific conditions
 */
export const aiApplicationDependentColumns = {
  cluster: new Set([CLUSTER_COLUMN_FILTER_NAME]),
  endpoints: new Set([EApplicationColumnName.Connections]),
};
