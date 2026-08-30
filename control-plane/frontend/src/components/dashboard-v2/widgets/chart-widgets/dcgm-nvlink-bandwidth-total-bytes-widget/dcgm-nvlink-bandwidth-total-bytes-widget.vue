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
  CHART_COLOR_PALETTE,
  LEGEND_RECT_ITEM_STYLE,
} from "@/models/chart.model";
import {
  WorkloadMetricType,
  type Workload,
  type MetricsResponse,
  type MeasurementResponseValuesInner,
  type MeasurementResponse,
} from "@/swagger-models/workloads-client";

// echarts
import type { EChartsOption } from "echarts";
import type { LineSeriesOption } from "echarts/charts";

// services
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// utils
import { chartUtil } from "@/utils/chart.util";
import { metricUtil } from "@/utils/metric.util";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { formatUtil } from "@/utils/format.util/format.util";
import type { IEChartsXAxisOptions, IEChartsYAxisOptions } from "@/utils/chart.util/chart.util";

export const DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_CHART_ID = "dcgm-nvlink-bandwidth-total-bytes-chart";

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
      measurementMap: {} as Record<string, Array<MeasurementResponseValuesInner>>,
      currWorkload: null as Workload | null,
      wrapperOptions: {
        title: "DCGM NVLink bandwidth total bytes",
      } as IWidgetWrapperOptions,
      chartId: DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_CHART_ID as string,
    };
  },
  created() {
    this.loadChartOptions();
    this.onPropsChanged();
  },
  computed: {
    metricType(): Array<WorkloadMetricType> {
      return [WorkloadMetricType.NvlinkBandwidthTotal];
    },
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
            color: CHART_COLOR_PALETTE.BLUE,
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
        itemGap: 20,
        textStyle: {
          fontSize: 11,
          fontWeight: "bold",
        },
        data: [{ name: "DCGM NVLink bandwidth total bytes", icon: LEGEND_RECT_ITEM_STYLE.icon }],
      };
    },
    async loadChartData(): Promise<void> {
      if (this.chartOptions === null) return;

      try {
        if (this.initialLoading) {
          this.currWorkload = await workloadService.getWorkloadById(this.workloadId);
          this.addSeriesDetails();
        }

        const updatedDates: IRangeDates = dateUtil.getUpdatedTime(this.filterByDates);
        this.timeframes = chartUtil.getTimeFrames(updatedDates.dateStart, updatedDates.dateEnd);
        const samples = this.timeframes.length;

        const metricsResponse: MetricsResponse = await workloadService.getWorkloadMetrics(
          this.workloadId,
          dateUtil.convertDateToISO(updatedDates.dateStart),
          dateUtil.convertDateToISO(updatedDates.dateEnd),
          this.metricType,
          samples,
        );

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
      this.addUsage();
    },
    getMeasurements(
      timeframes: string[],
      measurements: Array<MeasurementResponseValuesInner>,
    ): Array<TMeasurementsTimeAndValue> {
      return chartUtil.arrangeMeasurementsByTime(timeframes, measurements);
    },
    addSeriesDetails(): void {
      if (!this.chartOptions) return;

      this.chartOptions.series = [
        {
          id: "usage",
          name: "DCGM NVLink bandwidth total bytes",
          type: "line",
          showSymbol: false,
          ...chartUtil.getConnectedChartHoverMarker(CHART_COLOR_PALETTE.BLUE),
          lineStyle: {
            color: CHART_COLOR_PALETTE.BLUE,
          },
          itemStyle: {
            color: CHART_COLOR_PALETTE.BLUE,
          },
        },
      ] as LineSeriesOption[];
    },
    addUsage(): void {
      this.addChartData(WorkloadMetricType.NvlinkBandwidthTotal, "usage");
    },
    addChartData(measurementName: WorkloadMetricType, seriesId: string): void {
      const measurements: Array<MeasurementResponseValuesInner> | undefined = this.measurementMap[measurementName];
      const seriesConfig: LineSeriesOption | undefined = this.getSeriesConfig(seriesId);
      if (!seriesConfig || !measurements) return;
      const seriesData: Array<TMeasurementsTimeAndValue> = this.getMeasurements(this.timeframes, measurements);
      seriesConfig.data = seriesData;
    },
    getSeriesConfig(seriesId: string): LineSeriesOption | undefined {
      const series = (this.chartOptions?.series ?? []) as LineSeriesOption[];
      return series.find((s: LineSeriesOption) => s.id === seriesId);
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
    formatDatetimeAxisLabel(value: number): string {
      return chartUtil.formatDatetimeAxisLabel(value);
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
    async onPropsChanged(withLoading = true): Promise<void> {
      this.$emit("unregister");
      this.initialLoading = withLoading;
      intervalUtil.stopInterval(EIntervalLabels.DcgmNvlinkBandwidthTotalBytesWidget);
      await this.loadChartData();
      intervalUtil.startInterval(EIntervalLabels.DcgmNvlinkBandwidthTotalBytesWidget, this.loadChartData);
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
  },
  beforeUnmount() {
    this.$emit("unregister");
    intervalUtil.stopInterval(EIntervalLabels.DcgmNvlinkBandwidthTotalBytesWidget);
  },
});
</script>

<style lang="scss" scoped>
.metric-chart-shell {
  position: relative;
}
</style>
