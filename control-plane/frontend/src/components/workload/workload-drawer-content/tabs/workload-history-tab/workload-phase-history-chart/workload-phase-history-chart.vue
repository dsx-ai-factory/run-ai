<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <figure class="workload-phase-history-figure" role="img" :aria-label="accessibleDescription">
    <v-chart
      v-if="chartOptions"
      aid="workload-phase-history-chart"
      :option="chartOptions"
      autoresize
      class="workload-phase-history-chart"
      ref="workload-status-history-chart"
    />
  </figure>
</template>
<script lang="ts">
import "@/plugins/echarts";
import { defineComponent, type PropType } from "vue";
import VChart from "vue-echarts";
import type { EChartsOption } from "echarts";
import type { HistoryRecord } from "@/swagger-models/workloads-client";
import { debounce } from "quasar";
import { workloadChartUtil } from "@/utils/workload.util/workload-chart.util";
import { chartUtil } from "@/utils/chart.util";

export default defineComponent({
  components: {
    VChart,
  },
  props: {
    drawerWidth: {
      type: String as PropType<string>,
      required: true,
    },
    phaseUpdateRecords: {
      type: Array as PropType<HistoryRecord[]>,
      required: true,
    },
  },
  data() {
    return {
      chartOptions: null as null | EChartsOption,
    };
  },
  methods: {
    setChartOptions(): void {
      this.chartOptions = {
        ...workloadChartUtil.getWorkloadXRangeOptions(this.phaseUpdateRecords),
        aria: chartUtil.buildChartAria("Workload status history over time."),
      };
    },
    resizeChart() {
      this.$nextTick(() => {
        const chartRef = this.$refs["workload-status-history-chart"] as InstanceType<typeof VChart> | undefined;
        chartRef?.resize();
      });
    },
  },
  computed: {
    accessibleDescription(): string {
      return `Workload status history over time. ${this.phaseUpdateRecords.length} status records shown.`;
    },
  },
  watch: {
    drawerWidth: debounce(function () {
      //@ts-ignore
      this.resizeChart();
    }, 100),
    phaseUpdateRecords: {
      handler() {
        this.setChartOptions();
      },
      immediate: true,
    },
  },
});
</script>

<style lang="scss" scoped>
.workload-phase-history-chart {
  height: 320px;
}

.workload-phase-history-figure {
  margin: 0;
}
</style>
