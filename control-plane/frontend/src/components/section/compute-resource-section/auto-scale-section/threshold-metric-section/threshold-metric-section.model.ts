export interface IThresholdMetricData {
  thresholdMetric: string | undefined;
  thresholdValue: number | undefined;
}

export const METRIC_THROUGHPUT = "throughput";
export const METRIC_CONCURRENCY = "concurrency";
export const METRIC_LATENCY = "latency";
export const DEFAULT_THRESHOLD_VALUE = 100;

export interface IThresholdMetricOption {
  id: string;
  name: string;
}

export const thresholdMetricOptions: Array<IThresholdMetricOption> = [
  {
    id: METRIC_THROUGHPUT,
    name: "Throughput (Requests/sec)",
  },
  {
    id: METRIC_CONCURRENCY,
    name: "Concurrency (Requests)",
  },
  {
    id: METRIC_LATENCY,
    name: "Latency (millisecond)",
  },
];

export enum EThresholdMetricSectionEvents {
  UpdateModelValue = "update:model-value",
}
