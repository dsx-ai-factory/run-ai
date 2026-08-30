import { EIntervalLabels } from "@/models/interval.model";
import {
  MetricsType,
  type GroupedMetricsResponse,
  type GroupedMeasurementResponseValuesInner,
  type MetricsResponse,
} from "@/swagger-models/cluster-service-client";
import { OrgUnitMetricType } from "@/swagger-models/org-unit-service-client";
import {
  type IGenericMetricWidgetConfig,
  WORKLOAD_BY_CATEGORY_TIME_RANGE_WIDGET_CHART_ID,
  type IGenericWidgetWrapperConfig,
  EWidgetEmptyMessages,
} from "@/models/chart.model";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { EOrgUnitMetricSourceType } from "@/models/org-unit.model";

function _convertGroupedMetricsToMetricsResponse(response: GroupedMetricsResponse): MetricsResponse {
  return {
    measurements: response.measurements.flatMap((measurement) => {
      if (measurement.groups && measurement.values) {
        return measurement.groups.map((group) => ({
          type: group.value,
          labels: measurement.labels,
          values: measurement.values
            ? measurement?.values.map((val) => ({
                value: val.value,
                timestamp: val.timestamp,
              }))
            : [],
        }));
      }
      return [];
    }),
  };
}

const workloadByCategoryGetData = async (
  metricType: string[],
  start: string,
  end: string,
  numberOfSamples: number,
  args: string[],
): Promise<MetricsResponse> => {
  const clusterId = args[0];
  const nodePoolName = args[1] || undefined;
  const orgUnitType = args[2];
  const orgUnitId = args[3];

  if (orgUnitType && orgUnitId) {
    const fetcher =
      orgUnitType === EOrgUnitMetricSourceType.Department
        ? orgUnitService.getDepartmentMetrics
        : orgUnitService.getProjectMetrics;
    return (await fetcher(
      orgUnitId,
      [OrgUnitMetricType.WorkloadsCount],
      start,
      end,
      numberOfSamples,
      nodePoolName,
      "Category",
    )) as unknown as MetricsResponse;
  }

  const response: GroupedMetricsResponse = await clusterService.getClusterGroupedMetrics(
    clusterId,
    start,
    end,
    metricType as MetricsType[],
    numberOfSamples,
    "Category",
    nodePoolName,
  );

  return _convertGroupedMetricsToMetricsResponse(response);
};

const workloadByCategoryMeasurementParser = (
  metricsResponse: MetricsResponse,
): Record<string, Array<GroupedMeasurementResponseValuesInner>> => {
  const result: Record<string, Array<GroupedMeasurementResponseValuesInner>> = {};

  metricsResponse.measurements.forEach((measurement) => {
    const categoryName = measurement.type;
    result[categoryName] = measurement.values || [];
  });

  return result;
};

export const workloadByCategoryWidgetConfig: IGenericMetricWidgetConfig = {
  metricTypes: [MetricsType.WorkloadsCount],
  chartTitle: "Workloads by category",
  keyLabelMap: {
    [OrgUnitMetricType.WorkloadsCount]: "Workloads",
  },
  intervalLabel: EIntervalLabels.WorkloadByCategoryTimeRangeWidget,
  chartId: WORKLOAD_BY_CATEGORY_TIME_RANGE_WIDGET_CHART_ID,
  preserveFilterRange: true,
  useCanonicalTimeBuckets: true,
  useConnectedTooltipAnchor: true,
  getData: workloadByCategoryGetData,
  measurementParser: workloadByCategoryMeasurementParser,
  chartOptions: {
    type: "line",
    height: 250,
    grid: {
      top: 30,
      bottom: 85,
      left: 54,
      right: 20,
      containLabel: false,
    },
    xAxisIntervalMs: 12 * 60 * 60 * 1000,
    sharedTooltip: true,
    showSharedCrosshair: true,
    yAxisTitle: "Workloads",
  },
  wrapperProps: {
    hideBorder: false,
    flat: true,
    hideActions: true,
    hideSeparator: false,
    bodyNoPadding: false,
    options: {
      tooltipText: "Distribution of workloads by category over time to help monitor usage trends across the cluster",
      emptyMessage: EWidgetEmptyMessages.Default,
    },
  } as IGenericWidgetWrapperConfig,
};
