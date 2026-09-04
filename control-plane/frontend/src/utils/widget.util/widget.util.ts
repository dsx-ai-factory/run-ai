/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { TelemetryResponseValuesInner } from "@/swagger-models/workloads-client";
import type { TelemetryResponseValuesInnerGroupsInner } from "@/swagger-models/workloads-client";
import { ENodeTelemetryGroupBy } from "@/models/cluster.model";
import type { IResourceData, MetricsData } from "@/models/metrics.model";
import { deepCopy } from "@/utils/common.util";
import {
  EWorkloadTelemetryGroupBy,
  type EWorkloadTelemetryGroupByWithoutBucket,
  type EWorkloadTelemetryPointMeta,
  EWorkloadType,
} from "@/models/workload.model";
import type { GroupedMeasurementResponseValuesInner, MetricsResponse } from "@/swagger-models/cluster-service-client";
import { WIDGET_AVERAGE_LINE_LABEL } from "@/models/chart.model";

export interface IEChartsSeriesOption {
  name?: string;
  type?: string;
  data?: unknown[];
  color?: string;
  barWidth?: number;
  meta?: Record<string, unknown>;
  [key: string]: unknown;
}

const MAX_DISPLAYED_PROJECTS = 20;
interface ProjectPhaseCount {
  projectName: string;
  phaseCount: number;
}

export const widgetUtil = {
  aggregateTelemetryGroupedData: aggregateTelemetryGroupedDataForECharts,
  aggregateTelemetryGroupedDataForECharts,
  enrichTelemetryResponseByKeyValue,
  sumAllTelemetryValues,
  sortByNodePool,
  filterTopAllocationRatioProjects,
  limitProjectsTelemetryResponse,
  getWorkloadTelemetryDrilldownOptions,
  combineDistributedTelemetryIntoTraining,
  gpuUtilizationPerNodeParser,
};

function aggregateTelemetryGroupedDataForECharts(
  telemetryValues: Array<TelemetryResponseValuesInner>,
  groupBy: [string, string],
  includeDrilldown = false,
  prevGroupByKey?: string,
  prevEntityId?: string,
  includeZeroValues = false,
): IEChartsSeriesOption[] {
  if (!telemetryValues) return [];
  const seriesMap = new Map();
  const filteredTelemetryValues = includeZeroValues
    ? telemetryValues
    : telemetryValues.filter((telemetryValue: TelemetryResponseValuesInner) => parseFloat(telemetryValue.value) !== 0);

  filteredTelemetryValues.forEach((telemetryValue: TelemetryResponseValuesInner) => {
    const firstGroupBy: TelemetryResponseValuesInnerGroupsInner | undefined = telemetryValue.groups?.find(
      (group: TelemetryResponseValuesInnerGroupsInner) => group.key === groupBy[0],
    );
    const secondGroupByValue: string | undefined = telemetryValue.groups?.find(
      (group: TelemetryResponseValuesInnerGroupsInner) => group.key === groupBy[1],
    )?.value;
    const count = parseFloat(telemetryValue?.value || "0");

    let seriesItem = seriesMap.get(secondGroupByValue);

    if (!seriesItem) {
      seriesItem = {
        name: secondGroupByValue,
        data: [],
      };

      // Extend seriesItem with meta information about the previous data to enable drilldown refresh.
      if (!seriesItem.meta) {
        seriesItem.meta = {};
      }
      if (prevGroupByKey) {
        seriesItem.meta.groupByKey = prevGroupByKey;
      }
      if (prevEntityId) {
        seriesItem.meta.entityId = prevEntityId;
      }

      seriesMap.set(secondGroupByValue, seriesItem);
    }

    const existingDataIndex = seriesItem.data.findIndex(
      (data: { name: string; y: number; drilldown: boolean }) => data.name === firstGroupBy?.name,
    );

    if (existingDataIndex === -1) {
      seriesItem.data.push({
        name: firstGroupBy?.name || firstGroupBy?.value,
        y: count,
        drilldown: includeDrilldown,
        ...(firstGroupBy?.value && { meta: { entityId: firstGroupBy.value, groupByKey: firstGroupBy.key } }),
      });
    } else {
      seriesItem.data[existingDataIndex].y += count;
    }
  });

  return Array.from(seriesMap.values());
}

function enrichTelemetryResponseByKeyValue(
  array: TelemetryResponseValuesInner[],
  keyToEnrich: string,
  value: string,
  key: string,
): TelemetryResponseValuesInner[] {
  return array.map((item: TelemetryResponseValuesInner) => {
    const groups = item.groups?.map((group: TelemetryResponseValuesInnerGroupsInner) => {
      if (group.key === keyToEnrich) {
        return { ...group, [key]: value };
      }
      return group;
    });

    return { ...item, groups };
  });
}

function sumAllTelemetryValues(telemetryValues: Array<TelemetryResponseValuesInner>): number {
  return telemetryValues.reduce((acc: number, telemetryValue: TelemetryResponseValuesInner) => {
    return acc + parseFloat(telemetryValue.value || "0");
  }, 0);
}

