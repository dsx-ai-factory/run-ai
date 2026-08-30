import {
  type CHART_COLOR_PALETTE,
  CHART_AXIS_LABEL_DEFAULTS,
  CHART_CROSSHAIR_COLOR,
  CHART_DATA_LABEL_COLOR,
  INSIDE_DATAZOOM_CONFIG,
  LEGEND_RECT_ICON,
  PROJECT_RESOURCE_ALLOCATION_COLORS,
} from "@/models/chart.model";
import { type ECharts, getInstanceByDom } from "echarts/core";
import {
  chartColorPaletteArray,
  type IChartWidgetOptions,
  type ITooltipPoint,
  type TMeasurementsTimeAndValue,
  type IZoomEvent,
  TOOLTIP_BASE_COLOR,
  EYAxisLabelFormat,
} from "@/models/chart.model";
import type { EChartsOption, BarSeriesOption } from "echarts";
import type { LineSeriesOption } from "echarts/charts";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";
import { addSeconds } from "date-fns";
import type { MeasurementResponseValuesInner } from "@/swagger-models/workloads-client";
import { type MeasurementResponse } from "@/swagger-models/workloads-client";
import DOMPurify from "dompurify";

export const chartUtil = {
  getResourcesOptions,
  getBasicWidgetChartOptions,
  formatTooltipValue,
  formatSharedTooltip,
  getTimeFrames,
  getTimeFrameForHoursAgo,
  arrangeMeasurementsByTime,
  getDatetimeOptions,
  getDatetimeAxisLabelOptions,
  formatDatetimeAxisLabel,
  assignColorsToKeys,
  addZoomOptions,
  resetChartDataZoom,
  resetChartDataZoomInstance,
  getDashboardChartToolbox,
  getBarChartYAxisConfig,
  getLineChartYAxisConfig,
  formatCompactAxisValue,
  formatCategoricalAxisLabel,
  getDataZoomRange,
  getMinimumDataZoomValueSpan,
  getDashboardLineOnlyLegendIcon,
  getDashboardLineOnlyLegendData,
  getDashboardLineSeriesStyle,
  getDottedLineStyle,
  isDottedLineStyle,
  getConnectedChartHoverMarker,
  getConnectedChartAxisPointer,
  getConnectedTooltipAnchorSeries,
  isConnectedTooltipAnchorSeries,
  getBarSeriesLegendHoverState,
  getBarSeriesDefaults,
  getHiddenAxisPointer,
  updateYAxisLabelsDefaultsIfNeeded,
  getBasicTooltipOptions,
  buildStandardTooltip,
  buildChartAria,
  escapeTooltipHtml,
  sanitizeTooltipHtml,
  renderTooltipDot,
  formatDefaultTooltip,
  hideTooltipsWithin,
};

const DASHBOARD_LINE_LEGEND_ICON =
  "path://M0,5 L8,5 L8,7 L0,7 Z M8,6 A5,5 0 1,0 18,6 A5,5 0 1,0 8,6 M18,5 L26,5 L26,7 L18,7 Z";
const DASHBOARD_LINE_ONLY_LEGEND_ICON = LEGEND_RECT_ICON;
const LEGEND_HOVER_BLUR_OPACITY = 0.2;
const CONNECTED_TOOLTIP_ANCHOR_SERIES_ID = "connected-tooltip-anchor";
const DOTTED_LINE_DASH_PATTERN = [4, 4];
function formatCategoricalAxisLabel(value: string): string {
  // Preserve the complete category value here. The adaptive categorical
  // layout assigns a pixel width to each axis slot and lets ECharts apply an
  // ellipsis only when the rendered text does not fit that slot.
  return value;
}

export interface IDataZoomPayload {
  batch?: IDataZoomPayload[];
  startValue?: number | string;
  endValue?: number | string;
  start?: number | string;
  end?: number | string;
}

