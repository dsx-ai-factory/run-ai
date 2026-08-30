<template>
  <chart-widget-wrapper
    class="resource-utilization-time-range-widget"
    :options="wrapperOptions"
    body-no-padding
    :loading="isInitialLoader"
    :error="error"
    @export-csv="exportCsv"
    :hide-border="hideBorder"
    :hide-time-frame="hideTimeFrame"
    :hide-actions="hideActions"
    :hide-separator="hideSeparator"
    :flat="flat"
    slot-wrapper-class=""
    size="large"
    :announcement="legendAnnouncement"
  >
    <template v-if="hideAverageUtilizationStats" #header="{ titleId }">
      <div class="resource-utilization-drawer-header row items-center q-py-xs q-pl-md">
        <div class="title-tooltip-container">
          <h3 :id="titleId" class="chart-widget-title text-b-main">Resources utilization</h3>
          <runai-tooltip-wrapper use-html :tooltip-text="wrapperOptions.tooltipText ?? ''">
            <template #tooltip-content>
              <div v-dompurify-html="wrapperOptions.tooltipText" />
            </template>
            <q-icon class="resource-utilization-help-icon text-icon-default" :name="helpIcon" />
          </runai-tooltip-wrapper>
        </div>
      </div>
    </template>
    <div class="q-pt-sm chart-container">
      <div v-if="!hideAverageUtilizationStats" class="avg-utilization-wrapper">
        <average-utilization-summary :avg-data="averageMetrics" />
      </div>
      <div class="time-range-chart-container">
        <chart-zoom-controls v-show="false" ref="zoomControls" @zoom-change="handleZoomControls" />
        <chart-area-zoom-controls
          :external-active="zoomActive"
          @zoom-toggle="handleZoomToggle"
          @reset="handleAreaZoomReset"
        />
        <v-chart
          class="resource-utilization-time-range-chart"
          v-if="chartOptions"
          :id="chartId"
          :option="chartOptions"
          autoresize
          ref="chart"
          @datazoom="handleDataZoom"
          @restore="handleResetZoom"
          @zr:click="handleChartAreaClick"
        />
        <div
          v-if="resourceUtilizationLegendItems.length"
          class="resource-utilization-legend"
          :class="{ 'resource-utilization-legend--compact': hideAverageUtilizationStats }"
        >
          <button
            v-for="item in resourceUtilizationLegendItems"
            :key="item.name"
            class="resource-utilization-legend__item"
            :class="{ 'resource-utilization-legend__item--muted': !isLegendSelected(item.name) }"
            type="button"
            :aria-pressed="isLegendSelected(item.name)"
            @click="toggleLegendItem(item.name)"
            @mouseenter="highlightLegendItem(item.name)"
            @mouseleave="downplayLegendItem(item.name)"
            @focus="highlightLegendItem(item.name)"
            @blur="downplayLegendItem(item.name)"
          >
            <span class="resource-utilization-legend__glyph" :style="{ '--legend-color': item.color }" />
            <span class="resource-utilization-legend__label">{{ item.name }}</span>
          </button>
        </div>
      </div>
    </div>
  </chart-widget-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";

//models
import {
  CHART_AXIS_LABEL_DEFAULTS,
  CHART_COLOR_PALETTE,
  INSIDE_DATAZOOM_CONFIG,
  type IWidgetWrapperOptions,
  type ITooltipPoint,
  type TMeasurementsTimeAndValue,
  type IZoomEvent,
  RESOURCE_UTILIZATION_TIME_RANGE_WIDGET_CHART_ID,
} from "@/models/chart.model";
import {
  MetricsType,
  type MetricsResponse,
  type GroupedMeasurementResponseValuesInner,
} from "@/swagger-models/cluster-service-client";
import { OrgUnitMetricType } from "@/swagger-models/org-unit-service-client";

//cmps
import VChart from "vue-echarts";
import "@/plugins/echarts";
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { ChartZoomControls } from "@/components/common/charts/chart-zoom-controls";
import { ChartAreaZoomControls } from "@/components/common/charts/chart-area-zoom-controls";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";

