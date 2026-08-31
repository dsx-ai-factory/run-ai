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
  GPU_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID,
  INSIDE_DATAZOOM_CONFIG,
  chartColorPaletteArray,
  type IWidgetWrapperOptions,
  type IZoomEvent,
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

const seriesColor: Record<string, string> = {
  [MetricsType.TotalGpu]: chartColorPaletteArray[0],
  [MetricsType.AllocatedGpu]: chartColorPaletteArray[1],
  [MetricsType.UnallocatedGpu]: chartColorPaletteArray[2],
};

const seriesName: Record<string, string> = {
  [MetricsType.TotalGpu]: "Total GPUs",
  [MetricsType.AllocatedGpu]: "Allocated GPUs",
  [MetricsType.UnallocatedGpu]: "Unallocated GPUs",
};
const metricsTypes: MetricsType[] = [MetricsType.TotalGpu, MetricsType.AllocatedGpu, MetricsType.UnallocatedGpu];
export const GPU_ALLOCATION_RENDER_ORDER: MetricsType[] = [
  MetricsType.TotalGpu,
  MetricsType.AllocatedGpu,
  MetricsType.UnallocatedGpu,
];

export function getGpuAllocationSeriesLayout(metricType: MetricsType): { stack?: string; z: number } {
  return metricType === MetricsType.TotalGpu ? { z: 1 } : { stack: "allocation", z: 2 };
}
const TEN_MINUTES_IN_MS = 10 * 60 * 1000;
export enum EEmitName {
  Register = "register",
  Unregister = "unregister",
  OnZoom = "on-zoom",
  OnResetZoom = "on-reset-zoom",
}
export default defineComponent({
  name: "gpu-allocation-time-range-widget",
  components: { ChartWidgetWrapper, VChart, ChartZoomControls, ChartAreaZoomControls },
  emits: {
    [EEmitName.Register]: null,
    [EEmitName.Unregister]: null,
    [EEmitName.OnZoom]: (event: IZoomEvent) => event,
    [EEmitName.OnResetZoom]: (chartId: string) => chartId,
  },
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
      chartId: GPU_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID as string,
      chartOptions: null as null | EChartsOption,
      error: false as boolean,
      loading: false as boolean,
      isInitialLoader: true as boolean,
      timeframes: [] as Array<string>,
      measurementMap: {} as Record<string, Array<GroupedMeasurementResponseValuesInner>>,
      wrapperOptions: {
        title: "GPU allocation",
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
      return this.filterBy?.find((filter: IFilterModel) => filter.field === "nodepool")?.term;
    },
  },
  methods: {
    startRefreshLoadData(): void {
      intervalUtil.startInterval(EIntervalLabels.GpuAllocationTimeRangeWidget, () => {
        this.loadChartData(true);
      });
    },
    stopRefreshLoadData(): void {
      intervalUtil.stopInterval(EIntervalLabels.GpuAllocationTimeRangeWidget);
    },
    loadChartOptions(): void {
      this.chartOptions = {
        color: metricsTypes.map((metricType) => seriesColor[metricType]),
        textStyle: { fontFamily: "Roboto, sans-serif" },
        grid: { left: 24, right: 20, top: 30, bottom: 86 },
        legend: {
          show: true,
          left: "center",
          bottom: 16,
          itemWidth: 12,
          itemHeight: 12,
          itemGap: 24,
          textStyle: { ...CHART_LEGEND_TEXT_STYLE_DEFAULTS },
          data: metricsTypes.map((metricType) => ({
            name: seriesName[metricType],
            icon: "circle",
          })),
        },
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
          name: "",
          nameLocation: "middle",
          nameRotate: 90,
          nameGap: 40,
          boundaryGap: [0, "50%"],
          splitNumber: 3,
          splitLine: { lineStyle: { type: "solid", color: "#e6e6e6" } },
          axisLabel: { ...CHART_AXIS_LABEL_DEFAULTS, formatter: chartUtil.formatCompactAxisValue },
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
            const points = [...visibleItems]
              .sort((a, b) => this.getSeriesSortIndex(a.seriesName) - this.getSeriesSortIndex(b.seriesName))
              .map((p) => ({
                name: p.seriesName,
                y: p.value[1] !== null ? roundToDecimal(p.value[1]) : p.value[1],
                color: p.color,
              }));
            return chartUtil.formatSharedTooltip(points, (visibleItems[0] || paramsArr[0]).value[0]);
          },
        },
        dataZoom: [{ ...INSIDE_DATAZOOM_CONFIG }],
        toolbox: chartUtil.getDashboardChartToolbox({ showZoomToolbox: true }),
        series: [],
      };
    },
    getEndDate(): Date {
      return this.filterByDates.dateEnd && dateUtil.isDateIsToday(this.filterByDates.dateEnd)
        ? new Date()
        : this.filterByDates.dateEnd;
    },
    async loadChartData(refresh = false): Promise<void> {
      const endDate = this.getEndDate();
      this.timeframes = chartUtil.getTimeFrames(this.filterByDates.dateStart, endDate);
      try {
        this.loading = true;
        const metricsResponse: MetricsResponse = await this.getMetric();
        this.measurementMap = metricUtil.mapMeasurementsToMap(metricsResponse);
        this.setTooltipAnchorSeries(refresh);
        GPU_ALLOCATION_RENDER_ORDER.forEach((metricType: MetricsType, index) =>
          this.setSeriesData(metricType, index, refresh),
        );
      } catch (error: unknown) {
        this.error = true;
        console.error(error);
      } finally {
        this.loading = false;
        this.isInitialLoader = false;
      }
      this.$nextTick(() => {
        if (!this.error) {
          this.$emit(EEmitName.Register);
        }
      });
    },
    async getMetric(): Promise<MetricsResponse> {
      const samples = this.timeframes.length;
      const endDate = this.getEndDate();

      const start = dateUtil.convertDateToISO(this.filterByDates.dateStart);
      const end = dateUtil.convertDateToISO(endDate);
      if (this.nodePoolName) {
        return await clusterService.getNodepoolMetrics(
          this.clusterId,
          this.nodePoolName,
          start,
          end,
          metricsTypes,
          samples,
        );
      } else {
        return await clusterService.getClusterMetrics(this.clusterId, start, end, metricsTypes, samples);
      }
    },
    setSeriesData(metricType: MetricsType, seriesIndex: number, refresh = false): void {
      const measurements: Array<GroupedMeasurementResponseValuesInner> | undefined = this.measurementMap[metricType];
      const seriesData: Array<TMeasurementsTimeAndValue> = this.getMeasurements(this.timeframes, measurements);
      const isTotal = metricType === MetricsType.TotalGpu;
      const series: SeriesOption = {
        type: "line",
        data: seriesData,
        name: this.getSeriesName(metricType),
        itemStyle: { color: this.getSeriesColor(metricType) },
        lineStyle: isTotal ? chartUtil.getDottedLineStyle() : { width: 1, type: "solid" },
        showSymbol: false,
        ...chartUtil.getConnectedChartHoverMarker(this.getSeriesColor(metricType)),
        ...getGpuAllocationSeriesLayout(metricType),
        areaStyle: { opacity: 0.2 },
      };
      this.updateChartSeries(series, seriesData, seriesIndex, refresh);
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
    getSeriesName(metricType: MetricsType): string {
      return seriesName[metricType];
    },
    getSeriesSortIndex(series: string): number {
      const index = metricsTypes.findIndex((metricType) => seriesName[metricType] === series);
      return index === -1 ? metricsTypes.length : index;
    },
    getSeriesColor(metricType: MetricsType): string {
      return seriesColor[metricType];
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
        this.$emit(EEmitName.OnZoom, { ...range, chartId: this.chartId });
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
      this.$emit(EEmitName.OnResetZoom, this.chartId);
    },
    async onPropsChanged(): Promise<void> {
      this.$emit(EEmitName.Unregister);
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
    this.$emit(EEmitName.Unregister);
    this.stopRefreshLoadData();
  },
});
</script>

<style scoped lang="scss">
.time-range-chart-container {
  position: relative;
  width: 100%;
}
</style>