function getResourcesOptions(overQuotaValue: number, quotaValue: number, allocatedValue: number): EChartsOption {
  return {
    aria: buildChartAria("Project resource allocation, quota, and over-quota values."),
    animation: false,
    grid: {
      top: 0,
      left: 0,
      right: 0,
      bottom: 30,
    },
    tooltip: {
      trigger: "axis",
      hideDelay: 50,
      axisPointer: getHiddenAxisPointer(),
      ...STANDARD_TOOLTIP_STYLE,
      appendTo: "body",
      confine: false,
      transitionDuration: 0.6,
      displayTransition: true,
      // This compact chart kept Highcharts' light, translucent tooltip rather
      // than the dark dashboard tooltip. It is used only by Project cards.
      backgroundColor: "rgba(255, 255, 255, 0.85)",
      borderColor: TOOLTIP_BASE_COLOR,
      borderWidth: 1,
      borderRadius: 3,
      extraCssText: "box-shadow: 0 2px 6px rgba(0, 0, 0, 0.25);",
      textStyle: { color: "#333333", fontFamily: "Roboto, sans-serif", fontSize: 12 },
      formatter: (params: Record<string, unknown> | Record<string, unknown>[]) => {
        const items = Array.isArray(params) ? params : [params];
        const labels: Record<string, number> = {
          "Over quota": overQuotaValue,
          Quota: quotaValue,
          Allocated: allocatedValue,
        };
        const colors: Record<string, string> = {
          "Over quota": PROJECT_RESOURCE_ALLOCATION_COLORS.overQuota,
          Quota: PROJECT_RESOURCE_ALLOCATION_COLORS.quota,
          Allocated: PROJECT_RESOURCE_ALLOCATION_COLORS.allocated,
        };
        return items
          .map((item) => {
            const name = item.seriesName as string;
            const color = colors[name] || "";
            const val = labels[name] ?? 0;
            return `<div style="line-height:16px;white-space:nowrap;">${renderTooltipDot(color, 8)}${escapeTooltipHtml(name)}: <b>${escapeTooltipHtml(val)} GPU</b></div>`;
          })
          .join("");
      },
    } as unknown as EChartsOption["tooltip"],
    xAxis: {
      type: "category",
      show: false,
      data: [""],
    },
    yAxis: {
      type: "value",
      show: false,
    },
    legend: {
      show: true,
      bottom: 0,
      left: "center",
      icon: "path://M4,0A4,4,0,1,1,4,8A4,4,0,1,1,4,0Z",
      width: 224,
      textStyle: {
        color: CHART_CROSSHAIR_COLOR,
        fontWeight: "normal",
        fontSize: 12,
        padding: [0, 0, 0, -2],
      },
      itemWidth: 8,
      itemHeight: 8,
      itemGap: 14,
    },
    series: [
      {
        type: "bar",
        name: "Over quota",
        color: PROJECT_RESOURCE_ALLOCATION_COLORS.overQuota,
        data: [quotaValue + overQuotaValue],
        barWidth: 70,
        barGap: "-100%",
        z: 1,
      },
      {
        type: "bar",
        name: "Quota",
        color: PROJECT_RESOURCE_ALLOCATION_COLORS.quota,
        data: [quotaValue],
        barWidth: 70,
        barGap: "-100%",
        z: 2,
      },
      {
        type: "bar",
        name: "Allocated",
        color: PROJECT_RESOURCE_ALLOCATION_COLORS.allocated,
        data: [allocatedValue],
        barWidth: 50,
        barGap: "-100%",
        z: 3,
      },
    ],
  };
}

export const STANDARD_TOOLTIP_STYLE = {
  backgroundColor: TOOLTIP_BASE_COLOR,
  borderColor: TOOLTIP_BASE_COLOR,
  appendToBody: false,
  confine: true,
  textStyle: { color: "#F0F0F0", fontFamily: "Roboto, sans-serif", fontSize: 12 },
} as const;

function getBasicTooltipOptions(sharedTooltip?: boolean): EChartsOption["tooltip"] {
  return buildStandardTooltip({ trigger: sharedTooltip ? "axis" : "item" });
}

function buildStandardTooltip(options: Record<string, unknown> = {}): EChartsOption["tooltip"] {
  const { axisPointer, ...tooltipOptions } = options;
  const shouldUseVisibleAxisPointer = tooltipOptions.trigger === "axis" || Boolean(axisPointer);
  const axisPointerOptions = axisPointer as Record<string, unknown> | undefined;
  const axisPointerLineStyle = axisPointerOptions?.lineStyle as Record<string, unknown> | undefined;

  return {
    ...STANDARD_TOOLTIP_STYLE,
    ...tooltipOptions,
    ...(shouldUseVisibleAxisPointer
      ? {
          axisPointer: {
            ...getConnectedChartAxisPointer(),
            ...axisPointerOptions,
            lineStyle: {
              ...getConnectedChartAxisPointer().lineStyle,
              ...axisPointerLineStyle,
            },
          },
        }
      : {}),
  };
}

const DEFAULT_CHART_ARIA_DESCRIPTION = "Chart visualization. Refer to the chart heading and summary for details.";

function buildChartAria(description?: string, decal = false): EChartsOption["aria"] {
  return {
    enabled: true,
    description: description || DEFAULT_CHART_ARIA_DESCRIPTION,
    decal: { show: decal },
  };
}

