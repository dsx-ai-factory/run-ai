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
  LEGEND_RECT_ITEM_STYLE,
  type IWidgetWrapperOptions,
  type ITooltipPoint,
  type TMeasurementsTimeAndValue,
  type IZoomEvent,
} from "@/models/chart.model";
import {
  ElementMetricType,
  WorkloadMetricType,
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

export const GPU_COMPUTE_UTILIZATION_WIDGET_CHART_ID = "gpu-compute-utilization-chart";

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
    reloadTrigger: {
      type: Number,
      default: 0,
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
      initialLoading: false as boolean,
      timeframes: [] as Array<string>,
      measurementMap: {} as Record<string, Array<MeasurementResponseValuesInner>>,
      wrapperOptions: {
        title: "GPU compute utilization",
      } as IWidgetWrapperOptions,
      chartId: GPU_COMPUTE_UTILIZATION_WIDGET_CHART_ID as string,
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
        yAxisFormatFunction: (value: string | number): string => `${value}%`,
        grid: { top: 30, right: 10, bottom: 50, left: 34, containLabel: true },
      });

      this.chartOptions.yAxis = {
        ...(this.chartOptions.yAxis as IEChartsYAxisOptions),
        max: 100,
        interval: 25,
        axisLabel: { formatter: (value: number) => `${value}%` },
      };

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
            let pointY = undefined;
            if (item.value[1] != null) {
              pointY = formatToFixedOrZero(item.value[1]);
            }
            return {
              name: item.seriesName,
              y: pointY,
              color: item.color,
              symbol: "%",
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
              [ElementMetricType.GpuUtilization],
              start,
              end,
              samples,
            )
          : await workloadService.getWorkloadMetrics(
              this.workloadId,
              start,
              end,
              [WorkloadMetricType.GpuUtilization],
              samples,
            );

        this.measurementMap = metricUtil.mapMeasurementsToMap(metricsResponse);

        this.setSeries();
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
    setSeries(): void {
      if (!this.chartOptions?.series) return;

      let gpuUtilizationData: Array<TMeasurementsTimeAndValue> = [];
      if (this.measurementMap[WorkloadMetricType.GpuUtilization]) {
        gpuUtilizationData = this.getMeasurements(
          this.timeframes,
          this.measurementMap[WorkloadMetricType.GpuUtilization],
        );
      }

      const series = this.chartOptions.series as LineSeriesOption[];
      const gpuSeries = series.find((item) => !chartUtil.isConnectedTooltipAnchorSeries({ seriesId: item.id }));
      if (gpuSeries) gpuSeries.data = gpuUtilizationData;
    },
    setTooltipAnchorSeries(): void {
      if (!this.chartOptions) return;
      const series = Array.isArray(this.chartOptions.series) ? (this.chartOptions.series as LineSeriesOption[]) : [];
      const visibleSeries = series.filter((item) => !chartUtil.isConnectedTooltipAnchorSeries({ seriesId: item.id }));
      this.chartOptions.series = [chartUtil.getConnectedTooltipAnchorSeries(this.timeframes), ...visibleSeries];
    },
    addSeriesDetails(): void {
      if (!this.chartOptions) return;

      this.chartOptions.series = [
        {
          name: "GPU compute utilization",
          type: "line",
          showSymbol: false,
          ...chartUtil.getDashboardLineSeriesStyle(),
        },
      ];
      this.updateLegendOptions(["GPU compute utilization"]);
    },
    updateLegendOptions(names: string[]): void {
      if (!this.chartOptions) return;

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
        data: chartUtil.getDashboardLineOnlyLegendData(names),
      };
    },
    updateYAxisLabels(measurements: MeasurementResponse[]): void {
      if (!this.chartOptions) return;

      this.chartOptions.yAxis = chartUtil.updateYAxisLabelsDefaultsIfNeeded(
        measurements,
        this.chartOptions.yAxis as IEChartsYAxisOptions,
        { steps: [0, 25, 50, 75, 100] },
      );
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
    async onPropsChanged(withLoading = true): Promise<void> {
      this.$emit("unregister");
      this.initialLoading = withLoading;
      intervalUtil.stopInterval(EIntervalLabels.GpuComputeUtilizationWidget);
      await this.loadChartData();
      intervalUtil.startInterval(EIntervalLabels.GpuComputeUtilizationWidget, this.loadChartData);
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
    intervalUtil.stopInterval(EIntervalLabels.GpuComputeUtilizationWidget);
  },
});
</script>

<style lang="scss" scoped>
.metric-chart-shell {
  position: relative;
}
</style>
