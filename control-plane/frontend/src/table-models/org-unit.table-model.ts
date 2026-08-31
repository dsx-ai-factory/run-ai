/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ITableColumn } from "@/models/table.model";
import {
  type DepartmentGetRanking,
  DepartmentNodePoolResourcesFilterSortFields,
  type ProjectGetRanking,
  ProjectNodePoolResourcesFilterSortFields,
} from "@/swagger-models/org-unit-service-client";
import { numberFormat } from "@/utils/format.util";
import { ECustomCell, ECustomHeader, type ICustomHeaderRadioButtonsConfig } from "@/models/table.model";
import { orgUnitUtil } from "@/utils/org-unit.util";
import { resourceUtil } from "@/utils/resource.util";
import type { EQuotaEntity } from "@/models/resource.model";
import { EResourceType } from "@/models/resource.model";
import { ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG, ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";
import { EColumnFilterType } from "@/models/filter.model";

/** Ranking **/
export enum EOrgUnitRankingColumnName {
  AllocationValue = "allocationValue",
  UtilizationValue = "utilizationValue",
  MemoryUtilizationValue = "memoryUtilizationValue",
  GpuQuota = "gpuQuota",
  PendingWorkloadsCount = "pendingWorkloadsCount",
  RunningWorkloadsCount = "runningWorkloadsCount",
}

export enum EOrgUnitProjectsRankingColumnName {
  Project = "project",
  ParentName = "parentName",
}

export enum EOrgUnitDepartmentsRankingColumnName {
  Department = "department",
}

export enum EOrgUnitRankingColumnLabel {
  AllocationValue = "Avg. GPU devices allocation",
  UtilizationValue = "Avg. GPU utilization",
  MemoryUtilizationValue = "Avg. GPU memory utilization",
  GpuQuota = "GPU quota",
  PendingWorkloadsCount = "Pending workloads",
  RunningWorkloadsCount = "Running workloads",
}

export enum EOrgUnitProjectsRankingColumnLabel {
  Project = "Project",
  ParentName = "Department",
}

export enum EOrgUnitDepartmentsRankingColumnLabel {
  Department = "Department",
}

export type IOrgUnitRankingTable = Pick<
  ProjectGetRanking | DepartmentGetRanking,
  | "allocationValue"
  | "utilizationValue"
  | "memoryUtilizationValue"
  | "name"
  | "gpuQuota"
  | "pendingWorkloadsCount"
  | "runningWorkloadsCount"
>;

export const allOrgUnitRankingColumnsMap: Record<string, ITableColumn> = {
  [EOrgUnitProjectsRankingColumnName.Project]: {
    name: EOrgUnitProjectsRankingColumnName.Project,
    label: EOrgUnitProjectsRankingColumnLabel.Project,
    field: (row: IOrgUnitRankingTable) => row.name,
    display: true,
    align: "left",
    customCell: ECustomCell.LINK_COL,
    customCellEvent: { emitName: "project-clicked" },
    customCellConfig: {
      [ECustomCell.LINK_COL]: {
        tooltipText: "Go to Projects to view more details",
      },
    },
  },
  [EOrgUnitProjectsRankingColumnName.ParentName]: {
    name: EOrgUnitProjectsRankingColumnName.ParentName,
    label: EOrgUnitProjectsRankingColumnLabel.ParentName,
    field: (row: ProjectGetRanking) => row.parentName,
    display: true,
    align: "left",
  },
  [EOrgUnitDepartmentsRankingColumnName.Department]: {
    name: EOrgUnitDepartmentsRankingColumnName.Department,
    label: EOrgUnitDepartmentsRankingColumnLabel.Department,
    field: (row: IOrgUnitRankingTable) => row.name,
    display: true,
    align: "left",
    customCell: ECustomCell.LINK_COL,
    customCellEvent: { emitName: "department-clicked" },
    customCellConfig: {
      [ECustomCell.LINK_COL]: {
        tooltipText: "Go to Departments to view more details",
      },
    },
  },
  [EOrgUnitRankingColumnName.AllocationValue]: {
    name: EOrgUnitRankingColumnName.AllocationValue,
    label: EOrgUnitRankingColumnLabel.AllocationValue,
    field: (row: IOrgUnitRankingTable) =>
      row.allocationValue !== null ? parseFloat(numberFormat(row.allocationValue)) : "-",
    display: true,
    align: "left",
    customHeader: ECustomHeader.RadioButtonsTableHeader,
    customHeaderEventName: "ranking-timeframe-changed",
    customHeaderConfig: {
      ...ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG,
      selectedValueFilterByKey: "rankingTimeframe",
    } as ICustomHeaderRadioButtonsConfig,
  },
  [EOrgUnitRankingColumnName.UtilizationValue]: {
    name: EOrgUnitRankingColumnName.UtilizationValue,
    label: EOrgUnitRankingColumnLabel.UtilizationValue,
    field: (row: IOrgUnitRankingTable) =>
      row.utilizationValue !== null ? parseFloat(row.utilizationValue.toFixed(0)) : null,
    format: (utilizationValue: number | null) => orgUnitUtil.getProgressColOptions(utilizationValue),
    customCell: ECustomCell.PROGRESS_COL,
    display: true,
    align: "left",
    customHeader: ECustomHeader.RadioButtonsTableHeader,
    customHeaderEventName: "ranking-timeframe-changed",
    customHeaderConfig: {
      ...ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG,
      selectedValueFilterByKey: "rankingTimeframe",
    } as ICustomHeaderRadioButtonsConfig,
  },
  [EOrgUnitRankingColumnName.MemoryUtilizationValue]: {
    name: EOrgUnitRankingColumnName.MemoryUtilizationValue,
    label: EOrgUnitRankingColumnLabel.MemoryUtilizationValue,
    field: (row: IOrgUnitRankingTable) =>
      row.memoryUtilizationValue !== null && row.memoryUtilizationValue !== undefined
        ? parseFloat(row.memoryUtilizationValue.toFixed(0))
        : null,
    format: (memoryUtilizationValue: number | null) => orgUnitUtil.getProgressColOptions(memoryUtilizationValue),
    customCell: ECustomCell.PROGRESS_COL,
    display: true,
    align: "left",
    customHeader: ECustomHeader.RadioButtonsTableHeader,
    customHeaderEventName: "ranking-timeframe-changed",
    customHeaderConfig: {
      ...ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG,
      selectedValueFilterByKey: "rankingTimeframe",
    } as ICustomHeaderRadioButtonsConfig,
  },
  [EOrgUnitRankingColumnName.PendingWorkloadsCount]: {
    name: EOrgUnitRankingColumnName.PendingWorkloadsCount,
    label: EOrgUnitRankingColumnLabel.PendingWorkloadsCount,
    field: (row: IOrgUnitRankingTable) => row.pendingWorkloadsCount || 0,
    display: true,
    align: "left",
  },
  [EOrgUnitRankingColumnName.RunningWorkloadsCount]: {
    name: EOrgUnitRankingColumnName.RunningWorkloadsCount,
    label: EOrgUnitRankingColumnLabel.RunningWorkloadsCount,
    field: (row: IOrgUnitRankingTable) => row.runningWorkloadsCount || 0,
    display: true,
    align: "left",
  },
  [EOrgUnitRankingColumnName.GpuQuota]: {
    name: EOrgUnitRankingColumnName.GpuQuota,
    label: EOrgUnitRankingColumnLabel.GpuQuota,
    field: (row: IOrgUnitRankingTable) => resourceUtil.getResourceDisplayValue(row.gpuQuota, EResourceType.GPU) || 0,
    display: true,
    align: "left",
  },
};

export const projectsRankingColumns: ITableColumn[] = [
  { ...allOrgUnitRankingColumnsMap.project },
  { ...allOrgUnitRankingColumnsMap.parentName },
  { ...allOrgUnitRankingColumnsMap.gpuQuota },
  { ...allOrgUnitRankingColumnsMap.allocationValue },
  { ...allOrgUnitRankingColumnsMap.utilizationValue },
  { ...allOrgUnitRankingColumnsMap.memoryUtilizationValue },
  { ...allOrgUnitRankingColumnsMap.pendingWorkloadsCount },
  { ...allOrgUnitRankingColumnsMap.runningWorkloadsCount },
];

export const departmentsRankingColumns: ITableColumn[] = [
  { ...allOrgUnitRankingColumnsMap.department },
  { ...allOrgUnitRankingColumnsMap.gpuQuota },
  { ...allOrgUnitRankingColumnsMap.allocationValue },
  { ...allOrgUnitRankingColumnsMap.utilizationValue },
  { ...allOrgUnitRankingColumnsMap.memoryUtilizationValue },
  { ...allOrgUnitRankingColumnsMap.pendingWorkloadsCount },
  { ...allOrgUnitRankingColumnsMap.runningWorkloadsCount },
];

/**
 * ============================================
 * GPU Quota Step - Other Projects/Departments Table
 * ============================================
 */

/**
 * Row interface for quota table
 */
export interface IQuotaTableRow {
  id: string;
  name: string;
  gpuQuota: number | null;
  quotaDistribution: number;
  serverGpuQuota?: number | null; // Original server value before pending changes
  serverGpuLimit?: number | null; // Original GPU limit (null or -1 for unlimited)
  serverOverQuotaWeight?: number | null; // Original Over-Quota Weight
  allocatedNonPreemptibleGpu?: number | null; // For projects - allocated non-preemptible GPUs
  // for departments - projectsAggregatedResources - currently return only in get by id and not return as part of the node pool resources api
}

/**
 * Configuration for GPU Quota column in quota tables
 */
export interface ICustomCellGpuQuotaColConfig {
  availableQuota: number | null;
  totalQuota: number | null;
  nodePoolName: string;
  nodePoolId: string;
  entityType: EQuotaEntity;
  isDefaultDepartmentUnlimited?: boolean; // When true, show exceeded warning without amounts for departments
  onSave: (payload: {
    projectId: string;
    entityName: string;
    nodePoolName: string;
    nodePoolId: string;
    gpuQuota: number | null;
    originalValue: number | null;
    originalLimit: number | null;
    originalOverQuotaWeight: number | null;
  }) => void;
}

/**
 * Custom cell config for editable weight column
 */
export interface ICustomCellEditableWeightColConfig {
  nodePoolId: string;
  onSave: (payload: { orgUnitId: string; nodePoolId: string; weight: number; originalWeight: number | null }) => void;
  disabled?: boolean;
  disabledTooltip?: string;
}

/**
 * Custom cell config for weight distribution column
 */
export interface ICustomCellWeightDistributionColConfig {
  totalWeightInRank: number;
  entityType?: EQuotaEntity;
}

/**
 * Configuration for Quota Distribution column in quota tables
 */
export interface ICustomCellQuotaDistributionColConfig {
  totalQuota: number | null;
  isUnlimited: boolean; // True when parent quota is unlimited (null or -1)
  isDefaultDepartmentUnlimited?: boolean; // True when the "default" department has unlimited quota
  parentName?: string;
  totalAssignedQuota: number; // Sum of all projects' quotas (including pending changes)
  nodePoolName: string;
  currentEntityId: string;
  entityType: EQuotaEntity;
}

/**
 * Fetch context for project quota table
 */
export interface IProjectQuotaFetchContext {
  departmentId: string;
  currentProjectId: string;
  nodePoolName: string;
  clusterId: string;
}

/**
 * Fetch context for department quota table
 */
export interface IDepartmentQuotaFetchContext {
  currentDepartmentId: string;
  nodePoolName: string;
  clusterId: string;
}

/**
 * Generic fetch context (union type)
 */
export type QuotaTableFetchContext = IProjectQuotaFetchContext | IDepartmentQuotaFetchContext;

/**
 * Entity type for quota table
 */
export type QuotaTableEntityType = EQuotaEntity;

/**
 * Static project quota columns
 */
export const projectQuotaColumns: ITableColumn[] = [
  {
    name: ProjectNodePoolResourcesFilterSortFields.OrgUnitName,
    label: "Project",
    field: "name",
    sortable: true,
    align: "left",
  },
  {
    name: ProjectNodePoolResourcesFilterSortFields.GpuDeserved,
    label: "GPUs",
    field: "gpuQuota",
    sortable: false,
    align: "left",
    customCell: ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL,
  },
  {
    name: "quotaDistribution",
    label: "Quota distribution",
    field: "quotaDistribution",
    sortable: false,
    align: "left",
    customCell: ECustomCell.QUOTA_DISTRIBUTION_COL,
  },
];

/**
 * Static department quota columns
 */
export const departmentQuotaColumns: ITableColumn[] = [
  {
    name: DepartmentNodePoolResourcesFilterSortFields.OrgUnitName,
    label: "Department",
    field: "name",
    sortable: true,
    align: "left",
  },
  {
    name: "gpuQuota",
    label: "GPUs",
    field: "gpuQuota",
    sortable: false,
    align: "left",
    customCell: ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL,
  },
  {
    name: "quotaDistribution",
    label: "Quota distribution",
    field: "quotaDistribution",
    sortable: false,
    align: "left",
    customCell: ECustomCell.QUOTA_DISTRIBUTION_COL,
  },
];

/**
 * ============================================
 * Peer Org Unit Rank List
 * ============================================
 */

/**
 * Column names for peer rank list
 */
export enum EPeerOrgUnitRankColumnName {
  Name = "name",
  Rank = "rank",
  Weight = "weight",
}

/**
 * Column labels for peer rank list
 */
export enum EPeerOrgUnitRankColumnLabel {
  ProjectName = "Project",
  DepartmentName = "Department",
  Rank = "Rank",
  Weight = "Weight",
}

/**
 * Column definitions for peer org unit rank list
 */
export const peerProjectRankColumns: ITableColumn[] = [
  {
    name: ProjectNodePoolResourcesFilterSortFields.OrgUnitName,
    label: EPeerOrgUnitRankColumnLabel.ProjectName,
    field: "name",
    sortable: false,
    align: "left",
    display: true,
    filter: {
      type: EColumnFilterType.FreeText,
    },
  },
  {
    name: EPeerOrgUnitRankColumnName.Rank,
    label: EPeerOrgUnitRankColumnLabel.Rank,
    field: "rank",
    sortable: false,
    align: "left",
    display: true,
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: ENTITY_PRIORITY_SELECT_OPTIONS,
    },
  },
  {
    name: EPeerOrgUnitRankColumnName.Weight,
    label: EPeerOrgUnitRankColumnLabel.Weight,
    field: "weight",
    sortable: false,
    align: "left",
    display: true,
    hideFilter: true,
  },
];

/**
 * Column definitions for peer department rank list
 */
export const peerDepartmentRankColumns: ITableColumn[] = [
  {
    name: DepartmentNodePoolResourcesFilterSortFields.OrgUnitName,
    label: EPeerOrgUnitRankColumnLabel.DepartmentName,
    field: "name",
    sortable: false,
    align: "left",
    display: true,
    filter: {
      type: EColumnFilterType.FreeText,
    },
  },
  {
    name: EPeerOrgUnitRankColumnName.Rank,
    label: EPeerOrgUnitRankColumnLabel.Rank,
    field: "rank",
    sortable: false,
    align: "left",
    display: true,
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: ENTITY_PRIORITY_SELECT_OPTIONS,
    },
  },
  {
    name: EPeerOrgUnitRankColumnName.Weight,
    label: EPeerOrgUnitRankColumnLabel.Weight,
    field: "weight",
    sortable: false,
    align: "left",
    display: true,
    hideFilter: true,
  },
];
