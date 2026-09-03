<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="nim-workload-widgets">
    <requests-concurrency-by-status-widget
      :workload-id="workloadId"
      :filter-by-dates="filterByDates"
      :pod-id="podId"
      @register="registerChartToCrosshair(requestsConcurrencyByStatusWidget)"
      @unregister="unregisterChartFromCrosshair(requestsConcurrencyByStatusWidget)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
    />

    <throughput-by-status-widget
      :workload-id="workloadId"
      :filter-by-dates="filterByDates"
      :pod-id="podId"
      @register="registerChartToCrosshair(throughputByStatusWidget)"
      @unregister="unregisterChartFromCrosshair(throughputByStatusWidget)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
    />

    <time-to-first-token-widget
      :workload-id="workloadId"
      :filter-by-dates="filterByDates"
      :pod-id="podId"
      @register="registerChartToCrosshair(timeToFirstTokenWidget)"
      @unregister="unregisterChartFromCrosshair(timeToFirstTokenWidget)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
    />

    <time-to-first-token-by-percentiles-widget
      :workload-id="workloadId"
      :filter-by-dates="filterByDates"
      :pod-id="podId"
      @register="registerChartToCrosshair(timeToFirstTokenByPercentilesWidget)"
      @unregister="unregisterChartFromCrosshair(timeToFirstTokenByPercentilesWidget)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
    />

    <end-to-end-request-latency-widget
      :workload-id="workloadId"
      :filter-by-dates="filterByDates"
      :pod-id="podId"
      @register="registerChartToCrosshair(endToEndRequestLatencyWidget)"
      @unregister="unregisterChartFromCrosshair(endToEndRequestLatencyWidget)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
    />

    <end-to-end-latency-by-percentiles-widget
      :workload-id="workloadId"
      :filter-by-dates="filterByDates"
      :pod-id="podId"
      @register="registerChartToCrosshair(endToEndLatencyByPercentilesWidget)"
      @unregister="unregisterChartFromCrosshair(endToEndLatencyByPercentilesWidget)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
    />

    <gpu-kv-cache-utilization-widget
      :workload-id="workloadId"
      :filter-by-dates="filterByDates"
      :pod-id="podId"
      @register="registerChartToCrosshair(gpuKvCacheUtilizationWidget)"
      @unregister="unregisterChartFromCrosshair(gpuKvCacheUtilizationWidget)"
      @on-zoom="handleZoomIn"
      @on-reset-zoom="handleResetZoom"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { ThroughputByStatusWidget } from "@/components/dashboard-v2/widgets/nim-widgets/throughput-by-status-widget";
import { RequestsConcurrencyByStatusWidget } from "@/components/dashboard-v2/widgets/nim-widgets/requests-concurrency-by-status-widget";
import { GpuKvCacheUtilizationWidget } from "@/components/dashboard-v2/widgets/nim-widgets/gpu-kv-cache-utilization-widget";
import { EndToEndLatencyByPercentilesWidget } from "@/components/dashboard-v2/widgets/nim-widgets/end-to-end-latency-by-percentiles-widget";
import { TimeToFirstTokenByPercentilesWidget } from "@/components/dashboard-v2/widgets/nim-widgets/time-to-first-token-by-percentiles-widget";
import { TimeToFirstTokenWidget } from "@/components/dashboard-v2/widgets/nim-widgets/time-to-first-token-widget";
import { EndToEndRequestLatencyWidget } from "@/components/dashboard-v2/widgets/nim-widgets/end-to-end-request-latency-widget";

// types
import type { Workload } from "@/swagger-models/workloads-client";
import type { IRangeDates } from "@/utils/date.util";
import type { IZoomEvent } from "@/models/chart.model";

// composables
import {
  type IConnectedChartsOptions,
  useConnectedCharts,
} from "@/components/dashboard-v2/composables/use-connected-charts.composable";

