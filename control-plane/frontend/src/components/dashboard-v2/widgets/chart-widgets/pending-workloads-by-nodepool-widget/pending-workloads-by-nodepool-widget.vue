<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <chart-widget-wrapper
    class="pending-workloads-by-nodepool-widget"
    :options="wrapperOptions"
    :loading="displayLoading"
    :error="error"
    :empty="isChartEmpty"
    :announcement="drilldownAnnouncement"
    hide-actions
  >
    <div v-if="chartOption" ref="chartAreaRef" class="pending-workloads-by-nodepool-chart-area">
      <chart-drilldown-breadcrumbs :items="drillBreadcrumbs" @navigate="drillUpTo" />
      <v-chart
        aid="pending-workloads-by-nodepool-chart"
        ref="chartRef"
        :option="chartOption"
        autoresize
        class="pending-workloads-by-nodepool-chart"
        key="pending-workloads-by-nodepool-chart"
        @click="handleChartClick"
      />
      <div
        v-if="drilldownAxisCategories.length"
        class="pending-workloads-by-nodepool-axis-labels"
        role="group"
        aria-label="Explore chart data"
        :style="{ bottom: `${layout.externalCategoryLabelBottom}px` }"
      >
        <button
          v-for="category in drilldownAxisCategories"
          :key="category"
          class="pending-workloads-by-nodepool-axis-label"
          type="button"
          :aria-label="`View details for ${category}`"
          @click="drillDownByCategory(category)"
        >
          {{ category }}
        </button>
      </div>
    </div>
  </chart-widget-wrapper>
</template>

<script setup lang="ts">
import "@/plugins/echarts";
import { ref, computed, watch, onMounted } from "vue";
import VChart from "vue-echarts";
import type { EChartsOption, BarSeriesOption } from "echarts";

// store
import { useClusterStore } from "@/stores/cluster.store";

// models
import {
  EWidgetEmptyMessages,
  type IWidgetWrapperOptions,
  chartColorPaletteArray,
  UNKNOWN_COLOR,
  CHART_DATA_LABEL_COLOR,
  ACTIVE_AXIS_LABEL_COLOR,
  DRILLDOWN_CHART_ANIMATION_PRESET,
  getDrilldownTransitionPreset,
  DRILLDOWN_UNIVERSAL_TRANSITION,
} from "@/models/chart.model";
import {
  EWorkloadTelemetryGroupBy,
  type EWorkloadTelemetryGroupByWithoutBucket,
  type EWorkloadTelemetryPointMeta,
} from "@/models/workload.model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

// components
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { ChartDrilldownBreadcrumbs } from "@/components/common/charts/chart-drilldown-breadcrumbs";

// services
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { Phase, type TelemetryResponseValuesInner, WorkloadTelemetryType } from "@/swagger-models/workloads-client";

// utils
import { widgetUtil } from "@/utils/widget.util";
import {
  chartUtil,
  completeChartDrilldown,
  getChartDrillUpNavigation,
  getNextChartDrilldownBreadcrumbs,
  STANDARD_TOOLTIP_STYLE,
} from "@/utils/chart.util";
import { useAdaptiveCategoricalChartLayout } from "@/composables/use-adaptive-categorical-chart-layout.composable";

const chartGridOptions = {
  top: 11,
  bottom: 74,
  left: 56,
  right: 12,
};

const drilldownChartGridOptions = {
  top: 52,
  bottom: 86,
  left: 56,
  right: 12,
};

const yAxisTitleOptions = {
  name: "Workloads",
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
  orgUnitMetricSource?: IOrgUnitMetricSource | null;
}

const props = withDefaults(defineProps<Props>(), {
  orgUnitMetricSource: null,
});

const clusterStore = useClusterStore();

const isDataLoading = ref(false);
const displayLoading = ref(true);
const chartOption = ref<EChartsOption | null>(null);
const error = ref(false);
const isChartEmpty = ref(false);
const latestRequestId = ref(0);

const colorMap = new Map<string, string>();
let nextColorIndex = 0;

