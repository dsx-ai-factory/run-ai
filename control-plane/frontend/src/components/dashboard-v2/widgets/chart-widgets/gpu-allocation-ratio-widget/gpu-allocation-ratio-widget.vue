<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <chart-widget-wrapper
    :options="wrapperOptions"
    :loading="displayLoading"
    :error="error"
    @link-clicked="redirectToEntityIndex"
    hide-actions
    :empty="isChartEmpty"
    :announcement="drilldownAnnouncement"
  >
    <div ref="chartAreaRef" class="gpu-allocation-ratio-chart-container">
      <div v-show="breadcrumbPath.length > 0" class="breadcrumb-nav" role="navigation" aria-label="Chart drilldown">
        <span v-for="(crumb, idx) in breadcrumbPath" :key="idx" class="breadcrumb-item">
          <button
            v-if="idx < breadcrumbPath.length - 1 || breadcrumbPath.length === 1"
            type="button"
            class="breadcrumb-link"
            @click="drillUp(idx)"
          >
            {{ crumb.label }}
          </button>
          <span v-else class="breadcrumb-current" aria-current="page">{{ crumb.label }}</span>
          <span v-if="idx < breadcrumbPath.length - 1" class="breadcrumb-separator"> / </span>
        </span>
      </div>
      <v-chart
        aid="gpu-allocation-ratio-chart"
        v-if="chartOption"
        ref="chartRef"
        :option="chartOption"
        autoresize
        class="gpu-allocation-ratio-chart"
        @click="handleChartClick"
      />
      <div
        v-if="drilldownAxisCategories.length"
        class="gpu-allocation-ratio-axis-labels"
        role="group"
        aria-label="Explore chart data"
        :style="{ bottom: `${layout.externalCategoryLabelBottom}px` }"
      >
        <button
          v-for="category in drilldownAxisCategories"
          :key="category"
          class="gpu-allocation-ratio-axis-label"
          type="button"
          :aria-label="`View details for ${category}`"
          @click="handleAxisCategoryClick(category)"
        >
          {{ category }}
        </button>
      </div>
    </div>
  </chart-widget-wrapper>
</template>

<script setup lang="ts">
import "@/plugins/echarts";
import { ref, watch, onMounted } from "vue";
import VChart from "vue-echarts";
import type { BarSeriesOption, EChartsOption, SeriesOption } from "echarts";
import { useRouter } from "vue-router";

import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import {
  type ChartDataMapItem,
  type DepartmentsMetricData,
  EDrillDownType,
  type IChartArrayItemConfig,
  type IChartMapItemConfig,
  type IResourceData,
  type MetricsData,
} from "@/models/metrics.model";
import {
  ACTIVE_AXIS_LABEL_COLOR,
  CHART_DATA_LABEL_COLOR,
  EWidgetEntity,
  type IWidgetWrapperOptions,
  chartColorPaletteArray,
  DRILLDOWN_CHART_ANIMATION_PRESET,
  getDrilldownTransitionPreset,
  DRILLDOWN_UNIVERSAL_TRANSITION,
} from "@/models/chart.model";
import { departmentService } from "@/services/control-plane/department.service/department.service";
import { HttpErrorResponse } from "@/models/http-response.model";
import type { IFilterBy } from "@/models/filter.model";
import commonUtil from "@/utils/common.util/common.util";
import { widgetUtil } from "@/utils/widget.util";
import { chartUtil, STANDARD_TOOLTIP_STYLE } from "@/utils/chart.util";
import { useAdaptiveCategoricalChartLayout } from "@/composables/use-adaptive-categorical-chart-layout.composable";
import { filterService } from "@/services/filter.service/filter.service";
import { ETableFilters } from "@/models/table.model";
import { allProjectColumns } from "@/table-models/project.table-model";
import { departmentIndexColumns } from "@/table-models/department.table-model";
import { EDepartmentColumnName } from "@/models/department.model";
import { EProjectColumnName } from "@/models/project.model";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";

