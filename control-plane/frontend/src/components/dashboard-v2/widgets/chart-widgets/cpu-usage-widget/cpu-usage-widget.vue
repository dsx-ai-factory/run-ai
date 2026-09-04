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
  LEGEND_DASHED_LINE_ICON,
  LEGEND_RECT_ITEM_STYLE,
} from "@/models/chart.model";
import {
  ElementMetricType,
  WorkloadMetricType,
  type Workload,
  type MetricsResponse,
  type MeasurementResponse,
  type MeasurementResponseValuesInner,
} from "@/swagger-models/workloads-client";

// ECharts
import type { EChartsOption } from "echarts";
import type { LineSeriesOption } from "echarts/charts";

// services
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// utils
import { chartUtil } from "@/utils/chart.util";
import type { IEChartsXAxisOptions, IEChartsYAxisOptions } from "@/utils/chart.util/chart.util";
import { metricUtil } from "@/utils/metric.util";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import { formatToFixedOrZero } from "@/utils/format.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";

export enum ELinesInCharts {
  Request = "request",
  Usage = "usage",
  Limit = "limit",
}

export const CPU_USAGE_WIDGET_CHART_ID = "cpu-usage-chart";

export default defineComponent({
  components: { VChart, ChartWidgetWrapper, ChartAreaZoomControls },
  emits: ["register", "unregister", "on-zoom", "on-reset-zoom", "on-zoom-toggle"],
  props: {
    workloadId: {
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
    selectedElementId: {
      type: String as PropType<string | null>,
      default: null,
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
      initialLoading: true as boolean,
      currWorkload: null as Workload | null,
      timeframes: [] as Array<string>,
      measurementMap: {} as Record<string, Array<MeasurementResponseValuesInner>>,
      wrapperOptions: {
        title: "CPU usage",
      } as IWidgetWrapperOptions,
      chartId: CPU_USAGE_WIDGET_CHART_ID as string,
    };
  },
  created() {
    this.loadChartOptions();
    this.onPropsChanged();
  },
  computed: {
    shouldShowRequest(): boolean {
      const request: number | null = this.currWorkload?.workloadRequestedResources?.cpu?.request || null;
      return request !== null;
    },
    shouldShowLimit(): boolean {
      const limit: number | null = this.currWorkload?.workloadRequestedResources?.cpu?.limit || null;
      return limit !== null;
    },
    metricType(): Array<WorkloadMetricType> {
      const metricType: Array<WorkloadMetricType> = [WorkloadMetricType.CpuUsageCores];

      if (this.shouldShowLimit) {
        metricType.push(WorkloadMetricType.CpuLimitCores);
      }
      if (this.shouldShowRequest) {
        metricType.push(WorkloadMetricType.CpuRequestCores);
      }

      return metricType;
    },
    elementMetricType(): Array<ElementMetricType> {
      const metricType: Array<ElementMetricType> = [ElementMetricType.CpuUsageCores];

      if (this.shouldShowLimit) {
        metricType.push(ElementMetricType.CpuLimitCores);
      }
      if (this.shouldShowRequest) {
        metricType.push(ElementMetricType.CpuRequestCores);
      }

      return metricType;
    },
  },
  methods: {
    loadChartOptions(): void {
      this.chartOptions = chartUtil.getBasicWidgetChartOptions({
        yAxisTitle: "Core(s)",
        type: "line",
        height: 250,
        showSharedCrosshair: true,
        sharedTooltip: true,
        // Reserve enough room for the rotated "Core(s)" y-axis title. The
        // shared chart default uses nameGap: 56, so a 34px grid margin clips it
        // in the workload drawer.
        grid: { top: 30, right: 10, bottom: 50, left: 56, containLabel: true },
      });

      this.addTooltipOptions();
      this.addXAxisOptions();
      this.addZoomOptions();
    },
    addTooltipOptions(): void {
      if (!this.chartOptions) return;

      this.chartOptions.tooltip = {
        ...this.chartOptions.tooltip,
        axisPointer: {
          ...chartUtil.getConnectedChartAxisPointer(),
        },
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
            const pointY = item.value[1] != null ? formatToFixedOrZero(item.value[1]) : undefined;
            return {
              name: item.seriesName,
              y: item.value[1] != null ? `${pointY} Cores` : undefined,
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
        axisLabel: chartUtil.getDatetimeAxisLabelOptions((value: number | string) =>
          this.formatDatetimeAxisLabel(value),
        ),
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
    async loadChartData(): Promise<void> {
      if (this.chartOptions === null) return;
      try {
        const updatedDates: IRangeDates = dateUtil.getUpdatedTime(this.filterByDates);
        this.timeframes = chartUtil.getTimeFrames(updatedDates.dateStart, updatedDates.dateEnd);
        const samples = this.timeframes.length;

        if (this.initialLoading) {
          this.currWorkload = await workloadService.getWorkloadById(this.workloadId);
          this.addSeriesDetails();
        }
        this.setTooltipAnchorSeries();

        const start = dateUtil.convertDateToISO(updatedDates.dateStart);
        const end = dateUtil.convertDateToISO(updatedDates.dateEnd);
        const metricsResponse: MetricsResponse = this.selectedElementId
          ? await workloadService.getWorkloadElementMetrics(
              this.workloadId,
              this.selectedElementId,
              this.elementMetricType,
              start,
              end,
              samples,
            )
          : await workloadService.getWorkloadMetrics(this.workloadId, start, end, this.metricType, samples);

        this.measurementMap = metricUtil.mapMeasurementsToMap(metricsResponse);

        this.setSeriesData();
        this.updateYAxisLabels(metricsResponse.measurements);
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
    setSeriesData(): void {
      this.addRequest();
      this.addUsage();
      this.addLimit();
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
    addSeriesDetails(): void {
      if (!this.chartOptions || !this.currWorkload) return;

      const series: LineSeriesOption[] = [];

      if (this.shouldShowRequest) {
        series.push({
          id: ELinesInCharts.Request,
          name: "CPU request",
          type: "line",
          lineStyle: chartUtil.getDottedLineStyle(),
          showSymbol: false,
          symbol: "circle",
        });
      }

      series.push({
        id: ELinesInCharts.Usage,
        name: "CPU usage",
        type: "line",
        showSymbol: false,
        symbol: "circle",
      });

      if (this.shouldShowLimit) {
        series.push({
          id: ELinesInCharts.Limit,
          name: "CPU limit",
          type: "line",
          lineStyle: chartUtil.getDottedLineStyle(),
          showSymbol: false,
          symbol: "circle",
        });
      }

      this.chartOptions.series = this.applyDashboardLineSeriesStyle(series);
      this.updateLegendOptions();
    },
    applyDashboardLineSeriesStyle(series: LineSeriesOption[]): LineSeriesOption[] {
      return series.map((seriesOption: LineSeriesOption, index: number) => ({
        ...seriesOption,
        ...chartUtil.getDashboardLineSeriesStyle(index, seriesOption.lineStyle),
      }));
    },
    updateLegendOptions(): void {
      if (!this.chartOptions) return;

      const series = ((this.chartOptions.series ?? []) as LineSeriesOption[]).filter(
        (item) => !chartUtil.isConnectedTooltipAnchorSeries({ seriesId: item.id }),
      );
      this.chartOptions.legend = {
        ...(this.chartOptions.legend as Record<string, unknown>),
        bottom: 15,
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
    updateYAxisLabels(measurements: MeasurementResponse[]): void {
      if (!this.chartOptions) return;

      this.chartOptions.yAxis = chartUtil.updateYAxisLabelsDefaultsIfNeeded(
        measurements,
        this.chartOptions.yAxis as IEChartsYAxisOptions,
        {},
      );
      const yAxis = this.chartOptions.yAxis as IEChartsYAxisOptions;
      const maxValue = this.getMaxMeasurementValue(measurements);

      if (maxValue === null || maxValue <= 0) {
        this.chartOptions.yAxis = chartUtil.updateYAxisLabelsDefaultsIfNeeded(measurements, yAxis, {});
        return;
      }

      this.chartOptions.yAxis = {
        ...yAxis,
        ...chartUtil.getLineChartYAxisConfig(maxValue, { nearZeroThreshold: 0.005 }),
        axisLabel: {
          ...(yAxis.axisLabel as Record<string, unknown>),
          formatter: chartUtil.formatCompactAxisValue,
        },
      } as EChartsOption["yAxis"];
    },
    getMaxMeasurementValue(measurements: MeasurementResponse[]): number | null {
      const values = measurements.flatMap((measurement: MeasurementResponse) =>
        (measurement.values ?? []).map((point: MeasurementResponseValuesInner) => Number(point.value)),
      );
      const finiteValues = values.filter((value: number) => Number.isFinite(value));

      if (!finiteValues.length) return null;

      return Math.max(...finiteValues);
    },
    handleDataZoom(event: Parameters<typeof chartUtil.getDataZoomRange>[0]): void {
      const range = chartUtil.getDataZoomRange(event, this.timeframes);

      if (range) {
        this.$emit("on-zoom", { ...range, chartId: this.chartId } as IZoomEvent);
      }
    },
    handleResetZoom(): void {
      chartUtil.resetChartDataZoom(this.chartId);
      this.$emit("on-reset-zoom", this.chartId);
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
    formatDatetimeAxisLabel(value: number | string): string {
      return chartUtil.formatDatetimeAxisLabel(value);
    },
    addRequest(): void {
      this.addChartData(WorkloadMetricType.CpuRequestCores, ELinesInCharts.Request);
    },
    addLimit(): void {
      this.addChartData(WorkloadMetricType.CpuLimitCores, ELinesInCharts.Limit);
    },
    addUsage(): void {
      this.addChartData(WorkloadMetricType.CpuUsageCores, ELinesInCharts.Usage);
    },
    addChartData(measurementName: WorkloadMetricType, seriesId: ELinesInCharts): void {
      const measurements: Array<MeasurementResponseValuesInner> | undefined = this.measurementMap[measurementName];
      const seriesConfig = this.getSeriesConfig(seriesId);
      if (!seriesConfig || !measurements) return;

      seriesConfig.data = this.getMeasurements(this.timeframes, measurements);
    },
    getSeriesConfig(seriesId: ELinesInCharts): LineSeriesOption | undefined {
      const series = (this.chartOptions?.series ?? []) as LineSeriesOption[];
      return series.find((s: LineSeriesOption) => s.id === seriesId);
    },
    async onPropsChanged(withLoading = true): Promise<void> {
      this.$emit("unregister");
      this.initialLoading = withLoading;
      intervalUtil.stopInterval(EIntervalLabels.CpuUsageWidget);
      await this.loadChartData();
      intervalUtil.startInterval(EIntervalLabels.CpuUsageWidget, this.loadChartData);
    },
  },
  watch: {
    workloadId(): void {
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
    selectedElementId(): void {
      this.onPropsChanged();
    },
  },
  beforeUnmount() {
    this.$emit("unregister");
    intervalUtil.stopInterval(EIntervalLabels.CpuUsageWidget);
  },
});
</script>

<style lang="scss" scoped>
.metric-chart-shell {
  position: relative;
}
</style>
