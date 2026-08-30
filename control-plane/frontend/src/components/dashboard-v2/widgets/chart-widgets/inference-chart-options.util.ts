import {
  CHART_AXIS_LABEL_DEFAULTS,
  CHART_COLOR_PALETTE,
  LEGEND_RECT_ITEM_STYLE,
  type IChartWidgetOptions,
} from "@/models/chart.model";
import { chartUtil } from "@/utils/chart.util";
import { dateUtil } from "@/utils/date.util";

// Shared NIM widget defaults: every NIM widget renders as a 250px-tall line chart
// with shared tooltip + crosshair, and uses the same y-axis step ladder.
// Override yAxisTitle per widget ("Requests" vs "Milliseconds", etc.).
export const NIM_WIDGET_Y_AXIS_LABEL_STEPS = [0, 200, 400, 600, 800, 1000];

export function getNimWidgetChartOptions(yAxisTitle: string): IChartWidgetOptions {
  return {
    type: "line",
    height: 250,
    sharedTooltip: true,
    showSharedCrosshair: true,
    yAxisTitle,
    yAxisNameGap: 38,
    highchartsLikeYAxis: true,
  } as IChartWidgetOptions & { yAxisNameGap: number };
}

export const INFERENCE_CHART_BLUE = CHART_COLOR_PALETTE.BLUE;
export const INFERENCE_CHART_ORANGE = CHART_COLOR_PALETTE.ORANGE;

export const inferenceChartGrid = {
  top: 6,
  right: 10,
  bottom: 50,
  left: 34,
  containLabel: true,
};

export const inferenceChartAxisLabel = {
  color: CHART_AXIS_LABEL_DEFAULTS.color,
  fontSize: 12,
  formatter: chartUtil.formatCompactAxisValue,
};

export const inferenceChartSplitLine = {
  show: true,
  lineStyle: {
    color: "#e0e0e0",
  },
};

export const inferenceTooltipAxisPointer = {
  ...chartUtil.getConnectedChartAxisPointer(),
  label: {
    show: false,
  },
} as const;

export function formatInferenceAxisTimestamp(value: number | string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return dateUtil.dateAndTimeFormat(date).replace(", ", "\n");
}

export function getInferenceTimeXAxisOptions(): Record<string, unknown> {
  return {
    type: "time",
    axisLabel: {
      ...inferenceChartAxisLabel,
      fontSize: CHART_AXIS_LABEL_DEFAULTS.fontSize,
      formatter: formatInferenceAxisTimestamp,
      // A two-line date/time label cannot fit at every time tick in a narrow
      // drawer. Let ECharts recalculate the visible labels after resize rather
      // than forcing a viewport-specific tick interval.
      hideOverlap: true,
    },
    axisLine: {
      lineStyle: {
        color: "#b8c7e8",
      },
    },
    axisTick: {
      lineStyle: {
        color: "#b8c7e8",
      },
    },
    // The tooltip owns the hover content. Without this explicit native setting,
    // ECharts can display its default white timestamp label on empty points.
    axisPointer: {
      label: {
        show: false,
      },
    },
  };
}

export function getInferenceChartLegendOptions(names: string[]): Record<string, unknown> {
  return {
    show: true,
    bottom: 0,
    ...LEGEND_RECT_ITEM_STYLE,
    itemGap: 20,
    textStyle: {
      color: "#2b2b2b",
      fontSize: 11,
      fontWeight: 600,
    },
    data: chartUtil.getDashboardLineOnlyLegendData(names),
  };
}