import type { EChartsOption } from "echarts";
import type { LineSeriesOption } from "echarts/charts";

// services
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

//utils
import { chartUtil, STANDARD_TOOLTIP_STYLE } from "@/utils/chart.util";
import { metricUtil } from "@/utils/metric.util";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";
import { roundToDecimal } from "@/utils/format.util";
import { dashboardUtil } from "@/utils/dashboard.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import type { IFilterModel } from "@/models/filter.model";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";
import { AverageUtilizationSummary } from "@/components/dashboard-v2/widgets/time-range-widgets/resource-utilization-time-range-widget/average-utilization-summary";
import {
  getOverviewShortRangeAxisInterval,
  TEN_MINUTES_IN_MS,
} from "@/components/dashboard-v2/widgets/time-range-widgets/overview-time-axis.util";

const RESOURCE_UTILIZATION_LEGEND_NAMES = [
  "GPU compute utilization",
  "GPU memory utilization",
  "CPU compute utilization",
  "CPU memory utilization",
];
const RESOURCE_UTILIZATION_COLORS = [
  CHART_COLOR_PALETTE.BLUE,
  CHART_COLOR_PALETTE.ORANGE,
  CHART_COLOR_PALETTE.LIME,
  CHART_COLOR_PALETTE.PURPLE,
];
const TWO_HOURS_IN_MS = 2 * 60 * 60 * 1000;

interface IResourceUtilizationLegendItem {
  name: string;
  color: string;
}

