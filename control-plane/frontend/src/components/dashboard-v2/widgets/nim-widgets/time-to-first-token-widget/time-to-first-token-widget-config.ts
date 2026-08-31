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
  parseHistogramToColumnSeries,
  mapHistogramToRanges,
} from "@/components/dashboard-v2/widgets/chart-widgets/histogram-metric-widget/column-histogram-process-metrics-data.util";
import {
  NIM_HISTOGRAM_CHART_GRID,
  NIM_LEGEND_TEXT_STYLE,
} from "@/components/dashboard-v2/widgets/nim-widgets/nim-chart-legend.util";

export const TIME_TO_FIRST_TOKEN_WIDGET_CHART_ID = "time-to-first-token-widget-id";

const TIME_TO_FIRST_TOKEN_THRESHOLDS = {
  VERY_FAST: 0.25, // 0-0.25 sec
  FAST: 0.5, // 0.25-0.5 sec
  MODERATE: 1, // 0.5-1 sec
  ACCEPTABLE: 2.5, // 1-2.5 sec
  SLOW: 5, // 2.5-5 sec
} as const;

/**
 * Range keys for Time to First Token metrics in display order
 */
const timeToFirstTokenRangeKeys = ["0-0.25", "0.25-0.5", "0.5-1", "1-2.5", "2.5-5", "5+"];

/**
 * Determines which range a bucket value belongs to for Time to First Token metrics
 */
function getTimeToFirstTokenRangeKey(bucketValue: number): string {
  const { VERY_FAST, FAST, MODERATE, ACCEPTABLE, SLOW } = TIME_TO_FIRST_TOKEN_THRESHOLDS;

  if (bucketValue <= VERY_FAST) return timeToFirstTokenRangeKeys[0];
  if (bucketValue <= FAST) return timeToFirstTokenRangeKeys[1];
  if (bucketValue <= MODERATE) return timeToFirstTokenRangeKeys[2];
  if (bucketValue <= ACCEPTABLE) return timeToFirstTokenRangeKeys[3];
  if (bucketValue <= SLOW) return timeToFirstTokenRangeKeys[4];
  return timeToFirstTokenRangeKeys[5];
}

const commonConfig: Omit<IGenericMetricWidgetConfigWithCompositeResponse, "metricTypes" | "getData"> = {
  chartTitle: "Time to first token (TTFT)",
  intervalLabel: EIntervalLabels.TimeToFirstTokenWidget,
  chartId: TIME_TO_FIRST_TOKEN_WIDGET_CHART_ID,
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
    metricTypes: [WorkloadMetricType.NimTimeToFirstTokenSeconds],
    getData: getWorkloadRequestData,
    columnHistogramParser: (metricsResponse: MetricsCompositeResponse) => {
      return mapHistogramToRanges(
        parseHistogramToColumnSeries(metricsResponse),
        getTimeToFirstTokenRangeKey,
        timeToFirstTokenRangeKeys,
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
    metricTypes: [PodMetricType.NimTimeToFirstTokenSeconds],
    getData: getPodRequestData,
    columnHistogramParser: (metricsResponse: MetricsCompositeResponse) => {
      return mapHistogramToRanges(
        parseHistogramToColumnSeries(metricsResponse),
        getTimeToFirstTokenRangeKey,
        timeToFirstTokenRangeKeys,
        undefined,
        "sec",
      );
    },
  };
}

export const workloadTimeToFirstTokenWidgetConfig = createWorkloadWidgetConfig();
export const podTimeToFirstTokenWidgetConfig = createPodWidgetConfig();
