import { ref, type Ref } from "vue";
import type { EChartsOption } from "echarts";
import type { LineSeriesOption } from "echarts/charts";
import { chartUtil, STANDARD_TOOLTIP_STYLE } from "@/utils/chart.util";
import type { IEChartsXAxisOptions, IEChartsYAxisOptions } from "@/utils/chart.util/chart.util";
import { formatToFixedOrZero } from "@/utils/format.util";
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import { intervalUtil } from "@/utils/interval.util";
import type { ITooltipPoint, IZoomEvent, TMeasurementsTimeAndValue } from "@/models/chart.model";
import type { IGenericMetricWidgetConfigWithCompositeResponse } from "@/models/chart.model";
import type { MeasurementResponseValuesInner } from "@/swagger-models/workloads-client";
import {
  CHART_AXIS_LABEL_DEFAULTS,
  CHART_LEGEND_TEXT_STYLE_DEFAULTS,
  EGenericMetricKeySeriesTypeOptions,
  EGenericMetricKeySeriesStyleOptions,
  WIDGET_AVERAGE_LINE_LABEL,
  LEGEND_DASHED_LINE_ICON,
  LEGEND_RECT_ICON,
  LEGEND_RECT_ITEM_STYLE,
} from "@/models/chart.model";

interface IEChartsSeriesData {
  name: string;
  type: "line" | "bar" | "scatter";
  data: TMeasurementsTimeAndValue[];
  lineStyle?: LineSeriesOption["lineStyle"];
  showSymbol?: LineSeriesOption["showSymbol"];
  symbol?: LineSeriesOption["symbol"];
  symbolSize?: LineSeriesOption["symbolSize"];
  itemStyle?: LineSeriesOption["itemStyle"];
  areaStyle?: LineSeriesOption["areaStyle"];
  emphasis?: LineSeriesOption["emphasis"];
}

const GENERIC_RECT_LEGEND_ICON = LEGEND_RECT_ICON;

function getHighchartsLikeYAxisScale(series: IEChartsSeriesData[]): { max?: number; interval?: number } {
  const dataMax = series.reduce((maxValue, currentSeries) => {
    const seriesMax = currentSeries.data.reduce((currentMax, point) => {
      const value = point[1];
      return typeof value === "number" && value > currentMax ? value : currentMax;
    }, 0);
    return seriesMax > maxValue ? seriesMax : maxValue;
  }, 0);

  // When all data is zero, force a 0-1 range with interval 2 so ECharts renders
  // exactly one tick at "0 Bytes" — matching Highcharts prod behavior.
  if (dataMax <= 0) return { max: 1, interval: 2 };

  const rawInterval = dataMax / 3;
  const magnitude = Math.pow(10, Math.floor(Math.log10(rawInterval)));
  const normalized = rawInterval / magnitude;
  const multiplier = [1, 2.5, 5, 10].find((candidate) => normalized <= candidate) || 10;
  const interval = multiplier * magnitude;

  return {
    interval,
    max: Math.ceil(dataMax / interval) * interval,
  };
}

export function getRoundedDownTime(timestamp: string | number, intervalMs: number): number | undefined {
  const time = typeof timestamp === "number" ? timestamp : new Date(timestamp).getTime();
  if (!Number.isFinite(time)) return undefined;

  return Math.floor(time / intervalMs) * intervalMs;
}

function getRoundedUpTime(timestamp: string | number, intervalMs: number): number | undefined {
  const time = typeof timestamp === "number" ? timestamp : new Date(timestamp).getTime();
  if (!Number.isFinite(time)) return undefined;

  return Math.ceil(time / intervalMs) * intervalMs;
}

