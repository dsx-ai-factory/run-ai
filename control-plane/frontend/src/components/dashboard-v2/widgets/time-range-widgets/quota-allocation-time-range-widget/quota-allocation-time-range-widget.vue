<template>
  <chart-widget-wrapper
    class="quota-allocation-time-range-widget"
    :options="wrapperOptions"
    body-no-padding
    :loading="isInitialLoading"
    :error="error"
    hide-actions
    size="auto"
  >
    <template #header="{ titleId }">
      <div class="widget-header row items-center q-py-xs q-pl-md q-pr-sm">
        <div class="title-tooltip-container">
          <h3 :id="titleId" class="chart-widget-title text-b-main">Over time allocation / quota</h3>
          <runai-tooltip-wrapper use-html :tooltip-text="wrapperOptions.tooltipText ?? ''" class="q-ml-xs">
            <template #tooltip-content>
              <div v-dompurify-html="wrapperOptions.tooltipText" />
            </template>
            <q-icon class="text-icon-default" :name="helpIcon" size="20px" />
          </runai-tooltip-wrapper>
        </div>
        <runai-datepicker-with-predefined
          aid="quota-allocation-date-picker"
          @changed="onDateRangeChanged"
          :range-of-dates="rangeOfDatesCallback"
          info-bar-message="Select a period of up to 30 days within the last 90 days"
          :predefined-options="predefinedOptions"
          :default-option-index="defaultOptionIndex"
        />
      </div>
    </template>
    <chart-area-zoom-controls @zoom-toggle="handleZoomToggle" @reset="handleAreaZoomReset" />
    <v-chart
      class="quota-allocation-time-range-chart"
      v-if="chartOptions"
      ref="chartRef"
      :option="chartOptions"
      autoresize
      :id="CHART_ID"
      @restore="handleResetZoom"
    />
  </chart-widget-wrapper>
</template>

<script setup lang="ts">
import "@/plugins/echarts";
import { ref, computed, watch, onUnmounted, onMounted } from "vue";
import VChart from "vue-echarts";
import type { EChartsOption } from "echarts";

// components
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { ChartAreaZoomControls } from "@/components/common/charts/chart-area-zoom-controls";
import { RunaiDatepickerWithPredefined } from "@/components/common/runai-datepicker-with-predefined";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";

// models
import {
  CHART_AXIS_LABEL_DEFAULTS,
  CHART_LEGEND_TEXT_STYLE_DEFAULTS,
  LEGEND_RECT_ITEM_STYLE,
  type IWidgetWrapperOptions,
  type ITooltipPoint,
  type TMeasurementsTimeAndValue,
  QUOTA_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID as CHART_ID,
  chartColorPaletteArray,
  INSIDE_DATAZOOM_CONFIG,
} from "@/models/chart.model";
import {
  MetricsType,
  type MetricsResponse,
  type GroupedMeasurementResponseValuesInner,
} from "@/swagger-models/cluster-service-client";
import {
  OrgUnitMetricType,
  type MetricsResponse as OrgUnitMetricsResponse,
} from "@/swagger-models/org-unit-service-client";
import { EIntervalLabels } from "@/models/interval.model";
import type { PredefinedOption } from "@/models/date.model";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

// services
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";

// utils
import { chartUtil, STANDARD_TOOLTIP_STYLE } from "@/utils/chart.util";
import { metricUtil } from "@/utils/metric.util";
import { dateUtil, extendedHoursPredefinedOptions, type IRangeDates, TimeUnit } from "@/utils/date.util";
import { roundToDecimal } from "@/utils/format.util";
import { intervalUtil } from "@/utils/interval.util";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";
import {
  getOverviewShortRangeAxisInterval,
  TEN_MINUTES_IN_MS,
} from "@/components/dashboard-v2/widgets/time-range-widgets/overview-time-axis.util";

const helpIcon = toQuasarIconName(runaiIcons.object.help);

export interface Props {
  clusterId: string;
  nodePoolName?: string;
  orgUnitMetricSource?: IOrgUnitMetricSource | null;
}

const props = withDefaults(defineProps<Props>(), {
  orgUnitMetricSource: null,
});

const defaultOptionIndex = 6;
const predefinedOptions: PredefinedOption[] = extendedHoursPredefinedOptions;

const TOOLTIP_TEXT =
  "Allocated: The number of GPUs currently used by workloads over time<br>Quota: The total GPU quota assigned to subordinate departments or projects over time";
const TWO_HOURS_IN_MS = 2 * 60 * 60 * 1000;
const ONE_HOUR_IN_MS = 60 * 60 * 1000;
const ONE_DAY_IN_MS = 24 * ONE_HOUR_IN_MS;

const wrapperOptions = computed<IWidgetWrapperOptions>(() => ({
  title: "Over time allocation / quota",
  tooltipText: TOOLTIP_TEXT,
}));

