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
import { ChartWidgetWrapper } from "../../common/widget-wrapper/chart-widget-wrapper";
import { ChartAreaZoomControls } from "@/components/common/charts/chart-area-zoom-controls";

// stores
import { useSettingStore } from "@/stores/setting.store";

// models
import {
  type IWidgetWrapperOptions,
  type ITooltipPoint,
  type TMeasurementsTimeAndValue,
  type IZoomEvent,
  EYAxisLabelFormat,
  LEGEND_DASHED_LINE_ICON,
  LEGEND_RECT_ITEM_STYLE,
} from "@/models/chart.model";
import {
  ElementMetricType,
  WorkloadMetricType,
  type Workload,
  GpuRequestType,
  type MetricsResponse,
  type MeasurementResponse,
  type MeasurementResponseValuesInner,
  type WorkloadRequestResources,
} from "@/swagger-models/workloads-client";
import { EWorkloadType } from "@/models/workload.model";

// services
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// ECharts
import type { EChartsOption } from "echarts";
import type { LineSeriesOption } from "echarts/charts";

// utils
import { chartUtil } from "@/utils/chart.util";
import type { IEChartsXAxisOptions, IEChartsYAxisOptions } from "@/utils/chart.util/chart.util";
import { metricUtil } from "@/utils/metric.util";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import { intervalUtil } from "@/utils/interval.util";
import { formatK8sResourceMemory, convertToBytes, parseSizeString } from "@/utils/format.util/format.util";
import { EIntervalLabels } from "@/models/interval.model";

export const GPU_MEMORY_USAGE_CHART_ID = "gpu-memory-usage-chart";

