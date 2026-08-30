<template>
  <chart-widget-wrapper
    :options="wrapperOptions"
    body-no-padding
    :loading="isInitialLoader"
    :error="error"
    :hide-border="hideBorder"
    :flat="flat"
    :announcement="legendAnnouncement"
    @export-csv="exportCsv"
  >
    <div class="time-range-chart-container">
      <chart-area-zoom-controls
        :external-active="zoomActive"
        @zoom-toggle="handleZoomToggle"
        @reset="handleAreaZoomReset"
      />
      <v-chart
        class="resource-allocation-time-range-highchart"
        v-if="chartOptions"
        ref="chartRef"
        :id="chartId"
        :option="chartOptions"
        autoresize
        @datazoom="handleDataZoom"
        @restore="handleResetZoom"
        @zr:click="handleChartAreaClick"
      />
      <div v-if="resourceAllocationLegendItems.length" class="resource-allocation-legend">
        <button
          v-for="item in resourceAllocationLegendItems"
          :key="item.name"
          class="resource-allocation-legend__item"
          :class="{ 'resource-allocation-legend__item--muted': !isLegendSelected(item.name) }"
          type="button"
          :aria-pressed="isLegendSelected(item.name)"
          @click="toggleLegendItem(item.name)"
          @mouseenter="highlightLegendItem(item.name)"
          @mouseleave="downplayLegendItem(item.name)"
          @focus="highlightLegendItem(item.name)"
          @blur="downplayLegendItem(item.name)"
        >
          <span
            class="resource-allocation-legend__glyph"
            :class="`resource-allocation-legend__glyph--${item.marker}`"
            :style="{ '--legend-color': item.color }"
          />
          <span class="resource-allocation-legend__label">{{ item.name }}</span>
        </button>
      </div>
    </div>
  </chart-widget-wrapper>
</template>
<script lang="ts">
import "@/plugins/echarts";
import { defineComponent, type PropType } from "vue";
import VChart from "vue-echarts";

import {
  CHART_COLOR_PALETTE,
  type IWidgetWrapperOptions,
  type TMeasurementsTimeAndValue,
  type IZoomEvent,
  RESOURCE_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID,
  INSIDE_DATAZOOM_CONFIG,
} from "@/models/chart.model";
import {
  MetricsType,
  type MetricsResponse,
  type GroupedMeasurementResponseValuesInner,
} from "@/swagger-models/cluster-service-client";
import { OrgUnitMetricType } from "@/swagger-models/org-unit-service-client";

import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { ChartAreaZoomControls } from "@/components/common/charts/chart-area-zoom-controls";

import type { EChartsOption, SeriesOption } from "echarts";

import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

import { chartUtil, STANDARD_TOOLTIP_STYLE } from "@/utils/chart.util";
import { metricUtil } from "@/utils/metric.util";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";
import { roundToDecimal } from "@/utils/format.util";

import { useClusterStore } from "@/stores/cluster.store";
import { dashboardUtil } from "@/utils/dashboard.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import {
  getOverviewShortRangeAxisInterval,
  TEN_MINUTES_IN_MS,
} from "@/components/dashboard-v2/widgets/time-range-widgets/overview-time-axis.util";

export enum EEmitName {
  Register = "register",
  Unregister = "unregister",
  OnZoom = "on-zoom",
  OnResetZoom = "on-reset-zoom",
}

const DEPARTMENT_EMPTY_VALUE = "-1";
const RESOURCE_ALLOCATION_COLORS = {
  total: CHART_COLOR_PALETTE.BLUE,
  unallocated: CHART_COLOR_PALETTE.ORANGE,
  allocated: CHART_COLOR_PALETTE.LIME,
  quota: CHART_COLOR_PALETTE.PURPLE,
};
const TWO_HOURS_IN_MS = 2 * 60 * 60 * 1000;

interface IResourceAllocationLegendItem {
  name: string;
  color: string;
  marker: "dash" | "circle";
}