const UNKNOWN_RESOURCE = "Unknown";

interface AllocationRatioMeta {
  quota: number;
  allocated: number;
}

interface BreadcrumbItem {
  label: string;
  drilldownIds: string[];
}

interface DrilldownDataPoint {
  name: string;
  value: number;
  drilldown?: string | boolean;
  meta: AllocationRatioMeta;
}

const chartGridOptions = {
  top: 11,
  bottom: 74,
  left: 56,
  right: 12,
};

const drilldownChartGridOptions = {
  top: 52,
  bottom: 76,
  left: 56,
  right: 12,
};

const yAxisTitleOptions = {
  name: "Allocation ratio",
  nameLocation: "middle" as const,
  nameRotate: 90,
  nameGap: 36,
  nameTextStyle: {
    color: "#585f69",
    fontSize: 12,
    fontWeight: 400,
  },
};

const legendOptions = {
  show: true,
  bottom: 16,
  icon: "circle",
  itemWidth: 12,
  itemHeight: 12,
  itemGap: 20,
  textStyle: {
    fontWeight: "bold" as const,
    fontSize: 12,
    color: "#333",
    padding: [0, 0, 0, 5],
  },
};

export interface Props {
  clusterId: string;
  nodePoolName?: string;
  hasDepartments: boolean;
}

const props = defineProps<Props>();
const router = useRouter();

const nodePoolsNames = ref<string[]>([]);
const allocationClusterByNodePoolArray = ref<IChartArrayItemConfig[]>([]);
const allocationDepartmentsByNodePoolArray = ref<IChartArrayItemConfig[]>([]);
const displayLoading = ref(true);
const error = ref(false);
const isChartEmpty = ref(false);
const chartOption = ref<EChartsOption | null>(null);

const breadcrumbPath = ref<BreadcrumbItem[]>([]);
const currentDrilldownId = ref<string | null>(null);
const drilldownAxisCategories = ref<string[]>([]);
const drilldownAnnouncement = ref("");
const axisDrilldownMap = new Map<string, string[]>();
const chartRef = ref<InstanceType<typeof VChart> | null>(null);
const chartAreaRef = ref<HTMLElement | null>(null);
const { layout, getCategoryAxisLabelOptions } = useAdaptiveCategoricalChartLayout({
  chartRef,
  containerRef: chartAreaRef,
  horizontalGridPadding: chartGridOptions.left + chartGridOptions.right,
});

const wrapperOptions = ref<IWidgetWrapperOptions>({
  title: "Allocation ratio by node pool",
  linkText: "",
  entityType: EWidgetEntity.Department,
  tooltipText: "This ratio (allocation/quota) shows how well <br/>the department or project is using its quota",
});

function isAllocationRatioMeta(meta: unknown): meta is AllocationRatioMeta {
  return typeof meta === "object" && meta !== null && "quota" in meta && "allocated" in meta;
}

function initWrapperOptions(): void {
  wrapperOptions.value.linkText = props.hasDepartments ? "All departments" : "All projects";
}