interface DrillState {
  option: EChartsOption;
  meta: EWorkloadTelemetryPointMeta | null;
  categoryMetaMap: Map<string, EWorkloadTelemetryPointMeta>;
  drilldownAxisCategories: string[];
}
const drillStack = ref<DrillState[]>([]);
const drillBreadcrumbs = ref<string[]>([]);
const currentDrillMeta = ref<EWorkloadTelemetryPointMeta | null>(null);
const categoryMetaMap = ref<Map<string, EWorkloadTelemetryPointMeta>>(new Map());
const drilldownAxisCategories = ref<string[]>([]);
const drilldownAnnouncement = ref("");
const chartRef = ref<InstanceType<typeof VChart> | null>(null);
const chartAreaRef = ref<HTMLElement | null>(null);
const { layout, getCategoryAxisLabelOptions, getGroupedBarLayout } = useAdaptiveCategoricalChartLayout({
  chartRef,
  containerRef: chartAreaRef,
  horizontalGridPadding: chartGridOptions.left + chartGridOptions.right,
});

function getColor(type: string): string {
  if (!type) return UNKNOWN_COLOR;
  if (colorMap.has(type)) return colorMap.get(type) as string;
  const color = chartColorPaletteArray[nextColorIndex % chartColorPaletteArray.length];
  colorMap.set(type, color);
  nextColorIndex++;
  return color;
}

interface HcDataPoint {
  name: string;
  y: number;
  drilldown?: boolean;
  meta?: { entityId: string; groupByKey: string };
}

interface HcSeriesItem {
  name?: string;
  data: HcDataPoint[];
  color?: string;
}

function formatTooltip(params: unknown): string {
  const items = Array.isArray(params) ? params : [params];
  const firstItem = items[0] as { axisValueLabel?: string; name?: string } | undefined;
  const header = chartUtil.escapeTooltipHtml(firstItem?.axisValueLabel || firstItem?.name || "");
  const headerHtml = `<div style="padding-bottom: 4px;">${header}</div>`;

  const tooltip = items.reduce((html, item) => {
    const point = item as { color?: string; seriesName?: string; value?: unknown };
    const rawValue =
      typeof point.value === "object" && point.value !== null && "value" in point.value
        ? (point.value as { value?: unknown }).value
        : point.value;
    const value = rawValue === null || rawValue === undefined || rawValue === "" ? null : Number(rawValue);
    if (value === null || !Number.isFinite(value)) return html;

    return html + chartUtil.formatTooltipValue(point.seriesName || "", point.color, value, "");
  }, headerHtml);

  return tooltip === headerHtml ? "" : tooltip;
}

function convertAndColorSeries(
  hcSeries: HcSeriesItem[],
  _isDrilldown = false,
  parentGroupId?: string,
): {
  categories: string[];
  series: BarSeriesOption[];
} {
  const sortedSeries = [...hcSeries].sort((a, b) => (a.name || "").localeCompare(b.name || ""));
  sortedSeries.forEach((s) => {
    if (s.name) s.color = getColor(s.name);
  });

  const categorySet = new Set<string>();
  const categories: string[] = [];
  categoryMetaMap.value = new Map();
  for (const s of sortedSeries) {
    for (const d of s.data) {
      if (!categorySet.has(d.name)) {
        categorySet.add(d.name);
        categories.push(d.name);
      }
      if (d.drilldown && d.meta && !categoryMetaMap.value.has(d.name)) {
        categoryMetaMap.value.set(d.name, d.meta as EWorkloadTelemetryPointMeta);
      }
    }
  }

  const series: BarSeriesOption[] = sortedSeries.map((s) => {
    const catMap = new Map(s.data.map((d) => [d.name, d]));
    const seriesName = s.name || "";
    return {
      type: "bar",
      id: `pending-workloads-by-nodepool-${seriesName}`,
      name: seriesName,
      data: categories.map((cat) => {
        const pt = catMap.get(cat);
        return pt
          ? {
              value: pt.y,
              groupId: parentGroupId ?? pt.meta?.entityId ?? cat,
              childGroupId: pt.meta?.entityId,
              drilldown: pt.drilldown,
              meta: pt.meta,
            }
          : { value: null, groupId: parentGroupId ?? cat };
      }),
      universalTransition: DRILLDOWN_UNIVERSAL_TRANSITION,
      itemStyle: s.color ? { color: s.color } : undefined,
      ...chartUtil.getBarSeriesDefaults(),
      label: {
        show: true,
        position: "top" as const,
        formatter: (params: { value?: unknown }) => {
          return params.value === 0 ? "" : String(params.value ?? "");
        },
        color: CHART_DATA_LABEL_COLOR,
        fontSize: 12,
        fontWeight: 400,
      },
    };
  });

  return { categories, series };
}