export default defineComponent({
  name: "resource-allocation-time-range-widget",
  components: { ChartWidgetWrapper, VChart, ChartAreaZoomControls },
  emits: {
    [EEmitName.Register]: null,
    [EEmitName.Unregister]: null,
    [EEmitName.OnZoom]: (event: IZoomEvent) => event,
    [EEmitName.OnResetZoom]: (chartId: string) => chartId,
    "on-zoom-toggle": (active: boolean) => typeof active === "boolean",
  },
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
    hideBorder: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    flat: {
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
      chartId: RESOURCE_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID as string,
      clusterStore: useClusterStore(),
      wrapperOptions: {
        title: "Resources allocation",
        tooltipText:
          "Quota/Total shows whether the quota is suitable for the physical resources available. <br/>Allocated/Total shows how many resources are being used by the departments or projects.<br/>Click and drag to view a shorter timeframe.",
      } as IWidgetWrapperOptions,
      chartOptions: null as null | EChartsOption,
      measurementMap: {} as Record<string, Array<GroupedMeasurementResponseValuesInner>>,
      error: false as boolean,
      loading: false as boolean,
      isInitialLoader: true as boolean,
      gpuQuotaTitle: "GPU quota (All departments)",
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
    resourceAllocationLegendItems(): Array<IResourceAllocationLegendItem> {
      if (this.orgUnitMetricSource) {
        return [
          { name: "GPU quota", color: RESOURCE_ALLOCATION_COLORS.quota, marker: "dash" },
          { name: "GPU allocated", color: RESOURCE_ALLOCATION_COLORS.allocated, marker: "circle" },
        ];
      }

      const items: IResourceAllocationLegendItem[] = [
        { name: "GPU devices (total)", color: RESOURCE_ALLOCATION_COLORS.total, marker: "dash" },
        { name: "GPU devices (unallocated)", color: RESOURCE_ALLOCATION_COLORS.unallocated, marker: "circle" },
        { name: "GPU devices (allocated)", color: RESOURCE_ALLOCATION_COLORS.allocated, marker: "circle" },
        { name: this.gpuQuotaTitle, color: RESOURCE_ALLOCATION_COLORS.quota, marker: "dash" },
      ];
      return items.filter((item) => item.name);
    },
  },
  methods: {
    startRefreshLoadData(): void {
      intervalUtil.startInterval(EIntervalLabels.ResourceAllocationTimeRangeWidget, this.loadChartData);
    },
    stopRefreshLoadData(): void {
      intervalUtil.stopInterval(EIntervalLabels.ResourceAllocationTimeRangeWidget);
    },
    formatXAxisLabel(value: number | string): string {
      const timestamp = typeof value === "number" ? value : new Date(value).getTime();
      if (!Number.isFinite(timestamp)) return "";

      return dateUtil.dateFormat(new Date(timestamp), DATE_FORMATS.STANDARD_DATE_TIME).replace(", ", "\n");
    },
    getXAxisOptions(): EChartsOption["xAxis"] {
      const duration = this.filterByDates.dateEnd.getTime() - this.filterByDates.dateStart.getTime();
      const isShortRange = duration <= TWO_HOURS_IN_MS;
      const interval = getOverviewShortRangeAxisInterval(this.filterByDates.label);
      const intervalOptions = isShortRange
        ? {
            interval,
            min: Math.floor(this.filterByDates.dateStart.getTime() / interval) * interval,
          }
        : {};

      return {
        type: "time",
        splitNumber: 14,
        minInterval: isShortRange ? interval : TEN_MINUTES_IN_MS,
        min: this.filterByDates.dateStart.getTime(),
        max: this.filterByDates.dateEnd.getTime(),
        ...intervalOptions,
        axisLabel: {
          fontSize: 10,
          formatter: this.formatXAxisLabel,
          hideOverlap: true,
        },
        splitLine: { show: false },
        axisTick: { show: false },
      };
    },
    loadChartOptions(): void {
      this.chartOptions = {
        color: [
          RESOURCE_ALLOCATION_COLORS.total,
          RESOURCE_ALLOCATION_COLORS.unallocated,
          RESOURCE_ALLOCATION_COLORS.allocated,
          RESOURCE_ALLOCATION_COLORS.quota,
        ],
        grid: { left: 54, right: 20, top: 36, bottom: 84 },
        xAxis: this.getXAxisOptions(),
        yAxis: {
          type: "value",
          min: 0,
          name: "GPU devices",
          nameLocation: "middle",
          nameRotate: 90,
          nameGap: 40,
          splitNumber: 4,
          axisLabel: { formatter: chartUtil.formatCompactAxisValue },
          splitLine: { lineStyle: { type: "solid", color: "#e6e6e6" } },
        },
        tooltip: {
          trigger: "axis",
          ...STANDARD_TOOLTIP_STYLE,
          axisPointer: {
            ...chartUtil.getConnectedChartAxisPointer(),
          },
          formatter: (params: unknown) => {
            const paramsArr = params as Array<{
              seriesId?: string;
              seriesName: string;
              value: [number, number | null];
              color: string;
            }>;
            if (!Array.isArray(paramsArr) || paramsArr.length === 0) return "";
            const visibleItems = paramsArr.filter((item) => !chartUtil.isConnectedTooltipAnchorSeries(item));
            const points = visibleItems.map((p) => ({
              name: p.seriesName,
              y: p.value[1] !== null ? roundToDecimal(p.value[1]) : p.value[1],
              color: p.color,
            }));
            return chartUtil.formatSharedTooltip(points, (visibleItems[0] || paramsArr[0]).value[0]);
          },
        },
        dataZoom: [{ ...INSIDE_DATAZOOM_CONFIG }],
        legend: {
          show: false,
        },
        toolbox: chartUtil.getDashboardChartToolbox({ showZoomToolbox: true }),
        series: [],
      };
    },
    calculateUnallocatedValues(
      totalGPU: Array<TMeasurementsTimeAndValue>,
      allocatedGPU: Array<TMeasurementsTimeAndValue>,
    ): Array<TMeasurementsTimeAndValue> {
      if (totalGPU.length !== allocatedGPU.length) {
        throw new Error("Arrays must have the same length.");
      }

      const differenceArray: Array<TMeasurementsTimeAndValue> = [];

      for (let i = 0; i < totalGPU.length; i++) {
        const totalValue: number | null = totalGPU[i][1];
        const allocatedValue: number | null = allocatedGPU[i][1];
        if (totalValue === null || allocatedValue === null) {
          differenceArray.push([totalGPU[i][0], null]);
        } else {
          const difference: number = totalValue - allocatedValue;
          differenceArray.push([totalGPU[i][0], difference]);
        }
      }

      return differenceArray;
    },
    async loadChartData(): Promise<void> {
      if (this.chartOptions === null) return;

      try {
        this.loading = true;
        this.chartOptions.xAxis = this.getXAxisOptions();
        this.timeframes = chartUtil.getTimeFrames(this.filterByDates.dateStart, this.filterByDates.dateEnd);
        const metricsResponse: MetricsResponse = await this.getMetric();

        this.measurementMap = metricUtil.mapMeasurementsToMap(metricsResponse);
        this.error = false;
        this.setSeries();
      } catch (error: unknown) {
        this.error = true;
        console.error(error);
      } finally {
        this.loading = false;
        this.isInitialLoader = false;
      }

      this.$nextTick(() => {
        if (!this.error) {
          this.$emit(EEmitName.Register);
        }
      });
    },
    async getMetric(): Promise<MetricsResponse> {
      const samples = this.timeframes.length;
      const start = dateUtil.convertDateToISO(this.filterByDates.dateStart);
      const end = dateUtil.convertDateToISO(this.filterByDates.dateEnd);

      if (this.orgUnitMetricSource) {
        const orgUnitMetricTypes: Array<OrgUnitMetricType> = [
          OrgUnitMetricType.GpuAllocation,
          OrgUnitMetricType.GpuQuota,
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

      const metricType: Array<MetricsType> = [MetricsType.TotalGpu, MetricsType.AllocatedGpu, MetricsType.GpuQuota];
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
    async exportCsv(): Promise<void> {
      if (this.orgUnitMetricSource) {
        this.$q.notify(dashboardUtil.getCsvErrorMessage());
        return;
      }
      const metricType: Array<MetricsType> = [MetricsType.TotalGpu, MetricsType.AllocatedGpu, MetricsType.GpuQuota];

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
    setSeries(): void {
      if (this.chartOptions === null) return;

      if (this.orgUnitMetricSource) {
        this.setOrgUnitSeries();
        return;
      }

      const totalGpus: Array<TMeasurementsTimeAndValue> = this.getMeasurements(
        this.timeframes,
        this.measurementMap[MetricsType.TotalGpu],
      );

      const allocatedGpus: Array<TMeasurementsTimeAndValue> = this.getMeasurements(
        this.timeframes,
        this.measurementMap[MetricsType.AllocatedGpu],
      );

      const gpuQuotaMeasurements = this.measurementMap[MetricsType.GpuQuota]?.filter(
        (measurement) => measurement.value !== DEPARTMENT_EMPTY_VALUE,
      );

      const gpuQuota: Array<TMeasurementsTimeAndValue> = this.getMeasurements(this.timeframes, gpuQuotaMeasurements);

      this.chartOptions = {
        ...this.chartOptions,
        legend: {
          show: false,
          data: this.resourceAllocationLegendItems.map((item) => item.name),
        },
        series: [
          chartUtil.getConnectedTooltipAnchorSeries(this.timeframes),
          {
            name: "GPU devices (total)",
            type: "line",
            data: totalGpus,
            lineStyle: chartUtil.getDottedLineStyle({ color: RESOURCE_ALLOCATION_COLORS.total }),
            showSymbol: false,
            ...chartUtil.getConnectedChartHoverMarker(RESOURCE_ALLOCATION_COLORS.total),
            itemStyle: { color: RESOURCE_ALLOCATION_COLORS.total },
          },
          {
            name: "GPU devices (unallocated)",
            type: "line",
            data: this.calculateUnallocatedValues(totalGpus, allocatedGpus),
            showSymbol: false,
            stack: "allocation",
            areaStyle: { color: RESOURCE_ALLOCATION_COLORS.unallocated, opacity: 0.2 },
            lineStyle: { color: RESOURCE_ALLOCATION_COLORS.unallocated, width: 1 },
            itemStyle: { color: RESOURCE_ALLOCATION_COLORS.unallocated },
            ...chartUtil.getConnectedChartHoverMarker(RESOURCE_ALLOCATION_COLORS.unallocated),
          },
          {
            name: "GPU devices (allocated)",
            type: "line",
            data: allocatedGpus,
            showSymbol: false,
            stack: "allocation",
            areaStyle: { color: RESOURCE_ALLOCATION_COLORS.allocated, opacity: 0.2 },
            lineStyle: { color: RESOURCE_ALLOCATION_COLORS.allocated, width: 1 },
            itemStyle: { color: RESOURCE_ALLOCATION_COLORS.allocated },
            ...chartUtil.getConnectedChartHoverMarker(RESOURCE_ALLOCATION_COLORS.allocated),
          },
          {
            name: this.gpuQuotaTitle,
            type: "line",
            data: gpuQuota,
            lineStyle: chartUtil.getDottedLineStyle({ color: RESOURCE_ALLOCATION_COLORS.quota }),
            showSymbol: false,
            itemStyle: { color: RESOURCE_ALLOCATION_COLORS.quota },
            ...chartUtil.getConnectedChartHoverMarker(RESOURCE_ALLOCATION_COLORS.quota),
          },
        ] as SeriesOption[],
      };
    },
    setOrgUnitSeries(): void {
      if (this.chartOptions === null) return;

      const gpuAllocation: Array<TMeasurementsTimeAndValue> = this.getMeasurements(
        this.timeframes,
        this.measurementMap[OrgUnitMetricType.GpuAllocation],
      );

      const gpuQuotaMeasurements = this.measurementMap[OrgUnitMetricType.GpuQuota]?.filter(
        (measurement) => measurement.value !== DEPARTMENT_EMPTY_VALUE,
      );
      const gpuQuota: Array<TMeasurementsTimeAndValue> = this.getMeasurements(this.timeframes, gpuQuotaMeasurements);

      this.chartOptions = {
        ...this.chartOptions,
        legend: {
          show: false,
          data: this.resourceAllocationLegendItems.map((item) => item.name),
        },
        series: [
          chartUtil.getConnectedTooltipAnchorSeries(this.timeframes),
          {
            name: "GPU quota",
            type: "line",
            data: gpuQuota,
            lineStyle: chartUtil.getDottedLineStyle({ color: RESOURCE_ALLOCATION_COLORS.quota }),
            showSymbol: false,
            itemStyle: { color: RESOURCE_ALLOCATION_COLORS.quota },
            ...chartUtil.getConnectedChartHoverMarker(RESOURCE_ALLOCATION_COLORS.quota),
          },
          {
            name: "GPU allocated",
            type: "line",
            data: gpuAllocation,
            showSymbol: false,
            areaStyle: { color: RESOURCE_ALLOCATION_COLORS.allocated, opacity: 0.2 },
            lineStyle: { color: RESOURCE_ALLOCATION_COLORS.allocated, width: 1 },
            itemStyle: { color: RESOURCE_ALLOCATION_COLORS.allocated },
            ...chartUtil.getConnectedChartHoverMarker(RESOURCE_ALLOCATION_COLORS.allocated),
          },
        ] as SeriesOption[],
      };
    },
    getMeasurements(
      timeframes: string[],
      measurements: Array<GroupedMeasurementResponseValuesInner>,
    ): Array<TMeasurementsTimeAndValue> {
      return chartUtil.arrangeMeasurementsByTime(timeframes, measurements, { useTimeFrameTimestamps: true });
    },
    getChartRef(): InstanceType<typeof VChart> | undefined {
      return this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
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
      const chart = this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
      if (!chart?.chart) return;
      chart.chart.dispatchAction({
        type: "takeGlobalCursor",
        key: "dataZoomSelect",
        dataZoomSelectActive: active,
      });
      this.$emit("on-zoom-toggle", active);
    },
    handleAreaZoomReset(): void {
      const chart = this.$refs.chartRef as InstanceType<typeof VChart> | undefined;
      if (!chart?.chart) return;
      chartUtil.resetChartDataZoomInstance(chart.chart);
      this.$emit("on-zoom-toggle", false);
    },
    handleDataZoom(event: Parameters<typeof chartUtil.getDataZoomRange>[0]): void {
      const range = chartUtil.getDataZoomRange(event, this.timeframes);

      if (range) {
        this.$emit(EEmitName.OnZoom, { ...range, chartId: this.chartId });
      }
    },
    handleChartAreaClick(event: { target?: unknown }): void {
      if (event.target) return;
      this.handleResetZoom();
    },
    handleResetZoom(): void {
      this.getChartRef()?.chart?.dispatchAction({ type: "dataZoom", start: 0, end: 100 }, { silent: true });
      this.$emit(EEmitName.OnResetZoom, this.chartId);
    },
    onPropsChanged(): void {
      this.$emit(EEmitName.Unregister);
      this.isInitialLoader = true;
      this.loadChartData();
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
      handler(): void {
        this.onPropsChanged();
      },
    },
    clusterId: {
      handler(): void {
        this.onPropsChanged();
      },
    },
    orgUnitMetricSource: {
      handler(): void {
        this.onPropsChanged();
      },
      deep: true,
    },
  },
  unmounted() {
    this.$emit(EEmitName.Unregister);
    this.stopRefreshLoadData();
  },
});
</script>

<style scoped lang="scss">
.time-range-chart-container {
  position: relative;
  width: 100%;
}

.resource-allocation-time-range-highchart {
  width: 100%;
  height: 250px;
}

.resource-allocation-legend {
  position: absolute;
  right: 0;
  bottom: 20px;
  left: 0;
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
  gap: 6px 34px;
  padding: 0 16px;
  pointer-events: none;
}

.resource-allocation-legend__item {
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

.resource-allocation-legend__item--muted {
  opacity: 0.35;
}

.resource-allocation-legend__glyph {
  flex: 0 0 auto;
  color: var(--legend-color);
}

.resource-allocation-legend__glyph--dash {
  // Mirrors LEGEND_DASHED_LINE_ICON: four 3px stripes separated by 2px gaps.
  width: 20px;
  height: 8px;
  background-image: repeating-linear-gradient(
    to right,
    var(--legend-color) 0,
    var(--legend-color) 3px,
    transparent 3px,
    transparent 5px
  );
  // Do not center a repeating pattern: it can clip the first dash at fractional display scales.
  background-position: 0 center;
  background-repeat: repeat-x;
  background-size: 5px 8px;
}

.resource-allocation-legend__glyph--circle {
  width: 11px;
  height: 11px;
  border-radius: 50%;
  background: var(--legend-color);
}

.resource-allocation-legend__label {
  min-width: 0;
}
</style>
