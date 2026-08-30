<template>
  <generic-metric-widget
    ref="metricWidget"
    class="workload-by-category-chart"
    :config="widgetConfig"
    :filter-by-dates="filterByDates"
    :get-data-args="getDataArgs"
    show-area-zoom
    :zoom-active="zoomActive"
    @register="handleRegister"
    @unregister="$emit('unregister')"
    @on-zoom="$emit('on-zoom', $event)"
    @on-reset-zoom="$emit('on-reset-zoom', $event)"
    @on-zoom-toggle="$emit('on-zoom-toggle', $event)"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import GenericMetricWidget from "@/components/dashboard-v2/widgets/chart-widgets/generic-metric-widget/generic-metric-widget.vue";
import { workloadByCategoryWidgetConfig } from "./workload-by-category-time-range-widget-config";
import {
  CHART_LEGEND_TEXT_STYLE_DEFAULTS,
  LEGEND_RECT_ITEM_STYLE,
  type IGenericMetricWidgetConfig,
} from "@/models/chart.model";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import type { IFilterModel } from "@/models/filter.model";
import type { IOrgUnitMetricSource } from "@/models/org-unit.model";
import type { EChartsOption, SeriesOption } from "echarts";
import { DATE_FORMATS } from "@/constants/date-formats";
import { chartUtil } from "@/utils/chart.util";
import {
  getOverviewShortRangeAxisInterval,
  ONE_MINUTE_IN_MS,
} from "@/components/dashboard-v2/widgets/time-range-widgets/overview-time-axis.util";

type GenericMetricWidgetInstance = InstanceType<typeof GenericMetricWidget> & {
  chartOptions?: EChartsOption | null;
};

function getWorkloadYAxisScale(series: SeriesOption[] = []): { max: number; interval: number } {
  const dataMax = series.reduce<number>((maxValue, currentSeries) => {
    const data = Array.isArray((currentSeries as { data?: unknown[] }).data)
      ? (currentSeries as { data: unknown[] }).data
      : [];
    const seriesMax = data.reduce<number>((currentMax, point) => {
      const value = Array.isArray(point) ? point[1] : null;
      return typeof value === "number" && value > currentMax ? value : currentMax;
    }, 0);
    return seriesMax > maxValue ? seriesMax : maxValue;
  }, 0);

  if (dataMax <= 0) {
    return { max: 1, interval: 1 };
  }

  const rawInterval = dataMax / 2;
  const magnitude = Math.pow(10, Math.floor(Math.log10(rawInterval)));
  const normalized = rawInterval / magnitude;
  let multiplier = 10;

  if (normalized <= 1) {
    multiplier = 1;
  } else if (normalized <= 2) {
    multiplier = 2;
  } else if (normalized <= 5) {
    multiplier = 5;
  }

  const interval = Math.max(1, multiplier * magnitude);
  const max = Math.max(interval, Math.ceil(dataMax / interval) * interval);

  return {
    interval,
    max,
  };
}

function formatDatetimeAxisLabel(value: number | string): string {
  const timestamp = typeof value === "number" ? value : new Date(value).getTime();
  if (!Number.isFinite(timestamp)) return "";

  return dateUtil.dateFormat(new Date(timestamp), DATE_FORMATS.STANDARD_DATE_TIME).replace(", ", "\n");
}

export default defineComponent({
  name: "workload-by-category-time-range-widget",
  components: {
    GenericMetricWidget,
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
    orgUnitMetricSource: {
      type: Object as PropType<IOrgUnitMetricSource | null>,
      default: null,
    },
    zoomActive: {
      type: Boolean as PropType<boolean | undefined>,
      default: undefined,
    },
  },
  computed: {
    widgetConfig(): IGenericMetricWidgetConfig {
      return workloadByCategoryWidgetConfig;
    },
    getDataArgs(): string[] {
      const args: string[] = [this.clusterId, this.nodePoolName ?? ""];
      if (this.orgUnitMetricSource) {
        args.push(this.orgUnitMetricSource.type, this.orgUnitMetricSource.id);
      }
      return args;
    },
  },
  methods: {
    handleRegister(): void {
      this.applyHighchartsParityOptions();
      this.$emit("register");
    },
    applyHighchartsParityOptions(): void {
      const metricWidget = this.$refs.metricWidget as GenericMetricWidgetInstance | undefined;
      const chartOptions = metricWidget?.chartOptions;
      if (!chartOptions) return;

      const allSeries = Array.isArray(chartOptions.series) ? (chartOptions.series as SeriesOption[]) : [];
      const series = allSeries.filter((item) => !chartUtil.isConnectedTooltipAnchorSeries(item));
      const yAxisScale = getWorkloadYAxisScale(series);

      chartOptions.yAxis = {
        ...(chartOptions.yAxis as Record<string, unknown>),
        ...yAxisScale,
        nameGap: 28,
        nameTextStyle: {
          color: "#666666",
        },
      };

      chartOptions.grid = {
        ...(chartOptions.grid as Record<string, unknown>),
        top: 30,
      };

      chartOptions.toolbox = chartUtil.getDashboardChartToolbox({ showZoomToolbox: true });

      const shortRangeInterval = getOverviewShortRangeAxisInterval(this.filterByDates.label);
      const last15MinuteAxisOptions =
        shortRangeInterval === ONE_MINUTE_IN_MS
          ? {
              interval: ONE_MINUTE_IN_MS,
              minInterval: ONE_MINUTE_IN_MS,
              maxInterval: ONE_MINUTE_IN_MS,
            }
          : {
              maxInterval: workloadByCategoryWidgetConfig.chartOptions.xAxisIntervalMs,
            };

      const legendOptions = chartOptions.legend as Record<string, unknown>;

      chartOptions.legend = {
        ...legendOptions,
        left: "center",
        ...LEGEND_RECT_ITEM_STYLE,
        itemGap: 20,
        data: series.map((currentSeries) => (currentSeries as { name?: string }).name ?? ""),
        textStyle: {
          ...(legendOptions.textStyle as Record<string, unknown>),
          ...CHART_LEGEND_TEXT_STYLE_DEFAULTS,
        },
      };

      chartOptions.xAxis = {
        ...(chartOptions.xAxis as Record<string, unknown>),
        min: this.filterByDates.dateStart.getTime(),
        max: this.filterByDates.dateEnd.getTime(),
        ...last15MinuteAxisOptions,
        axisLabel: {
          ...((chartOptions.xAxis as Record<string, unknown>).axisLabel as Record<string, unknown>),
          formatter: formatDatetimeAxisLabel,
          hideOverlap: true,
          showMaxLabel: true,
        },
      };
    },
  },
});
</script>

<style scoped lang="scss">
.workload-by-category-chart,
.workload-by-category-chart :deep(.q-card),
.workload-by-category-chart :deep(.generic-metric-chart-shell),
.workload-by-category-chart :deep(.generic-metric-chart),
.workload-by-category-chart :deep(svg) {
  background: var(--bg-primary-default);
}
</style>