function getDashboardLineOnlyLegendIcon(): string {
  return DASHBOARD_LINE_ONLY_LEGEND_ICON;
}

function getDashboardLineOnlyLegendData(names: string[]): Array<{ name: string; icon: string }> {
  return names.map((name) => ({ name, icon: DASHBOARD_LINE_ONLY_LEGEND_ICON }));
}

function getDashboardLineSeriesStyle(
  index = 0,
  lineStyle: LineSeriesOption["lineStyle"] = {},
): Pick<LineSeriesOption, "symbol" | "symbolSize" | "lineStyle" | "itemStyle" | "emphasis"> {
  const color = chartColorPaletteArray[index % chartColorPaletteArray.length];

  return {
    symbol: "circle",
    symbolSize: 16,
    lineStyle: {
      ...lineStyle,
      color,
    },
    itemStyle: { color },
    emphasis: {
      scale: false,
      itemStyle: {
        color: "#FFFFFF",
        borderColor: color,
        borderWidth: 2,
      },
    },
  };
}

function getDottedLineStyle(overrides: LineSeriesOption["lineStyle"] = {}): NonNullable<LineSeriesOption["lineStyle"]> {
  return {
    width: 2,
    type: [...DOTTED_LINE_DASH_PATTERN],
    ...overrides,
  };
}

function isDottedLineStyle(lineStyle: LineSeriesOption["lineStyle"]): boolean {
  const type = lineStyle?.type;
  return type === "dotted" || Array.isArray(type);
}

function getConnectedChartHoverMarker(color: string): Pick<LineSeriesOption, "symbol" | "symbolSize" | "emphasis"> {
  return {
    symbol: "circle",
    symbolSize: 16,
    emphasis: {
      scale: false,
      itemStyle: {
        color: "#FFFFFF",
        borderColor: color,
        borderWidth: 2,
      },
    },
  };
}

function getConnectedChartAxisPointer() {
  return {
    z: 1,
    type: "line" as const,
    lineStyle: {
      width: 2,
      color: CHART_CROSSHAIR_COLOR,
      type: "solid" as const,
    },
  };
}

function getConnectedTooltipAnchorSeries(timeframes: string[]): LineSeriesOption {
  return {
    id: CONNECTED_TOOLTIP_ANCHOR_SERIES_ID,
    type: "line",
    data: timeframes.map((timeframe) => [new Date(timeframe).getTime(), 0]),
    showSymbol: false,
    symbol: "none",
    silent: true,
    animation: false,
    lineStyle: { color: "transparent", opacity: 0, width: 0 },
    itemStyle: { color: "transparent", opacity: 0 },
    emphasis: { disabled: true },
    z: -1,
  };
}

function isConnectedTooltipAnchorSeries(item: { id?: unknown; seriesId?: unknown }): boolean {
  return item.id === CONNECTED_TOOLTIP_ANCHOR_SERIES_ID || item.seriesId === CONNECTED_TOOLTIP_ANCHOR_SERIES_ID;
}

function getBarSeriesLegendHoverState(): Partial<BarSeriesOption> {
  return {
    legendHoverLink: true,
    emphasis: {
      focus: "series",
    },
    blur: {
      itemStyle: {
        opacity: LEGEND_HOVER_BLUR_OPACITY,
      },
      label: {
        opacity: LEGEND_HOVER_BLUR_OPACITY,
      },
    },
  };
}

// Shared defaults for drilldown/category bar series: legend hover state
// + minimum visible bar height so real zero values still render a tiny bar
// (matches Highcharts minPointLength: 3 behavior in prod).
function getBarSeriesDefaults(): Partial<BarSeriesOption> {
  return {
    ...getBarSeriesLegendHoverState(),
    barMinHeight: 3,
  };
}

function getHiddenAxisPointer(): {
  type: "line";
  lineStyle: { opacity: 0; width: 0 };
  label: { show: false };
} {
  return {
    type: "line",
    lineStyle: { opacity: 0, width: 0 },
    label: { show: false },
  };
}

