<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="gpu-memory-and-compute-widget">
    <gpu-compute-utilization-widget
      ref="gpuComputeUtilizationWidgetRef"
      :filter-by-dates="filterByDates"
      :workload-id="workloadId"
      :selected-element-id="selectedElementId"
      :reload-trigger="reloadTrigger"
      :silent-date-reload="silentDateReload"
      :zoom-active="zoomActive"
      @register="registerChart(gpuComputeUtilizationChart, 'gpuComputeUtilizationWidgetRef')"
      @unregister="unregisterChart(gpuComputeUtilizationChart)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
      @on-zoom-toggle="handleZoomToggle"
    />
    <gpu-memory-usage-widget
      ref="gpuMemoryUsageWidgetRef"
      :filter-by-dates="filterByDates"
      :workload-id="workloadId"
      :selected-element-id="selectedElementId"
      :reload-trigger="reloadTrigger"
      :silent-date-reload="silentDateReload"
      :zoom-active="zoomActive"
      @register="registerChart(gpuMemoryUsageChart, 'gpuMemoryUsageWidgetRef')"
      @unregister="unregisterChart(gpuMemoryUsageChart)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
      @on-zoom-toggle="handleZoomToggle"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// charts
import {
  GpuComputeUtilizationWidget,
  GPU_COMPUTE_UTILIZATION_WIDGET_CHART_ID,
} from "@/components/dashboard-v2/widgets/chart-widgets/gpu-compute-utilization-widget";
import {
  GpuMemoryUsageWidget,
  GPU_MEMORY_USAGE_CHART_ID,
} from "@/components/dashboard-v2/widgets/chart-widgets/gpu-memory-usage-widget";

// utils
import type { IRangeDates } from "@/utils/date.util";

// models
import {
  useConnectedCharts,
  type IConnectedChartsOptions,
} from "@/components/dashboard-v2/composables/use-connected-charts.composable";
import type { IZoomEvent } from "@/models/chart.model";

export default defineComponent({
  components: {
    GpuComputeUtilizationWidget,
    GpuMemoryUsageWidget,
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
    reloadTrigger: {
      type: Number,
      default: 0,
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
      gpuComputeUtilizationChart: GPU_COMPUTE_UTILIZATION_WIDGET_CHART_ID as string,
      gpuMemoryUsageChart: GPU_MEMORY_USAGE_CHART_ID as string,
      zoomActive: false,
    };
  },
  setup() {
    const options: IConnectedChartsOptions = {
      chartIds: [GPU_COMPUTE_UTILIZATION_WIDGET_CHART_ID, GPU_MEMORY_USAGE_CHART_ID],
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
