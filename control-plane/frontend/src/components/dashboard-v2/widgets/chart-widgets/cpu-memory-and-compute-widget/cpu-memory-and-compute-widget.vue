<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="cpu-memory-and-compute-widget">
    <cpu-usage-widget
      ref="cpuUsageWidgetRef"
      :filter-by-dates="filterByDates"
      :workload-id="workloadId"
      :selected-element-id="selectedElementId"
      :zoom-active="zoomActive"
      @register="registerChart(cpuUsageChart, 'cpuUsageWidgetRef')"
      @unregister="unregisterChart(cpuUsageChart)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
      @on-zoom-toggle="handleZoomToggle"
      :silent-date-reload="silentDateReload"
    />
    <cpu-memory-usage-widget
      ref="cpuMemoryUsageWidgetRef"
      :filter-by-dates="filterByDates"
      :workload-id="workloadId"
      :selected-element-id="selectedElementId"
      :zoom-active="zoomActive"
      @register="registerChart(cpuMemoryUsageChart, 'cpuMemoryUsageWidgetRef')"
      @unregister="unregisterChart(cpuMemoryUsageChart)"
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
  CpuUsageWidget,
  CPU_USAGE_WIDGET_CHART_ID,
} from "@/components/dashboard-v2/widgets/chart-widgets/cpu-usage-widget";
import {
  CpuMemoryUsageWidget,
  CPU_MEMORY_USAGE_WIDGET_CHART_ID,
} from "@/components/dashboard-v2/widgets/chart-widgets/cpu-memory-usage-widget";

// utils
import type { IRangeDates } from "@/utils/date.util";

import type { IZoomEvent } from "@/models/chart.model";
import {
  useConnectedCharts,
  type IConnectedChartsOptions,
} from "@/components/dashboard-v2/composables/use-connected-charts.composable";

export default defineComponent({
  components: {
    CpuUsageWidget,
    CpuMemoryUsageWidget,
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
    selectedElementId: {
      type: String as PropType<string | null>,
      default: null,
    },
  },
  data() {
    return {
      cpuUsageChart: CPU_USAGE_WIDGET_CHART_ID as string,
      cpuMemoryUsageChart: CPU_MEMORY_USAGE_WIDGET_CHART_ID as string,
      zoomActive: false,
    };
  },
  setup() {
    const options: IConnectedChartsOptions = { chartIds: [CPU_USAGE_WIDGET_CHART_ID, CPU_MEMORY_USAGE_WIDGET_CHART_ID] };
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
