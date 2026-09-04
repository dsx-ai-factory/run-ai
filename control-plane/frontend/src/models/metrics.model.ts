/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IColumnChartSeries } from "@/models/chart.model";

export enum EChartDataType {
  QUOTA = "quota",
  ALLOCATED = "allocated",
  PENDING = "pending",
}

export enum EDrillDownType {
  Clusters,
  Departments,
  Projects,
}

export const DRILL_DOWN_TITLES = ["Clusters", "Departments", "Projects"];

export const DRILL_DOWN_TITLES_DEPARTMENTS = ["Departments", "Projects"];

export const DRILL_DOWN_TITLES_PROJECTS = ["Clusters", "Projects", ""];
export interface IResourceData {
  numberOfPendingWorkloads: number;
  gpu: {
    quota: number;
    allocated: number;
  };
  cpu: {
    quota: number;
    allocated: number;
  };
  memory: {
    quota: number;
    allocated: number;
  };
  nodepoolName: string;
  departmentName?: string;
  projectName?: string;
  timestamp?: string;
  projectId?: string;
  departmentId?: string;
}

export interface TimeRangeData {
  resources: IResourceData[];
}

export interface CurrentData {
  resources: IResourceData[];
  projectResources: IResourceData[];
}

export interface Metadata {
  clusterId: string;
  clusterName: string;
  projectName?: string;
  projectId?: string;
  departmentName?: string;
  departmentId?: number;
  nodepoolName?: string;
}

export interface MetricsData {
  metadata: Metadata;
  current: CurrentData;
  timeRange: TimeRangeData;
}

export interface DepartmentsMetricData {
  data: MetricsData[];
  timeRange: TimeRangeData;
}

export interface MetricsFilter {
  department?: string;
  project?: string;
  nodePool?: string;
}

export interface ChartDataItemBase {
  name: string;
  drilldown?: string | boolean;
}

export interface ChartDataMapItem extends ChartDataItemBase {
  quotaY: number;
  allocatedY: number;
  pendingY?: number;
}

export interface ChartDataArrayItem extends ChartDataItemBase {
  y: number;
  optionValues?: Array<number>;
  meta?: {
    allocated?: number;
    correctName?: string;
    entityId?: string;
    groupByKey?: string;
    quota?: number;
  };
}

export interface IChartItemConfigBase {
  id: string;
  name: string;
}

export interface IChartMapItemConfig extends IChartItemConfigBase {
  data: Record<string, ChartDataMapItem>;
}

export interface IChartArrayItemConfig extends IChartItemConfigBase {
  data: Array<ChartDataArrayItem>;
  meta?: { correctName?: string } & Record<string, unknown>;
}

export interface DateRange {
  startDate: string;
  endDate: string;
}

export interface ITimeRangeChartData {
  columnSeries: IColumnChartSeries;
  splineSeries: IColumnChartSeries;
  categories: Array<string>;
}

export interface IAllocationItem {
  allocation: number;
  quota: number;
}