function sortByNodePool(array: TelemetryResponseValuesInner[]): TelemetryResponseValuesInner[] {
  return array.sort((a: TelemetryResponseValuesInner, b: TelemetryResponseValuesInner) => {
    const nodepoolA =
      (a.groups || []).find(
        (group: TelemetryResponseValuesInnerGroupsInner) => group.key === ENodeTelemetryGroupBy.Nodepool,
      )?.value || "";
    const nodepoolB =
      (b.groups || []).find(
        (group: TelemetryResponseValuesInnerGroupsInner) => group.key === ENodeTelemetryGroupBy.Nodepool,
      )?.value || "";
    return nodepoolA.localeCompare(nodepoolB);
  });
}

function filterTopAllocationRatioProjects(
  metricData: MetricsData[],
  maxDisplayProjects = MAX_DISPLAYED_PROJECTS,
): MetricsData[] {
  const data: MetricsData[] = deepCopy(metricData);
  const departmentHashmap: { [department: string]: { [project: string]: number } } = {};

  const getAllocationRatio = (project: IResourceData): number => {
    if (project.gpu.quota <= 0) return 0;
    return project.gpu.allocated / project.gpu.quota;
  };

  const calculateAllocationRatio = (departmentName: string, project: IResourceData): void => {
    const projectName: string = project.projectName as string;
    const allocationRatio: number = getAllocationRatio(project);

    if (!(departmentName in departmentHashmap)) {
      departmentHashmap[departmentName] = {};
    }

    if (projectName in departmentHashmap[departmentName]) {
      departmentHashmap[departmentName][projectName] += allocationRatio;
    } else {
      departmentHashmap[departmentName][projectName] = allocationRatio;
    }
  };

  const sortAndFilterProjects = (): MetricsData[] => {
    return data.map((department: MetricsData) => {
      const departmentName = department.metadata.departmentName as string;
      const projectsHashmap = departmentHashmap[departmentName] || {};

      // Sort projects within each department by allocation ratio
      const sortedProjects = Object.keys(projectsHashmap).sort(
        (a: string, b: string) => projectsHashmap[b] - projectsHashmap[a],
      );

      const topProjects = sortedProjects.slice(0, maxDisplayProjects);
      const filteredProjects = department.current.projectResources.filter((project: IResourceData) =>
        topProjects.includes(project.projectName as string),
      );

      return {
        ...department,
        current: {
          ...department.current,
          projectResources: filteredProjects,
        },
      };
    });
  };

  data.forEach((department: MetricsData) => {
    const departmentName = department.metadata.departmentName as string;
    department.current.projectResources.forEach((project: IResourceData) =>
      calculateAllocationRatio(departmentName, project),
    );
  });

  return sortAndFilterProjects();
}

function limitProjectsTelemetryResponse(
  projects: TelemetryResponseValuesInner[],
  maxDisplayedProjects = MAX_DISPLAYED_PROJECTS,
): TelemetryResponseValuesInner[] {
  const projectPhaseCounts: ProjectPhaseCount[] = _countPhasesPerProject(projects);
  const limitedProjectPhaseCounts: ProjectPhaseCount[] = projectPhaseCounts.slice(0, maxDisplayedProjects);
  return _filterProjectsByPhaseCount(projects, limitedProjectPhaseCounts);
}

const _countPhasesPerProject = (projects: TelemetryResponseValuesInner[]): ProjectPhaseCount[] => {
  const projectPhaseCountMap: { [projectName: string]: number } = {};

  const hasBuckets = projects.some((project) =>
    project?.groups?.some((group) => group.key === EWorkloadTelemetryGroupBy.Bucket),
  );

  for (const projectPhase of projects) {
    const projectIdGroup = projectPhase?.groups?.find((group) => group.key === EWorkloadTelemetryGroupBy.ProjectId);
    if (projectIdGroup && projectIdGroup.name) {
      const projectName = projectIdGroup.name;
      const value = parseFloat(projectPhase.value || "0");

      if (hasBuckets) {
        // For bucket-based telemetry, sum the values across all buckets
        projectPhaseCountMap[projectName] = (projectPhaseCountMap[projectName] || 0) + value;
      } else {
        // For phase-based telemetry, count the number of entries
        projectPhaseCountMap[projectName] = (projectPhaseCountMap[projectName] || 0) + 1;
      }
    }
  }
  const projectPhaseCounts: ProjectPhaseCount[] = Object.entries(projectPhaseCountMap)
    .map(([projectName, phaseCount]) => ({
      projectName,
      phaseCount,
    }))
    .sort((a, b) => b.phaseCount - a.phaseCount);

  return projectPhaseCounts;
};

