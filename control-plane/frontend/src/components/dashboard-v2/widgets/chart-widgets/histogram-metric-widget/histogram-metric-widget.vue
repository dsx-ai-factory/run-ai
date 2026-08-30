<template>
  <chart-widget-wrapper
    :options="wrapperOptions"
    :body-no-padding="wrapperProps.bodyNoPadding"
    :loading="initialLoading"
    :error="error"
    :hide-border="wrapperProps.hideBorder"
    :flat="wrapperProps.flat"
    :hide-actions="wrapperProps.hideActions"
    :hide-separator="wrapperProps.hideSeparator"
  >
    <v-chart
      v-if="chartOptions"
      :option="chartOptions"
      :id="config.chartId"
      :style="{ height: config.chartOptions.height + 'px' }"
      autoresize
      @datazoom="handleDataZoom"
      @restore="handleResetZoom"
      @zr:click="handleResetZoom"
    />
  </chart-widget-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import VChart from "vue-echarts";
import "@/plugins/echarts";
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";

// models
import type {
  IColumnHistogramDataParser,
  IGenericWidgetWrapperConfig,
  IWidgetWrapperOptions,
  IZoomEvent,
} from "@/models/chart.model";
import type {
  MeasurementResponse,
  MeasurementResponseValuesInner,
  MetricsCompositeResponse,
} from "@/swagger-models/workloads-client";

import type { BarSeriesOption } from "echarts";

// utils
import { chartUtil, type IEChartsYAxisOptions } from "@/utils/chart.util";
import { type IRangeDates } from "@/utils/date.util";
import { type IGenericMetricWidgetConfigWithCompositeResponse } from "@/models/chart.model";

// composables
import { useGenericMetricWidgetOptions } from "@/composables/use-generic-metric-widget-options.composable";

// utils - column spacing
import { calculatePointInterval, millisecondsToMinutes, getColumnSpacingConfig } from "./histogram-column-spacing.util";

export type THistogramMap = Record<string, Array<MeasurementResponseValuesInner>>;

