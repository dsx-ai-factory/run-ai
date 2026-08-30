import type {
  GroupedMeasurementResponseValuesInner,
  MetricsResponse,
  MetricsType,
  NodeMetricType,
} from "@/swagger-models/cluster-service-client";
import type { EIntervalLabels } from "@/models/interval.model";
import type { IYAxisLabelFormattingOptions } from "@/utils/chart.util/chart.util";
import type { MetricsCompositeResponse, PodMetricType, WorkloadMetricType } from "@/swagger-models/workloads-client";

export enum CHART_COLOR_PALETTE {
  // Don't change the order of the colors!!!
  // The order of the colors is important for the chart it will be consistent and look nice.
  // Colors are aligned with Figma design system (Graphs colors)
  BLUE = "#2CAFFE",
  ORANGE = "#FE6A35",
  LIME = "#00E272",
  PURPLE = "#544FC5",
  LIGHT_GREEN = "#91E8E1",
  PINK = "#E560FF",
  MUSTARD = "#FFB15A",
  LAVENDER = "#6B8ABC",
  CYAN = "#00E4CA",
}

export const TURQUOISE = "#208EB7";

// Shared chart axis label defaults (matches prod Highcharts axis style).
// Override fontSize per widget when needed (e.g. NIM widgets use 14).
export const CHART_AXIS_LABEL_DEFAULTS = {
  color: "#666666",
  fontSize: 11,
} as const;

// Shared chart legend text style defaults (matches prod Highcharts legend).
// Override fontWeight when widgets want a different weight (e.g. "bold").
export const CHART_LEGEND_TEXT_STYLE_DEFAULTS = {
  color: "#333333",
  fontSize: 11,
  fontWeight: 700,
} as const;

export const CHART_CROSSHAIR_COLOR = "#0000008A";

// UX-approved colors for project quota/allocation cards. Keep these semantic
// names even though some values overlap the generic chart palette.
export const PROJECT_RESOURCE_ALLOCATION_COLORS = {
  overQuota: "#544FC5",
  quota: "#50E991",
  allocated: "#0DB6FF",
} as const;

// Canonical 16x7 filled rectangle used by line-only and rectangular chart legends.
export const LEGEND_RECT_ICON = "rect";
export const LEGEND_RECT_ITEM_STYLE = {
  icon: LEGEND_RECT_ICON,
  itemWidth: 16,
  itemHeight: 7,
} as const;

// Canonical four-stripe icon for dotted/dashed line-series legends.
// Keep the 3px stripe + 2px gap geometry aligned with the custom Resources allocation legend.
export const LEGEND_DASHED_LINE_ICON = "path://M0,0 H3 V8 H0 Z M5,0 H8 V8 H5 Z M10,0 H13 V8 H10 Z M15,0 H18 V8 H15 Z";

export const DRILLDOWN_UNIVERSAL_TRANSITION = { enabled: true, divideShape: "clone" } as const;
export const DRILLDOWN_CHART_ANIMATION_PRESET = {
  animation: true,
  animationDuration: 200,
  animationDurationUpdate: 200,
  animationEasingUpdate: "cubicInOut",
} as const;
export const DRILLDOWN_TRANSITION_PRESET = {
  animation: true,
  animationDuration: 500,
  animationDurationUpdate: 500,
  animationEasing: "sinusoidalInOut",
  animationEasingUpdate: "sinusoidalInOut",
} as const;
export function getDrilldownTransitionPreset(): {
  readonly animation: true;
  readonly animationDuration: number;
  readonly animationDurationUpdate: number;
  readonly animationEasing: "sinusoidalInOut";
  readonly animationEasingUpdate: "sinusoidalInOut";
} {
  const reduceMotion = typeof window !== "undefined" && window.matchMedia?.("(prefers-reduced-motion: reduce)").matches;
  return reduceMotion
    ? { ...DRILLDOWN_TRANSITION_PRESET, animationDuration: 0, animationDurationUpdate: 0 }
    : DRILLDOWN_TRANSITION_PRESET;
}
export const INSIDE_DATAZOOM_CONFIG = {
  type: "inside",
  xAxisIndex: 0,
  filterMode: "none",
  disabled: true,
  zoomOnMouseWheel: false,
  moveOnMouseWheel: false,
  moveOnMouseMove: false,
  preventDefaultMouseMove: false,
} as const;
export enum EChartType {
  COLUMN = "column",
  STACKED_COLUMN = "stacked_column",
  MULTI_BAR = "multi_bar",
}