const _filterProjectsByPhaseCount = (
  projects: TelemetryResponseValuesInner[],
  projectPhaseCounts: ProjectPhaseCount[],
): TelemetryResponseValuesInner[] => {
  return projects.filter((project: TelemetryResponseValuesInner) => {
    const projectIdGroup: TelemetryResponseValuesInnerGroupsInner | undefined = project?.groups?.find(
      (group) => group.key === EWorkloadTelemetryGroupBy.ProjectId,
    );

    if (projectIdGroup) {
      return projectPhaseCounts.some((projectPhaseCount) => projectPhaseCount.projectName === projectIdGroup.name);
    }

    return false;
  });
};

function getWorkloadTelemetryDrilldownOptions(
  meta: EWorkloadTelemetryPointMeta,
  hasDepartments: boolean,
): {
  includeDrilldown: boolean;
  departmentIdFilter: string | undefined;
  groupByKey: EWorkloadTelemetryGroupByWithoutBucket;
} {
  const { entityId, groupByKey } = meta;
  let nextGroupBy: EWorkloadTelemetryGroupBy;
  if (groupByKey === EWorkloadTelemetryGroupBy.ClusterId && hasDepartments) {
    nextGroupBy = EWorkloadTelemetryGroupBy.DepartmentId;
  } else {
    nextGroupBy = EWorkloadTelemetryGroupBy.ProjectId;
  }
  const includeDrilldown: boolean = nextGroupBy !== EWorkloadTelemetryGroupBy.ProjectId;
  const departmentIdFilter: string | undefined =
    nextGroupBy === EWorkloadTelemetryGroupBy.ProjectId && hasDepartments ? entityId : undefined;
  return { includeDrilldown, departmentIdFilter, groupByKey: nextGroupBy };
}

/**
 * Combines all "Distributed" telemetry values into the "Training" type, same as we display in workload index.
 *
 * @param {TelemetryResponseValuesInner[]} data - The array of telemetry data to process.
 * @returns {TelemetryResponseValuesInner[]} - The processed telemetry data with combined values.
 */
function combineDistributedTelemetryIntoTraining(data: TelemetryResponseValuesInner[]): TelemetryResponseValuesInner[] {
  const { distributedSum, trainingIndex } = _calculateDistributedSumAndTrainingIndex(data);

  if (trainingIndex !== -1) {
    return _updateTrainingWithDistributedSum(data, distributedSum, trainingIndex);
  } else {
    return _updateDistributedToTraining(data, distributedSum);
  }
}

function _calculateDistributedSumAndTrainingIndex(data: TelemetryResponseValuesInner[]): {
  distributedSum: number;
  trainingIndex: number;
} {
  let distributedSum = 0;
  let trainingIndex = -1;

  data.forEach((item, index) => {
    const typeGroup = item.groups?.find((group) => group.key === EWorkloadTelemetryGroupBy.Type);
    if (typeGroup?.value === EWorkloadType.Distributed) {
      distributedSum += parseFloat(item.value);
    } else if (typeGroup?.value === EWorkloadType.Training) {
      trainingIndex = index;
    }
  });

  return { distributedSum, trainingIndex };
}

function _updateTrainingWithDistributedSum(
  data: TelemetryResponseValuesInner[],
  distributedSum: number,
  trainingIndex: number,
): TelemetryResponseValuesInner[] {
  data[trainingIndex].value = (parseFloat(data[trainingIndex].value) + distributedSum).toString();
  return data.filter((item) => {
    const typeGroup = item.groups?.find((group) => group.key === "Type");
    return typeGroup?.value !== EWorkloadType.Distributed;
  });
}

function _updateDistributedToTraining(
  data: TelemetryResponseValuesInner[],
  distributedSum: number,
): TelemetryResponseValuesInner[] {
  return data.map((item) => {
    const typeGroup = item.groups?.find((group) => group.key === "Type");
    if (typeGroup?.value === EWorkloadType.Distributed) {
      typeGroup.value = EWorkloadType.Training;
      item.value = distributedSum.toString();
    }
    return item;
  });
}

function gpuUtilizationPerNodeParser(
  metricsResponse: MetricsResponse,
  averageMetricType?: string,
): Record<string, Array<GroupedMeasurementResponseValuesInner>> {
  const measurementMap: Record<string, Array<GroupedMeasurementResponseValuesInner>> = {};
  metricsResponse.measurements
    .sort((a, b) => {
      if (a.type === averageMetricType) return -1;
      if (b.type === averageMetricType) return 1;
      const labelA = parseInt(a.labels?.gpu || "0", 10);
      const labelB = parseInt(b.labels?.gpu || "0", 10);
      return labelA - labelB;
    })
    .forEach((measurement) => {
      if (measurement.type === averageMetricType) {
        measurementMap[WIDGET_AVERAGE_LINE_LABEL] = [
          ...(measurementMap[WIDGET_AVERAGE_LINE_LABEL] || []),
          ...(measurement.values || []),
        ];
      } else {
        const label = measurement.labels?.gpu !== undefined ? `-${measurement.labels?.gpu}` : "";
        const key = `GPU${label}`;
        measurementMap[key] = [...(measurementMap[key] || []), ...(measurement.values || [])];
      }
    });

  return measurementMap;
}
