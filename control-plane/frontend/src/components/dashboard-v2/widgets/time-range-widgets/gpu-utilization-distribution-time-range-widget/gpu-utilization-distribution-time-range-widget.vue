<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <chart-widget-wrapper
    :options="wrapperOptions"
    :loading="isInitialLoader"
    hide-separator
    hide-actions
    hide-time-frame
    hide-border
    flat
    body-no-padding
  >
    <div class="time-range-chart-container">
      <chart-zoom-controls v-show="false" ref="zoomControls" @zoom-change="handleZoomControls" />
      <chart-area-zoom-controls :vertical-offset="10" @zoom-toggle="handleZoomToggle" @reset="handleAreaZoomReset" />
      <v-chart
        v-if="chartOptions"
        ref="chartRef"
        :id="chartId"
        :option="chartOptions"
        style="width: 100%; height: 250px"
        autoresize
        @datazoom="handleDataZoom"
        @restore="handleResetZoom"
        @zr:click="handleChartAreaClick"
      />
    </div>
  </chart-widget-wrapper>
</template>
<script lang="ts">
import "@/plugins/echarts";
import { defineComponent, type PropType } from "vue";
import VChart from "vue-echarts";
import ChartWidgetWrapper from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper/chart-widget-wrapper.vue";
import { ChartZoomControls } from "@/components/common/charts/chart-zoom-controls";
import { ChartAreaZoomControls } from "@/components/common/charts/chart-area-zoom-controls";
import {
  CHART_AXIS_LABEL_DEFAULTS,
  CHART_LEGEND_TEXT_STYLE_DEFAULTS,
  LEGEND_DASHED_LINE_ICON,
  GPU_UTILIZATION_DISTRIBUTION_TIME_RANGE_WIDGET_CHART_ID,
  INSIDE_DATAZOOM_CONFIG,
  chartColorPaletteArray,
  type IWidgetWrapperOptions,
  type TMeasurementsTimeAndValue,
} from "@/models/chart.model";
import type { EChartsOption, SeriesOption } from "echarts";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import type { IFilterModel } from "@/models/filter.model";
import { chartUtil, STANDARD_TOOLTIP_STYLE } from "@/utils/chart.util";
import {
  type GroupedMeasurementResponseValuesInner,
  type MetricsResponse,
  MetricsType,
} from "@/swagger-models/cluster-service-client";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { metricUtil } from "@/utils/metric.util";
import { roundToDecimal } from "@/utils/format.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
export interface RangeData {
  [key: string]: string;
}
export enum ESeriesType {
  Total = "Total",
  ZeroToTwentyFive = "0%-25%",
  TwentyFiveToFifty = "25%-50%",
  FiftyToSeventyFive = "50%-75%",
  SeventyFiveToOneHundred = "75%-100%",
}

const seriesColor: Record<ESeriesType, string> = {
  [ESeriesType.Total]: chartColorPaletteArray[0],
  [ESeriesType.ZeroToTwentyFive]: chartColorPaletteArray[4],
  [ESeriesType.TwentyFiveToFifty]: chartColorPaletteArray[3],
  [ESeriesType.FiftyToSeventyFive]: chartColorPaletteArray[2],
  [ESeriesType.SeventyFiveToOneHundred]: chartColorPaletteArray[1],
};

const seriesName: Record<ESeriesType, string> = {
  [ESeriesType.Total]: "Total GPUs",
  [ESeriesType.ZeroToTwentyFive]: "GPUs with 0–25% utilization",
  [ESeriesType.TwentyFiveToFifty]: "GPUs with 25–50% utilization",
  [ESeriesType.FiftyToSeventyFive]: "GPUs with 50–75% utilization",
  [ESeriesType.SeventyFiveToOneHundred]: "GPUs with 75–100% utilization",
};

const TEN_MINUTES_IN_MS = 10 * 60 * 1000;

export const GPU_DISTRIBUTION_RENDER_ORDER: ESeriesType[] = [
  ESeriesType.ZeroToTwentyFive,
  ESeriesType.TwentyFiveToFifty,
  ESeriesType.FiftyToSeventyFive,
  ESeriesType.SeventyFiveToOneHundred,
];

export interface IMeasurementByRange {
  type: ESeriesType;
  values: GroupedMeasurementResponseValuesInner[];
}

