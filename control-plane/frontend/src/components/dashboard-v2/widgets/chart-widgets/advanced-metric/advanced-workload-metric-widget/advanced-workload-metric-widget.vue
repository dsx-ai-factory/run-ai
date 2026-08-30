<template>
  <section aid="advanced-workload-metric-container">
    <generic-metric-widget
      v-for="widget in podMetricAdvancedWidgets"
      :key="widget.key"
      :config="widget.config"
      :filter-by-dates="filterByDates"
      :get-data-args="[workloadId, podId]"
      :confine-tooltip="confineTooltip"
      @register="registerChart(widget.config.chartId)"
      @unregister="unregisterChart(widget.config.chartId)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
    />
  </section>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IZoomEvent } from "@/models/chart.model";
import { hoursGranularityPredefinedOptions, type IRangeDates } from "@/utils/date.util";
import { podMetricAdvancedWidgets } from "./widget-config-mapper";
import GenericMetricWidget from "@/components/dashboard-v2/widgets/chart-widgets/generic-metric-widget/generic-metric-widget.vue";
import {
  type IConnectedChartsOptions,
  useConnectedCharts,
} from "@/components/dashboard-v2/composables/use-connected-charts.composable";

export default defineComponent({
  name: "advanced-workload-metric-widget",
  components: { GenericMetricWidget },
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
    confineTooltip: {
      type: Boolean,
      default: false,
    },
  },
  setup() {
    const chartIds = podMetricAdvancedWidgets.map((widget) => widget.config.chartId);
    const options: IConnectedChartsOptions = { chartIds };
    const { registerChartToCrosshair, unregisterChartFromCrosshair, zoomIn, resetZoom, resizeConnectedCharts } =
      useConnectedCharts(options);

    const registerChart = (chartId: string): void => {
      registerChartToCrosshair(chartId);
    };

    const unregisterChart = (chartId: string): void => {
      unregisterChartFromCrosshair(chartId);
    };

    const handleZoomIn = (event: IZoomEvent): void => {
      zoomIn(event.min, event.max, event.chartId);
    };

    const handleResetZoom = (currentChartId: string): void => {
      resetZoom(currentChartId);
    };

    return {
      podMetricAdvancedWidgets,
      hoursGranularityPredefinedOptions,
      chartIds,
      registerChart,
      unregisterChart,
      handleZoomIn,
      handleResetZoom,
      resizeConnectedCharts,
    };
  },
  methods: {
    resizeCharts(): void {
      this.resizeConnectedCharts();
    },
  },
});
</script>

<style scoped lang="scss"></style>