export function useGenericMetricWidgetOptions() {
  const chartOptions: Ref<EChartsOption | null> = ref(null);
  const error: Ref<boolean> = ref(false);
  const empty: Ref<boolean> = ref(false);
  const initialLoading: Ref<boolean> = ref(false);
  const timeframes: Ref<string[]> = ref([]);
  const config: Ref<IGenericMetricWidgetConfigWithCompositeResponse | null> = ref(null);
  const isFirstLoad: Ref<boolean> = ref(true);

  const initialize = (widgetConfig: IGenericMetricWidgetConfigWithCompositeResponse): void => {
    config.value = widgetConfig;
  };

  const addBasicOptions = (): void => {
    if (!config.value) return;
    chartOptions.value = chartUtil.getBasicWidgetChartOptions({
      yAxisTitle: config.value.chartOptions.yAxisTitle,
      type: config.value.chartOptions.type,
      height: config.value.chartOptions.height,
      showSharedCrosshair: config.value.chartOptions.showSharedCrosshair,
      sharedTooltip: config.value.chartOptions.sharedTooltip,
      yAxisFormatFunction: config.value.chartOptions.yAxisFormatFunction,
      colors: config.value.chartOptions.colors,
      grid: config.value.chartOptions.grid,
      accessibility: config.value.chartOptions.accessibility,
      allowWheelZoom: config.value.chartOptions.allowWheelZoom,
    });

    const yAxisNameGap = (config.value.chartOptions as typeof config.value.chartOptions & { yAxisNameGap?: number })
      .yAxisNameGap;
    if (yAxisNameGap !== undefined) {
      chartOptions.value.yAxis = {
        ...(chartOptions.value.yAxis as Record<string, unknown>),
        nameGap: yAxisNameGap,
      };
    }

    applyChartTypeSpecificOptions();
  };

  const applyChartTypeSpecificOptions = (): void => {
    if (!chartOptions.value || !config.value) return;

    const chartType = config.value.chartOptions.type;

    if (chartType === "bar") {
      const existingSeries = Array.isArray(chartOptions.value.series) ? chartOptions.value.series : [];
      chartOptions.value.series = existingSeries.map((s) => ({
        ...s,
        type: "bar" as const,
        stack: "total",
        barMaxWidth: 80,
        barGap: "5%",
        barCategoryGap: "5%",
      })) as EChartsOption["series"];
    }
  };

  const addYAxisOptions = (): void => {
    if (!chartOptions.value) return;

    chartOptions.value.yAxis = {
      ...(chartOptions.value.yAxis as IEChartsYAxisOptions),
      axisLabel: {
        ...((chartOptions.value.yAxis as IEChartsYAxisOptions).axisLabel || {}),
        ...CHART_AXIS_LABEL_DEFAULTS,
      },
    } as EChartsOption["yAxis"];

    if (config.value?.isPercentage) {
      chartOptions.value.yAxis = {
        ...(chartOptions.value.yAxis as IEChartsYAxisOptions),
        min: 0,
        max: 100,
        interval: 25,
        axisLabel: {
          ...((chartOptions.value.yAxis as IEChartsYAxisOptions).axisLabel || {}),
          formatter: (value: number) => value + "%",
        },
      } as EChartsOption["yAxis"];
    }
  };

  const addTooltipOptions = (): void => {
    if (!chartOptions.value || !config.value) return;

    if (config.value.formatter) {
      chartOptions.value.tooltip = {
        ...(chartOptions.value.tooltip as Record<string, unknown>),
        ...STANDARD_TOOLTIP_STYLE,
        formatter: config.value.formatter,
      } as unknown as EChartsOption["tooltip"];
    } else {
      const yFormatter = config.value.chartOptions.yAxisFormatFunction || formatToFixedOrZero;
      const tooltipSuffix = config.value.tooltipSuffix;
      chartOptions.value.tooltip = {
        ...(chartOptions.value.tooltip as Record<string, unknown>),
        ...STANDARD_TOOLTIP_STYLE,
        trigger: "axis",
        formatter: (params: Record<string, unknown> | Record<string, unknown>[]) => {
          const items = Array.isArray(params) ? params : [params];
          if (items.length === 0) return "";

          const visibleItems = items.filter((item) => !chartUtil.isConnectedTooltipAnchorSeries(item));

          const points: Array<ITooltipPoint> = visibleItems.map((item) => {
            let pointY = undefined;
            const value = item.value as [number, number | null] | undefined;
            const yVal = Array.isArray(value) ? value[1] : undefined;
            if (yVal !== undefined && yVal !== null) {
              pointY = yFormatter(yVal);
            }
            return {
              name: item.seriesName as string,
              y: pointY,
              color: (item.color as string) || undefined,
              symbol: tooltipSuffix,
            };
          });

          const firstValue = (visibleItems[0] || items[0]).value as [number, number | null] | undefined;
          const xValue = Array.isArray(firstValue) ? firstValue[0] : undefined;
          return chartUtil.formatSharedTooltip(points, xValue);
        },
      } as unknown as EChartsOption["tooltip"];
    }
  };

  const addXAxisOptions = (): void => {
    if (!chartOptions.value || !config.value) return;

    const intervalMs = config.value.chartOptions.xAxisIntervalMs;
    const xAxis = chartUtil.getDatetimeOptions(chartOptions.value.xAxis as IEChartsXAxisOptions);
    const formatXAxisLabel = (value: string | number): string => {
      const timestamp = typeof value === "number" ? value : new Date(value).getTime();
      if (!Number.isFinite(timestamp)) return "";

      const date = new Date(timestamp);
      const pad = (part: number): string => part.toString().padStart(2, "0");
      return `${pad(date.getDate())}/${pad(date.getMonth() + 1)}/${date.getFullYear()}\n${pad(date.getHours())}:${pad(
        date.getMinutes(),
      )}`;
    };

    chartOptions.value.xAxis = {
      ...xAxis,
      // minInterval + maxInterval together enforce an exact tick stride on ECharts time axis.
      // Do NOT add `interval` here — it's treated as a hint on time axes and causes ECharts
      // to fall back to auto-calculation (e.g. 5-min instead of 10-min ticks).
      ...(intervalMs && { minInterval: intervalMs, maxInterval: intervalMs }),
      axisLabel: {
        ...(xAxis.axisLabel || {}),
        ...CHART_AXIS_LABEL_DEFAULTS,
        formatter: formatXAxisLabel,
        hideOverlap: true,
        showMaxLabel: false,
      },
    };
  };

  const addZoomOptions = (onZoom: (event: IZoomEvent) => void, onResetZoom: (chartId: string) => void): void => {
    if (!chartOptions.value || !config.value) return;

    chartOptions.value = chartUtil.addZoomOptions(chartOptions.value, config.value.chartId, onZoom, onResetZoom, {
      allowWheelZoom: config.value.chartOptions.allowWheelZoom,
    });
  };

  const loadChartOptions = (onZoom: (event: IZoomEvent) => void, onResetZoom: (chartId: string) => void): void => {
    addBasicOptions();
    addYAxisOptions();
    addTooltipOptions();
    addXAxisOptions();
    addZoomOptions(onZoom, onResetZoom);
  };

  const processDateRange = (filterByDates: IRangeDates): { updatedDates: IRangeDates; samples: number } => {
    const updatedDates: IRangeDates = config.value?.preserveFilterRange
      ? filterByDates
      : dateUtil.getUpdatedTime(filterByDates);
    timeframes.value = chartUtil.getTimeFrames(updatedDates.dateStart, updatedDates.dateEnd);
    const samples = timeframes.value.length;
    return { updatedDates, samples };
  };

  const handleDataLoadError = (err: unknown): void => {
    error.value = true;
    console.error(err);
  };

  const handleDataLoadSuccess = (onRegister: () => void, isEmpty: boolean): void => {
    error.value = false;
    empty.value = isEmpty;
    onRegister();
    // Mark that first load is complete
    if (isFirstLoad.value) {
      isFirstLoad.value = false;
    }
  };

  const handleDataLoadFinally = (): void => {
    initialLoading.value = false;
    isFirstLoad.value = false;
  };

  const handlePropsChanged = async (onUnregister: () => void, onLoadData: () => Promise<void>): Promise<void> => {
    if (!config.value) return;

    onUnregister();
    // Only show loading on the very first load
    if (isFirstLoad.value) {
      initialLoading.value = true;
    }
    intervalUtil.stopInterval(config.value.intervalLabel);
    await onLoadData();
    intervalUtil.startInterval(config.value.intervalLabel, onLoadData);
  };

  const handleBeforeUnmount = (onUnregister: () => void): void => {
    if (!config.value) return;

    onUnregister();
    intervalUtil.stopInterval(config.value.intervalLabel);
  };

  const createLineSeriesFromMap = (
    measurementMap: Record<string, Array<MeasurementResponseValuesInner>>,
    timeframes: string[],
  ): boolean => {
    if (!chartOptions.value || !config.value) return true;

    const currentConfig = config.value;

    const measurementKeys = Object.keys(measurementMap);
    const seriesKeys = measurementKeys.length > 0 ? measurementKeys : currentConfig.emptySeriesNames || [];
    const shouldRenderEmptyChart = Boolean(currentConfig.emptySeriesNames?.length);

    const series: IEChartsSeriesData[] = seriesKeys.map((key, index) => {
      const metricData: Array<TMeasurementsTimeAndValue> = chartUtil.arrangeMeasurementsByTime(
        timeframes,
        measurementMap[key] || [],
        { useTimeFrameTimestamps: currentConfig.useCanonicalTimeBuckets },
      );
      const seriesName = currentConfig.keyLabelMap?.[key] || key;
      const seriesType = currentConfig.keySeriesTypeMap?.[key] || EGenericMetricKeySeriesTypeOptions.Line;
      const isDashed =
        seriesName === WIDGET_AVERAGE_LINE_LABEL ||
        currentConfig.keySeriesStyleMap?.[key] === EGenericMetricKeySeriesStyleOptions.ShortDot;
      const lineStyle = isDashed ? chartUtil.getDottedLineStyle() : {};
      const lineSeriesStyle: Partial<IEChartsSeriesData> =
        seriesType === EGenericMetricKeySeriesTypeOptions.Line || seriesType === EGenericMetricKeySeriesTypeOptions.Area
          ? chartUtil.getDashboardLineSeriesStyle(index, lineStyle)
          : {};
      const configuredColor = currentConfig.chartOptions.colors?.[index % currentConfig.chartOptions.colors.length];

      const seriesData: IEChartsSeriesData = {
        name: seriesName,
        type: seriesType === EGenericMetricKeySeriesTypeOptions.Area ? "line" : seriesType,
        data: metricData,
        showSymbol: false,
        ...lineSeriesStyle,
        ...(configuredColor && {
          lineStyle: {
            ...(lineSeriesStyle.lineStyle || {}),
            color: configuredColor,
          },
          itemStyle: {
            ...(lineSeriesStyle.itemStyle || {}),
            color: configuredColor,
          },
          emphasis: {
            ...(lineSeriesStyle.emphasis || {}),
            itemStyle: {
              ...((lineSeriesStyle.emphasis?.itemStyle as Record<string, unknown> | undefined) || {}),
              borderColor: configuredColor,
            },
          },
        }),
        ...(seriesType === EGenericMetricKeySeriesTypeOptions.Area && {
          areaStyle: { opacity: 0.2 },
        }),
      };
      return seriesData;
    });

    chartOptions.value.series = (
      currentConfig.useConnectedTooltipAnchor
        ? [chartUtil.getConnectedTooltipAnchorSeries(timeframes), ...series]
        : series
    ) as EChartsOption["series"];
    const minimumZoomSampleCount = currentConfig.chartOptions.minimumZoomSampleCount;
    const minimumZoomValueSpan = minimumZoomSampleCount
      ? chartUtil.getMinimumDataZoomValueSpan(
          timeframes,
          minimumZoomSampleCount,
          currentConfig.chartOptions.xAxisIntervalMs,
        )
      : undefined;
    if (minimumZoomValueSpan !== undefined) {
      const dataZoomOptions = Array.isArray(chartOptions.value.dataZoom)
        ? chartOptions.value.dataZoom
        : chartOptions.value.dataZoom
          ? [chartOptions.value.dataZoom]
          : [];
      chartOptions.value.dataZoom = dataZoomOptions.map((dataZoom) => ({
        ...dataZoom,
        minValueSpan: minimumZoomValueSpan,
      }));
    }
    chartOptions.value.legend = {
      ...(chartOptions.value.legend as Record<string, unknown>),
      bottom: 16,
      ...(currentConfig.chartOptions.adaptiveLegendLayout && {
        left: "center",
        width: "90%",
      }),
      ...LEGEND_RECT_ITEM_STYLE,
      itemGap: 20,
      textStyle: {
        ...((chartOptions.value.legend as Record<string, unknown>)?.textStyle as Record<string, unknown>),
        ...CHART_LEGEND_TEXT_STYLE_DEFAULTS,
        // Allow per-widget overrides (e.g. NIM widgets need fontSize 12 / fontWeight "bold").
        ...(currentConfig.legendTextStyle || {}),
      },
      data: series.map((s) => {
        const fallbackIcon = chartUtil.isDottedLineStyle(s.lineStyle)
          ? LEGEND_DASHED_LINE_ICON
          : GENERIC_RECT_LEGEND_ICON;
        return {
          name: s.name,
          icon: currentConfig.legendIconMap?.[s.name] || fallbackIcon,
        };
      }),
    };

    if (currentConfig.chartOptions.highchartsLikeYAxis) {
      chartOptions.value.yAxis = {
        ...(chartOptions.value.yAxis as Record<string, unknown>),
        min: 0,
        ...getHighchartsLikeYAxisScale(series),
      };
    }

    const intervalMs = currentConfig.chartOptions.xAxisIntervalMs;
    const roundedXAxisMin =
      intervalMs && timeframes.length > 0 ? getRoundedDownTime(timeframes[0], intervalMs) : undefined;
    // Also snap the max to the next interval boundary so the visible range is a clean
    // multiple of the tick interval. This prevents ECharts from rendering denser ticks
    // (e.g. 5-min instead of 10-min) when the raw range doesn't divide evenly.
    const roundedXAxisMax =
      intervalMs && timeframes.length > 0 ? getRoundedUpTime(timeframes[timeframes.length - 1], intervalMs) : undefined;

    if (roundedXAxisMin !== undefined || roundedXAxisMax !== undefined) {
      chartOptions.value.xAxis = {
        ...(chartOptions.value.xAxis as Record<string, unknown>),
        ...(roundedXAxisMin !== undefined && { min: roundedXAxisMin }),
        ...(roundedXAxisMax !== undefined && { max: roundedXAxisMax }),
      };
    }

    if (shouldRenderEmptyChart && timeframes.length > 0) {
      chartOptions.value.xAxis = {
        ...(chartOptions.value.xAxis as Record<string, unknown>),
        max: new Date(timeframes[timeframes.length - 1]).getTime(),
      };
    }

    const isEmpty =
      !shouldRenderEmptyChart && (series.length === 0 || series.every((s) => !s.data || s.data.length === 0));
    return isEmpty;
  };

  const loadData = async <T>(
    filterByDates: IRangeDates,
    getDataArgs: string[],
    onRegister: () => void,
    onDataLoaded: (metricsResponse: T) => boolean,
  ): Promise<void> => {
    if (chartOptions.value === null || !config.value) return;

    try {
      const { updatedDates, samples } = processDateRange(filterByDates);

      const metricsResponse = await config.value.getData(
        config.value.metricTypes,
        dateUtil.convertDateToISO(updatedDates.dateStart),
        dateUtil.convertDateToISO(updatedDates.dateEnd),
        samples,
        getDataArgs,
      );

      const isEmpty = onDataLoaded(metricsResponse as T);
      handleDataLoadSuccess(onRegister, isEmpty);
    } catch (error: unknown) {
      handleDataLoadError(error);
    } finally {
      handleDataLoadFinally();
    }
  };

  return {
    chartOptions,
    error,
    empty,
    initialLoading,
    timeframes,
    initialize,
    loadChartOptions,
    loadData,
    createLineSeriesFromMap,
    updateYAxisLabelsDefaultsIfNeeded: chartUtil.updateYAxisLabelsDefaultsIfNeeded,
    handlePropsChanged,
    handleBeforeUnmount,
  };
}