export default defineComponent({
  name: "gpu-utilization-distribution-time-range-widget",
  components: { ChartWidgetWrapper, VChart, ChartZoomControls, ChartAreaZoomControls },
  emits: ["register", "unregister", "on-zoom", "on-reset-zoom"],
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    filterByDates: {
      type: Object as PropType<IRangeDates>,
      required: true,
    },
    filterBy: {
      type: Array as PropType<IFilterModel[]>,
      required: false,
    },
    confineTooltip: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      chartId: GPU_UTILIZATION_DISTRIBUTION_TIME_RANGE_WIDGET_CHART_ID as string,
      chartOptions: null as null | EChartsOption,
      error: false as boolean,
      loading: false as boolean,
      isInitialLoader: true as boolean,
      timeframes: [] as Array<string>,
      measurementMap: {} as Record<string, Array<GroupedMeasurementResponseValuesInner>>,
      wrapperOptions: {
        title: "GPU utilization distribution",
      } as IWidgetWrapperOptions,
    };
  },
  created() {
    this.loadChartOptions();
    this.loadChartData();
    this.startRefreshLoadData();
  },
  computed: {
    nodePoolName(): string | undefined {
      const nodePoolName = this.filterBy?.find((filter: IFilterModel) => filter.field === "nodepool");
      if (nodePoolName) {
        return nodePoolName.term;
      }
      return undefined;
    },
  },
  methods: {
    startRefreshLoadData(): void {
      intervalUtil.startInterval(EIntervalLabels.GpuUtilizationDistributionWidgetTimeRangeWidget, () => {
        this.loadChartData(true);
      });
    },
    stopRefreshLoadData(): void {
      intervalUtil.stopInterval(EIntervalLabels.GpuUtilizationDistributionWidgetTimeRangeWidget);
    },
    loadChartOptions(): void {
      this.chartOptions = {
        color: [
          seriesColor[ESeriesType.Total],
          seriesColor[ESeriesType.SeventyFiveToOneHundred],
          seriesColor[ESeriesType.FiftyToSeventyFive],
          seriesColor[ESeriesType.TwentyFiveToFifty],
          seriesColor[ESeriesType.ZeroToTwentyFive],
        ],
        textStyle: { fontFamily: "Roboto, sans-serif" },
        grid: { left: 46, right: 20, top: 30, bottom: 100 },
        // Two centered legend rows (per user feedback: visually centered on chart canvas).
        legend: [
          {
            // Row 1: Total GPUs + 75-100% + 50-75% — centered
            show: true,
            orient: "horizontal",
            left: "center",
            bottom: 31,
            itemWidth: 12,
            itemHeight: 12,
            itemGap: 30,
            textStyle: { ...CHART_LEGEND_TEXT_STYLE_DEFAULTS },
            data: [
              { name: seriesName[ESeriesType.Total], icon: LEGEND_DASHED_LINE_ICON },
              { name: seriesName[ESeriesType.SeventyFiveToOneHundred], icon: "circle" },
              { name: seriesName[ESeriesType.FiftyToSeventyFive], icon: "circle" },
            ],
          },
          {
            // Row 2: 25-50% + 0-25% — centered
            show: true,
            orient: "horizontal",
            left: "center",
            bottom: 16,
            itemWidth: 12,
            itemHeight: 12,
            itemGap: 30,
            textStyle: { ...CHART_LEGEND_TEXT_STYLE_DEFAULTS },
            data: [
              { name: seriesName[ESeriesType.TwentyFiveToFifty], icon: "circle" },
              { name: seriesName[ESeriesType.ZeroToTwentyFive], icon: "circle" },
            ],
          },
        ],
        xAxis: {
          type: "time",
          interval: TEN_MINUTES_IN_MS,
          axisLabel: {
            formatter: "{dd}/{MM}/{yyyy}\n{HH}:{mm}",
            hideOverlap: true,
            ...CHART_AXIS_LABEL_DEFAULTS,
          },
          splitLine: { show: false },
        },
        yAxis: {
          type: "value",
          min: 0,
          max: 100,
          interval: 25,
          axisLabel: {
            formatter: (value: number) => value + "%",
            ...CHART_AXIS_LABEL_DEFAULTS,
          },
          splitLine: { lineStyle: { type: "solid", color: "#e6e6e6" } },
        },
        tooltip: {
          trigger: "axis",
          ...STANDARD_TOOLTIP_STYLE,
          ...(this.confineTooltip ? { appendToBody: false, confine: true } : {}),
          axisPointer: {
            ...chartUtil.getConnectedChartAxisPointer(),
          },
          formatter: (params: unknown) => {
            const paramsArr = params as Array<{
              seriesId?: string;
              seriesName: string;
              value: [number, number | null];
              color: string;
            }>;
            if (!Array.isArray(paramsArr) || paramsArr.length === 0) return "";
            const visibleItems = paramsArr.filter((item) => !chartUtil.isConnectedTooltipAnchorSeries(item));
            const points = visibleItems.map((p) => ({
              name: p.seriesName,
              y: p.value[1] !== null ? roundToDecimal(p.value[1]) : p.value[1],
              color: p.color,
              symbol: p.seriesName === seriesName.Total ? "" : "%",
            }));
            return chartUtil.formatSharedTooltip(points, (visibleItems[0] || paramsArr[0]).value[0]);
          },
        },
        dataZoom: [{ ...INSIDE_DATAZOOM_CONFIG }],
        toolbox: chartUtil.getDashboardChartToolbox({ showZoomToolbox: true }),
        series: [],
      };
    },
    async loadChartData(refresh = false): Promise<void> {
      const endDate = this.getEndDate();
      this.timeframes = chartUtil.getTimeFrames(this.filterByDates.dateStart, endDate);
      try {
        this.loading = true;
        const metricsResponse: MetricsResponse = await this.getMetric();
        this.measurementMap = metricUtil.mapMeasurementsToMap(metricsResponse);
        this.setTooltipAnchorSeries(refresh);
        this.setSeriesData(refresh);
        this.setDistributionSeriesData(refresh);
      } catch (error: unknown) {
        this.error = true;
        console.error(error);
      } finally {
        this.loading = false;
        this.isInitialLoader = false;
      }

      this.$nextTick(() => {
        if (!this.error) {
          this.$emit("register");
        }
      });
    },
    getEndDate(): Date {
      return this.filterByDates.dateEnd && dateUtil.isDateIsToday(this.filterByDates.dateEnd)
        ? new Date()
        : this.filterByDates.dateEnd;
    },
    async getMetric(): Promise<MetricsResponse> {
      const samples = this.timeframes.length;
      const endDate = this.getEndDate();
      const metricsType: MetricsType[] = [MetricsType.GpuUtilizationDistribution, MetricsType.TotalGpu];
      const start = dateUtil.convertDateToISO(this.filterByDates.dateStart);
      const end = dateUtil.convertDateToISO(endDate);
      if (this.nodePoolName) {
        return await clusterService.getNodepoolMetrics(
          this.clusterId,
          this.nodePoolName,
          start,
          end,
          metricsType,
          samples,
        );
      } else {
        return await clusterService.getClusterMetrics(this.clusterId, start, end, metricsType, samples);
      }
    },
    setSeriesData(refresh = false): void {
      const measurements: Array<GroupedMeasurementResponseValuesInner> | undefined =
        this.measurementMap[MetricsType.TotalGpu];
      const seriesData: Array<TMeasurementsTimeAndValue> = this.getMeasurements(this.timeframes, measurements);
      const series: SeriesOption = {
        type: "line",
        data: seriesData,
        name: this.getSeriesName(ESeriesType.Total),
        itemStyle: { color: this.getSeriesColor(ESeriesType.Total) },
        lineStyle: chartUtil.getDottedLineStyle(),
        showSymbol: false,
        ...chartUtil.getConnectedChartHoverMarker(this.getSeriesColor(ESeriesType.Total)),
      };
      this.updateChartSeries(series, seriesData, 0, refresh);
    },
    setTooltipAnchorSeries(refresh: boolean): void {
      if (!this.chartOptions) return;

      const anchor = chartUtil.getConnectedTooltipAnchorSeries(this.timeframes);
      const series = Array.isArray(this.chartOptions.series) ? (this.chartOptions.series as SeriesOption[]) : [];
      if (refresh && series[0]?.id === anchor.id) {
        series[0].data = anchor.data;
      } else {
        this.chartOptions.series = [anchor];
      }
    },
    setDistributionSeriesData(refresh = false): void {
      const measurements: Array<GroupedMeasurementResponseValuesInner> | undefined =
        this.measurementMap[MetricsType.GpuUtilizationDistribution];

      // Highcharts reverses normal stacks by default. ECharts stacks in series
      // order, so render low-to-high buckets to preserve the same visible bands.
      const measurementsByRange: IMeasurementByRange[] = GPU_DISTRIBUTION_RENDER_ORDER.map((type) => ({
        type,
        values: [],
      }));

      measurements.forEach((item: GroupedMeasurementResponseValuesInner) => {
        const parsedValue: Record<ESeriesType, string>[] = JSON.parse(item.value);
        measurementsByRange.forEach((range) => {
          range.values.push({
            value: parsedValue.find((value: RangeData) => value[range.type])?.[range.type] || "0",
            timestamp: item.timestamp,
          });
        });
      });

      measurementsByRange.forEach((range: IMeasurementByRange, seriesIndex) => {
        const seriesData: Array<TMeasurementsTimeAndValue> = this.getMeasurements(this.timeframes, range.values);
        const series: SeriesOption = {
          type: "line",
          data: seriesData,
          name: this.getSeriesName(range.type),
          itemStyle: { color: this.getSeriesColor(range.type) },
          lineStyle: { width: 1 },
          showSymbol: false,
          ...chartUtil.getConnectedChartHoverMarker(this.getSeriesColor(range.type)),
          stack: "distribution",
          areaStyle: { opacity: 0.2 },
        };
        this.updateChartSeries(series, seriesData, seriesIndex + 1, refresh);
      });
    },
    updateChartSeries(
      series: SeriesOption,
      seriesData: Array<TMeasurementsTimeAndValue>,
      seriesIndex: number,
      refresh: boolean,
    ): void {
      if (this.chartOptions) {
        if (!this.chartOptions.series) {
          this.chartOptions.series = [];
        }
        const seriesArr = this.chartOptions.series as SeriesOption[];
        if (refresh) {
          const existing = seriesArr[seriesIndex + 1] as Record<string, unknown>;
          if (existing) {
            existing.data = seriesData;
          }
        } else {
          seriesArr.push(series);
        }
        this.chartOptions = { ...this.chartOptions, series: [...seriesArr] };
      }
    },
    getSeriesName(seriesType: ESeriesType): string {
      return seriesName[seriesType];
    },
    getSeriesColor(seriesType: ESeriesType): (typeof seriesColor)[ESeriesType] {
      return seriesColor[seriesType];
    },
    getMeasurements(
      timeframes: string[],
      measurements: Array<GroupedMeasurementResponseValuesInner>,
    ): Array<TMeasurementsTimeAndValue> {
      return chartUtil.arrangeMeasurementsByTime(timeframes, measurements, { useTimeFrameTimestamps: true });
    },
    handleZoomToggle(active: boolean): void {
      const chart = this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
      if (!chart?.chart) return;
      chart.chart.dispatchAction({
        type: "takeGlobalCursor",
        key: "dataZoomSelect",
        dataZoomSelectActive: active,
      });
    },
    handleAreaZoomReset(): void {
      const chart = this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
      if (!chart?.chart) return;
      chartUtil.resetChartDataZoomInstance(chart.chart);
    },
    handleZoomControls(payload: { start: number; end: number }): void {
      const chartRef = this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
      if (!chartRef) return;
      chartRef.chart?.dispatchAction({ type: "dataZoom", start: payload.start, end: payload.end });
    },
    handleDataZoom(event: Parameters<typeof chartUtil.getDataZoomRange>[0]): void {
      const range = chartUtil.getDataZoomRange(event, this.timeframes);

      if (range) {
        this.$emit("on-zoom", { ...range, chartId: this.chartId });
      }
    },
    handleChartAreaClick(event: { target?: unknown }): void {
      if (event.target) return;
      this.handleResetZoom();
    },
    handleResetZoom(): void {
      const chartRef = this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
      chartRef?.chart?.dispatchAction({ type: "dataZoom", start: 0, end: 100 }, { silent: true });

      const zoomControls = this.$refs.zoomControls as InstanceType<typeof ChartZoomControls> | undefined;
      zoomControls?.resetZoom();
      this.$emit("on-reset-zoom", this.chartId);
    },
    async onPropsChanged(): Promise<void> {
      this.$emit("unregister");
      this.isInitialLoader = true;
      await this.loadChartData(true);
    },
  },
  watch: {
    filterByDates: {
      handler(): void {
        this.onPropsChanged();
      },
      deep: true,
    },
  },
  beforeUnmount() {
    this.$emit("unregister");
    this.stopRefreshLoadData();
  },
});
</script>

<style scoped lang="scss">
.time-range-chart-container {
  position: relative;
  width: 100%;
}

.node-charts {
  height: 100%;
  display: flex;
  flex-direction: column;

  .node-charts-header {
    border-bottom: solid 1px $black-20;
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 56px;
    padding: 0 16px;

    .metric-title {
      font-weight: 500;
    }
  }
}
</style>
