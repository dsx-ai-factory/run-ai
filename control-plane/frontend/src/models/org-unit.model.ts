import type { INodePoolsNameAndId } from "@/models/node-pool.model";
import type { ICustomHeaderRadioButtonsConfig } from "@/models/table.model";
import type { ISelectOption } from "@/models/global.model";
import type { Units } from "@/swagger-models/org-unit-service-client";
import { RankingMetricType, RankingOrderType } from "@/swagger-models/org-unit-service-client";
import type { IFilterBy } from "@/models/filter.model";

/** Overtime metric source - used when fetching metrics from org-unit-service instead of cluster-service */
export enum EOrgUnitMetricSourceType {
  Department = "department",
  Project = "project",
}

export interface IOrgUnitMetricSource {
  type: EOrgUnitMetricSourceType;
  id: string;
}

export interface IEmptyProjectModelConfig {
  nodePoolsNamesAndIds: INodePoolsNameAndId[];
  clusterId: string;
  isCpuEnabled: boolean;
  parentId?: string;
  defaultNodePools?: string[];
  memoryUnits?: Units;
}

export interface IEmptyDepartmentModelConfig {
  nodePoolsNamesAndIds: INodePoolsNameAndId[];
  clusterId: string;
  isCpuEnabled: boolean;
  defaultNodePools?: string[];
  memoryUnits?: Units;
}
/** Over time data **/
export enum EOrgUnitOverTimeValue {
  TwentyFourHours = "24h",
  SevenDays = "7d",
  ThirtyDays = "30d",
}

export const ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG = {
  label: "Show the last",
  options: [
    { label: "24 hours", value: EOrgUnitOverTimeValue.TwentyFourHours, shortLabel: "24h" },
    { label: "7 days", value: EOrgUnitOverTimeValue.SevenDays, shortLabel: "7d" },
    { label: "30 days", value: EOrgUnitOverTimeValue.ThirtyDays, shortLabel: "30d" },
  ],
  iconName: "ph-bold ph-calendar-dot q-mr-sm text-black-30",
  chipPrefix: "Last",
} as Partial<ICustomHeaderRadioButtonsConfig>;

export enum EOrgUnitOverTimeColumnName {
  AverageGpuUtilization = "averageGpuUtilization",
  AverageGpuAllocation = "averageGpuAllocation",
  AverageGpuMemoryUtilization = "averageGpuMemoryUtilization",
}

export const AVG_RESOURCE_TIMEFRAME_FILTER_KEY = "avgResourceTimeframe";

/** Ranking **/

export const NUMBER_OF_RANKING_ENTITIES_PER_PAGE = 5;

export interface IOrgUnitRankingTableFilterBy extends IFilterBy {
  rankingTimeframe: string;
}

export enum EOrgUnitEntityRankingValue {
  Department = "Department",
  Project = "Project",
}

export enum EOrgUnitRankingMetricValue {
  TopAverageGpuAllocation = "topAverageGpuAllocation",
  TopAverageGpuUtilization = "topAverageGpuUtilization",
  TopAverageGpuMemoryUtilization = "topAverageGpuMemoryUtilization",
  BottomAverageGpuAllocation = "bottomAverageGpuAllocation",
  BottomAverageGpuUtilization = "bottomAverageGpuUtilization",
  BottomAverageGpuMemoryUtilization = "bottomAverageGpuMemoryUtilization",
}

export interface IOrgUnitEntitySelectOption extends ISelectOption {
  value: EOrgUnitEntityRankingValue;
}

export const ORG_UNIT_RANKING_ENTITY_SELECT_OPTIONS: IOrgUnitEntitySelectOption[] = [
  { label: "Projects", value: EOrgUnitEntityRankingValue.Project },
  { label: "Departments", value: EOrgUnitEntityRankingValue.Department },
];

export interface IOrgUnitRankingMetricSelectOption extends ISelectOption {
  value: EOrgUnitRankingMetricValue;
}