function buildChartOption(
  categories: string[],
  seriesData: Array<{ name: string; data: Array<DrilldownDataPoint | null> }>,
  parentGroupId?: string,
  animateDrilldown = false,
): EChartsOption {
  axisDrilldownMap.clear();
  categories.forEach((category) => {
    const drilldownIds = Array.from(
      new Set(
        seriesData
          .flatMap((s) => s.data)
          .filter((d): d is DrilldownDataPoint => d !== null && d.name === category && typeof d.drilldown === "string")
          .map((d) => d.drilldown as string),
      ),
    );
    if (drilldownIds.length) {
      axisDrilldownMap.set(category, drilldownIds);
    }
  });
  drilldownAxisCategories.value = Array.from(axisDrilldownMap.keys());

  const series = seriesData.map((s) => ({
    type: "bar" as const,
    id: `gpu-allocation-ratio-${s.name}`,
    name: s.name,
    data: s.data.map((d) =>
      d
        ? {
            value: d.value,
            name: d.name,
            groupId: parentGroupId ?? d.name,
            childGroupId: d.name,
            drilldown: d.drilldown,
            meta: d.meta,
          }
        : null,
    ),
    barMaxWidth: 30,
    barGap: "30%",
    barCategoryGap: "40%",
    universalTransition: DRILLDOWN_UNIVERSAL_TRANSITION,
    ...chartUtil.getBarSeriesDefaults(),
    label: {
      show: true,
      position: "top" as const,
      color: CHART_DATA_LABEL_COLOR,
      fontSize: 12,
      fontWeight: 400 as const,
      formatter: (params: { data: { value: number; meta: AllocationRatioMeta } }) => {
        if (isAllocationRatioMeta(params.data.meta) && params.data.meta.quota < 0) return "-";
        return `${params.data.value}%`;
      },
    },
  })) as unknown as BarSeriesOption[];
  const yAxisConfig = getAllocationRatioYAxisConfig(series);

  return {
    ...DRILLDOWN_CHART_ANIMATION_PRESET,
    ...(animateDrilldown ? getDrilldownTransitionPreset() : {}),
    color: chartColorPaletteArray,
    grid: breadcrumbPath.value.length ? drilldownChartGridOptions : chartGridOptions,
    xAxis: {
      type: "category",
      data: categories,
      axisLabel: drilldownAxisCategories.value.length
        ? {
            ...getCategoryAxisLabelOptions(categories.length),
            interval: 0,
            rotate: 0,
            show: false,
            margin: 16,
            color: ACTIVE_AXIS_LABEL_COLOR,
            fontSize: 11,
          }
        : {
            ...getCategoryAxisLabelOptions(categories.length),
            show: true,
            interval: 0,
            rotate: 0,
            margin: 10,
            color: "#666",
            fontSize: 11,
            overflow: "truncate",
            ellipsis: "...",
            lineHeight: 16,
            hideOverlap: false,
            formatter: chartUtil.formatCategoricalAxisLabel,
          },
      axisLine: { show: true, lineStyle: { color: "#ccd6eb", width: 1 } },
      axisTick: { show: false },
      splitLine: { show: false },
      triggerEvent: drilldownAxisCategories.value.length > 0,
    },
    yAxis: {
      type: "value",
      min: 0,
      ...yAxisConfig,
      axisLabel: { formatter: "{value}%", margin: 12, color: "#666", fontSize: 11 },
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: {
        show: true,
        lineStyle: { color: "#e6e6e6", width: 1, type: "solid" },
      },
      ...yAxisTitleOptions,
    },
    tooltip: {
      trigger: "item",
      ...STANDARD_TOOLTIP_STYLE,
      formatter: (params: {
        data: { name: string; value: number; meta: AllocationRatioMeta };
        color: string;
        seriesName: string;
      }) => {
        const meta = params.data.meta;
        if (!isAllocationRatioMeta(meta)) return "";
        const nodePoolName = chartUtil.escapeTooltipHtml(params.seriesName);
        const isUnlimited = meta.quota < 0;
        const allocationRatio = chartUtil.escapeTooltipHtml(
          isUnlimited ? "-" : commonUtil.formatNumberWithFixedDecimals(params.data.value, 2) + "%",
        );
        const gpuAllocation = chartUtil.escapeTooltipHtml(commonUtil.formatNumberWithFixedDecimals(meta.allocated, 2));
        const gpuQuota = chartUtil.escapeTooltipHtml(
          isUnlimited ? "Unlimited" : commonUtil.formatNumberWithFixedDecimals(meta.quota, 2),
        );
        return `${chartUtil.renderTooltipDot(params.color)}Node pool: <b>${nodePoolName}</b><br/>
          &nbsp;&nbsp;&nbsp;GPU allocation ratio: <b>${allocationRatio}</b><br/>
          &nbsp;&nbsp;&nbsp;GPU allocation: <b>${gpuAllocation}</b><br/>
          &nbsp;&nbsp;&nbsp;GPU quota: <b>${gpuQuota}</b>`;
      },
    } as unknown as EChartsOption["tooltip"],
    legend: { ...legendOptions, data: seriesData.map((s) => s.name) },
    series,
  };
}

