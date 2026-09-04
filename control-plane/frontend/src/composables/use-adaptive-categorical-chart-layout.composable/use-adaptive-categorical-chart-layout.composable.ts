/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { onMounted, onUnmounted, ref, watch, type Ref } from "vue";
import type { BarSeriesOption } from "echarts";

type EChartsLayoutInstance = {
  getWidth: () => number;
  getOption: () => {
    xAxis?: Array<{ data?: unknown[] }>;
    series?: Array<{ id?: string }>;
  };
  setOption: (option: Record<string, unknown>, options?: { lazyUpdate?: boolean }) => void;
};

type EChartsRef = {
  chart?: EChartsLayoutInstance;
};

interface IAdaptiveCategoricalChartLayoutOptions {
  // vue-echarts exposes a richer component instance than this composable needs.
  // Keep the ref opaque at the boundary and narrow its chart API internally.
  chartRef: { readonly value: unknown };
  containerRef: Ref<HTMLElement | null>;
  horizontalGridPadding: number;
}

const LABEL_GAP = 8;
const LABEL_LINE_HEIGHT = 16;
// All categorical drilldown charts use a one-line legend in the same slot.
// Keeping the external category labels at a stable offset prevents a second
// layout pass from moving the plot or legend after a drilldown transition.
const LEGEND_BOTTOM_GAP = 16;
const LAYOUT_GAP = 16;
const LEGEND_LINE_HEIGHT = 16;
const EXTERNAL_CATEGORY_LABEL_BOTTOM = LEGEND_BOTTOM_GAP + LEGEND_LINE_HEIGHT + LAYOUT_GAP;
// Match Highcharts' column defaults used by the legacy dashboards:
// groupPadding = 0.2, pointPadding = 0.1, maxPointWidth = 30.
const GROUP_PADDING = 0.2;
const POINT_PADDING = 0.1;
const MAX_BAR_WIDTH = 30;
const DEFAULT_BAR_GAP_PERCENT = (POINT_PADDING * 2 * 100) / (1 - POINT_PADDING * 2);

function getChartInstance(value: unknown): EChartsLayoutInstance | undefined {
  if (!value || typeof value !== "object" || !("chart" in value)) return undefined;
  return (value as EChartsRef).chart;
}

export function useAdaptiveCategoricalChartLayout(options: IAdaptiveCategoricalChartLayoutOptions) {
  const layout = ref({
    externalCategoryLabelBottom: EXTERNAL_CATEGORY_LABEL_BOTTOM,
    categoryLabelWidth: 0,
  });

  let resizeObserver: ResizeObserver | null = null;
  let frameId: number | null = null;

  const getAvailablePlotWidth = (): number => {
    const chart = getChartInstance(options.chartRef.value);
    const containerWidth = options.containerRef.value?.clientWidth || 0;
    const chartWidth = containerWidth || chart?.getWidth() || 0;
    return Math.max(1, chartWidth - options.horizontalGridPadding);
  };

  const getCategoryLabelWidth = (categoryCount: number): number => {
    const availablePlotWidth = getAvailablePlotWidth();
    return Math.max(1, Math.floor(availablePlotWidth / Math.max(1, categoryCount)) - LABEL_GAP);
  };

  const getGroupedBarLayout = (categoryCount: number, seriesCount: number): Partial<BarSeriesOption> => {
    const chart = getChartInstance(options.chartRef.value);
    const hasMeasuredWidth = Boolean(options.containerRef.value?.clientWidth || chart?.getWidth());
    if (!hasMeasuredWidth) {
      return {
        barMaxWidth: MAX_BAR_WIDTH,
        barGap: `${DEFAULT_BAR_GAP_PERCENT}%`,
        barCategoryGap: `${GROUP_PADDING * 200}%`,
      };
    }

    const safeCategoryCount = Math.max(1, categoryCount);
    const safeSeriesCount = Math.max(1, seriesCount);
    const categoryBand = getAvailablePlotWidth() / safeCategoryCount;
    const groupWidth = categoryBand * (1 - GROUP_PADDING * 2);
    const seriesSlot = groupWidth / safeSeriesCount;
    const barWidth = Math.max(1, Math.min(MAX_BAR_WIDTH, seriesSlot * (1 - POINT_PADDING * 2)));
    const gapPixels = Math.max(0, seriesSlot - barWidth);
    const barGapPercent = safeSeriesCount > 1 ? (gapPixels / barWidth) * 100 : 0;

    return {
      barWidth: Number(barWidth.toFixed(2)),
      barMaxWidth: MAX_BAR_WIDTH,
      barGap: `${Number(barGapPercent.toFixed(2))}%`,
      barCategoryGap: `${GROUP_PADDING * 200}%`,
    };
  };

  const getCategoryAxisLabelOptions = (categoryCount: number) => ({
    width: getCategoryLabelWidth(categoryCount),
    overflow: "truncate" as const,
    ellipsis: "…",
    lineHeight: LABEL_LINE_HEIGHT,
    interval: 0,
  });

  const updateLayout = (): void => {
    frameId = null;
    const chart = getChartInstance(options.chartRef.value);
    const container = options.containerRef.value;
    if (!chart || !container) return;

    const currentOption = chart.getOption();
    const categoryCount = currentOption.xAxis?.[0]?.data?.length || 1;
    const series = currentOption.series || [];
    const categoryLabelWidth = getCategoryLabelWidth(categoryCount);
    const groupedBarLayout = getGroupedBarLayout(categoryCount, series.length);

    const previous = layout.value;
    if (previous.categoryLabelWidth === categoryLabelWidth) return;

    layout.value = { ...previous, categoryLabelWidth };
    // This path is reserved for a real container resize. Drilldown changes
    // receive the same options synchronously in their primary chart update.
    chart.setOption(
      {
        xAxis: { axisLabel: getCategoryAxisLabelOptions(categoryCount) },
        series: series.map(({ id }) => ({ id, ...groupedBarLayout })),
      },
      { lazyUpdate: true },
    );
  };

  const scheduleLayoutUpdate = (): void => {
    if (frameId !== null) cancelAnimationFrame(frameId);
    frameId = requestAnimationFrame(updateLayout);
  };

  const observeContainer = (container: HTMLElement | null): void => {
    resizeObserver?.disconnect();
    resizeObserver = null;
    if (!container || typeof ResizeObserver === "undefined") return;
    resizeObserver = new ResizeObserver(scheduleLayoutUpdate);
    resizeObserver.observe(container);
    scheduleLayoutUpdate();
  };

  onMounted(() => observeContainer(options.containerRef.value));
  watch(options.containerRef, observeContainer, { flush: "post" });

  onUnmounted(() => {
    resizeObserver?.disconnect();
    if (frameId !== null) cancelAnimationFrame(frameId);
  });

  return { layout, getCategoryAxisLabelOptions, getGroupedBarLayout };
}
