<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <chart-widget-wrapper
    class="quota-by-nodepool-widget"
    :options="wrapperOptions"
    :loading="displayLoading"
    :error="error"
    :empty="isChartEmpty"
    :announcement="drilldownAnnouncement"
    hide-actions
  >
    <div v-if="chartOption" ref="chartAreaRef" class="quota-by-nodepool-chart-area">
      <chart-drilldown-breadcrumbs :items="drillBreadcrumbs" @navigate="drillUpTo" />
      <v-chart
        aid="quota-by-nodepool-chart"
        ref="chartRef"
        :option="chartOption"
        autoresize
        class="quota-by-nodepool-chart"
        key="quota-by-nodepool-chart"
        @click="handleChartClick"
      />
      <div
        v-if="drilldownAxisCategories.length"
        class="quota-by-nodepool-axis-labels"
        role="group"
        aria-label="Explore chart data"
        :style="{ bottom: `${layout.externalCategoryLabelBottom}px` }"
      >
        <button
          v-for="category in drilldownAxisCategories"
          :key="category"
          class="quota-by-nodepool-axis-label"
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
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
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
import { HttpErrorResponse } from "@/models/http-response.model";
import { EIntervalLabels } from "@/models/interval.model";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

// components
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { ChartDrilldownBreadcrumbs } from "@/components/common/charts/chart-drilldown-breadcrumbs";

// services
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { OrgUnitTelemetryType, type TelemetryResponseValuesInner } from "@/swagger-models/org-unit-service-client";

// utils
import { widgetUtil } from "@/utils/widget.util";
import { intervalUtil } from "@/utils/interval.util";
import {
  chartUtil,
  completeChartDrilldown,
  getChartDrillUpNavigation,
  getNextChartDrilldownBreadcrumbs,
  STANDARD_TOOLTIP_STYLE,
} from "@/utils/chart.util";
import { useAdaptiveCategoricalChartLayout } from "@/composables/use-adaptive-categorical-chart-layout.composable";

const NODEPOOL_GROUP = "NodepoolId";
const CLUSTER_GROUP = "ClusterId";
const PARENT_GROUP = "ParentId";
const PROJECT_GROUP = "ProjectId";

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
  name: "GPUs",
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
    padding: [0, 0, 0, 2],
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

const displayLoading = ref(true);
const chartOption = ref<EChartsOption | null>(null);
const error = ref(false);
const isChartEmpty = ref(false);
const isDataLoading = ref(false);
const latestRequestId = ref(0);
const colorMap = new Map<string, string>();
let nextColorIndex = 0;

interface DrillState {
  option: EChartsOption;
  meta: { groupByKey: string; entityId: string } | null;
  categoryMetaMap: Map<string, { entityId: string; groupByKey: string }>;
  drilldownAxisCategories: string[];
}
const drillStack = ref<DrillState[]>([]);
const drillBreadcrumbs = ref<string[]>([]);
const currentDrillMeta = ref<{ groupByKey: string; entityId: string } | null>(null);
const categoryMetaMap = ref<Map<string, { entityId: string; groupByKey: string }>>(new Map());
const drilldownAxisCategories = ref<string[]>([]);
const drilldownAnnouncement = ref("");
const chartRef = ref<InstanceType<typeof VChart> | null>(null);
const chartAreaRef = ref<HTMLElement | null>(null);
const { layout, getCategoryAxisLabelOptions, getGroupedBarLayout } = useAdaptiveCategoricalChartLayout({
  chartRef,
  containerRef: chartAreaRef,
  horizontalGridPadding: chartGridOptions.left + chartGridOptions.right,
});

const wrapperOptions: IWidgetWrapperOptions = {
  title: "GPU quota by node pool",
  tooltipText: "GPU quota distribution across node pools",
  emptyMessage: EWidgetEmptyMessages.Default,
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
  meta?: { groupByKey?: string; entityId?: string };
}