function getBasicWidgetChartOptions(chartOptions: IChartWidgetOptions): EChartsOption {
  const isLineChart = chartOptions.type === "line";
  const basicOptions: EChartsOption = {
    // Match prod (Highcharts) font: Roboto. ECharts default platform sans-serif renders wider.
    textStyle: { fontFamily: "Roboto, sans-serif" },
    tooltip: getBasicTooltipOptions(chartOptions.sharedTooltip),
    aria: buildChartAria(chartOptions.accessibility?.description, chartOptions.accessibility?.decal),
    legend: {
      show: true,
      bottom: 0,
      icon: isLineChart ? DASHBOARD_LINE_LEGEND_ICON : "circle",
      itemWidth: isLineChart ? 26 : 12,
      itemHeight: 12,
    },
    grid: {
      containLabel: true,
      top: 30,
      bottom: 50,
      ...chartOptions.grid,
    },
    yAxis: {
      type: "value",
      min: 0,
      name: chartOptions.yAxisTitle,
      nameLocation: "middle",
      nameRotate: 90,
      nameGap: 56,
    },
    xAxis: {
      type: "category",
    },
    color: chartOptions.colors || chartColorPaletteArray,
    series: [],
  };

  if (chartOptions.showSharedCrosshair) {
    (basicOptions.tooltip as Record<string, unknown>).axisPointer = getConnectedChartAxisPointer();
  }

  if (chartOptions.yAxisFormatFunction) {
    const yAxisFormatter = chartOptions.yAxisFormatFunction;
    basicOptions.yAxis = {
      ...(basicOptions.yAxis as Record<string, unknown>),
      axisLabel: {
        formatter: (value: number) => yAxisFormatter(value),
      },
    };
  }

  if (chartOptions.showDataLabels) {
    basicOptions.series = [
      {
        type: "bar",
        label: {
          show: true,
          position: "top",
          formatter: chartOptions.isPercentage ? "{c}%" : "{c}",
          color: CHART_DATA_LABEL_COLOR,
          fontSize: 12,
          fontWeight: 400,
        },
      },
    ];
  }

  return basicOptions;
}

// Fixed 6x6 round tooltip marker (Figma spec) \u2014 uniform across all charts,
// independent of tooltip font size (replaces the size-varying \u25CF glyph).
function renderTooltipDot(color: string, size = 6): string {
  const safeSize = Number.isFinite(size) && size > 0 ? size : 6;
  return `<span style="display:inline-block;flex:0 0 ${safeSize}px;width:${safeSize}px;height:${safeSize}px;border-radius:999px;background:${escapeTooltipHtml(color)};margin-right:6px;vertical-align:middle;"></span>`;
}

