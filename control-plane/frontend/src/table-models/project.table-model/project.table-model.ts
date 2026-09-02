/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { EProjectColumnName, type IProjectTable } from "@/models/project.model";
import { EResourceState, EResourceType } from "@/models/resource.model";
import {
  ECustomCell,
  ECustomHeader,
  type ICustomHeaderRadioButtonsConfig,
  type IStatusColOptions,
  type ITableColumn,
} from "@/models/table.model";
import { dateUtil } from "@/utils/date.util";
import { convertToBytes, memoryFormat, numberFormat, toPercent } from "@/utils/format.util";
import { resourceUtil } from "@/utils/resource.util";
import { CLUSTER_COLUMN_FILTER_NAME, EColumnFilterType } from "@/models/filter.model";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";
import { ProjectFilterSortFields, ProjectPhase, type Resources } from "@/swagger-models/org-unit-service-client";
import { orgUnitUtil } from "@/utils/org-unit.util";
import { viewLinkColumn } from "@/table-models/common-columns";
import {
  AVG_RESOURCE_TIMEFRAME_FILTER_KEY,
  EOrgUnitOverTimeColumnName,
  ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG,
} from "@/models/org-unit.model";
import { filterUtil } from "@/utils/filter.util/filter.util";

export enum EProjectColumnLabel {
  ProjectName = "Project",
  Description = "Description",
  DepartmentName = "Department",
  Status = "Status",
  NodePools = "Node pool(s)",
  AccessRules = "Subject(s)",
  GpuQuota = "GPU quota",
  CpuQuota = "CPU quota (Cores)",
  CpuMemoryQuota = "CPU memory quota",
  NodeAffinityTrain = "Node type (affinity) - Training",
  NodeAffinityInteractive = "Node type (affinity) - Workspaces",
  TrainingJobMaxIdleDurationSecs = "Idle GPU time limit - Training",
  InteractivePreemptibleJobMaxIdleDurationSecs = "Idle GPU time limit - preemptible workspaces",
  InteractiveJobMaxIdleDurationSecs = "Idle GPU time limit - non-preemptible workspaces",
  InteractiveJobTimeLimitSecs = "Workspaces time limit",
  TrainingJobTimeLimitSecs = "Training time limit",
  AllocatedCpu = "Allocated CPUs (Cores)",
  AllocatedMemory = "Allocated CPU memory",
  AllocatedGpus = "Allocated GPUs",
  AverageGpuAllocation = "Avg. GPU allocation",
  AverageGpuUtilization = "Avg. GPU utilization",
  AverageGpuMemoryUtilization = "Avg. GPU memory utilization",
  GpuUtilization = "GPU allocation ratio",
  CpuUtilization = "CPU allocation ratio",
  MemoryUtilization = "CPU memory allocation ratio",
  CreatedAt = "Creation time",
  Workloads = "Workload(s)",
  Cluster = "Cluster",
}