export const ORG_UNIT_RANKING_METRIC_SELECT_OPTIONS: IOrgUnitRankingMetricSelectOption[] = [
  { label: "Highest Avg. GPU allocation", value: EOrgUnitRankingMetricValue.TopAverageGpuAllocation },
  { label: "Lowest Avg. GPU allocation", value: EOrgUnitRankingMetricValue.BottomAverageGpuAllocation },
  { label: "Highest  Avg. GPU utilization", value: EOrgUnitRankingMetricValue.TopAverageGpuUtilization },
  { label: "Lowest  Avg. GPU utilization", value: EOrgUnitRankingMetricValue.BottomAverageGpuUtilization },
  { label: "Highest Avg. GPU memory utilization", value: EOrgUnitRankingMetricValue.TopAverageGpuMemoryUtilization },
  { label: "Lowest Avg. GPU memory utilization", value: EOrgUnitRankingMetricValue.BottomAverageGpuMemoryUtilization },
];

export const rankingMetricTypeMap: Record<
  EOrgUnitRankingMetricValue,
  Record<EOrgUnitOverTimeValue, RankingMetricType>
> = {
  [EOrgUnitRankingMetricValue.TopAverageGpuAllocation]: {
    [EOrgUnitOverTimeValue.TwentyFourHours]: RankingMetricType.AvgGpuAllocation24h,
    [EOrgUnitOverTimeValue.SevenDays]: RankingMetricType.AvgGpuAllocation7d,
    [EOrgUnitOverTimeValue.ThirtyDays]: RankingMetricType.AvgGpuAllocation30d,
  },
  [EOrgUnitRankingMetricValue.TopAverageGpuUtilization]: {
    [EOrgUnitOverTimeValue.TwentyFourHours]: RankingMetricType.AvgGpuUtilization24h,
    [EOrgUnitOverTimeValue.SevenDays]: RankingMetricType.AvgGpuUtilization7d,
    [EOrgUnitOverTimeValue.ThirtyDays]: RankingMetricType.AvgGpuUtilization30d,
  },
  [EOrgUnitRankingMetricValue.BottomAverageGpuAllocation]: {
    [EOrgUnitOverTimeValue.TwentyFourHours]: RankingMetricType.AvgGpuAllocation24h,
    [EOrgUnitOverTimeValue.SevenDays]: RankingMetricType.AvgGpuAllocation7d,
    [EOrgUnitOverTimeValue.ThirtyDays]: RankingMetricType.AvgGpuAllocation30d,
  },
  [EOrgUnitRankingMetricValue.BottomAverageGpuUtilization]: {
    [EOrgUnitOverTimeValue.TwentyFourHours]: RankingMetricType.AvgGpuUtilization24h,
    [EOrgUnitOverTimeValue.SevenDays]: RankingMetricType.AvgGpuUtilization7d,
    [EOrgUnitOverTimeValue.ThirtyDays]: RankingMetricType.AvgGpuUtilization30d,
  },
  [EOrgUnitRankingMetricValue.TopAverageGpuMemoryUtilization]: {
    [EOrgUnitOverTimeValue.TwentyFourHours]: RankingMetricType.AvgGpuMemoryUtilization24h,
    [EOrgUnitOverTimeValue.SevenDays]: RankingMetricType.AvgGpuMemoryUtilization7d,
    [EOrgUnitOverTimeValue.ThirtyDays]: RankingMetricType.AvgGpuMemoryUtilization30d,
  },
  [EOrgUnitRankingMetricValue.BottomAverageGpuMemoryUtilization]: {
    [EOrgUnitOverTimeValue.TwentyFourHours]: RankingMetricType.AvgGpuMemoryUtilization24h,
    [EOrgUnitOverTimeValue.SevenDays]: RankingMetricType.AvgGpuMemoryUtilization7d,
    [EOrgUnitOverTimeValue.ThirtyDays]: RankingMetricType.AvgGpuMemoryUtilization30d,
  },
};

