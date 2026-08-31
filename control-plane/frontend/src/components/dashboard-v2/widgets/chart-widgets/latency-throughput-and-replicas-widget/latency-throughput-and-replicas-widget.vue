<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="latency-throughput-and-replicas-widget">
    <throughput-widget
      :workload-id="workloadId"
      :pod-id="podId"
      :filter-by-dates="filterByDates"
      :zoom-active="zoomActive"
      @register="registerChart(throughputWidget)"
      @unregister="unregisterChart(throughputWidget)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
      @on-zoom-toggle="handleZoomToggle"
      silent-date-reload
    />
    <latency-widget
      :workload-id="workloadId"
      :pod-id="podId"
      :filter-by-dates="filterByDates"
      :zoom-active="zoomActive"
      @register="registerChart(latencyWidget)"
      @unregister="unregisterChart(latencyWidget)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
      @on-zoom-toggle="handleZoomToggle"
      silent-date-reload
    />
    <replicas-widget
      v-if="!podId"
      :workload-id="workloadId"
      :filter-by-dates="filterByDates"
      :zoom-active="zoomActive"
      @register="registerChart(replicasWidget)"
      @unregister="unregisterChart(replicasWidget)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
      @on-zoom-toggle="handleZoomToggle"
      silent-date-reload
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// charts
import { LatencyWidget, LATENCY_WIDGET_CHART_ID } from "@/components/dashboard-v2/widgets/chart-widgets/latency-widget";
import {
  ThroughputWidget,
  THROUGHPUT_WIDGET_CHART_ID,
} from "@/components/dashboard-v2/widgets/chart-widgets/throughput-widget";

import {
  ReplicasWidget,
  REPLICAS_WIDGET_CHART_ID,
} from "@/components/dashboard-v2/widgets/chart-widgets/replicas-widget";

// stores
import { useClusterStore } from "@/stores/cluster.store";

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
    LatencyWidget,
    ThroughputWidget,
    ReplicasWidget,
  },
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
    clusterVersion: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      throughputWidget: THROUGHPUT_WIDGET_CHART_ID as string,
      latencyWidget: LATENCY_WIDGET_CHART_ID as string,
      replicasWidget: REPLICAS_WIDGET_CHART_ID as string,
      zoomActive: false,
    };
  },
  setup() {
    const options: IConnectedChartsOptions = {
      chartIds: [THROUGHPUT_WIDGET_CHART_ID, LATENCY_WIDGET_CHART_ID, REPLICAS_WIDGET_CHART_ID],
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
