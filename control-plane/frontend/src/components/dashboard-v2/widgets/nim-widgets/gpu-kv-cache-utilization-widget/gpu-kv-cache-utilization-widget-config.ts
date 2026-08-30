import { EIntervalLabels } from "@/models/interval.model";
import {
  type GroupedMeasurementResponseValuesInner,
  type MetricsResponse,
} from "@/swagger-models/cluster-service-client";
import { type IGenericMetricWidgetConfig, type TMetricType } from "@/models/chart.model";
import { WorkloadMetricType, PodMetricType } from "@/swagger-models/workloads-client";
import { dateUtil } from "@/utils/date.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { podService } from "@/services/control-plane/pod.service/pod.service";
import {
  NIM_LEGEND_TEXT_STYLE,
  NIM_LINE_LEGEND_ICON,
  NIM_METRIC_CHART_GRID,
} from "@/components/dashboard-v2/widgets/nim-widgets/nim-chart-legend.util";

export const GPU_KV_CACHE_UTILIZATION_WIDGET_CHART_ID = "gpu-kv-cache-utilization-widget-id";

const CommonLabelsMap = {
  GpuKvCache: "GPU KV cache",
};

function measurementParser(
  metricsResponse: MetricsResponse,
): Record<string, Array<GroupedMeasurementResponseValuesInner>> {
  const result: Record<string, Array<GroupedMeasurementResponseValuesInner>> = {};

  metricsResponse.measurements.forEach((measurement) => {
    const categoryName = measurement.type;
    result[categoryName] = measurement.values || [];
  });

  return result;
}

const commonConfig: Omit<IGenericMetricWidgetConfig, "metricTypes" | "getData"> = {
  chartTitle: "GPU KV cache utilization",
  intervalLabel: EIntervalLabels.GpuKvCacheUtilizationWidget,
  chartId: GPU_KV_CACHE_UTILIZATION_WIDGET_CHART_ID,
  useCanonicalTimeBuckets: true,
  useConnectedTooltipAnchor: true,
  tooltipSuffix: "%",
  isPercentage: true,
  chartOptions: {
    type: "line",
    height: 250,
    sharedTooltip: true,
    showSharedCrosshair: true,
    yAxisTitle: "",
    grid: NIM_METRIC_CHART_GRID,
  },
  measurementParser,
  yAxisLabelFormatting: {
    steps: [0, 200, 400, 600, 800, 1000],
  },
  emptySeriesNames: [CommonLabelsMap.GpuKvCache],
  legendIconMap: {
    [CommonLabelsMap.GpuKvCache]: NIM_LINE_LEGEND_ICON,
  },
  legendTextStyle: NIM_LEGEND_TEXT_STYLE,
};

export function createWorkloadWidgetConfig(): IGenericMetricWidgetConfig {
  async function getWorkloadRequestData(
    metricTypes: Array<TMetricType>,
    start: string,
    end: string,
    numberOfSamples: number,
    args: string[],
  ): Promise<MetricsResponse> {
    const workloadId = args[0];

    const metricsResponse: MetricsResponse = await workloadService.getWorkloadMetrics(
      workloadId,
      dateUtil.convertDateToISO(new Date(start)),
      dateUtil.convertDateToISO(new Date(end)),
      metricTypes as WorkloadMetricType[],
      numberOfSamples,
    );

    metricsResponse.measurements = metricsResponse.measurements.map((measurement) => {
      return {
        ...measurement,
        type: CommonLabelsMap.GpuKvCache,
        values:
          measurement.values?.map((val) => ({
            ...val,
            value: val.value,
            timestamp: val.timestamp,
          })) || [],
      };
    });

    return metricsResponse;
  }

  return {
    ...commonConfig,
    metricTypes: [WorkloadMetricType.NimGpuCacheUsagePerc],
    getData: getWorkloadRequestData,
  };
}

export function createPodWidgetConfig(): IGenericMetricWidgetConfig {
  async function getPodRequestData(
    metricTypes: Array<TMetricType>,
    start: string,
    end: string,
    numberOfSamples: number,
    args: string[],
  ): Promise<MetricsResponse> {
    const workloadId = args[0];
    const podId = args[1];

    const metricsResponse: MetricsResponse = await podService.getPodMetrics(
      workloadId,
      podId,
      dateUtil.convertDateToISO(new Date(start)),
      dateUtil.convertDateToISO(new Date(end)),
      metricTypes as PodMetricType[],
      numberOfSamples,
    );

    metricsResponse.measurements = metricsResponse.measurements.map((measurement) => {
      return {
        ...measurement,
        type: CommonLabelsMap.GpuKvCache,
        values:
          measurement.values?.map((val) => ({
            ...val,
            value: val.value,
            timestamp: val.timestamp,
          })) || [],
      };
    });

    return metricsResponse;
  }

  return {
    ...commonConfig,
    metricTypes: [PodMetricType.NimGpuCacheUsagePerc],
    getData: getPodRequestData,
  };
}

export const workloadGpuKvCacheUtilizationWidgetConfig = createWorkloadWidgetConfig();
export const podGpuKvCacheUtilizationWidgetConfig = createPodWidgetConfig();
