<template>
  <chart-widget-wrapper
    class="project-allocation-ratio-widget"
    :options="wrapperOptions"
    :loading="displayLoading"
    :error="error"
    :empty="isChartEmpty"
    hide-actions
  >
    <v-chart
      aid="project-allocation-ratio-chart"
      v-if="chartOption"
      ref="chartRef"
      :option="chartOption"
      autoresize
      style="height: 240px"
      key="project-allocation-ratio-chart"
    />
  </chart-widget-wrapper>
</template>

<script setup lang="ts">
import "@/plugins/echarts";
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from "vue";
import VChart from "vue-echarts";
import type { EChartsOption } from "echarts";

// models
import {
  EWidgetEmptyMessages,
  type IWidgetWrapperOptions,
  CHART_DATA_LABEL_COLOR,
  chartColorPaletteArray,
} from "@/models/chart.model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

// components
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";

// services
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { RankingMetricType, RankingOrderType, type ProjectGetRanking } from "@/swagger-models/org-unit-service-client";
import { chartUtil, STANDARD_TOOLTIP_STYLE } from "@/utils/chart.util";

export interface Props {
  clusterId: string;
  nodePoolName?: string;
  orgUnitMetricSource?: IOrgUnitMetricSource | null;
}

const props = withDefaults(defineProps<Props>(), {
  orgUnitMetricSource: null,
});

const displayLoading = ref(true);
const chartOption = ref<EChartsOption | null>(null);
const error = ref(false);
const isChartEmpty = ref(false);
const latestReqId = ref(0);
const chartRef = ref<InstanceType<typeof VChart> | null>(null);
const containerWidth = ref(0);
const currentCategories = ref<string[]>([]);
let resizeObserver: ResizeObserver | null = null;

const wrapperOptions: IWidgetWrapperOptions = {
  title: "Projects with lowest allocation ratio",
  tooltipText: "Projects ranked by GPU allocation / quota ratio (lowest first)",
  emptyMessage: EWidgetEmptyMessages.Default,
};

const departmentIdFromSource = computed(() => {
  if (props.orgUnitMetricSource?.type === EOrgUnitMetricSourceType.Department) {
    return props.orgUnitMetricSource.id;
  }
  return undefined;
});

interface AllocationDataPoint {
  value: number;
  quota: number;
  allocated: number;
}

const LABEL_CHAR_WIDTH = 6;
const LABEL_LINE_HEIGHT = 14;
const MIN_WRAPPABLE_LABEL_WIDTH = 42;
const BASE_GRID_BOTTOM = 87;
const WRAPPED_LABEL_GRID_OFFSET = 72;
const MAX_GRID_BOTTOM = 150;
const MAX_LABEL_WIDTH = 124;

const chartGridOptions = {
  top: 11,
  bottom: BASE_GRID_BOTTOM,
  left: 56,
  right: 12,
};

