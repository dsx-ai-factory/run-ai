<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <histogram-metric-widget
    ref="histogramMetricWidget"
    :config="widgetConfig"
    :filter-by-dates="filterByDates"
    :get-data-args="getDataArgs"
    @register="handleRegister"
    @unregister="$emit('unregister')"
    @on-zoom="$emit('on-zoom', $event)"
    @on-reset-zoom="$emit('on-reset-zoom', $event)"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { HistogramMetricWidget } from "@/components/dashboard-v2/widgets/chart-widgets/histogram-metric-widget";
import {
  workloadTimeToFirstTokenWidgetConfig,
  podTimeToFirstTokenWidgetConfig,
} from "./time-to-first-token-widget-config";
import type { IGenericMetricWidgetConfigWithCompositeResponse } from "@/models/chart.model";
import type { IRangeDates } from "@/utils/date.util";
import { chartUtil } from "@/utils/chart.util";
import { shouldHideFinalNimXAxisLabel } from "@/components/dashboard-v2/widgets/nim-widgets/nim-chart-legend.util";
import type { CustomSeriesOption, EChartsOption, SeriesOption } from "echarts";
import type { BarSeriesOption } from "echarts/charts";

type THistogramMetricWidgetInstance = InstanceType<typeof HistogramMetricWidget> & { chartOptions?: EChartsOption };
type TCustomHistogramPoint = [number, number, number, number];

const HISTOGRAM_BAR_WIDTH = 1;

interface ICustomSeriesRenderApi {
  value(dimension: number): unknown;
  coord(data: [unknown, unknown]): number[];
  style(): Record<string, unknown>;
}

function renderHistogramBar(_params: unknown, api: ICustomSeriesRenderApi): Record<string, unknown> | null {
  const value = Number(api.value(1));
  if (!Number.isFinite(value) || value <= 0) return null;

  const timestamp = api.value(0);
  const start = api.value(2);
  const end = api.value(3);
  const x = Math.round(api.coord([timestamp, 0])[0]);
  const yStart = api.coord([timestamp, start])[1];
  const yEnd = api.coord([timestamp, end])[1];

  return {
    type: "rect",
    shape: {
      x,
      y: Math.min(yStart, yEnd),
      width: HISTOGRAM_BAR_WIDTH,
      height: Math.max(0.5, Math.abs(yStart - yEnd)),
    },
    style: api.style(),
  };
}

