/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { EIntervalLabels } from "@/models/interval.model";
import type {
  IGenericMetricWidgetConfig,
  IGenericMetricWidgetConfigWithCompositeResponse,
  THistogramMetricData,
  TMetricType,
} from "@/models/chart.model";
import {
  PodMetricType,
  WorkloadMetricType,
  type HistogramValue,
  type MeasurementResponseValuesInner,
  type MetricsCompositeResponse,
  type MetricsResponse,
} from "@/swagger-models/workloads-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { podService } from "@/services/control-plane/pod.service/pod.service";
import { dateUtil } from "@/utils/date.util";
import {
  mapHistogramToRanges,
  parseHistogramToColumnSeries,
} from "@/components/dashboard-v2/widgets/chart-widgets/histogram-metric-widget/column-histogram-process-metrics-data.util";
import {
  NIM_HISTOGRAM_CHART_GRID,
  NIM_LEGEND_TEXT_STYLE,
  NIM_LINE_LEGEND_ICON,
  NIM_METRIC_CHART_GRID,
} from "@/components/dashboard-v2/widgets/nim-widgets/nim-chart-legend.util";
import { getNimWidgetChartOptions } from "@/components/dashboard-v2/widgets/chart-widgets/inference-chart-options.util";

type MetricPair = {
  workload: WorkloadMetricType;
  pod: PodMetricType;
  label: string;
};

type LinearWidgetDefinition = {
  id: string;
  workloadConfig: IGenericMetricWidgetConfig;
  podConfig: IGenericMetricWidgetConfig;
};

type HistogramWidgetDefinition = {
  id: string;
  workloadConfig: IGenericMetricWidgetConfigWithCompositeResponse;
  podConfig: IGenericMetricWidgetConfigWithCompositeResponse;
};

const PERCENTILES = [
  { key: "p50", label: "P50" },
  { key: "p95", label: "P95" },
  { key: "p99", label: "P99" },
] as const;

const VLLM_LATENCY_PHASE_CHART_GRID = {
  ...NIM_METRIC_CHART_GRID,
  bottom: 110,
};

async function getMetrics(
  isPod: boolean,
  metricTypes: Array<TMetricType>,
  start: string,
  end: string,
  numberOfSamples: number,
  args: string[],
): Promise<MetricsResponse> {
  const startIso = dateUtil.convertDateToISO(new Date(start));
  const endIso = dateUtil.convertDateToISO(new Date(end));
  return isPod
    ? podService.getPodMetrics(args[0], args[1], startIso, endIso, metricTypes as PodMetricType[], numberOfSamples)
    : workloadService.getWorkloadMetrics(
        args[0],
        startIso,
        endIso,
        metricTypes as WorkloadMetricType[],
        numberOfSamples,
      );
}

function createLinearConfig(
  title: string,
  id: string,
  intervalLabel: EIntervalLabels,
  metricPairs: MetricPair[],
  isPod: boolean,
  options: { yAxisTitle?: string; percentageRatio?: boolean; groupByFinishedReason?: boolean } = {},
): IGenericMetricWidgetConfig {
  const metrics = metricPairs.map((pair) => (isPod ? pair.pod : pair.workload));
  const labels = Object.fromEntries(metricPairs.map((pair) => [isPod ? pair.pod : pair.workload, pair.label]));
  const emptySeriesNames = options.groupByFinishedReason ? [] : metricPairs.map(({ label }) => label);

  return {
    chartTitle: title,
    intervalLabel,
    chartId: id,
    useCanonicalTimeBuckets: true,
    useConnectedTooltipAnchor: true,
    chartOptions: {
      ...getNimWidgetChartOptions(options.yAxisTitle || ""),
      grid: NIM_METRIC_CHART_GRID,
    },
    tooltipSuffix: options.percentageRatio ? "%" : undefined,
    isPercentage: options.percentageRatio,
    emptySeriesNames,
    legendIconMap: Object.fromEntries(emptySeriesNames.map((name) => [name, NIM_LINE_LEGEND_ICON])),
    legendTextStyle: NIM_LEGEND_TEXT_STYLE,
    metricTypes: metrics,
    getData: async (metricTypes, start, end, numberOfSamples, args) => {
      const response = await getMetrics(isPod, metricTypes, start, end, numberOfSamples, args);
      response.measurements = response.measurements.map((measurement) => {
        const finishedReason = measurement.labels?.finished_reason;
        const type = options.groupByFinishedReason && finishedReason ? finishedReason : labels[measurement.type];
        return {
          ...measurement,
          type: type || measurement.type,
          values: (measurement.values || []).map((value) => ({
            ...value,
            value: options.percentageRatio ? String(Number(value.value) * 100) : value.value,
          })),
        };
      });
      return response;
    },
    measurementParser: (response) =>
      Object.fromEntries(
        response.measurements.map((measurement) => [
          measurement.type,
          (measurement.values || []) as MeasurementResponseValuesInner[],
        ]),
      ),
  };
}

