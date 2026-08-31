<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <chart-widget-wrapper
    :options="wrapperOptions"
    body-no-padding
    :loading="initialLoading"
    :error="error"
    hide-border
    flat
    hide-actions
    hide-separator
  >
    <div class="metric-chart-shell">
      <chart-area-zoom-controls
        :external-active="zoomActive"
        :vertical-offset="28"
        @zoom-toggle="handleZoomToggle"
        @reset="handleAreaZoomReset"
      />
      <v-chart
        v-if="chartOptions"
        ref="chartRef"
        :id="chartId"
        :option="chartOptions"
        autoresize
        style="height: 250px; width: 100%"
        @datazoom="handleDataZoom"
        @restore="handleResetZoom"
      />
    </div>
  </chart-widget-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import VChart from "vue-echarts";
import "@/plugins/echarts";
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { ChartAreaZoomControls } from "@/components/common/charts/chart-area-zoom-controls";

// models
import {
  type IWidgetWrapperOptions,
  type ITooltipPoint,
  type TMeasurementsTimeAndValue,
  type IZoomEvent,
  EYAxisLabelFormat,
  LEGEND_DASHED_LINE_ICON,
  LEGEND_RECT_ITEM_STYLE,
  chartColorPaletteArray,
} from "@/models/chart.model";
import {
  PodMetricType,
  type Pod as WorkloadServicePod,
  type MetricsResponse,
  type MeasurementResponse,
  type MeasurementResponseValuesInner,
  type Workload,
} from "@/swagger-models/workloads-client";

// ECharts
import type { EChartsOption } from "echarts";
import type { LineSeriesOption } from "echarts/charts";

// services
import { podService } from "@/services/control-plane/pod.service/pod.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// utils
import { chartUtil } from "@/utils/chart.util";
import type { IEChartsXAxisOptions, IEChartsYAxisOptions } from "@/utils/chart.util/chart.util";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import { metricUtil } from "@/utils/metric.util";
import { convertToBytes, parseSizeString } from "@/utils/format.util";
import { intervalUtil } from "@/utils/interval.util";
import { formatK8sResourceMemory } from "@/utils/format.util/format.util";

// store
import { EIntervalLabels } from "@/models/interval.model";

export const CPU_MEMORY_USAGE_PER_GPU_WIDGET_CHART_ID = "cpu-memory-usage-per-gpu-chart";

export enum ELinesInCharts {
  Request = "request",
  Usage = "usage",
  Limit = "limit",
}

