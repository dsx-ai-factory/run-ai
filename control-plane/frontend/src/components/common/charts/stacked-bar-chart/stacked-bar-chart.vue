<template>
  <div v-if="isReady" class="stacked-bar-chart" role="img" :aria-label="accessibleDescription">
    <v-chart
      aid="stacked-bar-chart"
      ref="chartRef"
      :option="chartOptions"
      autoresize
      class="chart-container"
      :style="chartStyle"
    />
    <!-- Always render scale-labels to preserve vertical space, use visibility to hide -->
    <div class="scale-labels" :class="{ 'scale-labels--hidden': !showScale }">
      <span class="scale-min">{{ minValue }}</span>
      <span class="scale-max">{{ maxValue }}</span>
    </div>
  </div>
</template>

<script lang="ts" setup>
import "@/plugins/echarts";
import { computed, ref, onMounted } from "vue";
import VChart from "vue-echarts";
import type { EChartsOption, BarSeriesOption, LabelLayoutOptionCallbackParams } from "echarts";
import { chartUtil } from "@/utils/chart.util";
import type { IStackedBarSegment } from "./stacked-bar-chart.model";

const DEFAULT_HEIGHT = 20;
const DEFAULT_MIN_VALUE = 0;
const DEFAULT_MAX_VALUE = 100;
const ANIMATION_DURATION = 200;

type TooltipPositionSize = {
  contentSize: [number, number];
};

const props = withDefaults(
  defineProps<{
    segments: IStackedBarSegment[];
    height?: number;
    width?: number;
    showLabels?: boolean;
    minValue?: number;
    maxValue?: number;
    showScale?: boolean;
    entityName?: string;
    showPercentage?: boolean;
    tooltipSummary?: string;
    labelTextBorderWidth?: number;
  }>(),
  {
    height: DEFAULT_HEIGHT,
    width: undefined,
    showLabels: true,
    minValue: DEFAULT_MIN_VALUE,
    maxValue: DEFAULT_MAX_VALUE,
    showScale: true,
    entityName: "",
    showPercentage: false,
    tooltipSummary: "",
    labelTextBorderWidth: 1,
  },
);

const isReady = ref(false);
const chartRef = ref<InstanceType<typeof VChart>>();

onMounted(() => {
  isReady.value = true;
});

const totalValue = computed(() => {
  return props.segments.reduce((sum, segment) => sum + segment.value, 0);
});

const chartStyle = computed(() => ({
  height: `${props.height}px`,
  width: props.width ? `${props.width}px` : "100%",
}));

const accessibleDescription = computed(
  () => `${props.entityName || "Resource"} distribution. ${props.segments.length} segments shown.`,
);

function formatValueWithPercentage(value: number): string {
  if (!props.showPercentage || totalValue.value === 0) {
    return String(value);
  }
  const percentage = Math.round((value / totalValue.value) * 100);
  return `${value} (${percentage}%)`;
}

function keepLabelInsideChart(params: LabelLayoutOptionCallbackParams) {
  const chartWidth = chartRef.value?.$el?.clientWidth || 0;
  if (!chartWidth) {
    return { moveOverlap: "shiftX" as const, hideOverlap: false };
  }

  const minX = 0;
  const maxX = Math.max(minX, chartWidth - params.labelRect.width);
  const clampedX = Math.min(Math.max(params.labelRect.x, minX), maxX);

  // Preserve ECharts' normal vertical layout. An absolute `x` also resets the
  // vertical placement in these short stacked-bar charts.
  return { dx: clampedX - params.labelRect.x, moveOverlap: "shiftX" as const, hideOverlap: false };
}

