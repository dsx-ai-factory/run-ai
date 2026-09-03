<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div role="img" :aria-label="accessibleDescription">
    <v-chart aid="node-free-gpu-chart" :option="bubbleChartOptions" autoresize style="height: 240px" />
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import VChart from "vue-echarts";
import type { EChartsOption } from "echarts";
import { chartUtil, STANDARD_TOOLTIP_STYLE } from "@/utils/chart.util";
import type {
  TelemetryResponseValuesInner,
  TelemetryResponseValuesInnerGroupsInner,
} from "@/swagger-models/cluster-service-client";
import { ENodeTelemetryGroupBy } from "@/models/cluster.model";
import type { CHART_COLOR_PALETTE } from "@/models/chart.model";
import { chartColorPaletteArray } from "@/models/chart.model";

export interface IBubbleDataPoint {
  name: string;
  value: number;
  color: string;
  totalGpusDevices: number;
}

const SINGLE_BUBBLE_SIZE = 118;
const MIN_SYMBOL_SIZE = 54;
const MAX_SYMBOL_SIZE = 118;

export default defineComponent({
  name: "node-free-gpu-chart",
  components: { VChart },
  props: {
    nodePoolName: {
      type: String as PropType<string>,
      default: "",
    },
    nodePoolColor: {
      type: [String, null] as PropType<CHART_COLOR_PALETTE | null>,
      default: "",
    },
    nodePoolFreeGpusTelemetry: {
      type: Array as PropType<TelemetryResponseValuesInner[]>,
      default: () => [],
    },
    nodeTotalGpuDevicesTelemetry: {
      type: Array as PropType<TelemetryResponseValuesInner[]>,
      default: () => [],
    },
  },
  data() {
    return {
      bubbleChartOptions: {} as EChartsOption,
    };
  },
  created() {
    this.loadBubbleChart(this.nodePoolFreeGpusTelemetry, this.nodePoolName, this.nodePoolColor);
  },
  methods: {
    loadBubbleChart(
      telemetryData: TelemetryResponseValuesInner[],
      nodePoolName: string,
      nodePoolColor: CHART_COLOR_PALETTE | null,
    ): void {
      const dataPoints = this.filterTelemetryData(telemetryData, nodePoolName, nodePoolColor);
      const maxValue = Math.max(...dataPoints.map((d) => d.value), 0);

      this.bubbleChartOptions = {
        aria: chartUtil.buildChartAria("Free GPU devices for the selected node pool."),
        animation: false,
        grid: { left: 20, right: 20, top: 35, bottom: 70 },
        legend: {
          show: true,
          bottom: 22,
          left: "center",
          icon: "circle",
          itemWidth: 12,
          itemHeight: 12,
          textStyle: {
            color: "#333333",
            fontSize: 10,
            fontWeight: "normal",
          },
        },
        xAxis: { type: "value", show: false, min: 0, max: 1 },
        yAxis: { type: "value", show: false, min: 0, max: 1 },
        tooltip: {
          trigger: "item",
          ...STANDARD_TOOLTIP_STYLE,
          position: (point: number[]) => [point[0] + 80, Math.max(0, point[1] - 30)],
          formatter: (params: unknown) => {
            const p = params as {
              data: { name: string; value: number[]; meta: { totalGpusDevices: number }; itemStyle: { color: string } };
            };
            const d = p.data;
            const nodeTooltip = chartUtil.formatTooltipValue("Node", d.itemStyle.color, d.name, "");
            const totalGpusTooltip = chartUtil.formatTooltipValue("Total GPU devices", "", d.meta.totalGpusDevices, "");
            const freeGpusTooltip = chartUtil.formatTooltipValue("Free GPU devices  ", "", d.value[2], "");
            return `${nodeTooltip}${totalGpusTooltip}${freeGpusTooltip}`;
          },
        },
        series: [
          {
            type: "scatter",
            name: nodePoolName,
            itemStyle: {
              color: nodePoolColor || chartColorPaletteArray[0],
            },
            data: dataPoints.map((d, i) => ({
              name: d.name,
              value: [...this.getBubblePosition(i, dataPoints.length), d.value],
              meta: { totalGpusDevices: d.totalGpusDevices },
              itemStyle: {
                color: d.color,
                borderColor: d.color,
                borderWidth: 1,
                opacity: 0.5,
              },
            })),
            symbolSize: (val: number[]) => {
              if (dataPoints.length === 1) {
                return SINGLE_BUBBLE_SIZE;
              }
              const ratio = maxValue > 0 ? val[2] / maxValue : 1;
              return MIN_SYMBOL_SIZE + ratio * (MAX_SYMBOL_SIZE - MIN_SYMBOL_SIZE);
            },
            label: {
              show: false,
            },
            emphasis: {
              scale: true,
              itemStyle: {
                opacity: 0.7,
              },
            },
          },
        ],
      };
    },
    getBubblePosition(index: number, total: number): [number, number] {
      if (total <= 1) {
        return [0.5, 0.5];
      }

      const angle = (Math.PI * 2 * index) / total - Math.PI / 2;
      const radius = total <= 4 ? 0.22 : 0.28;
      return [0.5 + Math.cos(angle) * radius, 0.5 + Math.sin(angle) * radius];
    },
    filterTelemetryData(
      telemetryData: TelemetryResponseValuesInner[],
      nodePoolName: string,
      nodePoolColor: CHART_COLOR_PALETTE | null,
    ): IBubbleDataPoint[] {
      const telemetryDataMap = new Map<string, IBubbleDataPoint>();
      const paletteWithoutCurrentNodePoolColor: CHART_COLOR_PALETTE[] = chartColorPaletteArray.filter(
        (color: CHART_COLOR_PALETTE) => color !== nodePoolColor,
      );

      telemetryData.forEach((telemetryInner) => {
        const group = this.extractNodePoolGroupFromTelemetry(telemetryInner, nodePoolName);

        if (group) {
          const nodeName = this.getNodeName(telemetryInner);
          const totalGpusDevices = this.getTotalGpusDevicesByNodeName(nodeName || "");
          if (nodeName) {
            telemetryDataMap.set(nodeName, {
              name: nodeName,
              value: parseFloat(telemetryInner.value),
              color:
                paletteWithoutCurrentNodePoolColor[telemetryDataMap.size % paletteWithoutCurrentNodePoolColor.length] ||
                chartColorPaletteArray[0],
              totalGpusDevices,
            });
          }
        }
      });

      return Array.from(telemetryDataMap.values());
    },
    extractNodePoolGroupFromTelemetry(
      telemetryInner: TelemetryResponseValuesInner,
      nodePoolName: string,
    ): TelemetryResponseValuesInnerGroupsInner | undefined {
      return telemetryInner?.groups?.find(
        (groupInner: TelemetryResponseValuesInnerGroupsInner) =>
          groupInner.key === ENodeTelemetryGroupBy.Nodepool && groupInner.value === nodePoolName,
      );
    },
    getNodeName(telemetryValue: TelemetryResponseValuesInner): string | undefined {
      return telemetryValue?.groups?.find(
        (group: TelemetryResponseValuesInnerGroupsInner) => group.key === ENodeTelemetryGroupBy.Node,
      )?.value;
    },
    getTotalGpusDevicesByNodeName(nodeName: string): number {
      const node = this.nodeTotalGpuDevicesTelemetry.find((node: TelemetryResponseValuesInner) =>
        node.groups?.find(
          (group: TelemetryResponseValuesInnerGroupsInner) =>
            group.key === ENodeTelemetryGroupBy.Node && group.value === nodeName,
        ),
      );
      return node ? parseFloat(node.value) : 0;
    },
  },
  computed: {
    accessibleDescription(): string {
      return `Free GPU devices for ${this.nodePoolName || "the selected node pool"}.`;
    },
  },
});
</script>

<style scoped lang="scss"></style>
