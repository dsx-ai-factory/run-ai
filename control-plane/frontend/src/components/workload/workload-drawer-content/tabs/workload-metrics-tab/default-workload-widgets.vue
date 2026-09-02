<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="default-workload-widgets">
    <template v-if="selectedPodId">
      <gpu-memory-usage-and-compute-utilization-per-gpu-widget
        :ref="gpuMemoryUsageAndComputeUtilizationPerGpuWidgetRef"
        :workload-id="workloadId"
        :pod-id="selectedPodId"
        :filter-by-dates="filterByDates"
        silent-date-reload
      />
      <cpu-memory-and-compute-usage-per-gpu-widget
        :cluster-id="workload.clusterId"
        :ref="cpuMemoryAndComputeUsagePerGpuWidgetRef"
        :workload-id="workloadId"
        :pod-id="selectedPodId"
        :filter-by-dates="filterByDates"
        silent-date-reload
      />
      <dcgm-nvlink-bandwidth-section-per-gpu-widget
        v-if="isClusterSupportsDcgmNvlinkMetrics"
        :ref="dcgmNvlinkBandwidthSectionPerGpuWidgetRef"
        :workload-id="workloadId"
        :pod-id="selectedPodId"
        :filter-by-dates="filterByDates"
        silent-date-reload
      />
    </template>
    <template v-else>
      <gpu-memory-and-compute-widget
        :ref="gpuMemoryAndComputeWidgetRef"
        :workload-id="workloadId"
        :selected-element-id="selectedElementId"
        :filter-by-dates="filterByDates"
        :reload-trigger="reloadTrigger"
        silent-date-reload
      />
      <cpu-memory-and-compute-widget
        :ref="cpuMemoryAndComputeWidgetRef"
        :workload-id="workloadId"
        :selected-element-id="selectedElementId"
        :filter-by-dates="filterByDates"
        :reload-trigger="reloadTrigger"
        silent-date-reload
      />
      <dcgm-nvlink-bandwidth-total-bytes-section-widget
        v-if="isClusterSupportsDcgmNvlinkMetrics"
        :ref="dcgmNvlinkBandwidthTotalBytesSectionWidgetRef"
        :workload-id="workloadId"
        :filter-by-dates="filterByDates"
        :reload-trigger="reloadTrigger"
        silent-date-reload
      />
    </template>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { GpuMemoryAndComputeWidget } from "@/components/dashboard-v2/widgets/chart-widgets/gpu-memory-and-compute-widget";
import { CpuMemoryAndComputeWidget } from "@/components/dashboard-v2/widgets/chart-widgets/cpu-memory-and-compute-widget";
import { DcgmNvlinkBandwidthSectionPerGpuWidget } from "@/components/dashboard-v2/widgets/chart-widgets/dcgm-nvlink-bandwidth-section-per-gpu-widget";
import { DcgmNvlinkBandwidthTotalBytesSectionWidget } from "@/components/dashboard-v2/widgets/chart-widgets/dcgm-nvlink-bandwidth-total-bytes-section-widget";
import { GpuMemoryUsageAndComputeUtilizationPerGpuWidget } from "@/components/dashboard-v2/widgets/chart-widgets/gpu-memory-usage-and-compute-utilization-per-gpu-widget";
import { CpuMemoryAndComputeUsagePerGpuWidget } from "@/components/dashboard-v2/widgets/chart-widgets/cpu-memory-and-compute-usage-per-gpu-widget";

// types
import type { Workload } from "@/swagger-models/workloads-client";
import type { IRangeDates } from "@/utils/date.util";
import { MIN_CLUSTER_VERSION_FOR_DCGM_MNNVL_METRICS } from "@/common/version.constant";
import { useClusterStore } from "@/stores/cluster.store";

export type TWorkloadCharts = InstanceType<typeof GpuMemoryAndComputeWidget | typeof CpuMemoryAndComputeWidget>;

export default defineComponent({
  name: "default-workload-widgets",
  components: {
    GpuMemoryAndComputeWidget,
    CpuMemoryAndComputeWidget,
    DcgmNvlinkBandwidthSectionPerGpuWidget,
    DcgmNvlinkBandwidthTotalBytesSectionWidget,
    GpuMemoryUsageAndComputeUtilizationPerGpuWidget,
    CpuMemoryAndComputeUsagePerGpuWidget,
  },
  props: {
    selectedPodId: {
      type: String,
      required: false,
    },
    workload: {
      type: Object as PropType<Workload>,
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
    selectedElementId: {
      type: String as PropType<string | null>,
      default: null,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      gpuMemoryAndComputeWidgetRef: "gpu-memory-and-compute-widget" as string,
      cpuMemoryAndComputeWidgetRef: "cpu-memory-and-compute-widget" as string,
      dcgmNvlinkBandwidthSectionPerGpuWidgetRef: "dcgm-nvlink-bandwidth-section-per-gpu-widget" as string,
      dcgmNvlinkBandwidthTotalBytesSectionWidgetRef: "dcgm-nvlink-bandwidth-total-bytes-section-widget" as string,
      gpuMemoryUsageAndComputeUtilizationPerGpuWidgetRef:
        "gpu-memory-usage-and-compute-utilization-per-gpu-widget" as string,
      cpuMemoryAndComputeUsagePerGpuWidgetRef: "cpu-memory-and-compute-usage-per-gpu-widget" as string,
    };
  },
  computed: {
    workloadId(): string {
      return this.workload.id;
    },
    isClusterSupportsDcgmNvlinkMetrics(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workload.clusterId,
        MIN_CLUSTER_VERSION_FOR_DCGM_MNNVL_METRICS,
      );
    },
  },
  methods: {
    resizeCharts(): void {
      (this.$refs[this.gpuMemoryAndComputeWidgetRef] as TWorkloadCharts)?.resizeCharts();
      (this.$refs[this.cpuMemoryAndComputeWidgetRef] as TWorkloadCharts)?.resizeCharts();
      (this.$refs[this.dcgmNvlinkBandwidthSectionPerGpuWidgetRef] as TWorkloadCharts)?.resizeCharts();
      (this.$refs[this.dcgmNvlinkBandwidthTotalBytesSectionWidgetRef] as TWorkloadCharts)?.resizeCharts();
      (this.$refs[this.gpuMemoryUsageAndComputeUtilizationPerGpuWidgetRef] as TWorkloadCharts)?.resizeCharts();
      (this.$refs[this.cpuMemoryAndComputeUsagePerGpuWidgetRef] as TWorkloadCharts)?.resizeCharts();
    },
  },
});
</script>