function buildSeries(): BarSeriesOption[] {
  return props.segments.map((segment: IStackedBarSegment) => ({
    type: "bar" as const,
    name: segment.label || "",
    data: [segment.value],
    stack: "total",
    itemStyle: {
      color: segment.color,
    },
    emphasis: {
      disabled: true,
    },
    // Highcharts allowed data labels to remain visible when a stacked segment
    // was narrower than its label. Keep that behavior for small 0%-like slices.
    clip: false,
    labelLayout: keepLabelInsideChart,
    barCategoryGap: 0,
    barGap: "0%",
    label: {
      show: props.showLabels,
      position: "inside" as const,
      formatter: () => {
        if (segment.value === 0) return "";
        if (segment.displayValue) return segment.displayValue;
        return formatValueWithPercentage(segment.value);
      },
      color: "#46465C",
      fontSize: 14,
      fontWeight: "bold" as const,
      fontFamily: "Roboto, sans-serif",
      textBorderColor: "#ffffff",
      textBorderWidth: props.labelTextBorderWidth,
      overflow: "none",
    },
  }));
}

function formatTooltip(): string {
  let content = "";

  if (props.entityName) {
    content += `<div style="font-weight: 500; padding-bottom: 8px; white-space: normal; overflow-wrap: anywhere;">${chartUtil.escapeTooltipHtml(props.entityName)}</div>`;
  }

  props.segments.forEach((segment: IStackedBarSegment) => {
    if (segment.hideFromTooltip) return;

    const valueDisplay = chartUtil.escapeTooltipHtml(segment.displayValue ?? formatValueWithPercentage(segment.value));
    const label = chartUtil.escapeTooltipHtml(segment.label || "Value");

    content += `
      <div style="display: flex; align-items: center; min-width: 0; padding: 4px 0;">
        ${chartUtil.renderTooltipDot(segment.color)}
        <span style="flex: 1; min-width: 0; white-space: normal; overflow-wrap: anywhere;">${label}: <span style="font-weight: 700;">${valueDisplay}</span></span>
      </div>
    `;
  });

  if (props.tooltipSummary) {
    content += `<div style="padding-top: 4px; padding-left: 13px; white-space: normal; overflow-wrap: anywhere;">${chartUtil.sanitizeTooltipHtml(props.tooltipSummary)}</div>`;
  }

  return `<div style="max-width: 280px; padding: 4px; white-space: normal; overflow-wrap: anywhere;">${content}</div>`;
}

const chartOptions = computed<EChartsOption>(() => ({
  aria: chartUtil.buildChartAria(accessibleDescription.value),
  animation: true,
  animationDuration: ANIMATION_DURATION,
  animationDurationUpdate: ANIMATION_DURATION,
  grid: {
    top: 0,
    bottom: 0,
    left: 0,
    right: 0,
  },
  xAxis: {
    type: "value",
    show: false,
    min: props.minValue,
    max: props.maxValue,
  },
  yAxis: {
    type: "category",
    show: false,
    data: [""],
  },
  legend: {
    show: false,
  },
  tooltip: {
    trigger: "item",
    confine: false,
    appendToBody: false,
    backgroundColor: "#121212",
    borderColor: "#121212",
    textStyle: {
      color: "#F0F0F0",
      fontSize: 12,
      fontFamily: "Roboto, sans-serif",
    },
    formatter: formatTooltip,
    position: function (
      point: [number, number],
      _params: unknown,
      _dom: HTMLElement,
      _rect: unknown,
      size: TooltipPositionSize,
    ): [number, number] {
      const chartRect = chartRef.value?.$el?.getBoundingClientRect();
      if (!chartRect) return [0, 0];
      return [point[0] - size.contentSize[0] / 2, chartRect.height + 16];
    },
  } as unknown as EChartsOption["tooltip"],
  series: buildSeries(),
}));
</script>

<style lang="scss" scoped>
.stacked-bar-chart {
  width: 100%;

  .chart-container {
    width: 100%;
  }

  .scale-labels {
    display: flex;
    justify-content: space-between;
    padding: 4px 0 0 0;
    font-family: Roboto, sans-serif;
    font-size: 12px;
    color: rgba(0, 0, 0, 0.54);

    .scale-min,
    .scale-max {
      font-weight: 400;
    }

    &--hidden {
      visibility: hidden;
    }
  }
}
</style>
