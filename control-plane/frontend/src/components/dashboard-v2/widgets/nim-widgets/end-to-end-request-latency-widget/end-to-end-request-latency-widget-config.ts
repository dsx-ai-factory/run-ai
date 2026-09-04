/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { EIntervalLabels } from "@/models/interval.model";
import {
  type IGenericMetricWidgetConfigWithCompositeResponse,
  type TMetricType,
  type ITooltipPoint,
} from "@/models/chart.model";
import { WorkloadMetricType, PodMetricType, type MetricsCompositeResponse } from "@/swagger-models/workloads-client";
import { dateUtil } from "@/utils/date.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { podService } from "@/services/control-plane/pod.service/pod.service";
import { chartUtil } from "@/utils/chart.util/chart.util";
import { formatToFixedOrZero } from "@/utils/format.util/format.util";
import {
  mapHistogramToRanges,
  parseHistogramToColumnSeries,
} from "@/components/dashboard-v2/widgets/chart-widgets/histogram-metric-widget/column-histogram-process-metrics-data.util";
import {
  NIM_HISTOGRAM_CHART_GRID,
  NIM_LEGEND_TEXT_STYLE,
} from "@/components/dashboard-v2/widgets/nim-widgets/nim-chart-legend.util";

export const END_TO_END_REQUEST_LATENCY_WIDGET_CHART_ID = "end-to-end-request-latency-widget-id";

const END_TO_END_REQUEST_LATENCY_THRESHOLDS = {
  VERY_FAST: 1, // 0-1 sec
  FAST: 2.5, // 1-2.5 sec
  MODERATE: 5, // 2.5-5 sec
  ACCEPTABLE: 10, // 5-10 sec
  SLOW: 15, // 10-15 sec
  VERY_SLOW: 20, // 15-20 sec
  CRITICAL: 40, // 20-40 sec
  EXTREME: 60, // 40-60 sec
} as const;

/**
 * Range keys for End to End Request Latency metrics in display order
 */
const endToEndRequestLatencyRangeKeys = ["0-1", "1-2.5", "2.5-5", "5-10", "10-15", "15-20", "20-40", "40-60", "60+"];

/**
 * Determines which range a bucket value belongs to for End to End Request Latency metrics
 */
function getEndToEndRequestLatencyRangeKey(bucketValue: number): string {
  const { VERY_FAST, FAST, MODERATE, ACCEPTABLE, SLOW, VERY_SLOW, CRITICAL, EXTREME } =
    END_TO_END_REQUEST_LATENCY_THRESHOLDS;

  if (bucketValue <= VERY_FAST) return endToEndRequestLatencyRangeKeys[0];
  if (bucketValue <= FAST) return endToEndRequestLatencyRangeKeys[1];
  if (bucketValue <= MODERATE) return endToEndRequestLatencyRangeKeys[2];
  if (bucketValue <= ACCEPTABLE) return endToEndRequestLatencyRangeKeys[3];
  if (bucketValue <= SLOW) return endToEndRequestLatencyRangeKeys[4];
  if (bucketValue <= VERY_SLOW) return endToEndRequestLatencyRangeKeys[5];
  if (bucketValue <= CRITICAL) return endToEndRequestLatencyRangeKeys[6];
  if (bucketValue <= EXTREME) return endToEndRequestLatencyRangeKeys[7];
  return endToEndRequestLatencyRangeKeys[8];
}

const commonConfig: Omit<IGenericMetricWidgetConfigWithCompositeResponse, "metricTypes" | "getData"> = {
  chartTitle: "End to end request latency",
  intervalLabel: EIntervalLabels.EndToEndRequestLatencyWidget,
  chartId: END_TO_END_REQUEST_LATENCY_WIDGET_CHART_ID,
  useCanonicalTimeBuckets: true,
  useConnectedTooltipAnchor: true,
  chartOptions: {
    type: "bar",
    height: 250,
    sharedTooltip: true,
    showSharedCrosshair: true,
    yAxisTitle: "Requests",
    grid: NIM_HISTOGRAM_CHART_GRID,
    yAxisNameGap: 38,
  } as IGenericMetricWidgetConfigWithCompositeResponse["chartOptions"] & { yAxisNameGap: number },
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

export function createWorkloadWidgetConfig(): IGenericMetricWidgetConfigWithCompositeResponse {
  async function getWorkloadRequestData(
    metricTypes: Array<TMetricType>,
    start: string,
    end: string,
    numberOfSamples: number,
    args: string[],
  ): Promise<MetricsCompositeResponse> {
    const workloadId = args[0];

    const metricsResponse: MetricsCompositeResponse = await workloadService.getWorkloadMetrics(
      workloadId,
      dateUtil.convertDateToISO(new Date(start)),
      dateUtil.convertDateToISO(new Date(end)),
      metricTypes as WorkloadMetricType[],
      numberOfSamples,
    );

    return metricsResponse;
  }

  return {
    ...commonConfig,
    metricTypes: [WorkloadMetricType.NimE2ERequestLatencySeconds],
    getData: getWorkloadRequestData,
    columnHistogramParser: (metricsResponse: MetricsCompositeResponse) => {
      return mapHistogramToRanges(
        parseHistogramToColumnSeries(metricsResponse),
        getEndToEndRequestLatencyRangeKey,
        endToEndRequestLatencyRangeKeys,
        undefined,
        "sec",
      );
    },
  };
}

export function createPodWidgetConfig(): IGenericMetricWidgetConfigWithCompositeResponse {
  async function getPodRequestData(
    metricTypes: Array<TMetricType>,
    start: string,
    end: string,
    numberOfSamples: number,
    args: string[],
  ): Promise<MetricsCompositeResponse> {
    const workloadId = args[0];
    const podId = args[1];

    const metricsResponse: MetricsCompositeResponse = await podService.getPodMetrics(
      workloadId,
      podId,
      dateUtil.convertDateToISO(new Date(start)),
      dateUtil.convertDateToISO(new Date(end)),
      metricTypes as PodMetricType[],
      numberOfSamples,
    );

    return metricsResponse;
  }

  return {
    ...commonConfig,
    metricTypes: [PodMetricType.NimE2ERequestLatencySeconds],
    getData: getPodRequestData,
    columnHistogramParser: (metricsResponse: MetricsCompositeResponse) => {
      return mapHistogramToRanges(
        parseHistogramToColumnSeries(metricsResponse),
        getEndToEndRequestLatencyRangeKey,
        endToEndRequestLatencyRangeKeys,
        undefined,
        "sec",
      );
    },
  };
}

export const workloadEndToEndRequestLatencyWidgetConfig = createWorkloadWidgetConfig();
export const podEndToEndRequestLatencyWidgetConfig = createPodWidgetConfig();