const defaultDateEnd = new Date();
const defaultDateStart = dateUtil.adjustDateBy(TimeUnit.day, defaultDateEnd, -7);
const filterByDates = ref<IRangeDates>({
  dateStart: defaultDateStart,
  dateEnd: defaultDateEnd,
  label: "Last 7 days",
});
const chartRef = ref<InstanceType<typeof VChart> | null>(null);
const chartOptions = ref<EChartsOption | null>(null);
const measurementMap = ref<Record<string, Array<GroupedMeasurementResponseValuesInner>>>({});
const error = ref(false);
const isInitialLoading = ref(true);
const timeframes = ref<string[]>([]);

function getXAxisOptions(): EChartsOption["xAxis"] {
  const xAxis = chartUtil.getDatetimeOptions({});
  const duration = filterByDates.value.dateEnd.getTime() - filterByDates.value.dateStart.getTime();
  const isShortRange = duration <= TWO_HOURS_IN_MS;
  const isHourlyRange = duration > TWO_HOURS_IN_MS && duration <= ONE_DAY_IN_MS;
  const interval = isShortRange
    ? getOverviewShortRangeAxisInterval(filterByDates.value.label)
    : isHourlyRange
      ? ONE_HOUR_IN_MS
      : undefined;

  return {
    ...xAxis,
    splitNumber: 14,
    minInterval: interval ?? TEN_MINUTES_IN_MS,
    min: filterByDates.value.dateStart.getTime(),
    max: filterByDates.value.dateEnd.getTime(),
    ...(interval ? { interval, maxInterval: interval } : {}),
    axisLabel: {
      ...xAxis.axisLabel,
      ...CHART_AXIS_LABEL_DEFAULTS,
      hideOverlap: true,
      formatter: chartUtil.formatDatetimeAxisLabel,
    },
    axisLine: {
      lineStyle: {
        color: "#ccd6eb",
      },
    },
    axisTick: {
      show: false,
    },
    splitLine: {
      show: false,
    },
  };
}

function loadChartOptions(): void {
  const baseOptions = chartUtil.getBasicWidgetChartOptions({
    yAxisTitle: "GPU",
    type: "line",
    height: 250,
    showSharedCrosshair: true,
    sharedTooltip: true,
  });

  chartOptions.value = {
    ...baseOptions,
    color: chartColorPaletteArray,
    grid: { left: 50, right: 20, top: 30, bottom: 78 },
    xAxis: getXAxisOptions(),
    yAxis: {
      ...(baseOptions.yAxis as Record<string, unknown>),
      splitNumber: 2,
      nameTextStyle: { color: "#666666", fontSize: 12, fontFamily: "Roboto, sans-serif" },
      axisLabel: { ...CHART_AXIS_LABEL_DEFAULTS, formatter: chartUtil.formatCompactAxisValue },
      splitLine: {
        lineStyle: {
          type: "solid",
          color: "#e6e6e6",
        },
      },
    },
    tooltip: {
      trigger: "axis",
      ...STANDARD_TOOLTIP_STYLE,
      axisPointer: {
        ...chartUtil.getConnectedChartAxisPointer(),
      },
      formatter: (params: unknown) => {
        const items = params as Array<{ seriesName: string; value: [number, number | null]; color: string }>;
        if (!items || !Array.isArray(items) || items.length === 0) return "";
        const UNLIMITED_VALUE = -1;
        const points: Array<ITooltipPoint> = items.map((item) => ({
          name: item.seriesName,
          y:
            item.value[1] === UNLIMITED_VALUE
              ? "Unlimited"
              : item.value[1] != null
                ? roundToDecimal(item.value[1])
                : item.value[1],
          color: item.color,
        }));
        return chartUtil.formatSharedTooltip(points, items[0].value[0]);
      },
    },
    legend: {
      show: true,
      left: "center",
      bottom: 16,
      ...LEGEND_RECT_ITEM_STYLE,
      itemGap: 20,
      textStyle: { ...CHART_LEGEND_TEXT_STYLE_DEFAULTS },
      data: ["Allocation", "Quota"],
    },
    dataZoom: [{ ...INSIDE_DATAZOOM_CONFIG }],
    toolbox: chartUtil.getDashboardChartToolbox({ showZoomToolbox: true }),
    series: [],
  };
}

async function loadChartData(): Promise<void> {
  if (!chartOptions.value) return;

  try {
    timeframes.value = chartUtil.getTimeFrames(filterByDates.value.dateStart, filterByDates.value.dateEnd);
    const metricsResponse = await getMetrics();

    measurementMap.value = metricUtil.mapMeasurementsToMap(metricsResponse);
    error.value = false;
    setSeries();
  } catch (err: unknown) {
    error.value = true;
    console.error(err);
  } finally {
    isInitialLoading.value = false;
  }
}

