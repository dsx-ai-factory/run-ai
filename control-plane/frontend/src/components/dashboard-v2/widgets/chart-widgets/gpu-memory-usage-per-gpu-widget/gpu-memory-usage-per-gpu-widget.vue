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
  LEGEND_RECT_ITEM_STYLE,
  chartColorPaletteArray,
} from "@/models/chart.model";
import {
  PodMetricType,
  type MetricsResponse,
  type MeasurementResponseValuesInner,
  type MeasurementResponse,
} from "@/swagger-models/workloads-client";

// ECharts
import type { EChartsOption } from "echarts";
import type { LineSeriesOption } from "echarts/charts";

// services
import { podService } from "@/services/control-plane/pod.service/pod.service";

// utils
import { chartUtil } from "@/utils/chart.util";
import type { IEChartsXAxisOptions, IEChartsYAxisOptions } from "@/utils/chart.util/chart.util";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import { intervalUtil } from "@/utils/interval.util";
import { formatK8sResourceMemory } from "@/utils/format.util/format.util";
import { EIntervalLabels } from "@/models/interval.model";

export const GPU_MEMORY_USAGE_PER_GPU_CHART_ID = "gpu-memory-usage-per-gpu-chart";
const LINE_LEGEND_ITEM_GAP = 20;

export default defineComponent({
  components: { VChart, ChartWidgetWrapper, ChartAreaZoomControls },
  emits: ["register", "unregister", "on-zoom", "on-reset-zoom", "on-zoom-toggle"],
  props: {
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
        title: "GPU memory usage per GPU",
      } as IWidgetWrapperOptions,
      measurements: [] as Array<MeasurementResponse>,
      chartId: GPU_MEMORY_USAGE_PER_GPU_CHART_ID as string,
    };
  },
  created() {
    this.loadChartOptions();
    this.onPropsChanged();
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
        itemGap: LINE_LEGEND_ITEM_GAP,
        textStyle: {
          fontSize: 11,
          fontWeight: 700,
          color: "#333333",
        },
        data: chartUtil.getDashboardLineOnlyLegendData(
          series.map((seriesOption: LineSeriesOption) => String(seriesOption.name)),
        ),
      };
    },
    async loadChartData(): Promise<void> {
      if (this.chartOptions === null) return;

      try {
        if (this.initialLoading) {
          this.chartOptions.series = [];
        }

        const updatedDates: IRangeDates = dateUtil.getUpdatedTime(this.filterByDates);
        this.timeframes = chartUtil.getTimeFrames(updatedDates.dateStart, updatedDates.dateEnd);
        const samples = this.timeframes.length;

        const metricType: Array<PodMetricType> = [PodMetricType.GpuMemoryUsageBytesPerGpu];
        const metricsResponse: MetricsResponse = await podService.getPodMetrics(
          this.workloadId,
          this.podId,
          dateUtil.convertDateToISO(updatedDates.dateStart),
          dateUtil.convertDateToISO(updatedDates.dateEnd),
          metricType,
          samples,
        );

        this.measurements = metricsResponse.measurements;
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
      if (!this.chartOptions) return;
      this.chartOptions.series = [chartUtil.getConnectedTooltipAnchorSeries(this.timeframes)];

      this.measurements.forEach((measurement: MeasurementResponse) => {
        const name: string = measurement.labels?.gpu ? `GPU-${measurement.labels?.gpu}` : "N/A";
        let seriesConfig = this.getSeriesConfig(name);
        if (!measurement.values) return;

        if (seriesConfig) {
          seriesConfig.data = this.getMeasurements(this.timeframes, measurement.values);
        } else {
          seriesConfig = {
            name,
            type: "line",
            showSymbol: false,
            symbol: "none",
            connectNulls: true,
            data: this.getMeasurements(this.timeframes, measurement.values),
          };
          (this.chartOptions?.series as LineSeriesOption[]).push(seriesConfig);
        }
      });

      ((this.chartOptions.series ?? []) as LineSeriesOption[])
        .filter((item) => !chartUtil.isConnectedTooltipAnchorSeries({ seriesId: item.id }))
        .forEach((seriesOption, index) => {
          Object.assign(
            seriesOption,
            chartUtil.getConnectedChartHoverMarker(chartColorPaletteArray[index % chartColorPaletteArray.length]),
          );
        });
    },
    getSeriesConfig(name: string): LineSeriesOption | undefined {
      const series = (this.chartOptions?.series ?? []) as LineSeriesOption[];
      return series.find((s: LineSeriesOption) => s.name === name);
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
    getMeasurements(
      timeframes: string[],
      measurements: Array<MeasurementResponseValuesInner>,
    ): Array<TMeasurementsTimeAndValue> {
      if (!measurements) return [];
      return chartUtil.arrangeMeasurementsByTime(timeframes, measurements, { useTimeFrameTimestamps: true });
    },
    async onPropsChanged(withLoading = true): Promise<void> {
      this.$emit("unregister");
      this.initialLoading = withLoading;
      intervalUtil.stopInterval(EIntervalLabels.GpuMemoryUsagePerGpuWidget);
      await this.loadChartData();
      intervalUtil.startInterval(EIntervalLabels.GpuMemoryUsagePerGpuWidget, this.loadChartData);
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
    intervalUtil.stopInterval(EIntervalLabels.GpuMemoryUsagePerGpuWidget);
  },
});
</script>

<style lang="scss" scoped>
.metric-chart-shell {
  position: relative;
}
</style>