function buildChartOption(
  categories: string[],
  series: BarSeriesOption[],
  isDrilldown = false,
  animateDrilldown = false,
): EChartsOption {
  const hasDrilldown = categoryMetaMap.value.size > 0;
  drilldownAxisCategories.value = hasDrilldown ? categories : [];
  const yAxisConfig = chartUtil.getBarChartYAxisConfig(series, { targetTicks: 2 });
  const groupedBarLayout = getGroupedBarLayout(categories.length, series.length);
  return {
    ...DRILLDOWN_CHART_ANIMATION_PRESET,
    ...(animateDrilldown ? getDrilldownTransitionPreset() : {}),
    aria: chartUtil.buildChartAria("Pending workloads by node pool. Select a category to view its details."),
    grid: isDrilldown ? drilldownChartGridOptions : chartGridOptions,
    xAxis: {
      type: "category",
      data: categories,
      axisLabel: hasDrilldown
        ? {
            ...getCategoryAxisLabelOptions(categories.length),
            rotate: 0,
            color: ACTIVE_AXIS_LABEL_COLOR,
            margin: 16,
            show: false,
            fontSize: 11,
          }
        : {
            ...getCategoryAxisLabelOptions(categories.length),
            show: true,
            rotate: 0,
            margin: 10,
            color: "#666",
            fontSize: 11,
            interval: 0,
            overflow: "truncate",
            ellipsis: "...",
            lineHeight: 16,
            hideOverlap: false,
            formatter: chartUtil.formatCategoricalAxisLabel,
          },
      axisLine: { show: true, lineStyle: { color: "#ccd6eb", width: 1 } },
      axisTick: { show: false },
      axisPointer: chartUtil.getHiddenAxisPointer(),
      splitLine: { show: false },
      triggerEvent: hasDrilldown,
    },
    yAxis: {
      type: "value",
      min: 0,
      interval: yAxisConfig.interval,
      max: yAxisConfig.max,
      axisLabel: { margin: 12, color: "#666", fontSize: 11, formatter: chartUtil.formatCompactAxisValue },
      axisLine: { show: false },
      axisTick: { show: false },
      axisPointer: chartUtil.getHiddenAxisPointer(),
      splitLine: {
        show: true,
        lineStyle: { color: "#e6e6e6", width: 1, type: "solid" },
      },
      ...yAxisTitleOptions,
    },
    tooltip: {
      trigger: "axis",
      axisPointer: chartUtil.getHiddenAxisPointer(),
      ...STANDARD_TOOLTIP_STYLE,
      confine: true,
      padding: [7, 8],
      formatter: formatTooltip,
      position: (
        point: number[],
        _params: unknown,
        _dom: HTMLElement,
        _rect: unknown,
        size: { contentSize: number[]; viewSize: number[] },
      ) => {
        const [contentWidth, contentHeight] = size.contentSize;
        const [viewWidth] = size.viewSize;
        const horizontalMargin = 14;
        const maxLeft = Math.max(horizontalMargin, viewWidth - contentWidth - horizontalMargin);
        return [
          Math.max(horizontalMargin, Math.min(point[0] - contentWidth / 2, maxLeft)),
          Math.max(0, point[1] - contentHeight - 12),
        ];
      },
    } as unknown as EChartsOption["tooltip"],
    legend: legendOptions,
    series: series.map((item) => ({ ...item, ...groupedBarLayout })),
  };
}

