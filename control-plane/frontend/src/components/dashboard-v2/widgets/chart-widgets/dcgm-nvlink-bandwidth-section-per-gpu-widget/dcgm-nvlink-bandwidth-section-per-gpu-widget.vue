<template>
  <section class="dcgm-nvlink-bandwidth-section-per-gpu-widget">
    <dcgm-nvlink-bandwidth-total-bytes-per-gpu-widget
      ref="dcgmNvlinkBandwidthPerGpuWidget"
      :filter-by-dates="filterByDates"
      :workload-id="workloadId"
      :pod-id="podId"
      :zoom-active="zoomActive"
      @register="registerChart(dcgmNvlinkBandwidthPerGpuChart, 'dcgmNvlinkBandwidthPerGpuWidget')"
      @unregister="unregisterChart(dcgmNvlinkBandwidthPerGpuChart)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
      @on-zoom-toggle="handleZoomToggle"
      :silent-date-reload="silentDateReload"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// charts
import {
  DcgmNvlinkBandwidthTotalBytesPerGpuWidget,
  DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_PER_GPU_CHART_ID,
} from "@/components/dashboard-v2/widgets/chart-widgets/dcgm-nvlink-bandwidth-total-bytes-per-gpu-widget";

// utils
import type { IRangeDates } from "@/utils/date.util";

import type { IZoomEvent } from "@/models/chart.model";
import {
  type IConnectedChartsOptions,
  useConnectedCharts,
} from "@/components/dashboard-v2/composables/use-connected-charts.composable";

export default defineComponent({
  components: {
    DcgmNvlinkBandwidthTotalBytesPerGpuWidget,
  },
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
    silentDateReload: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      dcgmNvlinkBandwidthPerGpuChart: DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_PER_GPU_CHART_ID as string,
      zoomActive: false,
    };
  },
  setup() {
    const options: IConnectedChartsOptions = {
      chartIds: [DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_PER_GPU_CHART_ID],
    };
    const {
      zoomIn,
      resetZoom,
      setZoomMode,
      resizeConnectedCharts,
      registerChartToCrosshair,
      unregisterChartFromCrosshair,
    } = useConnectedCharts(options);
    return {
      zoomIn,
      resetZoom,
      setZoomMode,
      resizeConnectedCharts,
      registerChartToCrosshair,
      unregisterChartFromCrosshair,
    };
  },
  methods: {
    registerChart(chartId: string, chartRefName: string): void {
      this.$nextTick(() => {
        if (this.setChartElementId(chartId, chartRefName)) {
          this.registerChartToCrosshair(chartId);
          return;
        }

        requestAnimationFrame(() => {
          this.setChartElementId(chartId, chartRefName);
          this.registerChartToCrosshair(chartId);
        });
      });
    },
    unregisterChart(chartId: string): void {
      this.unregisterChartFromCrosshair(chartId);
    },
    resizeCharts(): void {
      this.resizeConnectedCharts();
    },
    handleZoomIn(event: IZoomEvent): void {
      this.zoomIn(event.min, event.max, event.chartId);
    },
    handleResetZoom(currentChartId: string): void {
      this.resetZoom(currentChartId);
      this.zoomActive = false;
      this.setZoomMode(false);
    },
    handleZoomToggle(active: boolean): void {
      this.zoomActive = active;
      this.setZoomMode(active);
    },
    setChartElementId(chartId: string, chartRefName: string): boolean {
      const ref = this.$refs[chartRefName];
      const chartComponent = Array.isArray(ref) ? ref[0] : ref;
      const rootElement =
        chartComponent instanceof HTMLElement
          ? chartComponent
          : (chartComponent as { $el?: HTMLElement } | undefined)?.$el;
      const chartElement = rootElement?.querySelector(".echarts") as HTMLElement | null;

      if (!chartElement) return false;

      chartElement.id = chartId;
      return true;
    },
  },
});
</script>
<style lang="scss" scoped></style>