//dashboard widgets
export enum EWidgetComponent {
  WorkloadStatus = "WorkloadStatusWidget",
  RunningWorkloads = "RunningWorkloadsWidget",
  PendingWorkloads = "PendingWorkloadsWidget",
  FailedWorkloads = "FailedWorkloadsWidget",
  IdleAllocatedGpuWidget = "IdleAllocatedGpuWidget",
  NodePoolFreeResourcesWidget = "NodePoolFreeResourcesWidget",
  WorkloadByTypeWidget = "WorkloadByTypeWidget",
  WorkloadPendingTimeDistributionWidget = "WorkloadPendingTimeDistributionWidget",
  ReadyNodesWidget = "ReadyNodesWidget",
  ReadyGpuWidget = "ReadyGpuWidget",
  AllocatedGpuWidget = "AllocatedGpuWidget",
  GpuAllocationRatioWidget = "GpuAllocationRatioWidget",
  OverviewTimeRangeWidgets = "OverviewTimeRangeWidgets",
  IdleGpuWorkloadWidget = "IdleGpuWorkloadWidget",
  OrgUnitRankingWidget = "OrgUnitRankingWidget",
  WorkloadByCategoryTimeRangeWidget = "WorkloadByCategoryTimeRangeWidget",
  // Quota dashboard widgets
  QuotaTotalGpuWidget = "QuotaTotalGpuWidget",
  AllocatedQuotaGpuWidget = "AllocatedQuotaGpuWidget",
  QuotaAllocationTimeRangeWidget = "QuotaAllocationTimeRangeWidget",
  PendingWorkloadsByNodepoolWidget = "PendingWorkloadsByNodepoolWidget",
  QuotaByNodepoolWidget = "QuotaByNodepoolWidget",
  AllocationByNodepoolWidget = "AllocationByNodepoolWidget",
  DeptAllocationRatioWidget = "DeptAllocationRatioWidget",
  ProjectAllocationRatioWidget = "ProjectAllocationRatioWidget",
}

export enum EWidgetEmptyMessages {
  Default = "To view data, first create workloads",
  NoNodes = "No nodes available",
  NoIdleWorkloads = "No idle workloads to display",
  NoRunningWorkloads = "No running workloads to display",
  NoPendingWorkloads = "No pending workloads to display",
}

export enum EWidgetEntity {
  Department = "Department",
  Project = "Project",
  Nodepool = "Node pool",
  Node = "Node",
  Workload = "Workload",
}

export type TMeasurementsTimeAndValue = [number, number | null];

export const chartColorPaletteArray = Object.entries(CHART_COLOR_PALETTE).map(([, value]) => value);
export const ACTIVE_AXIS_LABEL_COLOR = "#0654fe";
export const CHART_DATA_LABEL_COLOR = "#000000B2";
export const TOOLTIP_BASE_COLOR = "#121212";
export const UNKNOWN_COLOR = CHART_COLOR_PALETTE.PURPLE;

export interface IColumnChartSeries {
  title: string;
  data: number[];
  tooltipTitle?: string;
}

export interface IWidgetWrapperOptions {
  title: string;
  timeFrame?: string;
  linkText?: string;
  entityType?: EWidgetEntity;
  tooltipText?: string;
  emptyMessage?: string;
}

