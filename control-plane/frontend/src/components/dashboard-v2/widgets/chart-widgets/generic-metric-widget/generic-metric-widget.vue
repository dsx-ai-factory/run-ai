<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <chart-widget-wrapper
    class="generic-metric-widget"
    :options="wrapperOptions"
    :body-no-padding="wrapperProps.bodyNoPadding"
    :loading="initialLoading"
    :error="error"
    :empty="empty"
    :hide-border="wrapperProps.hideBorder"
    :flat="wrapperProps.flat"
    :hide-actions="wrapperProps.hideActions"
    :hide-separator="wrapperProps.hideSeparator"
  >
    <div class="generic-metric-chart-shell">
      <chart-area-zoom-controls
        v-if="showAreaZoom"
        :external-active="zoomActive"
        :vertical-offset="areaZoomVerticalOffset"
        @zoom-toggle="handleZoomToggle"
        @reset="handleAreaZoomReset"
      />
      <chart-zoom-controls v-show="false" ref="zoomControls" @zoom-change="handleZoomControls" />
      <v-chart
        v-if="chartOptions"
        ref="chartRef"
        class="generic-metric-chart"
        :option="chartOptions"
        :id="config.chartId"
        :style="{
          height: config.chartOptions.height + 'px',
          visibility: adaptiveLegendReady ? 'visible' : 'hidden',
        }"
        autoresize
        @datazoom="handleDataZoom"
        @restore="handleResetZoom"
        @finished="syncAdaptiveLegendLayout"
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
import { ChartZoomControls } from "@/components/common/charts/chart-zoom-controls";
import { ChartAreaZoomControls } from "@/components/common/charts/chart-area-zoom-controls";

// models
import type { IWidgetWrapperOptions, IZoomEvent, IGenericWidgetWrapperConfig } from "@/models/chart.model";
import {
  type MetricsResponse,
  type MeasurementResponseValuesInner,
  type MeasurementResponse,
} from "@/swagger-models/workloads-client";

// utils
import { metricUtil } from "@/utils/metric.util";
import { type IRangeDates } from "@/utils/date.util";
import { type IGenericMetricWidgetConfig } from "@/models/chart.model";
import { chartUtil } from "@/utils/chart.util";
import type { IEChartsYAxisOptions } from "@/utils/chart.util";

// composables
import { useGenericMetricWidgetOptions } from "@/composables/use-generic-metric-widget-options.composable";

const LEGEND_BOTTOM_GAP = 16;
const LEGEND_TO_AXIS_GAP = 12;
const MINIMUM_PLOT_HEIGHT = 96;

type LegendLayoutChart = {
  getHeight: () => number;
  getModel: () => { getComponent: (type: "legend" | "xAxis", index: number) => unknown };
  getViewOfComponentModel: (model: unknown) => { group: { getBoundingRect: () => { height: number } } } | undefined;
  setOption: (option: Record<string, unknown>, options?: { lazyUpdate?: boolean }) => void;
};