function createLinearWidget(
  title: string,
  id: string,
  intervalLabel: EIntervalLabels,
  metricPairs: MetricPair[],
  options?: { yAxisTitle?: string; percentageRatio?: boolean; groupByFinishedReason?: boolean },
): LinearWidgetDefinition {
  return {
    id,
    workloadConfig: createLinearConfig(title, id, intervalLabel, metricPairs, false, options),
    podConfig: createLinearConfig(title, id, intervalLabel, metricPairs, true, options),
  };
}

function parsePercentiles(values: HistogramValue[]): THistogramMetricData {
  const result: THistogramMetricData = {};
  PERCENTILES.forEach(({ label }) => {
    result[label] = [];
  });
  values.forEach(({ timestamp, data }) => {
    if (!timestamp || !data) return;
    PERCENTILES.forEach(({ key, label }) => {
      if (data[key] !== undefined) result[label].push({ timestamp, value: data[key] });
    });
  });
  return result;
}

function createHistogramConfig(
  title: string,
  id: string,
  intervalLabel: EIntervalLabels,
  metricPair: MetricPair,
  isPod: boolean,
  parser: Pick<IGenericMetricWidgetConfigWithCompositeResponse, "histogramParser" | "columnHistogramParser">,
): IGenericMetricWidgetConfigWithCompositeResponse {
  return {
    chartTitle: title,
    intervalLabel,
    chartId: id,
    useCanonicalTimeBuckets: true,
    useConnectedTooltipAnchor: true,
    chartOptions: parser.histogramParser
      ? {
          ...getNimWidgetChartOptions("Seconds"),
          grid: NIM_METRIC_CHART_GRID,
        }
      : {
          type: "bar",
          height: 250,
          sharedTooltip: true,
          showSharedCrosshair: true,
          yAxisTitle: "Requests",
          grid: NIM_HISTOGRAM_CHART_GRID,
        },
    emptySeriesNames: parser.histogramParser ? PERCENTILES.map(({ label }) => label) : undefined,
    legendTextStyle: NIM_LEGEND_TEXT_STYLE,
    metricTypes: [isPod ? metricPair.pod : metricPair.workload],
    getData: (metricTypes, start, end, numberOfSamples, args) =>
      getMetrics(isPod, metricTypes, start, end, numberOfSamples, args) as Promise<MetricsCompositeResponse>,
    ...parser,
  };
}

function createPercentileWidget(
  title: string,
  id: string,
  intervalLabel: EIntervalLabels,
  metricPair: MetricPair,
): HistogramWidgetDefinition {
  const parser = {
    histogramParser: (response: MetricsCompositeResponse) => parsePercentiles(response.histogram?.[0]?.values || []),
  };
  return {
    id,
    workloadConfig: createHistogramConfig(title, id, intervalLabel, metricPair, false, parser),
    podConfig: createHistogramConfig(title, id, intervalLabel, metricPair, true, parser),
  };
}