export interface IGenericWidgetWrapperConfig {
  hideBorder?: boolean;
  flat?: boolean;
  hideActions?: boolean;
  hideSeparator?: boolean;
  bodyNoPadding?: boolean;
  options?: IWidgetWrapperOptions;
}

export interface IChartWidgetOptions {
  type: "bar" | "line" | "scatter" | "custom";
  height: number;
  grid?: Record<string, unknown>;
  sharedTooltip?: boolean;
  showSharedCrosshair?: boolean;
  yAxisTitle?: string;
  yAxisFormatFunction?: (value: string | number) => string;
  xAxisIntervalMs?: number;
  colors?: string[];
  showDataLabels?: boolean;
  isPercentage?: boolean;
  highchartsLikeYAxis?: boolean;
  /** Human-readable summary announced by the ECharts SVG renderer. */
  accessibility?: { description?: string; decal?: boolean };
  /** Drawers disable wheel capture so their contents remain scrollable. */
  allowWheelZoom?: boolean;
  /** Reserve plot space for a wrapped, multi-row ECharts legend after each resize. */
  adaptiveLegendLayout?: boolean;
  /** Prevent area zoom from showing fewer than this many canonical samples. */
  minimumZoomSampleCount?: number;
}

export interface ITooltipPoint {
  name: string;
  y: string | number | null | undefined;
  color?: string;
  symbol?: string;
}

export interface IZoomEvent {
  max: number;
  min: number;
  chartId: string;
}

export enum EYAxisLabelFormat {
  Memory,
}

export const GPU_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID = "gpu-allocation-time-range-widget-chart-id";
export const GPU_UTILIZATION_DISTRIBUTION_TIME_RANGE_WIDGET_CHART_ID =
  "gpu-utilization-distribution-time-range-widget-chart-id";
export const RESOURCE_UTILIZATION_TIME_RANGE_WIDGET_CHART_ID = "resource-utilization-time-range-widget-chart-id";
export const RESOURCE_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID = "resource-allocation-time-range-widget-chart-id";
export const WORKLOAD_BY_CATEGORY_TIME_RANGE_WIDGET_CHART_ID = "workload-by-category-time-range-widget-chart-id";
export const QUOTA_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID = "quota-allocation-time-range-widget-chart-id";

/*** * Generic metric widget configuration interface ***/

export type TMetricType = NodeMetricType | WorkloadMetricType | PodMetricType | MetricsType;

export type TGenericMetricGetData = (
  metricTypes: Array<TMetricType>,
  start: string,
  end: string,
  numberOfSamples: number,
  args: string[],
) => Promise<MetricsResponse>;

export type TGenericMetricMeasurementParser = (
  metricsResponse: MetricsResponse,
  averageMetricType?: string,
) => Record<string, Array<GroupedMeasurementResponseValuesInner>>;

export enum EGenericMetricKeySeriesTypeOptions {
  Line = "line",
  Area = "area",
}

export enum EGenericMetricKeySeriesStyleOptions {
  ShortDot = "ShortDot",
  Solid = "Solid",
}

/**
 * @description Configuration interface for generic metric widgets that display time-series data charts.
 * This interface provides all the necessary configuration options to create customizable metric charts
 * with various display options, data sources, and formatting capabilities.
 */
export interface IGenericMetricWidgetConfig {
  /** Unique identifier for the chart instance, used for DOM element identification and event handling */
  chartId: string;

  /** Array of metric types to fetch data for. These correspond to the metrics available in the backend API */
  metricTypes: Array<TMetricType>;

  /** Display title shown at the top of the widget */
  chartTitle: string;

  /**
   * Interval label that determines how frequently the widget refreshes its data
   * @see {@link EIntervalLabels}
   * */
  intervalLabel: EIntervalLabels;