function getTooltipValue(point: { value?: unknown; data?: unknown }): number | null {
  const rawValue =
    typeof point.value === "object" && point.value !== null && "value" in point.value
      ? (point.value as { value?: unknown }).value
      : typeof point.data === "object" && point.data !== null && "value" in point.data
        ? (point.data as { value?: unknown }).value
        : point.value;

  if (rawValue === null || rawValue === undefined || rawValue === "") return null;

  const value = Number(rawValue);
  return Number.isFinite(value) ? value : null;
}

function formatTooltip(params: unknown): string {
  const items = Array.isArray(params) ? params : [params];
  const firstItem = items[0] as { axisValueLabel?: string; name?: string } | undefined;
  const header = chartUtil.escapeTooltipHtml(firstItem?.axisValueLabel || firstItem?.name || "");
  const headerHtml = header ? `<div style="padding-bottom: 4px;">${header}</div>` : "";

  const tooltip = items.reduce((html, item) => {
    const point = item as { color?: string; seriesName?: string; value?: unknown; data?: unknown };
    const value = getTooltipValue(point);
    if (value === null) return html;

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
        categoryMetaMap.value.set(d.name, d.meta);
      }
    }
  }

  const series: BarSeriesOption[] = sortedSeries.map((s) => {
    const catMap = new Map(s.data.map((d) => [d.name, d]));
    const seriesName = s.name || "";
    return {
      type: "bar",
      id: `quota-by-nodepool-${seriesName}`,
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
          const value = Number(params.value);
          return Number.isFinite(value) && value !== 0 ? String(params.value) : "";
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
    aria: chartUtil.buildChartAria("Quota by node pool. Select a category to view its details."),
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
      formatter: formatTooltip,
    },
    legend: legendOptions,
    series: series.map((item) => ({ ...item, ...groupedBarLayout })),
  };
}

function enrichNodepoolNames(values: TelemetryResponseValuesInner[]): TelemetryResponseValuesInner[] {
  return values.map((item) => ({
    ...item,
    groups: item.groups?.map((group) => {
      if (group.key === NODEPOOL_GROUP) {
        const displayName = group.name || group.value;
        return { ...group, value: displayName, name: displayName };
      }
      return group;
    }),
  }));
}