export default defineComponent({
  components: { VChart, ChartWidgetWrapper, ChartAreaZoomControls },
  emits: ["register", "unregister", "on-zoom", "on-reset-zoom", "on-zoom-toggle"],
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    workloadId: {
      type: String as PropType<string>,
      required: true,
    },
    podId: {
      type: String as PropType<string>,
      required: true,
    },
    filterByDates: {
      type: Object as PropType<IRangeDates>,
      required: true,
    },
    silentDateReload: {
      type: Boolean,
      default: false,
    },
    zoomActive: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      chartOptions: null as null | EChartsOption,
      error: false as boolean,
      initialLoading: false as boolean,
      timeframes: [] as Array<string>,
      wrapperOptions: {
        title: "CPU memory usage",
      } as IWidgetWrapperOptions,
      measurementMap: {} as Record<string, Array<MeasurementResponseValuesInner>>,
      currWorkload: null as null | Workload,
      chartId: CPU_MEMORY_USAGE_PER_GPU_WIDGET_CHART_ID as string,
      requestedCpuMemory: null as null | number,
      limitCpuMemory: null as null | number,
    };
  },
  created() {
    this.loadChartOptions();
    this.onPropsChanged();
  },
  computed: {
    shouldShowRequest(): boolean {
      return this.requestedCpuMemory !== null;
    },
    shouldShowLimit(): boolean {
      return this.limitCpuMemory !== null;
    },
  },
  methods: {
    loadChartOptions(): void {
      this.chartOptions = chartUtil.getBasicWidgetChartOptions({
        type: "line",
        height: 250,
        showSharedCrosshair: true,
        sharedTooltip: true,
        yAxisFormatFunction: formatK8sResourceMemory,
        grid: { top: 5, right: 20, bottom: 50, left: 34, containLabel: true },
      });

      this.addTooltipOptions();
      this.addXAxisOptions();
      this.addZoomOptions();
    },
    addTooltipOptions(): void {
      if (!this.chartOptions) return;

      this.chartOptions.tooltip = {
        ...this.chartOptions.tooltip,
        formatter: (params: unknown) => {
          const items = params as Array<{
            seriesId?: string;
            seriesName: string;
            value: [number, number | null];
            color: string;
          }>;
          if (!Array.isArray(items) || !items.length) return "";
          const visibleItems = items.filter((item) => !chartUtil.isConnectedTooltipAnchorSeries(item));
          const points: Array<ITooltipPoint> = visibleItems.map((item) => {
            return {
              name: item.seriesName,
              y: item.value[1] != null ? formatK8sResourceMemory(item.value[1]) : undefined,
              color: item.color,
            };
          });
          return chartUtil.formatSharedTooltip(points, (visibleItems[0] || items[0]).value[0]) as string;
        },
      };
    },
    addXAxisOptions(): void {
      if (!this.chartOptions) return;

      this.chartOptions.xAxis = {
        ...chartUtil.getDatetimeOptions(this.chartOptions.xAxis as IEChartsXAxisOptions),
        axisLine: {
          lineStyle: {
            color: "#CCD6EB",
          },
        },
        axisLabel: chartUtil.getDatetimeAxisLabelOptions((value: number) => this.formatDatetimeAxisLabel(value)),
      };
    },
    addZoomOptions(): void {
      if (!this.chartOptions) return;

      this.chartOptions = chartUtil.addZoomOptions(
        this.chartOptions as EChartsOption,
        this.chartId,
        (event: IZoomEvent) => this.$emit("on-zoom", event),
        (chartId: string) => this.$emit("on-reset-zoom", chartId),
      );
    },
    updateLegendOptions(): void {
      if (!this.chartOptions) return;

      const series = ((this.chartOptions.series ?? []) as LineSeriesOption[]).filter(
        (item) => !chartUtil.isConnectedTooltipAnchorSeries({ seriesId: item.id }),
      );
      this.chartOptions.legend = {
        ...(this.chartOptions.legend as Record<string, unknown>),
        bottom: 0,
        ...LEGEND_RECT_ITEM_STYLE,
        itemGap: 20,
        textStyle: {
          fontSize: 11,
          fontWeight: 700,
          color: "#333333",
        },
        data: series.map((seriesOption: LineSeriesOption) => {
          const isDotted = chartUtil.isDottedLineStyle(seriesOption.lineStyle);
          return {
            name: String(seriesOption.name),
            icon: isDotted ? LEGEND_DASHED_LINE_ICON : chartUtil.getDashboardLineOnlyLegendIcon(),
          };
        }),
      };
    },
    async loadChartData(): Promise<void> {
      if (this.chartOptions === null) return;

      try {
        if (this.initialLoading) {
          await this.loadPodData();
          this.currWorkload = await workloadService.getWorkloadById(this.workloadId);
          this.addSeriesDetails();
        }
        const updatedDates: IRangeDates = dateUtil.getUpdatedTime(this.filterByDates);
        this.timeframes = chartUtil.getTimeFrames(updatedDates.dateStart, updatedDates.dateEnd);
        this.setTooltipAnchorSeries();
        const samples = this.timeframes.length;

        const metricType: Array<PodMetricType> = [PodMetricType.CpuMemoryUsageBytes];
        const metricsResponse: MetricsResponse = await podService.getPodMetrics(
          this.workloadId,
          this.podId,
          dateUtil.convertDateToISO(updatedDates.dateStart),
          dateUtil.convertDateToISO(updatedDates.dateEnd),
          metricType,
          samples,
        );
        this.measurementMap = metricUtil.mapMeasurementsToMap(metricsResponse);
        this.setSeriesData();
        this.updateYAxisLabels(metricsResponse.measurements);
        this.updateLegendOptions();
        this.error = false;
      } catch (error: unknown) {
        this.error = true;
        console.error(error);
      } finally {
        this.initialLoading = false;
      }

      this.$nextTick(() => {
        if (!this.error) {
          this.$emit("register");
        }
      });
    },
    addSeriesDetails(): void {
      if (!this.chartOptions || !this.currWorkload) return;

      const series: LineSeriesOption[] = [];

      if (this.shouldShowRequest) {
        series.push({
          id: ELinesInCharts.Request,
          name: "CPU memory request",
          type: "line",
          lineStyle: chartUtil.getDottedLineStyle(),
          showSymbol: false,
          symbol: "none",
          connectNulls: true,
        });
      }

      series.push({
        id: ELinesInCharts.Usage,
        name: "CPU memory usage",
        type: "line",
        showSymbol: false,
        symbol: "none",
        connectNulls: true,
      });

      if (this.shouldShowLimit) {
        series.push({
          id: ELinesInCharts.Limit,
          name: "CPU memory limit",
          type: "line",
          lineStyle: chartUtil.getDottedLineStyle(),
          showSymbol: false,
          symbol: "none",
          connectNulls: true,
        });
      }

      this.chartOptions.series = series.map((seriesOption, index) => ({
        ...seriesOption,
        ...chartUtil.getConnectedChartHoverMarker(chartColorPaletteArray[index % chartColorPaletteArray.length]),
      }));
    },
    async loadPodData(): Promise<void> {
      const pods = await workloadService.getWorkloadPods(this.workloadId);
      const currentPod: WorkloadServicePod | undefined = pods.find((pod: WorkloadServicePod) => pod.id === this.podId);
      if (currentPod?.requestedResources?.cpuMemory) {
        this.requestedCpuMemory = currentPod.requestedResources.cpuMemory.request
          ? this.parseStringSizeToBytes(currentPod.requestedResources.cpuMemory.request)
          : null;
        this.limitCpuMemory = currentPod.requestedResources.cpuMemory.limit
          ? this.parseStringSizeToBytes(currentPod.requestedResources.cpuMemory.limit)
          : null;
      }
    },
    setSeriesData(): void {
      if (this.chartOptions === null) return;

      const cpuMemoryUsageMeasurements = this.measurementMap[PodMetricType.CpuMemoryUsageBytes];
      if (!cpuMemoryUsageMeasurements) {
        return;
      }

      const cpuMemoryUsageData: Array<TMeasurementsTimeAndValue> = this.getMeasurements(
        this.timeframes,
        cpuMemoryUsageMeasurements,
      );

      const updatedTimeframes: Array<number> = cpuMemoryUsageData.map(
        (cpuUsage: TMeasurementsTimeAndValue) => cpuUsage[0],
      );

      this.addRequest(updatedTimeframes);
      this.addUsage(cpuMemoryUsageData);
      this.addLimit(updatedTimeframes);
    },
    setTooltipAnchorSeries(): void {
      if (!this.chartOptions) return;
      const series = Array.isArray(this.chartOptions.series) ? (this.chartOptions.series as LineSeriesOption[]) : [];
      const visibleSeries = series.filter((item) => !chartUtil.isConnectedTooltipAnchorSeries({ seriesId: item.id }));
      this.chartOptions.series = [chartUtil.getConnectedTooltipAnchorSeries(this.timeframes), ...visibleSeries];
    },
    getMeasurements(
      timeframes: string[],
      measurements: Array<MeasurementResponseValuesInner>,
    ): Array<TMeasurementsTimeAndValue> {
      return chartUtil.arrangeMeasurementsByTime(timeframes, measurements, { useTimeFrameTimestamps: true });
    },
    parseStringSizeToBytes(sizeString: string): number | null {
      let memoryFormatted = formatK8sResourceMemory(sizeString);
      if (memoryFormatted?.includes(" ")) {
        memoryFormatted = memoryFormatted.replace(" ", "");
      }

      const sizeAndString: [number, string] | null = parseSizeString(memoryFormatted);
      if (!sizeAndString) return null;
      return convertToBytes(sizeAndString[0], sizeAndString[1]);
    },
    insertConstantValueIntoTimeframe(
      constantValue: number,
      timeframes: Array<number>,
    ): Array<TMeasurementsTimeAndValue> {
      const createdAt: number | undefined = this.currWorkload
        ? new Date(this.currWorkload.createdAt).getTime()
        : undefined;
      return metricUtil.insertConstantValueIntoTimeframe(constantValue, timeframes, createdAt);
    },
    addUsage(cpuMemoryUsageData: Array<TMeasurementsTimeAndValue>): void {
      const seriesConfig = this.getSeriesConfig(ELinesInCharts.Usage);
      if (!seriesConfig) return;
      seriesConfig.data = cpuMemoryUsageData;
    },
    addRequest(timeframe: Array<number>): void {
      const seriesConfig = this.getSeriesConfig(ELinesInCharts.Request);
      if (!seriesConfig || this.requestedCpuMemory === null) return;

      const cpuMemoryRequestData: Array<TMeasurementsTimeAndValue> = this.insertConstantValueIntoTimeframe(
        this.requestedCpuMemory,
        timeframe,
      );
      seriesConfig.data = cpuMemoryRequestData;
    },
    addLimit(timeframe: Array<number>): void {
      const seriesConfig = this.getSeriesConfig(ELinesInCharts.Limit);
      if (!seriesConfig || this.limitCpuMemory === null) return;

      const cpuMemoryLimitData: Array<TMeasurementsTimeAndValue> = this.insertConstantValueIntoTimeframe(
        this.limitCpuMemory,
        timeframe,
      );
      seriesConfig.data = cpuMemoryLimitData;
    },
    getSeriesConfig(seriesId: ELinesInCharts): LineSeriesOption | undefined {
      const series = (this.chartOptions?.series ?? []) as LineSeriesOption[];
      return series.find((s: LineSeriesOption) => s.id === seriesId);
    },
    updateYAxisLabels(measurements: MeasurementResponse[]): void {
      if (!this.chartOptions) return;

      const yAxis = this.chartOptions.yAxis as IEChartsYAxisOptions;
      const measurementMax = this.getMaxMeasurementValue(measurements);
      const maxValue = Math.max(measurementMax ?? 0, this.requestedCpuMemory ?? 0, this.limitCpuMemory ?? 0);

      if (maxValue <= 0) {
        if (measurementMax === null) {
          this.chartOptions.yAxis = chartUtil.updateYAxisLabelsDefaultsIfNeeded(measurements, yAxis, {
            type: EYAxisLabelFormat.Memory,
          });
          return;
        }

        this.chartOptions.yAxis = {
          ...yAxis,
          ...chartUtil.getLineChartYAxisConfig(maxValue),
          axisLabel: { formatter: (value: number) => formatK8sResourceMemory(value) },
        };
        return;
      }

      const interval = this.getNiceInterval(maxValue / 2);
      this.chartOptions.yAxis = {
        ...yAxis,
        min: 0,
        max: this.roundAxisValue(Math.ceil(maxValue / interval) * interval),
        interval,
        axisLabel: { formatter: (value: number) => formatK8sResourceMemory(value) },
      };
    },
    getMaxMeasurementValue(measurements: MeasurementResponse[]): number | null {
      const values = measurements.flatMap((measurement: MeasurementResponse) =>
        (measurement.values ?? []).map((point: MeasurementResponseValuesInner) => Number(point.value)),
      );
      const finiteValues = values.filter((value: number) => Number.isFinite(value));

      if (!finiteValues.length) return null;

      return Math.max(...finiteValues);
    },
    getNiceInterval(rawInterval: number): number {
      if (!Number.isFinite(rawInterval) || rawInterval <= 0) return 1;

      const magnitude = 10 ** Math.floor(Math.log10(rawInterval));
      const normalized = rawInterval / magnitude;
      const multiplier = normalized <= 1 ? 1 : normalized <= 2 ? 2 : normalized <= 2.5 ? 2.5 : normalized <= 5 ? 5 : 10;
      return this.roundAxisValue(multiplier * magnitude);
    },
    roundAxisValue(value: number): number {
      return Number(value.toPrecision(12));
    },
    handleDataZoom(event: Parameters<typeof chartUtil.getDataZoomRange>[0]): void {
      const range = chartUtil.getDataZoomRange(event, this.timeframes);

      if (range) {
        this.$emit("on-zoom", { ...range, chartId: this.chartId } as IZoomEvent);
      }
    },
    handleZoomToggle(active: boolean): void {
      const chartRef = this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
      chartRef?.chart?.dispatchAction({
        type: "takeGlobalCursor",
        key: "dataZoomSelect",
        dataZoomSelectActive: active,
      });
      this.$emit("on-zoom-toggle", active);
    },
    handleAreaZoomReset(): void {
      this.handleResetZoom();
      this.$emit("on-zoom-toggle", false);
    },
    handleResetZoom(): void {
      chartUtil.resetChartDataZoom(this.chartId);
      this.$emit("on-reset-zoom", this.chartId);
    },
    formatDatetimeAxisLabel(value: number): string {
      return chartUtil.formatDatetimeAxisLabel(value);
    },
    async onPropsChanged(withLoading = true): Promise<void> {
      this.$emit("unregister");
      this.initialLoading = withLoading;
      intervalUtil.stopInterval(EIntervalLabels.CpuMemoryUsagePerGpuWidget);
      await this.loadChartData();
      intervalUtil.startInterval(EIntervalLabels.CpuMemoryUsagePerGpuWidget, this.loadChartData);
    },
  },
  watch: {
    workloadId(): void {
      this.onPropsChanged();
    },
    podId(): void {
      this.onPropsChanged();
    },
    filterByDates: {
      handler(): void {
        if (this.silentDateReload) {
          this.onPropsChanged(false);
        } else {
          this.onPropsChanged();
        }
      },
      deep: true,
    },
  },
  beforeUnmount() {
    this.$emit("unregister");
    intervalUtil.stopInterval(EIntervalLabels.CpuMemoryUsagePerGpuWidget);
  },
});
</script>

<style lang="scss" scoped>
.metric-chart-shell {
  position: relative;
}
</style>