const wrapperOptions: IWidgetWrapperOptions = {
  title: "Pending workloads by node pool",
  tooltipText: "Distribution of pending workloads across node pools",
  emptyMessage: EWidgetEmptyMessages.NoPendingWorkloads,
};

const clusterName = computed(() => clusterStore.clusterNameByUuid(props.clusterId) || "");

const departmentIdFromSource = computed(() => {
  if (props.orgUnitMetricSource?.type === EOrgUnitMetricSourceType.Department) {
    return props.orgUnitMetricSource.id;
  }
  return undefined;
});

const projectIdFromSource = computed(() => {
  if (props.orgUnitMetricSource?.type === EOrgUnitMetricSourceType.Project) {
    return props.orgUnitMetricSource.id;
  }
  return undefined;
});

async function loadInitialChartData(): Promise<void> {
  const requestId = ++latestRequestId.value;
  try {
    isDataLoading.value = true;
    const res = await fetchTelemetry(EWorkloadTelemetryGroupBy.ClusterId);
    if (requestId !== latestRequestId.value) return;

    const enrichedRes = widgetUtil.enrichTelemetryResponseByKeyValue(
      res,
      EWorkloadTelemetryGroupBy.ClusterId,
      clusterName.value,
      "name",
    );

    const hcSeries = widgetUtil.aggregateTelemetryGroupedData(
      enrichedRes,
      [EWorkloadTelemetryGroupBy.ClusterId, EWorkloadTelemetryGroupBy.RequestedNodepools],
      true,
    ) as unknown as HcSeriesItem[];

    const { categories, series } = convertAndColorSeries(hcSeries);

    if (series.length === 0) {
      isChartEmpty.value = true;
    } else {
      isChartEmpty.value = false;
      chartOption.value = buildChartOption(categories, series);
    }

    drillStack.value = [];
    drillBreadcrumbs.value = [];
    currentDrillMeta.value = null;
    error.value = false;
  } catch (err: unknown) {
    if (requestId === latestRequestId.value) handleError(err);
  } finally {
    if (requestId === latestRequestId.value) {
      isDataLoading.value = false;
      displayLoading.value = false;
    }
  }
}

function handleChartClick(params: unknown): void {
  if (isDataLoading.value) return;

  const eventParams = params as { componentType?: string; seriesIndex?: number; dataIndex?: number; value?: unknown };
  const axisValue = typeof eventParams.value === "string" ? eventParams.value : undefined;
  if (eventParams.componentType === "xAxis" && axisValue) {
    const meta = categoryMetaMap.value.get(axisValue);
    if (meta) {
      drillDownTo(meta);
    }
    return;
  }

  if (eventParams.seriesIndex === undefined || eventParams.dataIndex === undefined) return;

  const seriesData = chartOption.value?.series as BarSeriesOption[] | undefined;
  if (!seriesData?.[eventParams.seriesIndex]) return;

  const dataArr = seriesData[eventParams.seriesIndex].data as
    | Array<{ value: number; drilldown?: boolean; meta?: EWorkloadTelemetryPointMeta }>
    | undefined;
  const point = dataArr?.[eventParams.dataIndex];
  if (!point?.drilldown || !point?.meta) return;

  drillDownTo(point.meta);
}

