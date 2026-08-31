<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <chart-widget-wrapper
    class="dept-allocation-ratio-widget"
    :options="wrapperOptions"
    :loading="displayLoading"
    :error="error"
    :empty="isChartEmpty"
    hide-actions
  >
    <div v-if="chartOption" ref="chartAreaRef" style="height: 240px">
      <v-chart
        aid="department-allocation-ratio-chart"
        ref="chartRef"
        :option="chartOption"
        autoresize
        style="height: 240px"
        key="dept-allocation-ratio-chart"
      />
    </div>
  </chart-widget-wrapper>
</template>

<script setup lang="ts">
import "@/plugins/echarts";
import { ref, watch } from "vue";
import VChart from "vue-echarts";
import type { EChartsOption } from "echarts";

// models
import {
  EWidgetEmptyMessages,
  type IWidgetWrapperOptions,
  CHART_DATA_LABEL_COLOR,
  chartColorPaletteArray,
} from "@/models/chart.model";
import { HttpErrorResponse } from "@/models/http-response.model";

// components
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";

// services
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import {
  RankingMetricType,
  RankingOrderType,
  type DepartmentGetRanking,
} from "@/swagger-models/org-unit-service-client";
import { chartUtil, STANDARD_TOOLTIP_STYLE } from "@/utils/chart.util";
import { useAdaptiveCategoricalChartLayout } from "@/composables/use-adaptive-categorical-chart-layout.composable";

export interface Props {
  clusterId: string;
  nodePoolName?: string;
  hasDepartments: boolean;
}

const props = defineProps<Props>();

const displayLoading = ref(true);
const chartOption = ref<EChartsOption | null>(null);
const error = ref(false);
const isChartEmpty = ref(false);
const latestReqId = ref(0);

const wrapperOptions: IWidgetWrapperOptions = {
  title: "Departments with lowest allocation ratio",
  tooltipText: "Departments ranked by GPU allocation / quota ratio (lowest first)",
  emptyMessage: EWidgetEmptyMessages.Default,
};

interface AllocationDataPoint {
  value: number;
  quota: number;
  allocated: number;
}

const chartGridOptions = {
  top: 11,
  bottom: 87,
  left: 56,
  right: 12,
};

const yAxisTitleOptions = {
  name: "Ratio",
  nameLocation: "middle" as const,
  nameRotate: 90,
  nameGap: 36,
  nameTextStyle: {
    color: "#585f69",
    fontSize: 12,
    fontWeight: 400,
  },
};

const legendOptions = {
  show: true,
  bottom: 16,
  icon: "circle",
  itemWidth: 12,
  itemHeight: 12,
  itemGap: 20,
  textStyle: {
    fontWeight: "bold" as const,
    fontSize: 12,
    color: "#333",
    padding: [0, 0, 0, 5],
  },
};

const chartRef = ref<InstanceType<typeof VChart> | null>(null);
const chartAreaRef = ref<HTMLElement | null>(null);
const { getCategoryAxisLabelOptions } = useAdaptiveCategoricalChartLayout({
  chartRef,
  containerRef: chartAreaRef,
  horizontalGridPadding: chartGridOptions.left + chartGridOptions.right,
});

function initChartOption(): void {
  chartOption.value = {
    color: chartColorPaletteArray,
    grid: chartGridOptions,
    xAxis: {
      type: "category",
      data: [] as string[],
      axisLabel: {
        ...getCategoryAxisLabelOptions(1),
        interval: 0,
        rotate: 0,
        margin: 10,
        color: "#666",
        fontSize: 11,
        overflow: "break",
        lineHeight: 16,
        hideOverlap: false,
      },
      axisLine: { show: true, lineStyle: { color: "#ccd6eb", width: 1 } },
      axisTick: { show: false },
      splitLine: { show: false },
    },
    yAxis: {
      type: "value",
      min: 0,
      axisLabel: { formatter: "{value}%", margin: 12, color: "#666", fontSize: 11 },
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: {
        show: true,
        lineStyle: { color: "#e6e6e6", width: 1, type: "solid" },
      },
      ...yAxisTitleOptions,
    },
    tooltip: {
      trigger: "axis",
      axisPointer: chartUtil.getHiddenAxisPointer(),
      ...STANDARD_TOOLTIP_STYLE,
    },
    legend: { ...legendOptions, data: ["Allocation ratio"] },
    series: [],
  };
}