  /**
   * Chart-specific display options including type, height, tooltip behavior, and formatting
   * @see {@link IChartWidgetOptions}
   *  */
  chartOptions: IChartWidgetOptions;

  /**
   * Function to fetch metric data from the backend API with specified time range and sample count
   * @see {@link TGenericMetricGetData}
   * */
  getData: TGenericMetricGetData;

  /** Optional mapping of metric keys to display labels. If not provided, raw metric keys are used */
  keyLabelMap?: Record<string, string>;

  /** Optional suffix to append to tooltip values (e.g., "%", "GB", "requests/sec") */
  tooltipSuffix?: string;

  /** When true, displays Y-axis values as percentages (0-100%) with automatic tick positioning */
  isPercentage?: boolean;

  /** Optional array of fields to group measurements by. Affects how data is aggregated and displayed */
  groupBy?: Array<string>;

  /**
   * Optional custom function to parse and transform raw metric response data into the expected format
   * @see {@link TGenericMetricMeasurementParser}
   * */
  measurementParser?: TGenericMetricMeasurementParser;

  /** Optional metric type to use when calculating and displaying average values across all data points */
  averageMetricType?: string;

  /**
   * Optional configuration for customizing Y-axis label formatting and tick positions
   * For more information, see {@link IYAxisLabelFormattingOptions}
   * */
  yAxisLabelFormatting?: IYAxisLabelFormattingOptions;

  /** Optional mapping of metric keys to display series type. If not provided, line is used as default */
  keySeriesTypeMap?: Record<string, EGenericMetricKeySeriesTypeOptions>;

  /** Optional mapping of metric keys to display series style. If not provided, Solid is used as default */
  keySeriesStyleMap?: Record<string, EGenericMetricKeySeriesStyleOptions>;

  /** Optional series names to render when the API returns no measurement series */
  emptySeriesNames?: string[];

  /** Optional mapping of series names to ECharts legend icons */
  legendIconMap?: Record<string, string>;

  /** Custom ECharts tooltip formatter function */
  formatter?: (params: unknown) => string;

  /**
   * Optional chart widget wrapper config for styling and behavior
   * This can include options like hiding borders, flat design, or custom actions.
   * @see {@link IGenericWidgetWrapperConfig}
   * */
  wrapperProps?: IGenericWidgetWrapperConfig;

  /** Optional value of minimal cluster version supporting widget data */
  minSupportedClusterVersion?: string;

  /** Optional legend text style overrides (fontSize, fontWeight, color, etc.) applied on top of composable defaults */
  legendTextStyle?: Record<string, unknown>;

  /** Keep a parent-owned range unchanged when this chart participates in a connected group. */
  preserveFilterRange?: boolean;

  /** Place every measurement on the parent-generated bucket timestamp for connected hover alignment. */
  useCanonicalTimeBuckets?: boolean;

  /** Add a non-rendering first series so native connected tooltips also resolve all-null buckets. */
  useConnectedTooltipAnchor?: boolean;
}

export const WIDGET_AVERAGE_LINE_LABEL = "Average";

export type THistogramMetricData = Record<string, Array<GroupedMeasurementResponseValuesInner>>;

export type THistogramMetricParser = (metricsResponse: MetricsCompositeResponse) => THistogramMetricData;

export interface IColumnHistogramDataParser {
  name: string;
  data: GroupedMeasurementResponseValuesInner[];
}

export type THistogramMetricGetData = (
  metricTypes: Array<TMetricType>,
  start: string,
  end: string,
  numberOfSamples: number,
  args: string[],
) => Promise<MetricsCompositeResponse>;

export interface IGenericMetricWidgetConfigWithCompositeResponse
  extends Omit<IGenericMetricWidgetConfig, "getData" | "measurementParser"> {
  getData: THistogramMetricGetData;
  histogramParser?: THistogramMetricParser;
  columnHistogramParser?: (metricsResponse: MetricsCompositeResponse) => IColumnHistogramDataParser[];
}