export default defineComponent({
  name: "time-to-first-token-widget",
  components: {
    HistogramMetricWidget,
  },
  emits: ["register", "unregister", "on-zoom", "on-reset-zoom"],
  props: {
    workloadId: {
      type: String as PropType<string>,
      required: true,
    },
    podId: {
      type: String as PropType<string>,
      required: false,
    },
    filterByDates: {
      type: Object as PropType<IRangeDates>,
      required: true,
    },
  },
  computed: {
    widgetConfig(): IGenericMetricWidgetConfigWithCompositeResponse {
      return this.podId ? podTimeToFirstTokenWidgetConfig : workloadTimeToFirstTokenWidgetConfig;
    },
    getDataArgs(): string[] {
      const args: string[] = [this.workloadId];
      if (this.podId) {
        args.push(this.podId);
      }

      return args;
    },
  },
  methods: {
    handleRegister(): void {
      this.applyChartParityOptions();
      this.$emit("register");
    },
    applyChartParityOptions(): void {
      const widget = this.$refs.histogramMetricWidget as THistogramMetricWidgetInstance | undefined;
      const chartOptions = widget?.chartOptions;
      if (!widget || !chartOptions) return;

      const series = Array.isArray(chartOptions.series)
        ? this.createChartSeries(chartOptions.series)
        : chartOptions.series;

      widget.chartOptions = {
        ...chartOptions,
        legend: this.getLegendOptions(chartOptions),
        xAxis: this.getXAxisOptions(chartOptions),
        yAxis: this.getYAxisOptions(chartOptions),
        series,
      };
    },
    createChartSeries(series: SeriesOption[]): SeriesOption[] {
      const anchorSeries = series.filter((item) => chartUtil.isConnectedTooltipAnchorSeries(item));
      const histogramSeries = series.filter(
        (item) => !chartUtil.isConnectedTooltipAnchorSeries(item),
      ) as BarSeriesOption[];

      return [...anchorSeries, ...this.createCustomHistogramSeries(histogramSeries)];
    },
    createCustomHistogramSeries(series: BarSeriesOption[]): CustomSeriesOption[] {
      const remainingByTimestamp = series.reduce((totalsByTimestamp, currSeries) => {
        const data = Array.isArray(currSeries.data) ? currSeries.data : [];
        data.forEach((point) => {
          const timestamp = Array.isArray(point) ? Number(point[0]) : Number(point);
          const value = Array.isArray(point) ? Number(point[1]) : 0;
          const segmentValue = Number.isFinite(value) && value > 0 ? value : 0;
          totalsByTimestamp.set(timestamp, (totalsByTimestamp.get(timestamp) || 0) + segmentValue);
        });
        return totalsByTimestamp;
      }, new Map<number, number>());

      return series.map((currSeries) => {
        const data = Array.isArray(currSeries.data) ? currSeries.data : [];
        const customData: TCustomHistogramPoint[] = data.map((point) => {
          const timestamp = Array.isArray(point) ? Number(point[0]) : Number(point);
          const value = Array.isArray(point) ? Number(point[1]) : 0;
          const segmentValue = Number.isFinite(value) && value > 0 ? value : 0;
          const end = remainingByTimestamp.get(timestamp) || 0;
          const start = Math.max(0, end - segmentValue);
          remainingByTimestamp.set(timestamp, start);
          return [timestamp, segmentValue, start, end];
        });

        return {
          name: currSeries.name,
          ...chartUtil.getBarSeriesLegendHoverState(),
          type: "custom" as const,
          data: customData,
          renderItem: renderHistogramBar,
          encode: {
            x: 0,
            y: [2, 3],
            tooltip: [1],
          },
          clip: true,
          itemStyle: currSeries.itemStyle,
        } as unknown as CustomSeriesOption;
      });
    },
    getLegendOptions(chartOptions: EChartsOption): EChartsOption["legend"] {
      const legend = Array.isArray(chartOptions.legend) ? chartOptions.legend[0] : chartOptions.legend;
      const prevTextStyle = (legend as Record<string, unknown>)?.textStyle as Record<string, unknown> | undefined;
      return {
        ...(legend as Record<string, unknown>),
        bottom: 16,
        textStyle: { ...prevTextStyle, fontSize: 12, fontWeight: "bold" },
      } as EChartsOption["legend"];
    },
    getXAxisOptions(chartOptions: EChartsOption): EChartsOption["xAxis"] {
      const xAxis = Array.isArray(chartOptions.xAxis) ? chartOptions.xAxis[0] : chartOptions.xAxis;
      return {
        ...(xAxis as Record<string, unknown>),
        axisLabel: {
          ...((xAxis as Record<string, unknown> | undefined)?.axisLabel as Record<string, unknown> | undefined),
          formatter: this.formatXAxisTimestamp,
        },
      } as EChartsOption["xAxis"];
    },
    getYAxisOptions(chartOptions: EChartsOption): EChartsOption["yAxis"] {
      const yAxis = Array.isArray(chartOptions.yAxis) ? chartOptions.yAxis[0] : chartOptions.yAxis;

      return {
        ...(yAxis as Record<string, unknown>),
        axisLabel: {
          ...((yAxis as Record<string, unknown> | undefined)?.axisLabel as Record<string, unknown> | undefined),
          formatter: this.formatYAxisValue,
        },
      } as EChartsOption["yAxis"];
    },
    formatXAxisTimestamp(value: number | string): string {
      if (shouldHideFinalNimXAxisLabel(value, this.filterByDates)) return "";
      return chartUtil.formatDatetimeAxisLabel(value);
    },
    formatYAxisValue(value: number | string): string {
      const numericValue = Number(value);
      if (!Number.isFinite(numericValue)) return String(value);
      if (Math.abs(numericValue) < 1000) return String(numericValue);

      const valueInThousands = numericValue / 1000;
      return `${Number.isInteger(valueInThousands) ? valueInThousands : Number(valueInThousands.toFixed(1))}k`;
    },
  },
});
</script>