export const allProjectColumnsMap: Record<string, ITableColumn> = {
  projectName: {
    name: ProjectFilterSortFields.Name,
    label: EProjectColumnLabel.ProjectName,
    field: (row: IProjectTable) => row.name,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.projects(),
    },
  },
  description: {
    name: "description",
    label: EProjectColumnLabel.Description,
    field: (row: IProjectTable) => row.description,
    format: (val: string) => val || "-",
    sortable: false,
    align: "left",
    customCell: ECustomCell.DESCRIPTION_COL,
    hideFilter: true,
    sanitizeCell: true,
  },
  departmentName: {
    name: ProjectFilterSortFields.ParentName,
    label: EProjectColumnLabel.DepartmentName,
    field: (row: IProjectTable) => row?.parent?.name,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.departmentNamesViaProjects(),
    },
  },
  status: {
    name: ProjectFilterSortFields.Phase,
    label: EProjectColumnLabel.Status,
    field: (row: IProjectTable) => row,
    sortable: true,
    align: "left",
    format: (project: IProjectTable): IStatusColOptions => orgUnitUtil.getStatusColOptions(project),
    customCell: ECustomCell.STATUS_COL,
    exportFormat: (project: IProjectTable) => project?.status?.phase || "NA",
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: filterUtil.parseEnumToSelectOptions(ProjectPhase),
    },
    customCellEvent: { emitName: "status-clicked" },
  },
  nodePools: {
    name: EProjectColumnName.NodePools,
    label: EProjectColumnLabel.NodePools,
    field: (row: IProjectTable) => row.resources,
    sortable: false,
    align: "left",
    format: (resources: Resources[]): string[] => resources.map((resource) => resource.nodePool?.name || "") || [],
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "node-pools-clicked" },
    hideFilter: true,
  },
  accessRules: {
    name: EProjectColumnName.AccessRules,
    label: EProjectColumnLabel.AccessRules,
    field: (row: IProjectTable) => row.rolesNames,
    sortable: false,
    align: "left",
    customCell: ECustomCell.ROLE_ASSIGMENT_COL,
    customCellEvent: { emitName: "access-rules-clicked" },
    hideFilter: true,
    hideExport: true,
  },
  allocatedCpu: {
    name: EProjectColumnName.AllocatedCpu,
    label: EProjectColumnLabel.AllocatedCpu,
    field: (row: IProjectTable) => row.status?.quotaStatus?.allocated?.cpu,
    sortable: false,
    align: "left",
    format: (allocatedCpu: number | undefined) =>
      allocatedCpu ? resourceUtil.getResourceDisplayValue(allocatedCpu, EResourceType.CPU) : "-",
    hideFilter: true,
  },
  allocatedMemory: {
    name: EProjectColumnName.AllocatedMemory,
    label: EProjectColumnLabel.AllocatedMemory,
    field: (row: IProjectTable) => row.status?.quotaStatus?.allocated?.memory,
    sortable: false,
    align: "left",
    format: (allocatedMemory: number | undefined) =>
      allocatedMemory ? memoryFormat(convertToBytes(allocatedMemory, "MB")) : "-",
    hideFilter: true,
  },
  allocatedGpus: {
    name: ProjectFilterSortFields.GpuAllocated,
    label: EProjectColumnLabel.AllocatedGpus,
    field: (row: IProjectTable) => row.status?.quotaStatus?.allocated?.gpu,
    sortable: true,
    align: "left",
    format: (allocatedGpus: number | undefined) =>
      allocatedGpus ? resourceUtil.getResourceDisplayValue(allocatedGpus, EResourceType.GPU) : "-",
    filter: {
      type: EColumnFilterType.Numeric,
    },
  },
  averageGpuAllocation: {
    name: EOrgUnitOverTimeColumnName.AverageGpuAllocation,
    label: EProjectColumnLabel.AverageGpuAllocation,
    field: (row: IProjectTable) => row.averageGpuAllocation,
    sortable: true,
    align: "left",
    format: (averageGpuAllocation: number | null) => (averageGpuAllocation ? numberFormat(averageGpuAllocation) : "-"),
    customHeader: ECustomHeader.RadioButtonsTableHeader,
    customHeaderEventName: "avg-resource-timeframe-changed",
    customHeaderConfig: {
      ...ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG,
      selectedValueFilterByKey: AVG_RESOURCE_TIMEFRAME_FILTER_KEY,
    } as ICustomHeaderRadioButtonsConfig,
    filter: {
      type: EColumnFilterType.Numeric,
    },
  },
  averageGpuUtilization: {
    name: EOrgUnitOverTimeColumnName.AverageGpuUtilization,
    label: EProjectColumnLabel.AverageGpuUtilization,
    field: (row: IProjectTable) =>
      row.averageGpuUtilization !== null && row.averageGpuUtilization !== undefined
        ? parseFloat(row.averageGpuUtilization.toFixed(0))
        : null,
    format: (averageGpuUtilization: number | null) => orgUnitUtil.getProgressColOptions(averageGpuUtilization),
    customCell: ECustomCell.PROGRESS_COL,
    sortable: true,
    align: "left",
    customHeader: ECustomHeader.RadioButtonsTableHeader,
    customHeaderEventName: "avg-resource-timeframe-changed",
    customHeaderConfig: {
      ...ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG,
      selectedValueFilterByKey: AVG_RESOURCE_TIMEFRAME_FILTER_KEY,
    } as ICustomHeaderRadioButtonsConfig,
    filter: {
      type: EColumnFilterType.Numeric,
    },
  },
  averageGpuMemoryUtilization: {
    name: EOrgUnitOverTimeColumnName.AverageGpuMemoryUtilization,
    label: EProjectColumnLabel.AverageGpuMemoryUtilization,
    field: (row: IProjectTable) =>
      row.averageGpuMemoryUtilization !== null && row.averageGpuMemoryUtilization !== undefined
        ? parseFloat(row.averageGpuMemoryUtilization.toFixed(0))
        : null,
    format: (averageGpuMemoryUtilization: number | null) =>
      orgUnitUtil.getProgressColOptions(averageGpuMemoryUtilization),
    customCell: ECustomCell.PROGRESS_COL,
    sortable: true,
    align: "left",
    customHeader: ECustomHeader.RadioButtonsTableHeader,
    customHeaderEventName: "avg-resource-timeframe-changed",
    customHeaderConfig: {
      ...ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG,
      selectedValueFilterByKey: AVG_RESOURCE_TIMEFRAME_FILTER_KEY,
    } as ICustomHeaderRadioButtonsConfig,
    filter: {
      type: EColumnFilterType.Numeric,
    },
  },
  gpuUtilization: {
    name: EProjectColumnName.GpuUtilization,
    label: EProjectColumnLabel.GpuUtilization,
    field: (row: IProjectTable) => row.status?.quotaStatus?.allocated?.gpu,
    sortable: false,
    align: "left",
    format: (allocatedGpus: number | undefined, project: IProjectTable) =>
      orgUnitUtil.getAllocationRatioDisplayValue(allocatedGpus || 0, project?.totalResources),
    hideFilter: true,
  },
  cpuUtilization: {
    name: EProjectColumnName.CpuUtilization,
    label: EProjectColumnLabel.CpuUtilization,
    field: (row: IProjectTable) => row.status?.quotaStatus?.allocated?.cpu,
    sortable: false,
    align: "left",
    format: (allocatedCpu: number | undefined, project: IProjectTable) =>
      toPercent(allocatedCpu || 0, project?.totalResources?.cpuQuota || 0),
    hideFilter: true,
  },
  memoryUtilization: {
    name: EProjectColumnName.MemoryUtilization,
    label: EProjectColumnLabel.MemoryUtilization,
    field: (row: IProjectTable) => row.status?.quotaStatus?.allocated?.memory,
    sortable: true,
    align: "left",
    format: (allocatedMemory: number | undefined, project: IProjectTable) =>
      toPercent(resourceUtil.fromBytesToMiB(allocatedMemory || 0), project.totalResources?.memoryQuota || 0),
    hideFilter: true,
  },
  gpuQuota: {
    name: ProjectFilterSortFields.TotalGpuQuota,
    label: EProjectColumnLabel.GpuQuota,
    field: (row: IProjectTable) =>
      resourceUtil.getResourceDisplayValue(row?.totalResources?.gpuQuota, EResourceType.GPU) || 0,
    sortable: true,
    align: "left",
    exportFormat: (row: IProjectTable) => row.totalResources?.gpuQuota,
    filter: {
      type: EColumnFilterType.Numeric,
    },
  },
  cpuQuota: {
    name: EProjectColumnName.CpuQuota,
    label: EProjectColumnLabel.CpuQuota,
    field: (row: IProjectTable) => row.totalResources?.cpuQuota,
    sortable: false,
    align: "left",
    format: (cpuQuota: number | null | undefined) => resourceUtil.getResourceDisplayValue(cpuQuota, EResourceType.CPU),
    hideFilter: true,
    exportFormat: (row: IProjectTable) => row.totalResources?.cpuQuota,
  },
  cpuMemoryQuota: {
    name: EProjectColumnName.CpuMemoryQuota,
    label: EProjectColumnLabel.CpuMemoryQuota,
    field: (row: IProjectTable) => row.totalResources?.memoryQuota,
    sortable: false,
    align: "left",
    format: (memoryQuota: number | null | undefined) => {
      if (memoryQuota === null || memoryQuota === undefined || memoryQuota === -1) return EResourceState.Unlimited;
      const memoryQuotaInMib = resourceUtil.fromMbToMib(memoryQuota);
      return resourceUtil.getResourceDisplayValue(memoryQuotaInMib, EResourceType.MEMORY);
    },
    hideFilter: true,
    exportFormat: (row: IProjectTable) => row.totalResources?.memoryQuota,
  },
  nodeAffinityTrain: {
    name: EProjectColumnName.NodeAffinityTrain,
    label: EProjectColumnLabel.NodeAffinityTrain,
    field: (row: IProjectTable) => {
      return (
        row?.effective?.nodeTypes?.training?.map((nodeTypeId: string) => {
          return row.effective?.nodeTypes?.names?.[nodeTypeId];
        }) || []
      );
    },
    sortable: false,
    align: "left",
    customCell: ECustomCell.STRINGS_LIST_COL,
    hideFilter: true,
  },
  nodeAffinityInteractive: {
    name: EProjectColumnName.NodeAffinityInteractive,
    label: EProjectColumnLabel.NodeAffinityInteractive,
    field: (row: IProjectTable) => {
      return (
        row?.effective?.nodeTypes?.workspace?.map((nodeTypeId: string) => {
          return row.effective?.nodeTypes?.names?.[nodeTypeId];
        }) || []
      );
    },
    sortable: false,
    align: "left",
    customCell: ECustomCell.STRINGS_LIST_COL,
    hideFilter: true,
  },
  trainingJobMaxIdleDurationSecs: {
    name: EProjectColumnName.TrainingJobMaxIdleDurationSecs,
    label: EProjectColumnLabel.TrainingJobMaxIdleDurationSecs,
    field: (row: IProjectTable) => row?.effective?.schedulingRules?.trainingJobMaxIdleDurationSeconds,
    sortable: false,
    align: "left",
    format: (trainingJobMaxIdleDurationSecs) =>
      trainingJobMaxIdleDurationSecs ? dateUtil.formatDuration(trainingJobMaxIdleDurationSecs) : "-",
    hideFilter: true,
  },
  interactivePreemptibleJobMaxIdleDurationSecs: {
    name: EProjectColumnName.InteractivePreemptibleJobMaxIdleDurationSecs,
    label: EProjectColumnLabel.InteractivePreemptibleJobMaxIdleDurationSecs,
    field: (row: IProjectTable) => row?.effective?.schedulingRules?.interactiveJobPreemptIdleDurationSeconds,
    sortable: false,
    align: "left",
    format: (interactivePreemptibleJobMaxIdleDurationSecs) =>
      interactivePreemptibleJobMaxIdleDurationSecs
        ? dateUtil.formatDuration(interactivePreemptibleJobMaxIdleDurationSecs)
        : "-",
    hideFilter: true,
  },
  interactiveJobMaxIdleDurationSecs: {
    name: EProjectColumnName.InteractiveJobMaxIdleDurationSecs,
    label: EProjectColumnLabel.InteractiveJobMaxIdleDurationSecs,
    field: (row: IProjectTable) => row?.effective?.schedulingRules?.interactiveJobMaxIdleDurationSeconds,
    sortable: false,
    align: "left",
    format: (interactiveJobMaxIdleDurationSecs) =>
      interactiveJobMaxIdleDurationSecs ? dateUtil.formatDuration(interactiveJobMaxIdleDurationSecs) : "-",
    hideFilter: true,
  },
  interactiveJobTimeLimitSecs: {
    name: EProjectColumnName.InteractiveJobTimeLimitSecs,
    label: EProjectColumnLabel.InteractiveJobTimeLimitSecs,
    field: (row: IProjectTable) => row?.effective?.schedulingRules?.interactiveJobTimeLimitSeconds,
    sortable: false,
    align: "left",
    format: (interactiveJobTimeLimitSecs) =>
      interactiveJobTimeLimitSecs ? dateUtil.formatDuration(interactiveJobTimeLimitSecs) : "-",
    hideFilter: true,
  },
  trainingJobTimeLimitSecs: {
    name: EProjectColumnName.TrainingJobTimeLimitSecs,
    label: EProjectColumnLabel.TrainingJobTimeLimitSecs,
    field: (row: IProjectTable) => row?.effective?.schedulingRules?.trainingJobTimeLimitSeconds,
    sortable: false,
    align: "left",
    format: (trainingJobTimeLimitSecs) =>
      trainingJobTimeLimitSecs ? dateUtil.formatDuration(trainingJobTimeLimitSecs) : "-",
    hideFilter: true,
  },
  createdAt: {
    name: ProjectFilterSortFields.CreatedAt,
    label: EProjectColumnLabel.CreatedAt,
    field: (row: IProjectTable) => row.createdAt,
    sortable: true,
    align: "left",
    format: (createdAt: string) => (createdAt ? dateUtil.dateAndTimeFormat(new Date(createdAt)) : "-"),
    filter: {
      type: EColumnFilterType.Date,
    },
  },
  workloads: viewLinkColumn({
    name: EProjectColumnName.Workloads,
    label: EProjectColumnLabel.Workloads,
    emitName: "workloads-clicked",
  }),
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: EProjectColumnLabel.Cluster,
    field: (row: IProjectTable): string | undefined => row.clusterId,
    sortable: true,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
  },
};