function getAllocationRatioYAxisConfig(series: SeriesOption[]): { interval: number; max: number } {
  const config = chartUtil.getBarChartYAxisConfig(series);
  const dataMax = Math.max(
    0,
    ...series.flatMap((seriesItem) =>
      ((seriesItem as { data?: Array<{ value?: number } | null> }).data || []).map((point) => point?.value || 0),
    ),
  );

  // ECharts renders top labels outside the bar. Reserve a full nice-number interval
  // only when a value is within the top 10% of the calculated scale.
  return dataMax > config.max * 0.9 ? { ...config, max: config.max + config.interval } : config;
}

function setChartData(animateDrilldown = false): void {
  const chartData = allocationClusterByNodePoolArray.value;
  if (chartData.length === 0) {
    isChartEmpty.value = true;
    return;
  }
  isChartEmpty.value = false;

  const sortedData = sortByNodepoolName(chartData);
  const categories = sortedData.length > 0 ? sortedData[0].data.map((d) => d.name) : [];
  const seriesData = sortedData.map((np) => ({
    name: np.name,
    data: np.data.map((d) => ({
      name: d.name,
      value: d.y,
      drilldown: d.drilldown,
      meta: d.meta as AllocationRatioMeta,
    })),
  }));

  chartOption.value = buildChartOption(categories, seriesData, undefined, animateDrilldown);
  breadcrumbPath.value = [];
  currentDrilldownId.value = null;
}

function handleChartClick(event: unknown): void {
  const params = event as { data?: { drilldown?: string; name?: string } };
  const category = params.data?.name;
  if (category) {
    handleAxisCategoryClick(category);
  }
}

function handleAxisCategoryClick(category: string): void {
  const drilldownIds = axisDrilldownMap.get(category);
  if (!drilldownIds?.length) return;

  const drilldownSeries = drilldownIds
    .map((drilldownId) => allocationDepartmentsByNodePoolArray.value.find((series) => series.id === drilldownId))
    .filter((series): series is IChartArrayItemConfig => Boolean(series));
  if (!drilldownSeries.length) return;

  const currentPath = breadcrumbPath.value;
  const drilldownLabel = props.hasDepartments ? "Departments" : "Projects";
  breadcrumbPath.value = currentPath.length
    ? [...currentPath, { label: category, drilldownIds }]
    : [
        { label: category, drilldownIds },
        { label: drilldownLabel, drilldownIds },
      ];
  currentDrilldownId.value = drilldownIds[0];

  showDrilldownData(drilldownSeries, category);
  drilldownAnnouncement.value = `${drilldownLabel} details loaded.`;
}

function showDrilldownData(drilldownSeries: IChartArrayItemConfig[], parentGroupId?: string): void {
  const categories = Array.from(new Set(drilldownSeries.flatMap((series) => series.data.map((d) => d.name))));
  const seriesData = drilldownSeries.map((series) => {
    const pointsByCategory = new Map(series.data.map((point) => [point.name, point]));
    return {
      name: series.name,
      data: categories.map((category) => {
        const point = pointsByCategory.get(category);
        return point
          ? {
              name: point.name,
              value: point.y,
              drilldown: point.drilldown,
              meta: point.meta as AllocationRatioMeta,
            }
          : null;
      }),
    };
  });
  chartOption.value = buildChartOption(categories, seriesData, parentGroupId, true);
}