async function drillDownTo(meta: EWorkloadTelemetryPointMeta): Promise<void> {
  if (isDataLoading.value) return;
  const requestId = ++latestRequestId.value;

  try {
    isDataLoading.value = true;

    const previousDrillState: DrillState = {
      option: { ...chartOption.value } as EChartsOption,
      meta: currentDrillMeta.value,
      categoryMetaMap: new Map(categoryMetaMap.value),
      drilldownAxisCategories: [...drilldownAxisCategories.value],
    };
    const nextBreadcrumbs = getNextChartDrilldownBreadcrumbs(drillBreadcrumbs.value, clusterName.value);

    const { departmentIdFilter, includeDrilldown, groupByKey } = widgetUtil.getWorkloadTelemetryDrilldownOptions(
      meta,
      props.hasDepartments,
    );
    const prevGroupByKey = getPreviousGroupByKey(groupByKey);
    const res = await fetchTelemetry(groupByKey, departmentIdFilter);
    if (requestId !== latestRequestId.value) return;

    const hcSeries = widgetUtil.aggregateTelemetryGroupedData(
      res,
      [groupByKey, EWorkloadTelemetryGroupBy.RequestedNodepools],
      includeDrilldown,
      prevGroupByKey,
      departmentIdFilter,
    ) as unknown as HcSeriesItem[];

    const { categories, series } = convertAndColorSeries(hcSeries, true, meta.entityId);
    chartOption.value = buildChartOption(categories, series, true, true);
    currentDrillMeta.value = meta;

    const breadcrumbLabel =
      props.hasDepartments && meta.groupByKey === EWorkloadTelemetryGroupBy.ClusterId ? "Departments" : "Projects";
    const navigation = completeChartDrilldown(
      drillStack.value,
      previousDrillState,
      nextBreadcrumbs,
      breadcrumbLabel,
      categories.length,
    );
    drillStack.value = navigation.stack;
    drillBreadcrumbs.value = navigation.breadcrumbs;
    drilldownAnnouncement.value = navigation.announcement;

    error.value = false;
  } catch (err: unknown) {
    if (requestId === latestRequestId.value) handleError(err);
  } finally {
    if (requestId === latestRequestId.value) isDataLoading.value = false;
  }
}

function drillUpTo(index: number): void {
  const navigation = getChartDrillUpNavigation(drillStack.value, drillBreadcrumbs.value, index);
  if (!navigation) return;

  const { targetState } = navigation;
  chartOption.value = { ...targetState.option, ...getDrilldownTransitionPreset() };
  currentDrillMeta.value = targetState.meta;
  categoryMetaMap.value = new Map(targetState.categoryMetaMap);
  drilldownAxisCategories.value = [...targetState.drilldownAxisCategories];
  drillStack.value = navigation.stack;
  drillBreadcrumbs.value = navigation.breadcrumbs;
  drilldownAnnouncement.value = navigation.announcement;
}

function drillDownByCategory(category: string): void {
  const meta = categoryMetaMap.value.get(category);
  if (meta) {
    drillDownTo(meta);
  }
}

async function fetchTelemetry(
  groupByKey: EWorkloadTelemetryGroupByWithoutBucket,
  departmentId?: string,
): Promise<TelemetryResponseValuesInner[]> {
  const deptId = departmentId ?? departmentIdFromSource.value;
  const res = await workloadService.getWorkloadsTelemetry(
    WorkloadTelemetryType.WorkloadsCount,
    props.clusterId,
    props.nodePoolName,
    deptId,
    projectIdFromSource.value,
    [groupByKey, EWorkloadTelemetryGroupBy.RequestedNodepools],
    [Phase.Pending],
  );
  return res.values;
}

function getPreviousGroupByKey(groupByKey: EWorkloadTelemetryGroupBy): EWorkloadTelemetryGroupBy {
  return groupByKey === EWorkloadTelemetryGroupBy.DepartmentId
    ? EWorkloadTelemetryGroupBy.ClusterId
    : EWorkloadTelemetryGroupBy.DepartmentId;
}

function handleError(err: unknown): void {
  error.value = true;
  if (err instanceof HttpErrorResponse) {
    console.error(err.message);
  } else {
    console.error("Error fetching data", err);
  }
}

loadInitialChartData();

onMounted(() => {
  // no-op — drilldown is handled via @click
});

watch(
  () => [props.clusterId, props.nodePoolName],
  () => {
    displayLoading.value = true;
    loadInitialChartData();
  },
);

watch(
  () => props.orgUnitMetricSource,
  () => {
    displayLoading.value = true;
    loadInitialChartData();
  },
  { deep: true },
);
</script>

<style lang="scss" scoped>
.pending-workloads-by-nodepool-chart-area {
  position: relative;
  height: 240px;
}

.pending-workloads-by-nodepool-chart {
  height: 240px;
}

.pending-workloads-by-nodepool-axis-labels {
  position: absolute;
  right: 20px;
  bottom: 51px;
  left: 72px;
  display: flex;
  pointer-events: none;
}

.pending-workloads-by-nodepool-axis-label {
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
</style>
