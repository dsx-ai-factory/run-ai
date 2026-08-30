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
        @legendselectchanged="handleLegendSelectChanged"
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

// echarts
import type { EChartsOption } from "echarts";
import type { LineSeriesOption } from "echarts/charts";

// services
import { podService } from "@/services/control-plane/pod.service/pod.service";

// utils
import { chartUtil } from "@/utils/chart.util";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { formatUtil } from "@/utils/format.util/format.util";
import type { IEChartsXAxisOptions, IEChartsYAxisOptions } from "@/utils/chart.util/chart.util";

export const DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_PER_GPU_CHART_ID = "dcgm-nvlink-bandwidth-total-bytes-per-gpu-chart";

const LINE_LEGEND_ITEM_GAP = 20;
const MAX_VISIBLE_GPU_SERIES = 8;

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
      initialLoading: true as boolean,
      timeframes: [] as Array<string>,
      wrapperOptions: {
        title: "DCGM NVLink bandwidth total bytes",
      } as IWidgetWrapperOptions,
      measurements: [] as Array<MeasurementResponse>,
      legendSelection: {} as Record<string, boolean>,
      chartId: DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_PER_GPU_CHART_ID as string,
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
        yAxisFormatFunction: formatUtil.formatBytesRate,
        grid: {
          top: 10,
          left: 90,
          right: 20,
          bottom: 87,
          containLabel: false,
        },
      });

      this.addTooltipOptions();
      this.addXAxisOptions();
      this.addZoomOptions();
      this.addLegendOptions();
    },
    addTooltipOptions(): void {
      if (!this.chartOptions) return;

      this.chartOptions.tooltip = {
        ...this.chartOptions.tooltip,
        axisPointer: {
          ...chartUtil.getConnectedChartAxisPointer(),
          label: {
            show: false,
          },
        },
        formatter: (params: unknown) => {
          const items = params as Array<{ seriesName: string; value: [number, number | null]; color: string }>;
          if (!Array.isArray(items) || !items.length) return "";
          const points: Array<ITooltipPoint> = items.map((item) => ({
            name: item.seriesName,
            y: item.value[1] != null ? formatUtil.formatBytesRate(item.value[1]) : undefined,
            color: item.color,
          }));
          return chartUtil.formatSharedTooltip(points, items[0].value[0]) as string;
        },
      };
    },
    addXAxisOptions(): void {
      if (!this.chartOptions) return;

      this.chartOptions.xAxis = {
        ...chartUtil.getDatetimeOptions(this.chartOptions.xAxis as IEChartsXAxisOptions),
        axisLine: {
          onZero: false,
          lineStyle: {
            color: "#CCD6EB",
          },
        },
        axisTick: {
          show: false,
        },
        axisLabel: chartUtil.getDatetimeAxisLabelOptions((value: number) => {
          return this.formatDatetimeAxisLabel(value);
        }),
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
    addLegendOptions(): void {
      if (!this.chartOptions) return;

      this.chartOptions.legend = {
        ...(this.chartOptions.legend as Record<string, unknown>),
        bottom: 14,
        ...LEGEND_RECT_ITEM_STYLE,
        itemGap: LINE_LEGEND_ITEM_GAP,
        textStyle: {
          fontWeight: "bold",
          fontSize: 11,
          color: "#333333",
        },
      };
    },
    async loadChartData(): Promise<void> {
      if (this.chartOptions === null) return;

      try {
        if (this.initialLoading) {
          this.addSeriesDetails();
        }

        const updatedDates: IRangeDates = dateUtil.getUpdatedTime(this.filterByDates);
        this.timeframes = chartUtil.getTimeFrames(updatedDates.dateStart, updatedDates.dateEnd);
        const samples = this.timeframes.length;

        const metricType: Array<PodMetricType> = [PodMetricType.NvlinkBandwidthTotal];
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
    addSeriesDetails(): void {
      if (!this.chartOptions) return;
      this.chartOptions.series = [
        {
          id: "usage",
          name: "DCGM NVLink bandwidth total bytes",
          type: "line",
          showSymbol: false,
          symbol: "none",
        },
      ];
    },
    setSeries(): void {
      if (!this.chartOptions) return;
      this.chartOptions.series = [];

      this.measurements.forEach((measurement: MeasurementResponse, index: number) => {
        const name: string = measurement.labels?.gpu ? `GPU-${measurement.labels?.gpu}` : "N/A";
        let seriesConfig = this.getSeriesConfig(name);
        if (!measurement.values) return;

        this.ensureLegendSelection(name, index);

        if (seriesConfig) {
          seriesConfig.data = this.getMeasurements(this.timeframes, measurement.values);
        } else {
          seriesConfig = this.getDefaultSeriesConfig(name);
          seriesConfig.data = this.getMeasurements(this.timeframes, measurement.values);
          (this.chartOptions?.series as LineSeriesOption[])?.push(seriesConfig);
        }
      });

      ((this.chartOptions.series ?? []) as LineSeriesOption[]).forEach((seriesOption, index) => {
        Object.assign(
          seriesOption,
          chartUtil.getConnectedChartHoverMarker(chartColorPaletteArray[index % chartColorPaletteArray.length]),
        );
      });
    },
    ensureLegendSelection(name: string, seriesIndex: number): void {
      if (Object.prototype.hasOwnProperty.call(this.legendSelection, name)) return;

      this.legendSelection[name] = seriesIndex < MAX_VISIBLE_GPU_SERIES;
    },
    getDefaultSeriesConfig(name: string): LineSeriesOption {
      return {
        name,
        type: "line",
        showSymbol: false,
        symbol: "none",
      };
    },
    getSeriesConfig(name: string): LineSeriesOption | undefined {
      const series = (this.chartOptions?.series ?? []) as LineSeriesOption[];
      return series.find((s: LineSeriesOption) => s.name === name);
    },
    updateYAxisLabels(measurements: MeasurementResponse[]): void {
      if (!this.chartOptions) return;

      const yAxis = this.chartOptions.yAxis as IEChartsYAxisOptions;
      const hasPositiveValue = measurements.some((measurement: MeasurementResponse) =>
        measurement.values?.some((point: MeasurementResponseValuesInner) => Number(point.value) > 0),
      );
      const hasNumericValue = measurements.some((measurement: MeasurementResponse) =>
        measurement.values?.some((point: MeasurementResponseValuesInner) => Number.isFinite(Number(point.value))),
      );

      if (hasPositiveValue) {
        const restYAxis = { ...yAxis };
        delete restYAxis.max;
        delete restYAxis.interval;
        delete restYAxis.splitNumber;
        this.chartOptions.yAxis = chartUtil.updateYAxisLabelsDefaultsIfNeeded(measurements, restYAxis, {
          type: EYAxisLabelFormat.Memory,
        });
        return;
      }

      if (hasNumericValue) {
        this.chartOptions.yAxis = {
          ...yAxis,
          ...chartUtil.getLineChartYAxisConfig(0),
          axisLabel: { formatter: (value: number) => formatUtil.formatBytesRate(value) },
        };
        return;
      }

      this.chartOptions.yAxis = {
        ...yAxis,
        min: 0,
        max: 80e9,
        interval: 20e9,
        axisLabel: {
          formatter: (value: number) => {
            if (value === 0) return "0 Bytes";
            return `${value / 1e9} G`;
          },
        },
      };
    },
    updateLegendOptions(): void {
      if (!this.chartOptions) return;

      this.chartOptions.legend = {
        ...(this.chartOptions.legend as Record<string, unknown>),
        data: ((this.chartOptions.series ?? []) as LineSeriesOption[]).map((seriesOption: LineSeriesOption) => ({
          name: String(seriesOption.name),
          icon: LEGEND_RECT_ITEM_STYLE.icon,
        })),
        selected: { ...this.legendSelection },
      };
    },
    handleLegendSelectChanged(params: { selected?: Record<string, boolean> }): void {
      if (!params.selected) return;

      this.legendSelection = { ...params.selected };
      this.updateLegendOptions();
    },
    formatDatetimeAxisLabel(value: number): string {
      return chartUtil.formatDatetimeAxisLabel(value);
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
    getMeasurements(
      timeframes: string[],
      measurements: Array<MeasurementResponseValuesInner>,
    ): Array<TMeasurementsTimeAndValue> {
      if (!measurements) return [];
      return chartUtil.arrangeMeasurementsByTime(timeframes, measurements);
    },
    async onPropsChanged(withLoading = true): Promise<void> {
      this.$emit("unregister");
      this.initialLoading = withLoading;
      intervalUtil.stopInterval(EIntervalLabels.DcgmNvlinkBandwidthTotalBytesPerGpuWidget);
      await this.loadChartData();
      intervalUtil.startInterval(EIntervalLabels.DcgmNvlinkBandwidthTotalBytesPerGpuWidget, this.loadChartData);
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
    intervalUtil.stopInterval(EIntervalLabels.DcgmNvlinkBandwidthTotalBytesPerGpuWidget);
  },
});
</script>

<style lang="scss" scoped>
.metric-chart-shell {
  position: relative;
}
</style>