function createLatencyPhaseConfig(
  metricPairs: MetricPair[],
  isPod: boolean,
): IGenericMetricWidgetConfigWithCompositeResponse {
  const metrics = metricPairs.map((metricPair) => (isPod ? metricPair.pod : metricPair.workload));
  const labels = Object.fromEntries(
    metricPairs.map((metricPair) => [isPod ? metricPair.pod : metricPair.workload, metricPair.label]),
  );
  return {
    chartTitle: "Latency phase breakdown",
    intervalLabel: EIntervalLabels.VllmLatencyPhaseWidget,
    chartId: "vllm-latency-phase",
    useCanonicalTimeBuckets: true,
    useConnectedTooltipAnchor: true,
    chartOptions: {
      ...getNimWidgetChartOptions("Seconds"),
      height: 300,
      adaptiveLegendLayout: true,
      grid: VLLM_LATENCY_PHASE_CHART_GRID,
    },
    legendTextStyle: NIM_LEGEND_TEXT_STYLE,
    metricTypes: metrics,
    getData: (metricTypes, start, end, numberOfSamples, args) =>
      getMetrics(isPod, metricTypes, start, end, numberOfSamples, args) as Promise<MetricsCompositeResponse>,
    histogramParser: (response) =>
      Object.assign(
        {},
        ...(response.histogram || []).map((series) => {
          const phase = labels[series.type] || series.type;
          const parsed = parsePercentiles(series.values);
          return Object.fromEntries(
            Object.entries(parsed).map(([percentile, values]) => [`${phase} ${percentile}`, values]),
          );
        }),
      ),
  };
}

function createDistributionWidget(
  title: string,
  id: string,
  intervalLabel: EIntervalLabels,
  metricPair: MetricPair,
  thresholds: number[],
): HistogramWidgetDefinition {
  const rangeNames = thresholds.map((threshold, index) => `${index ? thresholds[index - 1] : 0}-${threshold}`);
  rangeNames.push(`${thresholds.at(-1)}+`);
  const rangeKey = (value: number) => {
    const index = thresholds.findIndex((threshold) => value <= threshold);
    return rangeNames[index === -1 ? rangeNames.length - 1 : index];
  };
  const parser = {
    columnHistogramParser: (response: MetricsCompositeResponse) =>
      mapHistogramToRanges(parseHistogramToColumnSeries(response), rangeKey, rangeNames, undefined, "sec"),
  };
  return {
    id,
    workloadConfig: createHistogramConfig(title, id, intervalLabel, metricPair, false, parser),
    podConfig: createHistogramConfig(title, id, intervalLabel, metricPair, true, parser),
  };
}

const pair = (workload: WorkloadMetricType, pod: PodMetricType, label: string): MetricPair => ({
  workload,
  pod,
  label,
});

export const vllmLinearWidgets = [
  createLinearWidget("Request concurrency by status", "vllm-concurrency", EIntervalLabels.VllmConcurrencyWidget, [
    pair(WorkloadMetricType.VllmNumRequestsRunning, PodMetricType.VllmNumRequestsRunning, "Running"),
    pair(WorkloadMetricType.VllmNumRequestsWaiting, PodMetricType.VllmNumRequestsWaiting, "Waiting"),
  ]),
  createLinearWidget(
    "Request count by status",
    "vllm-throughput",
    EIntervalLabels.VllmThroughputWidget,
    [pair(WorkloadMetricType.VllmRequestSuccess, PodMetricType.VllmRequestSuccess, "Finished")],
    { yAxisTitle: "Requests / sec", groupByFinishedReason: true },
  ),
  createLinearWidget(
    "Request success rate",
    "vllm-success-rate",
    EIntervalLabels.VllmSuccessRateWidget,
    [pair(WorkloadMetricType.VllmRequestSuccessRate, PodMetricType.VllmRequestSuccessRate, "Success rate")],
    { yAxisTitle: "Percent" },
  ),
  createLinearWidget(
    "GPU KV cache utilization",
    "vllm-kv-cache",
    EIntervalLabels.VllmKvCacheWidget,
    [pair(WorkloadMetricType.VllmKvCacheUsagePerc, PodMetricType.VllmKvCacheUsagePerc, "KV cache")],
    { yAxisTitle: "Percent", percentageRatio: true },
  ),
  createLinearWidget(
    "Token throughput",
    "vllm-token-throughput",
    EIntervalLabels.VllmTokenThroughputWidget,
    [
      pair(WorkloadMetricType.VllmPromptTokensPerSecond, PodMetricType.VllmPromptTokensPerSecond, "Prompt"),
      pair(WorkloadMetricType.VllmGenerationTokensPerSecond, PodMetricType.VllmGenerationTokensPerSecond, "Generation"),
    ],
    { yAxisTitle: "Tokens / sec" },
  ),
] as const;