async function getMetrics(): Promise<MetricsResponse | OrgUnitMetricsResponse> {
  const dates = filterByDates.value;
  const samples = timeframes.value.length;
  const start = dateUtil.convertDateToISO(dates.dateStart);
  const end = dateUtil.convertDateToISO(dates.dateEnd);

  if (props.orgUnitMetricSource) {
    const orgUnitMetricTypes: Array<OrgUnitMetricType> = [OrgUnitMetricType.GpuAllocation, OrgUnitMetricType.GpuQuota];
    const fetcher =
      props.orgUnitMetricSource.type === EOrgUnitMetricSourceType.Department
        ? orgUnitService.getDepartmentMetrics
        : orgUnitService.getProjectMetrics;
    return await fetcher(props.orgUnitMetricSource.id, orgUnitMetricTypes, start, end, samples, props.nodePoolName);
  }

  const metricTypes: Array<MetricsType> = [MetricsType.AllocatedGpu, MetricsType.GpuQuota];
  if (props.nodePoolName) {
    return clusterService.getNodepoolMetrics(props.clusterId, props.nodePoolName, start, end, metricTypes, samples);
  }
  return clusterService.getClusterMetrics(props.clusterId, start, end, metricTypes, samples);
}

function setSeries(): void {
  if (!chartOptions.value) return;

  const allocationKey = props.orgUnitMetricSource ? OrgUnitMetricType.GpuAllocation : MetricsType.AllocatedGpu;
  const quotaKey = props.orgUnitMetricSource ? OrgUnitMetricType.GpuQuota : MetricsType.GpuQuota;

  const allocatedGpus: Array<TMeasurementsTimeAndValue> = chartUtil.arrangeMeasurementsByTime(
    timeframes.value,
    measurementMap.value[allocationKey],
  );

  const gpuQuota: Array<TMeasurementsTimeAndValue> = chartUtil.arrangeMeasurementsByTime(
    timeframes.value,
    measurementMap.value[quotaKey],
  );

  chartOptions.value = {
    ...chartOptions.value,
    series: [
      {
        name: "Allocation",
        type: "line",
        data: allocatedGpus,
        color: chartColorPaletteArray[0],
        showSymbol: false,
        ...chartUtil.getConnectedChartHoverMarker(chartColorPaletteArray[0]),
        lineStyle: { width: 2 },
      },
      {
        name: "Quota",
        type: "line",
        data: gpuQuota,
        color: chartColorPaletteArray[1],
        showSymbol: false,
        ...chartUtil.getConnectedChartHoverMarker(chartColorPaletteArray[1]),
        lineStyle: { width: 2 },
      },
    ],
  };
}

function handleResetZoom(): void {
  // The restore event from toolbox already resets the chart internally
}

function handleZoomToggle(active: boolean): void {
  const vchart = chartRef.value;
  if (!vchart?.chart) return;
  vchart.chart.dispatchAction({
    type: "takeGlobalCursor",
    key: "dataZoomSelect",
    dataZoomSelectActive: active,
  });
}
function handleAreaZoomReset(): void {
  const vchart = chartRef.value;
  if (!vchart?.chart) return;
  chartUtil.resetChartDataZoomInstance(vchart.chart);
}

function onDateRangeChanged(dateRange: Array<Date>, label: string): void {
  filterByDates.value = {
    dateStart: dateRange[0],
    dateEnd: dateRange[1],
    label,
  };
}

function rangeOfDatesCallback(date: string): boolean {
  const dateToCheck = new Date(date);
  if (Number.isNaN(dateToCheck.getTime())) return false;

  dateToCheck.setHours(0, 0, 0, 0);
  const back90Days = dateUtil.adjustDateBy(TimeUnit.day, new Date(), -90);
  back90Days.setHours(0, 0, 0, 0);

  const today = new Date();
  today.setHours(23, 59, 59, 999);

  return dateToCheck >= back90Days && dateToCheck <= today;
}

watch(filterByDates, () => {
  isInitialLoading.value = true;
  if (chartOptions.value) {
    chartOptions.value = {
      ...chartOptions.value,
      xAxis: getXAxisOptions(),
    };
  }
  loadChartData();
});

watch(
  () => [props.clusterId, props.nodePoolName],
  () => {
    isInitialLoading.value = true;
    loadChartData();
  },
);

watch(
  () => props.orgUnitMetricSource,
  () => {
    isInitialLoading.value = true;
    loadChartData();
  },
  { deep: true },
);

onMounted(() => {
  loadChartOptions();
  loadChartData();
  intervalUtil.startInterval(EIntervalLabels.QuotaAllocationTimeRangeWidget, loadChartData);
});

onUnmounted(() => {
  intervalUtil.stopInterval(EIntervalLabels.QuotaAllocationTimeRangeWidget);
});

defineExpose({
  chartOptions,
  timeframes,
  filterByDates,
  loadChartData,
});
</script>

<style scoped lang="scss">
.chart-widget-title {
  margin: 0;
  font: inherit;
  letter-spacing: inherit;
}

.widget-header {
  height: 50px;
}
.title-tooltip-container {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  margin-right: auto;
}
.quota-allocation-time-range-chart {
  width: 100%;
  height: 250px;
}
</style>