function drillUp(targetIdx: number): void {
  if (targetIdx < 0) {
    setChartData(true);
    drilldownAnnouncement.value = "Top chart level loaded.";
    return;
  }

  breadcrumbPath.value = breadcrumbPath.value.slice(0, targetIdx + 1);
  const targetCrumb = breadcrumbPath.value[targetIdx];

  if (targetIdx === 0 && breadcrumbPath.value.length === 1) {
    breadcrumbPath.value = [];
    setChartData(true);
    drilldownAnnouncement.value = "Top chart level loaded.";
    return;
  }

  const drilldownSeries = targetCrumb.drilldownIds
    .map((drilldownId) => allocationDepartmentsByNodePoolArray.value.find((series) => series.id === drilldownId))
    .filter((series): series is IChartArrayItemConfig => Boolean(series));
  if (drilldownSeries.length) {
    currentDrilldownId.value = targetCrumb.drilldownIds[0];
    const parentGroupId = breadcrumbPath.value[targetIdx - 1]?.label ?? targetCrumb.label;
    showDrilldownData(drilldownSeries, parentGroupId);
    drilldownAnnouncement.value = "Previous drilldown level loaded.";
  }
}

async function loadInitialChartData(): Promise<void> {
  try {
    error.value = false;
    const res: DepartmentsMetricData = await departmentService.getDepartmentsMetric(props.clusterId, props.nodePoolName);
    const topAllocationRatioProjects = widgetUtil.filterTopAllocationRatioProjects(res.data);
    nodePoolsNames.value = extractNodepoolNames(topAllocationRatioProjects);
    transformDataIntoChartFormat(topAllocationRatioProjects);
  } catch (e: unknown) {
    handleError(e);
  } finally {
    displayLoading.value = false;
  }
}

function buildResourcesByNodePoolMap(
  resourceMap: Record<string, IChartMapItemConfig>,
  resourceKey: string,
  resourceName: string,
  drilldownTo: string,
  projectResource: IResourceData,
  drilldownId: string,
): void {
  if (!resourceMap[resourceKey]) {
    resourceMap[resourceKey] = { id: drilldownId, name: resourceName, data: {} };
  }
  if (!resourceMap[resourceKey].data[resourceName]) {
    resourceMap[resourceKey].data[resourceName] = {
      name: resourceName,
      quotaY: projectResource.gpu.quota,
      allocatedY: projectResource.gpu.allocated,
      drilldown: drilldownTo ? `${drilldownTo}-${resourceName}` : "",
    };
  } else {
    const existing = resourceMap[resourceKey].data[resourceName];
    if (projectResource.gpu.quota < 0 || existing.quotaY < 0) {
      existing.quotaY = -1;
    } else {
      existing.quotaY += projectResource.gpu.quota;
    }
    existing.allocatedY += projectResource.gpu.allocated;
  }
}

function sortByNodepoolName(array: IChartArrayItemConfig[]): IChartArrayItemConfig[] {
  return [...array].sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()));
}