export default defineComponent({
  name: "histogram-metric-widget",
  components: { VChart, ChartWidgetWrapper },
  emits: ["register", "unregister", "on-zoom", "on-reset-zoom"],
  props: {
    config: {
      type: Object as PropType<IGenericMetricWidgetConfigWithCompositeResponse>,
      required: true,
    },
    filterByDates: {
      type: Object as PropType<IRangeDates>,
      required: true,
    },
    getDataArgs: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
  },
  data() {
    return {
      wrapperOptions: {
        title: this.config.chartTitle,
      } as IWidgetWrapperOptions,
    };
  },
  setup() {
    const {
      chartOptions,
      error,
      initialLoading,
      timeframes,
      initialize,
      loadChartOptions,
      loadData,
      createLineSeriesFromMap,
      updateYAxisLabelsDefaultsIfNeeded,
      handlePropsChanged,
      handleBeforeUnmount,
    } = useGenericMetricWidgetOptions();

    return {
      chartOptions,
      error,
      initialLoading,
      timeframes,
      initialize,
      loadChartOptions,
      loadData,
      createLineSeriesFromMap,
      updateYAxisLabelsDefaultsIfNeeded,
      handlePropsChanged,
      handleBeforeUnmount,
    };
  },
  created() {
    this.initialize(this.config);
    this.loadChartOptions(
      (event: IZoomEvent) => this.$emit("on-zoom", event),
      (chartId: string) => this.$emit("on-reset-zoom", chartId),
    );
    this.onPropsChanged();
  },
  computed: {
    wrapperProps(): IGenericWidgetWrapperConfig {
      return {
        bodyNoPadding: true,
        hideBorder: true,
        flat: true,
        hideActions: true,
        hideSeparator: true,
        ...this.config.wrapperProps,
      };
    },
  },
  methods: {
    async loadChartDataMethod(): Promise<void> {
      await this.loadData(
        this.filterByDates,
        this.getDataArgs,
        () => this.$emit("register"),
        (metricsResponse: MetricsCompositeResponse) => {
          this.setSeries(metricsResponse, this.timeframes);
          const series = this.chartOptions?.series;
          const visibleSeries = Array.isArray(series)
            ? series.filter((item) => !chartUtil.isConnectedTooltipAnchorSeries(item as Record<string, unknown>))
            : series;
          const isEmpty = !visibleSeries || (Array.isArray(visibleSeries) && visibleSeries.length === 0);
          this.error = isEmpty;
          return isEmpty;
        },
      );
    },
    updateYAxisLabels(measurements: MeasurementResponse[]): void {
      if (!this.config.yAxisLabelFormatting || !this.chartOptions) return;

      this.chartOptions.yAxis = this.updateYAxisLabelsDefaultsIfNeeded(
        measurements,
        this.chartOptions.yAxis as IEChartsYAxisOptions,
        this.config.yAxisLabelFormatting,
      );
    },
    setSeries(metricsResponse: MetricsCompositeResponse, timeframes: Array<string>): void {
      if (!this.chartOptions) return;

      const chartType = this.config.chartOptions.type;
      let measurements: MeasurementResponse[] | undefined;

      switch (chartType) {
        case "bar":
          measurements = this.handleColumnChart(metricsResponse, timeframes);
          break;
        case "line":
          measurements = this.handleLineChart(metricsResponse, timeframes);
          break;
        default:
          throw new Error(`Unsupported chart type: ${chartType}`);
      }

      if (measurements) {
        this.updateYAxisLabels(measurements);
      }
    },
    handleColumnChart(
      metricsResponse: MetricsCompositeResponse,
      timeframes: Array<string>,
    ): MeasurementResponse[] | undefined {
      if (!this.config.columnHistogramParser) return;

      const seriesData = this.config.columnHistogramParser(metricsResponse);
      this.setColumnSeries(seriesData, timeframes);

      const measurements: MeasurementResponse[] = seriesData.map((series) => {
        return {
          type: series.name,
          values: series.data,
        };
      });

      return measurements;
    },
    handleLineChart(
      metricsResponse: MetricsCompositeResponse,
      timeframes: Array<string>,
    ): MeasurementResponse[] | undefined {
      if (!this.config.histogramParser) return;

      const histogramMap: THistogramMap = this.mapMeasurements(metricsResponse);
      this.createLineSeriesFromMap(histogramMap, timeframes);

      const measurements: MeasurementResponse[] = Object.keys(histogramMap).map((type: string) => {
        return {
          type,
          values: histogramMap[type],
        };
      });

      return measurements;
    },
    setColumnSeries(seriesData: IColumnHistogramDataParser[], timeframes: Array<string>): void {
      if (!this.config.columnHistogramParser || !this.chartOptions) return;

      const pointRange = calculatePointInterval(timeframes);
      const intervalInMinutes = pointRange ? millisecondsToMinutes(pointRange) : 1;
      const { maxPointWidth, pointPadding, groupPadding } = getColumnSpacingConfig(intervalInMinutes);

      const series = this.createColumnChartSeries(seriesData, timeframes, {
        maxPointWidth,
        pointPadding,
        groupPadding,
      });
      this.chartOptions.series = this.config.useConnectedTooltipAnchor
        ? [chartUtil.getConnectedTooltipAnchorSeries(timeframes), ...series]
        : series;
      this.updateColumnYAxisScale(series);
      this.updateColumnLegendOptions(seriesData);
    },
    createColumnChartSeries(
      seriesData: IColumnHistogramDataParser[],
      timeframes: Array<string>,
      spacing: { maxPointWidth: number; pointPadding: number; groupPadding: number },
    ): BarSeriesOption[] {
      return seriesData.map((currSeries) => {
        const data = chartUtil.arrangeMeasurementsByTime(timeframes, currSeries.data, {
          useTimeFrameTimestamps: this.config.useCanonicalTimeBuckets,
        });
        return {
          name: currSeries.name,
          data: data,
          type: "bar" as const,
          stack: "requests",
          barMaxWidth: spacing.maxPointWidth,
          barGap: `${Math.round(spacing.pointPadding * 100)}%`,
          barCategoryGap: `${Math.round(spacing.groupPadding * 100)}%`,
          ...chartUtil.getBarSeriesLegendHoverState(),
          itemStyle: { borderWidth: 0 },
        };
      });
    },
    updateColumnLegendOptions(seriesData: IColumnHistogramDataParser[]): void {
      if (!this.chartOptions) return;

      this.chartOptions.legend = {
        ...(this.chartOptions.legend as Record<string, unknown>),
        bottom: 0,
        icon: "circle",
        itemWidth: 12,
        itemHeight: 12,
        itemGap: 18,
        width: "80%",
        textStyle: {
          ...((this.chartOptions.legend as Record<string, unknown>)?.textStyle as Record<string, unknown> | undefined),
          fontWeight: 700,
          color: "#333333",
          ...(this.config.legendTextStyle || {}),
        },
        data: seriesData.map((series) => ({ name: series.name, icon: "circle" })),
      };
    },
    updateColumnYAxisScale(series: BarSeriesOption[]): void {
      if (!this.chartOptions) return;

      const stackedTotals = new Map<number | string, number>();

      const individualMax = series.reduce<number>((maxValue, currentSeries) => {
        const seriesData = (Array.isArray(currentSeries.data) ? currentSeries.data : []) as unknown[];
        const seriesMax = seriesData.reduce<number>((currentMax, point) => {
          const value = Array.isArray(point) ? Number(point[1]) : Number(point);
          const timestamp = Array.isArray(point) ? String(point[0]) : "";
          if (Array.isArray(point) && Number.isFinite(value)) {
            stackedTotals.set(timestamp, (stackedTotals.get(timestamp) || 0) + value);
          }
          return Number.isFinite(value) && value > currentMax ? value : currentMax;
        }, 0);
        return seriesMax > maxValue ? seriesMax : maxValue;
      }, 0);
      const stackedMax = Math.max(0, ...Array.from(stackedTotals.values()));
      const dataMax = stackedMax || individualMax;

      if (dataMax <= 0) {
        this.chartOptions.yAxis = {
          ...(this.chartOptions.yAxis as Record<string, unknown>),
          min: 0,
          max: 1,
          interval: 2,
        };
        return;
      }

      const rawInterval = dataMax / 3;
      const magnitude = Math.pow(10, Math.floor(Math.log10(rawInterval)));
      const normalized = rawInterval / magnitude;
      const multiplier = [1, 2.5, 5, 10].find((candidate) => normalized <= candidate) || 10;
      const interval = multiplier * magnitude;

      this.chartOptions.yAxis = {
        ...(this.chartOptions.yAxis as Record<string, unknown>),
        min: 0,
        max: Math.ceil(dataMax / interval) * interval,
        interval,
      };
    },
    handleDataZoom(event: Parameters<typeof chartUtil.getDataZoomRange>[0]): void {
      const range = chartUtil.getDataZoomRange(event, this.timeframes);
      if (range) {
        this.$emit("on-zoom", { ...range, chartId: this.config.chartId } as IZoomEvent);
      }
    },
    handleResetZoom(): void {
      chartUtil.resetChartDataZoom(this.config.chartId);
      this.$emit("on-reset-zoom", this.config.chartId);
    },
    mapMeasurements(metricsResponse: MetricsCompositeResponse): Record<string, Array<MeasurementResponseValuesInner>> {
      if (!this.config.histogramParser) {
        throw new Error("Histogram parser is required");
      }
      return this.config.histogramParser(metricsResponse);
    },
    async onPropsChanged(): Promise<void> {
      await this.handlePropsChanged(() => this.$emit("unregister"), this.loadChartDataMethod);
    },
  },
  watch: {
    getDataArgs: {
      handler(): void {
        this.onPropsChanged();
      },
      deep: true,
    },
    filterByDates: {
      handler(): void {
        this.onPropsChanged();
      },
      deep: true,
    },
  },
  beforeUnmount() {
    this.handleBeforeUnmount(() => this.$emit("unregister"));
  },
});
</script>