async function loadChartData(): Promise<void> {
  if (chartOption.value === null) return;
  const reqId = ++latestReqId.value;
  try {
    const departments = await orgUnitService.getDepartmentsRanking(
      RankingMetricType.GpuAllocationRatio,
      RankingOrderType.Bottom,
      props.clusterId,
      props.nodePoolName,
    );

    if (reqId !== latestReqId.value) return;

    if (departments.length === 0) {
      isChartEmpty.value = true;
    } else {
      isChartEmpty.value = false;
      const categories = departments.map((dept: DepartmentGetRanking) => dept.name || "");
      const data: AllocationDataPoint[] = departments.map((dept: DepartmentGetRanking) => {
        const ratio = dept.allocationValue ?? 0;
        const quota = dept.gpuQuota ?? 0;
        const allocated = Math.round(ratio * quota * 100) / 100;
        return { value: Math.round(ratio * 100), quota, allocated };
      });
      const yAxisConfig = chartUtil.getBarChartYAxisConfig([{ data }]);

      chartOption.value = {
        ...chartOption.value,
        xAxis: {
          type: "category",
          data: categories,
          axisLabel: {
            ...getCategoryAxisLabelOptions(categories.length),
            interval: 0,
            rotate: 0,
            margin: 10,
            color: "#666",
            fontSize: 11,
            overflow: "break",
            lineHeight: 16,
            hideOverlap: false,
          },
          axisLine: { show: true, lineStyle: { color: "#ccd6eb", width: 1 } },
          axisTick: { show: false },
          splitLine: { show: false },
        },
        yAxis: {
          type: "value",
          min: 0,
          ...yAxisConfig,
          axisLabel: { formatter: "{value}%", margin: 12, color: "#666", fontSize: 11 },
          axisLine: { show: false },
          axisTick: { show: false },
          splitLine: {
            show: true,
            lineStyle: { color: "#e6e6e6", width: 1, type: "solid" },
          },
          ...yAxisTitleOptions,
        },
        series: [
          {
            type: "bar",
            name: "Allocation ratio",
            data,
            barMaxWidth: 30,
            barCategoryGap: "40%",
            ...chartUtil.getBarSeriesDefaults(),
            label: {
              show: true,
              position: "top",
              formatter: "{c}%",
              color: CHART_DATA_LABEL_COLOR,
              fontSize: 12,
              fontWeight: 400,
            },
          },
        ],
        tooltip: {
          trigger: "axis",
          axisPointer: chartUtil.getHiddenAxisPointer(),
          ...STANDARD_TOOLTIP_STYLE,
          formatter: (params: unknown) => {
            const items = Array.isArray(params) ? params : [params];
            const param = items[0] as {
              color: string;
              seriesName: string;
              data: AllocationDataPoint;
              axisValueLabel?: string;
              name?: string;
            };
            if (!param?.data) return "";
            const header = chartUtil.escapeTooltipHtml(param.axisValueLabel || param.name || "");
            const color = chartUtil.escapeTooltipHtml(param.color);
            const seriesName = chartUtil.escapeTooltipHtml(param.seriesName);
            const value = chartUtil.escapeTooltipHtml(param.data.value);
            const quota = chartUtil.escapeTooltipHtml(param.data.quota);
            const allocated = chartUtil.escapeTooltipHtml(param.data.allocated);
            return `${header ? `<div style="padding-bottom: 4px;">${header}</div>` : ""}<span style="color:${color}">\u25CF</span> ${seriesName}: <b>${value}%</b><br/>GPU Quota: <b>${quota}</b><br/>GPU Allocated: <b>${allocated}</b>`;
          },
        },
      };
    }

    error.value = false;
  } catch (err: unknown) {
    if (reqId !== latestReqId.value) return;
    error.value = true;
    if (err instanceof HttpErrorResponse) {
      console.error(err.message);
    } else {
      console.error("Error fetching data", err);
    }
  } finally {
    if (reqId === latestReqId.value) {
      displayLoading.value = false;
    }
  }
}

initChartOption();
loadChartData();

watch(
  () => [props.clusterId, props.nodePoolName],
  () => {
    displayLoading.value = true;
    loadChartData();
  },
);
</script>