const GENERIC_WRAPPER_DEFAULT_PROPS: IGenericWidgetWrapperConfig = {
  hideBorder: true,
  flat: true,
  hideActions: true,
  hideSeparator: true,
  bodyNoPadding: true,
};
export default defineComponent({
  components: { VChart, ChartWidgetWrapper, ChartZoomControls, ChartAreaZoomControls },
  emits: ["register", "unregister", "on-zoom", "on-reset-zoom", "on-zoom-toggle"],
  props: {
    config: {
      type: Object as PropType<IGenericMetricWidgetConfig>,
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
    showAreaZoom: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    areaZoomVerticalOffset: {
      type: Number as PropType<number>,
      default: 0,
    },
    zoomActive: {
      type: Boolean as PropType<boolean | undefined>,
      default: undefined,
    },
    confineTooltip: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      wrapperOptions: {
        title: this.config.chartTitle,
        ...this.config.wrapperProps?.options,
      } as IWidgetWrapperOptions,
      adaptiveLegendBottom: null as number | null,
      adaptiveXAxisHeight: null as number | null,
      adaptiveLegendFrame: null as number | null,
      adaptiveLegendReady: !this.config.chartOptions.adaptiveLegendLayout,
    };
  },
  setup() {
    const {
      chartOptions,
      error,
      empty,
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
      empty,
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
    if (this.confineTooltip && this.chartOptions?.tooltip) {
      this.chartOptions.tooltip = {
        ...(this.chartOptions.tooltip as Record<string, unknown>),
        appendToBody: false,
        confine: true,
      };
    }
    this.onPropsChanged();
  },
  computed: {
    wrapperProps(): IGenericWidgetWrapperConfig {
      return (
        this.config.wrapperProps || {
          ...GENERIC_WRAPPER_DEFAULT_PROPS,
        }
      );
    },
  },
  methods: {
    async loadChartDataMethod(): Promise<void> {
      await this.loadData(
        this.filterByDates,
        this.getDataArgs,
        () => this.$emit("register"),
        (metricsResponse: MetricsResponse): boolean => {
          const measurementMap: Record<string, Array<MeasurementResponseValuesInner>> = this.mapMeasurements(
            metricsResponse,
          );

          const isEmpty = this.createLineSeriesFromMap(measurementMap, this.timeframes);
          this.updateYAxisLabels(metricsResponse.measurements);
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
    mapMeasurements(metricsResponse: MetricsResponse): Record<string, Array<MeasurementResponseValuesInner>> {
      if (this.config.measurementParser) {
        return this.config.measurementParser(metricsResponse, this.config.averageMetricType);
      }
      return metricUtil.mapMeasurementsToMap(metricsResponse);
    },
    handleZoomToggle(active: boolean): void {
      const vchart = this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
      if (!vchart?.chart) return;
      vchart.chart.dispatchAction({
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
    handleZoomControls(payload: { start: number; end: number }): void {
      const chartRef = this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
      if (!chartRef) return;

      chartRef.chart?.dispatchAction({
        type: "dataZoom",
        start: payload.start,
        end: payload.end,
      });
    },
    handleDataZoom(event: Parameters<typeof chartUtil.getDataZoomRange>[0]): void {
      const range = chartUtil.getDataZoomRange(event, this.timeframes);

      if (range) {
        this.$emit("on-zoom", { ...range, chartId: this.config.chartId } as IZoomEvent);
      }

      const zoomControls = this.$refs.zoomControls as InstanceType<typeof ChartZoomControls> | undefined;
      if (zoomControls && event.batch?.[0]) {
        const zoom = event.batch[0];
        if (zoom.start !== undefined && zoom.end !== undefined) {
          zoomControls.syncFromExternal(Number(zoom.start), Number(zoom.end));
        }
      }
    },
    handleResetZoom(): void {
      const chartRef = this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
      chartRef?.chart?.dispatchAction({ type: "dataZoom", start: 0, end: 100 }, { silent: true });
      this.$emit("on-reset-zoom", this.config.chartId);
      const zoomControls = this.$refs.zoomControls as InstanceType<typeof ChartZoomControls> | undefined;
      zoomControls?.resetZoom();
    },
    syncAdaptiveLegendLayout(): void {
      if (!this.config.chartOptions.adaptiveLegendLayout) return;

      if (this.adaptiveLegendFrame !== null) cancelAnimationFrame(this.adaptiveLegendFrame);
      this.adaptiveLegendFrame = requestAnimationFrame(() => {
        this.adaptiveLegendFrame = null;
        const chart = (this.$refs.chartRef as InstanceType<typeof VChart> | undefined)?.chart as
          | LegendLayoutChart
          | undefined;
        if (!chart) return;

        const legend = chart.getModel().getComponent("legend", 0);
        const legendHeight = Math.ceil(chart.getViewOfComponentModel(legend)?.group.getBoundingRect().height || 0);
        const xAxis = chart.getModel().getComponent("xAxis", 0);
        const measuredXAxisHeight = Math.ceil(chart.getViewOfComponentModel(xAxis)?.group.getBoundingRect().height || 0);
        const configuredGrid = this.config.chartOptions.grid || {};
        const baseBottom = typeof configuredGrid.bottom === "number" ? configuredGrid.bottom : 0;
        const gridTop = typeof configuredGrid.top === "number" ? configuredGrid.top : 0;
        if (this.adaptiveXAxisHeight === null && measuredXAxisHeight > 0) {
          this.adaptiveXAxisHeight = measuredXAxisHeight;
        }
        const stableXAxisHeight = this.adaptiveXAxisHeight || 0;
        const desiredBottom = Math.max(
          baseBottom,
          legendHeight + stableXAxisHeight + LEGEND_BOTTOM_GAP + LEGEND_TO_AXIS_GAP,
        );
        const maximumBottom = Math.max(baseBottom, chart.getHeight() - gridTop - MINIMUM_PLOT_HEIGHT);
        const nextBottom = Math.min(desiredBottom, maximumBottom);
        if (nextBottom === this.adaptiveLegendBottom) {
          this.adaptiveLegendReady = true;
          return;
        }

        this.adaptiveLegendBottom = nextBottom;
        chart.setOption(
          {
            grid: { bottom: nextBottom },
            legend: { bottom: LEGEND_BOTTOM_GAP, left: "center", width: "90%" },
          },
          { lazyUpdate: true },
        );
        requestAnimationFrame(() => {
          this.adaptiveLegendReady = true;
        });
      });
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
    if (this.adaptiveLegendFrame !== null) cancelAnimationFrame(this.adaptiveLegendFrame);
    this.handleBeforeUnmount(() => this.$emit("unregister"));
  },
});
</script>

<style scoped lang="scss">
.generic-metric-widget,
.generic-metric-chart-shell,
.generic-metric-chart {
  width: 100%;
  max-width: 100%;
  min-width: 0;
}

.generic-metric-chart-shell {
  overflow: hidden;
}

.generic-metric-chart {
  display: block;
}
</style>