export default defineComponent({
  name: "resource-utilization-time-range-widget",
  components: {
    AverageUtilizationSummary,
    ChartWidgetWrapper,
    VChart,
    ChartZoomControls,
    ChartAreaZoomControls,
    RunaiTooltipWrapper,
  },
  emits: ["register", "unregister", "on-zoom", "on-reset-zoom", "on-zoom-toggle"],
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    nodePoolName: {
      type: String as PropType<string>,
      required: false,
    },
    filterByDates: {
      type: Object as PropType<IRangeDates>,
      required: true,
    },
    filterBy: {
      type: Array as PropType<IFilterModel[]>,
      required: false,
    },
    confineTooltip: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideAverageUtilizationStats: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideBorder: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    flat: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideActions: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideTimeFrame: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideSeparator: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    orgUnitMetricSource: {
      type: Object as PropType<IOrgUnitMetricSource | null>,
      default: null,
    },
    zoomActive: {
      type: Boolean as PropType<boolean | undefined>,
      default: undefined,
    },
  },
  data() {
    return {
      helpIcon: toQuasarIconName(runaiIcons.object.help),
      wrapperOptions: {
        title: "Resources utilization",
        tooltipText: "GPU and CPU utilization over time.<br/> Click and drag to view a shorter timeframe.",
      } as IWidgetWrapperOptions,
      chartId: RESOURCE_UTILIZATION_TIME_RANGE_WIDGET_CHART_ID as string,
      chartOptions: null as null | EChartsOption,
      averageMetrics: [] as Array<{ name: string; value: string }>,
      measurementMap: {} as Record<string, Array<GroupedMeasurementResponseValuesInner>>,
      error: false as boolean,
      isInitialLoader: true as boolean,
      loading: false as boolean,
      timeframes: [] as Array<string>,
      legendSelected: {} as Record<string, boolean>,
      legendAnnouncement: "",
    };
  },
  async created() {
    this.loadChartOptions();
    this.loadChartData();
    this.startRefreshLoadData();
  },
  computed: {
    resourceUtilizationLegendItems(): Array<IResourceUtilizationLegendItem> {
      return RESOURCE_UTILIZATION_LEGEND_NAMES.map((name, index) => ({
        name,
        color: RESOURCE_UTILIZATION_COLORS[index],
      }));
    },
  },
  methods: {
    startRefreshLoadData(): void {
      intervalUtil.startInterval(EIntervalLabels.ResourceUtilizationTimeRangeWidget, () => this.loadChartData(true));
    },
    stopRefreshLoadData(): void {
      intervalUtil.stopInterval(EIntervalLabels.ResourceUtilizationTimeRangeWidget);
    },
    formatXAxisLabel(value: number, hiddenMin?: number): string {
      const timestamp = Number(value);
      if (hiddenMin !== undefined && Math.abs(timestamp - hiddenMin) < 1) return "";

      return dateUtil.dateFormat(new Date(timestamp), DATE_FORMATS.STANDARD_DATE_TIME).replace(", ", "\n");
    },
    getXAxisOptions(endDate: Date): EChartsOption["xAxis"] {
      const duration = endDate.getTime() - this.filterByDates.dateStart.getTime();
      const isShortRange = duration <= TWO_HOURS_IN_MS;
      const interval = getOverviewShortRangeAxisInterval(this.filterByDates.label);
      const hiddenMin = isShortRange
        ? Math.floor(this.filterByDates.dateStart.getTime() / interval) * interval
        : undefined;
      const intervalOptions = isShortRange
        ? {
            interval,
            min: hiddenMin,
          }
        : {};

      return {
        type: "time",
        splitNumber: 14,
        minInterval: isShortRange ? interval : TEN_MINUTES_IN_MS,
        min: this.filterByDates.dateStart.getTime(),
        max: endDate.getTime(),
        ...intervalOptions,
        axisLabel: {
          formatter: (value: number) => this.formatXAxisLabel(value, hiddenMin),
          hideOverlap: true,
          ...CHART_AXIS_LABEL_DEFAULTS,
        },
      };
    },
    loadChartOptions(): void {
      const endDate = this.getEndDate();

      this.chartOptions = {
        color: RESOURCE_UTILIZATION_COLORS,
        textStyle: { fontFamily: "Roboto, sans-serif" },
        grid: { top: 30, right: 12, bottom: 72, left: 8, containLabel: true },
        xAxis: this.getXAxisOptions(endDate),
        yAxis: {
          type: "value",
          min: 0,
          max: 100,
          interval: 25,
          axisLabel: { formatter: "{value}%", ...CHART_AXIS_LABEL_DEFAULTS },
        },
        tooltip: {
          trigger: "axis",
          ...STANDARD_TOOLTIP_STYLE,
          ...(this.confineTooltip ? { appendToBody: false, confine: true } : {}),
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
            const points: Array<ITooltipPoint> = visibleItems.map((item) => ({
              name: item.seriesName,
              y: item.value[1] ? roundToDecimal(item.value[1]) : item.value[1],
              color: item.color,
              symbol: "%",
            }));
            return chartUtil.formatSharedTooltip(points, (visibleItems[0] || items[0]).value[0]) as string;
          },
        },
        legend: {
          show: false,
          data: RESOURCE_UTILIZATION_LEGEND_NAMES,
        },
        dataZoom: [{ ...INSIDE_DATAZOOM_CONFIG }],
        toolbox: chartUtil.getDashboardChartToolbox({ showZoomToolbox: true }),
        series: [],
      };
    },
    async loadChartData(refresh = false): Promise<void> {
      if (this.chartOptions === null) return;
      const endDate = this.getEndDate();
      try {
        this.loading = true;
        this.chartOptions.xAxis = this.getXAxisOptions(endDate);
        this.timeframes = chartUtil.getTimeFrames(this.filterByDates.dateStart, endDate);

        const metricsResponse: MetricsResponse = await this.getMetric();
        this.measurementMap = metricUtil.mapMeasurementsToMap(metricsResponse);
        this.error = false;
        this.setSeries(refresh);
        this.setAverageMetrics();
      } catch (error: unknown) {
        this.error = true;
        console.error(error);
      } finally {
        this.loading = false;
        this.isInitialLoader = false;
      }

      this.$nextTick(() => {
        if (!this.error) {
          this.$emit("register");
        }
      });
    },
    getEndDate(): Date {
      return this.filterByDates.dateEnd;
    },
    async exportCsv(): Promise<void> {
      if (this.orgUnitMetricSource) {
        this.$q.notify(dashboardUtil.getCsvErrorMessage());
        return;
      }
      const metricType: Array<MetricsType> = [
        MetricsType.GpuUtilization,
        MetricsType.GpuMemoryUtilization,
        MetricsType.CpuUtilization,
        MetricsType.CpuMemoryUtilization,
      ];

      try {
        const samples = this.timeframes.length;

        await clusterService.getClusterMetricsCsv(
          this.clusterId,
          dateUtil.convertDateToISO(this.filterByDates.dateStart),
          dateUtil.convertDateToISO(this.filterByDates.dateEnd),
          metricType,
          samples,
        );
      } catch (error: unknown) {
        this.$q.notify(dashboardUtil.getCsvErrorMessage());
        console.error(error);
      }
    },
    setSeries(refresh = false): void {
      if (this.chartOptions === null) return;

      const gpuUtilizationData: Array<TMeasurementsTimeAndValue> = this.getMeasurements(
        this.timeframes,
        this.measurementMap[MetricsType.GpuUtilization],
      );
      const gpuMemoryUtilizationData: Array<TMeasurementsTimeAndValue> = this.getMeasurements(
        this.timeframes,
        this.measurementMap[MetricsType.GpuMemoryUtilization],
      );
      const cpuUtilizationData: Array<TMeasurementsTimeAndValue> = this.getMeasurements(
        this.timeframes,
        this.measurementMap[MetricsType.CpuUtilization],
      );
      const cpuMemoryUtilizationData: Array<TMeasurementsTimeAndValue> = this.getMeasurements(
        this.timeframes,
        this.measurementMap[MetricsType.CpuMemoryUtilization],
      );
      const tooltipAnchorSeries = chartUtil.getConnectedTooltipAnchorSeries(this.timeframes);
      const existingSeries = Array.isArray(this.chartOptions.series)
        ? (this.chartOptions.series as LineSeriesOption[])
        : [];
      if (refresh && existingSeries[0]?.id === tooltipAnchorSeries.id) {
        const series = existingSeries;
        series[0].data = tooltipAnchorSeries.data;
        series[1].data = gpuUtilizationData;
        series[2].data = gpuMemoryUtilizationData;
        series[3].data = cpuUtilizationData;
        series[4].data = cpuMemoryUtilizationData;
      } else {
        this.chartOptions.series = [
          tooltipAnchorSeries,
          {
            name: "GPU compute utilization",
            type: "line",
            data: gpuUtilizationData,
            color: RESOURCE_UTILIZATION_COLORS[0],
            showSymbol: false,
            ...chartUtil.getConnectedChartHoverMarker(RESOURCE_UTILIZATION_COLORS[0]),
          },
          {
            name: "GPU memory utilization",
            type: "line",
            data: gpuMemoryUtilizationData,
            color: RESOURCE_UTILIZATION_COLORS[1],
            showSymbol: false,
            ...chartUtil.getConnectedChartHoverMarker(RESOURCE_UTILIZATION_COLORS[1]),
          },
          {
            name: "CPU compute utilization",
            type: "line",
            data: cpuUtilizationData,
            color: RESOURCE_UTILIZATION_COLORS[2],
            showSymbol: false,
            ...chartUtil.getConnectedChartHoverMarker(RESOURCE_UTILIZATION_COLORS[2]),
          },
          {
            name: "CPU memory utilization",
            type: "line",
            data: cpuMemoryUtilizationData,
            color: RESOURCE_UTILIZATION_COLORS[3],
            showSymbol: false,
            ...chartUtil.getConnectedChartHoverMarker(RESOURCE_UTILIZATION_COLORS[3]),
          },
        ] as LineSeriesOption[];
      }
    },
    getMeasurements(
      timeframes: string[],
      measurements: Array<GroupedMeasurementResponseValuesInner>,
    ): Array<TMeasurementsTimeAndValue> {
      return chartUtil.arrangeMeasurementsByTime(timeframes, measurements, { useTimeFrameTimestamps: true });
    },
    getChartRef(): InstanceType<typeof VChart> | undefined {
      return this.$refs.chart as InstanceType<typeof VChart> | undefined;
    },
    isLegendSelected(name: string): boolean {
      return this.legendSelected[name] !== false;
    },
    toggleLegendItem(name: string): void {
      const selected = !this.isLegendSelected(name);
      this.legendSelected = {
        ...this.legendSelected,
        [name]: selected,
      };
      this.getChartRef()?.chart?.dispatchAction({ type: "legendToggleSelect", name });
      this.legendAnnouncement = `${name} ${selected ? "shown" : "hidden"}.`;
    },
    highlightLegendItem(name: string): void {
      this.getChartRef()?.chart?.dispatchAction({ type: "highlight", seriesName: name });
    },
    downplayLegendItem(name: string): void {
      this.getChartRef()?.chart?.dispatchAction({ type: "downplay", seriesName: name });
    },
    handleZoomToggle(active: boolean): void {
      const chart = this.$refs.chart as InstanceType<typeof VChart> | undefined;
      if (!chart?.chart) return;
      chart.chart.dispatchAction({
        type: "takeGlobalCursor",
        key: "dataZoomSelect",
        dataZoomSelectActive: active,
      });
      this.$emit("on-zoom-toggle", active);
    },
    handleAreaZoomReset(): void {
      const chart = this.$refs.chart as InstanceType<typeof VChart> | undefined;
      if (!chart?.chart) return;
      chartUtil.resetChartDataZoomInstance(chart.chart);
      this.$emit("on-zoom-toggle", false);
    },
    handleZoomControls(payload: { start: number; end: number }): void {
      const chartRef = this.$refs.chart as InstanceType<typeof VChart> | undefined;
      if (!chartRef) return;
      chartRef.chart?.dispatchAction({ type: "dataZoom", start: payload.start, end: payload.end });
    },
    handleDataZoom(event: Parameters<typeof chartUtil.getDataZoomRange>[0]): void {
      const range = chartUtil.getDataZoomRange(event, this.timeframes);

      if (range) {
        this.setAverageMetrics(range.min, range.max);
        this.$emit("on-zoom", { ...range, chartId: this.chartId } as IZoomEvent);
      }
    },
    handleChartAreaClick(event: { target?: unknown }): void {
      if (event.target) return;
      this.handleResetZoom();
    },
    handleResetZoom(): void {
      this.getChartRef()?.chart?.dispatchAction({ type: "dataZoom", start: 0, end: 100 }, { silent: true });
      const zoomControls = this.$refs.zoomControls as InstanceType<typeof ChartZoomControls> | undefined;
      zoomControls?.resetZoom();
      this.setAverageMetrics();
      this.$emit("on-reset-zoom", this.chartId);
    },
    setAverageMetrics(min?: number, max?: number): void {
      this.averageMetrics = [
        {
          name: "Avg. GPU memory",
          value: this.getAverageMetric(MetricsType.GpuMemoryUtilization, min, max),
        },
        {
          name: "Avg. GPU compute",
          value: this.getAverageMetric(MetricsType.GpuUtilization, min, max),
        },
        {
          name: "Avg. CPU compute",
          value: this.getAverageMetric(MetricsType.CpuUtilization, min, max),
        },
        {
          name: "Avg. CPU memory",
          value: this.getAverageMetric(MetricsType.CpuMemoryUtilization, min, max),
        },
      ];
    },
    calculateAverage(numbers: number[]): number {
      if (numbers.length === 0) return 0;

      const sum = numbers.reduce((acc: number, num: number) => acc + num, 0);
      return sum / numbers.length;
    },
    formatPercentage(num: number): string {
      return `${num.toFixed(2)}%`;
    },
    getAverageMetric(type: MetricsType, min?: number, max?: number): string {
      let measurements: Array<GroupedMeasurementResponseValuesInner> = this.measurementMap[type]
        ? this.measurementMap[type]
        : [];

      if (min !== undefined && max !== undefined) {
        measurements = measurements.filter((measurement: GroupedMeasurementResponseValuesInner) => {
          if (!measurement.timestamp) return false;

          const currentTimestamp = new Date(measurement.timestamp).getTime();
          return currentTimestamp >= min && currentTimestamp <= max;
        });
      }

      const measurementsValue: number[] = measurements.map((measurement: GroupedMeasurementResponseValuesInner) =>
        Number(measurement.value),
      );

      return this.formatPercentage(this.calculateAverage(measurementsValue));
    },
    async getMetric(): Promise<MetricsResponse> {
      const endDate = this.getEndDate();
      const samples = this.timeframes.length;
      const start = dateUtil.convertDateToISO(this.filterByDates.dateStart);
      const end = dateUtil.convertDateToISO(endDate);

      if (this.orgUnitMetricSource) {
        const orgUnitMetricTypes: Array<OrgUnitMetricType> = [
          OrgUnitMetricType.GpuUtilization,
          OrgUnitMetricType.GpuMemoryUtilization,
          OrgUnitMetricType.CpuUtilization,
          OrgUnitMetricType.CpuMemoryUtilization,
        ];
        const fetcher =
          this.orgUnitMetricSource.type === EOrgUnitMetricSourceType.Department
            ? orgUnitService.getDepartmentMetrics
            : orgUnitService.getProjectMetrics;
        return (await fetcher(
          this.orgUnitMetricSource.id,
          orgUnitMetricTypes,
          start,
          end,
          samples,
          this.nodePoolName,
        )) as unknown as MetricsResponse;
      }

      const metricType: Array<MetricsType> = [
        MetricsType.GpuUtilization,
        MetricsType.GpuMemoryUtilization,
        MetricsType.CpuUtilization,
        MetricsType.CpuMemoryUtilization,
      ];
      if (this.nodePoolName) {
        return await clusterService.getNodepoolMetrics(
          this.clusterId,
          this.nodePoolName,
          start,
          end,
          metricType,
          samples,
        );
      } else {
        return await clusterService.getClusterMetrics(this.clusterId, start, end, metricType, samples);
      }
    },
    async onPropsChanged(): Promise<void> {
      this.$emit("unregister");
      this.isInitialLoader = true;
      await this.loadChartData(true);
    },
  },
  watch: {
    filterByDates: {
      handler(): void {
        this.onPropsChanged();
      },
      deep: true,
    },
    nodePoolName: {
      async handler(): Promise<void> {
        this.onPropsChanged();
      },
    },
    clusterId: {
      async handler(): Promise<void> {
        this.onPropsChanged();
      },
    },
    orgUnitMetricSource: {
      async handler(): Promise<void> {
        this.onPropsChanged();
      },
      deep: true,
    },
  },
  unmounted() {
    this.$emit("unregister");
    this.stopRefreshLoadData();
  },
});
</script>