export const allProjectColumns: Array<ITableColumn> = [
  allProjectColumnsMap.projectName,
  allProjectColumnsMap.departmentName,
  allProjectColumnsMap.status,
  allProjectColumnsMap.nodePools,
  allProjectColumnsMap.accessRules,
  allProjectColumnsMap.gpuQuota,
  allProjectColumnsMap.cpuQuota,
  allProjectColumnsMap.cpuMemoryQuota,
  allProjectColumnsMap.createdAt,
  allProjectColumnsMap.nodeAffinityTrain,
  allProjectColumnsMap.nodeAffinityInteractive,
  allProjectColumnsMap.interactiveJobMaxIdleDurationSecs,
  allProjectColumnsMap.trainingJobMaxIdleDurationSecs,
  allProjectColumnsMap.interactivePreemptibleJobMaxIdleDurationSecs,
  allProjectColumnsMap.interactiveJobTimeLimitSecs,
  allProjectColumnsMap.trainingJobTimeLimitSecs,
];

export const projectIndexColumns: Array<ITableColumn> = [
  { ...allProjectColumnsMap.projectName, display: true, mandatory: true },
  { ...allProjectColumnsMap.description, display: false },
  { ...allProjectColumnsMap.departmentName, display: true },
  { ...allProjectColumnsMap.cluster, display: true },
  { ...allProjectColumnsMap.status, display: true },
  { ...allProjectColumnsMap.nodePools, display: true },
  { ...allProjectColumnsMap.accessRules, display: true },
  { ...allProjectColumnsMap.gpuQuota, display: true },
  { ...allProjectColumnsMap.allocatedGpus, display: true },
  { ...allProjectColumnsMap.averageGpuAllocation, display: true },
  { ...allProjectColumnsMap.averageGpuUtilization, display: true },
  { ...allProjectColumnsMap.averageGpuMemoryUtilization, display: true },
  { ...allProjectColumnsMap.allocatedCpu, display: false },
  { ...allProjectColumnsMap.allocatedMemory, display: false },
  { ...allProjectColumnsMap.gpuUtilization, display: false },
  { ...allProjectColumnsMap.cpuUtilization, display: false },
  { ...allProjectColumnsMap.memoryUtilization, display: false },
  { ...allProjectColumnsMap.cpuQuota, display: false },
  { ...allProjectColumnsMap.cpuMemoryQuota, display: false },
  { ...allProjectColumnsMap.nodeAffinityTrain, display: false },
  { ...allProjectColumnsMap.nodeAffinityInteractive, display: false },
  { ...allProjectColumnsMap.trainingJobMaxIdleDurationSecs, display: false },
  { ...allProjectColumnsMap.interactivePreemptibleJobMaxIdleDurationSecs, display: false },
  { ...allProjectColumnsMap.interactiveJobMaxIdleDurationSecs, display: false },
  { ...allProjectColumnsMap.trainingJobTimeLimitSecs, display: false },
  { ...allProjectColumnsMap.interactiveJobTimeLimitSecs, display: false },
  { ...allProjectColumnsMap.workloads, display: true },
  { ...allProjectColumnsMap.createdAt, display: true },
];