function transformDataIntoChartFormat(data: Array<MetricsData>): void {
  const clusterByNodePoolMap: Record<string, IChartMapItemConfig> = {};
  const departmentsByNodePoolMap: Record<string, IChartMapItemConfig> = {};
  const projectsByNodePoolAndDepartmentsMap: Record<string, IChartMapItemConfig> = {};

  for (const departmentData of data) {
    const clusterName = departmentData.metadata.clusterName;
    const resources: IResourceData[] = departmentData.current.projectResources || departmentData.current.resources;

    if (props.hasDepartments) {
      for (const departmentResource of departmentData.current.resources) {
        const nodepoolName = departmentResource.nodepoolName;
        const departmentName = departmentData.metadata.departmentName || UNKNOWN_RESOURCE;

        buildResourcesByNodePoolMap(
          clusterByNodePoolMap,
          nodepoolName,
          clusterName,
          getDrilldownId(EDrillDownType.Clusters, clusterName, nodepoolName, departmentName),
          departmentResource,
          "",
        );
        buildResourcesByNodePoolMap(
          departmentsByNodePoolMap,
          nodepoolName,
          departmentName,
          `projects-${nodepoolName}`,
          departmentResource,
          getDrilldownId(EDrillDownType.Departments, clusterName, nodepoolName, departmentName),
        );
      }
    }

    for (const projectResource of resources) {
      const departmentName =
        projectResource.departmentName || departmentData.metadata.departmentName || UNKNOWN_RESOURCE;
      const projectName = projectResource.projectName || departmentData.metadata.projectName || UNKNOWN_RESOURCE;
      const nodepoolName = projectResource.nodepoolName;

      if (!props.hasDepartments) {
        buildResourcesByNodePoolMap(
          clusterByNodePoolMap,
          nodepoolName,
          clusterName,
          getDrilldownId(EDrillDownType.Clusters, clusterName, nodepoolName, departmentName),
          projectResource,
          "",
        );
        buildResourcesByNodePoolMap(
          departmentsByNodePoolMap,
          nodepoolName,
          projectName,
          "",
          projectResource,
          getDrilldownId(EDrillDownType.Departments, clusterName, nodepoolName, departmentName),
        );
      }
      buildResourcesByNodePoolMap(
        projectsByNodePoolAndDepartmentsMap,
        props.hasDepartments ? `${nodepoolName}-${departmentName}` : projectName,
        projectName,
        "",
        projectResource,
        getDrilldownId(EDrillDownType.Projects, clusterName, nodepoolName, departmentName),
      );
    }
  }

  allocationClusterByNodePoolArray.value = [];
  allocationDepartmentsByNodePoolArray.value = [];

  convertResourceMapToChartArrayData(clusterByNodePoolMap, allocationClusterByNodePoolArray.value);
  convertResourceMapToChartArrayData(departmentsByNodePoolMap, allocationDepartmentsByNodePoolArray.value);
  if (props.hasDepartments) {
    convertResourceMapToChartArrayData(projectsByNodePoolAndDepartmentsMap, allocationDepartmentsByNodePoolArray.value);
  }
  fixNodePoolsNames();
}

function getDrilldownId(drillDownType: EDrillDownType, cluster: string, nodepool: string, department: string): string {
  switch (drillDownType) {
    case EDrillDownType.Clusters:
      return props.hasDepartments ? `departments-${nodepool}` : `projects-${nodepool}`;
    case EDrillDownType.Departments:
      return props.hasDepartments ? `departments-${nodepool}-${cluster}` : `projects-${nodepool}-${cluster}`;
    case EDrillDownType.Projects:
      return props.hasDepartments ? `projects-${nodepool}-${department}` : `projects-${nodepool}`;
  }
}

function convertResourceMapToChartArrayData(
  resourceMap: Record<string, IChartMapItemConfig>,
  chartArray: IChartArrayItemConfig[],
): void {
  const calculateAllocationRatio = (quota: number, allocated: number): number => {
    if (quota <= 0) return 0;
    return parseFloat(((allocated / quota) * 100).toFixed(2));
  };
  const getCorrectNodePoolName = (mixedName: string) => {
    return nodePoolsNames.value.find((name) => mixedName.startsWith(name) && mixedName.length !== name.length);
  };

  for (const nodepoolName in resourceMap) {
    const nodePoolData = resourceMap[nodepoolName];
    const data: Array<ChartDataMapItem> = Object.values(nodePoolData.data);
    const item: IChartArrayItemConfig = {
      id: nodePoolData.id,
      name: nodepoolName,
      meta: { correctName: getCorrectNodePoolName(nodepoolName) },
      data: data.map((entry) => ({
        name: entry.name,
        y: calculateAllocationRatio(entry.quotaY, entry.allocatedY),
        ...(entry.drilldown ? { drilldown: entry.drilldown } : {}),
        meta: { quota: entry.quotaY, allocated: entry.allocatedY },
      })),
    };
    chartArray.push(item);
  }
}

function extractNodepoolNames(data: MetricsData[]): string[] {
  const names = new Set<string>();
  data.forEach((item) => {
    item.current.resources.forEach((r) => names.add(r.nodepoolName));
    item.current.projectResources.forEach((r) => names.add(r.nodepoolName));
  });
  return Array.from(names);
}