const yAxisTitleOptions = {
  name: "Ratio",
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

function getEffectiveWidth(): number {
  return containerWidth.value || 0;
}

function computePerCategoryWidth(categoryCount: number): number {
  const width = getEffectiveWidth();
  if (categoryCount <= 0 || width <= 0) return MAX_LABEL_WIDTH;
  const plotWidth = width - chartGridOptions.left - chartGridOptions.right;
  return Math.floor(plotWidth / categoryCount);
}

function getLabelWidth(categoryCount: number): number {
  return Math.max(1, Math.min(MAX_LABEL_WIDTH, computePerCategoryWidth(categoryCount)));
}

function getLabelLineCapacity(categoryCount: number): number {
  return Math.max(1, Math.floor(getLabelWidth(categoryCount) / LABEL_CHAR_WIDTH));
}

function wrapLongSegment(segment: string, lineCapacity: number): string[] {
  const lines: string[] = [];
  for (let start = 0; start < segment.length; start += lineCapacity) {
    lines.push(segment.slice(start, start + lineCapacity));
  }
  return lines;
}

function wrapCategoryLabel(value: string, lineCapacity: number): string {
  if (!value) return "";

  const segments = value.split("-").map((segment, index, all) => (index < all.length - 1 ? `${segment}-` : segment));
  const lines: string[] = [];
  let line = "";

  for (const segment of segments) {
    if (segment.length > lineCapacity) {
      if (line) lines.push(line);
      const longSegmentLines = wrapLongSegment(segment, lineCapacity);
      lines.push(...longSegmentLines.slice(0, -1));
      line = longSegmentLines.at(-1) || "";
    } else if (!line || line.length + segment.length <= lineCapacity) {
      line += segment;
    } else {
      lines.push(line);
      line = segment;
    }
  }
  if (line) lines.push(line);
  return lines.join("\n");
}

function usesCompactLabels(categoryCount: number): boolean {
  return getLabelWidth(categoryCount) < MIN_WRAPPABLE_LABEL_WIDTH;
}

function getGridOptions(categories: string[]): typeof chartGridOptions {
  if (usesCompactLabels(categories.length)) return chartGridOptions;
  const lineCapacity = getLabelLineCapacity(categories.length);
  const lineCount = Math.max(
    ...categories.map((category) => wrapCategoryLabel(category, lineCapacity).split("\n").length),
    1,
  );
  return {
    ...chartGridOptions,
    bottom: Math.min(
      MAX_GRID_BOTTOM,
      Math.max(BASE_GRID_BOTTOM, WRAPPED_LABEL_GRID_OFFSET + lineCount * LABEL_LINE_HEIGHT),
    ),
  };
}

function getXAxisLabel(categoryCount: number): Record<string, unknown> {
  const labelWidth = getLabelWidth(categoryCount);
  const compactLabels = usesCompactLabels(categoryCount);
  return {
    interval: 0,
    rotate: 0,
    margin: 10,
    color: "#666",
    fontSize: 11,
    width: labelWidth,
    overflow: compactLabels ? "truncate" : "break",
    ellipsis: "…",
    lineHeight: LABEL_LINE_HEIGHT,
    hideOverlap: compactLabels,
    ...(compactLabels
      ? {}
      : { formatter: (value: string) => wrapCategoryLabel(value, getLabelLineCapacity(categoryCount)) }),
  };
}

function initChartOption(): void {
  chartOption.value = {
    color: chartColorPaletteArray,
    grid: chartGridOptions,
    xAxis: {
      type: "category",
      data: [] as string[],
      axisLabel: getXAxisLabel(0),
      axisLine: { show: true, lineStyle: { color: "#ccd6eb", width: 1 } },
      axisTick: { show: false },
      splitLine: { show: false },
    },
    yAxis: {
      type: "value",
      min: 0,
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
      trigger: "axis",
      axisPointer: chartUtil.getHiddenAxisPointer(),
      ...STANDARD_TOOLTIP_STYLE,
    },
    legend: { ...legendOptions, data: ["Allocation ratio"] },
    series: [],
  };
}

function buildFullChartOption(
  categories: string[],
  data: AllocationDataPoint[],
  yAxisConfig: Record<string, unknown>,
): EChartsOption {
  return {
    color: chartColorPaletteArray,
    grid: getGridOptions(categories),
    xAxis: {
      type: "category",
      data: categories,
      axisLabel: getXAxisLabel(categories.length),
      axisLine: { show: true, lineStyle: { color: "#ccd6eb", width: 1 } },
      axisTick: { show: false },
      splitLine: { show: false },
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
    series: [
      {
        type: "bar",
        name: "Allocation ratio",
        data,
        barMaxWidth: 30,
        barCategoryGap: "40%",
        ...chartUtil.getBarSeriesDefaults(),
        label: {
          show: true,
          position: "top",
          formatter: "{c}%",
          color: CHART_DATA_LABEL_COLOR,
          fontSize: 12,
          fontWeight: 400,
        },
      },
    ],
    tooltip: {
      trigger: "axis",
      axisPointer: chartUtil.getHiddenAxisPointer(),
      ...STANDARD_TOOLTIP_STYLE,
      formatter: (params: unknown) => {
        const items = Array.isArray(params) ? params : [params];
        const param = items[0] as {
          color: string;
          seriesName: string;
          data: AllocationDataPoint;
          axisValueLabel?: string;
          name?: string;
        };
        if (!param?.data) return "";
        const header = chartUtil.escapeTooltipHtml(param.axisValueLabel || param.name || "");
        const seriesName = chartUtil.escapeTooltipHtml(param.seriesName);
        const value = chartUtil.escapeTooltipHtml(param.data.value);
        const quota = chartUtil.escapeTooltipHtml(param.data.quota);
        const allocated = chartUtil.escapeTooltipHtml(param.data.allocated);
        return `${header ? `<div style="padding-bottom: 4px;">${header}</div>` : ""}${chartUtil.renderTooltipDot(param.color)}${seriesName}: <b>${value}%</b><br/>GPU Quota: <b>${quota}</b><br/>GPU Allocated: <b>${allocated}</b>`;
      },
    },
    legend: { ...legendOptions, data: ["Allocation ratio"] },
  };
}

let lastData: AllocationDataPoint[] = [];
let lastYAxisConfig: Record<string, unknown> = {};

async function loadChartData(): Promise<void> {
  if (chartOption.value === null) return;
  const reqId = ++latestReqId.value;
  try {
    const projects = await orgUnitService.getProjectsRanking(
      RankingMetricType.GpuAllocationRatio,
      RankingOrderType.Bottom,
      props.clusterId,
      props.nodePoolName,
      departmentIdFromSource.value,
    );

    if (reqId !== latestReqId.value) return;

    if (projects.length === 0) {
      isChartEmpty.value = true;
    } else {
      isChartEmpty.value = false;
      const categories = projects.map((project: ProjectGetRanking) => project.name || "");
      const data: AllocationDataPoint[] = projects.map((project: ProjectGetRanking) => {
        const ratio = project.allocationValue ?? 0;
        const quota = project.gpuQuota ?? 0;
        const allocated = Math.round(ratio * quota * 100) / 100;
        return { value: Math.round(ratio * 100), quota, allocated };
      });
      const yAxisConfig = chartUtil.getBarChartYAxisConfig([{ data }]);

      currentCategories.value = categories;
      lastData = data;
      lastYAxisConfig = yAxisConfig;
      chartOption.value = buildFullChartOption(categories, data, yAxisConfig);
    }

    error.value = false;
  } catch (err: unknown) {
    if (reqId !== latestReqId.value) return;
    error.value = true;
    if (err instanceof HttpErrorResponse) {
      console.error(err.message);
    } else {
      console.error("Error fetching data", err);
    }
  } finally {
    if (reqId === latestReqId.value) {
      displayLoading.value = false;
    }
  }
}

function rebuildChart(): void {
  if (currentCategories.value.length === 0 || !chartOption.value) return;
  chartOption.value = buildFullChartOption(currentCategories.value, lastData, lastYAxisConfig);
}

function readContainerWidth(): void {
  const el = (chartRef.value as unknown as { $el?: HTMLElement })?.$el;
  if (el && el.clientWidth > 0) {
    containerWidth.value = el.clientWidth;
  }
}

function setupResizeObserver(): void {
  const el = (chartRef.value as unknown as { $el?: HTMLElement })?.$el;
  if (!el) return;
  containerWidth.value = el.clientWidth;
  rebuildChart();
  resizeObserver = new ResizeObserver((entries) => {
    const entry = entries[0];
    if (entry) {
      const newWidth = entry.contentRect.width;
      if (Math.abs(newWidth - containerWidth.value) > 2) {
        containerWidth.value = newWidth;
        rebuildChart();
      }
    }
  });
  resizeObserver.observe(el);
}

initChartOption();
loadChartData();

onMounted(async () => {
  await nextTick();
  readContainerWidth();
  rebuildChart();
  setupResizeObserver();
});

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
  resizeObserver = null;
});

watch(
  () => [props.clusterId, props.nodePoolName],
  () => {
    displayLoading.value = true;
    loadChartData();
  },
);

watch(
  () => props.orgUnitMetricSource,
  () => {
    displayLoading.value = true;
    loadChartData();
  },
  { deep: true },
);

watch(chartRef, () => {
  if (chartRef.value && !resizeObserver) {
    setupResizeObserver();
  }
});
</script>