function escapeTooltipHtml(value: unknown): string {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function sanitizeTooltipHtml(value: unknown): string {
  return DOMPurify.sanitize(String(value ?? ""), {
    ALLOWED_TAGS: ["b", "strong", "br"],
    ALLOWED_ATTR: [],
  });
}

function formatDefaultTooltip(params: unknown): string {
  const points = Array.isArray(params) ? params : [params];
  const first = points[0] as Record<string, unknown> | undefined;
  if (!first) return "";

  const header = escapeTooltipHtml(first.axisValueLabel ?? first.name ?? "");
  const content = points.reduce((html, point) => {
    const item = point as Record<string, unknown>;
    const data = item.data as Record<string, unknown> | undefined;
    const rawValue = item.value ?? data?.value ?? "";
    const value = Array.isArray(rawValue) ? rawValue[rawValue.length - 1] : rawValue;
    const name = escapeTooltipHtml(item.seriesName ?? item.name ?? "");
    const color = String(item.color ?? "");

    return `${html}${renderTooltipDot(color)}${name}: <b>${escapeTooltipHtml(value)}</b><br>`;
  }, "");

  return `${header ? `<div style="padding-bottom: 4px;">${header}</div>` : ""}${content}`;
}

function formatTooltipValue(
  name: string,
  color: string | undefined,
  value: number | string,
  symbol = "",
  highlightName = false,
): string {
  const colorPoint = color ? renderTooltipDot(color) : `<span>&nbsp;&nbsp;&nbsp;&nbsp;</span>`;

  const escapedName = escapeTooltipHtml(name);
  const nameText = highlightName ? `<b>${escapedName}</b>` : escapedName;
  return `${colorPoint ? colorPoint + " " : ""}${nameText}: <b>${escapeTooltipHtml(value)}${escapeTooltipHtml(symbol)}</b><br>`;
}

const MULTI_COLUMN_TOOLTIP_THRESHOLD = 8;

function formatSharedTooltip(points: Array<ITooltipPoint>, time?: string | number): string {
  const header = time
    ? `<div style="padding-left: 3px; padding-bottom: 5px;">${dateUtil.dateFormat(
        new Date(time),
        DATE_FORMATS.STANDARD_DATE_TIME,
      )}</div>`
    : "";

  const validPoints = points.filter((point): point is ITooltipPoint & { y: string | number } => {
    return point.y !== undefined && point.y !== null;
  });

  if (validPoints.length > MULTI_COLUMN_TOOLTIP_THRESHOLD) {
    const mid = Math.ceil(validPoints.length / 2);
    const renderColumn = (pts: Array<ITooltipPoint & { y: string | number }>) =>
      pts.reduce((html, pt) => html + chartUtil.formatTooltipValue(pt.name, pt.color, pt.y, pt.symbol), "");
    return `${header}<div style="display: flex; gap: 12px;"><div>${renderColumn(validPoints.slice(0, mid))}</div><div>${renderColumn(validPoints.slice(mid))}</div></div>`;
  }

  return validPoints.reduce(
    (tooltip, point) => tooltip + chartUtil.formatTooltipValue(point.name, point.color, point.y, point.symbol),
    header,
  );
}

function getTimeFrames(dateStart: Date, dateEnd: Date): string[] {
  const MAX_DATA_POINTS = 500;

  // Calculate total duration in seconds
  const durationInSeconds = Math.floor((dateEnd.getTime() - dateStart.getTime()) / 1000);

  // Validate date range
  if (durationInSeconds <= 0) {
    throw new Error(`Invalid date range: dateStart (${dateStart}) must be before dateEnd (${dateEnd})`);
  }

  // Validate maximum range (5 years)
  const FIVE_YEARS_IN_SECONDS = 5 * 365 * 24 * 60 * 60; // ~157,680,000 seconds
  if (durationInSeconds > FIVE_YEARS_IN_SECONDS) {
    throw new Error(`Range time: ${dateStart} - ${dateEnd} is not supported. Maximum range is 5 years.`);
  }

  // Define reasonable interval tiers (in seconds) that scale naturally
  const intervalTiers = [
    // Seconds (for very short ranges)
    15, // 15 seconds
    30, // 30 seconds
    // Minutes (for short-medium ranges)
    60, // 1 minute
    120, // 2 minutes
    300, // 5 minutes
    600, // 10 minutes
    900, // 15 minutes
    1800, // 30 minutes
    // Hours (for medium-long ranges)
    3600, // 1 hour
    7200, // 2 hours
    10800, // 3 hours
    21600, // 6 hours
    43200, // 12 hours
    // Days (for long ranges)
    86400, // 1 day
    172800, // 2 days
    259200, // 3 days
    604800, // 7 days
    1209600, // 14 days
    2592000, // 30 days
    // Extended ranges (for very long periods)
    7776000, // 90 days (3 months)
    15552000, // 180 days (6 months)
    31536000, // 365 days (1 year)
    63072000, // 730 days (2 years)
  ];

  // Find the smallest interval that keeps data points <= MAX_DATA_POINTS
  let intervalInSeconds = intervalTiers[0];

  for (const tier of intervalTiers) {
    const estimatedPoints = Math.ceil(durationInSeconds / tier);
    if (estimatedPoints <= MAX_DATA_POINTS) {
      intervalInSeconds = tier;
      break;
    }
  }

  // Fallback: if even the largest tier exceeds MAX_DATA_POINTS,
  // calculate a custom interval to ensure we stay under the limit
  const estimatedPointsWithInterval = Math.ceil(durationInSeconds / intervalInSeconds);
  if (estimatedPointsWithInterval > MAX_DATA_POINTS) {
    intervalInSeconds = Math.ceil(durationInSeconds / MAX_DATA_POINTS);
  }

  return getTimeFrameForHoursAgo(dateStart, dateEnd, intervalInSeconds);
}

function getTimeFrameForHoursAgo(dateStart: Date, dateEnd: Date, interval: number): string[] {
  const timeFrames: string[] = [];
  let startTime: Date = new Date(dateStart);

  while (startTime.getTime() <= dateEnd.getTime()) {
    const dateTimeString: string = startTime.toISOString();
    timeFrames.push(dateTimeString);
    startTime = addSeconds(startTime, interval);
  }
  return timeFrames;
}

function arrangeMeasurementsByTime(
  timeFrames: string[],
  measurements: Array<MeasurementResponseValuesInner>,
  options: { useTimeFrameTimestamps?: boolean } = {},
): Array<TMeasurementsTimeAndValue> {
  if (!measurements || measurements.length === 0) {
    return timeFrames.map((timeFrame: string) => [new Date(timeFrame).getTime(), null]);
  }
  const chartData: Array<TMeasurementsTimeAndValue> = [];
  for (let i = 0; i < timeFrames.length - 1; i++) {
    const startTime = new Date(timeFrames[i]).getTime();
    const endTime = new Date(timeFrames[i + 1]).getTime();

    const pointOnCurrTime: Array<MeasurementResponseValuesInner> = measurements.filter(
      (currentPoint: MeasurementResponseValuesInner) => {
        if (!currentPoint.timestamp) return;
        const currentTimestamp = new Date(currentPoint.timestamp).getTime();
        return currentTimestamp >= startTime && currentTimestamp < endTime;
      },
    );

    chartData.push(..._generateChartDataPoint(timeFrames[i], pointOnCurrTime, options));
  }

  // last item
  const startTime = new Date(timeFrames[timeFrames.length - 1]).getTime();
  const pointOnCurrTime: Array<MeasurementResponseValuesInner> = measurements.filter(
    (currentPoint: MeasurementResponseValuesInner) => {
      if (!currentPoint.timestamp) return;
      const currentTimestamp = new Date(currentPoint.timestamp).getTime();
      return currentTimestamp === startTime;
    },
  );

  chartData.push(..._generateChartDataPoint(timeFrames[timeFrames.length - 1], pointOnCurrTime, options));
  return chartData;
}

function _generateChartDataPoint(
  timeFrame: string,
  pointOnCurrTime: MeasurementResponseValuesInner[],
  options: { useTimeFrameTimestamps?: boolean },
): Array<TMeasurementsTimeAndValue> {
  const chartDataPoint: Array<TMeasurementsTimeAndValue> = [];
  if (pointOnCurrTime.length === 0) {
    chartDataPoint.push([new Date(timeFrame).getTime(), null]);
  } else if (options.useTimeFrameTimestamps) {
    const bucketTimestamp = new Date(timeFrame).getTime();
    const nearestPoint = pointOnCurrTime.reduce((nearest, point) => {
      if (!point.timestamp) return nearest;
      if (!nearest?.timestamp) return point;

      const pointDistance = Math.abs(new Date(point.timestamp).getTime() - bucketTimestamp);
      const nearestDistance = Math.abs(new Date(nearest.timestamp).getTime() - bucketTimestamp);
      return pointDistance < nearestDistance ? point : nearest;
    });

    chartDataPoint.push([bucketTimestamp, nearestPoint?.value == null ? null : Number(nearestPoint.value)]);
  } else {
    pointOnCurrTime.forEach((point: MeasurementResponseValuesInner) => {
      if (point.timestamp) {
        chartDataPoint.push([new Date(point.timestamp).getTime(), Number(point.value)]);
      }
    });
  }
  return chartDataPoint;
}

export interface IEChartsXAxisOptions {
  type?: "category" | "value" | "time" | "log";
  data?: string[];
  axisLabel?: Record<string, unknown>;
  axisPointer?: Record<string, unknown>;
  [key: string]: unknown;
}

function getDatetimeOptions(xAxis: IEChartsXAxisOptions): IEChartsXAxisOptions {
  return {
    ...xAxis,
    type: "time",
    axisLabel: {
      formatter: "{dd}/{MM}/{yyyy}\n{HH}:{mm}",
      ...CHART_AXIS_LABEL_DEFAULTS,
      // Let ECharts suppress labels that no longer fit after a drawer resize.
      // This preserves readable time axes without per-drawer breakpoints.
      hideOverlap: true,
    },
  };
}

function getDatetimeAxisLabelOptions<T extends number | string>(
  formatter: (value: T) => string,
): Record<string, unknown> {
  return {
    ...CHART_AXIS_LABEL_DEFAULTS,
    hideOverlap: true,
    formatter,
  };
}

function formatDatetimeAxisLabel(value: number | string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";

  return dateUtil.dateFormat(date, DATE_FORMATS.STANDARD_DATE_TIME).replace(", ", "\n");
}

function assignColorsToKeys<T>(keys: T[]): Map<T, string> {
  const colorsMap = new Map<T, string>();
  let colorIndex = 0;

  keys.forEach((key) => {
    const color: CHART_COLOR_PALETTE = chartColorPaletteArray[colorIndex % chartColorPaletteArray.length];
    colorsMap.set(key, color);
    colorIndex++;
  });

  return colorsMap;
}

function addZoomOptions(
  chartOptions: EChartsOption,
  chartId: string,
  onZoom: (event: IZoomEvent) => void,
  onZoomReset: (chartId: string) => void,
  options?: { showZoomToolbox?: boolean; allowWheelZoom?: boolean },
): EChartsOption {
  const allowWheelZoom = options?.allowWheelZoom === true;

  return {
    ...chartOptions,
    dataZoom: [
      {
        ...INSIDE_DATAZOOM_CONFIG,
        disabled: !allowWheelZoom,
        zoomOnMouseWheel: allowWheelZoom,
        moveOnMouseMove: allowWheelZoom,
        preventDefaultMouseMove: allowWheelZoom,
      },
    ],
    // ECharts' `dataZoomSelect` global cursor is owned by the toolbox dataZoom
    // feature. Keep that feature mounted (off-screen) whenever zoom options are
    // installed so the application-owned controls can activate area selection,
    // including across connected chart groups.
    toolbox: getDashboardChartToolbox({ showZoomToolbox: options?.showZoomToolbox ?? true }),
  };
}

function resetChartDataZoom(chartId: string): void {
  if (typeof document === "undefined") return;

  const chartDom = document.getElementById(chartId);
  if (!chartDom) return;

  const chart = getInstanceByDom(chartDom);
  if (chart) resetChartDataZoomInstance(chart);
}

function resetChartDataZoomInstance(chart: ECharts): void {
  chart.dispatchAction(
    { type: "takeGlobalCursor", key: "dataZoomSelect", dataZoomSelectActive: false },
    { silent: true },
  );
  chart.dispatchAction({ type: "dataZoom", start: 0, end: 100 }, { silent: true });
  chart.getZr().setCursorStyle("default");
}

/**
 * Hide active ECharts tooltips before a scroll container moves. A tooltip that
 * is appended to the document body otherwise retains its pre-scroll position.
 */
function hideTooltipsWithin(container: HTMLElement | null): void {
  if (!container) return;

  container.querySelectorAll<HTMLElement>(".echarts").forEach((chartDom) => {
    getInstanceByDom(chartDom)?.dispatchAction({ type: "hideTip" });
  });
}

function getDashboardChartToolbox(options?: { showZoomToolbox?: boolean }): EChartsOption["toolbox"] {
  if (!options?.showZoomToolbox) {
    return { show: false };
  }
  return {
    show: true,
    right: -9999,
    feature: {
      dataZoom: { yAxisIndex: "none" },
      restore: {},
    },
  };
}

/**
 * Calculates a Highcharts-like Y-axis interval and max value for bar charts.
 * This ensures "nice numbers" and adds top padding so bars don't hit the top of the chart.
 */
function getBarChartYAxisConfig(
  series: Array<{ data?: unknown }>,
  options: { integerTicks?: boolean; targetTicks?: number; topPadding?: boolean } = {},
): { interval: number; max: number } {
  let dataMax = 0;
  for (const s of series) {
    const data = (Array.isArray(s.data) ? s.data : []) as Array<{ value?: number } | number>;
    for (const point of data) {
      const value = typeof point === "number" ? point : (point?.value ?? 0);
      if (value > dataMax) dataMax = value;
    }
  }
  if (dataMax <= 0) return { interval: 1, max: 1 };
  const targetTicks = options.targetTicks ?? 4;
  const rawInterval = dataMax / targetTicks;
  const magnitude = Math.pow(10, Math.floor(Math.log10(rawInterval)));
  const normalized = rawInterval / magnitude;
  let niceMultiplier;
  if (normalized <= 1) niceMultiplier = 1;
  else if (normalized <= 2) niceMultiplier = 2;
  else if (normalized <= 5) niceMultiplier = 5;
  else niceMultiplier = 10;
  const interval = niceMultiplier * magnitude;
  let max = Math.ceil(dataMax / interval) * interval;
  if (options.topPadding !== false && dataMax > 0 && dataMax >= max - interval * 0.2) {
    max += interval;
  }
  if (options.integerTicks) {
    return {
      interval: Math.max(1, Math.ceil(interval)),
      max: Math.max(1, Math.ceil(max)),
    };
  }

  return { interval, max };
}

function getLineChartYAxisConfig(
  dataMax: number | null,
  options: { nearZeroThreshold?: number; targetTicks?: number } = {},
): { min: number; max?: number; interval?: number } {
  const nearZeroThreshold = options.nearZeroThreshold ?? 0;
  if (dataMax === null || !Number.isFinite(dataMax) || dataMax <= 0 || dataMax < nearZeroThreshold) {
    return { min: 0, max: 1, interval: 2 };
  }

  const targetTicks = options.targetTicks ?? 3;
  const rawInterval = dataMax / targetTicks;
  const magnitude = 10 ** Math.floor(Math.log10(rawInterval));
  const normalized = rawInterval / magnitude;
  const multiplier = normalized <= 1 ? 1 : normalized <= 2 ? 2 : normalized <= 2.5 ? 2.5 : normalized <= 5 ? 5 : 10;
  const interval = Number((multiplier * magnitude).toPrecision(12));

  return {
    min: 0,
    max: Number((Math.ceil(dataMax / interval) * interval).toPrecision(12)),
    interval,
  };
}

function formatCompactAxisValue(value: number): string {
  const absoluteValue = Math.abs(value);
  const format = (divisor: number, suffix: string): string => {
    const scaled = value / divisor;
    const rounded = Number(scaled.toFixed(Math.abs(scaled) < 10 ? 1 : 0));
    return `${rounded}${suffix}`;
  };

  if (absoluteValue >= 1_000_000_000) return format(1_000_000_000, "G");
  if (absoluteValue >= 1_000_000) return format(1_000_000, "M");
  if (absoluteValue >= 1_000) return format(1_000, "k");
  return String(value);
}

function getDataZoomRange(event: IDataZoomPayload, timeframes: string[] = []): { min: number; max: number } | null {
  const zoom = event.batch?.[0] || event;
  const minFromValue = Number(zoom.startValue);
  const maxFromValue = Number(zoom.endValue);

  if (Number.isFinite(minFromValue) && Number.isFinite(maxFromValue)) {
    return { min: minFromValue, max: maxFromValue };
  }

  const start = Number(zoom.start);
  const end = Number(zoom.end);
  if (!Number.isFinite(start) || !Number.isFinite(end) || timeframes.length === 0) {
    return null;
  }

  const firstTime = new Date(timeframes[0]).getTime();
  const lastTime = new Date(timeframes[timeframes.length - 1]).getTime();
  const range = lastTime - firstTime;

  return {
    min: firstTime + (range * start) / 100,
    max: firstTime + (range * end) / 100,
  };
}

function getMinimumDataZoomValueSpan(
  timeframes: string[],
  minimumSampleCount = 5,
  minimumValueSpan = 0,
): number | undefined {
  if (timeframes.length < 2 || minimumSampleCount <= 0) return undefined;

  let smallestInterval = Number.POSITIVE_INFINITY;
  for (let index = 1; index < timeframes.length; index++) {
    const previousTimestamp = new Date(timeframes[index - 1]).getTime();
    const currentTimestamp = new Date(timeframes[index]).getTime();
    const interval = currentTimestamp - previousTimestamp;
    if (Number.isFinite(interval) && interval > 0) {
      smallestInterval = Math.min(smallestInterval, interval);
    }
  }

  return Number.isFinite(smallestInterval)
    ? Math.max(smallestInterval * minimumSampleCount, minimumValueSpan)
    : undefined;
}

/**
 * @description Optional settings for customizing the Y-axis.
 * @type {Object}
 * @property {number[]} [steps] - Optional array of tick positions for the Y-axis (e.g., [0, 20, 40, 60, 80]).
 * @property {EYAxisLabelFormat} [type] - The format type for the Y-axis labels, such as 'Memory'.
 * */
export interface IYAxisLabelFormattingOptions {
  steps?: Array<number>;
  type?: EYAxisLabelFormat;
}

export interface IEChartsYAxisOptions {
  type?: "value";
  min?: number;
  name?: string;
  axisLabel?: Record<string, unknown>;
  [key: string]: unknown;
}

/**
 * Updates the Y-axis configuration if all the metrics are zero or it is an empty array.
 * If the type is set to 'Memory', it updates the tick positions and label formatter to display
 * memory units (e.g., Bytes, G).
 */
function updateYAxisLabelsDefaultsIfNeeded(
  measurements: MeasurementResponse[],
  yAxis: IEChartsYAxisOptions,
  options: IYAxisLabelFormattingOptions,
): IEChartsYAxisOptions {
  const yAxisUpdated: IEChartsYAxisOptions = { ...yAxis };
  const filteredMeasurements = measurements.filter((measurement: MeasurementResponse) => {
    return measurement.values && measurement.values.length > 0;
  });

  if (filteredMeasurements.length > 0) return yAxisUpdated;

  const steps = options.steps || [0, 20, 40, 60, 80];

  if (options.type === EYAxisLabelFormat.Memory) {
    yAxisUpdated.min = steps[0];
    yAxisUpdated.max = steps[steps.length - 1];
    yAxisUpdated.interval = steps[1] - steps[0];
    yAxisUpdated.axisLabel = {
      formatter: (value: number) => {
        return value === 0 ? `${value} Bytes` : `${value} G`;
      },
    };
  } else {
    yAxisUpdated.min = steps[0];
    yAxisUpdated.max = steps[steps.length - 1];
    yAxisUpdated.interval = steps[1] - steps[0];
  }

  return yAxisUpdated;
}