export const rankingOrderMap: Record<EOrgUnitRankingMetricValue, RankingOrderType> = {
  [EOrgUnitRankingMetricValue.TopAverageGpuAllocation]: RankingOrderType.Top,
  [EOrgUnitRankingMetricValue.TopAverageGpuUtilization]: RankingOrderType.Top,
  [EOrgUnitRankingMetricValue.BottomAverageGpuAllocation]: RankingOrderType.Bottom,
  [EOrgUnitRankingMetricValue.BottomAverageGpuUtilization]: RankingOrderType.Bottom,
  [EOrgUnitRankingMetricValue.TopAverageGpuMemoryUtilization]: RankingOrderType.Top,
  [EOrgUnitRankingMetricValue.BottomAverageGpuMemoryUtilization]: RankingOrderType.Bottom,
};

/** Quota Management */
export enum EQuotaTab {
  Scheduling = "scheduling",
  Quota = "quota",
}

export enum EEntityPriorityValue {
  Lowest = "Lowest",
  VeryLow = "VeryLow",
  Low = "Low",
  MediumLow = "MediumLow",
  Medium = "Medium",
  MediumHigh = "MediumHigh",
  High = "High",
  VeryHigh = "VeryHigh",
  Highest = "Highest",
}

export enum EEntityPriorityLabel {
  Lowest = "Lowest",
  VeryLow = "Very Low",
  Low = "Low",
  MediumLow = "Medium Low",
  Medium = "Medium",
  MediumHigh = "Medium High",
  High = "High",
  VeryHigh = "Very High",
  Highest = "Highest",
}

export const ENTITY_PRIORITY_DEFAULT_VALUE = EEntityPriorityValue.MediumLow;
export const entityPriorityMap: Record<string, EEntityPriorityLabel> = {
  [EEntityPriorityValue.Lowest]: EEntityPriorityLabel.Lowest,
  [EEntityPriorityValue.VeryLow]: EEntityPriorityLabel.VeryLow,
  [EEntityPriorityValue.Low]: EEntityPriorityLabel.Low,
  [EEntityPriorityValue.MediumLow]: EEntityPriorityLabel.MediumLow,
  [EEntityPriorityValue.Medium]: EEntityPriorityLabel.Medium,
  [EEntityPriorityValue.MediumHigh]: EEntityPriorityLabel.MediumHigh,
  [EEntityPriorityValue.High]: EEntityPriorityLabel.High,
  [EEntityPriorityValue.VeryHigh]: EEntityPriorityLabel.VeryHigh,
  [EEntityPriorityValue.Highest]: EEntityPriorityLabel.Highest,
};

export interface IEntityPrioritySelectOption extends ISelectOption {
  value: EEntityPriorityValue;
  label: EEntityPriorityLabel;
}

export const ENTITY_PRIORITY_SELECT_OPTIONS: IEntityPrioritySelectOption[] = [
  { label: EEntityPriorityLabel.Lowest, value: EEntityPriorityValue.Lowest },
  { label: EEntityPriorityLabel.VeryLow, value: EEntityPriorityValue.VeryLow },
  { label: EEntityPriorityLabel.Low, value: EEntityPriorityValue.Low },
  { label: EEntityPriorityLabel.MediumLow, value: EEntityPriorityValue.MediumLow },
  { label: EEntityPriorityLabel.Medium, value: EEntityPriorityValue.Medium },
  { label: EEntityPriorityLabel.MediumHigh, value: EEntityPriorityValue.MediumHigh },
  { label: EEntityPriorityLabel.High, value: EEntityPriorityValue.High },
  { label: EEntityPriorityLabel.VeryHigh, value: EEntityPriorityValue.VeryHigh },
  { label: EEntityPriorityLabel.Highest, value: EEntityPriorityValue.Highest },
];

/*** QUOTA INPUT ***/
export enum EQuotaInputState {
  Default = "Default",
  Error = "Error",
  Warning = "Warning",
}

/*** MEMORY UNITS ***/
export enum ECpuMemoryUnit {
  MB = "MB",
  GB = "GB",
  MiB = "Mib",
}
