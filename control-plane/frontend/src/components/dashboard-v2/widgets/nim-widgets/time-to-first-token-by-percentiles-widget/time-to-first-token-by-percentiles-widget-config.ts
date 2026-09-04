/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { EIntervalLabels } from "@/models/interval.model";
import {
  type THistogramMetricData,
  type IGenericMetricWidgetConfigWithCompositeResponse,
  type TMetricType,
} from "@/models/chart.model";
import {
  WorkloadMetricType,
  PodMetricType,
  type MetricsCompositeResponse,
  type HistogramValue,
} from "@/swagger-models/workloads-client";
import { dateUtil } from "@/utils/date.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { podService } from "@/services/control-plane/pod.service/pod.service";
import {
  NIM_LEGEND_TEXT_STYLE,
  NIM_LINE_LEGEND_ICON,
  NIM_METRIC_CHART_GRID,
} from "@/components/dashboard-v2/widgets/nim-widgets/nim-chart-legend.util";
import {
  getNimWidgetChartOptions,
  NIM_WIDGET_Y_AXIS_LABEL_STEPS,
} from "@/components/dashboard-v2/widgets/chart-widgets/inference-chart-options.util";

export const TIME_TO_FIRST_TOKEN_BY_PERCENTILES_WIDGET_CHART_ID = "time-to-first-token-by-percentiles-widget-id";
const PERCENTILE_SERIES_CONFIG = [
  { key: "p50", label: "P50 requests" },
  { key: "p90", label: "P90 requests" },
  { key: "p99", label: "P99 requests" },
] as const;
const PERCENTILE_SERIES_NAMES = PERCENTILE_SERIES_CONFIG.map(({ label }) => label);

/**
 * Converts an array of HistogramValue objects to a map format
 *
 * Example input:
 * [{
 *   "timestamp": "2025-09-07T09:24:43.92Z",
 *   "data": {
 *     "p50": "0",
 *     "p90": "0",
 *     "p99": "0"
 *   }
 * },
 * {
 *   "timestamp": "2025-09-07T10:24:43.92Z",
 *   "data": {
 *     "p50": "1",
 *     "p90": "2",
 *     "p99": "3"
 *   }
 * }
 * ]
 *
 * Example output:
 * {
 *  "p50": [{ timestamp: "2025-09-07T09:24:43.92Z", value: "0" }, { timestamp: "2025-09-07T10:24:43.92Z", value: "1" }],
 *  "p90": [{ timestamp: "2025-09-07T09:24:43.92Z", value: "0" }, { timestamp: "2025-09-07T10:24:43.92Z", value: "2" }],
 *  "p99": [{ timestamp: "2025-09-07T09:24:43.92Z", value: "0" }, { timestamp: "2025-09-07T10:24:43.92Z", value: "3" }]
 * }
 *
 * @param histogramValues - Array of HistogramValue objects to convert
 * @returns Map of percentile keys to arrays of timestamp-value pairs
 */
export function convertHistogramValueToArray(
  histogramValues: HistogramValue[],
): Record<string, Array<{ timestamp: string; value: string }>> {
  const result = new Map<string, Array<{ timestamp: string; value: string }>>();

  histogramValues.forEach((histogramValue) => {
    if (!histogramValue.timestamp || !histogramValue.data) {
      return;
    }

    PERCENTILE_SERIES_CONFIG.forEach(({ key, label }) => {
      const value = histogramValue.data?.[key];
      if (value === undefined) return;

      const points = result.get(label) ?? [];
      points.push({
        timestamp: histogramValue.timestamp as string,
        value,
      });
      result.set(label, points);
    });
  });

  return Object.fromEntries(result);
}

const commonConfig: Omit<IGenericMetricWidgetConfigWithCompositeResponse, "metricTypes" | "getData"> = {
  chartTitle: "Time to first token (TTFT) by percentiles",
  intervalLabel: EIntervalLabels.TimeToFirstTokenByPercentilesWidget,
  chartId: TIME_TO_FIRST_TOKEN_BY_PERCENTILES_WIDGET_CHART_ID,
  useCanonicalTimeBuckets: true,
  useConnectedTooltipAnchor: true,
  chartOptions: {
    ...getNimWidgetChartOptions("Milliseconds"),
    grid: NIM_METRIC_CHART_GRID,
  },
  tooltipSuffix: " ms",
  yAxisLabelFormatting: {
    steps: NIM_WIDGET_Y_AXIS_LABEL_STEPS,
  },
  emptySeriesNames: PERCENTILE_SERIES_NAMES,
  legendIconMap: Object.fromEntries(PERCENTILE_SERIES_NAMES.map((name) => [name, NIM_LINE_LEGEND_ICON])),
  legendTextStyle: NIM_LEGEND_TEXT_STYLE,
};

function histogramParser(metricsResponse: MetricsCompositeResponse): THistogramMetricData {
  if (!metricsResponse.histogram?.[0]?.values) {
    return {};
  }

  return convertHistogramValueToArray(metricsResponse.histogram[0].values);
}

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
    metricTypes: [WorkloadMetricType.NimTimeToFirstTokenSecondsPercentiles],
    getData: getWorkloadRequestData,
    histogramParser,
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
    metricTypes: [PodMetricType.NimTimeToFirstTokenSecondsPercentiles],
    getData: getPodRequestData,
    histogramParser,
  };
}

export const workloadTimeToFirstTokenByPercentilesWidgetConfig = createWorkloadWidgetConfig();
export const podTimeToFirstTokenByPercentilesWidgetConfig = createPodWidgetConfig();