export enum ELinesInCharts {
  Request = "request",
  Usage = "usage",
  Limit = "limit",
}

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
    selectedElementId: {
      type: String as PropType<string | null>,
      default: null,
    },
    zoomActive: {
      type: Boolean,
      default: false,
    },
    reloadTrigger: {
      type: Number,
      default: 0,
    },
    silentDateReload: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      settingStore: useSettingStore(),
      chartOptions: null as null | EChartsOption,
      error: false as boolean,
      initialLoading: true as boolean,
      timeframes: [] as Array<string>,
      measurementMap: {} as Record<string, Array<MeasurementResponseValuesInner>>,
      currWorkload: null as Workload | null,
      wrapperOptions: {
        title: "GPU memory usage",
      } as IWidgetWrapperOptions,
      chartId: GPU_MEMORY_USAGE_CHART_ID as string,
      gpuMemoryLimit: null as number | null,
    };
  },
  created() {
    this.loadChartOptions();
    this.onPropsChanged();
  },
  computed: {
    shouldDisplayUsageOnly(): boolean {
      const workloadRequestedResources: WorkloadRequestResources | undefined | null =
        this.currWorkload?.workloadRequestedResources;

      return (
        this.currWorkload?.type === EWorkloadType.Distributed ||
        !workloadRequestedResources ||
        workloadRequestedResources.gpuRequestType === GpuRequestType.Portion
      );
    },
    shouldShowRequest(): boolean {
      const request: string | null = this.currWorkload?.workloadRequestedResources?.gpuMemory?.request || null;
      return request !== null;
    },
    shouldShowLimit(): boolean {
      return this.gpuMemoryLimit !== null && this.settingStore.isGPUOverProvisioningEnabled;
    },
    metricType(): Array<WorkloadMetricType> {
      const metricType: Array<WorkloadMetricType> = [WorkloadMetricType.GpuMemoryUsageBytes];

      if (this.shouldShowRequest) {
        metricType.push(WorkloadMetricType.GpuMemoryRequestBytes);
      }

      return metricType;
    },
    elementMetricType(): Array<ElementMetricType> {
      const metricType: Array<ElementMetricType> = [ElementMetricType.GpuMemoryUsageBytes];

      if (this.shouldShowRequest) {
        metricType.push(ElementMetricType.GpuMemoryRequestBytes);
      }

      return metricType;
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
        grid: { top: 30, right: 10, bottom: 50, left: 34, containLabel: true },
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
        if (this.initialLoading) {
          this.currWorkload = await workloadService.getWorkloadById(this.workloadId);

          const gpuMemoryLimit: string = this.currWorkload?.workloadRequestedResources?.gpuMemory?.limit || "";
          this.gpuMemoryLimit = this.parseStringSizeToBytes(gpuMemoryLimit);
          this.addSeriesDetails();
        }

        const updatedDates: IRangeDates = dateUtil.getUpdatedTime(this.filterByDates);
        this.timeframes = chartUtil.getTimeFrames(updatedDates.dateStart, updatedDates.dateEnd);
        this.setTooltipAnchorSeries();
        const samples = this.timeframes.length;

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

        this.setSeries();
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
    setSeries(): void {
      const gpuMemoryUsageMeasurements: Array<MeasurementResponseValuesInner> | undefined =
        this.measurementMap[WorkloadMetricType.GpuMemoryUsageBytes];

      if (!gpuMemoryUsageMeasurements) return;

      const gpuMemoryUsageData: Array<TMeasurementsTimeAndValue> = this.getMeasurements(
        this.timeframes,
        gpuMemoryUsageMeasurements,
      );

      this.addRequest();
      this.addUsage(gpuMemoryUsageData);
      this.addLimit(gpuMemoryUsageData);
    },
    setTooltipAnchorSeries(): void {
      if (!this.chartOptions) return;
      const series = Array.isArray(this.chartOptions.series) ? (this.chartOptions.series as LineSeriesOption[]) : [];
      const visibleSeries = series.filter((item) => !chartUtil.isConnectedTooltipAnchorSeries({ seriesId: item.id }));
      this.chartOptions.series = [chartUtil.getConnectedTooltipAnchorSeries(this.timeframes), ...visibleSeries];
    },
    addRequest(): void {
      const measurements: Array<MeasurementResponseValuesInner> | undefined =
        this.measurementMap[WorkloadMetricType.GpuMemoryRequestBytes];
      const seriesConfig = this.getSeriesConfig(ELinesInCharts.Request);
      if (!seriesConfig || !measurements) return;
      const seriesData: Array<TMeasurementsTimeAndValue> = this.getMeasurements(this.timeframes, measurements);
      seriesConfig.data = seriesData;
    },
    addUsage(gpuMemoryUsageData: Array<TMeasurementsTimeAndValue>): void {
      const seriesConfig = this.getSeriesConfig(ELinesInCharts.Usage);
      if (!seriesConfig) return;
      seriesConfig.data = gpuMemoryUsageData;
    },
    addLimit(gpuMemoryUsageData: Array<TMeasurementsTimeAndValue>): void {
      if (!this.gpuMemoryLimit) return;

      const updatedTimeframes: Array<number> = gpuMemoryUsageData.map(
        (gpuUsage: TMeasurementsTimeAndValue) => gpuUsage[0],
      );

      const gpuMemoryLimitData: Array<TMeasurementsTimeAndValue> = this.insertConstantValueIntoTimeframe(
        this.gpuMemoryLimit,
        updatedTimeframes,
      );

      const seriesConfig = this.getSeriesConfig(ELinesInCharts.Limit);
      if (!seriesConfig) return;
      seriesConfig.data = gpuMemoryLimitData;
    },
    addSeriesDetails(): void {
      if (!this.chartOptions || !this.currWorkload) return;

      const series: LineSeriesOption[] = [];

      if (this.shouldShowRequest) {
        series.push({
          id: ELinesInCharts.Request,
          name: "GPU memory request",
          type: "line",
          lineStyle: chartUtil.getDottedLineStyle(),
          showSymbol: false,
          symbol: "circle",
        });
      }

      series.push({
        id: ELinesInCharts.Usage,
        name: "GPU memory usage",
        type: "line",
        showSymbol: false,
        symbol: "circle",
      });

      if (this.shouldShowLimit) {
        series.push({
          id: ELinesInCharts.Limit,
          name: "GPU memory limit",
          type: "line",
          lineStyle: chartUtil.getDottedLineStyle(),
          showSymbol: false,
          symbol: "circle",
        });
      }

      this.chartOptions.series = this.applyDashboardLineSeriesStyle(series);
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

      const yAxis = this.chartOptions.yAxis as IEChartsYAxisOptions;
      const maxValue = this.getMaxMeasurementValue(measurements);

      if (maxValue === null) {
        this.chartOptions.yAxis = chartUtil.updateYAxisLabelsDefaultsIfNeeded(measurements, yAxis, {
          type: EYAxisLabelFormat.Memory,
        });
        return;
      }

      if (maxValue <= 0) {
        this.chartOptions.yAxis = {
          ...yAxis,
          ...chartUtil.getLineChartYAxisConfig(maxValue),
          axisLabel: { formatter: (value: number) => formatK8sResourceMemory(value) },
        };
        return;
      }

      const interval = this.getNiceInterval(maxValue / 3);
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
    getMeasurements(
      timeframes: string[],
      measurements: Array<MeasurementResponseValuesInner>,
    ): Array<TMeasurementsTimeAndValue> {
      return chartUtil.arrangeMeasurementsByTime(timeframes, measurements, { useTimeFrameTimestamps: true });
    },
    getSeriesConfig(seriesId: ELinesInCharts): LineSeriesOption | undefined {
      const series = (this.chartOptions?.series ?? []) as LineSeriesOption[];
      return series.find((s: LineSeriesOption) => s.id === seriesId);
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
    parseStringSizeToBytes(sizeString: string): number | null {
      const sizeAndString: [number, string] | null = parseSizeString(sizeString);
      if (!sizeAndString) return null;
      return convertToBytes(sizeAndString[0], sizeAndString[1]);
    },
    async onPropsChanged(withLoading = true): Promise<void> {
      this.$emit("unregister");
      intervalUtil.stopInterval(EIntervalLabels.GpuMemoryUsageWidget);
      this.initialLoading = withLoading;
      await this.loadChartData();
      intervalUtil.startInterval(EIntervalLabels.GpuMemoryUsageWidget, this.loadChartData);
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
    reloadTrigger(): void {
      this.onPropsChanged();
    },
    selectedElementId(): void {
      this.onPropsChanged();
    },
  },
  beforeUnmount() {
    this.$emit("unregister");
    intervalUtil.stopInterval(EIntervalLabels.GpuMemoryUsageWidget);
  },
});
</script>

<style lang="scss" scoped>
.metric-chart-shell {
  position: relative;
}
</style>
