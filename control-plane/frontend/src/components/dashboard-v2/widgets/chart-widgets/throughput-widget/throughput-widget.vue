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
} from "@/models/chart.model";
import {
  InferenceWorkloadMetricType,
  InferencePodMetricType,
  type MetricsResponse,
  type MeasurementResponseValuesInner,
} from "@/swagger-models/workloads-client";

// echarts
import type { EChartsOption } from "echarts";
import type { LineSeriesOption } from "echarts/charts";

// services
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";

// utils
import { chartUtil } from "@/utils/chart.util";
import type { IEChartsYAxisOptions } from "@/utils/chart.util/chart.util";
import { metricUtil } from "@/utils/metric.util";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import { intervalUtil } from "@/utils/interval.util";
import { formatToFixedOrZero } from "@/utils/format.util";
import { EIntervalLabels } from "@/models/interval.model";
import {
  getInferenceChartLegendOptions,
  getInferenceTimeXAxisOptions,
  INFERENCE_CHART_BLUE,
  inferenceChartAxisLabel,
  inferenceChartGrid,
  inferenceChartSplitLine,
  inferenceTooltipAxisPointer,
} from "@/components/dashboard-v2/widgets/chart-widgets/inference-chart-options.util";

export const THROUGHPUT_WIDGET_CHART_ID = "throughput-chart";
const THROUGHPUT_LEGEND_NAMES = ["Total throughput"];

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
      required: false,
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
      measurementMap: {} as Record<string, Array<MeasurementResponseValuesInner>>,
      wrapperOptions: {
        title: "Throughput",
      } as IWidgetWrapperOptions,
      chartId: THROUGHPUT_WIDGET_CHART_ID as string,
    };
  },
  created() {
    this.loadChartOptions();
    this.onPropsChanged();
  },
  methods: {
    loadChartOptions(): void {
      this.chartOptions = {
        color: [INFERENCE_CHART_BLUE],
        grid: inferenceChartGrid,
        xAxis: this.getXAxisOptions(),
        yAxis: {
          type: "value",
          name: "Requests / Sec",
          nameLocation: "middle",
          nameRotate: 90,
          nameGap: 50,
          min: 0,
          axisLabel: inferenceChartAxisLabel,
          axisLine: { show: false },
          axisTick: { show: false },
          splitLine: inferenceChartSplitLine,
        },
        tooltip: {
          ...chartUtil.buildStandardTooltip({ trigger: "axis" }),
          trigger: "axis",
          axisPointer: inferenceTooltipAxisPointer,
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
              const pointY = item.value[1] != null ? `${formatToFixedOrZero(item.value[1])} Requests/Sec` : undefined;
              return {
                name: item.seriesName,
                y: pointY,
                color: item.color,
              };
            });
            return chartUtil.formatSharedTooltip(points, (visibleItems[0] || items[0]).value[0]) as string;
          },
        },
        legend: {
          ...getInferenceChartLegendOptions(THROUGHPUT_LEGEND_NAMES),
          bottom: 16,
        },
        aria: chartUtil.buildChartAria("Inference request throughput over time."),
        dataZoom: [
          {
            type: "inside",
            disabled: true,
            zoomOnMouseWheel: false,
            moveOnMouseWheel: false,
            moveOnMouseMove: false,
            preventDefaultMouseMove: false,
          },
        ],
        series: [],
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
        this.setTooltipAnchorSeries();
        const samples = this.timeframes.length;

        let metricsResponse: MetricsResponse;
        if (this.podId) {
          const podMetricType: Array<InferencePodMetricType> = [InferencePodMetricType.Throughput];
          metricsResponse = await inferenceService.getPodMetrics(
            this.workloadId,
            this.podId,
            dateUtil.convertDateToISO(updatedDates.dateStart),
            dateUtil.convertDateToISO(updatedDates.dateEnd),
            podMetricType,
            samples,
          );
        } else {
          const metricType: Array<InferenceWorkloadMetricType> = [InferenceWorkloadMetricType.Throughput];
          metricsResponse = await inferenceService.getMetrics(
            this.workloadId,
            dateUtil.convertDateToISO(updatedDates.dateStart),
            dateUtil.convertDateToISO(updatedDates.dateEnd),
            metricType,
            samples,
          );
        }

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

      let throughputData: Array<TMeasurementsTimeAndValue> = [];
      const metricKey = this.podId ? InferencePodMetricType.Throughput : InferenceWorkloadMetricType.Throughput;
      if (this.measurementMap[metricKey]) {
        throughputData = this.getMeasurements(this.timeframes, this.measurementMap[metricKey]);
      }

      const series = this.chartOptions.series as LineSeriesOption[];
      const throughputSeries = series.find((item) => !chartUtil.isConnectedTooltipAnchorSeries({ seriesId: item.id }));
      if (throughputSeries) {
        throughputSeries.data = throughputData;
      }
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
          name: "Total throughput",
          type: "line",
          showSymbol: false,
          ...chartUtil.getConnectedChartHoverMarker(INFERENCE_CHART_BLUE),
          lineStyle: { color: INFERENCE_CHART_BLUE },
          itemStyle: { color: INFERENCE_CHART_BLUE },
        },
      ] as LineSeriesOption[];
    },
    getMeasurements(
      timeframes: string[],
      measurements: Array<MeasurementResponseValuesInner>,
    ): Array<TMeasurementsTimeAndValue> {
      return chartUtil.arrangeMeasurementsByTime(timeframes, measurements, { useTimeFrameTimestamps: true });
    },
    updateYAxisLabels(measurements: MetricsResponse["measurements"]): void {
      if (!this.chartOptions) return;

      this.chartOptions.yAxis = chartUtil.updateYAxisLabelsDefaultsIfNeeded(
        measurements,
        this.chartOptions.yAxis as IEChartsYAxisOptions,
        { steps: [0, 20, 40, 60, 80, 100] },
      );
    },
    getXAxisOptions(): Record<string, unknown> {
      const xAxisOptions = getInferenceTimeXAxisOptions();
      return {
        ...xAxisOptions,
        axisLabel: {
          ...((xAxisOptions.axisLabel || {}) as Record<string, unknown>),
          formatter: this.formatXAxisTimestamp,
        },
      };
    },
    formatXAxisTimestamp(value: number | string): string {
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
    async onPropsChanged(withLoading = true): Promise<void> {
      this.$emit("unregister");
      this.initialLoading = withLoading;
      intervalUtil.stopInterval(EIntervalLabels.ThroughputWidget);
      await this.loadChartData();
      intervalUtil.startInterval(EIntervalLabels.ThroughputWidget, this.loadChartData);
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
    intervalUtil.stopInterval(EIntervalLabels.ThroughputWidget);
  },
});
</script>

<style lang="scss" scoped>
.metric-chart-shell {
  position: relative;
}
</style>
