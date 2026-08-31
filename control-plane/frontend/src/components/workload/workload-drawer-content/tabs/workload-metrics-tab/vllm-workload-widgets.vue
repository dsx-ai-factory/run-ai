<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="vllm-workload-widgets">
    <vllm-metric-widget
      v-for="widget in linearWidgets"
      :key="widget.id"
      :workload-config="widget.workloadConfig"
      :pod-config="widget.podConfig"
      :workload-id="workload.id"
      :pod-id="podId"
      :filter-by-dates="filterByDates"
      @register="registerChartToCrosshair(widget.id)"
      @unregister="unregisterChartFromCrosshair(widget.id)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
    />
    <vllm-histogram-widget
      v-for="widget in histogramWidgets"
      :key="widget.id"
      :workload-config="widget.workloadConfig"
      :pod-config="widget.podConfig"
      :workload-id="workload.id"
      :pod-id="podId"
      :filter-by-dates="filterByDates"
      @register="registerChartToCrosshair(widget.id)"
      @unregister="unregisterChartFromCrosshair(widget.id)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
    />
  </section>
</template>

<script setup lang="ts">
import type { Workload } from "@/swagger-models/workloads-client";
import type { IRangeDates } from "@/utils/date.util";
import type { IZoomEvent } from "@/models/chart.model";
import {
  type IConnectedChartsOptions,
  useConnectedCharts,
} from "@/components/dashboard-v2/composables/use-connected-charts.composable";
import VllmMetricWidget from "@/components/dashboard-v2/widgets/vllm-widgets/vllm-metric-widget.vue";
import VllmHistogramWidget from "@/components/dashboard-v2/widgets/vllm-widgets/vllm-histogram-widget.vue";
import {
  vllmHistogramWidgets,
  vllmLinearWidgets,
} from "@/components/dashboard-v2/widgets/vllm-widgets/vllm-widget-configs";

defineProps<{
  workload: Workload;
  podId?: string;
  filterByDates: IRangeDates;
}>();

const linearWidgets = vllmLinearWidgets;
const histogramWidgets = vllmHistogramWidgets;
const options: IConnectedChartsOptions = {
  chartIds: [...linearWidgets, ...histogramWidgets].map(({ id }) => id),
};
const { resizeConnectedCharts, zoomIn, resetZoom, registerChartToCrosshair, unregisterChartFromCrosshair } =
  useConnectedCharts(options);

function resizeCharts(): void {
  resizeConnectedCharts();
}

function handleZoomIn(event: IZoomEvent): void {
  zoomIn(event.min, event.max, event.chartId);
}

function handleResetZoom(chartId: string): void {
  resetZoom(chartId);
}

defineExpose({ resizeCharts });
</script>
