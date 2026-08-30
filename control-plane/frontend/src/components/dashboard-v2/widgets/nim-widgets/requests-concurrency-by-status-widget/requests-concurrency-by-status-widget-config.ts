import { EIntervalLabels } from "@/models/interval.model";
import {
  type GroupedMeasurementResponseValuesInner,
  type MetricsResponse,
} from "@/swagger-models/cluster-service-client";
import {
  type IGenericMetricWidgetConfig,
  type TMetricType,
  EGenericMetricKeySeriesTypeOptions,
  EGenericMetricKeySeriesStyleOptions,
} from "@/models/chart.model";
import { WorkloadMetricType, PodMetricType } from "@/swagger-models/workloads-client";
import { dateUtil } from "@/utils/date.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { podService } from "@/services/control-plane/pod.service/pod.service";
import {
  NIM_CIRCLE_LEGEND_ICON,
  NIM_LEGEND_TEXT_STYLE,
  NIM_METRIC_CHART_GRID,
  NIM_VERTICAL_BARS_LEGEND_ICON,
} from "@/components/dashboard-v2/widgets/nim-widgets/nim-chart-legend.util";
import {
  getNimWidgetChartOptions,
  NIM_WIDGET_Y_AXIS_LABEL_STEPS,
} from "@/components/dashboard-v2/widgets/chart-widgets/inference-chart-options.util";

export const REQUESTS_CONCURRENCY_BY_STATUS_WIDGET_CHART_ID = "request-concurrency-by-status-widget-chart-id";

const CommonLabelsMap = {
  Running: "Running",
  Waiting: "Waiting",
  Max: "Max requests allowed",
};

const WorkloadLabelsMap = {
  [WorkloadMetricType.NimNumRequestsRunning]: CommonLabelsMap.Running,
  [WorkloadMetricType.NimNumRequestsWaiting]: CommonLabelsMap.Waiting,
  [WorkloadMetricType.NimNumRequestMax]: CommonLabelsMap.Max,
};

const PodLabelsMap = {
  [PodMetricType.NimNumRequestsRunning]: CommonLabelsMap.Running,
  [PodMetricType.NimNumRequestsWaiting]: CommonLabelsMap.Waiting,
  [PodMetricType.NimNumRequestMax]: CommonLabelsMap.Max,
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
  chartTitle: "Request concurrency by status",
  intervalLabel: EIntervalLabels.RequestsConcurrencyByStatusWidget,
  chartId: REQUESTS_CONCURRENCY_BY_STATUS_WIDGET_CHART_ID,
  useCanonicalTimeBuckets: true,
  useConnectedTooltipAnchor: true,
  measurementParser,
  chartOptions: {
    ...getNimWidgetChartOptions("Requests"),
    grid: NIM_METRIC_CHART_GRID,
  },
  yAxisLabelFormatting: {
    steps: NIM_WIDGET_Y_AXIS_LABEL_STEPS,
  },
  keySeriesTypeMap: {
    [CommonLabelsMap.Running]: EGenericMetricKeySeriesTypeOptions.Line,
    [CommonLabelsMap.Waiting]: EGenericMetricKeySeriesTypeOptions.Line,
    [CommonLabelsMap.Max]: EGenericMetricKeySeriesTypeOptions.Line,
  },
  keySeriesStyleMap: {
    [CommonLabelsMap.Max]: EGenericMetricKeySeriesStyleOptions.ShortDot,
  },
  emptySeriesNames: [CommonLabelsMap.Running, CommonLabelsMap.Waiting, CommonLabelsMap.Max],
  legendIconMap: {
    [CommonLabelsMap.Running]: NIM_CIRCLE_LEGEND_ICON,
    [CommonLabelsMap.Waiting]: NIM_CIRCLE_LEGEND_ICON,
    [CommonLabelsMap.Max]: NIM_VERTICAL_BARS_LEGEND_ICON,
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
        type: WorkloadLabelsMap[measurement.type.toString() as keyof typeof WorkloadLabelsMap] || measurement.type,
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
    metricTypes: [
      WorkloadMetricType.NimNumRequestsRunning,
      WorkloadMetricType.NimNumRequestsWaiting,
      WorkloadMetricType.NimNumRequestMax,
    ],
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
        type: PodLabelsMap[measurement.type.toString() as keyof typeof PodLabelsMap] || measurement.type,
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
    metricTypes: [
      PodMetricType.NimNumRequestsRunning,
      PodMetricType.NimNumRequestsWaiting,
      PodMetricType.NimNumRequestMax,
    ],
    getData: getPodRequestData,
  };
}

export const workloadWidgetConfig = createWorkloadWidgetConfig();
export const podWidgetConfig = createPodWidgetConfig();