function fixNodePoolsNames(): void {
  allocationDepartmentsByNodePoolArray.value.forEach((series) => {
    if (series.meta?.correctName) {
      series.name = series.meta.correctName;
    }
  });
}

function handleError(err: unknown): void {
  error.value = true;
  if (err instanceof HttpErrorResponse) {
    console.error(err.message);
  } else {
    console.error("Error fetching data", err);
  }
}

function resetData(): void {
  chartOption.value = null;
  nodePoolsNames.value = [];
  allocationClusterByNodePoolArray.value = [];
  allocationDepartmentsByNodePoolArray.value = [];
  breadcrumbPath.value = [];
  currentDrilldownId.value = null;
  drilldownAxisCategories.value = [];
  axisDrilldownMap.clear();
}

function redirectToEntityIndex(): void {
  setColumnFilter();
  const routeName = props.hasDepartments ? DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX : PROJECT_ROUTE_NAMES.PROJECT_INDEX;
  router.push({ name: routeName });
}

function setColumnFilter(): void {
  if (props.nodePoolName) {
    setNodePoolColumnFilter(getNodesFilterBy());
  }
}

function getNodesFilterBy(): IFilterBy {
  const columns = props.hasDepartments ? departmentIndexColumns : allProjectColumns;
  const sortBy = props.hasDepartments ? EDepartmentColumnName.DepartmentName : EProjectColumnName.ProjectName;
  const tableFilterName = props.hasDepartments ? ETableFilters.DEPARTMENT : ETableFilters.PROJECT;
  const defaultFilters = filterService.getDefaultFilters(sortBy, columns);
  return filterService.loadFilters(window.location, tableFilterName, defaultFilters);
}

function setNodePoolColumnFilter(nodesFilterBy: IFilterBy): void {
  const filterName = props.hasDepartments ? EDepartmentColumnName.NodePools : EProjectColumnName.NodePools;
  const tableFilterName = props.hasDepartments ? ETableFilters.DEPARTMENT : ETableFilters.PROJECT;
  filterService.setColumnFilter(nodesFilterBy, props.nodePoolName as string, filterName, tableFilterName);
}

async function getInitialData(): Promise<void> {
  await loadInitialChartData();
  setChartData();
}

onMounted(async () => {
  initWrapperOptions();
  await getInitialData();
  displayLoading.value = false;
});

watch(
  () => [props.nodePoolName, props.clusterId],
  async () => {
    displayLoading.value = true;
    resetData();
    await getInitialData();
  },
);
</script>

<style lang="scss" scoped>
.gpu-allocation-ratio-chart-container {
  position: relative;
  height: 240px;
}

.gpu-allocation-ratio-chart {
  height: 240px;
}

.gpu-allocation-ratio-axis-labels {
  position: absolute;
  right: 20px;
  bottom: 51px;
  left: 72px;
  display: flex;
  pointer-events: none;
}

.gpu-allocation-ratio-axis-label {
  min-width: 0;
  flex: 1 1 0;
  padding: 0;
  overflow: hidden;
  color: var(--text-link-default);
  font-size: 12px;
  line-height: 16px;
  font-weight: 600;
  text-align: center;
  text-decoration: underline;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
  background: transparent;
  border: 0;
  pointer-events: auto;
}

.breadcrumb-nav {
  position: absolute;
  top: 14px;
  left: 63px;
  z-index: 1;
  padding: 0;
  font-size: 12px;
  color: var(--b-text-sub);

  .breadcrumb-item {
    .breadcrumb-link {
      border: 0;
      background: transparent;
      padding: 0;
      font: inherit;
      color: var(--text-link-default);
      cursor: pointer;

      &:hover {
        text-decoration: underline;
      }
    }

    .breadcrumb-current {
      color: var(--b-text-main);
      font-weight: 700;
    }
  }

  .breadcrumb-separator {
    margin: 0 4px;
    color: var(--text-disabled);
  }
}
</style>