async function loadInitialChartData(): Promise<void> {
  const requestId = ++latestRequestId.value;
  try {
    isDataLoading.value = true;
    const res = await fetchTelemetry([CLUSTER_GROUP, NODEPOOL_GROUP]);
    if (requestId !== latestRequestId.value) return;

    const clusterEnriched = widgetUtil.enrichTelemetryResponseByKeyValue(res, CLUSTER_GROUP, clusterName.value, "name");
    const enrichedRes = enrichNodepoolNames(clusterEnriched);

    const hcSeries = widgetUtil.aggregateTelemetryGroupedData(
      enrichedRes,
      [CLUSTER_GROUP, NODEPOOL_GROUP],
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
    | Array<{ value: number; drilldown?: boolean; meta?: { entityId: string; groupByKey: string } }>
    | undefined;
  const point = dataArr?.[eventParams.dataIndex];
  if (!point?.drilldown || !point?.meta) return;

  drillDownTo(point.meta);
}

async function drillDownTo(meta: { entityId: string; groupByKey: string }): Promise<void> {
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

    const { nextGroupBy, departmentId, includeDrilldown } = getDrilldownOptionsFromMeta(meta);

    const res = await fetchTelemetry(
      [nextGroupBy as "ClusterId" | "NodepoolId" | "ParentId" | "ProjectId", NODEPOOL_GROUP],
      departmentId,
    );
    if (requestId !== latestRequestId.value) return;
    const enrichedRes = enrichNodepoolNames(res);

    const hcSeries = widgetUtil.aggregateTelemetryGroupedData(
      enrichedRes,
      [nextGroupBy, NODEPOOL_GROUP],
      includeDrilldown,
      meta.groupByKey,
      meta.entityId,
    ) as unknown as HcSeriesItem[];

    const { categories, series } = convertAndColorSeries(hcSeries, true, meta.entityId);
    chartOption.value = buildChartOption(categories, series, true, true);
    currentDrillMeta.value = meta;

    const breadcrumbLabel = props.hasDepartments && meta.groupByKey === CLUSTER_GROUP ? "Departments" : "Projects";
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

  latestRequestId.value++;
  isDataLoading.value = false;

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
  groupBy: Array<"ClusterId" | "NodepoolId" | "ParentId" | "ProjectId">,
  departmentId?: string,
): Promise<TelemetryResponseValuesInner[]> {
  const deptId = departmentId ?? departmentIdFromSource.value;
  const res = await orgUnitService.getProjectsTelemetry(
    OrgUnitTelemetryType.GpuQuota,
    props.clusterId,
    props.nodePoolName,
    deptId,
    projectIdFromSource.value,
    groupBy,
  );
  return res.values;
}

async function refreshChart(): Promise<void> {
  if (isDataLoading.value) return;

  if (currentDrillMeta.value) {
    await refreshDrilldownData();
  } else {
    await loadInitialChartData();
  }
}

async function refreshDrilldownData(): Promise<void> {
  if (!currentDrillMeta.value) return;
  const requestId = ++latestRequestId.value;
  const meta = currentDrillMeta.value;
  try {
    const { nextGroupBy, departmentId, includeDrilldown } = getDrilldownOptionsFromMeta(meta);

    const res = await fetchTelemetry(
      [nextGroupBy as "ClusterId" | "NodepoolId" | "ParentId" | "ProjectId", NODEPOOL_GROUP],
      departmentId,
    );
    if (requestId !== latestRequestId.value) return;

    const enrichedRes = enrichNodepoolNames(res);
    const hcSeries = widgetUtil.aggregateTelemetryGroupedData(
      enrichedRes,
      [nextGroupBy, NODEPOOL_GROUP],
      includeDrilldown,
      meta.groupByKey,
      meta.entityId,
    ) as unknown as HcSeriesItem[];

    const { categories, series } = convertAndColorSeries(hcSeries, true, meta.entityId);
    chartOption.value = buildChartOption(categories, series, true);
    error.value = false;
  } catch (err: unknown) {
    if (requestId === latestRequestId.value) handleError(err);
  }
}

function getDrilldownOptionsFromMeta(meta: { groupByKey: string; entityId: string }): {
  nextGroupBy: string;
  departmentId: string | undefined;
  includeDrilldown: boolean;
} {
  if (meta.groupByKey === CLUSTER_GROUP && departmentIdFromSource.value) {
    return { nextGroupBy: PROJECT_GROUP, departmentId: departmentIdFromSource.value, includeDrilldown: false };
  }
  if (meta.groupByKey === CLUSTER_GROUP && props.hasDepartments) {
    return { nextGroupBy: PARENT_GROUP, departmentId: undefined, includeDrilldown: true };
  } else if (meta.groupByKey === CLUSTER_GROUP) {
    return { nextGroupBy: PROJECT_GROUP, departmentId: undefined, includeDrilldown: false };
  }
  return { nextGroupBy: PROJECT_GROUP, departmentId: meta.entityId, includeDrilldown: false };
}

function handleError(err: unknown): void {
  error.value = true;
  if (err instanceof HttpErrorResponse) {
    console.error(err.message);
  } else {
    console.error("Error fetching data", err);
  }
}

onMounted(() => {
  loadInitialChartData();
  intervalUtil.startInterval(EIntervalLabels.QuotaByNodepoolWidget, refreshChart);
});

onUnmounted(() => {
  latestRequestId.value++;
  intervalUtil.stopInterval(EIntervalLabels.QuotaByNodepoolWidget);
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
.quota-by-nodepool-chart-area {
  position: relative;
  height: 240px;
}

.quota-by-nodepool-chart {
  height: 240px;
}

.quota-by-nodepool-axis-labels {
  position: absolute;
  right: 20px;
  bottom: 51px;
  left: 72px;
  display: flex;
  pointer-events: none;
}

.quota-by-nodepool-axis-label {
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