<style scoped lang="scss">
.time-range-chart-container {
  position: relative;
  width: 100%;
}

.resource-utilization-time-range-chart {
  width: 100%;
  height: 250px;
}

.resource-utilization-drawer-header {
  height: 50px;

  .title-tooltip-container {
    display: flex;
    align-items: center;
    flex-wrap: nowrap;
    margin-right: auto;
  }

  .resource-utilization-help-icon {
    margin-left: 8px;
    margin-bottom: 2px;
    font-size: 20px;
  }
}

.resource-utilization-legend {
  position: absolute;
  right: 0;
  bottom: 20px;
  left: 0;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  width: calc(100% - 16px);
  margin: 0 auto;
  gap: 1px 12px;
  pointer-events: none;
}

.resource-utilization-legend--compact {
  width: calc(100% - 16px);
}

.resource-utilization-legend__item {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  min-width: 0;
  padding: 0;
  border: 0;
  background: transparent;
  color: #333333;
  font-family: Roboto, sans-serif;
  font-size: 11px;
  font-weight: 700;
  line-height: 14px;
  white-space: nowrap;
  cursor: pointer;
  pointer-events: auto;
}

.resource-utilization-legend__item--muted {
  opacity: 0.35;
}

.resource-utilization-legend__glyph {
  flex: 0 0 auto;
  width: 16px;
  height: 7px;
  background-color: var(--legend-color);
  border-radius: 1px;
}

.resource-utilization-legend__label {
  min-width: 0;
}

.resource-utilization-time-range-widget {
  .chart-widget-title {
    margin: 0;
    font: inherit;
    letter-spacing: inherit;
  }

  overflow: hidden;

  .avg-utilization-wrapper {
    padding-left: 55px;
  }
}
</style>