export const vllmHistogramWidgets = [
  createDistributionWidget(
    "Time to first token (TTFT)",
    "vllm-ttft-distribution",
    EIntervalLabels.VllmTtftDistributionWidget,
    pair(WorkloadMetricType.VllmTimeToFirstTokenSeconds, PodMetricType.VllmTimeToFirstTokenSeconds, "TTFT"),
    [0.25, 0.5, 1, 2.5, 5],
  ),
  createPercentileWidget(
    "Time to first token (TTFT) by percentiles",
    "vllm-ttft-percentiles",
    EIntervalLabels.VllmTtftPercentilesWidget,
    pair(
      WorkloadMetricType.VllmTimeToFirstTokenSecondsPercentiles,
      PodMetricType.VllmTimeToFirstTokenSecondsPercentiles,
      "TTFT",
    ),
  ),
  createDistributionWidget(
    "End to end request latency",
    "vllm-e2e-distribution",
    EIntervalLabels.VllmE2EDistributionWidget,
    pair(WorkloadMetricType.VllmE2ERequestLatencySeconds, PodMetricType.VllmE2ERequestLatencySeconds, "End-to-end"),
    [1, 2.5, 5, 10, 20, 40, 60],
  ),
  createPercentileWidget(
    "End to end request latency by percentiles",
    "vllm-e2e-percentiles",
    EIntervalLabels.VllmE2EPercentilesWidget,
    pair(
      WorkloadMetricType.VllmE2ERequestLatencySecondsPercentiles,
      PodMetricType.VllmE2ERequestLatencySecondsPercentiles,
      "End-to-end",
    ),
  ),
  createPercentileWidget(
    "Queue time percentiles",
    "vllm-queue-percentiles",
    EIntervalLabels.VllmQueuePercentilesWidget,
    pair(
      WorkloadMetricType.VllmRequestQueueTimeSecondsPercentiles,
      PodMetricType.VllmRequestQueueTimeSecondsPercentiles,
      "Queue",
    ),
  ),
  {
    id: "vllm-latency-phase",
    workloadConfig: createLatencyPhaseConfig(
      [
        pair(
          WorkloadMetricType.VllmRequestPrefillTimeSecondsPercentiles,
          PodMetricType.VllmRequestPrefillTimeSecondsPercentiles,
          "Prefill",
        ),
        pair(
          WorkloadMetricType.VllmRequestDecodeTimeSecondsPercentiles,
          PodMetricType.VllmRequestDecodeTimeSecondsPercentiles,
          "Decode",
        ),
        pair(
          WorkloadMetricType.VllmInterTokenLatencySecondsPercentiles,
          PodMetricType.VllmInterTokenLatencySecondsPercentiles,
          "Inter-token latency",
        ),
      ],
      false,
    ),
    podConfig: createLatencyPhaseConfig(
      [
        pair(
          WorkloadMetricType.VllmRequestPrefillTimeSecondsPercentiles,
          PodMetricType.VllmRequestPrefillTimeSecondsPercentiles,
          "Prefill",
        ),
        pair(
          WorkloadMetricType.VllmRequestDecodeTimeSecondsPercentiles,
          PodMetricType.VllmRequestDecodeTimeSecondsPercentiles,
          "Decode",
        ),
        pair(
          WorkloadMetricType.VllmInterTokenLatencySecondsPercentiles,
          PodMetricType.VllmInterTokenLatencySecondsPercentiles,
          "Inter-token latency",
        ),
      ],
      true,
    ),
  },
] as const;