export const projectMiniTableColumns = [
  { ...allProjectColumnsMap.projectName, display: true },
  { ...allProjectColumnsMap.gpuQuota, display: true },
  { ...allProjectColumnsMap.cpuQuota, display: true },
  { ...allProjectColumnsMap.cpuMemoryQuota, display: true },
  { ...allProjectColumnsMap.allocatedGpus, display: true },
  { ...allProjectColumnsMap.allocatedCpu, display: true },
  { ...allProjectColumnsMap.allocatedMemory, display: true },
  { ...allProjectColumnsMap.gpuUtilization, display: true },
  { ...allProjectColumnsMap.cpuUtilization, display: true },
];

//columns that displayed depends on feature flags or versions
export const projectDependentColumns = {
  cpu: new Set([
    allProjectColumnsMap.cpuMemoryQuota.name,
    allProjectColumnsMap.cpuQuota.name,
    allProjectColumnsMap.memoryUtilization.name,
    allProjectColumnsMap.cpuUtilization.name,
    allProjectColumnsMap.allocatedCpu.name,
    allProjectColumnsMap.allocatedMemory.name,
  ]),
  department: new Set([allProjectColumnsMap.departmentName.name]),
  nodePools: new Set([allProjectColumnsMap.nodePools.name]),
  accessRules: new Set([allProjectColumnsMap.accessRules.name]),
  workloads: new Set([allProjectColumnsMap.workloads.name]),
  cluster: new Set([allProjectColumnsMap.cluster.name]),
};
