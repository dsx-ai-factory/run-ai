/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { EIntervalLabels } from "@/models/interval.model";
import {
  type GroupedMeasurementResponseValuesInner,
  type MetricsResponse,
} from "@/swagger-models/cluster-service-client";
import {
  type IGenericMetricWidgetConfig,
  type TMetricType,
  EGenericMetricKeySeriesTypeOptions,
  type ITooltipPoint,
} from "@/models/chart.model";
import { WorkloadMetricType, PodMetricType } from "@/swagger-models/workloads-client";
import { dateUtil } from "@/utils/date.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { podService } from "@/services/control-plane/pod.service/pod.service";
import { formatToFixedOrZero } from "@/utils/format.util/format.util";
import { chartUtil } from "@/utils/chart.util";
import {
  NIM_CIRCLE_LEGEND_ICON,
  NIM_LEGEND_TEXT_STYLE,
  NIM_METRIC_CHART_GRID,
} from "@/components/dashboard-v2/widgets/nim-widgets/nim-chart-legend.util";
import {
  getNimWidgetChartOptions,
  NIM_WIDGET_Y_AXIS_LABEL_STEPS,
} from "@/components/dashboard-v2/widgets/chart-widgets/inference-chart-options.util";

export const THROUGHPUT_BY_STATUS_WIDGET_CHART_ID = "throughput-by-status-widget-id";

const CommonLabelsMap = {
  Success: " Succeeded",
  Failure: "Failed",
};

const WorkloadLabelsMap = {
  [WorkloadMetricType.NimRequestSuccessTotal]: CommonLabelsMap.Success,
  [WorkloadMetricType.NimRequestFailureTotal]: CommonLabelsMap.Failure,
};

const PodLabelsMap = {
  [PodMetricType.NimRequestSuccessTotal]: CommonLabelsMap.Success,
  [PodMetricType.NimRequestFailureTotal]: CommonLabelsMap.Failure,
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
  chartTitle: "Request count by status",
  intervalLabel: EIntervalLabels.ThroughputByStatusWidget,
  chartId: THROUGHPUT_BY_STATUS_WIDGET_CHART_ID,
  useCanonicalTimeBuckets: true,
  useConnectedTooltipAnchor: true,
  chartOptions: {
    ...getNimWidgetChartOptions("Requests"),
    grid: NIM_METRIC_CHART_GRID,
  },
  measurementParser,
  yAxisLabelFormatting: {
    steps: NIM_WIDGET_Y_AXIS_LABEL_STEPS,
  },
  keySeriesTypeMap: {
    [CommonLabelsMap.Success]: EGenericMetricKeySeriesTypeOptions.Line,
    [CommonLabelsMap.Failure]: EGenericMetricKeySeriesTypeOptions.Line,
  },
  emptySeriesNames: [CommonLabelsMap.Success, CommonLabelsMap.Failure],
  legendIconMap: {
    [CommonLabelsMap.Success]: NIM_CIRCLE_LEGEND_ICON,
    [CommonLabelsMap.Failure]: NIM_CIRCLE_LEGEND_ICON,
  },
  legendTextStyle: NIM_LEGEND_TEXT_STYLE,
  formatter: (params: unknown) => {
    const items = Array.isArray(params) ? (params as Record<string, unknown>[]) : [params as Record<string, unknown>];
    const visibleItems = items.filter((item) => !chartUtil.isConnectedTooltipAnchorSeries(item));
    if (visibleItems.length === 0) return "";

    let total = 0;
    const points: Array<ITooltipPoint> = visibleItems.map((item) => {
      let pointY = undefined;
      const value = item.value as [number, number | null] | undefined;
      const yVal = Array.isArray(value) ? value[1] : undefined;
      if (yVal !== undefined && yVal !== null) {
        pointY = formatToFixedOrZero(yVal);
        total += pointY;
      }
      return {
        name: item.seriesName as string,
        y: pointY,
        color: (item.color as string) || undefined,
      };
    });

    const firstValue = visibleItems[0].value as [number, number | null] | undefined;
    const xValue = Array.isArray(firstValue) ? firstValue[0] : undefined;
    let tooltip = chartUtil.formatSharedTooltip(points, xValue);

    tooltip += chartUtil.formatTooltipValue("Total", undefined, total, "", true);
    return tooltip;
  },
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
    metricTypes: [WorkloadMetricType.NimRequestSuccessTotal, WorkloadMetricType.NimRequestFailureTotal],
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
    metricTypes: [PodMetricType.NimRequestSuccessTotal, PodMetricType.NimRequestFailureTotal],
    getData: getPodRequestData,
  };
}

export const workloadThroughputByStatusWidgetConfig = createWorkloadWidgetConfig();
export const podThroughputByStatusWidgetConfig = createPodWidgetConfig();