// consts
import { THROUGHPUT_BY_STATUS_WIDGET_CHART_ID } from "@/components/dashboard-v2/widgets/nim-widgets/throughput-by-status-widget/throughput-by-status-widget-config";
import { REQUESTS_CONCURRENCY_BY_STATUS_WIDGET_CHART_ID } from "@/components/dashboard-v2/widgets/nim-widgets/requests-concurrency-by-status-widget/requests-concurrency-by-status-widget-config";
import { GPU_KV_CACHE_UTILIZATION_WIDGET_CHART_ID } from "@/components/dashboard-v2/widgets/nim-widgets/gpu-kv-cache-utilization-widget/gpu-kv-cache-utilization-widget-config";
import { END_TO_END_LATENCY_BY_PERCENTILES_WIDGET_CHART_ID } from "@/components/dashboard-v2/widgets/nim-widgets/end-to-end-latency-by-percentiles-widget/end-to-end-latency-by-percentiles-widget-config";
import { TIME_TO_FIRST_TOKEN_BY_PERCENTILES_WIDGET_CHART_ID } from "@/components/dashboard-v2/widgets/nim-widgets/time-to-first-token-by-percentiles-widget/time-to-first-token-by-percentiles-widget-config";
import { TIME_TO_FIRST_TOKEN_WIDGET_CHART_ID } from "@/components/dashboard-v2/widgets/nim-widgets/time-to-first-token-widget/time-to-first-token-widget-config";
import { END_TO_END_REQUEST_LATENCY_WIDGET_CHART_ID } from "@/components/dashboard-v2/widgets/nim-widgets/end-to-end-request-latency-widget/end-to-end-request-latency-widget-config";

export default defineComponent({
  name: "nim-workload-widgets",
  components: {
    ThroughputByStatusWidget,
    RequestsConcurrencyByStatusWidget,
    GpuKvCacheUtilizationWidget,
    EndToEndLatencyByPercentilesWidget,
    TimeToFirstTokenByPercentilesWidget,
    TimeToFirstTokenWidget,
    EndToEndRequestLatencyWidget,
  },
  props: {
    workload: {
      type: Object as PropType<Workload>,
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
  data() {
    return {
      throughputByStatusWidget: THROUGHPUT_BY_STATUS_WIDGET_CHART_ID as string,
      requestsConcurrencyByStatusWidget: REQUESTS_CONCURRENCY_BY_STATUS_WIDGET_CHART_ID as string,
      gpuKvCacheUtilizationWidget: GPU_KV_CACHE_UTILIZATION_WIDGET_CHART_ID as string,
      endToEndLatencyByPercentilesWidget: END_TO_END_LATENCY_BY_PERCENTILES_WIDGET_CHART_ID as string,
      timeToFirstTokenByPercentilesWidget: TIME_TO_FIRST_TOKEN_BY_PERCENTILES_WIDGET_CHART_ID as string,
      timeToFirstTokenWidget: TIME_TO_FIRST_TOKEN_WIDGET_CHART_ID as string,
      endToEndRequestLatencyWidget: END_TO_END_REQUEST_LATENCY_WIDGET_CHART_ID as string,
    };
  },
  setup() {
    const options: IConnectedChartsOptions = {
      chartIds: [
        THROUGHPUT_BY_STATUS_WIDGET_CHART_ID,
        REQUESTS_CONCURRENCY_BY_STATUS_WIDGET_CHART_ID,
        GPU_KV_CACHE_UTILIZATION_WIDGET_CHART_ID,
        END_TO_END_LATENCY_BY_PERCENTILES_WIDGET_CHART_ID,
        TIME_TO_FIRST_TOKEN_BY_PERCENTILES_WIDGET_CHART_ID,
        TIME_TO_FIRST_TOKEN_WIDGET_CHART_ID,
        END_TO_END_REQUEST_LATENCY_WIDGET_CHART_ID,
      ],
    };

    const { zoomIn, resetZoom, resizeConnectedCharts, registerChartToCrosshair, unregisterChartFromCrosshair } =
      useConnectedCharts(options);
    return {
      zoomIn,
      resetZoom,
      resizeConnectedCharts,
      registerChartToCrosshair,
      unregisterChartFromCrosshair,
    };
  },
  computed: {
    workloadId(): string {
      return this.workload.id;
    },
  },
  methods: {
    resizeCharts(): void {
      this.resizeConnectedCharts();
    },
    handleZoomIn(event: IZoomEvent): void {
      this.zoomIn(event.min, event.max, event.chartId);
    },
    handleResetZoom(currentChartId: string): void {
      this.resetZoom(currentChartId);
    },
  },
});
</script>
