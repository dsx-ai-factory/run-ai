<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="dcgm-nvlink-bandwidth-total-bytes-section-widget">
    <dcgm-nvlink-bandwidth-total-bytes-widget
      ref="dcgmNvlinkBandwidthTotalBytesWidget"
      :filter-by-dates="filterByDates"
      :workload-id="workloadId"
      :zoom-active="zoomActive"
      @register="registerChart(dcgmNvlinkBandwidthTotalBytesChart)"
      @unregister="unregisterChart(dcgmNvlinkBandwidthTotalBytesChart)"
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
  DcgmNvlinkBandwidthTotalBytesWidget,
  DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_CHART_ID,
} from "@/components/dashboard-v2/widgets/chart-widgets/dcgm-nvlink-bandwidth-total-bytes-widget";

// utils
import type { IRangeDates } from "@/utils/date.util";

import type { IZoomEvent } from "@/models/chart.model";
import {
  type IConnectedChartsOptions,
  useConnectedCharts,
} from "@/components/dashboard-v2/composables/use-connected-charts.composable";

export default defineComponent({
  components: {
    DcgmNvlinkBandwidthTotalBytesWidget,
  },
  props: {
    workloadId: {
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
      dcgmNvlinkBandwidthTotalBytesChart: DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_CHART_ID as string,
      zoomActive: false,
    };
  },
  setup() {
    const options: IConnectedChartsOptions = {
      chartIds: [DCGM_NVLINK_BANDWIDTH_TOTAL_BYTES_CHART_ID],
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
    registerChart(chartId: string): void {
      this.registerChartToCrosshair(chartId);
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
  },
});
</script>
<style lang="scss" scoped></style>
