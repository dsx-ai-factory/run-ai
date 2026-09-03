<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="node-pool-free-gpu-chart__container" role="img" :aria-label="accessibleDescription">
    <div class="node-pool-free-gpu-chart__y-axis-title">Free GPU devices</div>
    <v-chart
      aid="node-pool-free-gpu-chart"
      :option="scatterChartOptions"
      autoresize
      style="height: 215px"
      @click="handlePointClick"
    />
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import VChart from "vue-echarts";
import type { EChartsOption } from "echarts";
import { chartUtil, STANDARD_TOOLTIP_STYLE } from "@/utils/chart.util";
import { chartColorPaletteArray } from "@/models/chart.model";
import { ENodeTelemetryGroupBy } from "@/models/cluster.model";
import type {
  TelemetryResponseValuesInner,
  TelemetryResponseValuesInnerGroupsInner,
} from "@/swagger-models/cluster-service-client";

export interface INodePoolSummary {
  totalFreeGpus: number;
  totalNodes: number;
}
export interface IScatterSeriesPointCoordinates {
  x: number;
  y: number;
}
export interface IScatterSeriesData {
  name: string;
  data: Array<[number, number]>;
  meta: INodePoolSummary;
  color: string;
}
const SMALL_MARKER_RADIUS = 4;
const MEDIUM_MARKER_RADIUS = 14;
const LARGE_MARKER_RADIUS = 20;
export default defineComponent({
  name: "node-pool-free-gpu-chart",
  components: { VChart },
  emits: ["scatter-point-click"],
  props: {
    nodePoolFreeGpusTelemetry: {
      type: Array as PropType<TelemetryResponseValuesInner[]>,
      default: () => [],
    },
  },
  data() {
    return {
      scatterChartOptions: {} as EChartsOption,
      nodePoolNames: [] as string[],
      chartData: [] as IScatterSeriesData[],
    };
  },
  created() {
    this.nodePoolNames = this.getNodePoolNames(this.nodePoolFreeGpusTelemetry);
    this.loadScatterChartData();
    this.loadScatterChartOptions();
  },
  computed: {
    accessibleDescription(): string {
      return `Free GPU devices by node pool. ${this.nodePoolNames.length} node pools shown.`;
    },
  },
  methods: {
    getMarkerRadius(): number {
      const nodesAmount = this.nodePoolFreeGpusTelemetry.length;
      if (nodesAmount < 50) {
        return LARGE_MARKER_RADIUS;
      }
      if (nodesAmount < 100) {
        return MEDIUM_MARKER_RADIUS;
      }
      return SMALL_MARKER_RADIUS;
    },
    loadScatterChartOptions(): void {
      const markerRadius = this.getMarkerRadius();
      const chartDataRef = this.chartData;
      const maxFreeGpus = this.getMaxFreeGpus();
      const isZeroOnly = maxFreeGpus === 0;

      this.scatterChartOptions = {
        aria: chartUtil.buildChartAria("Free GPU devices by node pool."),
        grid: { left: 60, right: 8, top: 10, bottom: 55 },
        legend: {
          show: true,
          bottom: 16,
          left: "center",
          icon: "circle",
          itemWidth: 10,
          itemHeight: 10,
          itemGap: 10,
          textStyle: {
            color: "#333333",
            fontSize: 11,
            fontWeight: 700,
          },
        },
        xAxis: {
          type: "value",
          show: false,
          min: this.nodePoolNames.length > 1 ? -0.5 : 0,
          max: this.nodePoolNames.length > 1 ? this.nodePoolNames.length - 0.5 : 1,
        },
        yAxis: {
          type: "value",
          min: 0,
          max: isZeroOnly ? 8 : undefined,
          interval: 8,
          name: "",
          nameLocation: "middle",
          nameRotate: 90,
          nameGap: 40,
          axisLine: { show: false },
          axisTick: { show: false },
          splitLine: { show: !isZeroOnly, lineStyle: { color: "#e6e6e6", type: "solid" } },
          axisLabel: {
            fontSize: 10,
            margin: 16,
            formatter: (value: number) => (isZeroOnly && value !== 0 ? "" : chartUtil.formatCompactAxisValue(value)),
          },
        },
        tooltip: {
          trigger: "item",
          ...STANDARD_TOOLTIP_STYLE,
          position: (point: number[]) => [0, Math.max(0, point[1] - 100)],
          formatter: (params: unknown) => {
            const p = params as { seriesIndex: number; value: [string, number]; color: string };
            const seriesData = chartDataRef[p.seriesIndex];
            if (!seriesData) return "";
            const totalNodes = seriesData.meta.totalNodes || 0;
            const nodePoolName = seriesData.name;
            const freeGpus = p.value[1] || 0;
            const nodePoolTooltip = chartUtil.formatTooltipValue("Node pool", p.color, nodePoolName, "");
            const totalNodesTooltip = chartUtil.formatTooltipValue("Nodes", "", totalNodes, "");
            const freeGpusTooltip = chartUtil.formatTooltipValue("Free GPUs devices", "", freeGpus, "");
            return `${nodePoolTooltip}${totalNodesTooltip}${freeGpusTooltip}`;
          },
        },
        series: this.chartData.map((s: IScatterSeriesData) => ({
          type: "scatter" as const,
          name: s.name,
          data: s.data,
          z: 3,
          symbolSize: markerRadius,
          itemStyle: {
            color: s.color,
            borderColor: "white",
            borderWidth: 2,
            opacity: 1,
          },
          emphasis: {
            scale: true,
            itemStyle: {
              opacity: 1,
            },
          },
          markLine:
            isZeroOnly && this.chartData[0]?.name === s.name
              ? {
                  silent: true,
                  z: 1,
                  symbol: "none",
                  label: { show: false },
                  lineStyle: {
                    color: "#e6e6e6",
                    width: 1,
                    type: "solid",
                  },
                  data: [{ yAxis: 0 }],
                }
              : undefined,
          cursor: "pointer",
        })),
        color: chartColorPaletteArray,
      };
    },
    getMaxFreeGpus(): number {
      const freeGpuValues = this.chartData.flatMap((series: IScatterSeriesData) =>
        series.data.map((point: [number, number]) => point[1]),
      );
      return Math.max(...freeGpuValues, 0);
    },
    handlePointClick(event: unknown): void {
      const params = event as {
        componentType?: string;
        seriesType?: string;
        seriesIndex?: number;
        dataIndex?: number;
        color?: string;
      };
      if (
        params.componentType !== "series" ||
        params.seriesType !== "scatter" ||
        params.seriesIndex === undefined ||
        params.dataIndex === undefined
      ) {
        return;
      }
      const seriesData = this.chartData[params.seriesIndex];
      if (!seriesData) return;
      this.$emit("scatter-point-click", seriesData.name, params.color || seriesData.color);
    },
    loadScatterChartData(): void {
      const chartData: IScatterSeriesData[] = [];
      const nodePoolSummary = new Map<string, INodePoolSummary>();

      this.nodePoolFreeGpusTelemetry.forEach((telemetryValue: TelemetryResponseValuesInner) => {
        const nodePoolName = this.extractNodePoolNameFromTelemetry(telemetryValue);
        const freeGpus = parseFloat(telemetryValue.value);

        if (nodePoolName) {
          this.updateNodePoolSummary(nodePoolName, freeGpus, nodePoolSummary);
          this.updateChartData(chartData, nodePoolName, freeGpus, nodePoolSummary);
        }
      });
      this.chartData = chartData;
    },
    extractNodePoolNameFromTelemetry(telemetryValue: TelemetryResponseValuesInner): string | undefined {
      return telemetryValue.groups?.find(
        (group: TelemetryResponseValuesInnerGroupsInner) => group.key === ENodeTelemetryGroupBy.Nodepool,
      )?.value;
    },
    updateNodePoolSummary(nodePoolName: string, freeGpus: number, nodePoolSummary: Map<string, INodePoolSummary>): void {
      const currentSummary = nodePoolSummary.get(nodePoolName);
      nodePoolSummary.set(nodePoolName, {
        totalFreeGpus: (currentSummary?.totalFreeGpus || 0) + freeGpus,
        totalNodes: (currentSummary?.totalNodes || 0) + 1,
      });
    },
    getNodePoolNames(data: TelemetryResponseValuesInner[]): string[] {
      const nodePoolSet = new Set<string>();
      data.forEach((item: TelemetryResponseValuesInner) => {
        const nodepool = (item.groups || []).find(
          (group: TelemetryResponseValuesInnerGroupsInner) => group.key === ENodeTelemetryGroupBy.Nodepool,
        )?.value;
        if (nodepool) {
          nodePoolSet.add(nodepool);
        }
      });
      return Array.from(nodePoolSet);
    },
    updateChartData(
      chartData: IScatterSeriesData[],
      nodePoolName: string,
      freeGpus: number,
      nodePoolSummary: Map<string, INodePoolSummary>,
    ): void {
      const nodePoolIndex: number = chartData.findIndex((series: IScatterSeriesData) => series.name === nodePoolName);
      const bubbleIndex = this.nodePoolNames.findIndex((name) => name === nodePoolName);

      const newDataPoint: [number, number] = [
        this.getJitteredXValue(Math.max(bubbleIndex, 0), chartData[nodePoolIndex]?.data.length || 0),
        freeGpus,
      ];

      const meta: INodePoolSummary = nodePoolSummary.get(nodePoolName) || {
        totalFreeGpus: freeGpus,
        totalNodes: 1,
      };

      if (nodePoolIndex === -1) {
        chartData.push({
          name: nodePoolName,
          data: [newDataPoint],
          meta,
          color: chartColorPaletteArray[chartData.length % chartColorPaletteArray.length],
        });
      } else {
        chartData[nodePoolIndex].data.push(newDataPoint);
        chartData[nodePoolIndex].meta = meta;
      }
    },
    getJitteredXValue(baseIndex: number, pointIndex: number): number {
      if (pointIndex === 0) {
        return baseIndex;
      }

      const direction = pointIndex % 2 === 1 ? 1 : -1;
      const distance = Math.min(0.45, Math.ceil(pointIndex / 2) * 0.18);
      return baseIndex + direction * distance;
    },
  },
});
</script>

<style scoped lang="scss">
.node-pool-free-gpu-chart__container {
  position: relative;
  height: 215px;
}

.node-pool-free-gpu-chart__y-axis-title {
  position: absolute;
  top: 39px;
  left: 12px;
  z-index: 1;
  color: #666666;
  font-size: 12px;
  line-height: 14px;
  pointer-events: none;
  writing-mode: vertical-rl;
  transform: rotate(180deg);
}
</style>
