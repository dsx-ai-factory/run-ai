<template>
  <generic-metric-widget
    ref="genericMetricWidget"
    :config="widgetConfig"
    :filter-by-dates="filterByDates"
    :get-data-args="getDataArgs"
    @register="handleRegister"
    @unregister="$emit('unregister')"
    @on-zoom="$emit('on-zoom', $event)"
    @on-reset-zoom="$emit('on-reset-zoom', $event)"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import GenericMetricWidget from "@/components/dashboard-v2/widgets/chart-widgets/generic-metric-widget/generic-metric-widget.vue";
import {
  workloadThroughputByStatusWidgetConfig,
  podThroughputByStatusWidgetConfig,
} from "./throughput-by-status-widget-config";
import type { IGenericMetricWidgetConfig } from "@/models/chart.model";
import type { IRangeDates } from "@/utils/date.util";
import { chartUtil } from "@/utils/chart.util";
import { shouldHideFinalNimXAxisLabel } from "@/components/dashboard-v2/widgets/nim-widgets/nim-chart-legend.util";
import type { EChartsOption } from "echarts";
import type { LineSeriesOption } from "echarts/charts";

type TGenericMetricWidgetInstance = InstanceType<typeof GenericMetricWidget> & { chartOptions?: EChartsOption };
const AREA_SERIES_NAMES = ["Succeeded", "Failed"];

export default defineComponent({
  name: "throughput-by-status-widget",
  components: {
    GenericMetricWidget,
  },
  emits: ["register", "unregister", "on-zoom", "on-reset-zoom"],
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
  },
  computed: {
    widgetConfig(): IGenericMetricWidgetConfig {
      return this.podId ? podThroughputByStatusWidgetConfig : workloadThroughputByStatusWidgetConfig;
    },
    getDataArgs(): string[] {
      const args: string[] = [this.workloadId];
      if (this.podId) {
        args.push(this.podId);
      }

      return args;
    },
  },
  methods: {
    handleRegister(): void {
      this.applyChartParityOptions();
      this.$emit("register");
    },
    applyChartParityOptions(): void {
      const widget = this.$refs.genericMetricWidget as TGenericMetricWidgetInstance | undefined;
      const chartOptions = widget?.chartOptions;
      if (!widget || !chartOptions) return;

      const series = Array.isArray(chartOptions.series)
        ? (chartOptions.series as LineSeriesOption[]).map((currSeries) => {
            const name = String(currSeries.name || "").trim();
            if (!AREA_SERIES_NAMES.includes(name)) return currSeries;
            return {
              ...currSeries,
              areaStyle: {
                ...(currSeries.areaStyle as Record<string, unknown> | undefined),
                opacity: 0.2,
              },
            };
          })
        : chartOptions.series;

      widget.chartOptions = {
        ...chartOptions,
        legend: this.getLegendOptions(chartOptions),
        xAxis: this.getXAxisOptions(chartOptions),
        yAxis: this.getYAxisOptions(chartOptions),
        series,
      };
    },
    getLegendOptions(chartOptions: EChartsOption): EChartsOption["legend"] {
      const legend = Array.isArray(chartOptions.legend) ? chartOptions.legend[0] : chartOptions.legend;
      const prevTextStyle = (legend as Record<string, unknown>)?.textStyle as Record<string, unknown> | undefined;
      return {
        ...(legend as Record<string, unknown>),
        bottom: 16,
        textStyle: { ...prevTextStyle, fontSize: 12, fontWeight: "bold" },
      } as EChartsOption["legend"];
    },
    getXAxisOptions(chartOptions: EChartsOption): EChartsOption["xAxis"] {
      const xAxis = Array.isArray(chartOptions.xAxis) ? chartOptions.xAxis[0] : chartOptions.xAxis;
      return {
        ...(xAxis as Record<string, unknown>),
        axisLabel: {
          ...((xAxis as Record<string, unknown> | undefined)?.axisLabel as Record<string, unknown> | undefined),
          formatter: this.formatXAxisTimestamp,
        },
      } as EChartsOption["xAxis"];
    },
    getYAxisOptions(chartOptions: EChartsOption): EChartsOption["yAxis"] {
      const yAxis = Array.isArray(chartOptions.yAxis) ? chartOptions.yAxis[0] : chartOptions.yAxis;

      return {
        ...(yAxis as Record<string, unknown>),
        axisLabel: {
          ...((yAxis as Record<string, unknown> | undefined)?.axisLabel as Record<string, unknown> | undefined),
          formatter: this.formatYAxisValue,
        },
      } as EChartsOption["yAxis"];
    },
    formatXAxisTimestamp(value: number | string): string {
      if (shouldHideFinalNimXAxisLabel(value, this.filterByDates)) return "";
      return chartUtil.formatDatetimeAxisLabel(value);
    },
    formatYAxisValue(value: number | string): string {
      const numericValue = Number(value);
      if (!Number.isFinite(numericValue)) return String(value);
      if (Math.abs(numericValue) < 1000) return String(numericValue);

      const valueInThousands = numericValue / 1000;
      return `${Number.isInteger(valueInThousands) ? valueInThousands : Number(valueInThousands.toFixed(1))}k`;
    },
  },
});
</script>
